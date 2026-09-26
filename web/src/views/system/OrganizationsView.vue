<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createOrganization, deleteOrganization, getOrganizations, updateOrganization } from '@/api/organization'
import type { Organization } from '@/types'

const keyword = ref('')
const selectedId = ref<number | null>(null)
const organizations = ref<Organization[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const treePaneWidth = ref(220)
const treePaneDragging = ref(false)
const treePaneStartX = ref(0)
const treePaneStartWidth = ref(220)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const tableRows = ref<Organization[]>([])
const selectedOrganizations = ref<Organization[]>([])
const form = reactive({ parent_id: null as number | null, name: '', code: '', manager: '', status: true })

const treeData = computed(() => {
  const map = new Map<number, any>()
  const roots: any[] = []
  organizations.value.forEach(item => map.set(item.id, { id: item.id, label: item.name, children: [] }))
  organizations.value.forEach(item => {
    const node = map.get(item.id)
    if (item.parent_id) map.get(item.parent_id)?.children.push(node)
    else roots.push(node)
  })
  return roots
})

const selectedName = computed(() => organizations.value.find(item => item.id === selectedId.value)?.name || '全部组织')
function organizationActionButton(text: string, className: string, onClick: () => void) {
  const button = document.createElement('button')
  button.className = `grid-action ${className}`
  button.textContent = text
  button.onclick = onClick
  return button
}
const organizationColumnDefs = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: params => String((currentPage.value - 1) * pageSize.value + (params.node?.rowIndex ?? 0) + 1), width: 70 },
  { field: 'name', headerName: '组织名称' }, { field: 'code', headerName: '组织编码' }, { field: 'manager', headerName: '负责人' },
  { headerName: '状态', cellRenderer: (params: ICellRendererParams<Organization>) => { const tag = document.createElement('span'); tag.className = `grid-status${params.data?.status ? '' : ' is-disabled'}`; tag.textContent = params.data?.status ? '正常' : '停用'; return tag } },
  { headerName: '更新时间', valueGetter: params => formatDate(params.data.updated_at) },
  { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (params: any) => { const wrap = document.createElement('div'); wrap.className = 'grid-action-group'; wrap.append(organizationActionButton('编辑', 'edit', () => openEdit(params.data)), organizationActionButton('删除', 'delete', () => remove(params.data))); return wrap } },
])
watch([keyword, selectedId], () => { currentPage.value = 1; void loadTable() })

function selectNode(node: { id: number }) {
  selectedId.value = node.id
}

function formatDate(value: string) {
  return value ? value.replace('T', ' ').slice(0, 19) : '-'
}

function startTreePaneDrag(event: PointerEvent) {
  if (event.button !== 0) return
  treePaneDragging.value = true
  treePaneStartX.value = event.clientX
  treePaneStartWidth.value = treePaneWidth.value
  document.addEventListener('pointermove', moveTreePane)
  document.addEventListener('pointerup', stopTreePaneDrag)
}

function moveTreePane(event: PointerEvent) {
  if (!treePaneDragging.value) return
  treePaneWidth.value = Math.max(0, Math.min(360, treePaneStartWidth.value + event.clientX - treePaneStartX.value))
}

function stopTreePaneDrag() {
  if (!treePaneDragging.value) return
  treePaneDragging.value = false
  if (treePaneWidth.value < 90) treePaneWidth.value = 0
  document.removeEventListener('pointermove', moveTreePane)
  document.removeEventListener('pointerup', stopTreePaneDrag)
}

function toggleTreePane() {
  treePaneWidth.value = treePaneWidth.value === 0 ? 220 : 0
}

async function load() {
  const { data } = await getOrganizations('', 1, 1000)
  organizations.value = data.data.list
  await loadTable()
}
async function loadTable() {
  const { data } = await getOrganizations(keyword.value, currentPage.value, pageSize.value, selectedId.value || 0)
  tableRows.value = data.data.list
  total.value = data.data.total
}
function openCreate() { editingId.value = null; Object.assign(form, { parent_id: selectedId.value, name: '', code: '', manager: '', status: true }); dialogVisible.value = true }
function openEdit(row: Organization) { editingId.value = row.id; Object.assign(form, { parent_id: row.parent_id, name: row.name, code: row.code, manager: row.manager, status: row.status }); dialogVisible.value = true }
async function save() { if (!form.name || !form.code) return ElMessage.warning('请填写组织名称和编码'); if (editingId.value) await updateOrganization(editingId.value, form); else await createOrganization(form); ElMessage.success('保存成功'); dialogVisible.value = false; await load() }
async function remove(row: Organization) { await ElMessageBox.confirm(`确定删除组织 ${row.name}？`, '提示', { type: 'warning' }); await deleteOrganization(row.id); ElMessage.success('删除成功'); if (selectedId.value === row.id) selectedId.value = null; await load() }
function handleOrganizationSelectionChanged(rows: unknown[]) { selectedOrganizations.value = rows as Organization[] }
async function batchRemoveOrganizations() {
  const deletable = selectedOrganizations.value.filter(organization => organization.code !== 'root')
  if (!deletable.length) return ElMessage.warning('默认根组织不支持删除')
  await ElMessageBox.confirm(`确定删除已选择的 ${deletable.length} 个组织？`, '批量删除组织', { type: 'warning' })
  await Promise.all(deletable.map(organization => deleteOrganization(organization.id)))
  selectedOrganizations.value = []
  ElMessage.success('批量删除成功')
  await load()
}
onMounted(load)
function organizationIndex(index: number) { return (currentPage.value - 1) * pageSize.value + index + 1 }
function handlePageChange(page: number) { currentPage.value = page; void loadTable() }
function handlePageSizeChange(size: number) { pageSize.value = size; currentPage.value = 1; void loadTable() }
</script>

<template>
  <div class="page-container organization-page">
    <div class="organization-layout" :style="{ gridTemplateColumns: `${treePaneWidth}px 2px minmax(0, 1fr)` }">
      <aside class="tree-panel">
        <div class="panel-title"><span>组织架构</span><el-button v-permission="'POST /api/v1/organizations/create'"
            :icon="Plus" circle text title="新增组织" @click="openCreate" /></div>
        <el-tree :data="treeData" node-key="id" highlight-current default-expand-all :current-node-key="selectedId"
          :expand-on-click-node="false" @node-click="selectNode" />
      </aside>
      <div class="content-divider" :class="{ dragging: treePaneDragging }" role="separator" aria-label="调整组织树宽度"
        @pointerdown="startTreePaneDrag" @dblclick="toggleTreePane">
        <button class="content-divider-handle" type="button" :aria-label="treePaneWidth ? '收起组织树' : '展开组织树'"
          @pointerdown.stop @click.stop="toggleTreePane">
          <el-icon>
            <component :is="treePaneWidth ? ArrowLeft : ArrowRight" />
          </el-icon>
        </button>
      </div>

      <section class="organization-content">
        <div class="section-card">
          <div class="page-toolbar">
            <div>
              <h2 class="page-title">{{ selectedName }}</h2>
              <p class="page-subtitle">查看和维护该组织下的部门信息</p>
            </div>
            <div class="toolbar-actions">
              <el-button v-if="selectedOrganizations.length" v-permission="'POST /api/v1/organizations/delete'"
                type="danger" plain @click="batchRemoveOrganizations">批量删除（{{ selectedOrganizations.length
                }}）</el-button>
              <el-button v-permission="'POST /api/v1/organizations/create'" type="primary" :icon="Plus"
                @click="openCreate">新增组织</el-button>
            </div>
          </div>
          <div class="list-toolbar">
            <el-input v-model="keyword" placeholder="搜索组织名称或编码" clearable :prefix-icon="Search" />
          </div>
          <AppDataGrid class="organization-grid" fill-height :row-data="tableRows" :column-defs="organizationColumnDefs"
            @selection-changed="handleOrganizationSelectionChanged" />
          <div class="pagination-bar">
            <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize"
              :page-sizes="[20, 50, 100, 500, 1000]" :total="total" layout="total, sizes, prev, pager, next, jumper"
              @current-change="handlePageChange" @size-change="handlePageSizeChange" />
          </div>
        </div>
      </section>
    </div>
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑组织' : '新增组织'" width="480px">
      <el-form :model="form" label-width="82px"><el-form-item label="上级组织"><el-select v-model="form.parent_id" clearable
            style="width:100%"><el-option v-for="item in organizations.filter(item => item.id !== editingId)"
              :key="item.id" :label="item.name" :value="item.id" /></el-select></el-form-item><el-form-item label="组织名称"
          required><el-input v-model="form.name" /></el-form-item><el-form-item label="组织编码" required><el-input
            v-model="form.code" /></el-form-item><el-form-item label="负责人"><el-input
            v-model="form.manager" /></el-form-item><el-form-item label="状态"><el-switch
            v-model="form.status" /></el-form-item></el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary"
          @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.organization-layout {
  display: grid;
  gap: 0;
  height: 100%;
  min-height: 100%;
}

.organization-content {
  display: flex;
  min-width: 0;
  min-height: 0;
}

.tree-panel,
.section-card {
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--bg-card);
  box-shadow: var(--shadow);
}

.tree-panel {
  min-width: 0;
  overflow: hidden;
  padding: 2px 8px 0;
}

.content-divider {
  width: 2px;
  cursor: col-resize;
  touch-action: none;
  position: relative;
  background: transparent;
}

.content-divider::before {
  content: '';
  position: absolute;
  z-index: 1;
  top: 0;
  bottom: 0;
  left: -6px;
  width: 13px;
}

.content-divider::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 2px;
  background: var(--line);
}

.content-divider:hover::after,
.content-divider.dragging::after {
  background: var(--primary);
}

.content-divider-handle {
  position: absolute;
  z-index: 2;
  top: 50%;
  left: 50%;
  width: 15px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transform: translate(-50%, -50%);
  border: 1px solid var(--line);
  border-radius: 4px;
  background: #fff;
  color: #667085;
  box-shadow: 0 2px 5px rgba(31, 42, 68, 0.12);
  cursor: pointer;
}

.content-divider-handle:hover {
  color: var(--primary);
  border-color: var(--primary);
}

.content-divider-handle :deep(.el-icon) {
  font-size: 10px;
}

.panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2px 8px 10px;
  color: var(--text);
  font-size: 13px;
  font-weight: 700;
  border-bottom: 1px solid var(--line);
}

.panel-title :deep(.el-button) {
  color: var(--primary);
}

.tree-panel :deep(.el-tree) {
  --el-tree-node-hover-bg-color: var(--primary-soft);
  background: transparent;
  padding-top: 8px;
  font-size: 12px;
}

.tree-panel :deep(.el-tree-node__content) {
  height: 34px;
  border-radius: 6px;
  margin: 1px 0;
}

.tree-panel :deep(.is-current > .el-tree-node__content) {
  color: var(--primary);
  font-weight: 700;
  background: var(--primary-soft);
}

.section-card {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: 4px 2px 0;
  min-width: 0;
  min-height: 0;
}

.page-toolbar {
  margin-bottom: 4px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-subtitle {
  margin: 5px 0 0;
  color: var(--text-soft);
  font-size: 11px;
}

.list-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 4px;
}

.list-toolbar :deep(.el-input) {
  width: 230px;
}

.data-table {
  width: 100%;
}

.organization-grid {
  flex: 1;
  min-height: 0;
}

.organization-grid :deep(.app-data-grid-shell) {
  height: 100%;
}

:deep(.grid-action-group) {
  position: absolute;
  inset: 0;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  gap: 8px;
  white-space: nowrap;
}

:deep(.grid-action) {
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 4px 9px;
  background: #fff;
  font: inherit;
  font-size: 11px;
  cursor: pointer;
  line-height: 1.25;
  box-shadow: 0 1px 3px rgba(24, 34, 53, .08);
}

:deep(.grid-action.edit) {
  border-color: var(--primary);
  color: var(--primary);
  background: #f7f8ff;
}

:deep(.grid-action.delete) {
  border-color: #ffd5d5;
  color: #ef6262;
  background: #fff7f7;
}

:deep(.grid-status) {
  display: inline-flex;
  padding: 4px 9px;
  border: 1px solid #d9f0cf;
  border-radius: 6px;
  color: #58b83d;
  background: #f0f9eb;
  line-height: 1.25;
}

:deep(.grid-status.is-disabled) {
  border-color: #e4e7ed;
  color: #909399;
  background: #f4f4f5;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: auto;
  padding-top: 4px;
  overflow-x: auto;
}

.pagination-bar :deep(.el-pagination) {
  flex-wrap: nowrap;
  min-width: max-content;
  --el-pagination-font-size: 10px;
  font-size: 10px;
}

.pagination-bar :deep(.el-pagination .el-select .el-input__wrapper),
.pagination-bar :deep(.el-pagination .el-input__wrapper),
.pagination-bar :deep(.el-pagination .el-input__inner) {
  font-size: 10px;
}

@media (max-width: 900px) {
  .organization-layout {
    grid-template-columns: 1fr !important;
  }

  .tree-panel {
    max-height: 220px;
    overflow: auto;
  }

  .content-divider {
    display: none;
  }
}
</style>
