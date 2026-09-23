<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteAuditOperations, listAuditOperations, type AuditOperationLog } from '@/api/audit'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const rows = ref<AuditOperationLog[]>([])
const selected = ref<AuditOperationLog[]>([])
const loading = ref(false)
const keyword = ref('')
const module = ref('')
const action = ref('')
const status = ref('')
const dateRange = ref<[string, string] | null>(null)
let loadTimer = 0

const columns: ColDef[] = [
  { headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 },
  { field: 'username', headerName: '操作用户', minWidth: 110 },
  { field: 'module', headerName: '模块', minWidth: 100, cellRenderer: (p: any) => tagRenderer(String(p.value || '-'), 'neutral') },
  { field: 'action', headerName: '操作类型', width: 90, cellRenderer: (p: any) => tagRenderer(String(p.value || '-'), actionTone(String(p.value || ''))) },
  { field: 'description', headerName: '操作描述', minWidth: 180, tooltipField: 'description' },
  { field: 'method', headerName: '请求方法', width: 90, cellRenderer: (p: any) => tagRenderer(String(p.value || '-'), `method-${String(p.value || '').toLowerCase()}`) },
  { field: 'path', headerName: '请求路径', minWidth: 180, tooltipField: 'path' },
  { field: 'status', headerName: '状态码', width: 80, cellRenderer: (p: any) => tagRenderer(String(p.value ?? '-'), Number(p.value) >= 200 && Number(p.value) < 300 ? 'success' : 'failed') },
  { field: 'cost_time', headerName: '耗时', width: 90, valueFormatter: (p) => `${p.value ?? 0} ms` },
  { field: 'ip', headerName: 'IP 地址', minWidth: 120 },
  { headerName: '操作', pinned: 'right', width: 80, headerClass: 'grid-action-header', cellClass: 'grid-action-cell', cellRenderer: (p: any) => actionRenderer(p.data) },
]

function tagRenderer(label: string, tone: string) {
  const tag = document.createElement('span')
  tag.className = `audit-tag ${tone}`
  tag.textContent = label
  return tag
}

function actionTone(value: string) {
  return ({ 创建: 'success', 保存: 'success', 更新: 'warning', 配置: 'primary', 删除: 'failed', 查询: 'neutral' } as Record<string, string>)[value] || 'neutral'
}

async function load() {
  loading.value = true
  try {
    const response = await listAuditOperations({
      keyword: keyword.value,
      module: module.value,
      action: action.value,
      status: status.value,
      start_time: dateRange.value?.[0] || '',
      end_time: dateRange.value?.[1] || '',
    })
    rows.value = response.data.data
    selected.value = []
  } finally {
    loading.value = false
  }
}

function scheduleLoad() {
  window.clearTimeout(loadTimer)
  loadTimer = window.setTimeout(load, 250)
}

async function remove(ids: number[]) {
  await ElMessageBox.confirm(`确定删除选中的 ${ids.length} 条操作日志吗？`, '删除确认', { type: 'warning' })
  await deleteAuditOperations(ids)
  ElMessage.success('删除成功')
  await load()
}

function actionRenderer(row: AuditOperationLog) {
  const wrap = document.createElement('div')
  if (!auth.hasPermission('POST /api/v1/audit/operation-logs/delete')) return wrap
  const button = document.createElement('button')
  button.className = 'grid-delete-button'
  button.textContent = '删除'
  button.onclick = () => remove([row.id])
  wrap.append(button)
  return wrap
}

function reset() {
  keyword.value = ''
  module.value = ''
  action.value = ''
  status.value = ''
  dateRange.value = null
}

watch([keyword, module, action, status, dateRange], scheduleLoad, { immediate: true })
onBeforeUnmount(() => window.clearTimeout(loadTimer))
</script>

<template>
  <section class="audit-page" v-loading="loading">
    <header class="page-header">
      <div><h2>操作日志</h2><p>记录系统内的操作行为、请求状态及耗时</p></div>
      <div class="header-actions">
        <el-button text @click="reset">重置筛选</el-button>
        <el-button v-if="selected.length" v-permission="'POST /api/v1/audit/operation-logs/delete'" type="danger" plain @click="remove(selected.map(item => item.id))">批量删除（{{ selected.length }}）</el-button>
      </div>
    </header>
    <div class="filters">
      <el-input v-model="keyword" placeholder="搜索用户、模块、描述、路径或 IP" clearable />
      <el-select v-model="module" placeholder="全部模块" clearable><el-option v-for="item in ['系统管理','个人信息','操作审计','资产管理','容器管理','监控中心','任务中心']" :key="item" :label="item" :value="item" /></el-select>
      <el-select v-model="action" placeholder="全部操作" clearable><el-option v-for="item in ['查询','创建','更新','删除','登录','登出']" :key="item" :label="item" :value="item" /></el-select>
      <el-select v-model="status" placeholder="全部状态" clearable><el-option label="成功 (2xx)" value="2xx" /><el-option label="重定向 (3xx)" value="3xx" /><el-option label="客户端错误 (4xx)" value="4xx" /><el-option label="服务器错误 (5xx)" value="5xx" /></el-select>
      <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
    </div>
    <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination :get-row-id="(row: any) => String(row.id)" @selection-changed="selected = $event as AuditOperationLog[]" />
  </section>
</template>

<style scoped>
.audit-page{display:flex;flex-direction:column;width:100%;height:100%;min-width:0;min-height:0;gap:8px}.page-header{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:12px 14px;border:1px solid var(--line);border-radius:8px;background:#fff}.page-header h2{margin:0;color:var(--text);font-size:17px;line-height:1.3}.page-header p{margin:3px 0 0;color:var(--muted);font-size:10px}.header-actions,.filters{display:flex;align-items:center;gap:8px}.filters{flex-wrap:wrap;padding:10px;border:1px solid var(--line);border-radius:8px;background:#fff}.filters .el-input{width:250px}.filters .el-select{width:125px}.filters :deep(.el-date-editor){width:230px}.grid{flex:1;min-height:0}:deep(.grid-action-header),:deep(.grid-action-header .ag-header-cell-comp-wrapper),:deep(.grid-action-header .ag-header-cell-label){height:100%;align-items:center;justify-content:center}:deep(.grid-action-cell){display:flex;height:100%;align-items:center;justify-content:center;padding:0}:deep(.grid-action-cell .ag-cell-wrapper),:deep(.grid-action-cell .ag-cell-value),:deep(.grid-action-cell>div){display:flex;width:100%;height:100%;align-items:center;justify-content:center}:deep(.audit-tag){display:inline-flex;align-items:center;min-height:22px;padding:0 8px;border:1px solid #dfe4ed;border-radius:6px;background:#f7f8fb;color:#727b8c;font-size:11px;line-height:20px}:deep(.audit-tag.success){border-color:#cce8c2;background:#f0f9eb;color:#54b72c}:deep(.audit-tag.failed){border-color:#f2c3c3;background:#fff1f1;color:var(--danger)}:deep(.audit-tag.warning){border-color:#f1d9aa;background:#fff8e8;color:#b57a19}:deep(.audit-tag.primary),:deep(.audit-tag.method-post){border-color:#c7d2f7;background:#f1f4ff;color:var(--primary)}:deep(.audit-tag.method-get){border-color:#dfe4ed;background:#f5f6f8;color:#727b8c}:deep(.grid-delete-button){height:24px;padding:0 10px;border:1px solid #f2b8b8;border-radius:8px;background:#fff7f7;color:var(--danger);font-size:11px;line-height:22px;cursor:pointer}:deep(.grid-delete-button:hover){background:#ffecec}@media(max-width:900px){.page-header{flex-wrap:wrap}.filters>*{flex:1 1 160px}.filters .el-input,.filters .el-select,.filters :deep(.el-date-editor){width:auto}}
</style>
