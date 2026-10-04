<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, Refresh } from '@element-plus/icons-vue'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, getKubernetesObject, listKubernetesEvents, type KubernetesEvent } from '@/api/kubernetes'

type RelatedResource = { kind: string; name: string; detail: string }
type PVCManifest = {
  metadata?: { name?: string; namespace?: string; uid?: string; resourceVersion?: string; creationTimestamp?: string; labels?: Record<string, string>; annotations?: Record<string, string>; ownerReferences?: Array<{ kind?: string; name?: string }> }
  spec?: { accessModes?: string[]; resources?: { requests?: { storage?: string } }; storageClassName?: string; volumeName?: string; volumeMode?: string; selector?: unknown }
  status?: { phase?: string; capacity?: { storage?: string }; accessModes?: string[] }
}

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const loading = ref(false), activeTab = ref('overview'), yaml = ref(''), manifest = ref<PVCManifest>({}), events = ref<KubernetesEvent[]>([]), error = ref('')
const dialog = ref<'describe' | 'clone' | 'delete' | ''>(''), cloneYaml = ref(''), deleteConfirm = ref(''), saving = ref(false), deleting = ref(false)
const dialogVisible = computed({ get: () => dialog.value !== '', set: value => { if (!value) dialog.value = '' } })
const labels = computed(() => manifest.value.metadata?.labels || {})
const annotations = computed(() => manifest.value.metadata?.annotations || {})
const status = computed(() => manifest.value.status?.phase || '未知')
const capacity = computed(() => manifest.value.status?.capacity?.storage || manifest.value.spec?.resources?.requests?.storage || '-')
const accessModes = computed(() => manifest.value.status?.accessModes || manifest.value.spec?.accessModes || [])
const owner = computed(() => manifest.value.metadata?.ownerReferences?.map(item => `${item.kind}/${item.name}`).join(', ') || '无')
const related = computed<RelatedResource[]>(() => {
  const items: RelatedResource[] = []
  if (manifest.value.spec?.storageClassName) items.push({ kind: 'StorageClass', name: manifest.value.spec.storageClassName, detail: '存储类' })
  if (manifest.value.spec?.volumeName) items.push({ kind: 'PV', name: manifest.value.spec.volumeName, detail: '已绑定卷' })
  return items
})
const createdAt = computed(() => formatTime(manifest.value.metadata?.creationTimestamp))
const age = computed(() => formatAge(manifest.value.metadata?.creationTimestamp))
const description = computed(() => {
  const lines = [
    `Name:          ${name.value}`,
    `Namespace:     ${namespace.value}`,
    `StorageClass:  ${manifest.value.spec?.storageClassName || ''}`,
    `Status:        ${status.value}`,
    `Volume:        ${manifest.value.spec?.volumeName || ''}`,
    `Labels:        ${mapDescription(labels.value)}`,
    `Annotations:   ${mapDescription(annotations.value)}`,
    `Finalizers:    ${(manifest.value as any).metadata?.finalizers?.join(', ') || '<none>'}`,
    `Capacity:      ${capacity.value}`,
    `Access Modes:  ${accessModes.value.join(', ') || '<none>'}`,
    `VolumeMode:    ${manifest.value.spec?.volumeMode || 'Filesystem'}`,
    `Used By:       <not discovered>`,
    '',
    'Events:',
  ]
  if (!events.value.length) lines.push('  <none>')
  else events.value.forEach(event => lines.push(`  ${event.type.padEnd(8)} ${event.reason.padEnd(22)} ${event.count || 1}  ${event.message}`))
  return lines.join('\n')
})

function formatTime(value?: string) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }
function formatAge(value?: string) {
  if (!value) return '-'
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000))
  if (seconds >= 86400) return `${Math.floor(seconds / 86400)} 天前`
  if (seconds >= 3600) return `${Math.floor(seconds / 3600)} 小时前`
  if (seconds >= 60) return `${Math.floor(seconds / 60)} 分钟前`
  return `${seconds} 秒前`
}
function mapDescription(value: Record<string, string>) { const entries = Object.entries(value); return entries.length ? entries.map(([key, val]) => `${key}=${val}`).join(', ') : '<none>' }

async function load() {
  error.value = ''
  if (!clusterID.value || !name.value) { error.value = '缺少集群 ID 或 PVC 名称，无法查询详情。'; return }
  loading.value = true
  try {
    const [objectResult, eventResult] = await Promise.all([
      getKubernetesObject({ cluster_id: clusterID.value, resource: 'persistentvolumeclaims', namespace: namespace.value, name: name.value }),
      listKubernetesEvents(clusterID.value, namespace.value),
    ])
    yaml.value = objectResult.data.data.yaml
    const { parse } = await import('yaml')
    manifest.value = parse(yaml.value) || {}
    events.value = (eventResult.data.data.list || []).filter(event => event.involvedObject?.kind === 'PersistentVolumeClaim' && event.involvedObject?.name === name.value)
  } catch (cause: any) { error.value = cause?.response?.data?.message || cause?.message || 'PVC 详情请求失败。' }
  finally { loading.value = false }
}
function openDescribe() { dialog.value = 'describe' }
async function openClone() {
  try {
    const { parse, stringify } = await import('yaml')
    const copy = parse(yaml.value)
    copy.metadata.name = `${name.value}-copy`
    delete copy.metadata.uid; delete copy.metadata.resourceVersion; delete copy.metadata.creationTimestamp; delete copy.metadata.managedFields; delete copy.status
    cloneYaml.value = stringify(copy)
    dialog.value = 'clone'
  } catch { ElMessage.error('无法解析当前 PVC YAML') }
}
function openDelete() { deleteConfirm.value = ''; dialog.value = 'delete' }
async function saveClone() {
  saving.value = true
  try {
    const { parse } = await import('yaml'); const copy = parse(cloneYaml.value)
    if (copy?.kind !== 'PersistentVolumeClaim' || !copy.metadata?.name) { ElMessage.error('请输入有效的 PVC YAML'); return }
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: 'persistentvolumeclaims', namespace: copy.metadata.namespace || namespace.value, yaml: cloneYaml.value })
    dialog.value = ''; ElMessage.success('PVC 副本创建成功')
  } finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: 'persistentvolumeclaims', namespace: namespace.value, name: name.value, wait: true })
    ElMessage.success('PVC 删除成功')
    await router.push({ name: 'kubernetes-storage', query: { cluster_id: clusterID.value, resource: 'persistentvolumeclaims' } })
  } finally { deleting.value = false }
}
watch([clusterID, namespace, name], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
</script>

<template>
  <section v-loading="loading" class="pvc-detail">
    <header class="detail-header">
      <div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>Namespace: <b>{{ namespace }}</b></p></div></div>
      <div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDescribe">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" :disabled="!yaml" @click="openClone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="openDelete">删除</el-button></div>
    </header>
    <el-alert v-if="error" type="error" :title="error" :closable="false" show-icon/>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview"><div class="overview-grid">
        <main><section class="panel info-panel"><h2>信息</h2><div class="summary"><div><span>状态</span><b class="bound"><i/>{{ status }}</b></div><div><span>容量</span><strong>{{ capacity }}</strong></div><div><span>创建时间</span><b>{{ createdAt }}<small>{{ age }}</small></b></div><div><span>所有者</span><b>{{ owner }}</b></div></div><div class="info-grid"><span>存储类</span><b>{{ manifest.spec?.storageClassName || '-' }}</b><span>卷</span><b>{{ manifest.spec?.volumeName || '-' }}</b><span>访问模式</span><b>{{ accessModes.join(', ') || '-' }}</b><span>卷模式</span><b>{{ manifest.spec?.volumeMode || 'Filesystem' }}</b><span>请求容量</span><b>{{ manifest.spec?.resources?.requests?.storage || '-' }}</b><span>资源版本</span><b>{{ manifest.metadata?.resourceVersion || '-' }}</b></div><div class="uid-row"><span>UID</span><code>{{ manifest.metadata?.uid || '-' }}</code></div></section></main>
        <aside><section class="panel"><h2>事件 ({{ events.length }})</h2><div v-for="event in events.slice(0,5)" :key="`${event.reason}-${event.lastTimestamp}`" class="event-row"><b :class="{ warning: event.type === 'Warning' }">{{ event.reason }}</b><span>{{ event.message }}</span><small>{{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><p v-if="!events.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>相关资源 ({{ related.length }})</h2><div v-for="item in related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i/><span>{{ item.name }}</span></div><p v-if="!related.length" class="empty">暂无相关资源</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value,key) in labels" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value,key) in annotations" :key="key" class="metadata-entry">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside>
      </div></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane :label="`关联资源 (${related.length})`" name="related"><section class="panel related-panel"><h2>关联资源</h2><div v-for="item in related" :key="`${item.kind}/${item.name}`" class="related-row"><b>{{ item.kind }}</b><i/><span>{{ item.name }}</span><small>{{ item.detail }}</small></div><el-empty v-if="!related.length" description="暂无相关资源"/></section></el-tab-pane>
      <el-tab-pane :label="`事件 (${events.length})`" name="events"><section class="panel"><div v-for="event in events" :key="`${event.reason}-${event.lastTimestamp}`" class="event-detail"><b :class="{ warning: event.type === 'Warning' }">{{ event.type }}</b><span>{{ event.reason }}</span><span>{{ event.message }}</span><small>{{ event.count || 1 }} 次 · {{ event.lastTimestamp || event.firstTimestamp || '-' }}</small></div><el-empty v-if="!events.length" description="暂无事件"/></section></el-tab-pane>
      <el-tab-pane label="历史" name="history"><section class="panel history"><div><span>当前资源版本</span><b>{{ manifest.metadata?.resourceVersion || '-' }}</b></div><div><span>创建时间</span><b>{{ createdAt }}</b></div><p>PersistentVolumeClaim 不保存可回滚的历史版本。</p></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialog === 'describe' ? 'PVC 描述' : dialog === 'clone' ? '克隆 PVC' : '删除 PVC'" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialog === 'describe'"><div class="command">kubectl describe persistentvolumeclaims -n {{ namespace }} {{ name }}</div><pre class="description">{{ description }}</pre></template><KubernetesYamlEditor v-else-if="dialog === 'clone'" v-model="cloneYaml"/><div v-else class="delete-confirm"><p>删除 PVC 可能导致数据不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialog = ''">取消</el-button><el-button v-if="dialog === 'clone'" type="primary" :loading="saving" @click="saveClone">创建副本</el-button><el-button v-else-if="dialog === 'delete'" type="danger" :loading="deleting" :disabled="deleteConfirm !== name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.pvc-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;background:#fff;color:var(--text);font-size:11px}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:18px}.identity p{margin:3px 0 0;color:var(--text-secondary)}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:9px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 11px;font-size:11px}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(250px,1fr);gap:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{min-width:0;padding:10px 12px;border:1px solid var(--line);border-radius:7px;background:#fff}.panel h2{margin:0 0 9px;font-size:12px}.summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px;padding:6px 0 12px;border-bottom:1px solid var(--line)}.summary>div{display:flex;min-width:0;flex-direction:column;gap:5px}.summary span,.empty,.history p{color:var(--text-secondary)}.summary b,.summary strong{font-weight:500;overflow-wrap:anywhere}.summary small{display:block;margin-top:3px;color:var(--text-secondary)}.bound{color:#079455}.bound i,.related-row i{display:inline-block;width:7px;height:7px;margin-right:6px;border-radius:50%;background:#12b76a}.info-grid{display:grid;grid-template-columns:100px minmax(0,1fr) 100px minmax(0,1fr);gap:12px;padding:13px 0}.info-grid b{font-weight:500;overflow-wrap:anywhere}.uid-row{display:grid;grid-template-columns:70px minmax(0,1fr);gap:10px;padding-top:10px;border-top:1px solid var(--line)}.uid-row code{overflow-wrap:anywhere}.metadata-entry{display:block;margin:5px 0;padding:3px 6px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;overflow-wrap:anywhere}.empty{padding:18px 4px;text-align:center}.related-row{display:grid;grid-template-columns:100px 8px minmax(0,1fr) auto;align-items:center;gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.related-row i{margin:0}.related-row span{overflow-wrap:anywhere}.related-row small{color:var(--text-secondary)}.event-row{display:grid;grid-template-columns:minmax(70px,100px) minmax(0,1fr);gap:8px;padding:9px 4px;border-bottom:1px solid var(--line)}.event-row small{grid-column:2;color:var(--text-secondary)}.event-detail{display:grid;grid-template-columns:70px 120px minmax(0,1fr) 180px;gap:9px;padding:9px 4px;border-bottom:1px solid var(--line)}.warning{color:#d92d20}.history{display:flex;gap:30px;align-items:flex-start}.history div{display:flex;flex-direction:column;gap:5px}.command{margin-bottom:9px;padding:8px 10px;border:1px solid var(--line);border-radius:5px;background:#f8f9fb;font-weight:600}.description{max-height:65vh;margin:0;padding:12px;overflow:auto;border:1px solid var(--line);border-radius:5px;background:#fff;white-space:pre-wrap;overflow-wrap:anywhere;font:12px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}@media(max-width:800px){.detail-header{align-items:flex-start;flex-direction:column}.overview-grid{grid-template-columns:1fr}.summary{grid-template-columns:1fr 1fr}.info-grid{grid-template-columns:90px minmax(0,1fr)}.event-detail{grid-template-columns:70px minmax(0,1fr)}.event-detail span:nth-of-type(2),.event-detail small{grid-column:2}}
</style>
