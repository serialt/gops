<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesNetworkPolicySupport, getKubernetesObject, type KubernetesEvent, type KubernetesNetworkPolicyPeer, type KubernetesNetworkPolicyRule, type KubernetesNetworkPolicySupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0)), namespace = computed(() => String(route.query.namespace || 'default')), name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref(''), supportError = ref(''), yamlError = ref('')
const support = ref<KubernetesNetworkPolicySupport>({ name: '', namespace: '', pod_selector: {}, pod_selector_text: '', policy_types: [], ingress: [], egress: [], pods: [], events: [], labels: {}, annotations: {}, owner: '', uid: '', created_at: '', age: '', description: '' })
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit'), editorYaml = ref('')
const saving = ref(false), deleting = ref(false), deleteConfirm = ref('')
const events = computed<KubernetesEvent[]>(() => support.value.events || [])
const relatedCount = computed(() => support.value.pods.length + (support.value.owner ? 1 : 0))

function unwrapSupport(value: unknown): KubernetesNetworkPolicySupport | null {
  let payload = value
  while (payload && typeof payload === 'object' && !Array.isArray(payload) && !('ingress' in payload) && 'data' in payload) payload = (payload as { data: unknown }).data
  if (!payload || typeof payload !== 'object' || !('ingress' in payload)) return null
  const data = payload as KubernetesNetworkPolicySupport
  return { ...data, policy_types: data.policy_types || [], ingress: data.ingress || [], egress: data.egress || [], pods: data.pods || [], events: data.events || [], labels: data.labels || {}, annotations: data.annotations || {} }
}
async function load() {
  supportError.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { supportError.value = '缺少集群 ID 或 NetworkPolicy 名称。'; return }
  loading.value = true
  try {
    const supportRequest = getKubernetesNetworkPolicySupport({ cluster_id: clusterID.value, namespace: namespace.value, name: name.value }).then(result => {
      const data = unwrapSupport(result.data)
      if (!data || data.name !== name.value) throw new Error('NetworkPolicy support 响应中没有匹配的资源数据。')
      support.value = data
      loading.value = false
      return data
    }).catch(error => { supportError.value = error?.response?.data?.message || error?.message || 'NetworkPolicy 详情请求失败。'; throw error })
    const [yamlResult, supportResult] = await Promise.allSettled([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'networkpolicies', namespace: namespace.value, name: name.value }), supportRequest,
    ])
    if (supportResult.status === 'rejected') return
    if (yamlResult.status === 'fulfilled') yaml.value = yamlResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = yamlResult.reason?.response?.data?.message || yamlResult.reason?.message || 'YAML 读取失败。' }
  } finally { loading.value = false }
}
function peerText(peer: KubernetesNetworkPolicyPeer) {
  if (peer.ip_block_cidr) return `${peer.ip_block_cidr}${peer.ip_block_except?.length ? `（排除 ${peer.ip_block_except.join(', ')}）` : ''}`
  const selectors = [peer.namespace_selector_text && `Namespace: ${peer.namespace_selector_text}`, peer.pod_selector_text && `Pod: ${peer.pod_selector_text}`].filter(Boolean)
  return selectors.join(' · ') || '任意来源'
}
function portsText(rule: KubernetesNetworkPolicyRule) { return rule.ports.length ? rule.ports.map(port => `${port.protocol}/${port.port}${port.end_port ? `-${port.end_port}` : ''}`).join(', ') : '所有端口' }
function eventTime(event: KubernetesEvent) { return event.lastTimestamp || event.firstTimestamp || '-' }
function openPod(podName: string) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace: namespace.value, name: podName } }) }
function openDialog(mode: typeof dialogMode.value) { dialogMode.value = mode; dialogVisible.value = true }
function editYaml() { yamlMode.value = 'edit'; editorYaml.value = yaml.value; openDialog('yaml') }
function clone() { const object = parse(yaml.value); object.metadata = object.metadata || {}; object.metadata.name = `${name.value}-copy`; for (const key of ['uid', 'resourceVersion', 'creationTimestamp', 'managedFields', 'ownerReferences']) delete object.metadata[key]; delete object.status; editorYaml.value = stringify(object); yamlMode.value = 'clone'; openDialog('yaml') }
async function saveYaml() {
  saving.value = true
  try {
    const object = parse(editorYaml.value)
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'networkpolicies', namespace: object?.metadata?.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false; ElMessage.success(yamlMode.value === 'clone' ? 'NetworkPolicy 已克隆' : 'NetworkPolicy 已保存')
    if (yamlMode.value === 'clone') await router.replace({ name: 'kubernetes-networkpolicy-detail', query: { cluster_id: clusterID.value, namespace: object?.metadata?.namespace || namespace.value, name: object?.metadata?.name } })
    await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try { await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'networkpolicies', namespace: namespace.value, name: name.value, wait: true }); ElMessage.success('删除成功'); await router.push({ name: 'kubernetes-network', query: { cluster_id: clusterID.value } }) } finally { deleting.value = false }
}
watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="network-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>NetworkPolicy · Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="editYaml">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div></header>
    <el-alert v-if="supportError" type="error" :title="supportError" :closable="false" show-icon/><el-alert v-if="yamlError" class="detail-alert" type="warning" :title="yamlError" description="NetworkPolicy 详情仍可查看；YAML 编辑与克隆暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="overview-grid"><main>
        <section class="panel"><h2>目标 Pods ({{ support.pods.length }})</h2><div class="pod-table"><div class="pod-head"><b>Pods</b><b>状态</b><b>就绪</b><b>IP</b><b>节点</b></div><div v-for="pod in support.pods" :key="pod.name" class="pod-row"><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><el-empty v-if="!support.pods.length" description="没有匹配的 Pod" :image-size="42"/></div></section>
        <section class="panel policy-summary"><h2>策略范围</h2><div class="info-grid"><span>Pod Selector</span><code>{{ support.pod_selector_text || '<none>' }}</code><span>策略类型</span><b>{{ support.policy_types.join(', ') || '-' }}</b><span>Ingress 规则</span><b>{{ support.ingress.length }}</b><span>Egress 规则</span><b>{{ support.egress.length }}</b><span>创建时间</span><b>{{ support.created_at }} ({{ support.age }})</b><span>所有者</span><b>{{ support.owner || '-' }}</b><span>UID</span><code>{{ support.uid || '-' }}</code></div></section>
      </main><aside><section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ eventTime(event) }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关资源 ({{ relatedCount }})</h2><div v-for="pod in support.pods" :key="pod.name" class="related-row" @click="openPod(pod.name)"><b>Pod</b><i class="status-dot"/><span>{{ pod.name }}</span></div><div v-if="support.owner" class="related-row"><b>Owner</b><i class="status-dot"/><span>{{ support.owner }}</span></div><p v-if="!relatedCount" class="empty">暂无相关资源</p></section><section class="panel"><h2>标签 ({{ Object.keys(support.labels).length }})</h2><code v-for="(value,key) in support.labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(support.labels).length" class="empty">暂无标签</p></section></aside></div></el-tab-pane>
      <el-tab-pane :label="`规则 (${support.ingress.length + support.egress.length})`" name="rules"><div class="rule-columns"><section class="panel"><h2>Ingress ({{ support.ingress.length }})</h2><article v-for="(rule,index) in support.ingress" :key="`in-${index}`" class="rule-card"><header><b>规则 {{ index + 1 }}</b><span>{{ portsText(rule) }}</span></header><div class="peer-list"><span class="subheading">允许来源</span><div v-for="(peer,peerIndex) in rule.peers" :key="peerIndex" class="peer-row">{{ peerText(peer) }}</div><div v-if="!rule.peers.length" class="peer-row">任意来源</div></div></article><el-empty v-if="!support.ingress.length" description="未配置 Ingress 规则"/></section><section class="panel"><h2>Egress ({{ support.egress.length }})</h2><article v-for="(rule,index) in support.egress" :key="`out-${index}`" class="rule-card"><header><b>规则 {{ index + 1 }}</b><span>{{ portsText(rule) }}</span></header><div class="peer-list"><span class="subheading">允许目标</span><div v-for="(peer,peerIndex) in rule.peers" :key="peerIndex" class="peer-row">{{ peerText(peer) }}</div><div v-if="!rule.peers.length" class="peer-row">任意目标</div></div></article><el-empty v-if="!support.egress.length" description="未配置 Egress 规则"/></section></div></el-tab-pane>
      <el-tab-pane :label="`相关资源 (${support.pods.length})`" name="related"><section class="panel table-panel"><h2>匹配的 Pods</h2><div class="pod-table"><div class="pod-head related-head"><b>Name</b><b>Status</b><b>Ready</b><b>IP</b><b>Node</b></div><div v-for="pod in support.pods" :key="pod.name" class="pod-row related-head"><b class="link" @click="openPod(pod.name)">{{ pod.name }}</b><span>{{ pod.status }}</span><span>{{ pod.ready || '-' }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span></div><el-empty v-if="!support.pods.length" description="没有匹配的 Pod"/></div></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ eventTime(event) }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'describe' ? 'NetworkPolicy 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 NetworkPolicy' : '编辑 NetworkPolicy YAML') : '删除 NetworkPolicy'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe networkpolicy -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.network-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-alert{margin-top:6px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid,.rule-columns{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.pod-head,.pod-row{display:grid;grid-template-columns:minmax(200px,1.7fr) 100px 70px minmax(110px,1fr) minmax(120px,1fr);align-items:center;gap:9px;padding:8px 4px;border-bottom:1px solid var(--line)}.pod-head,.related-head{font-weight:600}.link{color:#1683d8;cursor:pointer}.pod-row code,.info-grid code{overflow-wrap:anywhere}.info-grid{display:grid;grid-template-columns:110px minmax(0,1fr) 110px minmax(0,1fr);gap:9px 14px;padding:4px}.info-grid>span{font-weight:600}.info-grid>b{min-width:0;overflow-wrap:anywhere;color:#626b79;font-weight:500}.rule-columns{grid-template-columns:repeat(2,minmax(0,1fr))}.rule-card{padding:8px 4px;border-bottom:1px solid var(--line)}.rule-card>header{display:flex;justify-content:space-between;gap:10px;padding:4px 0 8px}.rule-card>header span{color:var(--text-secondary)}.peer-list{display:flex;flex-direction:column;gap:5px}.subheading{color:var(--text-secondary);font-weight:600}.peer-row{padding:6px 8px;border:1px solid var(--line);border-radius:5px;background:#fafbfc;overflow-wrap:anywhere}.related-row{display:grid;grid-template-columns:80px 8px minmax(0,1fr);align-items:center;gap:7px;padding:8px 2px;border-bottom:1px solid var(--line)}.related-row>b{color:var(--text-secondary);font-weight:500}.status-dot{width:6px;height:6px;border-radius:50%;background:#12b76a}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:4px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.warning{color:#d92d20}.empty{color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(180px,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-detail small{color:var(--text-secondary)}.table-panel{padding:12px}.command{margin-bottom:9px;font-weight:600}.description{max-height:60vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:6px;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.delete-confirm{display:flex;flex-direction:column;gap:10px}.delete-confirm p{margin:0}@media(max-width:850px){.overview-grid,.rule-columns{grid-template-columns:1fr}}@media(max-width:600px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.info-grid{grid-template-columns:100px minmax(0,1fr)}.event-detail{grid-template-columns:70px 100px minmax(180px,1fr)}.event-detail small{grid-column:2/-1}}
</style>