<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, EditPen, Refresh } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesObject, getKubernetesRoleBindingSupport, type KubernetesRoleBindingSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref('')
const support = ref<KubernetesRoleBindingSupport>({ name: '', namespace: '', created_at: '', age: '', owner: '', uid: '', resource_version: '', labels: {}, annotations: {}, role_ref: { api_group: '', kind: '', name: '' }, subjects: [], events: [], description: '' })
const error = ref(''), yamlError = ref(''), dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'delete'>('describe'), yamlMode = ref<'edit' | 'clone'>('edit')
const editorYaml = ref(''), deleteConfirm = ref(''), saving = ref(false), deleting = ref(false), describing = ref(false)
const labels = computed(() => support.value.labels || {})
const annotations = computed(() => support.value.annotations || {})

async function fetchSupport() {
  const { data } = await getKubernetesRoleBindingSupport(clusterID.value, namespace.value, name.value)
  return data.data
}
async function load() {
  error.value = ''; yamlError.value = ''
  if (!clusterID.value || !name.value) { error.value = '缺少集群 ID 或 RoleBinding 名称，无法查询详情。'; return }
  loading.value = true
  try {
    const [supportResult, yamlResult] = await Promise.allSettled([
      fetchSupport(),
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'rolebindings', namespace: namespace.value, name: name.value }),
    ])
    if (supportResult.status === 'rejected') throw supportResult.reason
    support.value = supportResult.value
    if (support.value.name !== name.value) throw new Error('RoleBinding 详情响应中没有匹配的资源数据。')
    if (yamlResult.status === 'fulfilled') yaml.value = yamlResult.value.data.data.yaml
    else { yaml.value = ''; yamlError.value = yamlResult.reason?.response?.data?.message || yamlResult.reason?.message || 'RoleBinding YAML 请求失败。' }
  } catch (cause: any) {
    error.value = cause?.response?.data?.message || cause?.message || 'RoleBinding 详情请求失败。'
  } finally { loading.value = false }
}
async function openDescribe() {
  describing.value = true
  try {
    const data = await fetchSupport()
    if (data.name !== name.value) throw new Error('RoleBinding 描述响应中没有匹配的资源数据。')
    support.value = data
    dialogMode.value = 'describe'
    dialogVisible.value = true
  } catch (cause: any) {
    ElMessage.error(cause?.response?.data?.message || cause?.message || 'RoleBinding 描述请求失败。')
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
    if (parsed?.kind !== 'RoleBinding' || !parsed.metadata?.name || (yamlMode.value === 'edit' && parsed.metadata.name !== name.value)) { ElMessage.error('只能保存当前 RoleBinding；克隆时请填写有效名称。'); return }
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'rolebindings', namespace: parsed.metadata.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false
    if (yamlMode.value === 'clone') {
      ElMessage.success('RoleBinding 已克隆')
      await router.replace({ name: 'kubernetes-rolebinding-detail', query: { cluster_id: clusterID.value, namespace: parsed.metadata.namespace || namespace.value, name: parsed.metadata.name } })
    } else {
      ElMessage.success('RoleBinding 已保存')
      await load()
    }
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'rolebindings', namespace: namespace.value, name: name.value, wait: true })
    ElMessage.success('RoleBinding 删除成功')
    await router.push({ name: 'kubernetes-access', query: { cluster_id: clusterID.value } })
  } finally { deleting.value = false }
}
watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="rolebinding-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" :loading="describing" @click="openDescribe">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" :disabled="!yaml" @click="openYamlEditor">编辑 YAML</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="openDeleteDialog">删除</el-button></div></header>
    <el-alert v-if="error" type="error" :title="error" :closable="false" show-icon/><el-alert v-if="yamlError" type="warning" :title="yamlError" description="RoleBinding 详情仍可查看；YAML 暂不可用。" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="overview-grid">
        <main><section class="panel"><h2>信息</h2><div class="info-summary"><div><span>创建时间</span><b>{{ support.created_at || '-' }} ({{ support.age || '-' }})</b></div><div><span>所有者</span><b>{{ support.owner || '无' }}</b></div></div><div class="info-grid"><span>RoleRef 类型</span><b>{{ support.role_ref.kind || '-' }}</b><span>RoleRef 名称</span><b>{{ support.role_ref.name || '-' }}</b><span>API Group</span><b>{{ support.role_ref.api_group || '-' }}</b><span>Subjects</span><b>{{ support.subjects.length }}</b><span>资源版本</span><b>{{ support.resource_version || '-' }}</b></div><div class="uid-row"><span>UID</span><code>{{ support.uid || '-' }}</code></div></section></main>
        <aside><section class="panel"><h2>Subjects ({{ support.subjects.length }})</h2><div v-for="(subject, index) in support.subjects" :key="`${subject.kind}/${subject.namespace}/${subject.name}/${index}`" class="subject-row"><b>{{ subject.kind }}</b><span>{{ subject.name }}</span><small v-if="subject.namespace">{{ subject.namespace }}</small></div><p v-if="!support.subjects.length" class="empty">暂无 Subjects</p></section><section class="panel"><h2>事件 ({{ support.events.length }})</h2><div v-for="event in support.events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><p v-if="!support.events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value, key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value, key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside>
      </div></el-tab-pane>
      <el-tab-pane :label="`Related (${support.subjects.length})`" name="related"><section class="panel"><h2>Subjects ({{ support.subjects.length }})</h2><div v-for="(subject, index) in support.subjects" :key="`${subject.kind}/${subject.namespace}/${subject.name}/${index}`" class="subject-row"><b>{{ subject.kind }}</b><span>{{ subject.name }}</span><small v-if="subject.namespace">{{ subject.namespace }}</small></div><el-empty v-if="!support.subjects.length" description="暂无关联对象"/></section></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`Events (${support.events.length})`" name="events"><section class="panel"><div v-for="event in support.events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count }} 次 · {{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><el-empty v-if="!support.events.length" description="暂无事件"/></section></el-tab-pane>
      <el-tab-pane label="History" name="history"><section class="panel"><el-empty description="Kubernetes 不保存 RoleBinding 历史版本"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :class="{ 'is-describe': dialogMode === 'describe' }" class="rolebinding-dialog" :title="dialogMode === 'describe' ? 'RoleBinding 描述' : dialogMode === 'yaml' ? (yamlMode === 'clone' ? '克隆 RoleBinding' : '编辑 RoleBinding YAML') : '删除 RoleBinding'" :width="dialogMode === 'describe' ? 'min(1080px,calc(100vw - 24px))' : 'min(800px,calc(100vw - 32px))'" align-center destroy-on-close>
      <template v-if="dialogMode === 'describe'"><div class="command">kubectl describe rolebindings -n {{ namespace }} {{ name }}</div><pre class="description">{{ support.description || '暂无描述信息' }}</pre></template><KubernetesYamlEditor v-else-if="dialogMode === 'yaml'" v-model="editorYaml"/><div v-else class="delete-confirm"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button v-if="dialogMode === 'yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode === 'clone' ? '创建副本' : '保存' }}</el-button><el-button v-else-if="dialogMode === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.rolebinding-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);background:#fff;font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:11px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.info-summary{display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:6px 0 12px;border-bottom:1px solid var(--line)}.info-summary>div{display:flex;min-width:0;flex-direction:column;gap:5px}.info-summary span,.empty{color:var(--text-secondary)}.info-summary b,.info-grid b{overflow-wrap:anywhere;font-weight:500}.info-grid{display:grid;grid-template-columns:minmax(120px,1fr) minmax(80px,1fr) minmax(120px,1fr) minmax(80px,1fr);gap:12px;padding:12px 0}.uid-row{display:grid;grid-template-columns:70px minmax(0,1fr);gap:10px;padding-top:10px;border-top:1px solid var(--line)}.uid-row code{overflow-wrap:anywhere}.metadata-entry{display:block;max-width:100%;margin:5px 0;padding:3px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.empty{padding:18px 4px;text-align:center}.subject-row{display:grid;grid-template-columns:100px minmax(0,1fr) auto;align-items:center;gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.subject-row span{overflow-wrap:anywhere;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.subject-row small{color:var(--text-secondary)}.event-row{display:grid;grid-template-columns:minmax(70px,100px) minmax(0,1fr);gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-row small{grid-column:2;color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:80px 120px minmax(0,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.warning{color:#d92d20}.command{margin-bottom:9px;padding:8px 10px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere;white-space:pre-wrap;font-weight:600}.description{max-height:65vh;min-width:0;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:5px;background:#f5f5f6;color:var(--text);white-space:pre-wrap;overflow-wrap:anywhere;word-break:break-word;font:12px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.rolebinding-dialog.is-describe :deep(.el-dialog__body){max-height:calc(100vh - 140px);overflow:auto}@media(max-width:760px){.detail-header{align-items:flex-start;flex-direction:column}.overview-grid{grid-template-columns:1fr}.info-grid{grid-template-columns:1fr 1fr}.event-detail{grid-template-columns:70px minmax(0,1fr)}.event-detail span:nth-of-type(2),.event-detail small{grid-column:2}}
</style>