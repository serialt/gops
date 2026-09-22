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

function redirectToError(status: 401 | 403 | 429) {
  const targetPath = `/${status}`
  if (isRedirecting || location.pathname === targetPath) return
  isRedirecting = true
  const current = `${location.pathname}${location.search}${location.hash}`
  location.assign(`${targetPath}?redirect=${encodeURIComponent(current)}`)
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
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '请求失败'
    if (status === 401) {
      localStorage.removeItem('token')
      if (location.pathname !== '/login') redirectToError(401)
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
