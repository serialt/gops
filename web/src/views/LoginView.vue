<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const loading = ref(false)
const form = reactive({ username: 'admin', password: '' })

async function submit() {
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    await router.replace((route.query.redirect as string) || '/')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <el-card class="login-card" shadow="always">
      <div class="title">Gin Vue Casbin Admin</div>
      <div class="subtitle">基于 API 路径和 HTTP 方法的 RBAC 权限管理</div>
      <el-form :model="form" label-position="top" @keyup.enter="submit">
        <el-form-item label="用户名"><el-input v-model="form.username" size="large" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password size="large" /></el-form-item>
        <el-button type="primary" size="large" :loading="loading" class="submit" @click="submit">登录</el-button>
      </el-form>
      <div class="hint">默认账号：admin</div>
    </el-card>
  </div>
</template>

<style scoped>
.login-page { min-height: 100%; display: flex; align-items: center; justify-content: center; background: radial-gradient(circle at top left, #dbeafe, #f8fafc 45%, #e0e7ff); }
.login-card { width: 430px; border: 0; border-radius: 16px; padding: 14px; }
.title { font-size: 26px; font-weight: 700; margin-bottom: 8px; }
.subtitle { color: #6b7280; margin-bottom: 28px; }
.submit { width: 100%; margin-top: 8px; }
.hint { margin-top: 18px; color: #9ca3af; text-align: center; font-size: 13px; }
</style>
