<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { ColDef } from 'ag-grid-community'
import { Delete, FolderOpened, Monitor, Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import KubernetesLabelDialog from '@/components/KubernetesLabelDialog.vue'
import { applyKubernetesObject, deleteKubernetesNode, deleteKubernetesObject, drainKubernetesNode, getKubernetesClusterStats, getKubernetesObject, listKubernetesClusters, listKubernetesNamespaces, listKubernetesNodes, listKubernetesPods, listKubernetesWorkloads, updateKubernetesNamespace, updateKubernetesNode, type KubernetesCluster, type KubernetesNode, type KubernetesNamespace, type KubernetesTaint } from '@/api/kubernetes'
const route = useRoute(); const kind = computed(() => String(route.meta.kubernetesResource || 'nodes')); const title = computed(() => kind.value === 'nodes' ? '节点管理' : kind.value === 'namespaces' ? '命名空间' : kind.value === 'pods' ? 'Pod 管理' : '工作负载'); const clusters = ref<KubernetesCluster[]>([]), clusterID = ref(0), keyword = ref(''), rows = ref<any[]>([])
const statusFilter = ref(''), roleFilter = ref(''), namespaceStatus = ref(''), refreshing = ref(false)
const cloudTTYDialog = ref(false)
const cloudTTYCommand = `# 1. 安装并等待 Pod 运行起来
helm repo add cloudtty https://cloudtty.github.io/cloudtty
helm repo update
helm install cloudtty-operator cloudtty/cloudtty \\
  --version 0.5.0 \\
  --create-namespace \\
  --namespace cloudtty-system

# 2. 创建 cloudshell.yaml
cat <<'EOF' > cloudshell.yaml
apiVersion: cloudshell.cloudtty.io/v1alpha1
kind: CloudShell
metadata:
  name: permanent-terminal
  namespace: cloudtty-system
spec:
  commandAction: "bash -il"
  exposureMode: NodePort
  once: false
  cleanup: false
  ttlSecondsAfterStarted: 315360000
  urlArg: true
  env:
    - name: TTYD_WRITABLE
      value: "true"
    - name: TTYD_SERVER_BUFFER_SIZE
      value: "4096"
    - name: TTYD_CLIENT_BUFFER_SIZE
      value: "4096"
  ttydClientOptions:
    fontSize: "14"
    fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace"
    cursorBlink: "true"
    rendererType: "canvas"
  image: "cloudshell/cloudshell:latest"
EOF

kubectl apply -f cloudshell.yaml

# 3. 观察 CloudShell 状态并获取访问入口
kubectl get cloudshell -n cloudtty-system -w`
const nodeDialog = ref(false), nodeSaving = ref(false), nodeForm = ref({ name: '', labels: '', schedulable: true })
const namespaceLabelDialog = ref(false), namespaceLabelSaving = ref(false), selectedNamespace = ref<KubernetesNamespace | null>(null)
const nodeLabelDialog = ref(false), nodeLabelSaving = ref(false), selectedLabelNode = ref<KubernetesNode | null>(null)
const nodeTaintDialog = ref(false), nodeTaintEditing = ref(false), nodeTaintSaving = ref(false), selectedTaintNode = ref<KubernetesNode | null>(null), nodeTaintRows = ref<KubernetesTaint[]>([])
const nodeMenu = ref({ visible:false, x:0, y:0, node:null as KubernetesNode | null }), nodeYamlDialog = ref(false), nodeYamlSaving = ref(false), nodeYaml = ref('')
const namespaceMenu = ref({ visible:false, x:0, y:0, namespace:null as KubernetesNamespace | null }), namespaceYamlDialog = ref(false), namespaceYamlSaving = ref(false), namespaceYaml = ref('')
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
    if (kind.value === 'pods') {
        const r = await listKubernetesPods(clusterID.value, keyword.value)
        rows.value = r.data.data
        return
    }
    const r = await listKubernetesWorkloads(clusterID.value, keyword.value)
    rows.value = r.data.data
}
watch([clusterID, keyword, kind, statusFilter, roleFilter, namespaceStatus], load); loadClusters().then(load)
const displayRows = computed(() => rows.value)
async function refresh() { refreshing.value = true; try { await load() } finally { refreshing.value = false } }
async function copyCloudTTYCommand() { await navigator.clipboard.writeText(cloudTTYCommand); ElMessage.success('部署命令已复制') }
function finishCloudTTYDeployment() { cloudTTYDialog.value = false; ElMessage.success('已记录 CloudTTY 部署完成') }
function editNode(node: KubernetesNode) { nodeForm.value = { name: node.name, labels: JSON.stringify(node.labels || {}, null, 2), schedulable: node.schedulable }; nodeDialog.value = true }
async function saveNode() {
    let labels: Record<string, string>
    try { labels = JSON.parse(nodeForm.value.labels || '{}') } catch { ElMessage.error('节点标签必须是有效的 JSON 对象'); return }
    if (!labels || Array.isArray(labels) || typeof labels !== 'object') { ElMessage.error('节点标签必须是 JSON 对象'); return }
    nodeSaving.value = true
    try { await updateKubernetesNode({ cluster_id: clusterID.value, name: nodeForm.value.name, labels, schedulable: nodeForm.value.schedulable }); ElMessage.success('节点更新成功'); nodeDialog.value = false; await load() } finally { nodeSaving.value = false }
}
async function removeNode(node: KubernetesNode) {
    try { await ElMessageBox.confirm(`删除节点“${node.name}”后，该节点将从 Kubernetes 集群中移除。确定继续吗？`, '删除节点', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }) } catch { return }
    await deleteKubernetesNode(clusterID.value, node.name)
    ElMessage.success('节点删除成功')
    await load()
}
function openNodeActionMenu(node: KubernetesNode, button: HTMLElement) { const rect = button.getBoundingClientRect(); nodeMenu.value = { visible: true, x: Math.max(8, rect.right - 180), y: rect.bottom + 4, node } }
function closeNodeMenu() { nodeMenu.value.visible = false }
function openNodeShell() { closeNodeMenu(); ElMessage.info('Shell 需要先在目标集群部署并配置 CloudTTY') }
async function openNodeYaml() { const node = nodeMenu.value.node; closeNodeMenu(); if (!node) return; const { data } = await getKubernetesObject({ cluster_id: clusterID.value, resource: 'nodes', name: node.name }); nodeYaml.value = data.data.yaml; nodeYamlDialog.value = true }
async function saveNodeYaml() { nodeYamlSaving.value = true; try { await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'nodes', yaml: nodeYaml.value }); ElMessage.success('节点 YAML 保存成功'); nodeYamlDialog.value = false; await load() } finally { nodeYamlSaving.value = false } }
async function setNodeSchedulable(schedulable: boolean) { const node = nodeMenu.value.node; closeNodeMenu(); if (!node) return; await updateKubernetesNode({ cluster_id: clusterID.value, name: node.name, labels: node.labels || {}, schedulable, taints: node.taints || [] }); ElMessage.success(schedulable ? '节点已设为可调度' : '节点已设为不可调度'); await load() }
async function drainNode() { const node = nodeMenu.value.node; closeNodeMenu(); if (!node) return; try { await ElMessageBox.confirm(`排空节点“${node.name}”将驱逐节点上的非 DaemonSet Pod，并将节点设为不可调度。确定继续吗？`, '节点排空', { type: 'warning', confirmButtonText: '确认排空', cancelButtonText: '取消' }) } catch { return }; await drainKubernetesNode(clusterID.value, node.name); ElMessage.success('节点排空完成'); await load() }
function openNamespaceActionMenu(namespace: KubernetesNamespace, button: HTMLElement) { const rect = button.getBoundingClientRect(); namespaceMenu.value = { visible:true, x:Math.max(8, rect.right - 150), y:rect.bottom + 4, namespace } }
function closeNamespaceMenu() { namespaceMenu.value.visible = false }
function modifyNamespace() { const namespace = namespaceMenu.value.namespace; closeNamespaceMenu(); if (!namespace) return; openNamespaceLabels(namespace) }
async function openNamespaceYaml() { const namespace = namespaceMenu.value.namespace; closeNamespaceMenu(); if (!namespace) return; const { data } = await getKubernetesObject({ cluster_id:clusterID.value, resource:'namespaces', name:namespace.name }); namespaceYaml.value = data.data.yaml; namespaceYamlDialog.value = true }
async function saveNamespaceYaml() { namespaceYamlSaving.value = true; try { await applyKubernetesObject({ cluster_id:clusterID.value, resource:'namespaces', yaml:namespaceYaml.value }); ElMessage.success('命名空间 YAML 保存成功'); namespaceYamlDialog.value = false; await load() } finally { namespaceYamlSaving.value = false } }
async function removeNamespace(namespace: KubernetesNamespace) { try { await ElMessageBox.confirm(`删除命名空间“${namespace.name}”将同时删除其中的全部资源，且无法恢复。确定继续吗？`, '删除命名空间', { type:'warning', confirmButtonText:'确认删除', cancelButtonText:'取消' }) } catch { return }; await deleteKubernetesObject({ cluster_id:clusterID.value, resource:'namespaces', name:namespace.name }); ElMessage.success('命名空间删除成功'); await load() }
function openNamespaceLabels(namespace: KubernetesNamespace) { selectedNamespace.value = namespace; namespaceLabelDialog.value = true }
function openNodeLabels(node: KubernetesNode) { selectedLabelNode.value = node; nodeLabelDialog.value = true }
async function saveNamespaceLabelsFromDialog(labels:Record<string,string>){if(!selectedNamespace.value)return;namespaceLabelSaving.value=true;try{await updateKubernetesNamespace({cluster_id:clusterID.value,name:selectedNamespace.value.name,labels});namespaceLabelDialog.value=false;ElMessage.success('命名空间标签更新成功');await load()}finally{namespaceLabelSaving.value=false}}
async function saveNodeLabelsFromDialog(labels:Record<string,string>){if(!selectedLabelNode.value)return;nodeLabelSaving.value=true;try{await updateKubernetesNode({cluster_id:clusterID.value,name:selectedLabelNode.value.name,labels,schedulable:selectedLabelNode.value.schedulable});nodeLabelDialog.value=false;ElMessage.success('节点标签更新成功');await load()}finally{nodeLabelSaving.value=false}}
function openNodeTaints(node: KubernetesNode) { selectedTaintNode.value = node; nodeTaintRows.value = (node.taints || []).map(item => ({ ...item })); nodeTaintEditing.value = false; nodeTaintDialog.value = true }
function startNodeTaintEdit() { nodeTaintEditing.value = true; if (!nodeTaintRows.value.length) nodeTaintRows.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function addNodeTaint() { nodeTaintRows.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeNodeTaint(index: number) { nodeTaintRows.value.splice(index, 1) }
async function saveNodeTaints() {
    if (!selectedTaintNode.value) return
    const seen = new Set<string>()
    for (const item of nodeTaintRows.value) { item.key = item.key.trim(); if (!item.key) { ElMessage.warning('污点 Key 不能为空'); return }; const identity = `${item.key}:${item.effect}`; if (seen.has(identity)) { ElMessage.warning(`污点重复：${item.key}`); return }; seen.add(identity) }
    nodeTaintSaving.value = true
    try { await updateKubernetesNode({ cluster_id: clusterID.value, name: selectedTaintNode.value.name, labels: selectedTaintNode.value.labels || {}, schedulable: selectedTaintNode.value.schedulable, taints: nodeTaintRows.value }); ElMessage.success('节点污点更新成功'); nodeTaintDialog.value = false; await load() } finally { nodeTaintSaving.value = false }
}
const indexColumn: ColDef = { headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }
const columns = computed<ColDef[]>(() => kind.value === 'nodes' ? [indexColumn, { field: 'name', headerName: '节点名称' }, { field: 'status', headerName: '状态', cellRenderer: (p: any) => { const ready = isNodeReady(p.value); const tag = document.createElement('span'); tag.className = `node-status ${ready ? 'is-ready' : 'is-abnormal'}`; tag.textContent = ready ? '正常' : p.value; return tag } }, { field: 'roles', headerName: '角色', cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = 'node-role'; tag.textContent = p.value || 'Worker'; return tag } }, { field: 'internalIP', headerName: '内部 IP' }, { field: 'cpuCapacity', headerName: 'CPU' }, { field: 'memoryCapacity', headerName: '内存' }, { field: 'podCapacity', headerName: 'Pod 容量' }, { field: 'version', headerName: 'Kubelet 版本' }, { field: 'age', headerName: '存活时间' }, { headerName: '操作', width: 150, minWidth: 150, maxWidth: 150, pinned: 'right', sortable: false, cellRenderer: (p: any) => { const button = document.createElement('button'); button.className = 'grid-action-btn edit'; button.textContent = '编辑'; button.onclick = () => editNode(p.data); return button } }] : kind.value === 'namespaces' ? [indexColumn, { field: 'name', headerName: '命名空间', minWidth: 180 }, { field: 'status', headerName: '状态' }, { field: 'age', headerName: '运行时间' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }] : kind.value === 'pods' ? [indexColumn, { field: 'namespace', headerName: '命名空间' }, { field: 'name', headerName: 'Pod 名称', minWidth: 220 }, { field: 'ready', headerName: '就绪' }, { field: 'status', headerName: '状态' }, { field: 'restarts', headerName: '重启次数' }, { field: 'ip', headerName: 'Pod IP' }, { field: 'node', headerName: '节点', minWidth: 160 }, { field: 'age', headerName: '存活时间' }] : [indexColumn, { field: 'namespace', headerName: '命名空间' }, { field: 'name', headerName: '工作负载名称', minWidth: 180 }, { field: 'kind', headerName: '类型' }, { field: 'ready', headerName: '就绪副本' }, { field: 'status', headerName: '状态' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }])
const adaptiveColumns = computed<ColDef[]>(() => {
    const source = [...columns.value]
    if (kind.value === 'nodes') {
        const field = (name: string) => source.find(column => column.field === name)!
        const action = source.find(column => column.headerName === '操作')!
        const labelColumn: ColDef = { field: 'labels', headerName: '标签', cellRenderer: (p: any) => { const labels = p.value || {}; const wrap = document.createElement('span'); const icon = document.createElement('span'); const badge = document.createElement('span'); wrap.className = 'namespace-label-count'; icon.className = 'namespace-label-icon'; icon.textContent = '◇'; badge.className = 'namespace-label-badge'; badge.textContent = String(Object.keys(labels).length); wrap.title = '查看节点标签'; wrap.onclick = () => openNodeLabels(p.data); wrap.append(icon, badge); return wrap } }
        const schedulableColumn: ColDef = { field: 'schedulable', headerName: '调度', cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = `node-status ${p.value ? 'is-ready' : 'is-abnormal'}`; tag.textContent = p.value ? '可调度' : '不可调度'; return tag } }
        const taintColumn: ColDef = { field: 'taints', headerName: '污点', cellRenderer: (p: any) => { const taints = p.value || []; const wrap = document.createElement('span'); const icon = document.createElement('span'); const badge = document.createElement('span'); wrap.className = 'namespace-label-count taint-count'; icon.className = 'namespace-label-icon'; icon.textContent = '◇'; badge.className = 'namespace-label-badge'; badge.textContent = String(taints.length); wrap.title = '查看节点污点'; wrap.onclick = () => openNodeTaints(p.data); wrap.append(icon, badge); return wrap } }
        source.splice(0, source.length, field('name') ? source[0] : indexColumn, field('name'), field('status'), field('roles'), field('internalIP'), { ...field('version'), headerName: 'Kubelet 版本' }, labelColumn, field('cpuCapacity'), field('memoryCapacity'), { ...field('podCapacity'), headerName: 'Pod 数量', valueGetter: (p: any) => `${Number(p.data?.podCount || 0)} / ${Number(p.data?.podCapacity || 0)}` }, schedulableColumn, taintColumn, { ...field('age'), headerName: '运行时间' }, action)
    }
    if (kind.value === 'namespaces') {
        const statusIndex = source.findIndex(column => column.field === 'status')
        source.splice(statusIndex + 1, 0, { field: 'labels', headerName: '标签', cellRenderer: (p: any) => { const labels = p.value || {}; const wrap = document.createElement('span'); const icon = document.createElement('span'); const badge = document.createElement('span'); wrap.className = 'namespace-label-count'; icon.className = 'namespace-label-icon'; icon.textContent = '◇'; badge.className = 'namespace-label-badge'; badge.textContent = String(Object.keys(labels).length); wrap.title = '查看命名空间标签'; wrap.onclick = () => openNamespaceLabels(p.data); wrap.append(icon, badge); return wrap } })
        source.push({ headerName:'操作', pinned:'right', width:150, cellRenderer:(p:any) => { const wrap=document.createElement('div'); const editButton=document.createElement('button'); const deleteButton=document.createElement('button'); wrap.className='node-action-group'; editButton.className='grid-action-btn edit'; editButton.textContent='编辑'; editButton.onclick=()=>openNamespaceActionMenu(p.data,editButton); deleteButton.className='grid-action-btn delete'; deleteButton.textContent='删除'; deleteButton.onclick=()=>removeNamespace(p.data); wrap.append(editButton,deleteButton); return wrap } })
    }
    return source.map(column => {
    if (column.headerName === '操作' && kind.value === 'nodes') return { ...column, width: 150, minWidth: 150, maxWidth: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); const editButton = document.createElement('button'); const deleteButton = document.createElement('button'); wrap.className = 'node-action-group'; editButton.className = 'grid-action-btn edit'; editButton.textContent = '编辑'; editButton.onclick = () => openNodeActionMenu(p.data, editButton); deleteButton.className = 'grid-action-btn delete'; deleteButton.textContent = '删除'; deleteButton.onclick = () => removeNode(p.data); wrap.append(editButton, deleteButton); return wrap } }
    if (column.headerName === '操作') return { ...column, width:150, minWidth:150, maxWidth:150 }
    if (kind.value === 'namespaces' && column.field === 'status') return { ...column, cellRenderer: (p: any) => { const healthy = ['active', '正常'].includes(String(p.value || '').toLowerCase()); const tag = document.createElement('span'); tag.className = `node-status ${healthy ? 'is-ready' : 'is-abnormal'}`; tag.textContent = healthy ? '正常' : p.value; return tag } }
    if (kind.value === 'namespaces' && (column.field === 'name' || column.field === 'created_at')) return { ...column, minWidth: undefined }
    return column
    })
})
</script>
<template>
    <section class="page" :class="{ 'node-page': kind === 'nodes', 'compact-page': kind === 'namespaces' }">
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
        <div v-if="kind === 'nodes'" class="node-filters"><el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索节点名称或 IP" clearable/><el-select v-model="statusFilter" placeholder="节点状态" clearable><el-option label="正常" value="Ready"/><el-option label="异常" value="NotReady"/></el-select><el-select v-model="roleFilter" placeholder="节点角色" clearable><el-option label="Master" value="master"/><el-option label="Worker" value="worker"/></el-select><el-button type="primary" plain @click="cloudTTYDialog = true">部署 CloudTTY</el-button></div>
        <div v-else-if="kind === 'namespaces'" class="node-filters"><el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索命名空间" clearable/><el-select v-model="namespaceStatus" placeholder="命名空间状态" clearable><el-option label="正常" value="Active"/><el-option label="终止中" value="Terminating"/></el-select></div>
        <AppDataGrid class="grid" :row-data="displayRows" :column-defs="adaptiveColumns" pagination />
        <div v-if="nodeMenu.visible" class="node-action-menu" :style="{left:`${nodeMenu.x}px`,top:`${nodeMenu.y}px`}" @mouseleave="closeNodeMenu"><button @click="openNodeShell">Shell</button><button @click="openNodeYaml">YAML</button><div class="menu-divider"></div><button @click="drainNode">节点排空</button><button @click="setNodeSchedulable(false)">设为不可调度</button><button @click="setNodeSchedulable(true)">设为可调度</button></div>
        <div v-if="namespaceMenu.visible" class="node-action-menu namespace-action-menu" :style="{left:`${namespaceMenu.x}px`,top:`${namespaceMenu.y}px`}" @mouseleave="closeNamespaceMenu"><button @click="modifyNamespace">修改</button><button @click="openNamespaceYaml">YAML</button></div>
        <el-dialog v-model="nodeYamlDialog" :title="`节点 YAML · ${nodeMenu.node?.name || ''}`" width="60%" destroy-on-close><el-input v-model="nodeYaml" type="textarea" :rows="22" class="yaml-editor"/><template #footer><el-button @click="nodeYamlDialog=false">取消</el-button><el-button type="primary" :loading="nodeYamlSaving" @click="saveNodeYaml">保存</el-button></template></el-dialog>
        <el-dialog v-model="namespaceYamlDialog" :title="`命名空间 YAML · ${namespaceMenu.namespace?.name || ''}`" width="60%" destroy-on-close><el-input v-model="namespaceYaml" type="textarea" :rows="22" class="yaml-editor"/><template #footer><el-button @click="namespaceYamlDialog=false">取消</el-button><el-button type="primary" :loading="namespaceYamlSaving" @click="saveNamespaceYaml">保存</el-button></template></el-dialog>
        <el-dialog v-model="nodeDialog" title="编辑节点" width="520px" destroy-on-close><el-form label-width="90px"><el-form-item label="节点名称"><el-input v-model="nodeForm.name" disabled/></el-form-item><el-form-item label="允许调度"><el-switch v-model="nodeForm.schedulable"/></el-form-item><el-form-item label="节点标签"><el-input v-model="nodeForm.labels" type="textarea" :rows="10" placeholder="请输入 JSON 格式的节点标签"/></el-form-item></el-form><template #footer><el-button @click="nodeDialog=false">取消</el-button><el-button type="primary" :loading="nodeSaving" @click="saveNode">保存</el-button></template></el-dialog>
        <el-dialog v-model="cloudTTYDialog" class="cloudtty-dialog" title="部署 CloudTTY" width="min(980px, 82vw)" top="4vh" :close-on-click-modal="false" destroy-on-close>
            <div class="cloudtty-content">
                <el-alert type="info" :closable="false" show-icon><template #title>CloudTTY 是一个 Kubernetes Web Terminal 解决方案</template><div class="cloudtty-intro">可提供节点 Shell、文件上传/下载、多标签页和会话审计录制能力。</div></el-alert>
                <section class="cloudtty-steps"><strong>部署步骤</strong><ol><li class="active">复制下方命令</li><li>在集群控制节点执行命令</li><li>等待部署完成（约 2–3 分钟）</li><li>点击“我已完成部署”检查部署状态</li></ol></section>
                <section class="cloudtty-command"><div class="command-title"><strong>部署命令</strong><el-button type="primary" plain @click="copyCloudTTYCommand">复制命令</el-button></div><pre><code>{{ cloudTTYCommand }}</code></pre></section>
            </div>
            <template #footer><el-button @click="cloudTTYDialog = false">关闭</el-button><el-button type="success" @click="finishCloudTTYDeployment">我已完成部署</el-button></template>
        </el-dialog>
        <KubernetesLabelDialog v-model="namespaceLabelDialog" :title="`${selectedNamespace?.name||''} 标签`" :labels="selectedNamespace?.labels||{}" :saving="namespaceLabelSaving" @save="saveNamespaceLabelsFromDialog"/>
        <KubernetesLabelDialog v-model="nodeLabelDialog" :title="`${selectedLabelNode?.name||''} 标签`" :labels="selectedLabelNode?.labels||{}" :saving="nodeLabelSaving" @save="saveNodeLabelsFromDialog"/>
        <el-dialog v-model="nodeTaintDialog" class="namespace-label-dialog" :title="nodeTaintEditing ? '编辑节点污点' : '节点污点'" width="60%" destroy-on-close>
            <div v-if="!nodeTaintEditing" class="namespace-label-view"><div class="node-taint-head"><strong>Key</strong><strong>Value</strong><strong>Effect</strong></div><div v-for="item in nodeTaintRows" :key="`${item.key}:${item.effect}`" class="node-taint-view-row"><code>{{ item.key }}</code><span>{{ item.value || '-' }}</span><span>{{ item.effect }}</span></div><el-empty v-if="!nodeTaintRows.length" description="暂无污点" :image-size="54"/></div>
            <div v-else class="namespace-label-editor"><div class="namespace-label-editor-title"><span>编辑 {{ selectedTaintNode?.name }} 的污点</span><strong>共 {{ nodeTaintRows.length }} 个污点</strong></div><div v-for="(item,index) in nodeTaintRows" :key="index" class="node-taint-edit-row"><span class="taint-row-index">{{ index + 1 }}</span><div class="taint-field"><label>KEY</label><el-input v-model="item.key" placeholder="如: key1"/></div><span class="taint-separator">=</span><div class="taint-field"><label>VALUE</label><el-input v-model="item.value" placeholder="可选"/></div><span class="taint-separator">:</span><div class="taint-field effect-field"><label>EFFECT</label><el-select v-model="item.effect" popper-class="taint-effect-popper"><el-option label="NoSchedule" value="NoSchedule"><div class="taint-effect-option"><strong>NoSchedule</strong><span>Pod 不会被调度</span></div></el-option><el-option label="PreferNoSchedule" value="PreferNoSchedule"><div class="taint-effect-option"><strong>PreferNoSchedule</strong><span>尽量不调度</span></div></el-option><el-option label="NoExecute" value="NoExecute"><div class="taint-effect-option danger"><strong>NoExecute</strong><span>驱逐已有 Pod</span></div></el-option></el-select></div><el-button class="taint-delete" type="danger" circle :icon="Delete" @click="removeNodeTaint(index)"/></div><el-button class="add-taint-button" plain @click="addNodeTaint">＋ 添加污点</el-button></div>
            <template #footer><el-button @click="nodeTaintDialog=false">{{ nodeTaintEditing ? '取消' : '关闭' }}</el-button><el-button v-if="!nodeTaintEditing" type="primary" @click="startNodeTaintEdit">编辑污点</el-button><el-button v-else type="primary" :loading="nodeTaintSaving" @click="saveNodeTaints">保存更改</el-button></template>
        </el-dialog>
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

    .node-page { gap:2px }
    .node-page .stats { order:1 }
    .node-page .node-header { order:2 }
    .node-page .node-filters { order:3 }
    .node-page .grid { order:4 }
    .node-page .stat-card { padding:10px 12px }
    .node-page .node-header { padding:10px 14px }
    .node-page .node-filters { padding:8px 12px }
    .compact-page { gap:2px }
    .compact-page .stat-card { padding:10px 12px }
    .compact-page .node-header { padding:10px 14px }
    .compact-page .node-filters { padding:8px 12px }
    .compact-page .stats { order:1 }
    .compact-page .node-header { order:2 }
    .compact-page .node-filters { order:3 }
    .compact-page .grid { order:4 }

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
    .node-filters .el-button { margin-left:0; white-space:nowrap }
    .grid :deep(.node-status),.grid :deep(.node-role) { display:inline-flex; align-items:center; justify-content:center; padding:3px 9px; border-radius:5px; font-size:11px; line-height:1.2 }
    .grid :deep(.node-status.is-ready) { color:#247d31; background:#eaf7e8; border:1px solid #9bd393 }
    .grid :deep(.node-status.is-abnormal) { color:#b54747; background:#fff0f0; border:1px solid #efb1b1 }
    .grid :deep(.node-role) { color:var(--primary); background:color-mix(in srgb,var(--primary) 7%,white); border:1px solid color-mix(in srgb,var(--primary) 28%,white) }
    .grid :deep(.grid-action-btn) { border:1px solid var(--primary); border-radius:8px; padding:4px 9px; background:#f7f8ff; color:var(--primary); font:inherit; font-size:11px; line-height:1.25; cursor:pointer; box-shadow:0 1px 3px rgba(24,34,53,.08) }
    .grid :deep(.grid-action-btn:hover) { background:var(--primary-soft) }
    .grid :deep(.node-action-group) { display:flex; align-items:center; justify-content:center; gap:4px; white-space:nowrap }
    .grid :deep(.grid-action-btn.delete) { border-color:#ffd5d5; background:#fff7f7; color:#ef6262 }
    .grid :deep(.grid-action-btn.delete:hover) { background:#ffecec }
    .node-action-menu{position:fixed;z-index:5000;display:flex;width:180px;flex-direction:column;padding:6px;border:1px solid var(--line);border-radius:8px;background:#fff;box-shadow:0 8px 24px rgba(24,34,53,.18)}.node-action-menu button{padding:9px 12px;border:0;border-radius:6px;background:#fff;color:var(--text);font:inherit;text-align:left;cursor:pointer}.node-action-menu button:hover{background:var(--primary-soft);color:var(--primary)}.node-action-menu .menu-divider{height:1px;margin:5px -6px;background:var(--line)}.yaml-editor :deep(.el-textarea__inner){font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
    .grid :deep(.ag-pinned-right-cols-container .ag-cell) { display:flex; align-items:center; justify-content:center; padding-right:4px; padding-left:4px }
    .grid :deep(.namespace-label-count) { position:relative; display:inline-flex; width:30px; height:28px; align-items:center; justify-content:flex-start; color:#c99818; cursor:help }
    .grid :deep(.namespace-label-icon) { font-size:25px; line-height:1; transform:rotate(-45deg) }
    .grid :deep(.namespace-label-badge) { position:absolute; top:0; right:0; display:flex; min-width:16px; height:16px; align-items:center; justify-content:center; padding:0 3px; border-radius:8px; background:#d7a91f; color:#3d3107; font-size:9px; font-weight:700; line-height:16px }
    .namespace-label-dialog :deep(.el-dialog),.namespace-label-dialog :deep(.el-dialog__header),.namespace-label-dialog :deep(.el-dialog__body),.namespace-label-dialog :deep(.el-dialog__footer) { background:#fff }
    :global(.cloudtty-dialog),:global(.cloudtty-dialog .el-dialog__header),:global(.cloudtty-dialog .el-dialog__body),:global(.cloudtty-dialog .el-dialog__footer) { background:#fff }
    .cloudtty-dialog :deep(.el-dialog__body) { max-height:78vh; padding-top:10px; overflow:auto }
    .cloudtty-content { display:flex; flex-direction:column; gap:14px; color:var(--text) }
    .cloudtty-intro { padding-top:5px; color:var(--text-soft); font-size:11px }
    .cloudtty-steps { display:grid; grid-template-columns:90px minmax(0,1fr); align-items:start; gap:12px; padding:4px 2px }
    .cloudtty-steps>strong { padding-top:3px }
    .cloudtty-steps ol { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:8px; margin:0; padding:0; list-style-position:inside; color:var(--text-soft) }
    .cloudtty-steps li { padding:8px 10px; border:1px solid var(--line); border-radius:7px; background:#fff }
    .cloudtty-steps li.active { color:var(--primary); border-color:color-mix(in srgb,var(--primary) 35%,white); background:color-mix(in srgb,var(--primary) 5%,white) }
    .cloudtty-command { min-width:0 }
    .command-title { display:flex; align-items:center; justify-content:space-between; margin-bottom:8px }
    .cloudtty-command pre { max-height:48vh; margin:0; padding:16px 18px; overflow:auto; border:1px solid #252b37; border-radius:8px; background:#0e1117; color:#f4cc45; font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace; white-space:pre }
    .namespace-label-head,.namespace-label-view-row { display:grid; grid-template-columns:1fr 1fr; align-items:center; gap:20px; padding:10px 14px; border-bottom:1px solid var(--line) }
    .namespace-label-head { color:var(--text-soft) }.namespace-label-view-row code { width:max-content; max-width:100%; padding:5px 9px; border:1px solid #ead79b; border-radius:6px; background:#fff; color:#9a7312; overflow-wrap:anywhere }
    .namespace-label-editor { display:flex; flex-direction:column; gap:12px }.namespace-label-editor-title { display:flex; align-items:center; justify-content:space-between; padding:14px 16px; border:1px solid var(--line); border-radius:9px; background:#fff; color:var(--text) }.namespace-label-editor-title strong { color:#a77910 }
    .namespace-label-edit-row { display:grid; grid-template-columns:30px minmax(0,1fr) auto minmax(0,1fr) auto; align-items:center; gap:10px; padding:12px; border:1px solid var(--line); border-radius:8px; background:#fff }.row-index { display:flex; align-items:center; justify-content:center; width:26px; height:26px; border-radius:6px; background:#f5f6fa; color:var(--text); font-weight:700 }.add-label-button { width:100%; background:#fff }
    .node-taint-head,.node-taint-view-row { display:grid; grid-template-columns:1.2fr 1fr 1fr; align-items:center; gap:16px; padding:10px 14px; border-bottom:1px solid var(--line) }.node-taint-head{color:var(--text-soft)}.node-taint-view-row code{width:max-content;max-width:100%;padding:5px 9px;border:1px solid #ead79b;border-radius:6px;background:#fff;color:#9a7312;overflow-wrap:anywhere}.node-taint-edit-row{display:grid;grid-template-columns:46px minmax(140px,1fr) 18px minmax(140px,1fr) 18px minmax(180px,.72fr) 42px;align-items:end;gap:8px;padding:14px 16px;border:1px solid #dfb32d;border-radius:9px;background:#fff}.taint-row-index{display:flex;align-items:center;justify-content:center;width:36px;height:36px;margin-bottom:1px;border:1px solid #dfb32d;border-radius:7px;background:#fff;color:#a77910;font-weight:700}.taint-field{display:flex;min-width:0;flex-direction:column;gap:6px}.taint-field label{font-size:10px;font-weight:700;color:#8b909a;letter-spacing:.04em}.taint-separator{align-self:end;padding-bottom:8px;color:#8b909a;font-size:16px;font-weight:700}.taint-delete{align-self:end;width:36px!important;height:36px!important;margin-bottom:1px}.add-taint-button{width:100%;border:1px dashed #d7a91f!important;background:#f6faff!important;color:#409eff!important;font-weight:600}.taint-count{color:#b57816}
    :global(.taint-effect-popper .el-select-dropdown__item){height:auto;padding:7px 12px}.taint-effect-option{display:flex;align-items:center;justify-content:space-between;gap:14px}.taint-effect-option strong{padding:1px 5px;border:1px solid #edc77a;border-radius:4px;color:#cf8b22;background:#fff}.taint-effect-option span{color:#8b909a;font-size:10px}.taint-effect-option.danger strong{border-color:#ffb5b5;color:#ef6262}

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
        .node-filters .el-button { width:100% }
        .cloudtty-steps { grid-template-columns:1fr }
        .cloudtty-steps ol { grid-template-columns:1fr }
    }
</style>
