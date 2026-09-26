<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteMonitorChannel, listMonitorChannels, saveMonitorChannel, testMonitorChannel, type MonitorAlertChannel, type MonitorAlertChannelConfig, type MonitorAlertChannelRequest, type MonitorAlertChannelType } from '@/api/monitor'

const labels: Record<MonitorAlertChannelType, string> = { email: '邮件', dingtalk: '钉钉', feishu: '飞书', wechat: '企业微信', webhook: 'Webhook' }
const emptyConfig = (): MonitorAlertChannelConfig => ({ smtp_host: '', smtp_port: 465, smtp_user: '', smtp_password: '', from_email: '', from_name: '', webhook_url: '', wechat_webhook: '', dingtalk_webhook: '', dingtalk_secret: '', feishu_webhook: '' })
const emptyForm = (): MonitorAlertChannelRequest => ({ name: '', channel_type: 'email', enabled: true, config: emptyConfig() })
const rows = ref<MonitorAlertChannel[]>([])
const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const dialogVisible = ref(false)
const saving = ref(false)
const testing = ref(false)
const testRecipient = ref('')
const editing = ref<MonitorAlertChannel | null>(null)
const formRef = ref<FormInstance>()
const form = ref<MonitorAlertChannelRequest>(emptyForm())
const rules: FormRules = { name: [{ required: true, message: '请输入通道名称', trigger: 'blur' }], channel_type: [{ required: true, message: '请选择通道类型', trigger: 'change' }] }

function createTag(text: string, className: string) {
  const tag = document.createElement('span')
  tag.className = `channel-tag ${className}`
  tag.textContent = text
  return tag
}
function configPreview(row: MonitorAlertChannel) {
  const c = row.config
  if (row.channel_type === 'email') return `${c.smtp_host || '-'}:${c.smtp_port || 465} · ${c.from_email || '-'}`
  const value = row.channel_type === 'webhook' ? c.webhook_url : row.channel_type === 'wechat' ? c.wechat_webhook : row.channel_type === 'dingtalk' ? c.dingtalk_webhook : c.feishu_webhook
  return !value ? '-' : value.length > 64 ? `${value.slice(0, 64)}…` : value
}
function actionRenderer(params: { data: MonitorAlertChannel }) {
  const wrap = document.createElement('div'); wrap.className = 'channel-actions'
  const edit = document.createElement('button'); edit.className = 'channel-action edit'; edit.textContent = '编辑'; edit.onclick = () => open(params.data)
  const removeButton = document.createElement('button'); removeButton.className = 'channel-action delete'; removeButton.textContent = '删除'; removeButton.onclick = () => remove(params.data)
  wrap.append(edit, removeButton)
  return wrap
}
const columns: ColDef[] = [
  { headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node!.rowIndex! + 1, width: 70 },
  { field: 'name', headerName: '通道名称', minWidth: 150 },
  { field: 'channel_type', headerName: '通道类型', minWidth: 110, cellRenderer: (p: any) => createTag(labels[p.value as MonitorAlertChannelType] || p.value, `type-${p.value}`) },
  { field: 'enabled', headerName: '状态', width: 90, cellRenderer: (p: any) => createTag(p.value ? '启用' : '禁用', p.value ? 'enabled' : 'disabled') },
  { headerName: '配置预览', minWidth: 280, valueGetter: (p: any) => configPreview(p.data) },
  { headerName: '操作', pinned: 'right', width: 150, cellClass: 'channel-operation-cell', cellRenderer: actionRenderer },
]
async function load() { const { data } = await listMonitorChannels(keyword.value, page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total }
function open(row?: MonitorAlertChannel) {
  editing.value = row ?? null
  form.value = row ? { id: row.id, name: row.name, channel_type: row.channel_type, enabled: row.enabled, config: { ...emptyConfig(), ...row.config } } : emptyForm()
  testRecipient.value = ''
  dialogVisible.value = true
}
function validateConfig() {
  const c = form.value.config
  const required: Record<MonitorAlertChannelType, Array<[unknown, string]>> = {
    email: [[c.smtp_host, '请输入 SMTP 服务器'], [c.smtp_port, '请输入 SMTP 端口'], [c.from_email, '请输入发件人邮箱'], [c.smtp_user, '请输入 SMTP 用户名']],
    webhook: [[c.webhook_url, '请输入 Webhook URL']], wechat: [[c.wechat_webhook, '请输入企业微信 Webhook URL']],
    dingtalk: [[c.dingtalk_webhook, '请输入钉钉 Webhook URL']], feishu: [[c.feishu_webhook, '请输入飞书 Webhook URL']],
  }
  const missing = required[form.value.channel_type].find(([value]) => !value)
  if (missing) { ElMessage.warning(missing[1]); return false }
  return true
}
async function save() {
  if (!formRef.value) return
  await formRef.value.validate()
  if (!validateConfig()) return
  saving.value = true
  try { await saveMonitorChannel(form.value); ElMessage.success(editing.value ? '告警通道已更新' : '告警通道已创建'); dialogVisible.value = false; await load() }
  finally { saving.value = false }
}
async function remove(row: MonitorAlertChannel) {
  await ElMessageBox.confirm(`确定删除告警通道“${row.name}”？`, '确认删除', { type: 'warning' })
  await deleteMonitorChannel(row.id); ElMessage.success('删除成功'); await load()
}
async function batchRemove(selectedRows: unknown[]) {
  const selected = selectedRows as MonitorAlertChannel[]
  if (!selected.length) return
  await ElMessageBox.confirm(`确定删除选中的 ${selected.length} 个告警通道？`, '批量删除确认', { type: 'warning' })
  await Promise.all(selected.map(row => deleteMonitorChannel(row.id)))
  ElMessage.success('批量删除成功')
  await load()
}
async function testChannel() {
  if (!validateConfig()) return
  if (form.value.channel_type === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(testRecipient.value)) {
    ElMessage.warning('请输入正确的测试收件人邮箱')
    return
  }
  testing.value = true
  try {
    await testMonitorChannel({ channel_type: form.value.channel_type, config: form.value.config, recipient: testRecipient.value })
    ElMessage.success('测试消息发送成功')
  } finally { testing.value = false }
}
watch(keyword, () => { page.value = 1; void load() })
onMounted(load)
</script>

<template>
  <section class="channel-page">
    <div class="toolbar"><el-input v-model="keyword" placeholder="搜索通道名称" clearable />
      <div class="toolbar-actions"><el-button type="primary" @click="open()">新增通道</el-button></div>
    </div>
    <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total"
      :current-page="page" :page-size="pageSize" :bulk-delete="batchRemove"
      @page-change="value => { page = value; void load() }"
      @page-size-change="value => { pageSize = value; page = 1; void load() }" />
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑告警通道' : '新增告警通道'" width="min(720px, 72vw)"
      :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="105px" class="channel-form">
        <div class="form-grid">
          <el-form-item label="通道名称" prop="name" class="full"><el-input v-model="form.name"
              placeholder="请输入通道名称，如：企业微信通知" /></el-form-item>
          <el-form-item label="通道类型" prop="channel_type"><el-select v-model="form.channel_type"
              style="width: 100%"><el-option label="邮件" value="email" /><el-option label="钉钉"
                value="dingtalk" /><el-option label="飞书" value="feishu" /><el-option label="企业微信"
                value="wechat" /><el-option label="Webhook" value="webhook" /></el-select></el-form-item>
          <el-form-item label="启用状态"><el-switch v-model="form.enabled" /></el-form-item>
        </div>
        <template v-if="form.channel_type === 'email'">
          <el-divider content-position="left">邮件服务器配置</el-divider>
          <div class="form-grid"><el-form-item label="SMTP 服务器"><el-input v-model="form.config.smtp_host"
                placeholder="smtp.example.com" /></el-form-item><el-form-item label="SMTP 端口"><el-input-number
                v-model="form.config.smtp_port" :min="1" :max="65535"
                controls-position="right" /></el-form-item><el-form-item label="发件人邮箱"><el-input
                v-model="form.config.from_email" placeholder="noreply@example.com" /></el-form-item><el-form-item
              label="发件人名称"><el-input v-model="form.config.from_name" placeholder="系统监控" /></el-form-item><el-form-item
              label="SMTP 用户名"><el-input v-model="form.config.smtp_user"
                placeholder="user@example.com" /></el-form-item><el-form-item label="SMTP 密码"><el-input
                v-model="form.config.smtp_password" type="password" show-password
                autocomplete="new-password" /></el-form-item></div>
        </template>
        <template v-else-if="form.channel_type === 'dingtalk'"><el-divider
            content-position="left">钉钉机器人配置</el-divider><el-form-item label="Webhook URL"><el-input
              v-model="form.config.dingtalk_webhook"
              placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." /></el-form-item><el-form-item
            label="加签密钥"><el-input v-model="form.config.dingtalk_secret" type="password" show-password
              placeholder="SEC 开头的密钥（可选）" /></el-form-item></template>
        <template v-else-if="form.channel_type === 'feishu'"><el-divider
            content-position="left">飞书机器人配置</el-divider><el-form-item label="Webhook URL"><el-input
              v-model="form.config.feishu_webhook"
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." /></el-form-item></template>
        <template v-else-if="form.channel_type === 'wechat'"><el-divider
            content-position="left">企业微信机器人配置</el-divider><el-form-item label="Webhook URL"><el-input
              v-model="form.config.wechat_webhook"
              placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." /></el-form-item></template>
        <template v-else><el-divider content-position="left">Webhook 配置</el-divider><el-form-item
            label="Webhook URL"><el-input v-model="form.config.webhook_url"
              placeholder="https://example.com/webhook" /></el-form-item></template>
        <template v-if="form.channel_type === 'email'">
          <el-divider content-position="left">发送测试</el-divider>
          <el-form-item label="测试收件人"><el-input v-model="testRecipient" placeholder="test@example.com" /></el-form-item>
        </template>
      </el-form>
      <template #footer>
        <div class="dialog-footer"><el-button class="test-button" :loading="testing"
            @click="testChannel">发送测试</el-button>
          <div><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving"
              @click="save">保存</el-button></div>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.channel-page {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toolbar .el-input {
  width: 260px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.toolbar-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.grid {
  flex: 1;
  height: 100%;
  min-height: 0;
}

:deep(.grid-pagination) {
  width: 100%;
  margin-top: auto;
  justify-content: flex-end;
  align-items: center;
  padding-right: 1px;
  box-sizing: border-box;
}

.form-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  column-gap: 14px;
}

.form-grid .full {
  grid-column: 1 / -1;
}

.channel-form :deep(.el-divider) {
  margin: 8px 0 16px;
}

.channel-form :deep(.el-input-number) {
  width: 100%;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.test-button {
  color: var(--el-color-success);
  border-color: var(--el-color-success-light-5);
  background: var(--el-color-success-light-9);
}

:deep(.channel-operation-cell) {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

:deep(.channel-actions) {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
}

:deep(.channel-action) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  margin: 0;
  padding: 0 10px;
  line-height: 1;
  text-align: center;
  background: #fff;
  border: 1px solid;
  border-radius: 8px;
  cursor: pointer;
  box-shadow: 0 1px 3px rgb(31 41 55 / 8%);
}

:deep(.channel-action.edit) {
  color: var(--primary);
  border-color: var(--primary);
  background: color-mix(in srgb, var(--primary) 5%, white);
}

:deep(.channel-action.delete) {
  color: var(--el-color-danger);
  border-color: var(--el-color-danger-light-5);
  background: var(--el-color-danger-light-9);
}

:deep(.channel-tag) {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 9px;
  border: 1px solid;
  border-radius: 8px;
  line-height: 1;
}

:deep(.channel-tag.enabled) {
  color: var(--el-color-success);
  border-color: var(--el-color-success-light-7);
  background: var(--el-color-success-light-9);
}

:deep(.channel-tag.disabled) {
  color: var(--el-text-color-secondary);
  border-color: var(--el-border-color);
  background: var(--el-fill-color-light);
}

:deep(.channel-tag.type-email) {
  color: #d58a18;
  border-color: #f1d5a6;
  background: #fff9ef;
}

:deep(.channel-tag.type-dingtalk),
:deep(.channel-tag.type-feishu) {
  color: var(--primary);
  border-color: color-mix(in srgb, var(--primary) 30%, white);
  background: color-mix(in srgb, var(--primary) 6%, white);
}

:deep(.channel-tag.type-wechat) {
  color: var(--el-color-success);
  border-color: var(--el-color-success-light-7);
  background: var(--el-color-success-light-9);
}

:deep(.channel-tag.type-webhook) {
  color: var(--el-color-info);
  border-color: var(--el-color-info-light-7);
  background: var(--el-color-info-light-9);
}

@media (max-width: 760px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-grid .full {
    grid-column: auto;
  }
}
</style>
