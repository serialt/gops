<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef, ICellRendererParams } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createOrganization, deleteOrganization, getOrganizations, moveOrganization, updateOrganization } from '@/api/organization'
import { getRoles } from '@/api/role'
import { createUser, deleteUser, getUsers, updateUser } from '@/api/user'
import type { Organization, Role, User } from '@/types'

const loading = ref(false)
const users = ref<User[]>([])
const roles = ref<Role[]>([])
const organizations = ref<Organization[]>([])
const activeOrganizationId = ref<number | null>(null)
const organizationKeyword = ref('')
const organizationTreeRef = ref()
const organizationPaneWidth = ref(200)
const organizationPaneDragging = ref(false)
const organizationPaneStartX = ref(0)
const organizationPaneStartWidth = ref(200)
const keyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const selectedUsers = ref<User[]>([])
const dialogVisible = ref(false)
const editingID = ref<number | null>(null)
const title = computed(() => editingID.value ? '编辑用户' : '新增用户')
const form = reactive({ username: '', password: '', nickname: '', status: true, role_ids: [] as number[], organization_ids: [] as number[] })
function userActionButton(text: string, className: string, onClick: () => void) {
  const button = document.createElement('button')
  button.className = `grid-action ${className}`
  button.textContent = text
  button.onclick = onClick
  return button
}
const userColumnDefs = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: params => String((currentPage.value - 1) * pageSize.value + (params.node?.rowIndex ?? 0) + 1), width: 70 },
  { field: 'username', headerName: '用户名' },
  { field: 'nickname', headerName: '昵称' },
  { headerName: '所属组织', valueGetter: params => params.data.organizations.map((item: Organization) => item.name).join('、') },
  { headerName: '角色', valueGetter: params => params.data.roles.map((item: Role) => item.name).join('、'), minWidth: 140 },
  { headerName: '状态', minWidth: 100, cellRenderer: (params: ICellRendererParams<User>) => { const tag = document.createElement('span'); tag.className = `grid-status${params.data?.status ? '' : ' is-disabled'}`; tag.textContent = params.data?.status ? '启用' : '禁用'; return tag } },
  { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (params: any) => { const wrap = document.createElement('div'); wrap.className = 'grid-action-group'; wrap.append(userActionButton('编辑', 'edit', () => openEdit(params.data)), userActionButton('删除', 'delete', () => remove(params.data))); return wrap } },
])
const organizationDialogVisible = ref(false)
const organizationEditingId = ref<number | null>(null)
const organizationForm = reactive({ parent_id: null as number | null, name: '', code: '', manager: '', status: true })
const contextMenu = reactive({ visible: false, x: 0, y: 0, node: null as Organization | null })

const organizationTree = computed(() => {
  const map = new Map<number, any>(); const roots: any[] = []
  organizations.value.forEach(item => map.set(item.id, { ...item, label: item.name, children: [] }))
  organizations.value.forEach(item => item.parent_id ? map.get(item.parent_id)?.children.push(map.get(item.id)) : roots.push(map.get(item.id)))
  return roots
})
const visibleUsers = computed(() => {
  if (!activeOrganizationId.value) return users.value
  return users.value.filter(user => user.organizations.some(organization => organization.id === activeOrganizationId.value))
})
const pagedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return visibleUsers.value.slice(start, start + pageSize.value)
})
watch(organizationKeyword, value => organizationTreeRef.value?.filter(value))
watch(keyword, () => { loadUsers() })

function filterOrganizationNode(value: string, data: Organization) {
  if (!value) return true
  const keyword = value.trim().toLowerCase()
  return data.name.toLowerCase().includes(keyword) || data.code.toLowerCase().includes(keyword)
}

async function load() {
  currentPage.value = 1
  loading.value = true
  try {
    const [usersRes, rolesRes, organizationsRes] = await Promise.all([getUsers(keyword.value, activeOrganizationId.value), getRoles(), getOrganizations()])
    users.value = usersRes.data.data
    roles.value = rolesRes.data.data
    organizations.value = organizationsRes.data.data
  } finally { loading.value = false }
}

async function loadUsers() {
  currentPage.value = 1
  loading.value = true
  try {
    const usersRes = await getUsers(keyword.value, activeOrganizationId.value)
    users.value = usersRes.data.data
  } finally { loading.value = false }
}

function openCreate() {
  editingID.value = null
  Object.assign(form, { username: '', password: '', nickname: '', status: true, role_ids: [], organization_ids: activeOrganizationId.value ? [activeOrganizationId.value] : [] })
  dialogVisible.value = true
}

function openEdit(row: User) {
  editingID.value = row.id
  Object.assign(form, { username: row.username, password: '', nickname: row.nickname, status: row.status, role_ids: row.roles.map(r => r.id), organization_ids: row.organizations.map(o => o.id) })
  dialogVisible.value = true
}

function closeContextMenu() { contextMenu.visible = false }
function openOrganizationCreate(parent: Organization | null = null) {
  closeContextMenu(); organizationEditingId.value = null
  Object.assign(organizationForm, { parent_id: parent?.id ?? null, name: '', code: '', manager: '', status: true })
  organizationDialogVisible.value = true
}
function openOrganizationRename(row: Organization) {
  closeContextMenu(); organizationEditingId.value = row.id
  Object.assign(organizationForm, { parent_id: row.parent_id, name: row.name, code: row.code, manager: row.manager, status: row.status })
  organizationDialogVisible.value = true
}
async function saveOrganization() {
  if (!organizationForm.name || !organizationForm.code) return ElMessage.warning('请填写部门名称和编码')
  if (organizationEditingId.value) await updateOrganization(organizationEditingId.value, organizationForm)
  else await createOrganization(organizationForm)
  ElMessage.success('部门保存成功'); organizationDialogVisible.value = false; await load()
}
async function removeOrganization(row: Organization) {
  closeContextMenu()
  if (row.code === 'root') return ElMessage.warning('默认根组织不支持删除')
  await ElMessageBox.confirm(`确定删除部门 ${row.name}？请先移除其子部门。`, '删除部门', { type: 'warning' })
  await deleteOrganization(row.id); ElMessage.success('删除成功'); await load()
}
function handleOrganizationContextMenu(event: MouseEvent, data: Organization) {
  event.preventDefault(); Object.assign(contextMenu, { visible: true, x: event.clientX, y: event.clientY, node: data })
}
function siblingsOf(row: Organization) { return organizations.value.filter(item => (item.parent_id || null) === (row.parent_id || null)).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id) }
async function reorderOrganization(row: Organization, offset: number) {
  closeContextMenu(); const siblings = siblingsOf(row); const index = siblings.findIndex(item => item.id === row.id); const target = index + offset
  if (target < 0 || target >= siblings.length) return
  await moveOrganization(row.id, { parent_id: row.parent_id, sort_order: target }); await load()
}
async function handleOrganizationDrop(draggingNode: any, dropNode: any, dropType: 'before' | 'after' | 'inner') {
  const dragging = draggingNode.data as Organization; const target = dropNode.data as Organization
  if (dragging.code === 'root') return ElMessage.warning('默认根组织不支持移动')
  const parentID = dropType === 'inner' ? target.id : (target.parent_id || null)
  const siblings = organizations.value.filter(item => item.id !== dragging.id && (item.parent_id || null) === parentID).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
  let position = dropType === 'inner' ? siblings.length : siblings.findIndex(item => item.id === target.id)
  if (dropType === 'after') position += 1
  await moveOrganization(dragging.id, { parent_id: parentID, sort_order: Math.max(position, 0) }); ElMessage.success('部门层级已调整'); await load()
}
function allowOrganizationDrop(draggingNode: any, dropNode: any, type: 'prev' | 'inner' | 'next') {
  if (draggingNode.data.code === 'root') return false
  return !(type === 'inner' && draggingNode.data.id === dropNode.data.id)
}

async function save() {
  if (!form.username || !form.nickname || (!editingID.value && !form.password)) {
    ElMessage.warning('请填写必填字段')
    return
  }
  if (editingID.value) await updateUser(editingID.value, form)
  else await createUser(form)
  ElMessage.success('保存成功')
  dialogVisible.value = false
  await load()
}

async function remove(row: User) {
  await ElMessageBox.confirm(`确定删除用户 ${row.username}？`, '提示', { type: 'warning' })
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  await load()
}

function handleUserSelectionChanged(rows: unknown[]) { selectedUsers.value = rows as User[] }
async function batchRemoveUsers() {
  if (!selectedUsers.value.length) return
  await ElMessageBox.confirm(`确定删除已选择的 ${selectedUsers.value.length} 个用户？`, '批量删除用户', { type: 'warning' })
  await Promise.all(selectedUsers.value.map(user => deleteUser(user.id)))
  selectedUsers.value = []
  ElMessage.success('批量删除成功')
  await load()
}

onMounted(load)

function handleOrganizationNodeClick(node: Organization) { activeOrganizationId.value = node.id; loadUsers() }
function userIndex(index: number) { return (currentPage.value - 1) * pageSize.value + index + 1 }
function handlePageChange(page: number) { currentPage.value = page }
function handlePageSizeChange(size: number) { pageSize.value = size; currentPage.value = 1 }

function startOrganizationPaneDrag(event: PointerEvent) {
  if (event.button !== 0) return
  organizationPaneDragging.value = true
  organizationPaneStartX.value = event.clientX
  organizationPaneStartWidth.value = organizationPaneWidth.value
  document.addEventListener('pointermove', moveOrganizationPane)
  document.addEventListener('pointerup', stopOrganizationPaneDrag)
}

function moveOrganizationPane(event: PointerEvent) {
  if (!organizationPaneDragging.value) return
  organizationPaneWidth.value = Math.max(0, Math.min(360, organizationPaneStartWidth.value + event.clientX - organizationPaneStartX.value))
}

function stopOrganizationPaneDrag() {
  if (!organizationPaneDragging.value) return
  organizationPaneDragging.value = false
  if (organizationPaneWidth.value < 90) organizationPaneWidth.value = 0
  else if (organizationPaneWidth.value === 0) organizationPaneWidth.value = 200
  document.removeEventListener('pointermove', moveOrganizationPane)
  document.removeEventListener('pointerup', stopOrganizationPaneDrag)
}

function toggleOrganizationPane() {
  organizationPaneWidth.value = organizationPaneWidth.value === 0 ? 200 : 0
}
</script>

<template>
  <div class="page-container table-shell" @click="closeContextMenu">
    <div class="user-layout" :style="{ gridTemplateColumns: `${organizationPaneWidth}px 2px minmax(0, 1fr)` }">
      <aside class="organization-tree">
        <div class="tree-title"><span>组织架构</span><el-input v-model="organizationKeyword" class="organization-search"
            placeholder="搜索" clearable /></div><el-tree ref="organizationTreeRef" :data="organizationTree" node-key="id"
          draggable default-expand-all :filter-node-method="filterOrganizationNode" :allow-drop="allowOrganizationDrop"
          @node-click="handleOrganizationNodeClick" @node-contextmenu="handleOrganizationContextMenu"
          @node-drop="handleOrganizationDrop" />
      </aside>
      <div class="content-divider" :class="{ dragging: organizationPaneDragging }" role="separator" aria-label="调整组织树宽度"
        @pointerdown="startOrganizationPaneDrag" @dblclick="toggleOrganizationPane">
        <button class="content-divider-handle" type="button" :aria-label="organizationPaneWidth ? '收起组织树' : '展开组织树'"
          @pointerdown.stop @click.stop="toggleOrganizationPane">
          <el-icon>
            <component :is="organizationPaneWidth ? ArrowLeft : ArrowRight" />
          </el-icon>
        </button>
      </div>
      <div class="section-card">
        <div class="page-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="用户名或昵称" clearable class="search-input" />
          </div>
          <div class="toolbar-actions">
            <el-button v-if="selectedUsers.length" v-permission="'POST /api/v1/users/delete'" type="danger" plain
              @click="batchRemoveUsers">批量删除（{{ selectedUsers.length }}）</el-button>
            <el-button v-permission="'POST /api/v1/users/create'" type="primary" class="primary-btn"
              @click="openCreate">新增用户</el-button>
          </div>
        </div>

        <AppDataGrid v-loading="loading" :row-data="pagedUsers" :column-defs="userColumnDefs"
          @selection-changed="handleUserSelectionChanged" />
        <!--
        <el-table-column label="序号" type="index" :index="userIndex" width="70" :resizable="true" />
        <el-table-column prop="username" label="用户名" :width="userColumnWidths.username" :resizable="true" />
        <el-table-column prop="nickname" label="昵称" :width="userColumnWidths.nickname" :resizable="true" />
        <el-table-column label="所属组织" :width="userColumnWidths.organizations" :resizable="true"><template #default="{ row }"><el-tag v-for="organization in row.organizations" :key="organization.id" class="role-tag">{{ organization.name }}</el-tag></template></el-table-column>
<el-table-column label="角色" :width="userColumnWidths.roles" :resizable="true">
  <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role.id" class="role-tag">{{ role.name }}</el-tag>
          </template>
</el-table-column>
<el-table-column label="状态" width="100" :resizable="true">
  <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'danger'" class="status-tag">{{ row.status ? '启用' : '禁用' }}</el-tag>
          </template>
</el-table-column>
<el-table-column label="操作" width="150" fixed="right" :resizable="true">
  <template #default="{ row }">
            <div class="action-group">
              <el-button v-permission="'POST /api/v1/users/update'" link type="primary" class="edit-btn" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'POST /api/v1/users/delete'" link type="danger" @click="remove(row)">删除</el-button>
            </div>
          </template>
</el-table-column>
</el-table>
-->
        <div class="pagination-bar">
          <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]"
            :total="visibleUsers.length" layout="total, sizes, prev, pager, next, jumper"
            @current-change="handlePageChange" @size-change="handlePageSizeChange" />
        </div>
      </div>

      <el-dialog v-model="dialogVisible" :title="title" width="520px">
        <el-form :model="form" label-width="90px">
          <el-form-item label="用户名" required><el-input v-model="form.username" /></el-form-item>
          <el-form-item :label="editingID ? '新密码' : '密码'" :required="!editingID"><el-input v-model="form.password"
              type="password" show-password :placeholder="editingID ? '留空表示不修改' : ''" /></el-form-item>
          <el-form-item label="昵称" required><el-input v-model="form.nickname" /></el-form-item>
          <el-form-item label="角色"><el-select v-model="form.role_ids" multiple style="width: 100%"><el-option
                v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" /></el-select></el-form-item>
          <el-form-item label="所属组织"><el-tree-select v-model="form.organization_ids" :data="organizationTree"
              node-key="id" multiple show-checkbox check-strictly style="width:100%" /></el-form-item>
          <el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item>
        </el-form>
        <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary"
            @click="save">保存</el-button></template>
      </el-dialog>
      <el-dialog v-model="organizationDialogVisible" :title="organizationEditingId ? '修改部门' : '新建部门'" width="480px">
        <el-form :model="organizationForm" label-width="82px"><el-form-item label="上级部门"><el-tree-select
              v-model="organizationForm.parent_id" :data="organizationTree" node-key="id" check-strictly clearable
              style="width:100%" /></el-form-item><el-form-item label="部门名称" required><el-input
              v-model="organizationForm.name" /></el-form-item><el-form-item label="部门编码" required><el-input
              v-model="organizationForm.code"
              :disabled="organizations.find(item => item.id === organizationEditingId)?.code === 'root'" /></el-form-item><el-form-item
            label="负责人"><el-input v-model="organizationForm.manager" /></el-form-item><el-form-item
            label="状态"><el-switch v-model="organizationForm.status" /></el-form-item></el-form>
        <template #footer><el-button @click="organizationDialogVisible = false">取消</el-button><el-button type="primary"
            @click="saveOrganization">保存</el-button></template>
      </el-dialog>
      <div v-if="contextMenu.visible && contextMenu.node" class="organization-context-menu"
        :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }" @click.stop>
        <el-button v-permission="'POST /api/v1/organizations/update'" link
          @click="openOrganizationRename(contextMenu.node)">修改名称</el-button>
        <el-button v-permission="'POST /api/v1/organizations/create'" link
          @click="openOrganizationCreate(contextMenu.node)">新建部门</el-button>
        <el-button v-permission="'POST /api/v1/organizations/move'" link :disabled="contextMenu.node.code === 'root'"
          @click="reorderOrganization(contextMenu.node, -1)">同级上移</el-button>
        <el-button v-permission="'POST /api/v1/organizations/move'" link :disabled="contextMenu.node.code === 'root'"
          @click="reorderOrganization(contextMenu.node, 1)">同级下移</el-button>
        <el-button v-permission="'POST /api/v1/organizations/delete'" link type="danger"
          :disabled="contextMenu.node.code === 'root'" @click="removeOrganization(contextMenu.node)">删除</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.table-shell {
  background: var(--bg);
  height: calc(100% - 1px);
  min-height: calc(100% - 1px);
}

.user-layout {
  display: grid;
  gap: 0;
  height: 100%;
  min-height: 100%;
}

.organization-tree {
  min-width: 0;
  overflow: hidden;
  padding: 14px 8px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--bg-card);
  box-shadow: var(--shadow);
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

.tree-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 8px 10px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--line);
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;
}

.organization-search {
  width: 92px;
}

.organization-search :deep(.el-input__wrapper) {
  min-height: 28px;
  border-radius: 6px;
  box-shadow: 0 0 0 1px var(--line) inset;
}

.organization-tree :deep(.el-tree) {
  background: transparent;
  font-size: 11px;
}

.organization-tree :deep(.el-tree-node__content) {
  height: 32px;
  border-radius: 6px;
}

.section-card {
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 16px;
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  min-height: 100%;
}

.page-toolbar {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
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

.search-btn {
  border-radius: 6px;
  background: #fafafa;
  border: 1px solid var(--line);
  color: var(--text);
}

.primary-btn {
  border-radius: 6px;
}

.data-table {
  width: 100%;
  --el-table-border-color: var(--line);
  --el-table-text-color: var(--text);
  --el-table-header-text-color: var(--text);
  --el-table-header-bg-color: #fafafa;
  --el-table-row-hover-bg-color: #f0f0f0;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: auto;
  padding-top: 16px;
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

@media (max-width: 768px) {
  .user-layout {
    grid-template-columns: 1fr !important;
  }

  .content-divider {
    display: none;
  }

  .organization-tree {
    max-height: 260px;
    overflow: auto;
  }

  .page-toolbar,
  .toolbar-left {
    flex-wrap: wrap;
  }

  .search-input {
    width: min(240px, 100%);
  }
}

.section-card {
  min-width: 0;
}

.data-table :deep(.el-table__header-wrapper th) {
  background: #fafafa !important;
  font-weight: 600;
}

.data-table :deep(.el-table__body tr td) {
  padding-top: 12px;
  padding-bottom: 12px;
}

.role-tag,
.status-tag {
  border-radius: 4px;
  font-weight: 500;
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

.organization-context-menu {
  position: fixed;
  z-index: 3000;
  display: flex;
  flex-direction: column;
  min-width: 112px;
  padding: 5px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-card);
  box-shadow: 0 8px 22px rgb(31 45 61 / 18%);
}

.organization-context-menu :deep(.el-button) {
  justify-content: flex-start;
  margin: 0;
  padding: 7px 9px;
  font-size: 11px;
}
</style>
