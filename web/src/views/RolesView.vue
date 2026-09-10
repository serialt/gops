<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createRole, deleteRole, getPermissionCatalog, getRolePermissions, getRoles, setRolePermissions, updateRole } from '@/api'
import type { Permission, Role } from '@/types'

const roles = ref<Role[]>([])
const catalog = ref<Permission[]>([])
const roleDialog = ref(false)
const permissionDrawer = ref(false)
const editingID = ref<number | null>(null)
const selectedRole = ref<Role | null>(null)
const checkedKeys = ref<string[]>([])
const form = reactive({ code: '', name: '', description: '' })
const grouped = computed(() => catalog.value.reduce<Record<string, Permission[]>>((result, item) => {
  const group = result[item.group] ?? (result[item.group] = [])
  group.push(item)
  return result
}, {}))

function groupKeys(group: string) {
  return (grouped.value[group] ?? []).map(item => item.key)
}

function isGroupChecked(group: string) {
  const keys = groupKeys(group)
  return keys.length > 0 && keys.every(key => checkedKeys.value.includes(key))
}

function isGroupIndeterminate(group: string) {
  const keys = groupKeys(group)
  const selectedCount = keys.filter(key => checkedKeys.value.includes(key)).length
  return selectedCount > 0 && selectedCount < keys.length
}

function toggleGroup(group: string, checked: boolean | string | number) {
  const keys = groupKeys(group)
  const remaining = checkedKeys.value.filter(key => !keys.includes(key))
  checkedKeys.value = checked ? [...remaining, ...keys] : remaining
}

function togglePermission(key: string, checked: boolean | string | number) {
  checkedKeys.value = checked
    ? [...new Set([...checkedKeys.value, key])]
    : checkedKeys.value.filter(item => item !== key)
}

function clearPermissions() {
  checkedKeys.value = []
}

async function load() {
  const [rolesRes, catalogRes] = await Promise.all([getRoles(), getPermissionCatalog()])
  roles.value = rolesRes.data.data
  catalog.value = catalogRes.data.data
}

function openCreate() {
  editingID.value = null
  Object.assign(form, { code: '', name: '', description: '' })
  roleDialog.value = true
}

function openEdit(role: Role) {
  editingID.value = role.id
  Object.assign(form, { code: role.code, name: role.name, description: role.description })
  roleDialog.value = true
}

async function saveRole() {
  if (!form.code || !form.name) return ElMessage.warning('请填写角色编码和名称')
  if (editingID.value) await updateRole(editingID.value, form)
  else await createRole(form)
  ElMessage.success('保存成功')
  roleDialog.value = false
  await load()
}

async function remove(role: Role) {
  await ElMessageBox.confirm(`确定删除角色 ${role.name}？`, '提示', { type: 'warning' })
  await deleteRole(role.id)
  ElMessage.success('删除成功')
  await load()
}

async function openPermissions(role: Role) {
  selectedRole.value = role
  const { data } = await getRolePermissions(role.id)
  checkedKeys.value = data.data.map(item => item.key)
  permissionDrawer.value = true
}

async function savePermissions() {
  if (!selectedRole.value) return
  const selected = catalog.value.filter(item => checkedKeys.value.includes(item.key))
  await setRolePermissions(selectedRole.value.id, selected)
  ElMessage.success('权限保存成功')
  permissionDrawer.value = false
}

onMounted(load)
</script>

<template>
  <div class="page-container table-shell">
    <div class="section-card">
      <div class="page-toolbar">
        <h2 class="page-title">角色列表</h2>
        <el-button v-permission="'POST /api/v1/roles/create'" type="primary" class="primary-btn" @click="openCreate">新增角色</el-button>
      </div>

      <el-table :data="roles" stripe class="data-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="角色编码" />
        <el-table-column prop="name" label="角色名称" />
        <el-table-column prop="description" label="说明" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="action-group">
              <el-button v-permission="'POST /api/v1/roles/permissions/get'" link type="success" @click="openPermissions(row)">配置权限</el-button>
              <el-button v-permission="'POST /api/v1/roles/update'" link type="primary" class="edit-btn" @click="openEdit(row)">编辑</el-button>
              <el-button v-if="row.code !== 'admin'" v-permission="'POST /api/v1/roles/delete'" link type="danger" @click="remove(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="roleDialog" :title="editingID ? '编辑角色' : '新增角色'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="角色编码" required><el-input v-model="form.code" :disabled="Boolean(editingID)" placeholder="例如 viewer" /></el-form-item>
        <el-form-item label="角色名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="roleDialog = false">取消</el-button><el-button type="primary" @click="saveRole">保存</el-button></template>
    </el-dialog>

    <el-drawer v-model="permissionDrawer" :title="`配置权限：${selectedRole?.name || ''}`" size="560px">
      <el-alert v-if="selectedRole?.code === 'admin'" title="超级管理员始终拥有全部权限，无需配置。" type="info" :closable="false" />
      <div v-else class="permission-selector">
        <div class="permission-toolbar">
          <span>已选择 <strong>{{ checkedKeys.length }}</strong> / {{ catalog.length }} 项权限</span>
          <el-button link type="primary" class="clear-selection-btn" :disabled="!checkedKeys.length" @click="clearPermissions">清空选择</el-button>
        </div>
        <div v-for="(items, group) in grouped" :key="group" class="permission-group">
          <div class="group-title">
            <el-checkbox
              :model-value="isGroupChecked(group)"
              :indeterminate="isGroupIndeterminate(group)"
              @change="toggleGroup(group, $event)"
            >
              {{ group }}
            </el-checkbox>
            <span class="group-count">{{ groupKeys(group).filter(key => checkedKeys.includes(key)).length }}/{{ items.length }}</span>
          </div>
          <div class="permission-list">
            <el-checkbox
              v-for="item in items"
              :key="item.key"
              :model-value="checkedKeys.includes(item.key)"
              class="permission-item"
              @change="togglePermission(item.key, $event)"
            >
              <span class="permission-name">{{ item.name }}</span>
              <span class="permission-api"><b>{{ item.method }}</b> {{ item.path }}</span>
            </el-checkbox>
          </div>
        </div>
      </div>
      <template #footer><el-button @click="permissionDrawer = false">取消</el-button><el-button v-if="selectedRole?.code !== 'admin'" v-permission="'POST /api/v1/roles/permissions/set'" type="primary" @click="savePermissions">保存</el-button></template>
    </el-drawer>
  </div>
</template>

<style scoped>
.table-shell {
  background: var(--bg);
  min-height: 100%;
}

.section-card {
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 16px;
  box-shadow: var(--shadow);
}

.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
}

.page-title {
  margin: 0;
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

.permission-group {
  margin-bottom: 14px;
  border: 1px solid var(--line);
  border-radius: 10px;
  overflow: hidden;
  background: #fff;
}

.permission-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  margin-bottom: 14px;
  color: var(--text-soft);
  font-size: 12px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-soft);
}

.permission-toolbar strong {
  color: var(--primary);
}

.clear-selection-btn {
  min-height: 30px !important;
  padding: 5px 10px !important;
  color: #536fe8 !important;
  background: #fff !important;
  border: 1px solid #536fe8 !important;
  border-radius: 8px !important;
  box-shadow: 0 3px 8px rgba(83, 111, 232, 0.12) !important;
  font-size: 11px;
}

.clear-selection-btn:hover:not(.is-disabled) {
  color: #405bcc !important;
  background: #eef2ff !important;
  border-color: #405bcc !important;
  transform: none !important;
}

.clear-selection-btn.is-disabled {
  color: #98a2b3 !important;
  background: #fff !important;
  border-color: #d0d5dd !important;
  box-shadow: none !important;
}

.group-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  margin: 0;
  padding: 10px 12px;
  border-bottom: 1px solid var(--line);
  background: #f8faff;
  color: var(--text);
}

.group-count {
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 500;
}

.permission-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
  padding: 4px;
}

.permission-list :deep(.el-checkbox) {
  margin: 0;
  padding: 8px;
  height: auto;
  border-radius: 7px;
  align-items: flex-start;
}

.permission-list :deep(.el-checkbox:hover) {
  background: var(--primary-soft);
}

.permission-item :deep(.el-checkbox__label) {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px;
  padding-left: 8px;
}

.permission-name {
  color: var(--text);
  font-weight: 600;
}

.permission-api {
  color: var(--text-soft);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
}

.permission-api b {
  color: #536fe8;
  font-weight: 700;
}

small {
  margin-left: 8px;
  color: var(--text-soft);
}
</style>
