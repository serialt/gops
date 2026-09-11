<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'
import { getSystemConfig, testSystemConfigEmail, updateSystemConfig } from '@/api/system-config'
import { useSystemConfigStore } from '@/stores/system-config'

const activeTab = ref('basic')
const verificationTab = ref('email')
const saving = ref(false)
const testingEmail = ref(false)
const testRecipient = ref('')
const route = useRoute()
const form = reactive({
  system_name: '', system_description: '', password_min_length: 8,
  password_require_digit: false, password_require_letter: false, password_require_upper: false,
  password_require_lower: false, password_require_special: false,
  session_timeout_minutes: 30, captcha_enabled: false, max_login_failures: 5, account_lock_minutes: 30,
  verification_email_enabled: false, verification_email_sender_name: '', verification_smtp_host: '', verification_smtp_port: 465,
  verification_smtp_username: '', verification_smtp_password: '', verification_smtp_tls: true,
  verification_webhook_enabled: false, verification_webhook_url: '', verification_webhook_secret: '',
})
const systemConfig = useSystemConfigStore()
const verificationChannels = computed<string[]>({
  get: () => [
    ...(form.verification_email_enabled ? ['email'] : []),
    ...(form.verification_webhook_enabled ? ['webhook'] : []),
  ],
  set: (channels) => {
    form.verification_email_enabled = channels.includes('email')
    form.verification_webhook_enabled = channels.includes('webhook')
  },
})

async function load() {
  const { data } = await getSystemConfig()
  Object.assign(form, data.data)
  systemConfig.updateConfig(data.data.system_name, data.data.system_description)
}

async function save() {
  saving.value = true
  try {
    const { data } = await updateSystemConfig(form)
    Object.assign(form, data.data)
    systemConfig.updateConfig(data.data.system_name, data.data.system_description)
    ElMessage.success('保存成功')
  } finally {
    saving.value = false
  }
}

async function testEmail() {
  if (!testRecipient.value.trim()) {
    ElMessage.warning('请输入测试收件人地址')
    return
  }
  testingEmail.value = true
  try {
    await testSystemConfigEmail({ recipient: testRecipient.value.trim() })
    ElMessage.success('测试邮件已发送')
  } finally {
    testingEmail.value = false
  }
}
watch(
  () => route.path,
  (path) => {
    if (path === '/system-config') void load()
  },
  { immediate: true },
)
</script>

<template>
  <div class="page-container system-config-page">
    <section class="section-card">
      <div class="page-toolbar">
        <div>
          <h2 class="page-title">系统配置</h2>
          <p class="page-subtitle">查看系统基础信息与运行环境。</p>
        </div>
      </div>
      <el-tabs v-model="activeTab" class="system-tabs" tab-position="left">
        <el-tab-pane label="基础配置" name="basic">
          <el-form :model="form" label-width="82px" class="system-form">
            <el-form-item label="系统名称" required><el-input v-model="form.system_name" maxlength="128" show-word-limit /></el-form-item>
            <el-form-item label="系统描述"><el-input v-model="form.system_description" type="textarea" :rows="4" maxlength="500" show-word-limit /></el-form-item>
            <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving" @click="save">保存配置</el-button></el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="安全配置" name="security">
          <el-form :model="form" label-width="96px" class="system-form security-form">
            <el-form-item label="Session 超时" required>
              <el-input-number v-model="form.session_timeout_minutes" :min="1" :max="10080" controls-position="right" />
              <span class="form-hint">分钟，最长 7 天</span>
            </el-form-item>
            <el-form-item label="验证码验证">
              <el-switch v-model="form.captcha_enabled" active-text="开启" inactive-text="关闭" />
            </el-form-item>
            <el-form-item v-if="form.captcha_enabled" label="验证码方式">
              <el-checkbox-group v-model="verificationChannels">
                <el-checkbox label="email">邮件验证码</el-checkbox>
                <el-checkbox label="webhook">Webhook 验证码</el-checkbox>
              </el-checkbox-group>
              <span class="form-hint">请至少选择一种已启用的验证码方式</span>
            </el-form-item>
            <el-form-item label="登录失败上限" required>
              <el-input-number v-model="form.max_login_failures" :min="1" :max="100" controls-position="right" />
              <span class="form-hint">达到上限后锁定账户</span>
            </el-form-item>
            <el-form-item label="账户锁定时长" required>
              <el-input-number v-model="form.account_lock_minutes" :min="1" :max="10080" controls-position="right" />
              <span class="form-hint">分钟</span>
            </el-form-item>
            <el-form-item label="密码最小长度" required>
              <el-input-number v-model="form.password_min_length" :min="1" :max="128" controls-position="right" />
              <span class="form-hint">设置用户密码的最少字符数</span>
            </el-form-item>
            <el-form-item label="密码安全级别">
              <el-checkbox v-model="form.password_require_digit">数字</el-checkbox>
              <el-checkbox v-model="form.password_require_letter">字母</el-checkbox>
              <el-checkbox v-model="form.password_require_upper">大写字母</el-checkbox>
              <el-checkbox v-model="form.password_require_lower">小写字母</el-checkbox>
              <el-checkbox v-model="form.password_require_special">特殊字符</el-checkbox>
            </el-form-item>
            <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving" @click="save">保存配置</el-button></el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="验证码配置" name="verification">
          <el-tabs v-model="verificationTab" class="verification-tabs">
            <el-tab-pane label="邮件" name="email">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用邮件验证码"><el-switch v-model="form.verification_email_enabled" active-text="开启" inactive-text="关闭" /></el-form-item>
                <el-form-item label="发件人名称"><el-input v-model="form.verification_email_sender_name" placeholder="系统管理员" /></el-form-item>
                <el-form-item label="SMTP 服务器">
                  <div class="smtp-row">
                    <el-input v-model="form.verification_smtp_host" placeholder="smtp.example.com" />
                    <el-input-number v-model="form.verification_smtp_port" :min="1" :max="65535" controls-position="right" />
                  </div>
                </el-form-item>
                <el-form-item label="SMTP 账户">
                  <div class="mail-account-row">
                    <el-input v-model="form.verification_smtp_username" placeholder="SMTP 用户名" autocomplete="username" />
                    <el-input v-model="form.verification_smtp_password" placeholder="SMTP 密码" type="password" show-password autocomplete="new-password" />
                  </div>
                </el-form-item>
                <el-form-item label="SMTP TLS"><el-switch v-model="form.verification_smtp_tls" active-text="开启" inactive-text="关闭" /></el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving" @click="save">保存配置</el-button></el-form-item>
                <el-form-item label="测试收件人" class="test-email-item">
                  <div class="email-test-row">
                    <el-input v-model="testRecipient" placeholder="test@example.com" />
                    <el-button v-permission="'POST /api/v1/system-config/update'" :loading="testingEmail" @click="testEmail">发送</el-button>
                  </div>
                </el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="Webhook" name="webhook">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用 Webhook"><el-switch v-model="form.verification_webhook_enabled" active-text="开启" inactive-text="关闭" /></el-form-item>
                <el-form-item label="Webhook 地址"><el-input v-model="form.verification_webhook_url" placeholder="https://example.com/webhook" /></el-form-item>
                <el-form-item label="签名密钥"><el-input v-model="form.verification_webhook_secret" type="password" show-password autocomplete="new-password" /></el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving" @click="save">保存配置</el-button></el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </section>
  </div>
</template>

<style scoped>
.system-config-page { height: 100%; min-height: 100%; }
.section-card { min-height: 100%; padding: 12px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg-card); box-shadow: var(--shadow); }
.page-toolbar { margin-bottom: 10px; }
.page-title { margin: 0; font-size: 18px; line-height: 1.35; }
.page-subtitle { margin: 2px 0 0; color: var(--text-soft); font-size: 11px; line-height: 1.45; }
.system-details { max-width: 680px; font-size: 11px; }
.system-form { max-width: 680px; }
.security-form :deep(.el-checkbox) { margin-right: 14px; font-size: 11px; }
.form-hint { margin-left: 8px; color: var(--text-soft); font-size: 11px; }
.verification-tabs :deep(.el-tabs__header) { margin-bottom: 8px; }
.verification-tabs :deep(.el-tabs__item) { font-size: 11px; }
.system-tabs :deep(.el-tabs__header) { margin-right: 14px; }
.system-tabs :deep(.el-tabs__item) { height: 34px; padding: 0 12px; font-size: 12px; line-height: 34px; }
.system-form :deep(.el-form-item) { margin-bottom: 10px; }
.system-form :deep(.el-form-item__label) { height: 32px; font-size: 11px; line-height: 32px; }
.system-form :deep(.el-input__wrapper) { min-height: 32px; padding: 1px 9px; }
.system-form :deep(.el-input__inner) { font-size: 11px; }
.system-form :deep(.el-input-number) { height: 32px; }
.system-form :deep(.el-input-number .el-input__wrapper) { min-height: 32px; }
.system-form :deep(.el-switch) { --el-switch-on-color: var(--primary); transform: scale(.88); transform-origin: left center; }
.system-form :deep(.el-button) { min-height: 32px; padding: 6px 13px; font-size: 11px; }
.smtp-row { display: grid; grid-template-columns: minmax(0, 1fr) 118px; gap: 8px; width: 100%; }
.smtp-row :deep(.el-input-number) { width: 118px; }
.mail-account-row { display: grid; grid-template-columns: minmax(0, 3fr) minmax(0, 2fr); gap: 8px; width: 100%; }
.email-test-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 8px; width: 100%; }
.email-test-row :deep(.el-button) { margin: 0; white-space: nowrap; }
.test-email-item { margin-top: 64px; }
</style>
