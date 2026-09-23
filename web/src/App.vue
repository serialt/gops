<template>
  <router-view />
</template>
<script setup lang="ts">
import { useSystemConfigStore } from '@/stores/system-config'

useSystemConfigStore()
</script>
