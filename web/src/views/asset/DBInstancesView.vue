<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, render, watch, type Component } from 'vue'; import { ElMessage, ElMessageBox, ElTooltip } from 'element-plus'; import { ArrowLeft, ArrowRight, CollectionTag, Delete, Edit, Monitor, Refresh, Search } from '@element-plus/icons-vue'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { collectDBInstanceInfo, createCredential, createDBInstance, deleteDBInstance, getDatabaseTypeOptions, listAssetGroups, listCredentials, listDBInstances, testDBInstanceConnection, type AssetGroup, type Credential, type DatabaseTypeOption, type DbInstance, updateDBInstance } from '@/api/asset'
import { useAuthStore } from '@/stores/auth'
interface AssetTreeNode { id: number; label: string; count: number; children: AssetTreeNode[] }
const rows = ref<DbInstance[]>([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const groups = ref<AssetGroup[]>([]); const credentials = ref<Credential[]>([]); const databaseTypes = ref<DatabaseTypeOption[]>([]); const keyword = ref(''); const typeFilter = ref(''); const treeKeyword = ref(''); const selectedGroupID = ref<number | null>(null); const treePaneWidth = ref(200); const treePaneDragging = ref(false); const treePaneStartX = ref(0); const treePaneStartWidth = ref(200); const assetTreeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>(); const dialogVisible = ref(false); const credentialDialogVisible = ref(false); const testingConnection = ref(false); const editingID = ref<number | null>(null); const tagDraft = ref(''); const credentialForm = reactive({ name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); const form = reactive<{ group_id: number | null; name: string; db_type: string; host: string; port: number | null; version: string; credential_id: number | null; tags: string; description: string; status: boolean }>({ group_id: null, name: '', db_type: '', host: '', port: null, version: '', credential_id: null, tags: '', description: '', status: true })
const auth = useAuthStore()
const tagItems = computed(() => form.tags.split(',').map(tag => tag.trim()).filter(Boolean))
const visibleRows = computed(() => rows.value.filter(row => (!selectedGroupID.value || row.group_id === selectedGroupID.value) && (!typeFilter.value || row.db_type === typeFilter.value)))
const treeData = computed<AssetTreeNode[]>(() => { const counts = new Map<number, number>(); rows.value.forEach(row => counts.set(row.group_id, (counts.get(row.group_id) || 0) + 1)); const nodes = new Map<number, AssetTreeNode>(); const roots: AssetTreeNode[] = []; groups.value.forEach(group => nodes.set(group.id, { id: group.id, label: group.name, count: counts.get(group.id) || 0, children: [] })); groups.value.forEach(group => { const node = nodes.get(group.id); const parent = nodes.get(group.parent_id); if (!node) return; if (group.parent_id && parent) parent.children.push(node); else roots.push(node) }); return roots })
function actionIcon(title: string, kind: string, icon: Component, fn: () => void) { const root = document.createElement('span'); render(h(ElTooltip, { content: title, placement: 'top', showAfter: 250 }, { default: () => h('button', { type: 'button', 'aria-label': title, class: `db-action-icon ${kind}`, onClick: fn }, [h('span', { class: 'el-icon' }, [h(icon)])]) }), root); return root }
function actionRenderer(p: any) { const wrap = document.createElement('div'); wrap.className = 'db-actions'; if (auth.hasPermission('POST /api/v1/assets/db-instances/collect')) wrap.append(actionIcon('采集信息', 'collect', Refresh, () => { void collect(p.data) })); if (auth.hasPermission('POST /api/v1/assets/db-instances/update')) wrap.append(actionIcon('编辑', 'edit', Edit, () => edit(p.data))); if (auth.hasPermission('POST /api/v1/assets/db-instances/delete')) wrap.append(actionIcon('删除', 'delete', Delete, () => remove(p.data))); return wrap }
function badgeRenderer(kind: string, value: string) { const tag = document.createElement('span'); tag.className = `db-badge ${kind}`; tag.textContent = value || '-'; return tag }
const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '实例名称', minWidth: 130 }, { field: 'db_type', headerName: '类型', width: 110, cellRenderer: (p: any) => badgeRenderer('type', databaseTypes.value.find(item => item.value === p.value)?.label ?? p.value) }, { field: 'host', headerName: '地址', minWidth: 140 }, { field: 'port', headerName: '端口', width: 80, valueFormatter: p => p.value ? String(p.value) : '—' }, { field: 'version', headerName: '版本', width: 150, cellRenderer: (p: any) => badgeRenderer('version', p.value) }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => `<span class="grid-status${p.value ? '' : ' is-disabled'}">${p.value ? '在线' : '离线'}</span>` }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: actionRenderer }])
async function load(reset = false) { if (reset) page.value = 1; const { data } = await listDBInstances(keyword.value, selectedGroupID.value || 0, page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total }
function changePage(value: number) { page.value = value; void load() }
function changePageSize(value: number) { pageSize.value = value; page.value = 1; void load() }
async function loadCredentials() { const { data } = await listCredentials('', 1, 1000); credentials.value = data.data.list }
async function loadDatabaseTypes() { const { data } = await getDatabaseTypeOptions(); databaseTypes.value = data.data }
function create() { editingID.value = null; tagDraft.value = ''; Object.assign(form, { group_id: selectedGroupID.value, name: '', db_type: '', host: '', port: null, version: '', credential_id: null, tags: '', description: '', status: true }); dialogVisible.value = true }
function edit(row: DbInstance) { editingID.value = row.id; tagDraft.value = ''; Object.assign(form, { ...row, group_id: row.group_id || null, credential_id: row.credential_id || null }); dialogVisible.value = true }
function addTag() { const tag = tagDraft.value.trim(); if (tag && !tagItems.value.includes(tag)) form.tags = [...tagItems.value, tag].join(','); tagDraft.value = '' }
function removeTag(tag: string) { form.tags = tagItems.value.filter(item => item !== tag).join(',') }
function openCredentialCreate() { Object.assign(credentialForm, { name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); credentialDialogVisible.value = true }
async function saveCredential() { if (!credentialForm.name.trim()) return ElMessage.warning('请填写凭证名称'); if (credentialForm.credential_type === 'password' && !credentialForm.password) return ElMessage.warning('请填写密码'); if (credentialForm.credential_type === 'key' && !credentialForm.private_key.trim()) return ElMessage.warning('请填写私钥'); const { data } = await createCredential(credentialForm); await loadCredentials(); form.credential_id = data.data.id; credentialDialogVisible.value = false; ElMessage.success('凭证创建成功') }
function applyDefaultPort(type: string) { const option = databaseTypes.value.find(item => item.value === type); if (option) form.port = option.default_port }
async function save() { if (!form.name || !form.db_type || !form.host || !form.port || !form.credential_id) return ElMessage.warning('请填写数据库名称、类型、地址、端口并选择凭据'); const payload = { ...form, group_id: form.group_id ?? 0, port: form.port, credential_id: form.credential_id }; if (editingID.value) await updateDBInstance({ ...payload, id: editingID.value }); else await createDBInstance(payload); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function testConnection() { if (!form.db_type || !form.host || !form.port || !form.credential_id) return ElMessage.warning('请填写数据库类型、地址、端口并选择凭据'); testingConnection.value = true; try { const { data } = await testDBInstanceConnection({ db_type: form.db_type, host: form.host, port: form.port, credential_id: form.credential_id }); if (data.data.version) form.version = data.data.version; ElMessage.success(data.data.version ? `连接成功，版本：${data.data.version}` : '连接成功') } finally { testingConnection.value = false } }
async function remove(row: DbInstance) { await ElMessageBox.confirm(`确定删除数据库 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteDBInstance(row.id); ElMessage.success('删除成功'); await load() }
async function batchRemove(selectedRows: unknown[]) {
    const selectedInstances = selectedRows as DbInstance[]
    if (!selectedInstances.length) return
    await ElMessageBox.confirm(`确定删除选中的 ${selectedInstances.length} 个数据库实例？`, '批量删除确认', { type: 'warning' })
    await Promise.all(selectedInstances.map(instance => deleteDBInstance(instance.id)))
    ElMessage.success('批量删除成功')
    await load()
}
async function collect(row: DbInstance) { await collectDBInstanceInfo(row.id); ElMessage.success(`${row.name} 状态和版本采集成功`); await load() }
function openTerminal() { window.open('/assets/terminal', '_blank', 'noopener,noreferrer') }
function selectTreeNode(node: AssetTreeNode) { selectedGroupID.value = node.id; void load(true) }
function filterTreeNode(value: string, data: AssetTreeNode) { return !value || data.label.toLowerCase().includes(value.toLowerCase()) }
function startTreePaneDrag(event: PointerEvent) { if (event.button !== 0) return; treePaneDragging.value = true; treePaneStartX.value = event.clientX; treePaneStartWidth.value = treePaneWidth.value; document.addEventListener('pointermove', moveTreePane); document.addEventListener('pointerup', stopTreePaneDrag) }
function moveTreePane(event: PointerEvent) { if (treePaneDragging.value) treePaneWidth.value = Math.max(0, Math.min(360, treePaneStartWidth.value + event.clientX - treePaneStartX.value)) }
function stopTreePaneDrag() { treePaneDragging.value = false; if (treePaneWidth.value < 90) treePaneWidth.value = 0; document.removeEventListener('pointermove', moveTreePane); document.removeEventListener('pointerup', stopTreePaneDrag) }
function toggleTreePane() { treePaneWidth.value = treePaneWidth.value === 0 ? 200 : 0 }
watch(treeKeyword, value => assetTreeRef.value?.filter(value))
onMounted(async () => { await Promise.all([load(), loadCredentials(), loadDatabaseTypes(), listAssetGroups('', 1, 1000).then(({ data }) => { groups.value = data.data.list })]) })
</script>
<template>
    <div class="page-container db-page">
        <div class="db-layout" :style="{ gridTemplateColumns: `${treePaneWidth}px 2px minmax(0, 1fr)` }">
            <aside class="asset-tree-panel">
                <div class="tree-title"><span>资产树</span><el-input v-model="treeKeyword" class="tree-search"
                        placeholder="搜索" clearable /></div>
                <el-tree ref="assetTreeRef" :data="treeData" node-key="id" highlight-current default-expand-all
                    :current-node-key="selectedGroupID" :filter-node-method="filterTreeNode"
                    :expand-on-click-node="false" @node-click="selectTreeNode">
                    <template #default="{ data }"><span class="asset-tree-node"><el-icon>
                                <CollectionTag />
                            </el-icon><span class="asset-tree-label">{{ data.label }}</span><span
                                class="asset-tree-count">({{ data.count }})</span></span></template>
                </el-tree>
            </aside>
            <div class="content-divider" :class="{ dragging: treePaneDragging }" role="separator"
                @pointerdown="startTreePaneDrag" @dblclick="toggleTreePane"><button class="content-divider-handle"
                    type="button" @pointerdown.stop @click.stop="toggleTreePane"><el-icon>
                        <component :is="treePaneWidth ? ArrowLeft : ArrowRight" />
                    </el-icon></button></div>
            <div class="section-card">
                <div class="page-toolbar">
                    <div class="db-filters"><el-input v-model="keyword" class="db-search" placeholder="搜索数据库名称/IP..."
                            clearable :prefix-icon="Search" @input="load(true)" /><el-select v-model="typeFilter"
                            class="type-filter" placeholder="数据库类型" clearable><el-option v-for="item in databaseTypes"
                                :key="item.value" :label="item.label" :value="item.value" /></el-select></div>
                    <div class="toolbar-actions"><el-button :icon="Monitor"
                            @click="openTerminal">终端</el-button><el-button type="primary" class="primary-btn"
                            @click="create">新增数据库</el-button></div>
                </div>
                <AppDataGrid class="db-grid" :row-data="visibleRows" :column-defs="columns" pagination server-pagination
                    :bulk-delete="auth.hasPermission('POST /api/v1/assets/db-instances/delete') ? batchRemove : undefined"
                    :total="total" :current-page="page" :page-size="pageSize" @page-change="changePage"
                    @page-size-change="changePageSize" />
            </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑数据库' : '新增数据库'" width="60%">
                <el-form :model="form" label-width="92px" class="db-form">
                    <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="所属分组"><el-select
                                    v-model="form.group_id" placeholder="请选择分组" clearable><el-option
                                        v-for="group in groups" :key="group.id" :label="group.name"
                                        :value="group.id" /></el-select></el-form-item></el-col></el-row>
                    <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="数据库名称" required><el-input
                                    v-model="form.name" placeholder="请输入数据库名称" /></el-form-item></el-col><el-col
                            :xs="24" :sm="12"><el-form-item label="数据库类型" required><el-select v-model="form.db_type"
                                    placeholder="请选择数据库类型" clearable @change="applyDefaultPort"><el-option
                                        v-for="databaseType in databaseTypes" :key="databaseType.value"
                                        :label="databaseType.label"
                                        :value="databaseType.value" /></el-select></el-form-item></el-col></el-row>
                    <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="数据库地址" required><el-input
                                    v-model="form.host" placeholder="IP 地址或主机名" /></el-form-item></el-col><el-col
                            :xs="24" :sm="12"><el-form-item label="端口"><el-input-number v-model="form.port"
                                    placeholder="请选择类型" :min="1" :max="65535" controls-position="right"
                                    :disabled="!form.db_type" /></el-form-item></el-col></el-row>
                    <el-row :gutter="16"><el-col :xs="24" :sm="12"><el-form-item label="版本"><el-input
                                    v-model="form.version" placeholder="如：16.2" /></el-form-item></el-col><el-col
                            :xs="24" :sm="12"><el-form-item label="认证凭据">
                                <div class="credential-picker"><el-select v-model="form.credential_id"
                                        placeholder="选择凭据" clearable><el-option v-for="credential in credentials"
                                            :key="credential.id" :label="credential.name"
                                            :value="credential.id" /></el-select><el-button
                                        v-permission="'POST /api/v1/assets/credentials/create'" text type="primary"
                                        class="credential-create-button" @click="openCredentialCreate">新建凭据</el-button>
                                </div>
                            </el-form-item></el-col></el-row>
                    <el-form-item label="标签">
                        <div class="tag-input"><el-tag v-for="tag in tagItems" :key="tag" closable
                                @close="removeTag(tag)">{{ tag
                                }}</el-tag><el-input v-model="tagDraft" placeholder="输入后按回车添加标签"
                                @keyup.enter.prevent="addTag" @blur="addTag" /></div>
                    </el-form-item>
                    <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3"
                            placeholder="请输入备注信息" /></el-form-item>
                </el-form>
                <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button
                        v-permission="'POST /api/v1/assets/db-instances/test'" :loading="testingConnection"
                        @click="testConnection">测试连接</el-button><el-button type="primary"
                        @click="save">保存</el-button></template>
            </el-dialog>
            <el-dialog v-model="credentialDialogVisible" title="新建凭据" width="460px" append-to-body>
                <el-form :model="credentialForm" label-width="82px" class="credential-form">
                    <el-form-item label="凭证名称" required><el-input v-model="credentialForm.name"
                            placeholder="请输入凭证名称" /></el-form-item>
                    <el-form-item label="认证方式" required><el-radio-group
                            v-model="credentialForm.credential_type"><el-radio value="password">密码</el-radio><el-radio
                                value="key">密钥</el-radio></el-radio-group></el-form-item>
                    <el-form-item label="用户名"><el-input v-model="credentialForm.username"
                            placeholder="如：root" /></el-form-item>
                    <el-form-item v-if="credentialForm.credential_type === 'password'" label="密码" required><el-input
                            v-model="credentialForm.password" type="password" show-password /></el-form-item>
                    <el-form-item v-else label="私钥" required><el-input v-model="credentialForm.private_key"
                            type="textarea" :rows="3" placeholder="请输入私钥内容" /></el-form-item>
                    <el-form-item v-if="credentialForm.credential_type === 'key'" label="密钥口令"><el-input
                            v-model="credentialForm.passphrase" type="password" show-password /></el-form-item>
                </el-form>
                <template #footer><el-button @click="credentialDialogVisible = false">取消</el-button><el-button
                        type="primary" @click="saveCredential">创建并选择</el-button></template>
            </el-dialog>
        </div>
    </div>
</template>
<style scoped>
.db-page,
.db-layout,
.section-card {
    min-width: 0;
    min-height: 0;
}

.db-layout {
    display: grid;
    gap: 0;
    height: 100%;
    min-height: 100%;
}

.db-search {
    width: 156px;
}

.db-filters {
    display: flex;
    align-items: center;
    gap: 8px;
}

.toolbar-actions {
    display: flex;
    align-items: center;
    gap: 6px;
}

.toolbar-actions :deep(.el-button + .el-button) {
    margin-left: 0;
}

.type-filter {
    width: 120px;
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
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
}

.db-grid {
    flex: 1;
    min-height: 0;
}

.db-grid :deep(.app-data-grid-shell) {
    height: 100%;
}

.asset-tree-panel {
    min-width: 0;
    overflow: hidden;
    padding: 12px 8px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-card);
    box-shadow: var(--shadow);
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
}

.content-divider {
    position: relative;
    width: 2px;
    cursor: col-resize;
    touch-action: none;
}

.content-divider::after {
    position: absolute;
    inset: 0;
    width: 2px;
    background: var(--line);
    content: '';
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
    width: 15px;
    height: 20px;
    align-items: center;
    justify-content: center;
    padding: 0;
    transform: translate(-50%, -50%);
    border: 1px solid var(--line);
    border-radius: 4px;
    background: #fff;
    color: var(--text-soft);
    cursor: pointer;
}

.db-form :deep(.el-form-item) {
    margin-bottom: 16px;
}

.db-form :deep(.el-form-item__label) {
    color: var(--text);
    font-size: 12px;
    font-weight: 600;
}

.db-form :deep(.el-input__wrapper),
.db-form :deep(.el-textarea__inner) {
    box-shadow: 0 0 0 1px var(--line) inset;
}

.db-form :deep(.el-input-number) {
    width: 100%;
}

.credential-picker {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
}

.credential-picker :deep(.el-select) {
    flex: 1;
    min-width: 0;
}

.credential-picker :deep(.credential-create-button) {
    flex: none;
    min-width: auto;
    padding: 4px 8px;
    border: 1px solid var(--primary);
    background: var(--primary) !important;
    color: #fff !important;
    font-size: 11px;
    font-weight: 600;
    box-shadow: none;
}

.credential-picker :deep(.credential-create-button:hover) {
    background: var(--primary-strong) !important;
    border-color: var(--primary-strong);
    transform: none;
}

.tag-input {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    width: 100%;
    min-height: 32px;
    padding: 3px 8px;
    border: 1px solid var(--line);
    border-radius: 6px;
}

.tag-input :deep(.el-tag) {
    border-radius: 5px;
}

.tag-input :deep(.el-input) {
    flex: 1;
    min-width: 140px;
}

.tag-input :deep(.el-input__wrapper) {
    box-shadow: none;
    padding: 0;
}

:deep(.db-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: flex-start;
    padding-left: 12px
}

:deep(.db-action-icon) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--text-soft);
    cursor: pointer;
}

:deep(.db-action-icon .el-icon),
:deep(.db-action-icon svg) {
    width: 16px;
    height: 16px;
}

:deep(.db-action-icon:hover) {
    border-color: var(--line);
    background: #f6f7fa;
    color: var(--primary);
}

:deep(.db-action-icon.delete:hover) {
    border-color: #ffd5d5;
    background: #fff7f7;
    color: #ef6262;
}

:deep(.db-badge) {
    display: inline-flex;
    max-width: 100%;
    height: 23px;
    align-items: center;
    padding: 0 8px;
    overflow: hidden;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: #f7f8fa;
    color: var(--text-soft);
    font-size: 10px;
    line-height: 21px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.db-badge.type) {
    border-color: color-mix(in srgb, var(--primary) 24%, var(--line));
    background: var(--primary-soft);
    color: var(--primary);
    font-weight: 600;
}

:deep(.db-badge.version) {
    background: #fff;
    color: var(--text-soft);
    font-variant-numeric: tabular-nums;
}

@media (max-width: 900px) {
    .db-layout {
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
