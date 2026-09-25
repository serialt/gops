<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createCredential, deleteCredential, listCredentials, updateCredential, type Credential } from '@/api/asset'

const rows = ref<Credential[]>([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const keyword = ref(''); const dialogVisible = ref(false); const editingID = ref<number | null>(null)
const form = reactive({ name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' })
function action(text: string, kind: string, fn: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `credential-action ${kind}`; el.onclick = fn; return el }
const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '凭证名称', minWidth: 120 }, { field: 'credential_type', headerName: '类型', width: 100, cellRenderer: (p: { value: string }) => `<span class="credential-type ${p.value === 'key' ? 'is-key' : 'is-password'}">${p.value === 'key' ? '密钥' : '密码'}</span>` }, { field: 'username', headerName: '用户名', minWidth: 120 }, { field: 'description', headerName: '说明', minWidth: 160 }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'credential-actions'; wrap.append(action('编辑', 'edit', () => edit(p.data)), action('删除', 'delete', () => remove(p.data))); return wrap } }])
async function load(reset = false) { if (reset) page.value = 1; const { data } = await listCredentials(keyword.value, page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total }
function changePage(value: number) { page.value = value; void load() }
function changePageSize(value: number) { pageSize.value = value; page.value = 1; void load() }
function create() { editingID.value = null; Object.assign(form, { name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); dialogVisible.value = true }
function edit(row: Credential) { editingID.value = row.id; Object.assign(form, { name: row.name, credential_type: row.credential_type, username: row.username, password: '', private_key: '', passphrase: '', description: row.description }); dialogVisible.value = true }
async function save() {
    if (!form.name.trim()) return ElMessage.warning('请填写凭证名称')
    if (!editingID.value && form.credential_type === 'password' && !form.password) return ElMessage.warning('请填写密码')
    if (!editingID.value && form.credential_type === 'key' && !form.private_key.trim()) return ElMessage.warning('请填写私钥')
    if (editingID.value) await updateCredential({ ...form, id: editingID.value }); else await createCredential(form)
    dialogVisible.value = false; ElMessage.success('保存成功'); await load()
}
async function remove(row: Credential) { await ElMessageBox.confirm(`确定删除凭证 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteCredential(row.id); ElMessage.success('删除成功'); await load() }
onMounted(load)
</script>
<template>
    <div class="page-container table-shell">
        <div class="section-card">
            <div class="page-toolbar">
                <h2 class="page-title">凭证管理</h2><el-button v-permission="'POST /api/v1/assets/credentials/create'"
                    type="primary" class="primary-btn" @click="create">新增凭证</el-button>
            </div><el-input v-model="keyword" class="credential-search" placeholder="搜索凭证名称或用户名" clearable
                @input="load(true)" />
            <AppDataGrid :row-data="rows" :column-defs="columns" pagination server-pagination :total="total" :current-page="page" :page-size="pageSize" @page-change="changePage" @page-size-change="changePageSize" />
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑凭证' : '新增凭证'" width="560px" class="credential-dialog">
            <el-form :model="form" label-width="92px" class="credential-form">
                <el-form-item label="凭证名称" required><el-input v-model="form.name" placeholder="请输入凭证名称，如：生产环境 root 凭证" /></el-form-item>
                <el-form-item label="认证方式" required>
                    <el-radio-group v-model="form.credential_type"><el-radio value="password">密码认证</el-radio><el-radio value="key">密钥认证</el-radio></el-radio-group>
                </el-form-item>
                <el-form-item label="用户名"><el-input v-model="form.username" placeholder="如：root" /></el-form-item>
                <el-form-item v-if="form.credential_type === 'password'" label="密码" required>
                    <el-input v-model="form.password" type="password" show-password :placeholder="editingID ? '留空则不修改密码' : '请输入密码'" />
                </el-form-item>
                <template v-else>
                    <el-form-item label="私钥" required><el-input v-model="form.private_key" type="textarea" :rows="4" placeholder="请输入 SSH 私钥内容" /></el-form-item>
                    <el-form-item label="私钥口令"><el-input v-model="form.passphrase" type="password" show-password :placeholder="editingID ? '留空则不修改口令' : '可选'" /></el-form-item>
                </template>
                <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入备注信息" /></el-form-item>
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
    display: flex;
    flex-direction: column;
    min-height: 100%;
    padding: 12px 16px 16px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-card);
    box-shadow: var(--shadow);
}

.page-toolbar {
    margin-bottom: 8px;
}

.credential-search {
    width: 260px;
    margin-bottom: 8px;
}

.credential-form :deep(.el-form-item) {
    margin-bottom: 16px;
}

.credential-form :deep(.el-form-item__label) {
    color: var(--text);
    font-size: 12px;
    font-weight: 600;
}

.credential-form :deep(.el-input__wrapper),
.credential-form :deep(.el-textarea__inner) {
    box-shadow: 0 0 0 1px var(--line) inset;
}

.credential-form :deep(.el-radio) {
    margin-right: 28px;
    font-size: 12px;
}

:deep(.credential-type) {
    display: inline-flex;
    align-items: center;
    padding: 3px 9px;
    border: 1px solid #ffe2b9;
    border-radius: 8px;
    color: #e6a23c;
    background: #fdf6ec;
    line-height: 1.25;
}

:deep(.credential-type.is-key) {
    border-color: #d9f0d0;
    color: #67c23a;
    background: #f0f9eb;
}

:deep(.credential-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center
}

:deep(.credential-action) {
    padding: 4px 9px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: #fff;
    font-size: 11px;
    box-shadow: 0 1px 3px rgba(24, 34, 53, .08)
}

:deep(.credential-action.edit) {
    border-color: var(--primary);
    color: var(--primary);
    background: #f7f8ff
}

:deep(.credential-action.delete) {
    border-color: #ffd5d5;
    color: #ef6262;
    background: #fff7f7
}
</style>
