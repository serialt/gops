<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesConfigResourceSupport, getKubernetesObject, type KubernetesConfigResourceSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const resource = computed(() => String(route.meta.kubernetesConfigResource || 'resourcequotas'))
const kindName = computed(() => ({ resourcequotas: 'ResourceQuota', limitranges: 'LimitRange', horizontalpodautoscalers: 'HPA', poddisruptionbudgets: 'PDB' } as Record<string, string>)[resource.value] || '配置资源')
const support = ref<KubernetesConfigResourceSupport>({ kind: '', name: '', namespace: '', created_at: '', age: '', uid: '', resource_version: '', status: '', owner: '', labels: {}, annotations: {}, properties: [], related: [], events: [], description: '' })
const loading = ref(false), yamlLoading = ref(false), describing = ref(false), saving = ref(false), deleting = ref(false)
const activeTab = ref('overview'), yaml = ref(''), error = ref(''), yamlError = ref('')
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), editorYaml = ref(''), deleteConfirm = ref('')
const labels = computed(() => support.value.labels || {})
const annotations = computed(() => support.value.annotations || {})
const related = computed(() => support.value.related || [])
let requestID = 0
const detailRequest = () => ({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, name: name.value })

async function fetchSupport() {
  const { data } = await getKubernetesConfigResourceSupport(detailRequest())
  return data.data
}
async function load() {
  const currentRequest = ++requestID
  error.value = ''; yamlError.value = ''; yaml.value = ''
  if (!clusterID.value || !name.value) { error.value = `缺少集群 ID 或 ${kindName.value} 名称，无法查询详情。`; return }
  loading.value = true
  yamlLoading.value = true
  void getKubernetesObject(detailRequest())
    .then(({ data }) => { if (currentRequest === requestID) yaml.value = data.data.yaml })
    .catch(cause => { if (currentRequest === requestID) yamlError.value = cause?.response?.data?.message || cause?.message || `${kindName.value} YAML 请求失败。` })
    .finally(() => { if (currentRequest === requestID) yamlLoading.value = false })
  try {
    const data = await fetchSupport()
    if (currentRequest !== requestID) return
    if (data.name !== name.value) throw new Error(`${kindName.value} 响应中没有匹配的资源数据。`)
    support.value = data
  } catch (cause: any) {
    if (currentRequest === requestID) error.value = cause?.response?.data?.message || cause?.message || `${kindName.value} 详情请求失败。`
  } finally { if (currentRequest === requestID) loading.value = false }
}
async function openDescribe() {
  describing.value = true
  try {
    const data = await fetchSupport()
    if (data.name !== name.value) throw new Error(`${kindName.value} 描述响应中没有匹配的资源数据。`)
    support.value = data
    dialogMode.value = 'describe'
    dialogVisible.value = true
  } catch (cause: any) {
    ElMessage.error(cause?.response?.data?.message || cause?.message || `${kindName.value} 描述请求失败。`)
  } finally { describing.value = false }
}
function openYamlEditor() { editorYaml.value = yaml.value; dialogMode.value = 'yaml'; dialogVisible.value = true }
function openDeleteDialog() { deleteConfirm.value = ''; dialogMode.value = 'delete'; dialogVisible.value = true }
async function saveYaml() {
  saving.value = true
  try {
    const { parse } = await import('yaml')
    const parsed = parse(editorYaml.value)
    if (parsed?.kind !== support.value.kind || parsed.metadata?.name !== name.value) { ElMessage.error(`只能保存当前 ${kindName.value}，且名称不可修改。`); return }
    await applyKubernetesObject({ ...detailRequest(), yaml: editorYaml.value })
    dialogVisible.value = false
    ElMessage.success(`${kindName.value} 已保存`)
    await load()
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ ...detailRequest(), wait: true })
    ElMessage.success(`${kindName.value} 删除成功`)
    await router.push({ name: 'kubernetes-config', query: { cluster_id: clusterID.value, resource: resource.value } })
  } finally { deleting.value = false }
}
watch([clusterID, namespace, name, resource], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section class="config-resource-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>{{ kindName }} · Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" :loading="loading || yamlLoading" @click="load">刷新</el-button><el-button :icon="Document" :loading="describing" @click="openDescribe">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="openYamlEditor">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="openDeleteDialog">删除</el-button></div></header>
    <el-alert v-if="error" type="error" :title="error" :closable="false" show-icon/><el-alert v-if="yamlError" type="warning" :title="yamlError" :description="`${kindName} 详情仍可查看；YAML 暂不可用。`" :closable="false" show-icon/>
    <nav class="config-resource-tabs" role="tablist"><button role="tab" :aria-selected="activeTab === 'overview'" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">概览</button><button role="tab" :aria-selected="activeTab === 'yaml'" :class="{ active: activeTab === 'yaml' }" @click="activeTab = 'yaml'">YAML</button><button role="tab" :aria-selected="activeTab === 'related'" :class="{ active: activeTab === 'related' }" @click="activeTab = 'related'">Related ({{ related.length }})</button><button role="tab" :aria-selected="activeTab === 'events'" :class="{ active: activeTab === 'events' }" @click="activeTab = 'events'">Events ({{ support.events.length }})</button><button role="tab" :aria-selected="activeTab === 'history'" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">History</button></nav>
    <div class="config-resource-content">
      <section v-if="activeTab === 'overview'" role="tabpanel"><div class="overview-grid"><main><section class="panel"><h2>信息</h2><div class="info-summary"><div><span>状态</span><b>{{ support.status || '-' }}</b></div><div><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b></div></div><div class="property-list"><div v-for="item in support.properties" :key="item.name" class="property-row"><b>{{ item.name }}</b><span>{{ item.value }}</span></div><div class="property-row"><b>资源版本</b><span>{{ support.resource_version || '-' }}</span></div><div class="property-row"><b>UID</b><code>{{ support.uid || '-' }}</code></div></div></section></main>
        <aside><section class="panel"><h2>事件 ({{ support.events.length }})</h2><div v-for="event in support.events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><p v-if="!support.events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关资源 ({{ related.length }})</h2><div v-for="item in related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i class="status-dot"/><span>{{ item.name }}</span><small>{{ item.status }}</small></div><p v-if="!related.length" class="empty">暂无相关资源</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value, key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value, key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside>
      </div></section>
      <KubernetesYamlEditor v-else-if="activeTab === 'yaml'" :model-value="yaml" min-height="calc(100vh - 210px)" readonly/>
      <section v-else-if="activeTab === 'related'" role="tabpanel" class="panel"><h2>相关资源</h2><div v-for="item in related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i class="status-dot"/><span>{{ item.name }}</span><small>{{ item.status }}</small></div><el-empty v-if="!related.length" description="暂无相关资源"/></section>
      <section v-else-if="activeTab === 'events'" role="tabpanel" class="panel"><div v-for="event in support.events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><el-empty v-if="!support.events.length" description="暂无事件"/></section>
      <section v-else role="tabpanel" class="panel"><div class="history-info"><span>当前资源版本</span><b>{{ support.resource_version || '-' }}</b><span>创建时间</span><b>{{ support.created_at || '-' }}</b></div><p>此资源类型不保存可回滚的历史版本。</p></section>
    </div>
    <el-dialog v-model="dialogVisible" :class="{ 'is-describe': dialogMode === 'describe' }" class="config-resource-dialog" :title="dialogMode === 'describe' ? `${kindName} 描述` : dialogMode === 'yaml' ? `编辑 ${kindName} YAML` : `删除 ${kindName}`" :width="dialogMode === 'describe' ? 'min(1080px,calc(100vw - 24px))' : 'min(800px,calc(100vw - 32px))'" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe {{ resource }} -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除后可能影响配额、限额或可用副本，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">保存</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.config-resource-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.config-resource-tabs{display:flex;align-items:center;gap:20px;min-width:0;overflow-x:auto;margin-top:6px;border-bottom:1px solid var(--line)}.config-resource-tabs button{flex:none;height:36px;padding:0 2px;border:0;border-bottom:2px solid transparent;background:transparent;color:var(--text-secondary);font-family:inherit;font-size:11px;font-weight:600;line-height:1;cursor:pointer}.config-resource-tabs button.active{border-bottom-color:var(--primary);color:var(--primary)}.config-resource-content{min-width:0;padding-top:9px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.info-summary{display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:6px 0 12px;border-bottom:1px solid var(--line)}.info-summary>div{display:flex;min-width:0;flex-direction:column;gap:5px}.info-summary span,.empty,.panel p{color:var(--text-secondary)}.info-summary b{overflow-wrap:anywhere;font-weight:500}.property-list{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0 16px;padding:6px 0}.property-row{display:grid;grid-template-columns:minmax(110px,1fr) minmax(0,1.4fr);gap:8px;padding:8px 2px;border-bottom:1px solid var(--line)}.property-row b{font-weight:600}.property-row span,.property-row code{min-width:0;overflow-wrap:anywhere;color:var(--text-secondary)}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:3px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.empty{padding:18px 4px;text-align:center}.related-row{display:grid;grid-template-columns:120px 8px minmax(0,1fr) auto;align-items:center;gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.related-row .status-dot{width:7px;height:7px;border-radius:50%;background:#12b76a}.related-row span{overflow-wrap:anywhere}.event-row{display:grid;grid-template-columns:minmax(70px,100px) minmax(0,1fr);gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-row small{grid-column:2;color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(0,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.warning{color:#d92d20}.history-info{display:grid;grid-template-columns:160px minmax(0,1fr);gap:10px;padding-bottom:10px;border-bottom:1px solid var(--line)}.history-info span{color:var(--text-secondary)}.command{margin-bottom:9px;padding:8px 10px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere;white-space:pre-wrap;font-weight:600}.description{max-height:65vh;min-width:0;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:5px;background:#f5f5f6;color:var(--text);white-space:pre-wrap;overflow-wrap:anywhere;word-break:break-word;font:12px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.config-resource-dialog.is-describe :deep(.el-dialog__body){max-height:calc(100vh - 140px);overflow:auto}@media(max-width:760px){.detail-header{align-items:flex-start;flex-direction:column}.overview-grid{grid-template-columns:1fr}.property-list{grid-template-columns:1fr}.event-detail{grid-template-columns:70px minmax(0,1fr)}.event-detail span:nth-of-type(2),.event-detail small{grid-column:2}}
</style>