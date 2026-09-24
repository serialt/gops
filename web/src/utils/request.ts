import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 扩展 Axios 实例方法，调用 API 时可直接声明响应数据泛型。
declare module 'axios' {
  interface AxiosInstance {
    get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>>
    post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>>
    put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>>
    delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>>
    patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>>
  }
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

let isRedirecting = false
let refreshPromise: Promise<string> | null = null

function clearTokens() {
  localStorage.removeItem('token')
  localStorage.removeItem('refresh_token')
}

async function renewAccessToken() {
  if (refreshPromise) return refreshPromise
  const refreshToken = localStorage.getItem('refresh_token')
  if (!refreshToken) throw new Error('missing refresh token')
  refreshPromise = axios.post<{token:string;access_token:string;refresh_token:string}>('/api/v1/auth/refresh', { refresh_token: refreshToken })
    .then(({ data }) => {
      const accessToken = data.data.access_token || data.data.token
      localStorage.setItem('token', accessToken)
      localStorage.setItem('refresh_token', data.data.refresh_token)
      return accessToken
    })
    .finally(() => { refreshPromise = null })
  return refreshPromise
}

function redirectToError(status: 401 | 403 | 429) {
  const targetPath = `/${status}`
  if (isRedirecting || location.pathname === targetPath) return
  isRedirecting = true
  const current = `${location.pathname}${location.search}${location.hash}`
  location.assign(`${targetPath}?redirect=${encodeURIComponent(current)}`)
}

function redirectToLogin() {
  if (isRedirecting || location.pathname === '/login') return
  isRedirecting = true
  const current = `${location.pathname}${location.search}${location.hash}`
  location.assign(`/login?redirect=${encodeURIComponent(current)}`)
}

request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error),
)

request.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '请求失败'
    if (status === 401) {
      const original = error.config as AxiosRequestConfig & { _retry?: boolean }
      const isAuthRequest = String(original?.url || '').includes('/auth/login') || String(original?.url || '').includes('/auth/refresh')
      if (!original?._retry && !isAuthRequest && localStorage.getItem('refresh_token')) {
        original._retry = true
        try {
          const token = await renewAccessToken()
          original.headers = { ...original.headers, Authorization: `Bearer ${token}` }
          return request(original)
        } catch {
          clearTokens()
        }
      } else {
        clearTokens()
      }
      if (location.pathname !== '/login') redirectToLogin()
      else ElMessage.error(message)
    } else if (status === 403) {
      redirectToError(403)
    } else if (status === 429) {
      redirectToError(429)
    } else {
      ElMessage.error(message)
    }
    return Promise.reject(error)
  },
)

export default request
