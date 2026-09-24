import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { getLoginSecurityConfig } from '@/api/auth'

export const useSystemConfigStore = defineStore('system-config', () => {
  const fallbackName = 'Gops'
  const systemName = ref(localStorage.getItem('system_name')?.trim() || fallbackName)
  const systemDescription = ref('')

  async function fetchConfig() {
    try {
      const { data } = await getLoginSecurityConfig()
      systemName.value = data.data.system_name?.trim() || fallbackName
      systemDescription.value = data.data.system_description || ''
    } catch {
      systemName.value = localStorage.getItem('system_name')?.trim() || fallbackName
    }
  }

  function updateConfig(name: string, description: string) {
    systemName.value = name?.trim() || fallbackName
    systemDescription.value = description || ''
  }

  watch(systemName, (name) => {
    localStorage.setItem('system_name', name)
    document.title = name
  }, { immediate: true })

  return { systemName, systemDescription, fetchConfig, updateConfig }
})
