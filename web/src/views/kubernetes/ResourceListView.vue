<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { ColDef } from 'ag-grid-community'
import { FolderOpened, Monitor, Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { getKubernetesClusterStats, listKubernetesClusters, listKubernetesNamespaces, listKubernetesNodes, listKubernetesPods, listKubernetesWorkloads, updateKubernetesNode, type KubernetesCluster, type KubernetesNode } from '@/api/kubernetes'
const route = useRoute(); const kind = computed(() => String(route.meta.kubernetesResource || 'nodes')); const title = computed(() => kind.value === 'nodes' ? '节点管理' : kind.value === 'namespaces' ? '命名空间' : kind.value === 'pods' ? 'Pod 管理' : '工作负载'); const clusters = ref<KubernetesCluster[]>([]), clusterID = ref(0), keyword = ref(''), rows = ref<any[]>([])
const statusFilter = ref(''), roleFilter = ref(''), namespaceStatus = ref(''), refreshing = ref(false)
const nodeDialog = ref(false), nodeSaving = ref(false), nodeForm = ref({ name: '', labels: '', schedulable: true })
const nodeStats = ref({ total: 0, healthy: 0, pods: 0, podCapacity: 0, cpu: 0 })
const namespaceStats = ref({ total: 0, healthy: 0, labels: 0, averageAge: '0d' })
const isNodeReady = (status: unknown) => ['ready', '就绪', '正常'].includes(String(status || '').toLowerCase())
function calculateNamespaceStats(items: any[]) {
    const now = Date.now()
    const totalAge = items.reduce((sum, item) => { const created = new Date(String(item.created_at || '').replace(' ', 'T')).getTime(); return sum + (Number.isFinite(created) ? Math.max(0, now - created) : 0) }, 0)
    namespaceStats.value = { total: items.length, healthy: items.filter(item => ['active', '正常'].includes(String(item.status || '').toLowerCase())).length, labels: items.reduce((sum, item) => sum + Object.keys(item.labels || {}).length, 0), averageAge: items.length ? `${Math.floor(totalAge / items.length / 86400000)}d` : '0d' }
}
async function loadClusters() { const { data } = await listKubernetesClusters(); clusters.value = data.data; if (!clusterID.value && clusters.value[0]) clusterID.value = clusters.value[0].id }
async function load() {
    if (!clusterID.value) { rows.value = []; nodeStats.value = { total: 0, healthy: 0, pods: 0, podCapacity: 0, cpu: 0 }; namespaceStats.value = { total: 0, healthy: 0, labels: 0, averageAge: '0d' }; return }
    if (kind.value === 'nodes') {
        const [listResult, statsResult, allNodesResult] = await Promise.all([
            listKubernetesNodes(clusterID.value, keyword.value, statusFilter.value, roleFilter.value),
            getKubernetesClusterStats(clusterID.value),
            keyword.value || statusFilter.value || roleFilter.value ? listKubernetesNodes(clusterID.value) : Promise.resolve(null),
        ])
        rows.value = listResult.data.data
        const allNodes = allNodesResult?.data.data || rows.value
        const stats = statsResult.data.data
        nodeStats.value = { total: stats.nodeCount, healthy: allNodes.filter((node: any) => isNodeReady(node.status)).length, pods: stats.podCount, podCapacity: allNodes.reduce((sum: number, node: any) => sum + Number(node.podCapacity || 0), 0), cpu: stats.cpuCapacity }
        return
    }
    if (kind.value === 'namespaces') {
        const [listResult, allResult] = await Promise.all([listKubernetesNamespaces(clusterID.value, keyword.value, namespaceStatus.value), keyword.value || namespaceStatus.value ? listKubernetesNamespaces(clusterID.value) : Promise.resolve(null)])
        rows.value = listResult.data.data
        calculateNamespaceStats(allResult?.data.data || rows.value)
        return
    }
    const r = kind.value === 'pods' ? await listKubernetesPods(clusterID.value, keyword.value) : await listKubernetesWorkloads(clusterID.value, keyword.value)
    rows.value = r.data.data
}
watch([clusterID, keyword, kind, statusFilter, roleFilter, namespaceStatus], load); loadClusters().then(load)
const displayRows = computed(() => rows.value)
async function refresh() { refreshing.value = true; try { await load() } finally { refreshing.value = false } }
function editNode(node: KubernetesNode) { nodeForm.value = { name: node.name, labels: JSON.stringify(node.labels || {}, null, 2), schedulable: node.schedulable }; nodeDialog.value = true }
async function saveNode() {
    let labels: Record<string, string>
    try { labels = JSON.parse(nodeForm.value.labels || '{}') } catch { ElMessage.error('节点标签必须是有效的 JSON 对象'); return }
    if (!labels || Array.isArray(labels) || typeof labels !== 'object') { ElMessage.error('节点标签必须是 JSON 对象'); return }
    nodeSaving.value = true
    try { await updateKubernetesNode({ cluster_id: clusterID.value, name: nodeForm.value.name, labels, schedulable: nodeForm.value.schedulable }); ElMessage.success('节点更新成功'); nodeDialog.value = false; await load() } finally { nodeSaving.value = false }
}
const indexColumn: ColDef = { headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }
const columns = computed<ColDef[]>(() => kind.value === 'nodes' ? [indexColumn, { field: 'name', headerName: '节点名称' }, { field: 'status', headerName: '状态', cellRenderer: (p: any) => { const ready = isNodeReady(p.value); const tag = document.createElement('span'); tag.className = `node-status ${ready ? 'is-ready' : 'is-abnormal'}`; tag.textContent = ready ? '正常' : p.value; return tag } }, { field: 'roles', headerName: '角色', cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = 'node-role'; tag.textContent = p.value || 'Worker'; return tag } }, { field: 'internalIP', headerName: '内部 IP' }, { field: 'cpuCapacity', headerName: 'CPU' }, { field: 'memoryCapacity', headerName: '内存' }, { field: 'podCapacity', headerName: 'Pod 容量' }, { field: 'version', headerName: 'Kubelet 版本' }, { field: 'age', headerName: '存活时间' }, { headerName: '操作', width: 150, minWidth: 150, maxWidth: 150, pinned: 'right', sortable: false, cellRenderer: (p: any) => { const button = document.createElement('button'); button.className = 'grid-action-btn edit'; button.textContent = '编辑'; button.onclick = () => editNode(p.data); return button } }] : kind.value === 'namespaces' ? [indexColumn, { field: 'name', headerName: '命名空间', minWidth: 180 }, { field: 'status', headerName: '状态' }, { field: 'age', headerName: '存活时间' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }] : kind.value === 'pods' ? [indexColumn, { field: 'namespace', headerName: '命名空间' }, { field: 'name', headerName: 'Pod 名称', minWidth: 220 }, { field: 'ready', headerName: '就绪' }, { field: 'status', headerName: '状态' }, { field: 'restarts', headerName: '重启次数' }, { field: 'ip', headerName: 'Pod IP' }, { field: 'node', headerName: '节点', minWidth: 160 }, { field: 'age', headerName: '存活时间' }] : [indexColumn, { field: 'namespace', headerName: '命名空间' }, { field: 'name', headerName: '工作负载名称', minWidth: 180 }, { field: 'kind', headerName: '类型' }, { field: 'ready', headerName: '就绪副本' }, { field: 'status', headerName: '状态' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }])
const adaptiveColumns = computed<ColDef[]>(() => columns.value.map(column => {
    if (column.headerName === '操作') return { ...column, width: 100, minWidth: 100, maxWidth: 100 }
    if (kind.value === 'namespaces' && column.field === 'status') return { ...column, cellRenderer: (p: any) => { const healthy = ['active', '正常'].includes(String(p.value || '').toLowerCase()); const tag = document.createElement('span'); tag.className = `node-status ${healthy ? 'is-ready' : 'is-abnormal'}`; tag.textContent = healthy ? '正常' : p.value; return tag } }
    if (kind.value === 'namespaces' && (column.field === 'name' || column.field === 'created_at')) return { ...column, minWidth: undefined }
    return column
}))
</script>
<template>
    <section class="page">
        <div v-if="kind === 'nodes'" class="node-header">
            <div class="node-heading"><div class="node-heading-icon"><el-icon><Monitor /></el-icon></div><div><strong>节点管理</strong><span>管理 Kubernetes 集群节点，监控节点状态和资源使用</span></div></div>
            <div class="node-header-actions"><el-select v-model="clusterID" placeholder="选择集群"><el-option v-for="item in clusters" :key="item.id" :label="item.alias || item.name" :value="item.id" /></el-select><el-button type="primary" :icon="Refresh" :loading="refreshing" @click="refresh">刷新</el-button></div>
        </div>
        <div v-else-if="kind === 'namespaces'" class="node-header">
            <div class="node-heading"><div class="node-heading-icon"><el-icon><FolderOpened /></el-icon></div><div><strong>命名空间</strong><span>管理 Kubernetes 命名空间，查看状态、标签和运行时间</span></div></div>
            <div class="node-header-actions"><el-select v-model="clusterID" placeholder="选择集群"><el-option v-for="item in clusters" :key="item.id" :label="item.alias || item.name" :value="item.id" /></el-select><el-button type="primary" :icon="Refresh" :loading="refreshing" @click="refresh">刷新</el-button></div>
        </div>
        <div v-else class="toolbar">
            <div class="filters"><el-select v-model="clusterID" placeholder="选择集群"><el-option v-for="item in clusters"
                        :key="item.id" :label="item.alias || item.name" :value="item.id" /></el-select><el-input
                    v-model="keyword" :placeholder="`搜索${title}`" clearable /></div><strong>{{ title }}</strong>
        </div>
        <div v-if="kind === 'nodes'" class="stats">
            <div class="stat-card"><div class="stat-icon">▰</div><div class="stat-content"><span>节点总数</span><strong>{{ nodeStats.total }}</strong></div></div>
            <div class="stat-card healthy"><div class="stat-icon">✓</div><div class="stat-content"><span>运行正常</span><strong>{{ nodeStats.healthy }}</strong></div></div>
            <div class="stat-card pods"><div class="stat-icon">◇</div><div class="stat-content"><span>Pod 总数</span><strong>{{ nodeStats.pods }}/{{ nodeStats.podCapacity }}</strong></div></div>
            <div class="stat-card cpu"><div class="stat-icon">⌁</div><div class="stat-content"><span>总 CPU 核数</span><strong>{{ nodeStats.cpu }}</strong></div></div>
        </div>
        <div v-if="kind === 'namespaces'" class="stats">
            <div class="stat-card"><div class="stat-icon">▰</div><div class="stat-content"><span>命名空间数</span><strong>{{ namespaceStats.total }}</strong></div></div>
            <div class="stat-card healthy"><div class="stat-icon">✓</div><div class="stat-content"><span>正常状态</span><strong>{{ namespaceStats.healthy }}</strong></div></div>
            <div class="stat-card pods"><div class="stat-icon">◇</div><div class="stat-content"><span>标签总数</span><strong>{{ namespaceStats.labels }}</strong></div></div>
            <div class="stat-card cpu"><div class="stat-icon">⌁</div><div class="stat-content"><span>平均运行时间</span><strong>{{ namespaceStats.averageAge }}</strong></div></div>
        </div>
        <div v-if="kind === 'nodes'" class="node-filters"><el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索节点名称或 IP" clearable/><el-select v-model="statusFilter" placeholder="节点状态" clearable><el-option label="正常" value="Ready"/><el-option label="异常" value="NotReady"/></el-select><el-select v-model="roleFilter" placeholder="节点角色" clearable><el-option label="Master" value="master"/><el-option label="Worker" value="worker"/></el-select></div>
        <div v-else-if="kind === 'namespaces'" class="node-filters"><el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索命名空间" clearable/><el-select v-model="namespaceStatus" placeholder="命名空间状态" clearable><el-option label="正常" value="Active"/><el-option label="终止中" value="Terminating"/></el-select></div>
        <AppDataGrid class="grid" :row-data="displayRows" :column-defs="adaptiveColumns" pagination />
        <el-dialog v-model="nodeDialog" title="编辑节点" width="520px" destroy-on-close><el-form label-width="90px"><el-form-item label="节点名称"><el-input v-model="nodeForm.name" disabled/></el-form-item><el-form-item label="允许调度"><el-switch v-model="nodeForm.schedulable"/></el-form-item><el-form-item label="节点标签"><el-input v-model="nodeForm.labels" type="textarea" :rows="10" placeholder="请输入 JSON 格式的节点标签"/></el-form-item></el-form><template #footer><el-button @click="nodeDialog=false">取消</el-button><el-button type="primary" :loading="nodeSaving" @click="saveNode">保存</el-button></template></el-dialog>
    </section>
</template>
<style
    scoped>
    .page {
        height: 100%;
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-height: 0
    }

    .toolbar {
        display: flex;
        justify-content: space-between;
        align-items: center
    }

    .filters {
        display: flex;
        gap: 8px
    }

    .filters .el-select {
        width: 180px
    }

    .filters .el-input {
        width: 220px
    }

    .grid {
        flex: 1;
        min-height: 0
    }

    .node-header,.node-filters { display:flex; align-items:center; justify-content:space-between; padding:14px 18px; border:1px solid var(--line); border-radius:10px; background:#fff; box-shadow:0 3px 12px rgba(31,35,48,.05) }
    .node-heading,.node-header-actions,.node-filters { display:flex; align-items:center; gap:10px }
    .node-heading-icon { display:flex; align-items:center; justify-content:center; width:38px; height:38px; flex:none; border-radius:9px; color:var(--primary); background:color-mix(in srgb,var(--primary) 11%,white); border:1px solid color-mix(in srgb,var(--primary) 24%,white); font-size:18px }
    .node-heading>div:last-child { display:flex; flex-direction:column; gap:4px }
    .node-heading strong { font-size:15px; color:var(--text) }
    .node-heading span { font-size:11px; color:var(--text-soft) }
    .node-header-actions .el-select { width:210px }
    .node-header-actions .el-button { margin-left:0 }
    .node-filters { justify-content:flex-start; padding:10px 14px }
    .node-filters .el-input { width:240px }
    .node-filters .el-select { width:150px }
    .grid :deep(.node-status),.grid :deep(.node-role) { display:inline-flex; align-items:center; justify-content:center; padding:3px 9px; border-radius:5px; font-size:11px; line-height:1.2 }
    .grid :deep(.node-status.is-ready) { color:#247d31; background:#eaf7e8; border:1px solid #9bd393 }
    .grid :deep(.node-status.is-abnormal) { color:#b54747; background:#fff0f0; border:1px solid #efb1b1 }
    .grid :deep(.node-role) { color:var(--primary); background:color-mix(in srgb,var(--primary) 7%,white); border:1px solid color-mix(in srgb,var(--primary) 28%,white) }
    .grid :deep(.grid-action-btn) { border:1px solid var(--primary); border-radius:8px; padding:4px 9px; background:#f7f8ff; color:var(--primary); font:inherit; font-size:11px; line-height:1.25; cursor:pointer; box-shadow:0 1px 3px rgba(24,34,53,.08) }
    .grid :deep(.grid-action-btn:hover) { background:var(--primary-soft) }
    .grid :deep(.ag-pinned-right-cols-container .ag-cell) { display:flex; align-items:center; justify-content:center; padding-right:4px; padding-left:4px }

    .stats { display:grid; grid-template-columns:repeat(4,minmax(120px,144px)); justify-content:start; gap:10px }
    .stat-card { display:flex; align-items:center; width:100%; max-width:144px; min-width:0; padding:14px 16px; border:1px solid var(--line); border-radius:9px; background:#fff; box-shadow:0 3px 12px rgba(31,35,48,.06) }
    .stat-icon { display:flex; align-items:center; justify-content:center; width:34px; height:34px; flex:none; margin-right:12px; border-radius:8px; font-size:16px; font-weight:700; color:var(--primary); background:color-mix(in srgb,var(--primary) 10%,white) }
    .stat-content { display:flex; min-width:0; flex:1; flex-direction:column; gap:5px }
    .stat-content span { font-size:11px; color:var(--text-secondary,#717784); white-space:nowrap }
    .stat-content strong { font-size:22px; line-height:1; color:var(--text) }
    .stat-card.healthy .stat-icon,.stat-card.healthy strong { color:#20a162 }
    .stat-card.healthy .stat-icon { background:#eaf8ef }
    .stat-card.pods .stat-icon { color:#2878c8; background:#edf5ff }
    .stat-card.cpu .stat-icon { color:#b57816; background:#fff6df }

    @media(max-width:760px) {
        .stats { grid-template-columns:repeat(2,minmax(120px,144px)) }
        .node-header { align-items:flex-start; flex-direction:column }
        .node-header-actions,.node-filters { width:100%; flex-wrap:wrap }
        .node-header-actions .el-select,.node-filters .el-input,.node-filters .el-select { width:100% }
    }
</style>
