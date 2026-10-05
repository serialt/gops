<script setup lang="ts">
import { h, onBeforeUnmount, reactive, ref, render, watch, type Component } from 'vue'
import { ElMessage, ElMessageBox, ElTooltip } from 'element-plus'
import { Delete, Edit } from '@element-plus/icons-vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createIOTExternalBroker, deleteIOTExternalBroker, listIOTExternalBrokers, updateIOTExternalBroker, type IOTExternalBroker } from '@/api/iot'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const rows = ref<IOTExternalBroker[]>([])
const keyword = ref('')
const enabledFilter = ref<boolean | undefined>()
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const loading = ref(false)
const saving = ref(false)
const dialog = ref(false)
const form = reactive({ id: 0, name: '', url: 'tcp://', client_id: '', username: '', password: '', topic_filter: '#', enabled: true, insecure_skip_verify: false, has_password: false })
let timer: ReturnType<typeof setTimeout> | undefined

function sequence(params: any) { return (page.value - 1) * limit.value + params.node.rowIndex + 1 }
function actionButton(label: string, kind: string, icon: Component, onClick: () => void) {
  const root = document.createElement('span')
  render(h(ElTooltip, { content: label, placement: 'top', showAfter: 250 }, { default: () => h('button', { type: 'button', class: `grid-action compact icon-only ${kind}`, 'aria-label': label, onClick }, [h('span', { class: 'el-icon' }, [h(icon)])]) }), root)
  return root
}
function actions(params: any) {
  const container = document.createElement('div')
  container.className = 'grid-action-group'
  if (auth.hasPermission('POST /api/v1/iot/brokers/update')) container.append(actionButton('编辑', 'edit', Edit, () => openEdit(params.data)))
  if (auth.hasPermission('POST /api/v1/iot/brokers/delete')) container.append(actionButton('删除', 'delete', Delete, () => void remove(params.data)))
  return container
}
function statusCell(params: any) {
  const element = document.createElement('span')
  const connected = params.data.connected
  element.className = `broker-state ${connected ? 'is-connected' : params.data.enabled ? 'is-pending' : 'is-disabled'}`
  element.textContent = connected ? '已连接' : params.data.enabled ? '未连接' : '已禁用'
  return element
}
const columns: ColDef[] = [
  { headerName: '序号', valueGetter: sequence, width: 62 },
  { field: 'name', headerName: '名称', minWidth: 120 },
  { field: 'url', headerName: '连接地址', minWidth: 210 },
  { field: 'client_id', headerName: 'Client ID', minWidth: 150 },
  { field: 'username', headerName: '用户名', minWidth: 110, valueFormatter: ({ value }) => value || '-' },
  { field: 'topic_filter', headerName: '订阅 Topic', minWidth: 130, flex: 1 },
  { headerName: '状态', width: 90, cellRenderer: statusCell },
  { headerName: '操作', width: 150, minWidth: 150, maxWidth: 150, pinned: 'right', resizable: false, sortable: false, cellRenderer: actions },
]

async function load() {
  loading.value = true
  try {
    const response = await listIOTExternalBrokers({ page: page.value, limit: limit.value, keyword: keyword.value.trim(), enabled: enabledFilter.value })
    rows.value = response.data.data.list
    total.value = response.data.data.total
  } finally { loading.value = false }
}
function reset() { Object.assign(form, { id: 0, name: '', url: 'tcp://', client_id: `gops-${Date.now()}`, username: '', password: '', topic_filter: '#', enabled: true, insecure_skip_verify: false, has_password: false }) }
function openCreate() { reset(); dialog.value = true }
function openEdit(row: IOTExternalBroker) { Object.assign(form, { ...row, password: '' }); dialog.value = true }
async function save() {
  if (!form.name.trim() || !form.url.trim() || !form.client_id.trim() || !form.topic_filter.trim()) { ElMessage.warning('请填写名称、连接地址、Client ID 和订阅 Topic'); return }
  saving.value = true
  try {
    if (form.id) await updateIOTExternalBroker({ ...form })
    else await createIOTExternalBroker({ ...form })
    dialog.value = false
    ElMessage.success('保存成功')
    await load()
  } finally { saving.value = false }
}
async function remove(row: IOTExternalBroker) {
  await ElMessageBox.confirm(`确定删除 Broker「${row.name}」吗？`, '删除确认', { type: 'warning' })
  await deleteIOTExternalBroker(row.id)
  ElMessage.success('删除成功')
  if (rows.value.length === 1 && page.value > 1) page.value -= 1
  await load()
}
watch(keyword, () => { if (timer) clearTimeout(timer); timer = setTimeout(() => { page.value = 1; void load() }, 300) })
watch(enabledFilter, () => { page.value = 1; void load() })
watch([page, limit], () => void load(), { immediate: true })
onBeforeUnmount(() => { if (timer) clearTimeout(timer) })
</script>

<template>
  <section class="iot-page">
    <header><div><h1>Broker 管理</h1><p>管理外部 MQTT Broker 连接与订阅。</p></div><el-button v-permission="'POST /api/v1/iot/brokers/create'" type="primary" @click="openCreate">新增 Broker</el-button></header>
    <div class="toolbar"><el-input v-model="keyword" placeholder="搜索 Broker" clearable/><el-select v-model="enabledFilter" placeholder="全部状态" clearable><el-option label="已启用" :value="true"/><el-option label="已禁用" :value="false"/></el-select></div>
    <AppDataGrid v-loading="loading" class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total" :current-page="page" :page-size="limit" @page-change="value=>{page=value}" @page-size-change="value=>{limit=value;page=1}"/>
    <el-dialog v-model="dialog" :title="form.id ? '编辑 Broker' : '新增 Broker'" width="620px" destroy-on-close>
      <el-form label-position="top">
        <div class="form-grid"><el-form-item label="名称" required><el-input v-model="form.name" placeholder="例如生产环境"/></el-form-item><el-form-item label="Client ID" required><el-input v-model="form.client_id"/></el-form-item></div>
        <el-form-item label="连接地址" required><el-input v-model="form.url" placeholder="tcp://host:1883、tls://host:8883、ws://host/mqtt 或 wss://host/mqtt"/></el-form-item>
        <div class="form-grid"><el-form-item label="用户名"><el-input v-model="form.username"/></el-form-item><el-form-item :label="form.id && form.has_password ? '密码（留空保持不变）' : '密码'"><el-input v-model="form.password" type="password" show-password autocomplete="new-password"/></el-form-item></div>
        <el-form-item label="订阅 Topic" required><el-input v-model="form.topic_filter" placeholder="#、devices/# 或 devices/+/status"/></el-form-item>
        <div class="switches"><el-form-item label="启用"><el-switch v-model="form.enabled"/></el-form-item><el-form-item label="跳过 TLS 证书校验"><el-switch v-model="form.insecure_skip_verify"/></el-form-item></div>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.iot-page{height:100%;min-width:0;min-height:0;display:flex;flex-direction:column;gap:8px;padding:0 8px 8px;box-sizing:border-box;font-size:11px}.iot-page header{display:flex;align-items:center;justify-content:space-between;gap:12px}.iot-page h1{margin:0;font-size:18px}.iot-page header p{margin:3px 0 0;color:var(--text-muted)}.toolbar{display:flex;align-items:center;gap:8px}.toolbar :deep(.el-input){width:118px}.toolbar :deep(.el-select){width:108px}.grid{min-width:0;min-height:0;flex:1}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}.switches{display:flex;gap:40px}:deep(.broker-state){display:inline-flex;align-items:center;gap:5px}:deep(.broker-state::before){width:7px;height:7px;border-radius:50%;background:var(--text-muted);content:''}:deep(.broker-state.is-connected::before){background:var(--success)}:deep(.broker-state.is-pending::before){background:var(--danger)}@media(max-width:700px){.form-grid{grid-template-columns:1fr}.toolbar{flex-wrap:wrap}}
</style>
