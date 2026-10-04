<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesNamespaceSupport, getKubernetesObject, type KubernetesEvent, type KubernetesNamespaceSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref(''), supportError = ref(''), yamlError = ref('')
const support = ref<KubernetesNamespaceSupport>({ name: '', status: '', created_at: '', age: '', resource_version: '', uid: '', labels: {}, annotations: {}, resources: [], pods: [], events: [], description: '' })
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), editorYaml = ref(''), saving = ref(false), deleting = ref(false), deleteConfirm = ref('')

const events = computed<KubernetesEvent[]>(() => support.value.events || [])
const resources = computed(() => support.value.resources || [])
const pods = computed(() => support.value.pods || [])
const labels = computed(() => support.value.labels || {})
const annotations = computed(() => support.value.annotations || {})
const resourceCounts = computed(() => [...new Set(resources.value.map(item => item.kind))].map(kind => ({ kind, count: resources.value.filter(item => item.kind === kind).length })))

function unwrapSupport(value: unknown): KubernetesNamespaceSupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('resources' in payload) && 'data' in payload) payload = (payload as { data: unknown }).data
  if (!payload || typeof payload !== 'object' || !('resources' in payload)) return null
  const data = payload as KubernetesNamespaceSupport
  return { ...data, resources: data.resources || [], pods: data.pods || [], events: data.events || [], labels: data.labels || {}, annotations: data.annotations || {} }
}

async function load() {
  supportError.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { supportError.value = '缺少集群 ID 或 Namespace 名称。'; return }
  loading.value = true
  try {
    const supportRequest = getKubernetesNamespaceSupport(clusterID.value, name.value).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.name !== name.value) throw new Error('Namespace support 响应中没有匹配的资源数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => { supportError.value = error?.response?.data?.message || error?.message || 'Namespace 详情请求失败。'; throw error })
    const [yamlResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'namespaces', name: name.value }), supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (yamlResult.status === 'fulfilled') yaml.value = yamlResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = yamlResult.reason?.response?.data?.message || yamlResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}

function detailRoute(kind: string) {
  return ({ Pod: 'kubernetes-pod-detail', Service: 'kubernetes-service-detail', Ingress: 'kubernetes-ingress-detail', NetworkPolicy: 'kubernetes-networkpolicy-detail', Deployment: 'kubernetes-deployment-detail', StatefulSet: 'kubernetes-statefulset-detail', DaemonSet: 'kubernetes-daemonset-detail', Job: 'kubernetes-job-detail', CronJob: 'kubernetes-cronjob-detail' } as Record<string, string>)[kind]
}
function openResource(item: { kind: string; name: string }) {
  const routeName = detailRoute(item.kind)
  if (routeName) void router.push({ name: routeName, query: { cluster_id: clusterID.value, namespace: name.value, name: item.name } })
}
function openPod(podName: string) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace: name.value, name: podName } }) }
function openDialog(mode: typeof dialogMode.value) { dialogMode.value = mode; dialogVisible.value = true }
function editYaml() { editorYaml.value = yaml.value; openDialog('yaml') }
async function saveYaml() {
  saving.value = true
  try {
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'namespaces', name: name.value, yaml: editorYaml.value })
    dialogVisible.value = false; ElMessage.success('Namespace YAML 保存成功'); await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'namespaces', name: name.value })
    ElMessage.success('Namespace 删除成功')
    await router.push({ name: 'kubernetes-namespaces', query: { cluster_id: clusterID.value } })
  } finally { deleting.value = false }
}
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }

watch([clusterID, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="namespace-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Namespace · Kubernetes</p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="editYaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div></header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/><el-alert v-if="yamlError" class="detail-alert" type="warning" :title="yamlError" description="Namespace 详情仍可查看；YAML 编辑暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="metrics"><article><span>状态</span><strong :class="support.status.toLowerCase() === 'active' ? 'healthy' : 'unhealthy'">● {{ support.status || '-' }}</strong></article><article><span>资源</span><strong>{{ resources.length }}</strong><small>命名空间内</small></article><article><span>Pods</span><strong>{{ pods.length }}</strong><small>当前数量</small></article><article><span>事件</span><strong>{{ events.length }}</strong><small>当前记录</small></article><article><span>运行时间</span><strong>{{ support.age || '-' }}</strong></article></div>
        <div class="overview-grid"><main><section class="panel"><h2>信息</h2><div class="info-grid"><span>创建时间</span><b>{{ support.created_at || '-' }}</b><span>资源版本</span><b>{{ support.resource_version || '-' }}</b><span>UID</span><code>{{ support.uid || '-' }}</code><span>标签数量</span><b>{{ Object.keys(labels).length }}</b></div></section><section class="panel"><h2>资源概览</h2><div class="resource-counts"><div v-for="item in resourceCounts" :key="item.kind" class="count-row"><b>{{ item.kind }}</b><span>{{ item.count }}</span></div><el-empty v-if="!resourceCounts.length" description="暂无命名空间资源" :image-size="40"/></div></section><section class="panel"><h2>Pods ({{ pods.length }})</h2><div class="pod-table"><div class="pod-head"><b>Name</b><b>Status</b><b>Ready</b><b>Restarts</b><b>IP</b><b>Node</b></div><div v-for="pod in pods.slice(0,8)" :key="pod.name" class="pod-row"><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><el-empty v-if="!pods.length" description="暂无 Pods" :image-size="40"/></div></section></main>
          <aside><section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,6)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关资源 ({{ resources.length }})</h2><div v-for="item in resources.slice(0,10)" :key="`${item.kind}-${item.name}`" class="related-row" :class="{ clickable: detailRoute(item.kind) }" @click="openResource(item)"><b>{{ item.kind }}</b><i class="resource-dot"/><span>{{ item.name }}</span><small>{{ item.status || item.age }}</small></div><p v-if="!resources.length" class="empty">暂无相关资源</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value,key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value,key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside></div>
      </el-tab-pane>
      <el-tab-pane :label="`相关资源 (${resources.length})`" name="related"><section class="panel table-panel"><h2>Namespace Resources</h2><div class="resource-table"><div class="resource-row resource-head"><b>类型</b><b>名称</b><b>状态</b><b>运行时间</b></div><div v-for="item in resources" :key="`${item.kind}-${item.name}`" class="resource-row"><b>{{ item.kind }}</b><b :class="{ link: detailRoute(item.kind) }" @click="openResource(item)">{{ item.name }}</b><span>{{ item.status || '-' }}</span><span>{{ item.age || '-' }}</span></div></div><el-empty v-if="!resources.length" description="暂无相关资源"/></section></el-tab-pane>
      <el-tab-pane :label="`Pods (${pods.length})`" name="pods"><section class="panel table-panel"><h2>Pods</h2><div class="pod-table"><div class="pod-row pod-head"><b>Name</b><b>Status</b><b>Ready</b><b>Restarts</b><b>IP</b><b>Node</b></div><div v-for="pod in pods" :key="pod.name" class="pod-row"><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><el-empty v-if="!pods.length" description="暂无 Pods"/></div></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'describe' ? 'Namespace 描述' : dialogMode === 'yaml' ? '编辑 Namespace YAML' : '删除 Namespace'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe namespace {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除命名空间会同时删除其中资源。请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">保存</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.namespace-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-alert{margin-top:6px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.metrics{display:grid;grid-template-columns:repeat(5,minmax(110px,1fr));gap:9px;margin-bottom:9px}.metrics article,.panel{min-width:0;border:1px solid var(--line);border-radius:7px;background:#fff}.metrics article{display:flex;min-height:72px;flex-direction:column;gap:5px;padding:10px}.metrics article>span,.metrics article>small,.empty,.event-row small{color:var(--text-secondary)}.metrics article>strong{font-size:15px}.healthy{color:#12a66a}.unhealthy,.warning{color:#d92d20}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{padding:10px 12px}.panel h2{margin:0 0 9px;font-size:12px}.info-grid{display:grid;grid-template-columns:110px minmax(0,1fr) 110px minmax(0,1fr);gap:9px 14px}.info-grid>span{font-weight:600}.info-grid>b,.info-grid code{min-width:0;overflow-wrap:anywhere;color:#626b79;font-weight:500}.resource-counts{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0 14px}.count-row,.related-row{display:grid;grid-template-columns:minmax(100px,1fr) auto;align-items:center;gap:8px;padding:8px 4px;border-bottom:1px solid var(--line)}.count-row span{font-weight:700}.pod-table,.resource-table{overflow:auto}.pod-row,.resource-row{display:grid;align-items:center;gap:9px;min-width:780px;padding:8px 5px;border-bottom:1px solid var(--line)}.pod-row{grid-template-columns:minmax(200px,1.6fr) 100px 75px 85px minmax(120px,1fr) minmax(120px,1fr)}.resource-row{grid-template-columns:140px minmax(200px,1.6fr) 130px 180px}.pod-head,.resource-head{font-weight:600}.related-row{grid-template-columns:85px 8px minmax(0,1fr)}.related-row>b{color:var(--text-secondary);font-weight:500}.related-row small{grid-column:3;color:var(--text-secondary)}.clickable{cursor:pointer}.resource-dot{width:6px;height:6px;border-radius:50%;background:#12b76a}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:4px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.link{color:#1683d8;cursor:pointer}.event-row{display:flex;flex-direction:column;gap:4px;padding:8px 2px;border-bottom:1px solid var(--line)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(180px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-detail small{color:var(--text-secondary)}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:900px){.overview-grid{grid-template-columns:1fr}.metrics{grid-template-columns:repeat(3,minmax(100px,1fr))}}@media(max-width:600px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.metrics{grid-template-columns:repeat(2,minmax(100px,1fr))}.info-grid{grid-template-columns:100px minmax(0,1fr)}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
</style>