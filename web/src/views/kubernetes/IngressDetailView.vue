<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesIngressSupport, getKubernetesObject, type KubernetesEvent, type KubernetesIngressSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref('')
const support = ref<KubernetesIngressSupport>({ name: '', namespace: '', ingress_class: '', hosts: [], addresses: [], rules: [], tls: [], backends: [], events: [], labels: {}, annotations: {}, owner: '', uid: '', resource_version: '', created_at: '', age: '', description: '' })
const supportError = ref(''), yamlError = ref('')
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit'), editorYaml = ref('')
const saving = ref(false), deleting = ref(false), deleteConfirm = ref('')

const events = computed<KubernetesEvent[]>(() => support.value.events || [])
const labels = computed(() => support.value.labels || {})
const annotations = computed(() => support.value.annotations || {})
const relatedCount = computed(() => support.value.backends.length + (support.value.default_backend ? 1 : 0))

function unwrapSupport(value: unknown): KubernetesIngressSupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('rules' in payload) && 'data' in payload) {
    payload = (payload as { data: unknown }).data
  }
  if (!payload || typeof payload !== 'object' || !('rules' in payload)) return null
  const data = payload as KubernetesIngressSupport
  return {
    ...data,
    hosts: data.hosts || [], addresses: data.addresses || [], rules: data.rules || [], tls: data.tls || [],
    backends: data.backends || [], events: data.events || [], labels: data.labels || {}, annotations: data.annotations || {},
  }
}

async function load() {
  supportError.value = ''
  yamlError.value = ''
  if (!clusterID.value || !name.value) {
    supportError.value = '缺少集群 ID 或 Ingress 名称，无法查询详情。'
    return
  }
  loading.value = true
  try {
    const supportRequest = getKubernetesIngressSupport(clusterID.value, namespace.value, name.value).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.name !== name.value) throw new Error('Ingress support 响应中没有匹配的资源数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => {
      supportError.value = error?.response?.data?.message || error?.message || 'Ingress 详情接口请求失败。'
      throw error
    })
    const [detailResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'ingresses', namespace: namespace.value, name: name.value }),
      supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (detailResult.status === 'fulfilled') yaml.value = detailResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = detailResult.reason?.response?.data?.message || detailResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}

function openService(serviceName: string) { void router.push({ name: 'kubernetes-service-detail', query: { cluster_id: clusterID.value, namespace: namespace.value, name: serviceName } }) }
function openDialog(mode: typeof dialogMode.value) { dialogMode.value = mode; dialogVisible.value = true }
function editYaml() { yamlMode.value = 'edit'; editorYaml.value = yaml.value; openDialog('yaml') }
function clone() {
  const object = parse(yaml.value)
  object.metadata = object.metadata || {}
  object.metadata.name = `${name.value}-copy`
  for (const key of ['uid', 'resourceVersion', 'creationTimestamp', 'managedFields', 'ownerReferences']) delete object.metadata[key]
  delete object.status
  editorYaml.value = stringify(object)
  yamlMode.value = 'clone'
  openDialog('yaml')
}
async function saveYaml() {
  saving.value = true
  try {
    const object = parse(editorYaml.value)
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'ingresses', namespace: object?.metadata?.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false
    ElMessage.success(yamlMode.value === 'clone' ? 'Ingress 已克隆' : 'Ingress 已保存')
    if (yamlMode.value === 'clone') await router.replace({ name: 'kubernetes-ingress-detail', query: { cluster_id: clusterID.value, namespace: object?.metadata?.namespace || namespace.value, name: object?.metadata?.name } })
    await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'ingresses', namespace: namespace.value, name: name.value, wait: true })
    ElMessage.success('删除成功')
    await router.push({ name: 'kubernetes-network', query: { cluster_id: clusterID.value } })
  } finally { deleting.value = false }
}
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }

watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="ingress-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="editYaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div></header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/>
    <el-alert v-if="yamlError" class="yaml-warning" type="warning" :title="yamlError" description="Ingress 详情仍可查看；YAML 编辑和克隆暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="overview-grid"><main>
        <section class="panel rules-panel"><h2>路由规则 ({{ support.rules.length }})</h2><div v-for="rule in support.rules" :key="rule.host" class="rule-block"><header><span>Host</span><b>{{ rule.host || '*' }}</b></header><div class="rule-head"><b>路径</b><b>类型</b><b>后端 Service</b><b>端口</b></div><div v-for="path in rule.paths" :key="`${rule.host}-${path.path}-${path.service_name}-${path.service_port}`" class="rule-row"><code>{{ path.path || '/' }}</code><span>{{ path.path_type || '-' }}</span><b class="link" @click="path.service_name && openService(path.service_name)">{{ path.service_name || '-' }}</b><span>{{ path.service_port || '-' }}</span></div><el-empty v-if="!rule.paths.length" description="未配置 HTTP 路径" :image-size="36"/></div><el-empty v-if="!support.rules.length" description="暂无规则" :image-size="42"/></section>
        <section class="panel tls-panel"><h2>TLS ({{ support.tls.length }})</h2><div v-for="(tls,index) in support.tls" :key="`${tls.secret_name}-${index}`" class="tls-row"><div><span>主机</span><b>{{ tls.hosts.join(', ') || '-' }}</b></div><div><span>Secret</span><code>{{ tls.secret_name || '-' }}</code></div></div><el-empty v-if="!support.tls.length" description="未配置 TLS" :image-size="38"/></section>
        <section class="panel ingress-info"><h2>信息</h2><div class="info-summary"><div><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b></div><div><span>Ingress Class</span><b>{{ support.ingress_class || '-' }}</b></div></div><div class="info-grid"><span>地址</span><b>{{ support.addresses.join(', ') || '-' }}</b><span>默认后端</span><b v-if="support.default_backend" class="link" @click="openService(support.default_backend.service_name)">{{ support.default_backend.service_name }}:{{ support.default_backend.service_port }}</b><b v-else>-</b><span>所有者</span><b>{{ support.owner || '-' }}</b><span>资源版本</span><b>{{ support.resource_version || '-' }}</b><span>UID</span><b>{{ support.uid || '-' }}</b></div></section>
      </main><aside>
        <section class="panel"><h2>事件 ({{ support.events.length }})</h2><div v-for="event in support.events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!support.events.length" class="empty">暂无最近事件</p></section>
        <section class="panel related-panel"><h2>相关资源 ({{ relatedCount }})</h2><div v-for="backend in support.backends" :key="backend.name" class="related-row" @click="openService(backend.name)"><b>Service</b><i class="status-dot"/><span>{{ backend.name }}</span><small>{{ backend.type }} · {{ backend.cluster_ip }}</small></div><div v-if="support.default_backend && !support.backends.some(item => item.name === support.default_backend?.service_name)" class="related-row" @click="openService(support.default_backend.service_name)"><b>Service</b><i class="status-dot"/><span>{{ support.default_backend.service_name }}</span><small>默认后端</small></div><p v-if="!relatedCount" class="empty">暂无相关资源</p></section>
        <section class="panel"><h2>标签 ({{ Object.keys(support.labels).length }})</h2><code v-for="(value,key) in support.labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(support.labels).length" class="empty">暂无标签</p></section>
        <section class="panel"><h2>注解 ({{ Object.keys(support.annotations).length }})</h2><code v-for="(value,key) in support.annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(support.annotations).length" class="empty">暂无注解</p></section>
      </aside></div></el-tab-pane>
      <el-tab-pane label="规则" name="rules"><section class="panel table-panel"><h2>Ingress Rules</h2><div v-for="rule in support.rules" :key="`detail-${rule.host}`" class="rule-block"><header><span>Host</span><b>{{ rule.host || '*' }}</b></header><div class="rule-head"><b>路径</b><b>类型</b><b>后端 Service</b><b>端口</b></div><div v-for="path in rule.paths" :key="`${rule.host}-${path.path}-${path.service_name}-${path.service_port}`" class="rule-row"><code>{{ path.path || '/' }}</code><span>{{ path.path_type || '-' }}</span><b class="link" @click="path.service_name && openService(path.service_name)">{{ path.service_name || '-' }}</b><span>{{ path.service_port || '-' }}</span></div></div><el-empty v-if="!support.rules.length" description="暂无规则"/></section></el-tab-pane>
      <el-tab-pane label="相关资源" name="related"><section class="panel table-panel"><h2>后端 Service ({{ relatedCount }})</h2><div class="service-table"><div class="service-row service-head"><b>名称</b><b>类型</b><b>Cluster IP</b><b>端口</b></div><div v-for="backend in support.backends" :key="backend.name" class="service-row"><b class="link" @click="openService(backend.name)">{{ backend.name }}</b><span>{{ backend.type || '-' }}</span><span>{{ backend.cluster_ip || '-' }}</span><span>{{ backend.ports || '-' }}</span></div></div><el-empty v-if="!support.backends.length" description="暂无后端 Service"/></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${support.events.length})`" name="events"><section class="panel"><div v-for="event in support.events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!support.events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'describe' ? 'Ingress 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 Ingress' : '编辑 Ingress YAML') : '删除 Ingress'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe ingress -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template>
      <KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/>
      <div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.ingress-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.rule-block{padding:7px 0;border-bottom:1px solid var(--line)}.rule-block>header{display:flex;gap:10px;padding:5px 4px 8px}.rule-block>header span{color:var(--text-secondary);font-weight:600}.rule-head,.rule-row{display:grid;grid-template-columns:minmax(120px,1fr) minmax(100px,.8fr) minmax(180px,1.5fr) 90px;align-items:center;gap:9px;padding:8px 5px;border-bottom:1px solid var(--line)}.rule-head{font-weight:600}.rule-row code{overflow-wrap:anywhere}.link{color:#1683d8;cursor:pointer}.tls-row{display:grid;grid-template-columns:1.5fr 1fr;gap:12px;padding:9px 4px;border-bottom:1px solid var(--line)}.tls-row>div{display:flex;min-width:0;flex-direction:column;gap:5px}.tls-row span,.info-summary span{color:var(--text-secondary);font-weight:600}.tls-row code{overflow-wrap:anywhere}.info-summary{display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:4px 0 10px}.info-summary>div{display:flex;flex-direction:column;gap:5px}.info-grid{display:grid;grid-template-columns:110px minmax(0,1fr) 110px minmax(0,1fr);gap:9px 14px;padding:9px 0;border-top:1px solid var(--line)}.info-grid>span{font-weight:600}.info-grid>b{min-width:0;overflow-wrap:anywhere;color:#626b79;font-weight:500}.event-row{display:flex;flex-direction:column;gap:4px;padding:8px 2px;border-bottom:1px solid var(--line)}.event-row small,.empty{color:var(--text-secondary)}.warning{color:#d92d20}.related-row{display:grid;grid-template-columns:80px 8px minmax(0,1fr);align-items:center;gap:7px;padding:8px 2px;border-bottom:1px solid var(--line);cursor:pointer}.related-row>b{color:var(--text-secondary);font-weight:500}.related-row>span{overflow-wrap:anywhere}.related-row small{grid-column:3;color:var(--text-secondary)}.status-dot{width:6px;height:6px;border-radius:50%;background:#12b76a}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:4px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.service-table{overflow:auto}.service-row{display:grid;grid-template-columns:minmax(180px,1.4fr) 120px minmax(130px,1fr) minmax(140px,1fr);align-items:center;gap:10px;min-width:700px;padding:9px 6px;border-bottom:1px solid var(--line)}.service-head{font-weight:600}.service-row span{overflow-wrap:anywhere}.event-detail{display:grid;grid-template-columns:80px 120px minmax(180px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-detail small{color:var(--text-secondary)}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:900px){.overview-grid{grid-template-columns:1fr}}@media(max-width:600px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.info-summary,.info-grid{grid-template-columns:1fr 1fr}.tls-row{grid-template-columns:1fr}.rule-head,.rule-row{grid-template-columns:minmax(90px,.7fr) minmax(80px,.6fr) minmax(150px,1.4fr) 60px;min-width:520px}.rule-block{overflow:auto}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
</style>