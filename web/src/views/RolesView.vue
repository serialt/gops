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
        <el-button v-permission="'POST /api/v1/roles'" type="primary" class="primary-btn" @click="openCreate">新增角色</el-button>
      </div>

      <el-table :data="roles" stripe class="data-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="角色编码" />
        <el-table-column prop="name" label="角色名称" />
        <el-table-column prop="description" label="说明" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="action-group">
              <el-button v-permission="'GET /api/v1/roles/:id/permissions'" link type="success" @click="openPermissions(row)">配置权限</el-button>
              <el-button v-permission="'PUT /api/v1/roles/:id'" link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button v-if="row.code !== 'admin'" v-permission="'DELETE /api/v1/roles/:id'" link type="danger" @click="remove(row)">删除</el-button>
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
      <el-checkbox-group v-else v-model="checkedKeys">
        <div v-for="(items, group) in grouped" :key="group" class="permission-group">
          <div class="group-title">{{ group }}</div>
          <div class="permission-list"><el-checkbox v-for="item in items" :key="item.key" :value="item.key">{{ item.name }} <small>{{ item.method }} {{ item.path }}</small></el-checkbox></div>
        </div>
      </el-checkbox-group>
      <template #footer><el-button @click="permissionDrawer = false">取消</el-button><el-button v-if="selectedRole?.code !== 'admin'" v-permission="'PUT /api/v1/roles/:id/permissions'" type="primary" @click="savePermissions">保存</el-button></template>
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
  --el-table-border-color: var(--line);
  --el-table-text-color: var(--text);
  --el-table-header-text-color: var(--text);
  --el-table-header-bg-color: #fafafa;
  --el-table-row-hover-bg-color: #f0f0f0;
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

.permission-group {
  margin-bottom: 20px;
}

.group-title {
  font-weight: 600;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--line);
  color: var(--text);
}

.permission-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
}

.permission-list :deep(.el-checkbox) {
  height: auto;
  align-items: flex-start;
}

small {
  margin-left: 8px;
  color: var(--text-soft);
}
</style>
