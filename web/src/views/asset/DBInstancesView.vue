<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'; import { ElMessage, ElMessageBox } from 'element-plus'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { createDBInstance, deleteDBInstance, listDBInstances, type DbInstance, updateDBInstance } from '@/api/asset'
const rows = ref<DbInstance[]>([]); const keyword = ref(''); const dialogVisible = ref(false); const editingID = ref<number | null>(null); const form = reactive({ name: '', db_type: 'postgresql', host: '', port: 5432, username: '', password: '', version: '', credential_id: 0, tags: '', description: '', status: true })
function action(text: string, kind: string, fn: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `db-action ${kind}`; el.onclick = fn; return el }; const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '实例名称', minWidth: 130 }, { field: 'db_type', headerName: '类型', width: 110 }, { field: 'host', headerName: '地址', minWidth: 140 }, { field: 'port', headerName: '端口', width: 80 }, { field: 'username', headerName: '用户名', width: 110 }, { field: 'version', headerName: '版本', width: 100 }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => `<span class="grid-status${p.value ? '' : ' is-disabled'}">${p.value ? '启用' : '禁用'}</span>` }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'db-actions'; wrap.append(action('编辑', 'edit', () => edit(p.data)), action('删除', 'delete', () => remove(p.data))); return wrap } }])
async function load() { const { data } = await listDBInstances(keyword.value); rows.value = data.data } function create() { editingID.value = null; Object.assign(form, { name: '', db_type: 'postgresql', host: '', port: 5432, username: '', password: '', version: '', credential_id: 0, tags: '', description: '', status: true }); dialogVisible.value = true } function edit(row: DbInstance) { editingID.value = row.id; Object.assign(form, { ...row, password: '' }); dialogVisible.value = true } async function save() { if (!form.name || !form.host) return ElMessage.warning('请填写实例名称和地址'); if (editingID.value) await updateDBInstance({ ...form, id: editingID.value }); else await createDBInstance(form); dialogVisible.value = false; ElMessage.success('保存成功'); await load() } async function remove(row: DbInstance) { await ElMessageBox.confirm(`确定删除数据库实例 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteDBInstance(row.id); ElMessage.success('删除成功'); await load() } onMounted(load)
</script>
<template>
    <div class="page-container table-shell">
        <div class="section-card">
            <div class="page-toolbar">
                <h2 class="page-title">数据库实例</h2><el-button type="primary" class="primary-btn"
                    @click="create">新增实例</el-button>
            </div><el-input v-model="keyword" class="db-search" placeholder="搜索实例名称或地址" clearable @input="load" />
            <AppDataGrid :row-data="rows" :column-defs="columns" pagination />
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑数据库实例' : '新增数据库实例'" width="560px">
            <el-form :model="form" label-width="92px" class="db-form">
                <el-form-item label="实例名称" required><el-input v-model="form.name" placeholder="请输入实例名称，如：生产 MySQL 主库" /></el-form-item>
                <el-form-item label="数据库类型" required><el-select v-model="form.db_type" style="width: 100%"><el-option label="PostgreSQL" value="postgresql" /><el-option label="MySQL" value="mysql" /><el-option label="Redis" value="redis" /></el-select></el-form-item>
                <el-form-item label="数据库地址" required><el-input v-model="form.host" placeholder="IP 地址或主机名" /></el-form-item>
                <el-form-item label="端口"><el-input-number v-model="form.port" :min="1" :max="65535" controls-position="right" /></el-form-item>
                <el-form-item label="用户名"><el-input v-model="form.username" placeholder="请输入用户名" /></el-form-item>
                <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password :placeholder="editingID ? '留空则不修改密码' : '请输入密码'" /></el-form-item>
                <el-form-item label="版本"><el-input v-model="form.version" placeholder="如：16.2" /></el-form-item>
                <el-form-item label="凭证 ID"><el-input-number v-model="form.credential_id" :min="0" controls-position="right" /></el-form-item>
                <el-form-item label="标签"><el-input v-model="form.tags" placeholder="多个标签以逗号分隔" /></el-form-item>
                <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入备注信息" /></el-form-item>
            </el-form>
            <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
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

    :deep(.db-actions) {
        position: absolute;
        inset: 0;
        display: flex;
        gap: 8px;
        align-items: center;
        justify-content: center
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
