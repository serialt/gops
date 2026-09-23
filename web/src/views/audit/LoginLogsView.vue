<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteAuditLogins, listAuditLogins, type AuditLoginLog } from '@/api/audit'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const rows = ref<AuditLoginLog[]>([])
const selected = ref<AuditLoginLog[]>([])
const loading = ref(false)
const keyword = ref('')
const loginType = ref('')
const status = ref('')
const dateRange = ref<[string, string] | null>(null)
let loadTimer = 0

const columns: ColDef[] = [
  { headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 },
  { field: 'username', headerName: '用户名', minWidth: 120 },
  { field: 'login_type', headerName: '登录类型', width: 100, cellRenderer: (p: any) => tagRenderer(String(p.value || '-').toUpperCase(), `type-${String(p.value || '').toLowerCase()}`) },
  { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => tagRenderer(p.value === 'success' ? '成功' : '失败', p.value === 'success' ? 'success' : 'failed') },
  { field: 'ip', headerName: 'IP 地址', minWidth: 130 },
  { field: 'message', headerName: '说明 / 失败原因', minWidth: 220, tooltipField: 'message' },
  { headerName: '操作', pinned: 'right', width: 80, headerClass: 'grid-action-header', cellClass: 'grid-action-cell', cellRenderer: (p: any) => actionRenderer(p.data) },
]

function tagRenderer(label: string, tone: string) {
  const tag = document.createElement('span')
  tag.className = `audit-tag ${tone}`
  tag.textContent = label
  return tag
}

async function load() {
  loading.value = true
  try {
    const response = await listAuditLogins({ keyword: keyword.value, login_type: loginType.value, status: status.value, start_time: dateRange.value?.[0] || '', end_time: dateRange.value?.[1] || '' })
    rows.value = response.data.data
    selected.value = []
  } finally {
    loading.value = false
  }
}

function scheduleLoad() { window.clearTimeout(loadTimer); loadTimer = window.setTimeout(load, 250) }

async function remove(ids: number[]) {
  await ElMessageBox.confirm(`确定删除选中的 ${ids.length} 条登录日志吗？`, '删除确认', { type: 'warning' })
  await deleteAuditLogins(ids)
  ElMessage.success('删除成功')
  await load()
}

function actionRenderer(row: AuditLoginLog) {
  const wrap = document.createElement('div')
  if (!auth.hasPermission('POST /api/v1/audit/login-logs/delete')) return wrap
  const button = document.createElement('button')
  button.className = 'grid-delete-button'
  button.textContent = '删除'
  button.onclick = () => remove([row.id])
  wrap.append(button)
  return wrap
}

function reset() { keyword.value = ''; loginType.value = ''; status.value = ''; dateRange.value = null }

watch([keyword, loginType, status, dateRange], scheduleLoad, { immediate: true })
onBeforeUnmount(() => window.clearTimeout(loadTimer))
</script>

<template>
  <section class="audit-page" v-loading="loading">
    <header class="page-header">
      <div><h2>登录日志</h2><p>记录系统用户登录结果、来源及失败原因</p></div>
      <div class="header-actions"><el-button text @click="reset">重置筛选</el-button><el-button v-if="selected.length" v-permission="'POST /api/v1/audit/login-logs/delete'" type="danger" plain @click="remove(selected.map(item => item.id))">批量删除（{{ selected.length }}）</el-button></div>
    </header>
    <div class="filters">
      <el-input v-model="keyword" placeholder="搜索用户名、IP 或说明" clearable />
      <el-select v-model="loginType" placeholder="全部类型" clearable><el-option label="Web" value="web" /><el-option label="SSH" value="ssh" /><el-option label="API" value="api" /></el-select>
      <el-select v-model="status" placeholder="全部状态" clearable><el-option label="成功" value="success" /><el-option label="失败" value="failed" /></el-select>
      <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
    </div>
    <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination :get-row-id="(row: any) => String(row.id)" @selection-changed="selected = $event as AuditLoginLog[]" />
  </section>
</template>

<style scoped>
.audit-page{display:flex;flex-direction:column;width:100%;height:100%;min-width:0;min-height:0;gap:8px}.page-header{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:12px 14px;border:1px solid var(--line);border-radius:8px;background:#fff}.page-header h2{margin:0;color:var(--text);font-size:17px;line-height:1.3}.page-header p{margin:3px 0 0;color:var(--muted);font-size:10px}.header-actions,.filters{display:flex;align-items:center;gap:8px}.filters{flex-wrap:wrap;padding:10px;border:1px solid var(--line);border-radius:8px;background:#fff}.filters .el-input{width:250px}.filters .el-select{width:125px}.filters :deep(.el-date-editor){width:230px}.grid{flex:1;min-height:0}:deep(.grid-action-header),:deep(.grid-action-header .ag-header-cell-comp-wrapper),:deep(.grid-action-header .ag-header-cell-label){height:100%;align-items:center;justify-content:center}:deep(.grid-action-cell){display:flex;height:100%;align-items:center;justify-content:center;padding:0}:deep(.grid-action-cell .ag-cell-wrapper),:deep(.grid-action-cell .ag-cell-value),:deep(.grid-action-cell>div){display:flex;width:100%;height:100%;align-items:center;justify-content:center}:deep(.audit-tag){display:inline-flex;align-items:center;min-height:22px;padding:0 8px;border:1px solid #dfe4ed;border-radius:6px;background:#f7f8fb;color:#657085;font-size:11px;line-height:20px}:deep(.audit-tag.success),:deep(.audit-tag.type-web){border-color:#cce8c2;background:#f0f9eb;color:#54b72c}:deep(.audit-tag.failed){border-color:#f2c3c3;background:#fff1f1;color:var(--danger)}:deep(.audit-tag.type-ssh){border-color:#f1d9aa;background:#fff8e8;color:#b57a19}:deep(.audit-tag.type-api){border-color:#c7d2f7;background:#f1f4ff;color:var(--primary)}:deep(.grid-delete-button){height:24px;padding:0 10px;border:1px solid #f2b8b8;border-radius:8px;background:#fff7f7;color:var(--danger);font-size:11px;line-height:22px;cursor:pointer}:deep(.grid-delete-button:hover){background:#ffecec}@media(max-width:900px){.page-header{flex-wrap:wrap}.filters>*{flex:1 1 160px}.filters .el-input,.filters .el-select,.filters :deep(.el-date-editor){width:auto}}
</style>
