<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createUser, deleteUser, getRoles, getUsers, updateUser } from '@/api'
import type { Role, User } from '@/types'

const loading = ref(false)
const users = ref<User[]>([])
const roles = ref<Role[]>([])
const keyword = ref('')
const dialogVisible = ref(false)
const editingID = ref<number | null>(null)
const title = computed(() => editingID.value ? '编辑用户' : '新增用户')
const form = reactive({ username: '', password: '', nickname: '', status: true, role_ids: [] as number[] })

async function load() {
  loading.value = true
  try {
    const [usersRes, rolesRes] = await Promise.all([getUsers(keyword.value), getRoles()])
    users.value = usersRes.data.data
    roles.value = rolesRes.data.data
  } finally { loading.value = false }
}

function openCreate() {
  editingID.value = null
  Object.assign(form, { username: '', password: '', nickname: '', status: true, role_ids: [] })
  dialogVisible.value = true
}

function openEdit(row: User) {
  editingID.value = row.id
  Object.assign(form, { username: row.username, password: '', nickname: row.nickname, status: row.status, role_ids: row.roles.map(r => r.id) })
  dialogVisible.value = true
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
</script>

<template>
  <div class="page-container table-shell">
    <div class="section-card">
      <div class="page-toolbar">
        <div class="toolbar-left">
          <el-input v-model="keyword" placeholder="用户名或昵称" clearable class="search-input" @keyup.enter="load" />
          <el-button class="search-btn" @click="load">查询</el-button>
        </div>
        <el-button v-permission="'POST /api/v1/users'" type="primary" class="primary-btn" @click="openCreate">新增用户</el-button>
      </div>

      <el-table v-loading="loading" :data="users" stripe class="data-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="nickname" label="昵称" />
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
              <el-button v-permission="'PUT /api/v1/users/:id'" link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'DELETE /api/v1/users/:id'" link type="danger" @click="remove(row)">删除</el-button>
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
        <el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
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
</style>
