<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
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
  session_timeout_minutes: 30, captcha_enabled: false, captcha_expire_minutes: 1,
  login_verification_enabled: false, login_verification_method: '' as '' | 'email' | 'webhook' | 'passkey' | 'totp',
  max_login_failures: 5, account_lock_minutes: 30,
  verification_email_enabled: true, verification_email_expire_minutes: 5,
  verification_email_sender_name: '', verification_smtp_host: '', verification_smtp_port: 465,
  verification_smtp_username: '', verification_smtp_password: '', verification_smtp_tls: true,
  verification_webhook_enabled: false, verification_webhook_url: '', verification_webhook_secret: '',
  verification_passkey_enabled: true, verification_totp_enabled: true,
  direct_verification_email_enabled: true,
  direct_verification_passkey_enabled: true, direct_verification_totp_enabled: true,
})
const systemConfig = useSystemConfigStore()

async function load() {
  const { data } = await getSystemConfig()
  Object.assign(form, data.data)
  systemConfig.updateConfig(data.data.system_name, data.data.system_description)
}

async function save() {
  if (!form.verification_email_enabled && !form.verification_webhook_enabled && !form.verification_passkey_enabled && !form.verification_totp_enabled) {
    ElMessage.warning('验证配置至少需要启用一种方式')
    return
  }
  if (form.login_verification_enabled && !form.login_verification_method) {
    ElMessage.warning('开启登录二次验证后，请选择验证方式')
    return
  }
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

watch(
  () => form.login_verification_enabled,
  (enabled) => {
    if (enabled && !form.verification_email_enabled && !form.verification_webhook_enabled && !form.verification_passkey_enabled && !form.verification_totp_enabled) {
      ElMessage.warning('请先在“验证配置”中启用至少一种验证方式')
    }
  },
)

watch(
  () => [form.verification_email_enabled, form.verification_webhook_enabled, form.verification_passkey_enabled, form.verification_totp_enabled] as const,
  ([emailEnabled, webhookEnabled, passkeyEnabled, totpEnabled]) => {
    if (form.login_verification_method === 'email' && !emailEnabled) form.login_verification_method = ''
    if (form.login_verification_method === 'webhook' && !webhookEnabled) form.login_verification_method = ''
    if (form.login_verification_method === 'passkey' && !passkeyEnabled) form.login_verification_method = ''
    if (form.login_verification_method === 'totp' && !totpEnabled) form.login_verification_method = ''
  },
)

watch(
  () => form.verification_email_enabled,
  (enabled) => { if (!enabled) form.direct_verification_email_enabled = false },
)

watch(
  () => form.verification_passkey_enabled,
  (enabled) => { if (!enabled) form.direct_verification_passkey_enabled = false },
)

watch(
  () => form.verification_totp_enabled,
  (enabled) => { if (!enabled) form.direct_verification_totp_enabled = false },
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
            <el-form-item label="系统名称" required><el-input v-model="form.system_name" maxlength="128"
                show-word-limit /></el-form-item>
            <el-form-item label="系统描述"><el-input v-model="form.system_description" type="textarea" :rows="4"
                maxlength="500" show-word-limit /></el-form-item>
            <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving"
                @click="save">保存配置</el-button></el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="安全配置" name="security">
          <el-form :model="form" label-width="96px" class="system-form security-form">
            <section class="security-group">
              <h3>登录安全</h3>
              <div class="security-grid">
                <el-form-item label="Session 超时" required>
                  <el-input-number v-model="form.session_timeout_minutes" :min="1" :max="10080" controls-position="right" />
                  <span class="form-hint">分钟，最长 7 天</span>
                </el-form-item>
                <el-form-item label="登录失败上限" required>
                  <el-input-number v-model="form.max_login_failures" :min="1" :max="100" controls-position="right" />
                  <span class="form-hint">达到上限后锁定账户</span>
                </el-form-item>
                <el-form-item label="账户锁定时长" required>
                  <el-input-number v-model="form.account_lock_minutes" :min="1" :max="10080" controls-position="right" />
                  <span class="form-hint">分钟</span>
                </el-form-item>
              </div>
            </section>
            <section class="security-group">
              <h3>图形验证码</h3>
              <div class="security-grid">
                <el-form-item label="启用验证码">
                  <el-switch v-model="form.captcha_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">登录时校验图片验证码</span>
                </el-form-item>
                <el-form-item label="有效期" required>
                  <el-input-number v-model="form.captcha_expire_minutes" :min="1" :max="60"
                    :disabled="!form.captcha_enabled" controls-position="right" />
                  <span class="form-hint">分钟，默认 1 分钟</span>
                </el-form-item>
              </div>
            </section>
            <section class="security-group">
              <h3>登录二次验证</h3>
              <div class="security-grid">
                <el-form-item label="启用验证">
                  <el-switch v-model="form.login_verification_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">登录时通过已配置的方式发送验证码</span>
                </el-form-item>
                <el-form-item label="验证方式" :required="form.login_verification_enabled">
                  <el-select v-model="form.login_verification_method" :disabled="!form.login_verification_enabled"
                    placeholder="请选择验证码方式">
                    <el-option v-if="form.verification_email_enabled" label="邮件" value="email" />
                    <el-option v-if="form.verification_webhook_enabled" label="Webhook" value="webhook" />
                    <el-option v-if="form.verification_passkey_enabled" label="Passkey" value="passkey" />
                    <el-option v-if="form.verification_totp_enabled" label="TOTP" value="totp" />
                  </el-select>
                  <span v-if="!form.verification_email_enabled && !form.verification_webhook_enabled && !form.verification_passkey_enabled && !form.verification_totp_enabled" class="form-hint">
                    请先在“验证配置”中启用至少一种方式
                  </span>
                  <span v-else class="form-hint">仅显示已在“验证配置”中启用的方式</span>
                </el-form-item>
              </div>
            </section>
            <section class="security-group">
              <h3>直接登录配置</h3>
              <div class="security-grid">
                <el-form-item label="邮件">
                  <el-switch v-model="form.direct_verification_email_enabled"
                    :disabled="!form.verification_email_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">控制登录页邮件验证码直接登录入口</span>
                </el-form-item>
                <el-form-item label="Passkey">
                  <el-switch v-model="form.direct_verification_passkey_enabled"
                    :disabled="!form.verification_passkey_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">控制登录页 Passkey 直接登录入口</span>
                </el-form-item>
                <el-form-item label="TOTP">
                  <el-switch v-model="form.direct_verification_totp_enabled"
                    :disabled="!form.verification_totp_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">控制登录页 TOTP 直接登录入口</span>
                </el-form-item>
              </div>
            </section>
            <section class="security-group">
              <h3>密码策略</h3>
              <div class="security-grid">
                <el-form-item label="最小长度" required>
                  <el-input-number v-model="form.password_min_length" :min="1" :max="128" controls-position="right" />
                  <span class="form-hint">密码最少字符数</span>
                </el-form-item>
                <el-form-item label="安全级别">
                  <el-checkbox v-model="form.password_require_digit">数字</el-checkbox>
                  <el-checkbox v-model="form.password_require_letter">字母</el-checkbox>
                  <el-checkbox v-model="form.password_require_upper">大写</el-checkbox>
                  <el-checkbox v-model="form.password_require_lower">小写</el-checkbox>
                  <el-checkbox v-model="form.password_require_special">特殊字符</el-checkbox>
                </el-form-item>
              </div>
            </section>
            <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary" :loading="saving"
                @click="save">保存配置</el-button></el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="验证配置" name="verification">
          <el-tabs v-model="verificationTab" class="verification-tabs">
            <el-tab-pane label="邮件" name="email">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用邮件验证码"><el-switch v-model="form.verification_email_enabled" active-text="开启"
                    inactive-text="关闭" /></el-form-item>
                <el-form-item label="验证码有效期" required>
                  <el-input-number v-model="form.verification_email_expire_minutes" :min="1" :max="60"
                    controls-position="right" />
                  <span class="form-hint">分钟，默认 5 分钟</span>
                </el-form-item>
                <el-form-item label="发件人名称"><el-input v-model="form.verification_email_sender_name"
                    placeholder="系统管理员" /></el-form-item>
                <el-form-item label="SMTP 服务器">
                  <div class="smtp-row">
                    <el-input v-model="form.verification_smtp_host" placeholder="smtp.example.com" />
                    <el-input-number v-model="form.verification_smtp_port" :min="1" :max="65535"
                      controls-position="right" />
                  </div>
                </el-form-item>
                <el-form-item label="SMTP 账户">
                  <div class="mail-account-row">
                    <el-input v-model="form.verification_smtp_username" placeholder="SMTP 用户名"
                      autocomplete="username" />
                    <el-input v-model="form.verification_smtp_password" placeholder="SMTP 密码" type="password"
                      show-password autocomplete="new-password" />
                  </div>
                </el-form-item>
                <el-form-item label="SMTP TLS"><el-switch v-model="form.verification_smtp_tls" active-text="开启"
                    inactive-text="关闭" /></el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary"
                    :loading="saving" @click="save">保存配置</el-button></el-form-item>
                <el-form-item label="测试收件人" class="test-email-item">
                  <div class="email-test-row">
                    <el-input v-model="testRecipient" placeholder="test@example.com" />
                    <el-button v-permission="'POST /api/v1/system-config/update'" :loading="testingEmail"
                      @click="testEmail">发送</el-button>
                  </div>
                </el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="Webhook" name="webhook">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用 Webhook"><el-switch v-model="form.verification_webhook_enabled" active-text="开启"
                    inactive-text="关闭" /></el-form-item>
                <el-form-item label="Webhook 地址"><el-input v-model="form.verification_webhook_url"
                    placeholder="https://example.com/webhook" /></el-form-item>
                <el-form-item label="签名密钥"><el-input v-model="form.verification_webhook_secret" type="password"
                    show-password autocomplete="new-password" /></el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary"
                    :loading="saving" @click="save">保存配置</el-button></el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="Passkey" name="passkey">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用 Passkey">
                  <el-switch v-model="form.verification_passkey_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">控制 Passkey 验证入口与后端验证能力</span>
                </el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary"
                    :loading="saving" @click="save">保存配置</el-button></el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="TOTP" name="totp">
              <el-form :model="form" label-width="96px" class="system-form">
                <el-form-item label="启用 TOTP">
                  <el-switch v-model="form.verification_totp_enabled" active-text="开启" inactive-text="关闭" />
                  <span class="form-hint">控制 TOTP 验证入口与后端验证能力</span>
                </el-form-item>
                <el-form-item><el-button v-permission="'POST /api/v1/system-config/update'" type="primary"
                    :loading="saving" @click="save">保存配置</el-button></el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </section>
  </div>
</template>

<style scoped>
.system-config-page {
  height: 100%;
  min-height: 100%;
}

.section-card {
  min-height: 100%;
  padding: 20px 22px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-card);
  box-shadow: var(--shadow);
}

.page-toolbar {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--line);
}

.page-title {
  margin: 0;
  font-size: 18px;
  line-height: 1.35;
}

.page-subtitle {
  margin: 5px 0 0;
  color: var(--text-soft);
  font-size: 11px;
  line-height: 1.45;
}

.system-details {
  max-width: 680px;
  font-size: 11px;
}

.system-form {
  max-width: 760px;
  padding: 8px 18px 22px 10px;
}

.security-form :deep(.el-checkbox) {
  margin-right: 18px;
  font-size: 11px;
}

.security-form {
  max-width: 1080px;
}

.security-group {
  margin-bottom: 18px;
  padding: 16px 16px 0;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-card);
}

.security-group h3 {
  margin: 0 0 14px;
  color: var(--text);
  font-size: 12px;
  font-weight: 600;
}

.security-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
}

.security-grid :deep(.el-form-item) {
  min-width: 0;
}

.security-grid :deep(.el-form-item__content) {
  min-width: 0;
}

.security-grid :deep(.el-select),
.security-grid :deep(.el-input-number) {
  width: 150px;
}

.form-hint {
  margin-left: 12px;
  color: var(--text-soft);
  font-size: 11px;
}

.verification-tabs :deep(.el-tabs__header) {
  margin-bottom: 18px;
}

.verification-tabs :deep(.el-tabs__item) {
  font-size: 11px;
}

.system-tabs :deep(.el-tabs__header) {
  margin-right: 24px;
}

.system-tabs :deep(.el-tabs__item) {
  height: 42px;
  padding: 0 18px;
  font-size: 12px;
  line-height: 42px;
}

.system-tabs :deep(.el-tabs__content) {
  min-width: 0;
  padding: 4px 8px 20px;
}

.system-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.system-form :deep(.el-form-item__label) {
  height: 36px;
  padding-right: 14px;
  font-size: 11px;
  line-height: 36px;
}

.system-form :deep(.el-input__wrapper) {
  min-height: 36px;
  padding: 1px 11px;
}

.system-form :deep(.el-input__inner) {
  font-size: 11px;
}

.system-form :deep(.el-input-number) {
  height: 36px;
}

.system-form :deep(.el-input-number .el-input__wrapper) {
  min-height: 36px;
}

.system-form :deep(.el-switch) {
  --el-switch-on-color: var(--primary);
  transform: scale(.94);
  transform-origin: left center;
}

.system-form :deep(.el-button) {
  min-height: 34px;
  padding: 7px 16px;
  font-size: 11px;
}

.smtp-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 118px;
  gap: 12px;
  width: 100%;
}

.smtp-row :deep(.el-input-number) {
  width: 118px;
}

.mail-account-row {
  display: grid;
  grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
  gap: 12px;
  width: 100%;
}

.email-test-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  width: 100%;
}

.email-test-row :deep(.el-button) {
  margin: 0;
  white-space: nowrap;
}

.test-email-item {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px dashed var(--line);
}

@media (max-width: 980px) {
  .security-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .section-card {
    padding: 16px;
  }

  .page-toolbar {
    margin-bottom: 14px;
  }

  .system-tabs :deep(.el-tabs__header) {
    margin-right: 14px;
  }

  .system-tabs :deep(.el-tabs__item) {
    padding: 0 12px;
  }

  .system-tabs :deep(.el-tabs__content),
  .system-form {
    padding-right: 0;
  }

  .smtp-row,
  .mail-account-row,
  .email-test-row,
  .security-grid {
    grid-template-columns: 1fr;
  }

  .security-group {
    padding-right: 10px;
    padding-left: 10px;
  }

  .smtp-row :deep(.el-input-number) {
    width: 100%;
  }
}
</style>
