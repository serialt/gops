<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createTaskTemplate, deleteTaskTemplate, listTaskTemplates, updateTaskTemplate, type TaskTemplate, type TaskTemplateRequest } from '@/api/task'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const rows = ref<TaskTemplate[]>([])
const keyword = ref('')
const category = ref('')
const platform = ref('')
const status = ref<number | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const dialog = ref(false)
const form = reactive<TaskTemplateRequest>({ name: '', code: '', description: '', content: '', variables: '[]', category: 'shell', platform: 'linux', timeout: 300, sort_order: 0, status: 1 })

async function load() {
  loading.value = true
  try {
    const { data } = await listTaskTemplates(keyword.value, category.value, platform.value, status.value, page.value, pageSize.value)
    rows.value = data.data.list
    total.value = data.data.total
  } finally { loading.value = false }
}
function open(row?: TaskTemplate) {
  Object.assign(form, row ? { ...row } : { id: undefined, name: '', code: '', description: '', content: '', variables: '[]', category: 'shell', platform: 'linux', timeout: 300, sort_order: 0, status: 1 })
  dialog.value = true
}
async function save() {
  if (!form.name.trim() || !form.code.trim() || !form.content.trim()) return ElMessage.warning('请填写模板名称、编码和内容')
  if (form.id) await updateTaskTemplate(form as TaskTemplateRequest & { id: number })
  else await createTaskTemplate(form)
  dialog.value = false
  ElMessage.success('保存成功')
  await load()
}
async function remove(row: TaskTemplate, confirm = true) {
  if (confirm) await ElMessageBox.confirm(`确认删除模板“${row.name}”吗？`, '删除确认', { type: 'warning' })
  await deleteTaskTemplate(row.id)
}
async function bulkDelete(selected: unknown[]) {
  const items = selected as TaskTemplate[]
  await ElMessageBox.confirm(`确认删除选中的 ${items.length} 个模板吗？`, '批量删除', { type: 'warning' })
  for (const item of items) await remove(item, false)
  ElMessage.success('删除成功')
  await load()
}
function actionRenderer(params: { data: TaskTemplate }) {
  const wrap = document.createElement('div'); wrap.className = 'task-actions'
  if (auth.hasPermission('POST /api/v1/tasks/templates/update')) { const edit = document.createElement('button'); edit.className = 'task-edit'; edit.textContent = '编辑'; edit.onclick = () => open(params.data); wrap.append(edit) }
  if (auth.hasPermission('POST /api/v1/tasks/templates/delete')) { const del = document.createElement('button'); del.className = 'task-delete'; del.textContent = '删除'; del.onclick = () => { void remove(params.data).then(load) }; wrap.append(del) }
  return wrap
}
function tagRenderer(value: string, label: string) { const span = document.createElement('span'); span.className = `task-tag ${value}`; span.textContent = label; return span }
const centeredCell = { display: 'flex', alignItems: 'center', justifyContent: 'center', textAlign: 'center' }
const columns = computed<ColDef[]>(() => [
  { headerName: '序号', width: 64, headerClass: 'task-center-header', cellStyle: centeredCell, valueGetter: params => (page.value - 1) * pageSize.value + (params.node?.rowIndex || 0) + 1 },
  { field: 'name', headerName: '模板名称', minWidth: 150, headerClass: 'task-center-header', cellStyle: centeredCell },
  { field: 'code', headerName: '模板编码', minWidth: 130, headerClass: 'task-center-header', cellStyle: centeredCell },
  { field: 'category', headerName: '脚本类型', width: 100, headerClass: 'task-center-header', cellStyle: centeredCell, cellRenderer: (params: { value: string }) => tagRenderer(params.value, params.value === 'python' ? 'Python' : 'Shell') },
  { field: 'platform', headerName: '平台', width: 90, headerClass: 'task-center-header', cellStyle: centeredCell, cellRenderer: (params: { value: string }) => tagRenderer('platform', params.value === 'windows' ? 'Windows' : params.value === 'linux' ? 'Linux' : '不限') },
  { field: 'timeout', headerName: '超时', width: 82, headerClass: 'task-center-header', cellStyle: centeredCell, valueFormatter: params => `${params.value} 秒` },
  { field: 'description', headerName: '说明', minWidth: 180, headerClass: 'task-center-header', cellStyle: centeredCell, valueFormatter: params => params.value || '-' },
  { field: 'status', headerName: '状态', width: 82, headerClass: 'task-center-header', cellStyle: centeredCell, cellRenderer: (params: { value: number }) => tagRenderer(params.value ? 'enabled' : 'disabled', params.value ? '启用' : '停用') },
  { headerName: '操作', pinned: 'right', width: 150, cellStyle: { display: 'flex', alignItems: 'center' }, cellRenderer: actionRenderer },
])
watch([keyword, category, platform, status], () => { page.value = 1; void load() })
onMounted(load)
</script>

<template>
  <section class="page">
    <header><div><h2>模板管理</h2><p>维护可复用的 Shell 和 Python 执行模板。</p></div></header>
    <div class="toolbar">
      <div class="filters"><el-input v-model="keyword" placeholder="搜索模板名称或编码" clearable /><el-select v-model="category" placeholder="选择脚本类型" clearable><el-option label="Shell" value="shell" /><el-option label="Python" value="python" /></el-select><el-select v-model="platform" placeholder="选择平台" clearable><el-option label="Linux" value="linux" /><el-option label="Windows" value="windows" /></el-select><el-select v-model="status" placeholder="选择状态" clearable><el-option label="启用" :value="1" /><el-option label="停用" :value="0" /></el-select></div>
      <el-button v-permission="'POST /api/v1/tasks/templates/create'" type="primary" @click="open()">新增模板</el-button>
    </div>
    <AppDataGrid v-loading="loading" class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total" :current-page="page" :page-size="pageSize" :bulk-delete="auth.hasPermission('POST /api/v1/tasks/templates/delete') ? bulkDelete : undefined" @page-change="value => { page = value; void load() }" @page-size-change="value => { pageSize = value; page = 1; void load() }" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑模板' : '新增模板'" width="760px" destroy-on-close>
      <el-form label-width="76px">
        <div class="form-row"><el-form-item label="模板名称" required><el-input v-model="form.name" /></el-form-item><el-form-item label="模板编码" required><el-input v-model="form.code" /></el-form-item></div>
        <div class="form-row"><el-form-item label="脚本类型" required><el-select v-model="form.category"><el-option label="Shell" value="shell" /><el-option label="Python" value="python" /></el-select></el-form-item><el-form-item label="适用平台"><el-select v-model="form.platform" clearable><el-option label="Linux" value="linux" /><el-option label="Windows" value="windows" /></el-select></el-form-item></div>
        <div class="form-row"><el-form-item label="超时时间"><el-input-number v-model="form.timeout" :min="1" :max="3600" /></el-form-item><el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" /></el-form-item></div>
        <el-form-item label="模板内容" required><el-input v-model="form.content" type="textarea" :rows="11" class="code-input" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.page{height:100%;display:flex;flex-direction:column;gap:8px;min-width:0;padding:0 4px;box-sizing:border-box;font-size:11px}header h2{margin:0;color:var(--text);font-size:20px}header p{margin:3px 0 0;color:var(--text-secondary)}.toolbar{display:flex;align-items:center;justify-content:space-between;gap:8px}.filters{display:flex;gap:6px;min-width:0;flex-wrap:wrap}.filters .el-input{width:190px}.filters .el-select{width:145px}.grid{flex:1;min-height:0}.form-row{display:grid;grid-template-columns:1fr 1fr;gap:12px}.form-row .el-select{width:100%}.code-input :deep(textarea){font-family:ui-monospace,SFMono-Regular,Menlo,monospace}:deep(.task-center-header .ag-header-cell-label){justify-content:center}:deep(.task-actions){height:100%;display:flex;align-items:center;justify-content:flex-start;gap:6px}:deep(.task-edit),:deep(.task-delete){width:auto!important;height:26px!important;line-height:16px!important;border-radius:8px;padding:4px 9px;cursor:pointer;background:#fff}:deep(.task-edit){border:1px solid var(--primary);color:var(--primary);box-shadow:0 1px 2px rgba(65,80,190,.1)}:deep(.task-edit:hover){background:rgba(91,103,239,.06)}:deep(.task-delete){border:1px solid #f3b4b4;color:#ef5b5b;background:#fff8f8}:deep(.task-delete:hover){background:#fff1f1}:deep(.task-tag){display:inline-flex!important;align-items:center;width:auto!important;height:24px!important;box-sizing:border-box;border:1px solid var(--line);border-radius:7px;padding:2px 7px;line-height:16px;background:#f7f8fb}:deep(.task-tag.enabled){color:#43b936;background:#f2faef;border-color:#d6edcf}:deep(.task-tag.disabled){color:#8d94a3}:deep(.task-tag.shell),:deep(.task-tag.python){color:var(--primary);background:#f3f4ff;border-color:#d8dbff}:deep(.task-tag.platform){color:#596273;background:#f7f8fb;border-color:#dfe3eb}@media(max-width:720px){.form-row{grid-template-columns:1fr}}
.page :deep(.el-input__inner::placeholder),.page :deep(.el-textarea__inner::placeholder),.page :deep(.el-select__placeholder.is-transparent){font-size:10px}
</style>
