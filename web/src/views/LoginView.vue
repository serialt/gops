<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Lock, Message, Refresh, UserFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import illustrationUrl from '@/assets/login/illustration.svg'
import backgroundUrl from '@/assets/login/bg.png'
import gopsLogoUrl from '@/assets/gops.svg'
import { useAuthStore } from '@/stores/auth'
import { useSystemConfigStore } from '@/stores/system-config'
import { getCaptcha, getLoginSecurityConfig, sendLoginEmailCode } from '@/api/auth'

const auth = useAuthStore()
const systemConfig = useSystemConfigStore()
const router = useRouter()
const route = useRoute()
const loading = ref(false)
const sendingEmailCode = ref(false)
const security = reactive({
  captcha_enabled: false,
  login_verification_enabled: false,
  login_verification_method: '' as '' | 'email' | 'webhook',
})
const captcha = reactive({ id: '', image: '' })
const emailCodeRequired = ref(false)
const form = reactive({ username: 'admin', password: 'password', captcha_code: '', email_code: '' })

async function refreshCaptcha() {
  if (!security.captcha_enabled) return
  const { data } = await getCaptcha()
  captcha.id = data.data.captcha_id
  captcha.image = data.data.image
  form.captcha_code = ''
}

async function requestEmailCode() {
  if (security.captcha_enabled && !form.captcha_code.trim()) {
    ElMessage.warning('请输入图形验证码')
    return false
  }
  sendingEmailCode.value = true
  try {
    const { data } = await sendLoginEmailCode({ username: form.username, password: form.password, captcha_id: captcha.id, captcha_code: form.captcha_code })
    emailCodeRequired.value = data.data.required
    if (data.data.sent) ElMessage.success(security.login_verification_method === 'webhook' ? '验证码已通过 Webhook 发送' : '验证码已发送至用户邮箱')
    return data.data.required
  } catch (error) {
    await refreshCaptcha()
    throw error
  } finally {
    sendingEmailCode.value = false
  }
}

async function submit() {
  if (security.captcha_enabled && !form.captcha_code.trim()) return ElMessage.warning('请输入图形验证码')
  loading.value = true
  try {
    if (security.login_verification_enabled && !emailCodeRequired.value) {
      const required = await requestEmailCode()
      if (required) return
    }
    if (emailCodeRequired.value && !form.email_code.trim()) return ElMessage.warning('请输入登录验证码')
    await auth.login({ username: form.username, password: form.password, captcha_id: captcha.id, captcha_code: form.captcha_code, email_code: form.email_code })
    await router.replace((route.query.redirect as string) || '/')
  } catch (error) {
    await refreshCaptcha()
    throw error
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const { data } = await getLoginSecurityConfig()
  Object.assign(security, data.data)
  await refreshCaptcha()
})
</script>

<template>
  <main class="login-page">
    <img class="login-background" :src="backgroundUrl" alt="" aria-hidden="true" />
    <section class="login-visual" aria-hidden="true">
      <img class="login-illustration" :src="illustrationUrl" alt="" />
    </section>
    <section class="login-panel">
      <div class="login-form-wrap">
        <div class="login-heading">
          <img class="login-logo" :src="gopsLogoUrl" :alt="systemConfig.systemName" />
          <div>
            <h1>欢迎登录</h1>
            <p>登录管理控制台，开始你的工作。</p>
          </div>
        </div>
        <el-form :model="form" class="login-form" @keyup.enter="submit">
          <el-form-item>
            <el-input v-model="form.username" size="large" placeholder="请输入用户名" autocomplete="username">
              <template #prefix><el-icon>
                  <UserFilled />
                </el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item v-if="security.captcha_enabled">
            <div class="verification-row captcha-row">
              <el-input v-model="form.captcha_code" size="large" maxlength="4" placeholder="请输入图形验证码">
                <template #prefix><el-icon><Refresh /></el-icon></template>
              </el-input>
              <button class="captcha-image" type="button" title="点击刷新验证码" @click="refreshCaptcha"><img :src="captcha.image" alt="图形验证码" /></button>
            </div>
          </el-form-item>
          <el-form-item v-if="emailCodeRequired">
            <div class="verification-row">
              <el-input v-model="form.email_code" size="large" maxlength="6" placeholder="请输入登录验证码">
                <template #prefix><el-icon><Message /></el-icon></template>
              </el-input>
              <el-button :loading="sendingEmailCode" @click="requestEmailCode">重新发送</el-button>
            </div>
          </el-form-item>
          <el-form-item>
            <el-input v-model="form.password" type="password" show-password size="large" placeholder="请输入密码"
              autocomplete="current-password">
              <template #prefix><el-icon>
                  <Lock />
                </el-icon></template>
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
.login-page {
  position: relative;
  display: grid;
  grid-template-columns: minmax(520px, 1.15fr) minmax(420px, .85fr);
  min-height: 100%;
  overflow: hidden;
  background-color: #fff;
}

.login-background {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
}

.login-visual {
  position: relative;
  z-index: 1;
  min-width: 0;
  overflow: hidden;
}

.login-logo {
  width: 42px;
  height: 42px;
  border-radius: 11px;
  object-fit: contain;
  box-shadow: 0 10px 24px rgba(48, 73, 180, .26);
}

.login-illustration {
  position: absolute;
  left: 10%;
  bottom: 15%;
  width: min(80%, 680px);
  max-height: 58%;
  object-fit: contain;
  object-position: left bottom;
}

.login-panel {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  padding: 40px;
  background: #fff;
}

.login-form-wrap {
  width: min(360px, 100%);
}

.login-heading {
  display: flex;
  align-items: center;
  gap: 13px;
  margin-bottom: 32px;
}

.login-heading h1 {
  margin: 0 0 4px;
  color: #182235;
  font-size: 24px;
  line-height: 1.15;
}

.login-heading p {
  margin: 0;
  color: #98a2b3;
  font-size: 12px;
}

.login-logo {
  width: 46px;
  height: 46px;
  border-radius: 12px;
  flex: 0 0 auto;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.login-form :deep(.el-input__wrapper) {
  min-height: 42px;
  border-radius: 8px;
  box-shadow: 0 0 0 1px #e3e8f1 inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--primary) inset;
}

.login-form :deep(.el-input__prefix) {
  color: #a6afbf;
}

.verification-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 112px;
  gap: 8px;
  width: 100%;
}

.verification-row :deep(.el-button) {
  height: 42px;
  margin: 0;
  border-radius: 8px;
}

.captcha-image {
  height: 42px;
  padding: 0;
  overflow: hidden;
  border: 1px solid #e3e8f1;
  border-radius: 8px;
  background: #f3f5ff;
  cursor: pointer;
}

.captcha-image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.submit {
  width: 100%;
  height: 42px;
  margin-top: 4px;
  border-radius: 8px;
  font-weight: 600;
}

.login-tip {
  margin: 18px 0 0;
  color: #a2aaba;
  font-size: 11px;
  text-align: center;
}

@media (max-width: 900px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .login-background {
    display: none;
  }

  .login-visual {
    display: none;
  }

  .login-panel {
    padding: 28px 20px;
  }
}
</style>
