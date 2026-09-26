<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, ArrowRight, CollectionTag } from '@element-plus/icons-vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createAssetGroup, deleteAssetGroup, listAssetGroups, updateAssetGroup, type AssetGroup } from '@/api/asset'

interface AssetGroupTreeNode { id: number; label: string; count: number; children: AssetGroupTreeNode[] }

const rows = ref<AssetGroup[]>([])
const tableRows = ref<AssetGroup[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const keyword = ref('')
const treeKeyword = ref('')
const selectedID = ref<number | null>(null)
const treePaneWidth = ref(200)
const treePaneDragging = ref(false)
const treePaneStartX = ref(0)
const treePaneStartWidth = ref(200)
const groupTreeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>()
const dialogVisible = ref(false)
const editingID = ref<number | null>(null)
const selectedGroups = ref<AssetGroup[]>([])
const form = reactive({ name: '', code: '', parent_id: 0, description: '', sort_order: 0, status: true })

const parentOptions = computed(() => rows.value.filter((group) => group.id !== editingID.value))
const groupNameByID = computed(() => new Map(rows.value.map((group) => [group.id, group.name])))
const treeData = computed<AssetGroupTreeNode[]>(() => {
  const nodes = new Map<number, AssetGroupTreeNode>()
  const roots: AssetGroupTreeNode[] = []
  rows.value.forEach((group) => nodes.set(group.id, { id: group.id, label: group.name, count: 0, children: [] }))
  rows.value.forEach((group) => {
    const node = nodes.get(group.id)
    const parent = nodes.get(group.parent_id)
    if (!node) return
    if (group.parent_id && parent) parent.children.push(node)
    else roots.push(node)
  })
  nodes.forEach((node) => { node.count = node.children.length })
  return roots
})
const selectedName = computed(() => groupNameByID.value.get(selectedID.value ?? 0) ?? '全部资产分组')
function actionButton(text: string, kind: string, onClick: () => void) {
  const button = document.createElement('button')
  button.className = `asset-action ${kind}`
  button.textContent = text
  button.onclick = onClick
  return button
}

const columnDefs = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: (params) => String((page.value - 1) * pageSize.value + (params.node?.rowIndex ?? 0) + 1), width: 70 },
  { field: 'name', headerName: '分组名称', minWidth: 120 },
  { field: 'code', headerName: '分组编码', minWidth: 120 },
  { field: 'parent_id', headerName: '上级分组', valueFormatter: (params) => params.value ? (groupNameByID.value.get(params.value) ?? '未知分组') : '顶级分组', minWidth: 110 },
  { field: 'description', headerName: '说明', minWidth: 160 },
  { field: 'status', headerName: '状态', width: 90, cellRenderer: (params: { value: boolean }) => `<span class="grid-status${params.value ? '' : ' is-disabled'}">${params.value ? '启用' : '禁用'}</span>` },
  { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (params: { data: AssetGroup }) => { const wrap = document.createElement('div'); wrap.className = 'asset-actions'; wrap.append(actionButton('编辑', 'edit', () => openEdit(params.data)), actionButton('删除', 'delete', () => remove(params.data))); return wrap } },
])

function selectNode(node: AssetGroupTreeNode) { selectedID.value = node.id; page.value = 1; void loadTable() }
function filterTreeNode(value: string, data: AssetGroupTreeNode) { return !value || data.label.toLowerCase().includes(value.toLowerCase()) }
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
function toggleTreePane() { treePaneWidth.value = treePaneWidth.value === 0 ? 200 : 0 }
async function load() { const { data } = await listAssetGroups('', 1, 1000); rows.value = data.data.list; await loadTable() }
async function loadTable() { const { data } = await listAssetGroups(keyword.value, page.value, pageSize.value, selectedID.value || 0); tableRows.value = data.data.list; total.value = data.data.total }
function openCreate() { editingID.value = null; Object.assign(form, { name: '', code: '', parent_id: selectedID.value ?? 0, description: '', sort_order: 0, status: true }); dialogVisible.value = true }
function openEdit(row: AssetGroup) { editingID.value = row.id; Object.assign(form, row); dialogVisible.value = true }
async function save() { if (!form.name || !form.code) return ElMessage.warning('请填写分组名称和编码'); if (editingID.value) await updateAssetGroup({ ...form, id: editingID.value }); else await createAssetGroup(form); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: AssetGroup) { await ElMessageBox.confirm(`确定删除资产分组 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteAssetGroup(row.id); if (selectedID.value === row.id) selectedID.value = null; ElMessage.success('删除成功'); await load() }
function handleGroupSelectionChanged(groups: unknown[]) { selectedGroups.value = groups as AssetGroup[] }
async function batchRemoveGroups() {
  if (!selectedGroups.value.length) return
  await ElMessageBox.confirm(`确定删除已选择的 ${selectedGroups.value.length} 个资产分组？`, '批量删除确认', { type: 'warning' })
  await Promise.all(selectedGroups.value.map((group) => deleteAssetGroup(group.id)))
  if (selectedGroups.value.some((group) => group.id === selectedID.value)) selectedID.value = null
  selectedGroups.value = []
  ElMessage.success('批量删除成功')
  await load()
}

watch(treeKeyword, (value) => groupTreeRef.value?.filter(value))
watch(keyword, () => { page.value = 1; void loadTable() })
onMounted(load)
</script>

<template>
  <div class="page-container asset-group-page">
    <div class="asset-group-layout" :style="{ gridTemplateColumns: `${treePaneWidth}px 2px minmax(0, 1fr)` }">
      <aside class="asset-tree-panel">
        <div class="tree-title"><span>资产分组</span><el-input v-model="treeKeyword" class="tree-search" placeholder="搜索"
            clearable /></div>
        <el-tree ref="groupTreeRef" :data="treeData" node-key="id" highlight-current default-expand-all
          :current-node-key="selectedID" :filter-node-method="filterTreeNode" :expand-on-click-node="false"
          @node-click="selectNode">
          <template #default="{ data }"><span class="asset-tree-node"><el-icon>
                <CollectionTag />
              </el-icon><span class="asset-tree-label">{{ data.label }}</span><span class="asset-tree-count">({{
                data.count }})</span></span></template>
        </el-tree>
      </aside>
      <div class="content-divider" :class="{ dragging: treePaneDragging }" role="separator" aria-label="调整资产分组树宽度"
        @pointerdown="startTreePaneDrag" @dblclick="toggleTreePane">
        <button class="content-divider-handle" type="button" :aria-label="treePaneWidth ? '收起资产分组树' : '展开资产分组树'"
          @pointerdown.stop @click.stop="toggleTreePane">
          <el-icon>
            <component :is="treePaneWidth ? ArrowLeft : ArrowRight" />
          </el-icon>
        </button>
      </div>
      <section class="section-card asset-group-content">
        <div class="page-toolbar">
          <div class="toolbar-left"><el-input v-model="keyword" class="search-input" placeholder="搜索分组名称或编码"
              clearable />
          </div>
          <div class="toolbar-actions">
            <el-button v-if="selectedGroups.length" v-permission="'POST /api/v1/assets/groups/delete'" type="danger"
              plain @click="batchRemoveGroups">批量删除（{{ selectedGroups.length }}）</el-button>
            <el-button v-permission="'POST /api/v1/assets/groups/create'" type="primary" class="primary-btn"
              @click="openCreate">新增分组</el-button>
          </div>
        </div>
        <AppDataGrid class="asset-group-grid" :row-data="tableRows" :column-defs="columnDefs" pagination
          server-pagination :total="total" :current-page="page" :page-size="pageSize"
          @page-change="value => { page = value; void loadTable() }"
          @page-size-change="value => { pageSize = value; page = 1; void loadTable() }"
          @selection-changed="handleGroupSelectionChanged" />
      </section>
    </div>
    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑资产分组' : '新增资产分组'" width="460px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="上级分组"><el-select v-model="form.parent_id" placeholder="请选择上级分组"
            style="width: 100%"><el-option :value="0" label="顶级分组" /><el-option v-for="group in parentOptions"
              :key="group.id" :label="group.name" :value="group.id" /></el-select></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary"
          @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.asset-group-page,
.asset-group-layout,
.asset-group-content {
  min-width: 0;
  min-height: 0;
}

.asset-group-layout {
  display: grid;
  gap: 0;
  height: 100%;
  min-height: 100%;
}

.asset-tree-panel,
.section-card {
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-card);
  box-shadow: var(--shadow);
}

.asset-tree-panel {
  min-width: 0;
  overflow: hidden;
  padding: 4px 8px;
}

.tree-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 8px 10px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--line);
  color: var(--text);
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;
}

.tree-search {
  width: 92px;
}

.tree-search :deep(.el-input__wrapper) {
  min-height: 28px;
  border-radius: 6px;
  box-shadow: 0 0 0 1px var(--line) inset;
}

.asset-tree-panel :deep(.el-tree) {
  --el-tree-node-hover-bg-color: var(--primary-soft);
  background: transparent;
  font-size: 11px;
}

.asset-tree-panel :deep(.el-tree-node__content) {
  height: 32px;
  margin: 1px 0;
  border-radius: 7px;
}

.asset-tree-panel :deep(.is-current > .el-tree-node__content) {
  color: var(--primary);
  font-weight: 700;
  background: var(--primary-soft);
}

.asset-tree-node {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  gap: 7px;
  padding-right: 8px;
  color: var(--text-soft);
}

.asset-tree-node>.el-icon {
  width: 15px;
  height: 15px;
  flex: none;
  color: #55bd39;
  font-size: 15px;
}

.asset-tree-label {
  overflow: hidden;
  flex: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-tree-count {
  flex: none;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  font-weight: 400;
}

.asset-tree-panel :deep(.is-current > .el-tree-node__content .asset-tree-label) {
  color: var(--text);
  font-weight: 600;
}

.section-card {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding: 6px 8px 0;
}

.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input {
  width: 240px;
}

.primary-btn {
  border-radius: 6px;
}

.asset-group-grid {
  flex: 1;
  min-height: 0;
}

.asset-group-grid :deep(.app-data-grid-shell) {
  height: 100%;
}

.content-divider {
  position: relative;
  width: 2px;
  cursor: col-resize;
  touch-action: none;
  background: transparent;
}

.content-divider::before {
  position: absolute;
  z-index: 1;
  top: 0;
  bottom: 0;
  left: -6px;
  width: 13px;
  content: '';
}

.content-divider::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 2px;
  content: '';
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
  display: flex;
  align-items: center;
  justify-content: center;
  width: 15px;
  height: 20px;
  padding: 0;
  transform: translate(-50%, -50%);
  border: 1px solid var(--line);
  border-radius: 4px;
  background: #fff;
  color: #667085;
  box-shadow: 0 2px 5px rgba(31, 42, 68, .12);
  cursor: pointer;
}

.content-divider-handle:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.content-divider-handle :deep(.el-icon) {
  font-size: 10px;
}

:deep(.asset-actions) {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  padding-left: 12px;
}

:deep(.asset-action) {
  padding: 4px 9px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: #fff;
  font-size: 11px;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(24, 34, 53, .08);
}

:deep(.asset-action.edit) {
  border-color: var(--primary);
  color: var(--primary);
  background: #f7f8ff;
}

:deep(.asset-action.delete) {
  border-color: #ffd5d5;
  color: #ef6262;
  background: #fff7f7;
}

:deep(.grid-status) {
  display: inline-flex;
  padding: 3px 8px;
  border: 1px solid #d9f0d0;
  border-radius: 8px;
  color: #67c23a;
  background: #f0f9eb;
}

:deep(.grid-status.is-disabled) {
  border-color: #e4e7ed;
  color: #909399;
  background: #f4f4f5;
}

@media (max-width: 900px) {
  .asset-group-layout {
    grid-template-columns: 1fr !important;
  }

  .asset-tree-panel {
    max-height: 220px;
    overflow: auto;
  }

  .content-divider {
    display: none;
  }
}
</style>
