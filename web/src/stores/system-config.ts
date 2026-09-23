import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { getSystemConfig } from '@/api/system-config'

export const useSystemConfigStore = defineStore('system-config', () => {
  const fallbackName = 'Gops'
  const systemName = ref(localStorage.getItem('system_name')?.trim() || fallbackName)
  const systemDescription = ref('')

  async function fetchConfig() {
    try {
      const { data } = await getSystemConfig()
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
