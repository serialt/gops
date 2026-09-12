<script setup lang="ts">
import { reactive, ref } from 'vue'
import { Lock, UserFilled } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import illustrationUrl from '@/assets/login/illustration.svg'
import backgroundUrl from '@/assets/login/bg.png'
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
  <main class="login-page" :style="{ backgroundImage: `url(${backgroundUrl})` }">
    <section class="login-visual" aria-hidden="true">
      <img class="login-illustration" :src="illustrationUrl" alt="" />
    </section>
    <section class="login-panel">
      <div class="login-form-wrap">
        <div class="login-heading">
          <span class="login-logo">Gops</span>
          <div><h1>欢迎登录</h1><p>登录管理控制台，开始你的工作。</p></div>
        </div>
        <el-form :model="form" class="login-form" @keyup.enter="submit">
          <el-form-item>
            <el-input v-model="form.username" size="large" placeholder="请输入用户名" autocomplete="username">
              <template #prefix><el-icon><UserFilled /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-input v-model="form.password" type="password" show-password size="large" placeholder="请输入密码" autocomplete="current-password">
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-button type="primary" size="large" :loading="loading" class="submit" @click="submit">登录</el-button>
        </el-form>
        <p class="login-tip">使用您的账号和密码登录系统</p>
      </div>
    </section>
  </main>
</template>

<style scoped>
.login-page { display: grid; grid-template-columns: minmax(520px, 1.15fr) minmax(420px, .85fr); min-height: 100%; overflow: hidden; background-color: #fff; background-repeat: no-repeat; background-position: center; background-size: cover; }
.login-visual { position: relative; min-width: 0; overflow: hidden; }
.brand-mark, .login-logo { display: inline-flex; align-items: center; justify-content: center; width: 42px; height: 42px; border-radius: 11px; background: linear-gradient(135deg, #6078f4, #7a57e8); color: #fff; font-size: 17px; font-weight: 800; box-shadow: 0 10px 24px rgba(48, 73, 180, .26); }
.login-illustration { position: absolute; left: 10%; bottom: 15%; width: min(80%, 680px); max-height: 58%; object-fit: contain; object-position: left bottom; }
.login-panel { display: flex; align-items: center; justify-content: center; min-width: 0; padding: 40px; background: #fff; }
.login-form-wrap { width: min(360px, 100%); }
.login-heading { display: flex; align-items: center; gap: 13px; margin-bottom: 32px; }
.login-heading h1 { margin: 0 0 4px; color: #182235; font-size: 24px; line-height: 1.15; }
.login-heading p { margin: 0; color: #98a2b3; font-size: 12px; }
.login-logo { width: 46px; height: 46px; border-radius: 12px; flex: 0 0 auto; }
.login-form :deep(.el-form-item) { margin-bottom: 16px; }
.login-form :deep(.el-input__wrapper) { min-height: 42px; border-radius: 8px; box-shadow: 0 0 0 1px #e3e8f1 inset; }
.login-form :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 1px var(--primary) inset; }
.login-form :deep(.el-input__prefix) { color: #a6afbf; }
.submit { width: 100%; height: 42px; margin-top: 4px; border-radius: 8px; font-weight: 600; }
.login-tip { margin: 18px 0 0; color: #a2aaba; font-size: 11px; text-align: center; }

@media (max-width: 900px) {
  .login-page { grid-template-columns: 1fr; background-image: none !important; }
  .login-visual { display: none; }
  .login-panel { padding: 28px 20px; }
}
</style>
