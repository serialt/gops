import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getSystemConfig } from '@/api/system-config'

export const useSystemConfigStore = defineStore('system-config', () => {
  const systemName = ref('Admin')
  const systemDescription = ref('')

  async function fetchConfig() {
    const { data } = await getSystemConfig()
    systemName.value = data.data.system_name || 'Admin'
    systemDescription.value = data.data.system_description || ''
  }

  function updateConfig(name: string, description: string) {
    systemName.value = name || 'Admin'
    systemDescription.value = description || ''
  }

  return { systemName, systemDescription, fetchConfig, updateConfig }
})
