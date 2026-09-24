<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Lock, Message, Refresh, UserFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import illustrationUrl from '@/assets/login/illustration.svg'
import backgroundUrl from '@/assets/login/bg.png'
import gopsLogoUrl from '@/assets/gops.svg'
import { useAuthStore } from '@/stores/auth'
import { useSystemConfigStore } from '@/stores/system-config'
import { beginDirectPasskeyLogin, beginPasskeyLogin, finishDirectPasskeyLogin, finishPasskeyLogin, getCaptcha, getLoginSecurityConfig, resetPassword, sendMFACode, sendPasswordResetCode, verifyMFA, verifyPasswordResetCode, type MFAMethod } from '@/api/auth'
import { credentialJSON, requestOptions } from '@/utils/webauthn'

const auth = useAuthStore()
const systemConfig = useSystemConfigStore()
const router = useRouter()
const route = useRoute()
const loading = ref(false)
const sendingEmailCode = ref(false)
const resendCountdown = ref(0)
let resendTimer: ReturnType<typeof setInterval> | undefined
const verificationCountdown = ref(0)
let verificationTimer: ReturnType<typeof setInterval> | undefined
const loginStep = ref<'credentials' | 'verification' | 'forgot' | 'reset_verify' | 'reset_password'>('credentials')
const security = reactive({
  captcha_enabled: false,
  login_verification_enabled: false,
  login_verification_method: '' as '' | 'email' | 'webhook',
})
const captcha = reactive({ id: '', image: '' })
const emailCodeRequired = ref(false)
const mfaToken = ref('')
const mfaMethods = ref<MFAMethod[]>([])
const selectedMFAMethod = ref<MFAMethod>('code')
const form = reactive({ username: 'admin', password: 'password', captcha_code: '', email_code: '' })
const resetForm = reactive({ code: '', new_password: '', confirm_password: '' })
const resetToken = ref('')
const stepTitle = computed(() => ({ credentials: '欢迎登录', verification: '二次验证', forgot: '忘记密码', reset_verify: '身份验证', reset_password: '重置密码' })[loginStep.value])
const stepDescription = computed(() => ({
  credentials: '登录管理控制台，开始你的工作。',
  verification: '请输入收到的验证码以继续登录。',
  forgot: '验证身份后即可重新设置密码。',
  reset_verify: '请输入收到的验证码。',
  reset_password: '身份验证已通过，请设置新密码。',
})[loginStep.value])

function startResendCountdown(seconds: number) {
  resendCountdown.value = seconds
  if (resendTimer) clearInterval(resendTimer)
  if (seconds <= 0) return
  resendTimer = setInterval(() => {
    resendCountdown.value -= 1
    if (resendCountdown.value <= 0 && resendTimer) {
      clearInterval(resendTimer)
      resendTimer = undefined
    }
  }, 1000)
}

function startVerificationCountdown(seconds: number) {
  verificationCountdown.value = seconds
  if (verificationTimer) clearInterval(verificationTimer)
  if (seconds <= 0) return
  verificationTimer = setInterval(() => {
    verificationCountdown.value -= 1
    if (verificationCountdown.value <= 0 && verificationTimer) {
      clearInterval(verificationTimer)
      verificationTimer = undefined
    }
  }, 1000)
}

function formatCountdown(seconds: number) {
  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return `${String(minutes).padStart(2, '0')}:${String(remainder).padStart(2, '0')}`
}

async function refreshCaptcha() {
  if (!security.captcha_enabled) return
  const { data } = await getCaptcha()
  captcha.id = data.data.captcha_id
  captcha.image = data.data.image
  form.captcha_code = ''
}

async function requestEmailCode() {
	if (!mfaToken.value) return false
	sendingEmailCode.value = true
	try {
		const { data } = await sendMFACode(mfaToken.value)
    emailCodeRequired.value = data.data.required
    if (data.data.required) loginStep.value = 'verification'
    startResendCountdown(data.data.retry_after_seconds)
    startVerificationCountdown(data.data.expires_in_seconds)
    if (data.data.sent) ElMessage.success(security.login_verification_method === 'webhook' ? '验证码已通过 Webhook 发送' : '验证码已发送至用户邮箱')
    else if (data.data.required && data.data.retry_after_seconds > 0) ElMessage.info('验证码已发送，请勿重复获取')
    return data.data.required
	} finally {
    sendingEmailCode.value = false
  }
}

async function submit() {
  if (loginStep.value === 'credentials' && security.captcha_enabled && !form.captcha_code.trim()) return ElMessage.warning('请输入图形验证码')
	loading.value = true
	try {
		if (loginStep.value === 'credentials') {
			const result = await auth.login({ username: form.username, password: form.password, captcha_id: captcha.id, captcha_code: form.captcha_code })
			if (result.mfa_required) {
				mfaToken.value = result.mfa_token
				mfaMethods.value = result.mfa_methods
				selectedMFAMethod.value = result.mfa_methods[0]
				loginStep.value = 'verification'
				if (selectedMFAMethod.value === 'code') await requestEmailCode()
				return
			}
		} else if (selectedMFAMethod.value === 'passkey') {
			await loginWithPasskey()
			return
		} else {
			if (selectedMFAMethod.value === 'code' && verificationCountdown.value <= 0) return ElMessage.warning('验证码已过期，请重新获取')
			if (!form.email_code.trim()) return ElMessage.warning('请输入验证码')
			const { data } = await verifyMFA({ mfa_token: mfaToken.value, method: selectedMFAMethod.value, code: form.email_code.trim() })
			await auth.acceptLogin(data.data)
		}
		await router.replace((route.query.redirect as string) || '/')
  } catch (error) {
    if (loginStep.value === 'credentials') await refreshCaptcha()
    else form.email_code = ''
    throw error
  } finally {
    loading.value = false
  }
}

async function loginWithPasskey() {
	const { data } = await beginPasskeyLogin(mfaToken.value)
	const credential = await navigator.credentials.get(requestOptions(data.data.options)) as PublicKeyCredential | null
	if (!credential) throw new Error('Passkey 验证已取消')
	const result = await finishPasskeyLogin({ mfa_token: mfaToken.value, credential: credentialJSON(credential) })
	await auth.acceptLogin(result.data.data)
	await router.replace((route.query.redirect as string) || '/')
}

async function directPasskeyLogin() {
	if (!form.username.trim()) return ElMessage.warning('请输入用户名或邮箱')
	if (security.captcha_enabled && !form.captcha_code.trim()) return ElMessage.warning('请输入图形验证码')
	if (!window.PublicKeyCredential) return ElMessage.warning('当前浏览器不支持 Passkey')
	loading.value = true
	try {
		const { data } = await beginDirectPasskeyLogin({ username: form.username.trim(), captcha_id: captcha.id, captcha_code: form.captcha_code })
		const credential = await navigator.credentials.get(requestOptions(data.data.options)) as PublicKeyCredential | null
		if (!credential) return
		const result = await finishDirectPasskeyLogin({ login_token: data.data.login_token, credential: credentialJSON(credential) })
		await auth.acceptLogin(result.data.data)
		await router.replace((route.query.redirect as string) || '/')
	} catch (error) {
		await refreshCaptcha()
		throw error
	} finally {
		loading.value = false
	}
}

async function selectMFAMethod(method: MFAMethod) {
	selectedMFAMethod.value = method
	form.email_code = ''
	if (method === 'code' && verificationCountdown.value <= 0) await requestEmailCode()
}

function backToCredentials() {
  loginStep.value = 'credentials'
	emailCodeRequired.value = false
	mfaToken.value = ''
	mfaMethods.value = []
  form.email_code = ''
  resetToken.value = ''
  startResendCountdown(0)
  startVerificationCountdown(0)
}

async function openForgotPassword() {
  backToCredentials()
  form.password = ''
  loginStep.value = 'forgot'
  await refreshCaptcha()
}

async function requestPasswordResetCode() {
  if (!form.username.trim()) return ElMessage.warning('请输入用户名')
  if (security.captcha_enabled && !form.captcha_code.trim()) return ElMessage.warning('请输入图形验证码')
  sendingEmailCode.value = true
  try {
    const { data } = await sendPasswordResetCode({ username: form.username.trim(), captcha_id: captcha.id, captcha_code: form.captcha_code })
    startResendCountdown(data.data.retry_after_seconds)
    startVerificationCountdown(data.data.expires_in_seconds)
    loginStep.value = 'reset_verify'
    if (data.data.sent) ElMessage.success('密码重置验证码已发送')
    else ElMessage.info('验证码已发送，请勿重复获取')
  } catch (error) {
    await refreshCaptcha()
    throw error
  } finally {
    sendingEmailCode.value = false
  }
}

async function verifyPasswordReset() {
  if (verificationCountdown.value <= 0) return ElMessage.warning('验证码已过期，请重新获取')
  if (!resetForm.code.trim()) return ElMessage.warning('请输入验证码')
  loading.value = true
  try {
    const { data } = await verifyPasswordResetCode({ username: form.username.trim(), code: resetForm.code.trim() })
    resetToken.value = data.data.reset_token
    startVerificationCountdown(0)
    loginStep.value = 'reset_password'
  } finally {
    loading.value = false
  }
}

async function confirmPasswordReset() {
  if (!resetForm.new_password) return ElMessage.warning('请输入新密码')
  if (resetForm.new_password !== resetForm.confirm_password) return ElMessage.warning('两次输入的新密码不一致')
  loading.value = true
  try {
    await resetPassword({ reset_token: resetToken.value, new_password: resetForm.new_password })
    ElMessage.success('密码修改成功，请重新登录')
    resetForm.code = ''
    resetForm.new_password = ''
    resetForm.confirm_password = ''
    resetToken.value = ''
    backToCredentials()
    await refreshCaptcha()
  } finally {
    loading.value = false
  }
}

function handlePrimaryAction() {
  if (loginStep.value === 'forgot') return requestPasswordResetCode()
  if (loginStep.value === 'reset_verify') return verifyPasswordReset()
  if (loginStep.value === 'reset_password') return confirmPasswordReset()
  return submit()
}

onMounted(async () => {
  const { data } = await getLoginSecurityConfig()
  Object.assign(security, data.data)
  await refreshCaptcha()
})

onBeforeUnmount(() => {
  if (resendTimer) clearInterval(resendTimer)
  if (verificationTimer) clearInterval(verificationTimer)
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
            <h1>{{ stepTitle }}</h1>
            <p>{{ stepDescription }}</p>
          </div>
        </div>
        <el-form :model="form" class="login-form" @keyup.enter="handlePrimaryAction">
          <el-form-item v-if="loginStep === 'credentials' || loginStep === 'forgot'">
            <el-input v-model="form.username" size="large" :placeholder="loginStep === 'forgot' ? '请输入用户名' : '请输入用户名或邮箱'" autocomplete="username">
              <template #prefix><el-icon>
                  <UserFilled />
                </el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item v-if="loginStep === 'credentials'">
            <el-input v-model="form.password" type="password" show-password size="large" placeholder="请输入密码"
              autocomplete="current-password">
              <template #prefix><el-icon>
                  <Lock />
                </el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item v-if="(loginStep === 'credentials' || loginStep === 'forgot') && security.captcha_enabled">
            <div class="verification-row captcha-row">
              <el-input v-model="form.captcha_code" size="large" maxlength="4" placeholder="请输入图形验证码">
                <template #prefix><el-icon><Refresh /></el-icon></template>
              </el-input>
              <button class="captcha-image" type="button" title="点击刷新验证码" @click="refreshCaptcha"><img :src="captcha.image" alt="图形验证码" /></button>
            </div>
          </el-form-item>
			<el-form-item v-if="loginStep === 'verification'">
				<div class="mfa-methods">
					<el-button v-for="method in mfaMethods" :key="method" :type="selectedMFAMethod === method ? 'primary' : 'default'" plain @click="selectMFAMethod(method)">
						{{ method === 'code' ? '验证码' : method === 'totp' ? 'TOTP' : 'Passkey' }}
					</el-button>
				</div>
			</el-form-item>
			<el-form-item v-if="loginStep === 'verification' && selectedMFAMethod !== 'passkey'">
				<div :class="selectedMFAMethod === 'code' ? 'verification-row' : 'totp-row'">
					<el-input v-model="form.email_code" size="large" maxlength="6" :placeholder="selectedMFAMethod === 'totp' ? '请输入身份验证器动态码' : '请输入登录验证码'">
						<template #prefix><el-icon><Message /></el-icon></template>
					</el-input>
					<el-button v-if="selectedMFAMethod === 'code'" :loading="sendingEmailCode" :disabled="resendCountdown > 0" @click="requestEmailCode">
                重新发送
              </el-button>
            </div>
          </el-form-item>
          <el-form-item v-if="loginStep === 'reset_verify'">
            <el-input v-model="resetForm.code" size="large" maxlength="6" placeholder="请输入验证码">
              <template #prefix><el-icon><Message /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item v-if="loginStep === 'reset_password'">
            <el-input v-model="resetForm.new_password" type="password" show-password size="large" placeholder="请输入新密码"
              autocomplete="new-password"><template #prefix><el-icon><Lock /></el-icon></template></el-input>
          </el-form-item>
          <el-form-item v-if="loginStep === 'reset_password'">
            <el-input v-model="resetForm.confirm_password" type="password" show-password size="large" placeholder="请再次输入新密码"
              autocomplete="new-password"><template #prefix><el-icon><Lock /></el-icon></template></el-input>
          </el-form-item>
			<p v-if="(loginStep === 'verification' && selectedMFAMethod === 'code') || loginStep === 'reset_verify'" class="verification-expiry" :class="{ expired: verificationCountdown <= 0 }">
            {{ verificationCountdown > 0 ? `验证码有效期剩余 ${formatCountdown(verificationCountdown)}` : '验证码已过期，请重新获取' }}
          </p>
          
          <button v-if="loginStep === 'credentials'" type="button" class="forgot-password" @click="openForgotPassword">忘记密码？</button>
          <el-button type="primary" size="large" :loading="loading || sendingEmailCode" class="submit" @click="handlePrimaryAction">
            {{ loginStep === 'verification' ? '验证并登录' : loginStep === 'forgot' ? '发送验证码' : loginStep === 'reset_verify' ? '验证身份' : loginStep === 'reset_password' ? '修改密码' : '登录' }}
          </el-button>
          <el-button v-if="loginStep === 'credentials'" class="passkey-login-button" plain :disabled="loading" @click="directPasskeyLogin">
            使用 Passkey 直接登录
          </el-button>
          <el-button v-if="loginStep !== 'credentials'" class="back-button" text @click="backToCredentials">
            返回账号登录
          </el-button>
        </el-form>
        <p class="login-tip">
          {{ loginStep === 'verification' || loginStep === 'reset_verify' ? '验证码最多可尝试 5 次' : '使用您的账号和密码登录系统' }}
        </p>
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

.totp-row { width: 100%; }

.mfa-methods {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: 8px;
	width: 100%;
}

.mfa-methods :deep(.el-button) { margin: 0; }

.verification-row :deep(.el-button) {
  height: 42px;
  margin: 0;
  border-radius: 8px;
}

.verification-expiry {
  margin: -6px 0 12px;
  color: var(--text-soft);
  font-size: 11px;
  text-align: center;
}

.verification-expiry.expired {
  color: var(--danger);
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

.passkey-login-button {
	width: 100%;
	height: 40px;
	margin: 9px 0 0;
	border-color: var(--primary);
	color: var(--primary);
}

.forgot-password {
  display: block;
  margin: -4px 0 10px auto;
  padding: 0;
  border: 0;
  color: var(--primary);
  background: transparent;
  font-size: 11px;
  cursor: pointer;
}

.back-button {
  display: block;
  margin: 10px auto 0;
  color: var(--text-soft);
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
