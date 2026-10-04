<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesClusterRoleSupport, getKubernetesObject, type KubernetesClusterRoleSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref('')
const support = ref<KubernetesClusterRoleSupport>({ name: '', created_at: '', age: '', owner: '', uid: '', resource_version: '', labels: {}, annotations: {}, rules: [], related: [], events: [], description: '' })
const error = ref(''), yamlError = ref(''), dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit')
const editorYaml = ref(''), deleteConfirm = ref(''), saving = ref(false), deleting = ref(false), describing = ref(false)
const labels = computed(() => support.value.labels || {})
const annotations = computed(() => support.value.annotations || {})
let loadRequestID = 0

async function fetchSupport() {
  const { data } = await getKubernetesClusterRoleSupport(clusterID.value, name.value)
  return data.data
}
function supplementFromYaml(text: string) {
  const manifest = parse(text) || {}
  if (manifest.kind !== 'ClusterRole' || manifest.metadata?.name !== name.value) return
  const metadata = manifest.metadata || {}
  const rules = (manifest.rules || []).map((rule: any) => ({
    api_groups: rule.apiGroups || [], resources: rule.resources || [], resource_names: rule.resourceNames || [],
    verbs: rule.verbs || [], non_resource_urls: rule.nonResourceURLs || [],
  }))
  support.value = {
    ...support.value,
    created_at: support.value.created_at || metadata.creationTimestamp || '',
    uid: support.value.uid || metadata.uid || '',
    resource_version: support.value.resource_version || metadata.resourceVersion || '',
    labels: Object.keys(support.value.labels || {}).length ? support.value.labels : metadata.labels || {},
    annotations: Object.keys(support.value.annotations || {}).length ? support.value.annotations : metadata.annotations || {},
    rules: support.value.rules.length ? support.value.rules : rules,
  }
}
async function load() {
  const requestID = ++loadRequestID
  error.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { error.value = '缺少集群 ID 或 ClusterRole 名称，无法查询详情。'; return }
  loading.value = true
  yaml.value = ''
  void getKubernetesObject({ cluster_id: clusterID.value, resource: 'clusterroles', name: name.value })
    .then(({ data }) => { if (requestID === loadRequestID) { yaml.value = data.data.yaml; supplementFromYaml(yaml.value) } })
    .catch(cause => { if (requestID === loadRequestID) yamlError.value = cause?.response?.data?.message || cause?.message || 'ClusterRole YAML 请求失败。' })
  try {
    const data = await fetchSupport()
    if (requestID !== loadRequestID) return
    support.value = data
    if (support.value.name !== name.value) throw new Error('ClusterRole 详情响应中没有匹配的资源数据。')
    if (yaml.value) supplementFromYaml(yaml.value)
  } catch (cause: any) {
    if (requestID === loadRequestID) error.value = cause?.response?.data?.message || cause?.message || 'ClusterRole 详情请求失败。'
  } finally { if (requestID === loadRequestID) loading.value = false }
}
async function openDescribe() {
  describing.value = true
  try {
    const data = await fetchSupport()
    if (data.name !== name.value) throw new Error('ClusterRole 描述响应中没有匹配的资源数据。')
    support.value = data
    dialogMode.value = 'describe'
    dialogVisible.value = true
  } catch (cause: any) {
    ElMessage.error(cause?.response?.data?.message || cause?.message || 'ClusterRole 描述请求失败。')
  } finally { describing.value = false }
}
function openYamlEditor() { yamlMode.value = 'edit'; editorYaml.value = yaml.value; dialogMode.value = 'yaml'; dialogVisible.value = true }
function clone() {
  const object = parse(yaml.value) || {}
  object.metadata = object.metadata || {}
  object.metadata.name = `${name.value}-copy`
  for (const key of ['uid', 'resourceVersion', 'creationTimestamp', 'managedFields', 'ownerReferences', 'generation']) delete object.metadata[key]
  delete object.status
  editorYaml.value = stringify(object, { lineWidth: 0 })
  yamlMode.value = 'clone'
  dialogMode.value = 'yaml'
  dialogVisible.value = true
}
function openDeleteDialog() { deleteConfirm.value = ''; dialogMode.value = 'delete'; dialogVisible.value = true }
async function saveYaml() {
  saving.value = true
  try {
    const parsed = parse(editorYaml.value)
    if (parsed?.kind !== 'ClusterRole' || !parsed.metadata?.name || (yamlMode.value === 'edit' && parsed.metadata.name !== name.value)) { ElMessage.error('只能保存当前 ClusterRole；克隆时请填写有效名称。'); return }
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'clusterroles', yaml: editorYaml.value })
    dialogVisible.value = false
    if (yamlMode.value === 'clone') {
      ElMessage.success('ClusterRole 已克隆')
      await router.replace({ name: 'kubernetes-clusterrole-detail', query: { cluster_id: clusterID.value, name: parsed.metadata.name } })
    } else {
      ElMessage.success('ClusterRole 已保存')
      await load()
    }
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'clusterroles', name: name.value, wait: true })
    ElMessage.success('ClusterRole 删除成功')
    await router.push({ name: 'kubernetes-access', query: { cluster_id: clusterID.value } })
  } finally { deleting.value = false }
}
watch([clusterID, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section class="clusterrole-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Cluster scope</p></div></div><div class="actions"><el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button><el-button :icon="Document" :loading="describing" @click="openDescribe">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="openYamlEditor">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="openDeleteDialog">删除</el-button></div></header>
    <el-alert v-if="loading && !support.name" type="info" title="正在加载 ClusterRole 详情…" :closable="false" show-icon/>
    <el-alert v-if="error" type="error" :title="error" :closable="false" show-icon/><el-alert v-if="yamlError" type="warning" :title="yamlError" description="ClusterRole 详情仍可查看；YAML 暂不可用。" :closable="false" show-icon/>
    <nav class="clusterrole-tab-bar" role="tablist"><button role="tab" :aria-selected="activeTab === 'overview'" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">概览</button><button role="tab" :aria-selected="activeTab === 'yaml'" :class="{ active: activeTab === 'yaml' }" @click="activeTab = 'yaml'">YAML</button><button role="tab" :aria-selected="activeTab === 'related'" :class="{ active: activeTab === 'related' }" @click="activeTab = 'related'">Related ({{ support.related.length }})</button><button role="tab" :aria-selected="activeTab === 'events'" :class="{ active: activeTab === 'events' }" @click="activeTab = 'events'">Events ({{ support.events.length }})</button><button role="tab" :aria-selected="activeTab === 'history'" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">History</button></nav>
    <div class="clusterrole-tab-content">
      <section v-if="activeTab === 'overview'" role="tabpanel"><div class="overview-grid">
        <main><section class="panel"><h2>信息</h2><div class="info-summary"><div><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b></div><div><span>所有者</span><b>{{ support.owner || '无' }}</b></div></div><div class="info-grid"><span>资源版本</span><b>{{ support.resource_version || '-' }}</b></div><div class="uid-row"><span>UID</span><code>{{ support.uid || '-' }}</code></div></section></main>
        <aside><section class="panel"><h2>事件 ({{ support.events.length }})</h2><div v-for="event in support.events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><p v-if="!support.events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关资源 ({{ support.related.length }})</h2><div v-for="item in support.related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i class="status-dot"/><span>{{ item.name }}</span></div><p v-if="!support.related.length" class="empty">暂无相关资源</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value, key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value, key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside>
      </div></section>
      <KubernetesYamlEditor v-else-if="activeTab === 'yaml'" :model-value="yaml" min-height="calc(100vh - 210px)" readonly/>
      <section v-else-if="activeTab === 'related'" role="tabpanel" class="panel"><h2>相关资源</h2><div v-for="item in support.related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i class="status-dot"/><span>{{ item.name }}</span><small>{{ item.status }}</small></div><el-empty v-if="!support.related.length" description="暂无相关资源"/></section>
      <section v-else-if="activeTab === 'events'" role="tabpanel" class="panel"><div v-for="event in support.events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><el-empty v-if="!support.events.length" description="暂无事件"/></section>
      <section v-else role="tabpanel" class="panel"><el-empty description="Kubernetes 不保存 ClusterRole 历史版本"/></section>
    </div>
    <el-dialog v-model="dialogVisible" :class="{ 'is-describe': dialogMode === 'describe' }" class="clusterrole-dialog" :title="dialogMode === 'describe' ? 'ClusterRole 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 ClusterRole' : '编辑 ClusterRole YAML') : '删除 ClusterRole'" :width="dialogMode === 'describe' ? 'min(1080px,calc(100vw - 24px))' : 'min(800px,calc(100vw - 32px))'" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe clusterroles {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.clusterrole-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.info-summary{display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:6px 0 12px;border-bottom:1px solid var(--line)}.info-summary>div{display:flex;min-width:0;flex-direction:column;gap:5px}.info-summary span,.empty{color:var(--text-secondary)}.info-summary b,.info-grid b{overflow-wrap:anywhere;font-weight:500}.info-grid{display:grid;grid-template-columns:minmax(120px,1fr) minmax(80px,1fr) minmax(120px,1fr) minmax(80px,1fr);gap:12px;padding:12px 0}.uid-row{display:grid;grid-template-columns:70px minmax(0,1fr);gap:10px;padding-top:10px;border-top:1px solid var(--line)}.uid-row code{overflow-wrap:anywhere}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:3px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.empty{padding:18px 4px;text-align:center}.related-row{display:grid;grid-template-columns:110px 8px minmax(0,1fr) auto;align-items:center;gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.related-row .status-dot{width:7px;height:7px;border-radius:50%;background:#12b76a}.related-row span{overflow-wrap:anywhere}.event-row{display:grid;grid-template-columns:minmax(70px,100px) minmax(0,1fr);gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-row small{grid-column:2;color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(0,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.warning{color:#d92d20}.rule-head,.rule-row{display:grid;grid-template-columns:1.1fr 1.2fr 1.2fr 1fr 1.2fr;gap:10px;min-width:780px;padding:9px 4px;border-bottom:1px solid var(--line)}.rule-head{font-weight:600}.rule-row code{overflow-wrap:anywhere}.rules-panel{overflow:auto}.command{margin-bottom:9px;padding:8px 10px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere;white-space:pre-wrap;font-weight:600}.description{max-height:65vh;min-width:0;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:5px;background:#f5f5f6;color:var(--text);white-space:pre-wrap;overflow-wrap:anywhere;word-break:break-word;font:12px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.clusterrole-dialog.is-describe :deep(.el-dialog__body){max-height:calc(100vh - 140px);overflow:auto}@media(max-width:760px){.detail-header{align-items:flex-start;flex-direction:column}.overview-grid{grid-template-columns:1fr}.info-grid{grid-template-columns:1fr 1fr}.event-detail{grid-template-columns:70px minmax(0,1fr)}.event-detail span:nth-of-type(2),.event-detail small{grid-column:2}.related-row{grid-template-columns:80px 8px minmax(0,1fr)}}
</style>
<style scoped>
.clusterrole-tab-bar{display:flex;align-items:center;gap:20px;min-width:0;overflow-x:auto;border-bottom:1px solid var(--line)}.clusterrole-tab-bar button{flex:none;height:36px;padding:0 2px;border:0;border-bottom:2px solid transparent;background:transparent;color:var(--text-secondary);font:600 11px/1 inherit;cursor:pointer}.clusterrole-tab-bar button.active{border-bottom-color:var(--primary);color:var(--primary)}.clusterrole-tab-content{min-width:0;min-height:0;padding-top:9px}
</style>