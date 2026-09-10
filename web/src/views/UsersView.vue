<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createOrganization, createUser, deleteOrganization, deleteUser, getOrganizations, getRoles, getUsers, moveOrganization, updateOrganization, updateUser } from '@/api'
import type { Organization, Role, User } from '@/types'

const loading = ref(false)
const users = ref<User[]>([])
const roles = ref<Role[]>([])
const organizations = ref<Organization[]>([])
const activeOrganizationId = ref<number | null>(null)
const organizationKeyword = ref('')
const organizationTreeRef = ref()
const keyword = ref('')
const dialogVisible = ref(false)
const editingID = ref<number | null>(null)
const title = computed(() => editingID.value ? '编辑用户' : '新增用户')
const form = reactive({ username: '', password: '', nickname: '', status: true, role_ids: [] as number[], organization_ids: [] as number[] })
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
watch(organizationKeyword, value => organizationTreeRef.value?.filter(value))

function filterOrganizationNode(value: string, data: Organization) {
  if (!value) return true
  const keyword = value.trim().toLowerCase()
  return data.name.toLowerCase().includes(keyword) || data.code.toLowerCase().includes(keyword)
}

async function load() {
  loading.value = true
  try {
    const [usersRes, rolesRes, organizationsRes] = await Promise.all([getUsers(keyword.value, activeOrganizationId.value), getRoles(), getOrganizations()])
    users.value = usersRes.data.data
    roles.value = rolesRes.data.data
    organizations.value = organizationsRes.data.data
  } finally { loading.value = false }
}

async function loadUsers() {
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

onMounted(load)

function handleOrganizationNodeClick(node: Organization) { activeOrganizationId.value = node.id; loadUsers() }
</script>

<template>
  <div class="page-container table-shell" @click="closeContextMenu">
    <div class="user-layout">
      <aside class="organization-tree"><div class="tree-title"><span>组织架构</span><el-input v-model="organizationKeyword" class="organization-search" placeholder="搜索" clearable /></div><el-tree ref="organizationTreeRef" :data="organizationTree" node-key="id" draggable default-expand-all :filter-node-method="filterOrganizationNode" :allow-drop="allowOrganizationDrop" @node-click="handleOrganizationNodeClick" @node-contextmenu="handleOrganizationContextMenu" @node-drop="handleOrganizationDrop" /></aside>
      <div class="section-card">
      <div class="page-toolbar">
        <div class="toolbar-left">
          <el-input v-model="keyword" placeholder="用户名或昵称" clearable class="search-input" @keyup.enter="load" />
          <el-button class="search-btn" @click="load">查询</el-button>
        </div>
        <el-button v-permission="'POST /api/v1/users/create'" type="primary" class="primary-btn" @click="openCreate">新增用户</el-button>
      </div>

        <el-table v-loading="loading" :data="visibleUsers" stripe class="data-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column label="所属组织"><template #default="{ row }"><el-tag v-for="organization in row.organizations" :key="organization.id" class="role-tag">{{ organization.name }}</el-tag></template></el-table-column>
        <el-table-column label="角色">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role.id" class="role-tag">{{ role.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'danger'" class="status-tag">{{ row.status ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <div class="action-group">
              <el-button v-permission="'POST /api/v1/users/update'" link type="primary" class="edit-btn" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'POST /api/v1/users/delete'" link type="danger" @click="remove(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="title" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="用户名" required><el-input v-model="form.username" /></el-form-item>
        <el-form-item :label="editingID ? '新密码' : '密码'" :required="!editingID"><el-input v-model="form.password" type="password" show-password :placeholder="editingID ? '留空表示不修改' : ''" /></el-form-item>
        <el-form-item label="昵称" required><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="角色"><el-select v-model="form.role_ids" multiple style="width: 100%"><el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" /></el-select></el-form-item>
        <el-form-item label="所属组织"><el-tree-select v-model="form.organization_ids" :data="organizationTree" node-key="id" multiple show-checkbox check-strictly style="width:100%" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="organizationDialogVisible" :title="organizationEditingId ? '修改部门' : '新建部门'" width="480px">
      <el-form :model="organizationForm" label-width="82px"><el-form-item label="上级部门"><el-tree-select v-model="organizationForm.parent_id" :data="organizationTree" node-key="id" check-strictly clearable style="width:100%" /></el-form-item><el-form-item label="部门名称" required><el-input v-model="organizationForm.name" /></el-form-item><el-form-item label="部门编码" required><el-input v-model="organizationForm.code" :disabled="organizations.find(item => item.id === organizationEditingId)?.code === 'root'" /></el-form-item><el-form-item label="负责人"><el-input v-model="organizationForm.manager" /></el-form-item><el-form-item label="状态"><el-switch v-model="organizationForm.status" /></el-form-item></el-form>
      <template #footer><el-button @click="organizationDialogVisible=false">取消</el-button><el-button type="primary" @click="saveOrganization">保存</el-button></template>
    </el-dialog>
    <div v-if="contextMenu.visible && contextMenu.node" class="organization-context-menu" :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }" @click.stop>
      <el-button v-permission="'POST /api/v1/organizations/update'" link @click="openOrganizationRename(contextMenu.node)">修改名称</el-button>
      <el-button v-permission="'POST /api/v1/organizations/create'" link @click="openOrganizationCreate(contextMenu.node)">新建部门</el-button>
      <el-button v-permission="'POST /api/v1/organizations/move'" link :disabled="contextMenu.node.code === 'root'" @click="reorderOrganization(contextMenu.node, -1)">同级上移</el-button>
      <el-button v-permission="'POST /api/v1/organizations/move'" link :disabled="contextMenu.node.code === 'root'" @click="reorderOrganization(contextMenu.node, 1)">同级下移</el-button>
      <el-button v-permission="'POST /api/v1/organizations/delete'" link type="danger" :disabled="contextMenu.node.code === 'root'" @click="removeOrganization(contextMenu.node)">删除</el-button>
    </div>
      </div>
    </div>
</template>

<style scoped>
.table-shell {
  background: var(--bg);
  min-height: 100%;
}

.user-layout { display: grid; grid-template-columns: 200px minmax(0, 1fr); gap: 16px; }
.organization-tree { padding: 14px 8px; border: 1px solid var(--line); border-radius: 12px; background: var(--bg-card); box-shadow: var(--shadow); }
.tree-title { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 0 8px 10px; margin-bottom: 8px; border-bottom: 1px solid var(--line); font-size: 13px; font-weight: 700; white-space: nowrap; }
.organization-search { width: 92px; }
.organization-search :deep(.el-input__wrapper) { min-height: 28px; border-radius: 6px; box-shadow: 0 0 0 1px var(--line) inset; }
.organization-tree :deep(.el-tree) { background: transparent; font-size: 11px; }
.organization-tree :deep(.el-tree-node__content) { height: 32px; border-radius: 6px; }

.section-card {
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 16px;
  box-shadow: var(--shadow);
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

.action-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-btn {
  padding: 5px 9px !important;
  border-radius: 6px !important;
  color: #3658c9 !important;
  background: #eaf0ff !important;
}

.edit-btn:hover {
  color: #fff !important;
  background: #536fe8 !important;
}
.organization-context-menu { position: fixed; z-index: 3000; display: flex; flex-direction: column; min-width: 112px; padding: 5px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg-card); box-shadow: 0 8px 22px rgb(31 45 61 / 18%); }
.organization-context-menu :deep(.el-button) { justify-content: flex-start; margin: 0; padding: 7px 9px; font-size: 11px; }
</style>
