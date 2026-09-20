<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'; import { ElMessage, ElMessageBox } from 'element-plus'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { createCredential, createDBInstance, deleteDBInstance, getDatabaseTypeOptions, listCredentials, listDBInstances, type Credential, type DatabaseTypeOption, type DbInstance, updateDBInstance } from '@/api/asset'
const rows = ref<DbInstance[]>([]); const credentials = ref<Credential[]>([]); const databaseTypes = ref<DatabaseTypeOption[]>([]); const keyword = ref(''); const dialogVisible = ref(false); const credentialDialogVisible = ref(false); const editingID = ref<number | null>(null); const tagDraft = ref(''); const credentialForm = reactive({ name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); const form = reactive<{ name: string; db_type: string; host: string; port: number | null; username: string; password: string; version: string; credential_id: number | null; tags: string; description: string; status: boolean }>({ name: '', db_type: '', host: '', port: null, username: '', password: '', version: '', credential_id: null, tags: '', description: '', status: true })
const tagItems = computed(() => form.tags.split(',').map(tag => tag.trim()).filter(Boolean))
function action(text: string, kind: string, fn: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `db-action ${kind}`; el.onclick = fn; return el }; const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '实例名称', minWidth: 130 }, { field: 'db_type', headerName: '类型', width: 110, valueFormatter: p => databaseTypes.value.find(item => item.value === p.value)?.label ?? p.value }, { field: 'host', headerName: '地址', minWidth: 140 }, { field: 'port', headerName: '端口', width: 80, valueFormatter: p => p.value ? String(p.value) : '—' }, { field: 'username', headerName: '用户名', width: 110 }, { field: 'version', headerName: '版本', width: 100 }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => `<span class="grid-status${p.value ? '' : ' is-disabled'}">${p.value ? '启用' : '禁用'}</span>` }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'db-actions'; wrap.append(action('编辑', 'edit', () => edit(p.data)), action('删除', 'delete', () => remove(p.data))); return wrap } }])
async function load() { const { data } = await listDBInstances(keyword.value); rows.value = data.data }
async function loadCredentials() { const { data } = await listCredentials(); credentials.value = data.data }
async function loadDatabaseTypes() { const { data } = await getDatabaseTypeOptions(); databaseTypes.value = data.data }
function create() { editingID.value = null; tagDraft.value = ''; Object.assign(form, { name: '', db_type: '', host: '', port: null, username: '', password: '', version: '', credential_id: null, tags: '', description: '', status: true }); dialogVisible.value = true }
function edit(row: DbInstance) { editingID.value = row.id; tagDraft.value = ''; Object.assign(form, { ...row, password: '', credential_id: row.credential_id || null }); dialogVisible.value = true }
function addTag() { const tag = tagDraft.value.trim(); if (tag && !tagItems.value.includes(tag)) form.tags = [...tagItems.value, tag].join(','); tagDraft.value = '' }
function removeTag(tag: string) { form.tags = tagItems.value.filter(item => item !== tag).join(',') }
function openCredentialCreate() { Object.assign(credentialForm, { name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); credentialDialogVisible.value = true }
async function saveCredential() { if (!credentialForm.name.trim()) return ElMessage.warning('请填写凭证名称'); if (credentialForm.credential_type === 'password' && !credentialForm.password) return ElMessage.warning('请填写密码'); if (credentialForm.credential_type === 'key' && !credentialForm.private_key.trim()) return ElMessage.warning('请填写私钥'); const { data } = await createCredential(credentialForm); await loadCredentials(); form.credential_id = data.data.id; credentialDialogVisible.value = false; ElMessage.success('凭证创建成功') }
function applyDefaultPort(type: string) { const option = databaseTypes.value.find(item => item.value === type); if (option) form.port = option.default_port }
async function save() { if (!form.name || !form.db_type || !form.host || !form.port) return ElMessage.warning('请填写数据库名称、类型、地址和端口'); const payload = { ...form, port: form.port, credential_id: form.credential_id ?? 0 }; if (editingID.value) await updateDBInstance({ ...payload, id: editingID.value }); else await createDBInstance(payload); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: DbInstance) { await ElMessageBox.confirm(`确定删除数据库 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteDBInstance(row.id); ElMessage.success('删除成功'); await load() }
onMounted(() => { void load(); void loadCredentials(); void loadDatabaseTypes() })
</script>
<template>
    <div class="page-container table-shell">
        <div class="section-card">
            <div class="page-toolbar">
                <h2 class="page-title">数据库管理</h2><el-button type="primary" class="primary-btn"
                    @click="create">新增数据库</el-button>
            </div><el-input v-model="keyword" class="db-search" placeholder="搜索实例名称或地址" clearable @input="load" />
            <AppDataGrid :row-data="rows" :column-defs="columns" pagination />
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑数据库' : '新增数据库'" width="60%">
            <el-form :model="form" label-width="92px" class="db-form">
                <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="数据库名称" required><el-input v-model="form.name" placeholder="请输入数据库名称" /></el-form-item></el-col><el-col :xs="24" :sm="12"><el-form-item label="数据库类型" required><el-select v-model="form.db_type" placeholder="请选择数据库类型" clearable @change="applyDefaultPort"><el-option v-for="databaseType in databaseTypes" :key="databaseType.value" :label="databaseType.label" :value="databaseType.value" /></el-select></el-form-item></el-col></el-row>
                <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="数据库地址" required><el-input v-model="form.host" placeholder="IP 地址或主机名" /></el-form-item></el-col><el-col :xs="24" :sm="12"><el-form-item label="端口"><el-input-number v-model="form.port" placeholder="请选择类型" :min="1" :max="65535" controls-position="right" :disabled="!form.db_type" /></el-form-item></el-col></el-row>
                <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="用户名"><el-input v-model="form.username" placeholder="请输入用户名" /></el-form-item></el-col><el-col :xs="24" :sm="12"><el-form-item label="密码"><el-input v-model="form.password" type="password" show-password :placeholder="editingID ? '留空则不修改密码' : '请输入密码'" /></el-form-item></el-col></el-row>
                <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="版本"><el-input v-model="form.version" placeholder="如：16.2" /></el-form-item></el-col><el-col :xs="24" :sm="12"><el-form-item label="认证凭据"><div class="credential-picker"><el-select v-model="form.credential_id" placeholder="选择凭据" clearable><el-option v-for="credential in credentials" :key="credential.id" :label="credential.name" :value="credential.id" /></el-select><el-button v-permission="'POST /api/v1/assets/credentials/create'" text type="primary" class="credential-create-button" @click="openCredentialCreate">新建凭据</el-button></div></el-form-item></el-col></el-row>
                <el-form-item label="标签"><div class="tag-input"><el-tag v-for="tag in tagItems" :key="tag" closable @close="removeTag(tag)">{{ tag }}</el-tag><el-input v-model="tagDraft" placeholder="输入后按回车添加标签" @keyup.enter.prevent="addTag" @blur="addTag" /></div></el-form-item>
                <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入备注信息" /></el-form-item>
            </el-form>
            <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
        </el-dialog>
        <el-dialog v-model="credentialDialogVisible" title="新建凭据" width="460px" append-to-body>
            <el-form :model="credentialForm" label-width="82px" class="credential-form">
                <el-form-item label="凭证名称" required><el-input v-model="credentialForm.name" placeholder="请输入凭证名称" /></el-form-item>
                <el-form-item label="认证方式" required><el-radio-group v-model="credentialForm.credential_type"><el-radio value="password">密码</el-radio><el-radio value="key">密钥</el-radio></el-radio-group></el-form-item>
                <el-form-item label="用户名"><el-input v-model="credentialForm.username" placeholder="如：root" /></el-form-item>
                <el-form-item v-if="credentialForm.credential_type === 'password'" label="密码" required><el-input v-model="credentialForm.password" type="password" show-password /></el-form-item>
                <el-form-item v-else label="私钥" required><el-input v-model="credentialForm.private_key" type="textarea" :rows="3" placeholder="请输入私钥内容" /></el-form-item>
                <el-form-item v-if="credentialForm.credential_type === 'key'" label="密钥口令"><el-input v-model="credentialForm.passphrase" type="password" show-password /></el-form-item>
            </el-form>
            <template #footer><el-button @click="credentialDialogVisible = false">取消</el-button><el-button type="primary" @click="saveCredential">创建并选择</el-button></template>
        </el-dialog>
    </div>
</template>
<style
    scoped>
    .db-search {
        width: 260px;
        margin-bottom: 8px
    }

    .table-shell { background: var(--bg); min-height: 100%; }
    .section-card { display: flex; flex-direction: column; min-height: 100%; padding: 12px 16px 16px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg-card); box-shadow: var(--shadow); }
    .page-toolbar { margin-bottom: 8px; }
    .db-form :deep(.el-form-item) { margin-bottom: 16px; }
    .db-form :deep(.el-form-item__label) { color: var(--text); font-size: 12px; font-weight: 600; }
    .db-form :deep(.el-input__wrapper), .db-form :deep(.el-textarea__inner) { box-shadow: 0 0 0 1px var(--line) inset; }
    .db-form :deep(.el-input-number) { width: 100%; }
    .credential-picker { display: flex; align-items: center; gap: 6px; width: 100%; }
    .credential-picker :deep(.el-select) { flex: 1; min-width: 0; }
    .credential-picker :deep(.credential-create-button) { flex: none; min-width: auto; padding: 4px 8px; border: 1px solid var(--primary); background: var(--primary) !important; color: #fff !important; font-size: 11px; font-weight: 600; box-shadow: none; }
    .credential-picker :deep(.credential-create-button:hover) { background: var(--primary-strong) !important; border-color: var(--primary-strong); transform: none; }
    .tag-input { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; width: 100%; min-height: 32px; padding: 3px 8px; border: 1px solid var(--line); border-radius: 6px; }
    .tag-input :deep(.el-tag) { border-radius: 5px; }
    .tag-input :deep(.el-input) { flex: 1; min-width: 140px; }
    .tag-input :deep(.el-input__wrapper) { box-shadow: none; padding: 0; }

    :deep(.db-actions) {
        position: absolute;
        inset: 0;
        display: flex;
        gap: 8px;
        align-items: center;
        justify-content: flex-start;
        padding-left: 12px
    }

    :deep(.db-action) {
        padding: 4px 9px;
        border: 1px solid transparent;
        border-radius: 8px;
        background: #fff;
        font-size: 11px;
        box-shadow: 0 1px 3px rgba(24, 34, 53, .08)
    }

    :deep(.db-action.edit) {
        border-color: var(--primary);
        color: var(--primary);
        background: #f7f8ff
    }

    :deep(.db-action.delete) {
        border-color: #ffd5d5;
        color: #ef6262;
        background: #fff7f7
    }
</style>
