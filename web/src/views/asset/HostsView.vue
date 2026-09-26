<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, render, watch, type Component } from 'vue'; import { ElMessage, ElMessageBox, ElTooltip } from 'element-plus'; import { ArrowLeft, ArrowRight, CollectionTag, Cpu, Delete, Download, Edit, Folder, Monitor, Refresh, Search, Upload } from '@element-plus/icons-vue'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { collectHostInfo, createCredential, createHost, deleteHost, downloadHostImportTemplate, exportHosts, importHosts, listAssetGroups, listCloudAccounts, listCredentials, listHosts, testHostConnection, type AssetGroup, type CloudAccount, type Credential, type Host, updateHost } from '@/api/asset'
import { useAuthStore } from '@/stores/auth'
import { downloadAttachment } from '@/utils/download'
interface AssetTreeNode { id: number; label: string; count: number; children: AssetTreeNode[] }
const rows = ref<Host[]>([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const groups = ref<AssetGroup[]>([]); const credentials = ref<Credential[]>([]); const cloudAccounts = ref<CloudAccount[]>([]); const keyword = ref(''); const statusFilter = ref<number | null>(null); const treeKeyword = ref(''); const selectedGroupID = ref<number | null>(null); const treePaneWidth = ref(200); const treePaneDragging = ref(false); const treePaneStartX = ref(0); const treePaneStartWidth = ref(200); const assetTreeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>(); const dialogVisible = ref(false); const credentialDialogVisible = ref(false); const testingConnection = ref(false); const editingID = ref<number | null>(null); const tagDraft = ref(''); const credentialForm = reactive({ name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); const form = reactive<{ name: string; group_id: number | null; host_type: string; ip: string; port: number; credential_id: number | null; cloud_account_id: number | null; tags: string; description: string; status: number }>({ name: '', group_id: null, host_type: '', ip: '', port: 22, credential_id: null, cloud_account_id: null, tags: '', description: '', status: -1 })
const auth = useAuthStore()
const excelInput = ref<HTMLInputElement>()
const excelBusy = ref(false)
const selectedHosts = ref<Host[]>([])
const tagItems = computed(() => form.tags.split(',').map(tag => tag.trim()).filter(Boolean))
const hostTypeLabels: Record<string, string> = { self: '自建主机', cloud: '云主机' }
const treeData = computed<AssetTreeNode[]>(() => { const counts = new Map<number, number>(); rows.value.forEach(row => counts.set(row.group_id, (counts.get(row.group_id) || 0) + 1)); const nodes = new Map<number, AssetTreeNode>(); const roots: AssetTreeNode[] = []; groups.value.forEach(group => nodes.set(group.id, { id: group.id, label: group.name, count: counts.get(group.id) || 0, children: [] })); groups.value.forEach(group => { const node = nodes.get(group.id); const parent = nodes.get(group.parent_id); if (!node) return; if (group.parent_id && parent) parent.children.push(node); else roots.push(node) }); return roots })
function componentIcon(icon: Component, className = 'el-icon') { const root = document.createElement('span'); root.className = className; render(h(icon), root); return root }
function hostRenderer(p: any) { const wrap = document.createElement('div'); wrap.className = `host-summary status-${p.data.status}`; const icon = componentIcon(Monitor, 'host-summary-icon el-icon'); const text = document.createElement('span'); text.className = 'host-summary-text'; const name = document.createElement('strong'); name.textContent = p.data.name || '-'; const address = document.createElement('small'); address.textContent = `${p.data.ip || '-'}:${p.data.port || 22}`; text.append(name, address); wrap.append(icon, text); return wrap }
function typeRenderer(p: any) { const tag = document.createElement('span'); tag.className = `host-type is-${p.value || 'unknown'}`; tag.textContent = hostTypeLabels[String(p.value || '')] || '未知'; return tag }
function statusRenderer(p: any) { const tag = document.createElement('span'); tag.className = `host-status is-${p.value}`; tag.textContent = p.value === 1 ? '在线' : p.value === 0 ? '离线' : '未知'; return tag }
function tagsRenderer(p: any) { const wrap = document.createElement('div'); wrap.className = 'host-tags'; const tags = String(p.value || '').split(',').map((tag: string) => tag.trim()).filter(Boolean); if (!tags.length) { wrap.textContent = '-'; return wrap }; tags.slice(0, 2).forEach((value: string) => { const tag = document.createElement('span'); tag.textContent = value; wrap.append(tag) }); if (tags.length > 2) { const more = document.createElement('small'); more.textContent = `+${tags.length - 2}`; wrap.append(more) }; return wrap }
function actionIcon(title: string, kind: string, icon: Component, fn: () => void) { const root = document.createElement('span'); render(h(ElTooltip, { content: title, placement: 'top', showAfter: 250 }, { default: () => h('button', { type: 'button', 'aria-label': title, class: `host-action-icon ${kind}`, onClick: fn }, [h('span', { class: 'el-icon' }, [h(icon)])]) }), root); return root }
function openTerminal() { window.open('/assets/terminal', '_blank', 'noopener,noreferrer') }
async function collectHost(row: Host) { await collectHostInfo(row.id); ElMessage.success(`${row.name} 信息采集成功`); await load() }
function openFiles(row: Host) { ElMessage.info(`正在准备打开 ${row.name} 的文件管理`) }
function actionRenderer(p: any) { const wrap = document.createElement('div'); wrap.className = 'host-actions'; if (auth.hasPermission('POST /api/v1/assets/hosts/collect')) wrap.append(actionIcon('采集信息', 'collect', Refresh, () => { void collectHost(p.data) })); wrap.append(actionIcon('文件管理', 'files', Folder, () => openFiles(p.data))); if (auth.hasPermission('POST /api/v1/assets/hosts/update')) wrap.append(actionIcon('编辑', 'edit', Edit, () => edit(p.data))); if (auth.hasPermission('POST /api/v1/assets/hosts/delete')) wrap.append(actionIcon('删除', 'delete', Delete, () => remove(p.data))); return wrap }
function formatBytes(value: number) { if (!value) return '-'; const units = ['B', 'K', 'M', 'G', 'T']; let size = value; let index = 0; while (size >= 1024 && index < units.length - 1) { size /= 1024; index++ }; return `${size.toFixed(index > 2 ? 1 : 0)}${units[index]}` }
function metricRenderer(kind: 'cpu' | 'memory' | 'disk') { return (p: any) => { const row = p.data as Host; const usage = kind === 'cpu' ? row.cpu_usage : kind === 'memory' ? row.memory_usage : row.disk_usage; const headline = kind === 'cpu' ? (row.cpu_cores ? `${row.cpu_cores}核` : '-') : `${formatBytes(kind === 'memory' ? row.memory_used : row.disk_used)} / ${formatBytes(kind === 'memory' ? row.memory_total : row.disk_total)}`; const wrap = document.createElement('div'); wrap.className = `host-metric${usage >= 70 ? ' is-warning' : ''}`; const text = document.createElement('span'); text.textContent = headline; const details = document.createElement('div'); const bar = document.createElement('i'); const fill = document.createElement('b'); fill.style.width = `${Math.max(0, Math.min(100, usage || 0))}%`; bar.append(fill); const percent = document.createElement('small'); percent.textContent = usage ? `${usage.toFixed(1)}%` : '-'; details.append(bar, percent); wrap.append(text, details); return wrap } }
function systemRenderer(p: any) { const wrap = document.createElement('div'); wrap.className = 'host-system'; const os = document.createElement('span'); os.append(componentIcon(Monitor)); const osText = document.createElement('em'); osText.textContent = p.data.os || '-'; os.append(osText); const arch = document.createElement('small'); arch.append(componentIcon(Cpu)); const archText = document.createElement('em'); archText.textContent = p.data.arch || '-'; arch.append(archText); wrap.append(os, arch); return wrap }
const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((page.value - 1) * pageSize.value + (p.node?.rowIndex ?? 0) + 1), width: 58 }, { field: 'name', headerName: '主机', minWidth: 190, cellRenderer: hostRenderer }, { field: 'status', headerName: '状态', width: 86, cellRenderer: statusRenderer }, { field: 'host_type', headerName: '类型', width: 90, cellRenderer: typeRenderer }, { headerName: 'CPU', width: 132, cellRenderer: metricRenderer('cpu') }, { headerName: '内存', width: 160, cellRenderer: metricRenderer('memory') }, { headerName: '磁盘', width: 160, cellRenderer: metricRenderer('disk') }, { field: 'tags', headerName: '标签', width: 115, cellRenderer: tagsRenderer }, { headerName: '系统信息', minWidth: 170, flex: 1, cellRenderer: systemRenderer }, { headerName: '操作', pinned: 'right', width: 170, cellRenderer: actionRenderer }])
async function load(reset = false) { if (reset) page.value = 1; const { data } = await listHosts(keyword.value, selectedGroupID.value || 0, page.value, pageSize.value, statusFilter.value); rows.value = data.data.list; total.value = data.data.total }
function changePage(value: number) { page.value = value; void load() }
function changePageSize(value: number) { pageSize.value = value; page.value = 1; void load() }
async function loadFormOptions() { const [groupResult, credentialResult, cloudResult] = await Promise.all([listAssetGroups('', 1, 1000), listCredentials('', 1, 1000), listCloudAccounts('', 1, 1000)]); groups.value = groupResult.data.data.list; credentials.value = credentialResult.data.data.list; cloudAccounts.value = cloudResult.data.data.list }
function selectTreeNode(node: AssetTreeNode) { selectedGroupID.value = node.id; void load(true) }
function filterTreeNode(value: string, data: AssetTreeNode) { return !value || data.label.toLowerCase().includes(value.toLowerCase()) }
function startTreePaneDrag(event: PointerEvent) { if (event.button !== 0) return; treePaneDragging.value = true; treePaneStartX.value = event.clientX; treePaneStartWidth.value = treePaneWidth.value; document.addEventListener('pointermove', moveTreePane); document.addEventListener('pointerup', stopTreePaneDrag) }
function moveTreePane(event: PointerEvent) { if (treePaneDragging.value) treePaneWidth.value = Math.max(0, Math.min(360, treePaneStartWidth.value + event.clientX - treePaneStartX.value)) }
function stopTreePaneDrag() { if (!treePaneDragging.value) return; treePaneDragging.value = false; if (treePaneWidth.value < 90) treePaneWidth.value = 0; document.removeEventListener('pointermove', moveTreePane); document.removeEventListener('pointerup', stopTreePaneDrag) }
function toggleTreePane() { treePaneWidth.value = treePaneWidth.value === 0 ? 200 : 0 }
function create() { editingID.value = null; tagDraft.value = ''; Object.assign(form, { name: '', group_id: selectedGroupID.value, host_type: '', ip: '', port: 22, credential_id: null, cloud_account_id: null, tags: '', description: '', status: -1 }); dialogVisible.value = true }
function edit(row: Host) { editingID.value = row.id; tagDraft.value = ''; Object.assign(form, { ...row, group_id: row.group_id || null, credential_id: row.credential_id || null, cloud_account_id: row.cloud_account_id || null }); dialogVisible.value = true }
function addTag() { const tag = tagDraft.value.trim(); if (tag && !tagItems.value.includes(tag)) form.tags = [...tagItems.value, tag].join(','); tagDraft.value = '' }
function removeTag(tag: string) { form.tags = tagItems.value.filter(item => item !== tag).join(',') }
function openCredentialCreate() { Object.assign(credentialForm, { name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); credentialDialogVisible.value = true }
async function saveCredential() { if (!credentialForm.name.trim()) return ElMessage.warning('请填写凭证名称'); if (credentialForm.credential_type === 'password' && !credentialForm.password) return ElMessage.warning('请填写密码'); if (credentialForm.credential_type === 'key' && !credentialForm.private_key.trim()) return ElMessage.warning('请填写私钥'); const { data } = await createCredential(credentialForm); await loadFormOptions(); form.credential_id = data.data.id; credentialDialogVisible.value = false; ElMessage.success('凭证创建成功') }
async function testConnection() { if (!form.ip || !form.port || !form.credential_id) return ElMessage.warning('请填写主机地址、端口并选择凭据'); testingConnection.value = true; try { await testHostConnection({ ip: form.ip, port: form.port, credential_id: form.credential_id }); ElMessage.success('SSH 连接成功') } finally { testingConnection.value = false } }
async function save() { if (!form.name || !form.ip || !form.host_type || !form.credential_id) return ElMessage.warning('请填写主机名称、类型、地址并选择凭据'); const payload = { ...form, group_id: form.group_id ?? 0, credential_id: form.credential_id, cloud_account_id: form.cloud_account_id ?? 0 }; if (editingID.value) await updateHost({ ...payload, id: editingID.value }); else await createHost(payload); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: Host) { await ElMessageBox.confirm(`确定删除主机 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteHost(row.id); ElMessage.success('删除成功'); await load() }
async function batchRemove(selectedRows: unknown[]) {
    const selectedHosts = selectedRows as Host[]
    if (!selectedHosts.length) return
    await ElMessageBox.confirm(`确定删除选中的 ${selectedHosts.length} 台主机？`, '批量删除确认', { type: 'warning' })
    await Promise.all(selectedHosts.map(host => deleteHost(host.id)))
    ElMessage.success('批量删除成功')
    await load()
}
async function exportExcel() {
    if (!auth.hasPermission('POST /api/v1/assets/hosts/export')) return
    excelBusy.value = true
    try {
        const response = await exportHosts(keyword.value, selectedGroupID.value || 0, statusFilter.value, selectedHosts.value.map(host => host.id))
        downloadAttachment(response.data, response.headers['content-disposition'], '主机管理.xlsx')
        ElMessage.success('主机 Excel 已导出')
    } finally { excelBusy.value = false }
}
async function handleExcelCommand(command: string) {
    if (command === 'template') {
        if (!auth.hasPermission('POST /api/v1/assets/hosts/template')) return
        excelBusy.value = true
        try {
            const response = await downloadHostImportTemplate()
            downloadAttachment(response.data, response.headers['content-disposition'], '主机导入模板.xlsx')
        } finally { excelBusy.value = false }
    } else if (command === 'import') excelInput.value?.click()
}
async function importExcel(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    if (!auth.hasPermission('POST /api/v1/assets/hosts/import')) {
        input.value = ''
        return
    }
    excelBusy.value = true
    try {
        const { data } = await importHosts(file)
        await load()
        ElMessage.success(`导入完成：新增 ${data.data.created} 台，更新 ${data.data.updated} 台`)
    } finally {
        excelBusy.value = false
        input.value = ''
    }
}
watch(treeKeyword, value => assetTreeRef.value?.filter(value))
watch(statusFilter, () => { void load(true) })
watch(() => form.host_type, type => { if (type !== 'cloud') form.cloud_account_id = null })
onMounted(() => { void load(); void loadFormOptions() })
</script>
<template>
    <div class="page-container host-page">
        <div class="host-layout" :style="{ gridTemplateColumns: `${treePaneWidth}px 2px minmax(0, 1fr)` }">
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
            <div class="content-divider" :class="{ dragging: treePaneDragging }" role="separator" aria-label="调整资产分组树宽度"
                @pointerdown="startTreePaneDrag" @dblclick="toggleTreePane">
                <button class="content-divider-handle" type="button" :aria-label="treePaneWidth ? '收起资产分组树' : '展开资产分组树'"
                    @pointerdown.stop @click.stop="toggleTreePane"><el-icon>
                        <component :is="treePaneWidth ? ArrowLeft : ArrowRight" />
                    </el-icon></button>
            </div>
            <section class="section-card host-content">
                <div class="page-toolbar">
                    <div class="host-filters"><el-input v-model="keyword" class="search-input" placeholder="搜索主机名/IP..."
                            clearable :prefix-icon="Search" @input="load(true)" /><el-select v-model="statusFilter"
                            class="status-filter" placeholder="主机状态" clearable><el-option label="在线"
                                :value="1" /><el-option label="离线" :value="0" /><el-option label="未知"
                                :value="-1" /></el-select></div>
                    <div class="toolbar-actions"><el-button :icon="Monitor"
                            @click="openTerminal">终端</el-button><el-dropdown
                            v-if="auth.hasPermission('POST /api/v1/assets/hosts/export') || auth.hasPermission('POST /api/v1/assets/hosts/import') || auth.hasPermission('POST /api/v1/assets/hosts/template')"
                            split-button class="host-excel-dropdown" :loading="excelBusy"
                            :disabled="!auth.hasPermission('POST /api/v1/assets/hosts/export')" @click="exportExcel"
                            @command="handleExcelCommand">导出 Excel
                            <template #dropdown><el-dropdown-menu>
                                    <el-dropdown-item v-if="auth.hasPermission('POST /api/v1/assets/hosts/template')"
                                        command="template" :icon="Download">下载 Excel
                                        导入模板</el-dropdown-item>
                                    <el-dropdown-item v-if="auth.hasPermission('POST /api/v1/assets/hosts/import')"
                                        command="import" :icon="Upload">从 Excel 导入</el-dropdown-item>
                                </el-dropdown-menu></template>
                        </el-dropdown><input ref="excelInput" class="host-excel-input" type="file" accept=".xlsx" hidden
                            @change="importExcel" /><el-button type="primary" class="primary-btn"
                            @click="create">新增主机</el-button>
                    </div>
                </div>
                <AppDataGrid class="host-grid" :row-data="rows" :column-defs="columns" pagination
                    :bulk-delete="auth.hasPermission('POST /api/v1/assets/hosts/delete') ? batchRemove : undefined"
                    server-pagination :total="total" :current-page="page" :page-size="pageSize"
                    @page-change="changePage" @page-size-change="changePageSize"
                    @selection-changed="selectedHosts = $event as Host[]" />
            </section>
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑主机' : '新增主机'" width="60%">
            <el-form :model="form" label-width="92px" class="host-form">
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="主机名称" required><el-input v-model="form.name"
                                placeholder="请输入主机名称" /></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="主机类型" required><el-select v-model="form.host_type"
                                placeholder="请选择主机类型" clearable><el-option label="自建主机" value="self" /><el-option
                                    label="云主机" value="cloud" /></el-select></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="所属分组"><el-select v-model="form.group_id"
                                placeholder="请选择分组" clearable><el-option v-for="group in groups" :key="group.id"
                                    :label="group.name" :value="group.id" /></el-select></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="云账号"
                            :class="{ 'is-disabled': form.host_type !== 'cloud' }"><el-select
                                v-model="form.cloud_account_id"
                                :placeholder="form.host_type === 'cloud' ? '请选择云账号' : ''" clearable
                                :disabled="form.host_type !== 'cloud'"><el-option v-for="account in cloudAccounts"
                                    :key="account.id" :label="account.name"
                                    :value="account.id" /></el-select></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="主机地址" required><el-input v-model="form.ip"
                                placeholder="请输入 IP 地址" /></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="SSH 端口" required><el-input-number v-model="form.port"
                                :min="1" :max="65535" controls-position="right" /></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="认证凭据" required>
                            <div class="credential-picker"><el-select v-model="form.credential_id" placeholder="选择凭据"
                                    clearable><el-option v-for="credential in credentials" :key="credential.id"
                                        :label="credential.name" :value="credential.id" /></el-select><el-button
                                    v-permission="'POST /api/v1/assets/credentials/create'" text type="primary"
                                    class="credential-create-button" @click="openCredentialCreate">新建凭据</el-button>
                            </div>
                        </el-form-item></el-col>
                </el-row>
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
                    v-permission="'POST /api/v1/assets/hosts/test'" :loading="testingConnection"
                    @click="testConnection">测试连接</el-button><el-button type="primary"
                    @click="save">保存</el-button></template>
        </el-dialog>
        <el-dialog v-model="credentialDialogVisible" title="新建凭据" width="460px" append-to-body>
            <el-form :model="credentialForm" label-width="82px" class="credential-form">
                <el-form-item label="凭证名称" required><el-input v-model="credentialForm.name"
                        placeholder="请输入凭证名称" /></el-form-item>
                <el-form-item label="认证方式" required><el-radio-group v-model="credentialForm.credential_type"><el-radio
                            value="password">密码</el-radio><el-radio
                            value="key">密钥</el-radio></el-radio-group></el-form-item>
                <el-form-item label="用户名"><el-input v-model="credentialForm.username"
                        placeholder="如：root" /></el-form-item>
                <el-form-item v-if="credentialForm.credential_type === 'password'" label="密码" required><el-input
                        v-model="credentialForm.password" type="password" show-password /></el-form-item>
                <el-form-item v-else label="私钥" required><el-input v-model="credentialForm.private_key" type="textarea"
                        :rows="3" placeholder="请输入私钥内容" /></el-form-item>
                <el-form-item v-if="credentialForm.credential_type === 'key'" label="密钥口令"><el-input
                        v-model="credentialForm.passphrase" type="password" show-password /></el-form-item>
            </el-form>
            <template #footer><el-button @click="credentialDialogVisible = false">取消</el-button><el-button
                    type="primary" @click="saveCredential">创建并选择</el-button></template>
        </el-dialog>
    </div>
</template>
<style scoped>
.host-page,
.host-layout,
.host-content {
    min-width: 0;
    min-height: 0;
}

.host-layout {
    display: grid;
    gap: 0;
    height: 100%;
    min-height: 100%;
}

.asset-tree-panel,
.section-card {
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-card);
    box-shadow: var(--shadow);
}

.asset-tree-panel {
    min-width: 0;
    overflow: hidden;
    padding: 12px 8px;
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
    color: var(--primary);
    font-weight: 700;
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
    font-size: 15px;
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
    font-weight: 400;
}

.asset-tree-panel :deep(.is-current > .el-tree-node__content .asset-tree-label) {
    color: var(--text);
    font-weight: 600;
}

.section-card {
    display: flex;
    flex-direction: column;
    min-height: 100%;
    padding: 12px 16px 0;
}

.page-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
}

.search-input {
    width: 156px;
    flex: none;
}

.host-filters,
.toolbar-actions {
    display: flex;
    align-items: center;
    gap: 6px;
}

.host-filters {
    flex: 1;
    min-width: 0;
}

.toolbar-actions :deep(.el-button + .el-button) {
    margin-left: 0;
}

.host-excel-dropdown :deep(.el-button-group > .el-button:first-child) {
    border-radius: 8px 0 0 8px;
}

.host-excel-dropdown :deep(.el-button-group > .el-button:last-child) {
    border-radius: 0 8px 8px 0;
}

.status-filter {
    width: 120px;
    flex: none;
}

.host-grid {
    flex: 1;
    min-height: 0;
}

.host-grid :deep(.app-data-grid-shell) {
    height: 100%;
}

.content-divider {
    position: relative;
    width: 2px;
    cursor: col-resize;
    touch-action: none;
    background: transparent;
}

.content-divider::before {
    position: absolute;
    z-index: 1;
    top: 0;
    bottom: 0;
    left: -6px;
    width: 13px;
    content: '';
}

.content-divider::after {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 2px;
    height: 100%;
    content: '';
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
    display: flex;
    align-items: center;
    justify-content: center;
    width: 15px;
    height: 20px;
    padding: 0;
    transform: translate(-50%, -50%);
    border: 1px solid var(--line);
    border-radius: 4px;
    background: #fff;
    color: var(--text-soft);
    box-shadow: 0 2px 5px rgba(31, 42, 68, .12);
    cursor: pointer;
}

.content-divider-handle:hover {
    border-color: var(--primary);
    color: var(--primary);
}

.content-divider-handle :deep(.el-icon) {
    font-size: 10px;
}

.host-form :deep(.el-form-item) {
    margin-bottom: 12px;
}

.host-form :deep(.el-form-item__label) {
    color: var(--text);
    font-size: 12px;
    font-weight: 600;
}

.host-form :deep(.el-input__wrapper),
.host-form :deep(.el-textarea__inner) {
    box-shadow: 0 0 0 1px var(--line) inset;
}

.host-form :deep(.el-input-number) {
    width: 100%;
}

.host-form :deep(.el-form-item.is-disabled .el-form-item__label) {
    color: var(--text-muted);
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

:deep(.host-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 5px;
    align-items: center;
    justify-content: flex-start;
    padding-left: 12px
}

:deep(.host-summary) {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 100%;
}

:deep(.host-summary-icon) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    flex: none;
    border-radius: 9px;
    background: #f1f3f7;
    color: var(--text-muted);
    font-size: 18px;
}

:deep(.host-summary-icon svg) {
    width: 18px;
    height: 18px;
}

:deep(.host-summary.status-1 .host-summary-icon) {
    background: #eaf8e8;
    color: #58b83d;
}

:deep(.host-summary-text) {
    display: flex;
    min-width: 0;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    line-height: 1.35;
}

:deep(.host-summary-text strong) {
    overflow: hidden;
    color: var(--text);
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.host-summary-text small) {
    color: var(--text-muted);
    font-size: 11px;
}

:deep(.host-status) {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    color: var(--text-muted);
}

:deep(.host-status::before) {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
    content: '';
    box-shadow: 0 0 0 2px color-mix(in srgb, currentColor 18%, transparent);
}

:deep(.host-status.is-1) {
    color: #58b83d;
}

:deep(.host-status.is-0) {
    color: #ef6262;
}

:deep(.host-type) {
    display: inline-flex;
    align-items: center;
    height: 24px;
    padding: 0 8px;
    border: 1px solid var(--line);
    border-radius: 7px;
    background: #f6f7f9;
    color: var(--text-soft);
}

:deep(.host-type.is-cloud) {
    border-color: #f4d6a0;
    background: #fff9ef;
    color: #d79222;
}

:deep(.host-tags) {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 100%;
    color: var(--text-muted);
}

:deep(.host-tags span),
:deep(.host-tags small) {
    display: inline-flex;
    align-items: center;
    height: 22px;
    padding: 0 7px;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: #f7f8fa;
    color: var(--text-soft);
    font-size: 10px;
}

:deep(.host-metric) {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 5px;
    height: 100%;
    color: var(--text-soft);
    font-size: 11px;
    line-height: 1.2;
}

:deep(.host-metric > span) {
    font-weight: 600;
    white-space: nowrap;
}

:deep(.host-metric > div) {
    display: grid;
    grid-template-columns: minmax(42px, 1fr) 34px;
    align-items: center;
    gap: 6px;
}

:deep(.host-metric i) {
    display: block;
    height: 5px;
    overflow: hidden;
    border-radius: 3px;
    background: #eceef5;
}

:deep(.host-metric b) {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--primary);
}

:deep(.host-metric.is-warning b) {
    background: #e5a331;
}

:deep(.host-metric small) {
    color: var(--text-muted);
    font-size: 9px;
}

:deep(.host-system) {
    display: flex;
    min-width: 0;
    flex-direction: column;
    justify-content: center;
    gap: 5px;
    height: 100%;
    line-height: 1.35;
}

:deep(.host-system > span),
:deep(.host-system > small) {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 5px;
    color: var(--text-soft);
}

:deep(.host-system .el-icon) {
    width: 13px;
    height: 13px;
    flex: none;
    color: var(--text-muted);
}

:deep(.host-system em) {
    overflow: hidden;
    font-style: normal;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.host-system small) {
    color: var(--text-muted);
    font-size: 10px;
}

:deep(.host-action-icon) {
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

:deep(.host-action-icon .el-icon),
:deep(.host-action-icon svg) {
    width: 16px;
    height: 16px;
}

:deep(.host-action-icon:hover) {
    border-color: var(--line);
    background: #f6f7fa;
    color: var(--primary);
}

:deep(.host-action-icon.delete:hover) {
    border-color: #ffd5d5;
    background: #fff7f7;
    color: #ef6262;
}

@media (max-width: 900px) {
    .host-layout {
        grid-template-columns: 1fr !important;
    }

    .asset-tree-panel {
        max-height: 220px;
        overflow: auto;
    }

    .content-divider {
        display: none;
    }

    .search-input {
        width: min(260px, 60vw);
    }
}
</style>
