<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteMonitorReceiver, listMonitorReceivers, saveMonitorReceiver, type MonitorAlertReceiver, type MonitorAlertReceiverRequest } from '@/api/monitor'

const emptyForm = (): MonitorAlertReceiverRequest => ({ name: '', email: '', phone: '', webhook: '', wechat_id: '', dingtalk_id: '', feishu_id: '', enable_email: true, enable_webhook: false, enable_wechat: false, enable_dingtalk: false, enable_feishu: false, enable_system_msg: true, enabled: true })
const rows = ref<MonitorAlertReceiver[]>([])
const keyword = ref('')
const dialogVisible = ref(false)
const saving = ref(false)
const editing = ref<MonitorAlertReceiver | null>(null)
const formRef = ref<FormInstance>()
const form = ref<MonitorAlertReceiverRequest>(emptyForm())
const rules: FormRules = { name: [{ required: true, message: '请输入接收人姓名', trigger: 'blur' }] }

function tag(text: string, className = '') { const node = document.createElement('span'); node.className = `receiver-tag ${className}`; node.textContent = text; return node }
function notificationText(row: MonitorAlertReceiver) {
  const items = [[row.enable_email, '邮件'], [row.enable_webhook, 'Webhook'], [row.enable_wechat, '企业微信'], [row.enable_dingtalk, '钉钉'], [row.enable_feishu, '飞书'], [row.enable_system_msg, '系统消息']]
  return items.filter(([enabled]) => enabled).map(([, text]) => text).join('、') || '-'
}
function actionRenderer(params: { data: MonitorAlertReceiver }) {
  const wrap = document.createElement('div'); wrap.className = 'receiver-actions'
  const edit = document.createElement('button'); edit.className = 'receiver-action edit'; edit.textContent = '编辑'; edit.onclick = () => open(params.data)
  const removeButton = document.createElement('button'); removeButton.className = 'receiver-action delete'; removeButton.textContent = '删除'; removeButton.onclick = () => remove(params.data)
  wrap.append(edit, removeButton); return wrap
}
const columns: ColDef[] = [
  { headerName: '序号', valueGetter: (p: any) => p.node!.rowIndex! + 1, width: 70 },
  { field: 'name', headerName: '姓名', minWidth: 130 },
  { field: 'email', headerName: '邮箱地址', minWidth: 180, valueFormatter: (p: any) => p.value || '-' },
  { field: 'phone', headerName: '手机号码', minWidth: 130, valueFormatter: (p: any) => p.value || '-' },
  { headerName: '通知方式', minWidth: 220, valueGetter: (p: any) => notificationText(p.data) },
  { field: 'enabled', headerName: '状态', width: 90, cellRenderer: (p: any) => tag(p.value ? '启用' : '禁用', p.value ? 'enabled' : 'disabled') },
  { headerName: '操作', pinned: 'right', width: 150, cellClass: 'receiver-operation-cell', cellRenderer: actionRenderer },
]
async function load() { const { data } = await listMonitorReceivers(keyword.value); rows.value = data.data }
function open(row?: MonitorAlertReceiver) { editing.value = row ?? null; form.value = row ? { ...row } : emptyForm(); dialogVisible.value = true }
function selectedCount() { const f = form.value; return [f.enable_email, f.enable_webhook, f.enable_wechat, f.enable_dingtalk, f.enable_feishu, f.enable_system_msg].filter(Boolean).length }
async function save() {
  if (!formRef.value) return
  await formRef.value.validate()
  if (!selectedCount()) { ElMessage.warning('请至少选择一种通知方式'); return }
  saving.value = true
  try { await saveMonitorReceiver(form.value); ElMessage.success(editing.value ? '接收人已更新' : '接收人已创建'); dialogVisible.value = false; await load() }
  finally { saving.value = false }
}
async function remove(row: MonitorAlertReceiver) { await ElMessageBox.confirm(`确定删除接收人“${row.name}”？`, '确认删除', { type: 'warning' }); await deleteMonitorReceiver(row.id); ElMessage.success('删除成功'); await load() }
watch(keyword, load)
onMounted(load)
</script>

<template>
  <section class="receiver-page">
    <div class="toolbar"><el-input v-model="keyword" placeholder="搜索接收人姓名或邮箱" clearable /><el-button type="primary" @click="open()">新增接收人</el-button></div>
    <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination />
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑告警接收人' : '新增告警接收人'" width="min(760px, 76vw)" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="105px" class="receiver-form">
        <el-divider content-position="left">基本信息</el-divider>
        <div class="form-grid"><el-form-item label="姓名" prop="name"><el-input v-model="form.name" placeholder="请输入接收人姓名" /></el-form-item><el-form-item label="启用状态"><el-switch v-model="form.enabled" /></el-form-item></div>
        <el-divider content-position="left">联系方式</el-divider>
        <div class="form-grid">
          <el-form-item label="邮箱地址"><el-input v-model="form.email" placeholder="example@email.com" /></el-form-item>
          <el-form-item label="手机号码"><el-input v-model="form.phone" placeholder="13800138000" /></el-form-item>
          <el-form-item label="企业微信 ID"><el-input v-model="form.wechat_id" placeholder="企业微信 User ID" /></el-form-item>
          <el-form-item label="钉钉 ID"><el-input v-model="form.dingtalk_id" placeholder="钉钉 User ID" /></el-form-item>
          <el-form-item label="飞书 ID"><el-input v-model="form.feishu_id" placeholder="飞书 User ID" /></el-form-item>
          <el-form-item label="Webhook"><el-input v-model="form.webhook" placeholder="https://example.com/webhook" /></el-form-item>
        </div>
        <el-divider content-position="left">通知方式</el-divider>
        <el-form-item label="启用的通知" class="notification-options">
          <el-checkbox v-model="form.enable_email">邮件通知</el-checkbox><el-checkbox v-model="form.enable_webhook">Webhook 通知</el-checkbox><el-checkbox v-model="form.enable_wechat">企业微信</el-checkbox><el-checkbox v-model="form.enable_dingtalk">钉钉通知</el-checkbox><el-checkbox v-model="form.enable_feishu">飞书通知</el-checkbox><el-checkbox v-model="form.enable_system_msg">系统消息</el-checkbox>
        </el-form-item>
        <el-alert title="至少选择一种通知方式，并填写对应的联系方式。" type="info" :closable="false" show-icon />
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.receiver-page { height: 100%; min-height: 0; display: flex; flex-direction: column; gap: 8px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar .el-input { width: 280px; }
.grid { flex: 1; height: 100%; min-height: 0; }
:deep(.grid-pagination) { width: 100%; margin-top: auto; justify-content: flex-end; padding-right: 1px; box-sizing: border-box; }
.receiver-form :deep(.el-divider) { margin: 8px 0 16px; }
.form-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); column-gap: 14px; }
.notification-options :deep(.el-form-item__content) { display: grid; grid-template-columns: repeat(3, minmax(120px, 1fr)); gap: 4px 12px; }
.notification-options :deep(.el-checkbox) { margin-right: 0; }
:deep(.receiver-operation-cell) { display: flex; align-items: center; justify-content: flex-start; }
:deep(.receiver-actions) { width: 100%; height: 100%; display: flex; align-items: center; justify-content: flex-start; gap: 8px; }
:deep(.receiver-action) { display: inline-flex; align-items: center; justify-content: center; height: 28px; margin: 0; padding: 0 10px; line-height: 1; border: 1px solid; border-radius: 8px; cursor: pointer; }
:deep(.receiver-action.edit) { color: var(--primary); border-color: var(--primary); background: color-mix(in srgb, var(--primary) 5%, white); }
:deep(.receiver-action.delete) { color: var(--el-color-danger); border-color: var(--el-color-danger-light-5); background: var(--el-color-danger-light-9); }
:deep(.receiver-tag) { display: inline-flex; align-items: center; height: 24px; padding: 0 9px; border: 1px solid; border-radius: 8px; line-height: 1; }
:deep(.receiver-tag.enabled) { color: var(--el-color-success); border-color: var(--el-color-success-light-7); background: var(--el-color-success-light-9); }
:deep(.receiver-tag.disabled) { color: var(--el-text-color-secondary); border-color: var(--el-border-color); background: var(--el-fill-color-light); }
@media (max-width: 760px) { .form-grid { grid-template-columns: 1fr; } .notification-options :deep(.el-form-item__content) { grid-template-columns: repeat(2, minmax(110px, 1fr)); } }
</style>
