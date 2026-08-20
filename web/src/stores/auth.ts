import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getMe, login as loginApi } from '@/api'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const user = ref<User | null>(null)
  const permissions = ref<string[]>([])
  const loaded = ref(false)

  const isLoggedIn = computed(() => Boolean(token.value))

  async function login(username: string, password: string) {
    const { data } = await loginApi({ username, password })
    token.value = data.data.token
    localStorage.setItem('admin_token', token.value)
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
    localStorage.removeItem('admin_token')
  }

  return { token, user, permissions, loaded, isLoggedIn, login, fetchMe, hasPermission, logout }
})
