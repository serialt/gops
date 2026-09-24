import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getMe, login as loginApi, type LoginRequest } from '@/api/auth'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)
  const permissions = ref<string[]>([])
  const loaded = ref(false)

  const isLoggedIn = computed(() => Boolean(token.value))

  async function login(request: LoginRequest) {
	const { data } = await loginApi(request)
	if (data.data.mfa_required) return data.data
	await acceptLogin(data.data)
	return data.data
  }

  async function acceptLogin(result: {access_token:string;token:string;refresh_token:string}) {
	token.value = result.access_token || result.token
	localStorage.setItem('token', token.value)
	localStorage.setItem('refresh_token', result.refresh_token)
	await fetchMe()
  }

  async function fetchMe() {
    const { data } = await getMe()
    user.value = data.data.user
    permissions.value = data.data.permissions || []
    loaded.value = true
  }

  function hasPermission(permission?: string) {
    return !permission || permissions.value.includes(permission)
  }

  function logout() {
    token.value = ''
    user.value = null
    permissions.value = []
    loaded.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
  }

	return { token, user, permissions, loaded, isLoggedIn, login, acceptLogin, fetchMe, hasPermission, logout }
})
