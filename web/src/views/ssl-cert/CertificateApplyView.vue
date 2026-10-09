<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useRouter } from 'vue-router'
import { applyCertificate, listDNSProviders, type SSLDNSProvider, type SSLCertificateApplyForm } from '@/api/ssl-cert'

const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)
const providers = ref<SSLDNSProvider[]>([])
const form = reactive<SSLCertificateApplyForm>({
  name: '',
  domain: '',
  san_domains: '',
  acme_email: '',
  ca_provider: 'letsencrypt',
  key_algorithm: 'rsa2048',
  dns_provider_id: 0,
  auto_renew: true,
  renew_days_before: 30,
})
const rules: FormRules<SSLCertificateApplyForm> = {
  name: [{ required: true, message: '请输入证书名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入主域名', trigger: 'blur' }],
  acme_email: [
    { required: true, message: '请输入 ACME 邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效邮箱', trigger: 'blur' },
  ],
  dns_provider_id: [{ required: true, type: 'number', min: 1, message: '请选择 DNS 配置', trigger: 'change' }],
}

async function loadProviders() {
  const { data } = await listDNSProviders({ keyword: '', provider: '', page: 1, limit: 1000 })
  providers.value = data.data.list.filter(item => item.enabled)
}

async function submit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    await applyCertificate({ ...form })
    ElMessage.success('证书申请任务已创建，可在任务记录中查看进度')
    await router.push('/ssl-cert/tasks')
  } finally {
    submitting.value = false
  }
}

onMounted(loadProviders)
</script>

<template>
  <section class="apply-page">
    <header>
      <h2>证书申请</h2>
      <p>通过 ACME DNS-01 自动申请 SSL 证书，支持通配符域名。</p>
    </header>
    <div class="apply-card">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="apply-form">
        <div class="form-grid">
          <el-form-item label="证书名称" prop="name"><el-input v-model="form.name" placeholder="例如：生产环境通配符证书" /></el-form-item>
          <el-form-item label="主域名" prop="domain"><el-input v-model="form.domain" placeholder="例如：example.com 或 *.example.com" /></el-form-item>
          <el-form-item label="SAN 域名"><el-input v-model="form.san_domains" placeholder="多个域名使用逗号分隔" /></el-form-item>
          <el-form-item label="ACME 邮箱" prop="acme_email"><el-input v-model="form.acme_email" placeholder="用于接收证书到期提醒" /></el-form-item>
          <el-form-item label="DNS 配置" prop="dns_provider_id">
            <el-select v-model="form.dns_provider_id" placeholder="请选择已启用的 DNS 配置" filterable>
              <el-option v-for="item in providers" :key="item.id" :label="`${item.name}（${item.provider}）`" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="证书环境">
            <el-select v-model="form.ca_provider"><el-option label="Let's Encrypt 正式环境" value="letsencrypt" /><el-option label="Let's Encrypt 测试环境" value="letsencrypt_staging" /></el-select>
          </el-form-item>
          <el-form-item label="密钥算法">
            <el-select v-model="form.key_algorithm"><el-option label="RSA 2048" value="rsa2048" /><el-option label="RSA 4096" value="rsa4096" /><el-option label="EC P-256" value="ec256" /><el-option label="EC P-384" value="ec384" /></el-select>
          </el-form-item>
          <el-form-item label="提前续期"><el-input-number v-model="form.renew_days_before" :min="1" :max="90" /><span class="unit">天</span></el-form-item>
          <el-form-item label="自动续期"><el-switch v-model="form.auto_renew" /></el-form-item>
        </div>
        <div class="form-actions"><el-button type="primary" :loading="submitting" @click="submit">提交申请</el-button></div>
      </el-form>
    </div>
  </section>
</template>

<style scoped>
.apply-page{height:100%;min-width:0;padding:0 4px;box-sizing:border-box;font-size:11px}.apply-page header h2{margin:0;font-size:20px}.apply-page header p{margin:3px 0 10px;color:var(--text-secondary)}.apply-card{max-width:960px;padding:18px 20px;border:1px solid var(--line);border-radius:9px;background:#fff}.form-grid{display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr);gap:2px 16px}.apply-form :deep(.el-select){width:100%}.unit{margin-left:5px;color:var(--text-secondary)}.form-actions{display:flex;justify-content:flex-end;padding-top:8px;border-top:1px solid var(--line)}@media(max-width:760px){.form-grid{grid-template-columns:1fr}.apply-card{padding:14px 12px}}
</style>
