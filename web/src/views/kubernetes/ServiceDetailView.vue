<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesObject, getKubernetesServiceSupport, type KubernetesEvent, type KubernetesServiceSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref('')
const support = ref<KubernetesServiceSupport>({ name: '', namespace: '', type: '', cluster_ip: '', external_name: '', external_ips: [], selector: {}, labels: {}, annotations: {}, owner: '', ports: [], pods: [], endpoints: [], endpoint_slices: [], events: [], created_at: '', age: '', uid: '', resource_version: '', description: '' })
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit'), editorYaml = ref('')
const saving = ref(false), deleting = ref(false), deleteConfirm = ref('')
const supportError = ref(''), yamlError = ref('')

const labels = computed<Record<string, string>>(() => support.value.labels || {})
const annotations = computed<Record<string, string>>(() => support.value.annotations || {})
const selectorText = computed(() => Object.entries(support.value.selector || {}).map(([key, value]) => `${key}=${value}`).join(', ') || '-')
const events = computed<KubernetesEvent[]>(() => support.value.events || [])
const relatedOwner = computed(() => {
  const [kind, ...ownerName] = support.value.owner.split('/')
  return { kind, name: ownerName.join('/') }
})
const relatedCount = computed(() => support.value.pods.length + (relatedOwner.value.name ? 1 : 0))

function unwrapSupport(value: unknown): KubernetesServiceSupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('pods' in payload) && 'data' in payload) {
    payload = (payload as { data: unknown }).data
  }
  if (!payload || typeof payload !== 'object' || !('pods' in payload)) return null
  const data = payload as KubernetesServiceSupport
  return {
    ...data,
    external_ips: data.external_ips || [], selector: data.selector || {}, labels: data.labels || {}, annotations: data.annotations || {},
    ports: data.ports || [], pods: data.pods || [], endpoints: data.endpoints || [], endpoint_slices: data.endpoint_slices || [], events: data.events || [],
  }
}

async function load() {
  supportError.value = ''
  yamlError.value = ''
  if (!clusterID.value || !name.value) {
    supportError.value = '缺少集群 ID 或 Service 名称，无法查询详情。'
    return
  }
  loading.value = true
  try {
    const supportRequest = getKubernetesServiceSupport(clusterID.value, namespace.value, name.value).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.name !== name.value) throw new Error('Service support 响应中没有匹配的 Service 数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => {
      supportError.value = error?.response?.data?.message || error?.message || 'Service 详情接口请求失败。'
      throw error
    })
    const [detailResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'services', namespace: namespace.value, name: name.value }),
      supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (detailResult.status === 'fulfilled') yaml.value = detailResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = detailResult.reason?.response?.data?.message || detailResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}

function openPod(podName: string) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace: namespace.value, name: podName } }) }
function openDialog(mode: typeof dialogMode.value) { dialogMode.value = mode; dialogVisible.value = true }
function editYaml() { yamlMode.value = 'edit'; editorYaml.value = yaml.value; openDialog('yaml') }
function clone() {
  const object = parse(yaml.value)
  object.metadata = object.metadata || {}
  object.metadata.name = `${name.value}-copy`
  for (const key of ['uid', 'resourceVersion', 'creationTimestamp', 'managedFields', 'ownerReferences']) delete object.metadata[key]
  delete object.status
  object.spec = object.spec || {}
  delete object.spec.clusterIP
  delete object.spec.clusterIPs
  delete object.spec.healthCheckNodePort
  for (const port of object.spec.ports || []) delete port.nodePort
  editorYaml.value = stringify(object)
  yamlMode.value = 'clone'
  openDialog('yaml')
}
async function saveYaml() {
  saving.value = true
  try {
    const object = parse(editorYaml.value)
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'services', namespace: object?.metadata?.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false
    ElMessage.success(yamlMode.value === 'clone' ? 'Service 已克隆' : 'Service 已保存')
    if (yamlMode.value === 'clone') await router.replace({ name: 'kubernetes-service-detail', query: { cluster_id: clusterID.value, namespace: object?.metadata?.namespace || namespace.value, name: object?.metadata?.name } })
    await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'services', namespace: namespace.value, name: name.value, wait: true })
    ElMessage.success('删除成功')
    await router.push({ name: 'kubernetes-network', query: { cluster_id: clusterID.value } })
  } finally { deleting.value = false }
}
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }
watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="service-detail">
    <header class="detail-header">
      <div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Namespace: <b>{{ namespace }}</b></p></div></div>
      <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="editYaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div>
    </header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/>
    <el-alert v-if="yamlError" class="yaml-warning" type="warning" :title="yamlError" description="Service 的后端详情仍可查看；YAML 编辑和克隆暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview">
        <div class="overview-grid">
          <main>
            <section class="panel pods-panel"><h2>Pods ({{ support.pods.length }})</h2><div class="pod-table"><div class="pod-head"><b>Pods</b><b>状态</b><b>就绪</b><b>重启</b><b>节点</b><b>运行时间</b></div><div v-for="pod in support.pods" :key="pod.name" class="pod-row"><span><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><small>{{ pod.pod_ip || '-' }}</small></span><span><i class="status-dot" :class="pod.status === 'Running' ? 'ready' : 'not-ready'"/>{{ pod.status || '-' }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><span>{{ pod.node || '-' }}</span><span>{{ pod.age || '-' }}</span></div><el-empty v-if="!support.pods.length" description="暂无匹配的 Pod" :image-size="46"/></div></section>
            <div class="endpoint-grid">
              <section class="panel"><h2>Endpoints ({{ support.endpoints.length }})</h2><div class="endpoint-row endpoint-head"><b>名称</b><b>就绪地址</b><b>创建时间</b></div><div v-for="endpoint in support.endpoints" :key="endpoint.name" class="endpoint-row"><b>{{ endpoint.name }}</b><span>{{ endpoint.endpoint_count }}</span><span>{{ endpoint.created_at }} ({{ endpoint.age }})</span><div class="endpoint-addresses"><code v-for="address in endpoint.addresses" :key="address">{{ address }}</code><small v-for="address in endpoint.not_ready || []" :key="`not-ready-${address}`">未就绪 {{ address }}</small><span v-if="!endpoint.addresses.length && !endpoint.not_ready?.length">暂无地址</span></div></div><el-empty v-if="!support.endpoints.length" description="暂无 Endpoints" :image-size="38"/></section>
              <section class="panel"><h2>EndpointSlices ({{ support.endpoint_slices.length }})</h2><div class="endpoint-row slice-head"><b>名称</b><b>地址类型</b><b>端点</b></div><div v-for="slice in support.endpoint_slices" :key="slice.name" class="endpoint-row slice-row"><b>{{ slice.name }}</b><span>{{ slice.address_type }}</span><span>{{ slice.endpoint_count }}</span><div class="endpoint-addresses"><code v-for="address in slice.addresses" :key="address">{{ address }}</code><span v-if="!slice.addresses.length">暂无地址</span></div></div><el-empty v-if="!support.endpoint_slices.length" description="暂无 EndpointSlices" :image-size="38"/></section>
            </div>
            <section class="panel service-info"><h2>信息</h2><div class="info-summary"><div><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b></div><div><span>所有者</span><b class="link">{{ support.owner || '-' }}</b></div></div><div class="selector-block"><b>Selector</b><code>{{ selectorText }}</code></div><div class="info-grid"><span>类型</span><b>{{ support.type || '-' }}</b><span>Cluster IP</span><b>{{ support.cluster_ip || '-' }}</b><span>外部名称</span><b>{{ support.external_name || '-' }}</b><span>外部 IP</span><b>{{ support.external_ips.join(', ') || '-' }}</b><span>资源版本</span><b>{{ support.resource_version || '-' }}</b><span>UID</span><b>{{ support.uid || '-' }}</b></div><div class="ports"><h3>端口</h3><div v-for="port in support.ports" :key="`${port.name}-${port.port}-${port.protocol}`" class="port-row"><b>{{ port.name || `${port.port}/${port.protocol}` }}</b><span>{{ port.port }} → {{ port.target_port }}</span><small>{{ port.protocol }}</small><small v-if="port.node_port">NodePort {{ port.node_port }}</small></div><el-empty v-if="!support.ports.length" description="未配置端口" :image-size="36"/></div></section>
          </main>
          <aside>
            <section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section>
            <section class="panel related-panel"><h2>相关资源 ({{ relatedCount }})</h2><div v-for="pod in support.pods" :key="`pod-${pod.name}`" class="related-row" @click="openPod(pod.name)"><b>Pod</b><i class="status-dot"/><span>{{ pod.name }}</span></div><div v-if="relatedOwner.name" class="related-row"><b>{{ relatedOwner.kind || 'Owner' }}</b><i class="status-dot"/><span>{{ relatedOwner.name }}</span></div><p v-if="!relatedCount" class="empty">暂无相关资源</p></section>
            <section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value,key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section>
            <section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value,key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section>
          </aside>
        </div>
      </el-tab-pane>
      <el-tab-pane label="相关资源" name="related"><section class="panel table-panel"><h2>相关资源 ({{ relatedCount }})</h2><div class="wide-table"><div class="related-detail-row related-detail-head"><b>类型</b><b>名称</b><b>状态</b><b>就绪</b><b>IP</b><b>节点</b></div><div v-for="pod in support.pods" :key="`pod-${pod.name}`" class="related-detail-row"><b>Pod</b><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><div v-if="relatedOwner.name" class="related-detail-row"><b>{{ relatedOwner.kind || 'Owner' }}</b><b>{{ relatedOwner.name }}</b><span>关联控制器</span><span>-</span><span>-</span><span>-</span></div><el-empty v-if="!relatedCount" description="暂无相关资源"/></div></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'describe' ? 'Service 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 Service' : '编辑 Service YAML') : '删除 Service'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe service -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template>
      <KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/>
      <div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.service-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.pod-head,.pod-row{display:grid;grid-template-columns:minmax(200px,2fr) .75fr .55fr .55fr .8fr .6fr;align-items:center;gap:8px;padding:7px 4px;border-bottom:1px solid var(--line)}.pod-row>span:first-child{display:flex;min-width:0;flex-direction:column}.pod-row small,.event-row small{color:var(--text-secondary)}.link{color:#1683d8;cursor:pointer}.status-dot{display:inline-block;width:6px;height:6px;margin-right:7px;border-radius:50%;background:#12b76a}.endpoint-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:9px}.endpoint-row{display:grid;grid-template-columns:minmax(0,1.2fr) .7fr minmax(0,1fr);gap:8px;align-items:center;padding:7px 3px;border-bottom:1px solid var(--line)}.endpoint-head{font-weight:600}.endpoint-addresses{grid-column:1/-1;display:flex;flex-wrap:wrap;gap:5px}.endpoint-addresses code,.metadata-entry{max-width:100%;padding:3px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.slice-head,.slice-row{grid-template-columns:minmax(0,1.3fr) .8fr .5fr}.info-summary{display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:3px 0 10px}.info-summary>div{display:flex;flex-direction:column;gap:5px}.info-summary span,.selector-block>b{font-weight:600}.selector-block{display:flex;flex-direction:column;gap:6px;padding:8px 0;border-top:1px solid var(--line)}.selector-block code{align-self:flex-start;padding:3px 7px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.info-grid{display:grid;grid-template-columns:100px minmax(0,1fr) 100px minmax(0,1fr);gap:9px 14px;padding:10px 0;border-top:1px solid var(--line)}.info-grid>span{font-weight:600}.info-grid>b{min-width:0;overflow-wrap:anywhere;color:#626b79;font-weight:500}.ports{padding-top:8px;border-top:1px solid var(--line)}.ports h3{margin:4px 0 8px;font-size:11px}.port-row{display:flex;align-items:center;gap:12px;padding:7px 3px;border-bottom:1px solid var(--line)}.port-row b{color:#1683d8}.port-row small{margin-left:auto;color:var(--text-secondary)}.port-row small+small{margin-left:0}.event-row{display:flex;flex-direction:column;gap:4px;padding:8px 2px;border-bottom:1px solid var(--line)}.event-row .warning,.event-detail .warning{color:#d92d20}.empty{color:var(--text-secondary)}.related-row{display:grid;grid-template-columns:85px 10px minmax(0,1fr);align-items:center;gap:5px;padding:8px 2px;border-bottom:1px solid var(--line)}.related-row>b{color:var(--text-secondary);font-weight:500}.related-row span{overflow-wrap:anywhere}.metadata-entry{display:block;margin:5px 0}.table-panel{padding:13px}.wide-table{overflow:auto}.wide-row{display:grid;grid-template-columns:minmax(220px,2fr) 70px 80px 100px 110px 140px minmax(180px,1.5fr);align-items:center;gap:10px;min-width:1000px;padding:9px 6px;border-bottom:1px solid var(--line)}.wide-head{font-weight:600}.full-endpoint{padding:10px 4px;border-bottom:1px solid var(--line)}.full-endpoint header,.full-endpoint footer,.full-endpoint>div{display:flex;align-items:center;flex-wrap:wrap;gap:8px;padding:5px 0}.full-endpoint header{justify-content:space-between}.full-endpoint header small,.full-endpoint>div>small{color:var(--text-secondary)}.full-endpoint code{padding:3px 6px;border-radius:5px;background:#f5f6f8}.full-endpoint footer span{color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(200px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:900px){.overview-grid{grid-template-columns:1fr}.endpoint-grid{grid-template-columns:1fr}}@media(max-width:620px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.info-summary,.info-grid{grid-template-columns:1fr 1fr}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
.related-detail-row{display:grid;grid-template-columns:95px minmax(220px,1.8fr) 110px 70px 120px 140px;align-items:center;gap:10px;min-width:800px;padding:9px 6px;border-bottom:1px solid var(--line)}.related-detail-head{font-weight:600}.related-detail-row code{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.status-dot.not-ready{background:#e5484d}
</style>