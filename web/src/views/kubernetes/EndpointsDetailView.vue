<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesEndpointsSupport, getKubernetesObject, type KubernetesEndpointsSupport, type KubernetesEvent } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0)), namespace = computed(() => String(route.query.namespace || 'default')), name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref(''), supportError = ref(''), yamlError = ref('')
const support = ref<KubernetesEndpointsSupport>({ name: '', namespace: '', addresses: [], not_ready: [], ports: [], pods: [], events: [], labels: {}, annotations: {}, owner: '', uid: '', created_at: '', age: '', description: '' })
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit'), editorYaml = ref('')
const saving = ref(false), deleting = ref(false), deleteConfirm = ref('')
const events = computed<KubernetesEvent[]>(() => support.value.events || [])
const targetPods = computed(() => support.value.pods)
const relatedCount = computed(() => targetPods.value.length + (support.value.owner ? 1 : 0))
const addressCount = computed(() => support.value.addresses.length)
const notReadyCount = computed(() => support.value.not_ready.length)

function unwrapSupport(value: unknown): KubernetesEndpointsSupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('addresses' in payload) && 'data' in payload) payload = (payload as { data: unknown }).data
  if (!payload || typeof payload !== 'object' || !('addresses' in payload)) return null
  const data = payload as KubernetesEndpointsSupport
  return { ...data, addresses: data.addresses || [], not_ready: data.not_ready || [], ports: data.ports || [], pods: data.pods || [], events: data.events || [], labels: data.labels || {}, annotations: data.annotations || {} }
}

async function load() {
  supportError.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { supportError.value = '缺少集群 ID 或 Endpoints 名称。'; return }
  loading.value = true
  try {
    const supportRequest = getKubernetesEndpointsSupport({ cluster_id: clusterID.value, namespace: namespace.value, name: name.value }).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.name !== name.value) throw new Error('Endpoints support 响应中没有匹配的资源数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => { supportError.value = error?.response?.data?.message || error?.message || 'Endpoints 详情请求失败。'; throw error })
    const [yamlResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'endpoints', namespace: namespace.value, name: name.value }), supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (yamlResult.status === 'fulfilled') yaml.value = yamlResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = yamlResult.reason?.response?.data?.message || yamlResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}
function openPod(podName: string) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace: namespace.value, name: podName } }) }
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }
function openDialog(mode: typeof dialogMode.value) { dialogMode.value = mode; dialogVisible.value = true }
function editYaml() { yamlMode.value = 'edit'; editorYaml.value = yaml.value; openDialog('yaml') }
function clone() { const object = parse(yaml.value); object.metadata = object.metadata || {}; object.metadata.name = `${name.value}-copy`; for (const key of ['uid', 'resourceVersion', 'creationTimestamp', 'managedFields', 'ownerReferences']) delete object.metadata[key]; delete object.status; editorYaml.value = stringify(object); yamlMode.value = 'clone'; openDialog('yaml') }
async function saveYaml() {
  saving.value = true
  try {
    const object = parse(editorYaml.value)
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'endpoints', namespace: object?.metadata?.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false; ElMessage.success(yamlMode.value === 'clone' ? 'Endpoints 已克隆' : 'Endpoints 已保存')
    if (yamlMode.value === 'clone') await router.replace({ name: 'kubernetes-endpoints-detail', query: { cluster_id: clusterID.value, namespace: object?.metadata?.namespace || namespace.value, name: object?.metadata?.name } })
    await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try { await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'endpoints', namespace: namespace.value, name: name.value, wait: true }); ElMessage.success('删除成功'); await router.push({ name: 'kubernetes-network', query: { cluster_id: clusterID.value } }) } finally { deleting.value = false }
}
watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="endpoints-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Endpoints · Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="editYaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div></header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/><el-alert v-if="yamlError" class="detail-alert" type="warning" :title="yamlError" description="Endpoints 详情仍可查看；YAML 编辑和克隆暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="overview-grid"><main>
        <section class="panel"><h2>地址 ({{ addressCount + notReadyCount }})</h2><div class="address-table"><div class="address-row address-head"><b>状态</b><b>IP</b><b>Hostname</b><b>目标</b><b>节点</b></div><div v-for="address in support.addresses" :key="`ready-${address.ip}-${address.target_name}`" class="address-row"><b class="ready-label">就绪</b><code>{{ address.ip || '-' }}</code><span>{{ address.hostname || '-' }}</span><span v-if="address.target_kind === 'Pod'" class="link" @click="openPod(address.target_name)">{{ address.target_name }}</span><span v-else>{{ address.target_name || '-' }}</span><span>{{ address.node_name || '-' }}</span></div><div v-for="address in support.not_ready" :key="`notready-${address.ip}-${address.target_name}`" class="address-row"><b class="not-ready-label">未就绪</b><code>{{ address.ip || '-' }}</code><span>{{ address.hostname || '-' }}</span><span v-if="address.target_kind === 'Pod'" class="link" @click="openPod(address.target_name)">{{ address.target_name }}</span><span v-else>{{ address.target_name || '-' }}</span><span>{{ address.node_name || '-' }}</span></div><el-empty v-if="!addressCount && !notReadyCount" description="暂无地址" :image-size="42"/></div></section>
        <section class="panel"><h2>端口 ({{ support.ports.length }})</h2><div class="port-table"><div class="port-row port-head"><b>名称</b><b>协议</b><b>端口</b><b>目标端口</b></div><div v-for="port in support.ports" :key="`${port.name}-${port.port}-${port.protocol}`" class="port-row"><b>{{ port.name || '-' }}</b><span>{{ port.protocol }}</span><span>{{ port.port }}</span><span>{{ port.target_port || '-' }}</span></div><el-empty v-if="!support.ports.length" description="未配置端口" :image-size="38"/></div></section>
        <section class="panel info-panel"><h2>信息</h2><div class="info-grid"><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b><span>所有者</span><b>{{ support.owner || '-' }}</b><span>标签</span><b>{{ Object.keys(support.labels).length }}</b><span>注解</span><b>{{ Object.keys(support.annotations).length }}</b><span>UID</span><code>{{ support.uid || '-' }}</code></div></section>
      </main><aside>
        <section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section>
        <section class="panel related-panel"><h2>相关 Pods ({{ targetPods.length }})</h2><div v-for="pod in targetPods" :key="pod.name" class="related-row" @click="openPod(pod.name)"><b>Pod</b><i class="status-dot" :class="pod.status === 'Running' ? 'ready' : 'not-ready'"/><span>{{ pod.name }}</span></div><p v-if="!targetPods.length" class="empty">暂无关联 Pod</p></section>
        <section class="panel"><h2>标签 ({{ Object.keys(support.labels).length }})</h2><code v-for="(value,key) in support.labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(support.labels).length" class="empty">暂无标签</p></section>
        <section class="panel"><h2>注解 ({{ Object.keys(support.annotations).length }})</h2><code v-for="(value,key) in support.annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(support.annotations).length" class="empty">暂无注解</p></section>
      </aside></div></el-tab-pane>
      <el-tab-pane :label="`地址 (${addressCount + notReadyCount})`" name="addresses"><section class="panel table-panel"><h2>Endpoints Addresses</h2><div class="address-table"><div class="address-row address-head"><b>状态</b><b>IP</b><b>Hostname</b><b>目标</b><b>节点</b></div><div v-for="address in support.addresses" :key="`a-${address.ip}-${address.target_name}`" class="address-row"><b class="ready-label">就绪</b><code>{{ address.ip || '-' }}</code><span>{{ address.hostname || '-' }}</span><span v-if="address.target_kind === 'Pod'" class="link" @click="openPod(address.target_name)">{{ address.target_name }}</span><span v-else>{{ address.target_name || '-' }}</span><span>{{ address.node_name || '-' }}</span></div><div v-for="address in support.not_ready" :key="`nr-${address.ip}-${address.target_name}`" class="address-row"><b class="not-ready-label">未就绪</b><code>{{ address.ip || '-' }}</code><span>{{ address.hostname || '-' }}</span><span v-if="address.target_kind === 'Pod'" class="link" @click="openPod(address.target_name)">{{ address.target_name }}</span><span v-else>{{ address.target_name || '-' }}</span><span>{{ address.node_name || '-' }}</span></div><el-empty v-if="!addressCount && !notReadyCount" description="暂无地址"/></div></section></el-tab-pane>
      <el-tab-pane :label="`Pods (${targetPods.length})`" name="pods"><section class="panel table-panel"><h2>关联 Pods</h2><div class="pod-table"><div class="pod-row pod-head"><b>Name</b><b>Status</b><b>Ready</b><b>Restarts</b><b>IP</b><b>Node</b></div><div v-for="pod in targetPods" :key="pod.name" class="pod-row"><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><el-empty v-if="!targetPods.length" description="暂无关联 Pod"/></div></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'describe' ? 'Endpoints 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 Endpoints' : '编辑 Endpoints YAML') : '删除 Endpoints'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe endpoints -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.endpoints-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-alert{margin-top:6px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.address-table,.port-table{overflow:auto}.address-row{display:grid;grid-template-columns:90px minmax(140px,1.1fr) minmax(120px,1fr) minmax(170px,1.3fr) minmax(130px,1fr);align-items:center;gap:9px;min-width:760px;padding:8px 5px;border-bottom:1px solid var(--line)}.address-head,.port-head{font-weight:600}.address-row code{overflow-wrap:anywhere}.ready-label{color:#12a66a}.not-ready-label{color:#d92d20}.port-row{display:grid;grid-template-columns:minmax(140px,1.5fr) 100px 100px minmax(140px,1fr);align-items:center;gap:10px;padding:8px 5px;border-bottom:1px solid var(--line)}.info-grid{display:grid;grid-template-columns:110px minmax(0,1fr) 110px minmax(0,1fr);gap:9px 14px}.info-grid>span{font-weight:600}.info-grid>b{overflow-wrap:anywhere;color:#626b79;font-weight:500}.link{color:#1683d8;cursor:pointer}.status-dot{width:6px;height:6px;border-radius:50%;background:#12b76a}.status-dot.not-ready{background:#e5484d}.related-row{display:grid;grid-template-columns:70px 8px minmax(0,1fr);align-items:center;gap:8px;padding:8px 2px;border-bottom:1px solid var(--line);cursor:pointer}.related-row>b{color:var(--text-secondary);font-weight:500}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:4px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.event-row{display:flex;flex-direction:column;gap:4px;padding:8px 2px;border-bottom:1px solid var(--line)}.event-row small,.empty{color:var(--text-secondary)}.warning{color:#d92d20}.event-detail{display:grid;grid-template-columns:80px 120px minmax(180px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-detail small{color:var(--text-secondary)}.table-panel{padding:12px}.pod-table{overflow:auto}.pod-row{display:grid;grid-template-columns:minmax(200px,1.6fr) 100px 80px 90px minmax(120px,1fr) minmax(120px,1fr);align-items:center;gap:9px;min-width:850px;padding:8px 5px;border-bottom:1px solid var(--line)}.pod-head{font-weight:600}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:900px){.overview-grid{grid-template-columns:1fr}}@media(max-width:600px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.info-grid{grid-template-columns:100px minmax(0,1fr)}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
</style>