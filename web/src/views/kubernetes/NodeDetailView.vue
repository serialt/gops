<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete, Document, EditPen, Refresh, WarningFilled } from '@element-plus/icons-vue'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesNode, drainKubernetesNode, getKubernetesNodeSupport, getKubernetesObject, updateKubernetesNode, type KubernetesEvent, type KubernetesNodeSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0)), name = computed(() => String(route.query.name || ''))
const loading = ref(false), support = ref<KubernetesNodeSupport | null>(null), yaml = ref(''), supportError = ref(''), yamlError = ref('')
const activeTab = ref('overview'), dialog = ref<'describe' | 'yaml' | 'delete' | ''>(''), editorYaml = ref(''), saving = ref(false), deleting = ref(false), deleteConfirm = ref('')
const node = computed(() => support.value?.node)
const events = computed<KubernetesEvent[]>(() => support.value?.events || [])
const pods = computed(() => support.value?.pods || [])
const labels = computed(() => node.value?.labels || {})
const annotations = computed(() => node.value?.annotations || {})
const ready = computed(() => node.value?.status === '就绪')
const conditionCount = computed(() => node.value?.conditions?.length || 0)
const memoryCapacity = computed(() => node.value?.memoryCapacity || '-')
const cpuCapacity = computed(() => node.value?.cpuCapacity || '-')

function unwrapSupport(value: unknown): KubernetesNodeSupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('node' in payload) && 'data' in payload) payload = (payload as { data: unknown }).data
  if (!payload || typeof payload !== 'object' || !('node' in payload)) return null
  const result = payload as KubernetesNodeSupport
  return { ...result, events: result.events || [], pods: result.pods || [] }
}
async function load() {
  supportError.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { supportError.value = '缺少集群 ID 或节点名称。'; return }
  loading.value = true
  try {
    const supportRequest = getKubernetesNodeSupport(clusterID.value, name.value).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.node.name !== name.value) throw new Error('Node support 响应中没有匹配的节点数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => { supportError.value = error?.response?.data?.message || error?.message || 'Node 详情请求失败。'; throw error })
    const [yamlResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'nodes', name: name.value }), supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (yamlResult.status === 'fulfilled') yaml.value = yamlResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = yamlResult.reason?.response?.data?.message || yamlResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}
function openPod(podName: string, namespace: string) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace, name: podName } }) }
function openDialog(mode: typeof dialog.value) { dialog.value = mode }
async function saveYaml() {
  saving.value = true
  try { await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'nodes', name: name.value, yaml: editorYaml.value }); dialog.value = ''; ElMessage.success('节点 YAML 保存成功'); await load() } finally { saving.value = false }
}
async function setSchedulable(schedulable: boolean) {
  if (!node.value) return
  await updateKubernetesNode({ cluster_id: clusterID.value, name: name.value, labels: node.value.labels || {}, schedulable, taints: node.value.taints || [] })
  ElMessage.success(schedulable ? '节点已设为可调度' : '节点已设为不可调度')
  await load()
}
async function drain() {
  try { await ElMessageBox.confirm(`排空节点“${name.value}”会驱逐其上的非 DaemonSet Pod，并将节点设为不可调度。确认继续吗？`, '排空节点', { type: 'warning' }) } catch { return }
  await drainKubernetesNode(clusterID.value, name.value)
  ElMessage.success('节点排空完成')
  await load()
}
async function remove() {
  deleting.value = true
  try { await deleteKubernetesNode(clusterID.value, name.value); dialog.value = ''; ElMessage.success('节点删除成功'); await router.push({ name: 'kubernetes-nodes', query: { cluster_id: clusterID.value } }) } finally { deleting.value = false }
}
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }
function conditionTone(type: string, status: string) {
  if (type.endsWith('Pressure') || type === 'NetworkUnavailable') return status === 'True' ? 'unhealthy' : status === 'False' ? 'healthy' : 'unknown'
  return status === 'True' ? 'healthy' : status === 'False' ? 'unhealthy' : 'unknown'
}
function formatMemory(bytes: number) {
  if (!bytes) return '-'
  const gib = bytes / 1024 ** 3
  return gib >= 1 ? `${gib.toFixed(2)} GiB` : `${(bytes / 1024 ** 2).toFixed(1)} MiB`
}

watch([clusterID, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="node-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Node · Kubernetes</p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="dialog='describe'">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/nodes/update'" :icon="EditPen" @click="dialog='yaml'" :disabled="!yaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/nodes/drain'" @click="drain">排空</el-button><el-button v-if="node" v-permission="'POST /api/v1/kubernetes/nodes/update'" @click="setSchedulable(!node.schedulable)">{{ node.schedulable ? '停止调度' : '恢复调度' }}</el-button><el-button v-permission="'POST /api/v1/kubernetes/nodes/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; dialog='delete'">删除</el-button></div></header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/><el-alert v-if="yamlError" class="detail-alert" type="warning" :title="yamlError" description="节点信息仍可查看；YAML 编辑暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="metrics"><article><span>状态</span><strong :class="ready ? 'healthy' : 'unhealthy'">● {{ node?.status || '-' }}</strong></article><article><span>角色</span><strong>{{ node?.roles || '-' }}</strong></article><article><span>内部 IP</span><strong>{{ node?.internalIP || '-' }}</strong></article><article><span>Pods</span><strong>{{ node?.podCount ?? pods.length }} / {{ node?.podCapacity || '-' }}</strong><small>已分配 / 容量</small></article><article><span>CPU</span><strong>{{ cpuCapacity }}</strong></article><article><span>内存</span><strong>{{ memoryCapacity }}</strong></article></div>
        <div class="overview-grid"><main><section class="panel info-panel"><h2>信息</h2><div class="info-grid"><span>创建时间</span><b>{{ node?.created_at || '-' }} ({{ node?.age || '-' }})</b><span>主机名</span><b>{{ labels['kubernetes.io/hostname'] || name }}</b><span>外部 IP</span><b>{{ node?.externalIP || '-' }}</b><span>Pod CIDR</span><b>{{ node?.podCIDR || '-' }}</b><span>Kubelet 版本</span><b>{{ node?.version || '-' }}</b><span>容器运行时</span><b>{{ node?.containerRuntime || '-' }}</b><span>操作系统镜像</span><b>{{ node?.osImage || '-' }}</b><span>内核版本</span><b>{{ node?.kernelVersion || '-' }}</b><span>架构</span><b>{{ labels['kubernetes.io/arch'] || '-' }}</b><span>内存容量</span><b>{{ node?.memoryCapacity || '-' }}</b><span>CPU 容量</span><b>{{ node?.cpuCapacity || '-' }}</b><span>调度状态</span><b>{{ node?.schedulable ? '可调度' : '不可调度' }}</b><span>Provider ID</span><b>{{ node?.providerID || '-' }}</b></div><div class="uid-row"><span>UID</span><code>{{ node?.uid || '-' }}</code></div></section>
          <section class="panel"><h2>节点状态 ({{ node?.conditions.length || 0 }})</h2><div class="condition-table"><div class="condition-row condition-head"><b>状态</b><b>消息</b><b>原因</b><b>Active</b></div><div v-for="condition in node?.conditions || []" :key="condition.type" class="condition-row"><b :class="conditionTone(condition.type, condition.status)"><i class="status-dot"/>{{ condition.type }}</b><span>{{ condition.message || '-' }}</span><span>{{ condition.reason || '-' }}</span><span>{{ condition.status === 'True' ? 'True' : 'False' }}</span></div><el-empty v-if="!node?.conditions.length" description="暂无状态条件" :image-size="38"/></div></section>
        </main><aside><section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,6)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关 Pods ({{ pods.length }})</h2><div v-for="pod in pods.slice(0,8)" :key="pod.name" class="related-row" @click="openPod(pod.name,pod.namespace)"><b>Pod</b><i class="status-dot"/><span>{{ pod.name }}</span></div><p v-if="!pods.length" class="empty">暂无相关 Pods</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value,key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value,key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section><section class="panel"><h2>污点 ({{ node?.taints.length || 0 }})</h2><div v-for="taint in node?.taints || []" :key="`${taint.key}-${taint.effect}`" class="taint-row"><b>{{ taint.key }}</b><span>{{ taint.value || '-' }}</span><small>{{ taint.effect }}</small></div><p v-if="!node?.taints.length" class="empty">暂无污点</p></section></aside></div>
      </el-tab-pane>
      <el-tab-pane :label="`Pods (${pods.length})`" name="pods"><section class="panel table-panel"><h2>节点 Pods</h2><div class="pod-table"><div class="pod-list-row pod-list-head"><b>Name</b><b>Namespace</b><b>Ready</b><b>Restarts</b><b>Status</b><b>IP</b><b>Created</b></div><div v-for="pod in pods" :key="pod.name" class="pod-list-row"><b class="link" @click="openPod(pod.name,pod.namespace)">{{ pod.name }}</b><span>{{ pod.namespace }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><span>{{ pod.status }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.created_at }} ({{ pod.age }})</span></div><el-empty v-if="!pods.length" description="暂无 Pods"/></div></section></el-tab-pane>
      <el-tab-pane :label="`状态 (${conditionCount})`" name="conditions"><section class="panel"><h2>Node Conditions</h2><div v-for="condition in node?.conditions || []" :key="condition.type" class="condition-card"><header><b :class="conditionTone(condition.type, condition.status)"><i class="status-dot"/>{{ condition.type }}</b><span>{{ condition.status }}</span></header><p>{{ condition.message || '-' }}</p><footer><span>Reason: {{ condition.reason || '-' }}</span><span>Last transition: {{ condition.lastTransitionTime || '-' }}</span></footer></div><el-empty v-if="!node?.conditions.length" description="暂无状态条件"/></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialog" :title="dialog==='describe'?'Node 描述':dialog==='yaml'?'编辑 Node YAML':'删除 Node'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialog==='describe'"><div class="command">kubectl describe node {{ name }}</div><pre class="description">{{ support?.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialog==='yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除节点后该节点将从集群移除，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialog=''">取消</el-button><el-button v-if="dialog==='yaml'" type="primary" :loading="saving" @click="saveYaml">保存</el-button><el-button v-else-if="dialog==='delete'" type="danger" :loading="deleting" :disabled="deleteConfirm!==name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.node-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 9px;font-size:11px}.detail-alert{margin-top:6px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.metrics{display:grid;grid-template-columns:repeat(6,minmax(110px,1fr));gap:9px;margin-bottom:9px}.metrics article,.panel{min-width:0;border:1px solid var(--line);border-radius:7px;background:#fff}.metrics article{display:flex;min-height:72px;flex-direction:column;gap:5px;padding:10px}.metrics article>span,.metrics article>small,.empty,.event-row small{color:var(--text-secondary)}.metrics article>strong{overflow-wrap:anywhere;font-size:15px}.healthy{color:#12a66a}.unhealthy,.warning{color:#d92d20}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{padding:10px 12px}.panel h2{margin:0 0 9px;font-size:12px}.info-grid{display:grid;grid-template-columns:120px minmax(0,1fr) 130px minmax(0,1fr);gap:9px 14px}.info-grid>span{font-weight:600}.info-grid>b{overflow-wrap:anywhere;color:#626b79;font-weight:500}.uid-row{display:grid;grid-template-columns:70px 1fr;gap:10px;padding-top:9px;border-top:1px solid var(--line)}.uid-row code{overflow-wrap:anywhere}.condition-table{overflow:auto}.condition-row{display:grid;grid-template-columns:160px minmax(240px,1.6fr) minmax(120px,1fr) 70px;align-items:center;gap:10px;min-width:660px;padding:9px 5px;border-bottom:1px solid var(--line)}.condition-head,.condition-row> b{font-weight:600}.condition-row>span{color:var(--text-secondary)}.condition-row>span:last-child{text-align:right}.status-dot{display:inline-block;width:6px;height:6px;margin-right:7px;border-radius:50%;background:#12b76a}.related-row{display:grid;grid-template-columns:65px 8px minmax(0,1fr);align-items:center;gap:7px;padding:8px 2px;border-bottom:1px solid var(--line);cursor:pointer}.related-row>b{color:var(--text-secondary);font-weight:500}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:4px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.taint-row{display:grid;grid-template-columns:minmax(120px,1fr) minmax(80px,1fr) 120px;gap:8px;padding:7px 2px;border-bottom:1px solid var(--line)}.taint-row span,.taint-row small{color:var(--text-secondary)}.event-row{display:flex;flex-direction:column;gap:4px;padding:8px 2px;border-bottom:1px solid var(--line)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(180px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-detail small{color:var(--text-secondary)}.table-panel{padding:12px}.pod-table{overflow:auto}.pod-list-row{display:grid;grid-template-columns:minmax(180px,1.5fr) 140px 70px 80px 100px 110px minmax(170px,1.3fr);align-items:center;gap:9px;min-width:1000px;padding:8px 5px;border-bottom:1px solid var(--line)}.pod-list-head{font-weight:600}.link{color:#1683d8;cursor:pointer}.condition-card{padding:9px 4px;border-bottom:1px solid var(--line)}.condition-card header,.condition-card footer{display:flex;align-items:center;justify-content:space-between;gap:10px}.condition-card p{margin:7px 0;color:var(--text-secondary)}.condition-card footer{color:var(--text-secondary);font-size:10px}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:900px){.overview-grid{grid-template-columns:1fr}.metrics{grid-template-columns:repeat(3,minmax(100px,1fr))}}@media(max-width:620px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.metrics{grid-template-columns:repeat(2,minmax(100px,1fr))}.info-grid{grid-template-columns:110px minmax(0,1fr)}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
</style>