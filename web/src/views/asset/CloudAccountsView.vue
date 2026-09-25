<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createCloudAccount, deleteCloudAccount, getCloudProviderOptions, listCloudAccounts, updateCloudAccount, type CloudAccount, type CloudProviderOption } from '@/api/asset'
const rows = ref<CloudAccount[]>([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const keyword = ref(''); const dialogVisible = ref(false); const editingID = ref<number | null>(null); const form = reactive({ name: '', provider: '', access_key: '', secret_key: '', region: '', description: '', status: true })
const providerOptions = ref<CloudProviderOption[]>([])
const regionOptions = computed(() => providerOptions.value.find((provider) => provider.value === form.provider)?.regions ?? [])
function action(text: string, kind: string, fn: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `cloud-action ${kind}`; el.onclick = fn; return el }
const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '账号名称', minWidth: 130 }, { field: 'provider', headerName: '云服务商', width: 120, cellRenderer: (p: { value: string }) => `<span class="cloud-provider">${({ aliyun: '阿里云', tencent: '腾讯云', aws: 'AWS', huawei: '华为云' } as Record<string, string>)[p.value] ?? p.value}</span>` }, { field: 'region', headerName: '默认区域', width: 120 }, { field: 'description', headerName: '说明', minWidth: 160 }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => `<span class="grid-status${p.value ? '' : ' is-disabled'}">${p.value ? '启用' : '禁用'}</span>` }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'cloud-actions'; wrap.append(action('编辑', 'edit', () => edit(p.data)), action('删除', 'delete', () => remove(p.data))); return wrap } }])
function applyDefaultRegion() { if (!regionOptions.value.some((region) => region.value === form.region)) form.region = regionOptions.value[0]?.value ?? '' }
function formatRegionLabel(region: CloudProviderOption['regions'][number]) {
    return `${region.label.replace(/\s+(?=\d)/g, '').replace(/（([^）]+)）/g, '-$1')}（${region.value}）`
}
async function load(reset = false) { if (reset) page.value = 1; const { data } = await listCloudAccounts(keyword.value, page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total }
function changePage(value: number) { page.value = value; void load() }
function changePageSize(value: number) { pageSize.value = value; page.value = 1; void load() }
async function loadProviderOptions() { const { data } = await getCloudProviderOptions(); providerOptions.value = Array.isArray(data.data) ? data.data : []; applyDefaultRegion() }
function openCreate() { dialogVisible.value = true; editingID.value = null; Object.assign(form, { name: '', provider: '', access_key: '', secret_key: '', region: '', description: '', status: true }) }
function edit(row: CloudAccount) { editingID.value = row.id; Object.assign(form, { name: row.name, provider: row.provider, access_key: '', secret_key: '', region: row.region, description: row.description, status: row.status }); dialogVisible.value = true }
async function save() { if (!form.name.trim()) return ElMessage.warning('请填写账号名称'); if (!form.provider) return ElMessage.warning('请选择云服务商'); if (!form.region) return ElMessage.warning('请选择默认区域'); if (!editingID.value && (!form.access_key.trim() || !form.secret_key)) return ElMessage.warning('请填写 AccessKey 和 SecretKey'); if (editingID.value) await updateCloudAccount({ ...form, id: editingID.value }); else await createCloudAccount(form); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: CloudAccount) { await ElMessageBox.confirm(`确定删除云账号 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteCloudAccount(row.id); ElMessage.success('删除成功'); await load() }
async function batchRemove(selectedRows: unknown[]) {
    const selectedAccounts = selectedRows as CloudAccount[]
    if (!selectedAccounts.length) return
    await ElMessageBox.confirm(`确定删除选中的 ${selectedAccounts.length} 个云账号？`, '批量删除确认', { type: 'warning' })
    await Promise.all(selectedAccounts.map(account => deleteCloudAccount(account.id)))
    ElMessage.success('批量删除成功')
    await load()
}
watch(() => form.provider, applyDefaultRegion)
onMounted(async () => { await loadProviderOptions(); await load() })
</script>
<template>
    <div class="page-container table-shell">
        <div class="section-card">
            <div class="page-toolbar">
                <h2 class="page-title">云账号</h2>
                <div class="toolbar-actions"><el-button v-permission="'POST /api/v1/assets/cloud-accounts/create'"
                    type="primary" class="primary-btn" @click="openCreate">新增云账号</el-button></div>
            </div><el-input v-model="keyword" class="cloud-search" placeholder="搜索账号名称或云服务商" clearable
                @input="load(true)" />
            <AppDataGrid :row-data="rows" :column-defs="columns" pagination server-pagination :total="total"
                :current-page="page" :page-size="pageSize" :bulk-delete="batchRemove" @page-change="changePage"
                @page-size-change="changePageSize" />
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑云账号' : '新增云账号'" width="560px">
            <el-form :model="form" label-width="92px" class="cloud-form">
                <el-form-item label="账号名称" required><el-input v-model="form.name"
                        placeholder="请输入云账号名称，如：生产环境阿里云" /></el-form-item>
                <el-form-item label="云服务商" required><el-select v-model="form.provider" placeholder="请选择云服务商"
                        style="width: 100%"><el-option v-for="provider in providerOptions" :key="provider.value"
                            :label="provider.label" :value="provider.value" /></el-select></el-form-item>
                <el-form-item label="AccessKey" required><el-input v-model="form.access_key" autocomplete="off"
                        :placeholder="editingID ? '留空则不修改' : '请输入 AccessKey'" /></el-form-item>
                <el-form-item label="SecretKey" required><el-input v-model="form.secret_key" type="password"
                        show-password autocomplete="new-password"
                        :placeholder="editingID ? '留空则不修改' : '请输入 SecretKey'" /></el-form-item>
                <el-form-item label="默认区域" required><el-select v-model="form.region" placeholder="请选择默认区域"
                        style="width: 100%"><el-option v-for="region in regionOptions" :key="region.value"
                            :label="formatRegionLabel(region)" :value="region.value" /></el-select></el-form-item>
                <el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item>
                <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3"
                        placeholder="请输入备注信息" /></el-form-item>
            </el-form>
            <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template>
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

.toolbar-actions {
    display: flex;
    align-items: center;
    gap: 6px;
}

.toolbar-actions :deep(.el-button + .el-button) {
    margin-left: 0;
}

.cloud-search {
    width: 260px;
    margin-bottom: 8px
}

.cloud-form :deep(.el-form-item) {
    margin-bottom: 16px;
}

.cloud-form :deep(.el-form-item__label) {
    color: var(--text);
    font-size: 12px;
    font-weight: 600;
}

.cloud-form :deep(.el-input__wrapper),
.cloud-form :deep(.el-textarea__inner) {
    box-shadow: 0 0 0 1px var(--line) inset;
}

:deep(.cloud-provider) {
    display: inline-flex;
    padding: 3px 9px;
    border: 1px solid #d9e7ff;
    border-radius: 8px;
    color: var(--primary);
    background: #f2f6ff;
    line-height: 1.25;
}

:deep(.cloud-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center
}

:deep(.cloud-action) {
    padding: 4px 9px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: #fff;
    font-size: 11px;
    box-shadow: 0 1px 3px rgba(24, 34, 53, .08)
}

:deep(.cloud-action.edit) {
    border-color: var(--primary);
    color: var(--primary);
    background: #f7f8ff
}

:deep(.cloud-action.delete) {
    border-color: #ffd5d5;
    color: #ef6262;
    background: #fff7f7
}
</style>
