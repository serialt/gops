<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'; import { ElMessage, ElMessageBox } from 'element-plus'; import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { createCredential, createHost, deleteHost, listAssetGroups, listCloudAccounts, listCredentials, listHosts, type AssetGroup, type CloudAccount, type Credential, type Host, updateHost } from '@/api/asset'
interface AssetTreeNode { id: number; label: string; children: AssetTreeNode[] }
const rows = ref<Host[]>([]); const groups = ref<AssetGroup[]>([]); const credentials = ref<Credential[]>([]); const cloudAccounts = ref<CloudAccount[]>([]); const keyword = ref(''); const treeKeyword = ref(''); const selectedGroupID = ref<number | null>(null); const treePaneWidth = ref(200); const treePaneDragging = ref(false); const treePaneStartX = ref(0); const treePaneStartWidth = ref(200); const assetTreeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>(); const dialogVisible = ref(false); const credentialDialogVisible = ref(false); const editingID = ref<number | null>(null); const tagDraft = ref(''); const credentialForm = reactive({ name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); const form = reactive<{ name: string; group_id: number | null; host_type: string; ip: string; port: number; ssh_user: string; credential_id: number | null; cloud_account_id: number | null; tags: string; description: string; status: number }>({ name: '', group_id: null, host_type: '', ip: '', port: 22, ssh_user: 'root', credential_id: null, cloud_account_id: null, tags: '', description: '', status: -1 })
const tagItems = computed(() => form.tags.split(',').map(tag => tag.trim()).filter(Boolean))
const treeData = computed<AssetTreeNode[]>(() => { const nodes = new Map<number, AssetTreeNode>(); const roots: AssetTreeNode[] = []; groups.value.forEach(group => nodes.set(group.id, { id: group.id, label: group.name, children: [] })); groups.value.forEach(group => { const node = nodes.get(group.id); const parent = nodes.get(group.parent_id); if (!node) return; if (group.parent_id && parent) parent.children.push(node); else roots.push(node) }); return roots })
function action(text: string, kind: string, fn: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `host-action ${kind}`; el.onclick = fn; return el }; const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 }, { field: 'name', headerName: '主机名称', minWidth: 130 }, { field: 'ip', headerName: '地址', minWidth: 140 }, { field: 'port', headerName: '端口', width: 80 }, { field: 'ssh_user', headerName: 'SSH 用户', width: 110 }, { field: 'host_type', headerName: '类型', width: 100 }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => `<span class="grid-status${p.value === 1 ? '' : ' is-disabled'}">${p.value === 1 ? '在线' : p.value === 0 ? '离线' : '未知'}</span>` }, { field: 'tags', headerName: '标签', minWidth: 120 }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'host-actions'; wrap.append(action('编辑', 'edit', () => edit(p.data)), action('删除', 'delete', () => remove(p.data))); return wrap } }])
async function load() { const { data } = await listHosts(keyword.value, selectedGroupID.value ?? 0); rows.value = data.data }
async function loadFormOptions() { const [groupResult, credentialResult, cloudResult] = await Promise.all([listAssetGroups(), listCredentials(), listCloudAccounts()]); groups.value = groupResult.data.data; credentials.value = credentialResult.data.data; cloudAccounts.value = cloudResult.data.data }
function selectTreeNode(node: AssetTreeNode) { selectedGroupID.value = node.id }
function filterTreeNode(value: string, data: AssetTreeNode) { return !value || data.label.toLowerCase().includes(value.toLowerCase()) }
function startTreePaneDrag(event: PointerEvent) { if (event.button !== 0) return; treePaneDragging.value = true; treePaneStartX.value = event.clientX; treePaneStartWidth.value = treePaneWidth.value; document.addEventListener('pointermove', moveTreePane); document.addEventListener('pointerup', stopTreePaneDrag) }
function moveTreePane(event: PointerEvent) { if (treePaneDragging.value) treePaneWidth.value = Math.max(0, Math.min(360, treePaneStartWidth.value + event.clientX - treePaneStartX.value)) }
function stopTreePaneDrag() { if (!treePaneDragging.value) return; treePaneDragging.value = false; if (treePaneWidth.value < 90) treePaneWidth.value = 0; document.removeEventListener('pointermove', moveTreePane); document.removeEventListener('pointerup', stopTreePaneDrag) }
function toggleTreePane() { treePaneWidth.value = treePaneWidth.value === 0 ? 200 : 0 }
function create() { editingID.value = null; tagDraft.value = ''; Object.assign(form, { name: '', group_id: selectedGroupID.value, host_type: '', ip: '', port: 22, ssh_user: 'root', credential_id: null, cloud_account_id: null, tags: '', description: '', status: -1 }); dialogVisible.value = true }
function edit(row: Host) { editingID.value = row.id; tagDraft.value = ''; Object.assign(form, { ...row, group_id: row.group_id || null, credential_id: row.credential_id || null, cloud_account_id: row.cloud_account_id || null }); dialogVisible.value = true }
function addTag() { const tag = tagDraft.value.trim(); if (tag && !tagItems.value.includes(tag)) form.tags = [...tagItems.value, tag].join(','); tagDraft.value = '' }
function removeTag(tag: string) { form.tags = tagItems.value.filter(item => item !== tag).join(',') }
function openCredentialCreate() { Object.assign(credentialForm, { name: '', credential_type: 'password', username: '', password: '', private_key: '', passphrase: '', description: '' }); credentialDialogVisible.value = true }
async function saveCredential() { if (!credentialForm.name.trim()) return ElMessage.warning('请填写凭证名称'); if (credentialForm.credential_type === 'password' && !credentialForm.password) return ElMessage.warning('请填写密码'); if (credentialForm.credential_type === 'key' && !credentialForm.private_key.trim()) return ElMessage.warning('请填写私钥'); const { data } = await createCredential(credentialForm); await loadFormOptions(); form.credential_id = data.data.id; credentialDialogVisible.value = false; ElMessage.success('凭证创建成功') }
async function save() { if (!form.name || !form.ip || !form.host_type) return ElMessage.warning('请填写主机名称、类型和地址'); const payload = { ...form, group_id: form.group_id ?? 0, credential_id: form.credential_id ?? 0, cloud_account_id: form.cloud_account_id ?? 0 }; if (editingID.value) await updateHost({ ...payload, id: editingID.value }); else await createHost(payload); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: Host) { await ElMessageBox.confirm(`确定删除主机 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteHost(row.id); ElMessage.success('删除成功'); await load() }
watch(treeKeyword, value => assetTreeRef.value?.filter(value))
watch(selectedGroupID, () => { void load() })
watch(() => form.host_type, type => { if (type !== 'cloud') form.cloud_account_id = null })
onMounted(() => { void load(); void loadFormOptions() })
</script>
<template>
    <div class="page-container host-page">
        <div class="host-layout" :style="{ gridTemplateColumns: `${treePaneWidth}px 2px minmax(0, 1fr)` }">
            <aside class="asset-tree-panel">
                <div class="tree-title"><span>资产树</span><el-input v-model="treeKeyword" class="tree-search" placeholder="搜索" clearable /></div>
                <el-tree ref="assetTreeRef" :data="treeData" node-key="id" highlight-current default-expand-all :current-node-key="selectedGroupID" :filter-node-method="filterTreeNode" :expand-on-click-node="false" @node-click="selectTreeNode" />
            </aside>
            <div class="content-divider" :class="{ dragging: treePaneDragging }" role="separator" aria-label="调整资产分组树宽度" @pointerdown="startTreePaneDrag" @dblclick="toggleTreePane">
                <button class="content-divider-handle" type="button" :aria-label="treePaneWidth ? '收起资产分组树' : '展开资产分组树'" @pointerdown.stop @click.stop="toggleTreePane"><el-icon><component :is="treePaneWidth ? ArrowLeft : ArrowRight" /></el-icon></button>
            </div>
            <section class="section-card host-content">
                <div class="page-toolbar">
                    <el-input v-model="keyword" class="search-input" placeholder="搜索主机名称或地址" clearable @input="load" />
                    <el-button type="primary" class="primary-btn" @click="create">新增主机</el-button>
                </div>
                <AppDataGrid class="host-grid" :row-data="rows" :column-defs="columns" pagination />
            </section>
        </div><el-dialog v-model="dialogVisible" :title="editingID ? '编辑主机' : '新增主机'" width="60%">
            <el-form :model="form" label-width="92px" class="host-form">
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="主机名称" required><el-input v-model="form.name" placeholder="请输入主机名称" /></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="主机类型" required><el-select v-model="form.host_type" placeholder="请选择主机类型" clearable><el-option label="自建主机" value="self" /><el-option label="云主机" value="cloud" /></el-select></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="所属分组"><el-select v-model="form.group_id" placeholder="请选择分组" clearable><el-option v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" /></el-select></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="云账号" :class="{ 'is-disabled': form.host_type !== 'cloud' }"><el-select v-model="form.cloud_account_id" :placeholder="form.host_type === 'cloud' ? '请选择云账号' : ''" clearable :disabled="form.host_type !== 'cloud'"><el-option v-for="account in cloudAccounts" :key="account.id" :label="account.name" :value="account.id" /></el-select></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="主机地址" required><el-input v-model="form.ip" placeholder="请输入 IP 地址" /></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="SSH 端口" required><el-input-number v-model="form.port" :min="1" :max="65535" controls-position="right" /></el-form-item></el-col>
                </el-row>
                <el-row :gutter="16">
                    <el-col :xs="24" :sm="12"><el-form-item label="SSH 用户" required><el-input v-model="form.ssh_user" placeholder="如：root" /></el-form-item></el-col>
                    <el-col :xs="24" :sm="12"><el-form-item label="认证凭据"><div class="credential-picker"><el-select v-model="form.credential_id" placeholder="选择凭据" clearable><el-option v-for="credential in credentials" :key="credential.id" :label="credential.name" :value="credential.id" /></el-select><el-button v-permission="'POST /api/v1/assets/credentials/create'" text type="primary" class="credential-create-button" @click="openCredentialCreate">新建凭据</el-button></div></el-form-item></el-col>
                </el-row>
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
<style scoped>
.host-page, .host-layout, .host-content { min-width: 0; min-height: 0; }
.host-layout { display: grid; gap: 0; height: 100%; min-height: 100%; }
.asset-tree-panel, .section-card { border: 1px solid var(--line); border-radius: 8px; background: var(--bg-card); box-shadow: var(--shadow); }
.asset-tree-panel { min-width: 0; overflow: hidden; padding: 12px 8px; }
.tree-title { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 0 8px 10px; margin-bottom: 8px; border-bottom: 1px solid var(--line); color: var(--text); font-size: 13px; font-weight: 700; white-space: nowrap; }
.tree-search { width: 92px; }
.tree-search :deep(.el-input__wrapper) { min-height: 28px; border-radius: 6px; box-shadow: 0 0 0 1px var(--line) inset; }
.asset-tree-panel :deep(.el-tree) { --el-tree-node-hover-bg-color: var(--primary-soft); background: transparent; font-size: 11px; }
.asset-tree-panel :deep(.el-tree-node__content) { height: 32px; border-radius: 6px; }
.asset-tree-panel :deep(.is-current > .el-tree-node__content) { color: var(--primary); font-weight: 700; background: var(--primary-soft); }
.section-card { display: flex; flex-direction: column; min-height: 100%; padding: 12px 16px 16px; }
.page-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 8px; }
.search-input { width: 260px; }
.host-grid { flex: 1; min-height: 0; }
.host-grid :deep(.app-data-grid-shell) { height: 100%; }
.content-divider { position: relative; width: 2px; cursor: col-resize; touch-action: none; background: transparent; }
.content-divider::before { position: absolute; z-index: 1; top: 0; bottom: 0; left: -6px; width: 13px; content: ''; }
.content-divider::after { position: absolute; top: 0; bottom: 0; left: 0; width: 2px; height: 100%; content: ''; background: var(--line); }
.content-divider:hover::after, .content-divider.dragging::after { background: var(--primary); }
.content-divider-handle { position: absolute; z-index: 2; top: 50%; left: 50%; display: flex; align-items: center; justify-content: center; width: 15px; height: 20px; padding: 0; transform: translate(-50%, -50%); border: 1px solid var(--line); border-radius: 4px; background: #fff; color: var(--text-soft); box-shadow: 0 2px 5px rgba(31, 42, 68, .12); cursor: pointer; }
.content-divider-handle:hover { border-color: var(--primary); color: var(--primary); }
.content-divider-handle :deep(.el-icon) { font-size: 10px; }
.host-form :deep(.el-form-item) { margin-bottom: 12px; }
.host-form :deep(.el-form-item__label) { color: var(--text); font-size: 12px; font-weight: 600; }
.host-form :deep(.el-input__wrapper), .host-form :deep(.el-textarea__inner) { box-shadow: 0 0 0 1px var(--line) inset; }
.host-form :deep(.el-input-number) { width: 100%; }
.host-form :deep(.el-form-item.is-disabled .el-form-item__label) { color: var(--text-muted); }
.credential-picker { display: flex; align-items: center; gap: 6px; width: 100%; }
.credential-picker :deep(.el-select) { flex: 1; min-width: 0; }
.credential-picker :deep(.credential-create-button) { flex: none; min-width: auto; padding: 4px 8px; border: 1px solid var(--primary); background: var(--primary) !important; color: #fff !important; font-size: 11px; font-weight: 600; box-shadow: none; }
.credential-picker :deep(.credential-create-button:hover) { background: var(--primary-strong) !important; border-color: var(--primary-strong); transform: none; }
.tag-input { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; width: 100%; min-height: 32px; padding: 3px 8px; border: 1px solid var(--line); border-radius: 6px; }
.tag-input :deep(.el-tag) { border-radius: 5px; }
.tag-input :deep(.el-input) { flex: 1; min-width: 140px; }
.tag-input :deep(.el-input__wrapper) { box-shadow: none; padding: 0; }

:deep(.host-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: flex-start;
    padding-left: 12px
}

:deep(.host-action) {
    padding: 4px 9px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: #fff;
    font-size: 11px;
    box-shadow: 0 1px 3px rgba(24, 34, 53, .08)
}

:deep(.host-action.edit) {
    border-color: var(--primary);
    color: var(--primary);
    background: #f7f8ff
}

:deep(.host-action.delete) {
    border-color: #ffd5d5;
    color: #ef6262;
    background: #fff7f7
}
@media (max-width: 900px) { .host-layout { grid-template-columns: 1fr !important; } .asset-tree-panel { max-height: 220px; overflow: auto; } .content-divider { display: none; } .search-input { width: min(260px, 60vw); } }
</style>
