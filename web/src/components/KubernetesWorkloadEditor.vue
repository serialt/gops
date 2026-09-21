<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import KubernetesContainerConfig from '@/components/KubernetesContainerConfig.vue'

type Pair = { key: string; value: string }
type Container = Record<string, any> & { name?: string; image?: string; imagePullPolicy?: string; command?: string[]; args?: string[]; env?: Array<{ name: string; value?: string }> }

const props = defineProps<{ modelValue: boolean; yaml: string; saving?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; save: [yaml: string] }>()
const visible = computed({ get: () => props.modelValue, set: value => emit('update:modelValue', value) })
const activeTab = ref('containers')
const editorRoot = ref<HTMLElement>()
const sidebarWidth = ref(250)
const sidebarCollapsed = ref(false)
const schedulingMode = ref<'any'|'node'|'selector'>('any')
let dragStartX = 0
let dragStartWidth = 0
const manifest = reactive<Record<string, any>>({})
const labels = ref<Pair[]>([])
const annotations = ref<Pair[]>([])
const nodeSelectors = ref<Pair[]>([])

const kind = computed(() => String(manifest.kind || 'Workload'))
const metadata = computed(() => manifest.metadata ?? (manifest.metadata = {}))
const workloadSpec = computed(() => manifest.spec ?? (manifest.spec = {}))
const podSpec = computed<Record<string, any>>(() => {
  if (kind.value === 'Pod') return workloadSpec.value
  if (kind.value === 'CronJob') return workloadSpec.value.jobTemplate?.spec?.template?.spec ?? (workloadSpec.value.jobTemplate = { spec: { template: { spec: {} } } }).spec.template.spec
  return workloadSpec.value.template?.spec ?? (workloadSpec.value.template = { metadata: {}, spec: {} }).spec
})
const containers = computed<Container[]>(() => podSpec.value.containers ?? (podSpec.value.containers = []))
const initContainers = computed<Container[]>(() => podSpec.value.initContainers ?? (podSpec.value.initContainers = []))
const volumes = computed<any[]>(() => podSpec.value.volumes ?? (podSpec.value.volumes = []))
const supportsReplicas = computed(() => ['Deployment', 'StatefulSet'].includes(kind.value))

function pairs(value: Record<string, unknown> | undefined): Pair[] { return Object.entries(value || {}).map(([key, item]) => ({ key, value: String(item) })) }
function objectFromPairs(items: Pair[]) { return Object.fromEntries(items.filter(item => item.key.trim()).map(item => [item.key.trim(), item.value])) }
function replaceManifest(value: Record<string, any>) { Object.keys(manifest).forEach(key => delete manifest[key]); Object.assign(manifest, value) }
function initialize() {
  try { replaceManifest(parse(props.yaml) || {}) } catch { replaceManifest({}) }
  if (kind.value === 'Deployment') workloadSpec.value.strategy ??= { type: 'RollingUpdate' }
  if (kind.value === 'StatefulSet' || kind.value === 'DaemonSet') workloadSpec.value.updateStrategy ??= { type: 'RollingUpdate' }
  labels.value = pairs(metadata.value.labels)
  annotations.value = pairs(metadata.value.annotations)
  nodeSelectors.value = pairs(podSpec.value.nodeSelector)
  schedulingMode.value = podSpec.value.nodeName ? 'node' : Object.keys(podSpec.value.nodeSelector || {}).length ? 'selector' : 'any'
  activeTab.value = 'containers'
}
function addPair(target: Pair[]) { target.push({ key: '', value: '' }) }
function addContainer(target: Container[]) { target.push({ name: '', image: '', imagePullPolicy: 'IfNotPresent', command: [], args: [], env: [] }) }
function commandText(container: Container, field: 'command' | 'args') { return (container[field] || []).join(' ') }
function setCommand(container: Container, field: 'command' | 'args', value: string) { container[field] = value.trim() ? value.trim().split(/\s+/) : [] }
function addVolume() { volumes.value.push({ name: '', emptyDir: {} }) }
function volumeType(volume: Record<string, unknown>) { return Object.keys(volume).find(key => key !== 'name') || 'emptyDir' }
function volumeTypeLabel(volume: Record<string, unknown>) { return ({ persistentVolumeClaim:'PVC', configMap:'ConfigMap', secret:'Secret', emptyDir:'EmptyDir', hostPath:'HostPath' } as Record<string,string>)[volumeType(volume)] || volumeType(volume) }
function setVolumeType(volume: Record<string, any>, type: string) {
  for (const key of ['persistentVolumeClaim','configMap','secret','emptyDir','hostPath']) delete volume[key]
  if (type === 'persistentVolumeClaim') volume[type] = { claimName: '', readOnly: false }
  else if (type === 'configMap') volume[type] = { name: '', optional: false }
  else if (type === 'secret') volume[type] = { secretName: '', optional: false }
  else if (type === 'hostPath') volume[type] = { path: '', type: 'DirectoryOrCreate' }
  else volume.emptyDir = {}
}
function podList(name: string) { return podSpec.value[name] ?? (podSpec.value[name] = []) }
function dnsList(name: string) { podSpec.value.dnsConfig ??= {}; return podSpec.value.dnsConfig[name] ?? (podSpec.value.dnsConfig[name] = []) }
function changeSchedulingMode(mode: 'any'|'node'|'selector') { schedulingMode.value = mode; if(mode!=='node')delete podSpec.value.nodeName; if(mode!=='selector'){nodeSelectors.value=[];delete podSpec.value.nodeSelector} }
function save() {
  metadata.value.labels = objectFromPairs(labels.value)
  metadata.value.annotations = objectFromPairs(annotations.value)
  podSpec.value.nodeSelector = objectFromPairs(nodeSelectors.value)
  if (!Object.keys(metadata.value.labels).length) delete metadata.value.labels
  if (!Object.keys(metadata.value.annotations).length) delete metadata.value.annotations
  if (!Object.keys(podSpec.value.nodeSelector).length) delete podSpec.value.nodeSelector
  emit('save', stringify(manifest, { lineWidth: 0 }))
}
function stopSidebarResize() {
  window.removeEventListener('pointermove', resizeSidebar)
  window.removeEventListener('pointerup', stopSidebarResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}
function resizeSidebar(event: PointerEvent) {
  const maxWidth = Math.min(420, Math.max(220, (editorRoot.value?.clientWidth || 900) - 420))
  sidebarWidth.value = Math.min(maxWidth, Math.max(190, dragStartWidth + event.clientX - dragStartX))
}
function startSidebarResize(event: PointerEvent) {
  if (sidebarCollapsed.value) return
  dragStartX = event.clientX
  dragStartWidth = sidebarWidth.value
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', resizeSidebar)
  window.addEventListener('pointerup', stopSidebarResize)
}
function toggleSidebar() { sidebarCollapsed.value = !sidebarCollapsed.value }

watch(() => [props.modelValue, props.yaml] as const, ([open]) => { if (open) initialize() }, { immediate: true })
onBeforeUnmount(stopSidebarResize)
</script>

<template>
  <el-dialog v-model="visible" title="编辑工作负载" width="min(1040px,calc(100vw - 32px))" class="workload-editor-dialog" align-center :close-on-click-modal="false" destroy-on-close>
    <div ref="editorRoot" class="workload-editor" :class="{ 'sidebar-collapsed': sidebarCollapsed }" :style="{ gridTemplateColumns: sidebarCollapsed ? '0 2px minmax(0,1fr)' : `${sidebarWidth}px 2px minmax(0,1fr)` }">
      <aside class="basic-panel">
        <header><span>▣</span><strong>基础信息</strong></header>
        <div class="basic-content">
          <label>名称<el-input v-model="metadata.name" disabled/></label>
          <label>命名空间<el-input v-model="metadata.namespace" disabled/></label>
          <label v-if="supportsReplicas">副本数<el-input-number v-model="workloadSpec.replicas" :min="0" :max="1000"/></label>
          <section class="pair-section"><div><strong>标签</strong><el-button link :icon="Plus" @click="addPair(labels)">添加</el-button></div><div v-for="(item,index) in labels" :key="index" class="pair-row"><el-input v-model="item.key" placeholder="key"/><span>=</span><el-input v-model="item.value" placeholder="value"/><el-button link type="danger" :icon="Delete" @click="labels.splice(index,1)"/></div><el-empty v-if="!labels.length" description="暂无标签" :image-size="42"/></section>
          <section class="pair-section"><div><strong>注解</strong><el-button link :icon="Plus" @click="addPair(annotations)">添加</el-button></div><div v-for="(item,index) in annotations" :key="index" class="pair-row"><el-input v-model="item.key" placeholder="key"/><span>=</span><el-input v-model="item.value" placeholder="value"/><el-button link type="danger" :icon="Delete" @click="annotations.splice(index,1)"/></div><el-empty v-if="!annotations.length" description="暂无注解" :image-size="42"/></section>
        </div>
      </aside>
      <div class="panel-splitter" :class="{ dragging: !sidebarCollapsed }" @pointerdown="startSidebarResize"><button type="button" :title="sidebarCollapsed ? '展开基础信息' : '收起基础信息'" @pointerdown.stop @click="toggleSidebar">{{ sidebarCollapsed ? '›' : '‹' }}</button></div>
      <main>
        <el-tabs v-model="activeTab" class="config-tabs">
          <el-tab-pane label="容器配置" name="containers"><div class="tab-content"><section class="config-section"><header><strong>标准容器 (Containers)</strong><el-button link :icon="Plus" @click="addContainer(containers)">添加容器</el-button></header><el-collapse accordion><el-collapse-item v-for="(container,index) in containers" :key="index" :name="index"><template #title><div class="container-title"><span>▣</span><strong>{{container.name||'未命名容器'}}</strong><el-tag type="success">{{container.image||'无镜像'}}</el-tag><el-button link type="danger" :icon="Delete" @click.stop="containers.splice(index,1)">删除</el-button></div></template><KubernetesContainerConfig :container="container" :volumes="volumes"/></el-collapse-item></el-collapse><el-empty v-if="!containers.length" description="暂无标准容器"/></section><section class="config-section"><header><strong>初始化容器 (Init Containers)</strong><el-button link :icon="Plus" @click="addContainer(initContainers)">添加初始化容器</el-button></header><el-collapse accordion><el-collapse-item v-for="(container,index) in initContainers" :key="index" :name="index"><template #title><div class="container-title"><span>▣</span><strong>{{container.name||'未命名容器'}}</strong><el-tag type="warning">{{container.image||'无镜像'}}</el-tag><el-button link type="danger" :icon="Delete" @click.stop="initContainers.splice(index,1)">删除</el-button></div></template><KubernetesContainerConfig :container="container" :volumes="volumes"/></el-collapse-item></el-collapse><el-empty v-if="!initContainers.length" description="暂无初始化容器"/></section></div></el-tab-pane>
          <el-tab-pane label="存储" name="storage"><div class="tab-content"><section class="config-section storage-section"><header><strong>数据卷 (Volumes)</strong><el-button link :icon="Plus" @click="addVolume">添加数据卷</el-button></header><el-collapse accordion><el-collapse-item v-for="(volume,index) in volumes" :key="index" :name="index"><template #title><div class="volume-title"><strong>{{volume.name||'未命名数据卷'}}（{{volumeTypeLabel(volume)}}）</strong><el-button link type="danger" :icon="Delete" @click.stop="volumes.splice(index,1)">删除</el-button></div></template><div class="volume-form"><label>名称<el-input v-model="volume.name" placeholder="数据卷名称"/></label><label>类型<el-select :model-value="volumeType(volume)" @change="setVolumeType(volume,String($event))"><el-option label="PVC" value="persistentVolumeClaim"/><el-option label="ConfigMap" value="configMap"/><el-option label="Secret" value="secret"/><el-option label="EmptyDir" value="emptyDir"/><el-option label="HostPath" value="hostPath"/></el-select></label><template v-if="volume.persistentVolumeClaim"><label>声明名称<el-input v-model="volume.persistentVolumeClaim.claimName" placeholder="PersistentVolumeClaim 名称"/></label><label class="inline-field">只读<el-switch v-model="volume.persistentVolumeClaim.readOnly"/></label></template><template v-else-if="volume.configMap"><label>ConfigMap 名称<el-input v-model="volume.configMap.name" placeholder="ConfigMap 名称"/></label><label class="inline-field">可选引用<el-switch v-model="volume.configMap.optional"/></label></template><template v-else-if="volume.secret"><label>Secret 名称<el-input v-model="volume.secret.secretName" placeholder="Secret 名称"/></label><label class="inline-field">可选引用<el-switch v-model="volume.secret.optional"/></label></template><template v-else-if="volume.emptyDir"><label>存储介质<el-select v-model="volume.emptyDir.medium" clearable placeholder="节点磁盘"><el-option label="节点磁盘" value=""/><el-option label="内存 tmpfs" value="Memory"/></el-select></label><label>容量限制<el-input v-model="volume.emptyDir.sizeLimit" clearable placeholder="例如 1Gi"/></label></template><template v-else-if="volume.hostPath"><label>主机路径<el-input v-model="volume.hostPath.path" placeholder="例如 /data"/></label><label>路径类型<el-select v-model="volume.hostPath.type"><el-option label="DirectoryOrCreate" value="DirectoryOrCreate"/><el-option label="Directory" value="Directory"/><el-option label="FileOrCreate" value="FileOrCreate"/><el-option label="File" value="File"/></el-select></label></template></div></el-collapse-item></el-collapse><el-empty v-if="!volumes.length" description="暂无数据卷"/></section></div></el-tab-pane>
          <el-tab-pane label="扩容配置" name="scaling"><div class="tab-content"><section class="settings-card"><header><strong>扩容与更新配置</strong></header><div class="settings-grid"><label v-if="supportsReplicas">当前副本数<el-input-number v-model="workloadSpec.replicas" :min="0" :max="1000"/></label><label v-if="kind==='Deployment'">更新策略<el-select v-model="workloadSpec.strategy.type"><el-option label="RollingUpdate（滚动更新）" value="RollingUpdate"/><el-option label="Recreate（重建）" value="Recreate"/></el-select></label><label v-if="kind==='StatefulSet'||kind==='DaemonSet'">更新策略<el-select v-model="workloadSpec.updateStrategy.type"><el-option label="RollingUpdate" value="RollingUpdate"/><el-option label="OnDelete" value="OnDelete"/></el-select></label><label v-if="kind==='Deployment'">最小就绪时间（秒）<el-input-number v-model="workloadSpec.minReadySeconds" :min="0"/></label><label v-if="kind==='Deployment'">进度截止时间（秒）<el-input-number v-model="workloadSpec.progressDeadlineSeconds" :min="1"/></label><label v-if="['Deployment','StatefulSet','DaemonSet'].includes(kind)">版本历史限制<el-input-number v-model="workloadSpec.revisionHistoryLimit" :min="0"/></label><label v-if="kind==='StatefulSet'">Pod 管理策略<el-select v-model="workloadSpec.podManagementPolicy"><el-option label="OrderedReady" value="OrderedReady"/><el-option label="Parallel" value="Parallel"/></el-select></label><label v-if="kind==='Job'">并行数<el-input-number v-model="workloadSpec.parallelism" :min="1"/></label><label v-if="kind==='Job'">完成数<el-input-number v-model="workloadSpec.completions" :min="1"/></label><label v-if="kind==='Job'">失败重试次数<el-input-number v-model="workloadSpec.backoffLimit" :min="0"/></label><label v-if="kind==='CronJob'">Cron 表达式<el-input v-model="workloadSpec.schedule"/></label><label v-if="kind==='CronJob'">并发策略<el-select v-model="workloadSpec.concurrencyPolicy"><el-option label="Allow" value="Allow"/><el-option label="Forbid" value="Forbid"/><el-option label="Replace" value="Replace"/></el-select></label><label v-if="kind==='CronJob'" class="inline-field">暂停调度<el-switch v-model="workloadSpec.suspend"/></label></div><el-empty v-if="kind==='Pod'" description="Pod 不支持扩容与滚动更新配置"/></section></div></el-tab-pane>
          <el-tab-pane label="节点调度" name="scheduling"><div class="tab-content"><section class="settings-card"><header><strong>调度类型</strong></header><el-radio-group :model-value="schedulingMode" @change="changeSchedulingMode($event as any)"><el-radio-button value="any">任意可用节点</el-radio-button><el-radio-button value="node">指定节点</el-radio-button><el-radio-button value="selector">调度规则匹配</el-radio-button></el-radio-group><div v-if="schedulingMode==='node'" class="single-field"><label>节点名称<el-input v-model="podSpec.nodeName" placeholder="请输入节点名称"/></label></div><div v-if="schedulingMode==='selector'" class="rule-list"><div class="list-head"><strong>节点标签匹配规则</strong><el-button link :icon="Plus" @click="addPair(nodeSelectors)">添加规则</el-button></div><div v-for="(item,index) in nodeSelectors" :key="index" class="pair-row wide"><el-input v-model="item.key" placeholder="标签键"/><span>=</span><el-input v-model="item.value" placeholder="标签值"/><el-button link type="danger" :icon="Delete" @click="nodeSelectors.splice(index,1)"/></div><el-empty v-if="!nodeSelectors.length" description="暂无匹配规则" :image-size="38"/></div></section><section class="settings-card"><header><strong>容忍度配置</strong><el-button link :icon="Plus" @click="podList('tolerations').push({key:'',operator:'Equal',value:'',effect:'NoSchedule'})">添加容忍</el-button></header><div class="toleration-head"><span>键</span><span>运算符</span><span>值</span><span>影响</span><span>容忍时间</span><span>操作</span></div><div v-for="(item,index) in podList('tolerations')" :key="index" class="toleration-row"><el-input v-model="item.key"/><el-select v-model="item.operator"><el-option label="Equal" value="Equal"/><el-option label="Exists" value="Exists"/></el-select><el-input v-model="item.value" :disabled="item.operator==='Exists'"/><el-select v-model="item.effect" clearable><el-option label="NoSchedule" value="NoSchedule"/><el-option label="PreferNoSchedule" value="PreferNoSchedule"/><el-option label="NoExecute" value="NoExecute"/></el-select><el-input-number v-model="item.tolerationSeconds" :min="0"/><el-button link type="danger" :icon="Delete" @click="podList('tolerations').splice(index,1)"/></div><el-empty v-if="!podList('tolerations').length" description="暂无容忍度" :image-size="38"/></section></div></el-tab-pane>
          <el-tab-pane label="网络" name="network"><div class="tab-content"><section class="settings-card"><header><strong>主机网络</strong></header><div class="switch-setting"><span>使用主机网络</span><el-switch v-model="podSpec.hostNetwork"/><small>Pod 将使用主机的网络命名空间</small></div></section><section class="settings-card"><header><strong>DNS 策略</strong></header><div class="single-field"><label>DNS 策略<el-select v-model="podSpec.dnsPolicy"><el-option label="ClusterFirst" value="ClusterFirst"/><el-option label="ClusterFirstWithHostNet" value="ClusterFirstWithHostNet"/><el-option label="Default" value="Default"/><el-option label="None" value="None"/></el-select></label></div></section><section class="settings-card"><header><strong>主机名设置</strong></header><div class="settings-grid"><label>主机名<el-input v-model="podSpec.hostname" clearable placeholder="可选，指定 Pod 的主机名"/></label><label>子域名<el-input v-model="podSpec.subdomain" clearable placeholder="可选，指定 Pod 的子域名"/></label></div></section><section class="settings-card"><header><strong>DNS 配置</strong></header><div class="list-head"><span>服务器地址</span><el-button link :icon="Plus" @click="dnsList('nameservers').push('')">添加服务器</el-button></div><div v-for="(_,index) in dnsList('nameservers')" :key="`n${index}`" class="simple-list-row"><el-input v-model="podSpec.dnsConfig.nameservers[index]" placeholder="例如 8.8.8.8"/><el-button link type="danger" :icon="Delete" @click="dnsList('nameservers').splice(index,1)"/></div><div class="list-head"><span>搜索域</span><el-button link :icon="Plus" @click="dnsList('searches').push('')">添加搜索域</el-button></div><div v-for="(_,index) in dnsList('searches')" :key="`s${index}`" class="simple-list-row"><el-input v-model="podSpec.dnsConfig.searches[index]"/><el-button link type="danger" :icon="Delete" @click="dnsList('searches').splice(index,1)"/></div></section></div></el-tab-pane>
          <el-tab-pane label="其他" name="other"><div class="tab-content"><section class="settings-card"><header><strong>优雅终止</strong></header><div class="settings-grid"><label>优雅终止期限（秒）<el-input-number v-model="podSpec.terminationGracePeriodSeconds" :min="0"/></label><label>活动期限（秒）<el-input-number v-model="podSpec.activeDeadlineSeconds" :min="1" placeholder="可选"/></label></div></section><section class="settings-card"><header><strong>服务账户</strong></header><div class="settings-grid"><label>服务账户名<el-input v-model="podSpec.serviceAccountName" clearable placeholder="default"/></label><label class="inline-field">自动挂载令牌<el-switch v-model="podSpec.automountServiceAccountToken"/></label></div></section><section class="settings-card"><header><strong>优先级</strong></header><div class="single-field"><label>优先级类名称<el-input v-model="podSpec.priorityClassName" clearable placeholder="可选，如 high-priority"/></label></div></section><section class="settings-card"><header><strong>重启策略</strong></header><div class="single-field"><label>重启策略<el-select v-model="podSpec.restartPolicy" :disabled="!['Pod','Job','CronJob'].includes(kind)"><el-option label="Always" value="Always"/><el-option label="OnFailure" value="OnFailure"/><el-option label="Never" value="Never"/></el-select></label></div></section></div></el-tab-pane>
        </el-tabs>
      </main>
    </div>
    <template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
  </el-dialog>
</template>

<style scoped>
.workload-editor{display:grid;grid-template-columns:250px minmax(0,1fr);width:100%;height:clamp(380px,58vh,530px);min-width:0;min-height:0;border:1px solid var(--line);background:#fff;font-size:11px}.basic-panel{min-width:0;min-height:0;overflow:auto;border-right:4px solid #dfb62e}.basic-panel>header{position:sticky;top:0;z-index:2;display:flex;align-items:center;gap:10px;padding:12px 14px;border-bottom:2px solid #dfb62e;background:#fff}.basic-panel>header span{display:grid;width:30px;height:30px;place-items:center;border-radius:8px;background:#dfb62e;color:#fff;font-size:14px}.basic-panel>header strong{font-size:13px}.basic-content{display:flex;flex-direction:column;gap:12px;padding:14px}.basic-content>label,.form-grid>label,.container-form>label{display:flex;flex-direction:column;gap:6px;font-size:11px;font-weight:600}.pair-section{padding:10px;border:1px solid var(--line);border-radius:9px;background:#fafafa}.pair-section>div:first-child,.config-section>header{display:flex;align-items:center;justify-content:space-between}.pair-section strong,.config-section>header strong{font-size:12px}.pair-row{display:grid;grid-template-columns:minmax(0,1fr) 12px minmax(0,1fr) 26px;align-items:center;gap:5px;margin-top:8px}.pair-row.wide{max-width:680px}main{min-width:0;min-height:0;overflow:hidden}.config-tabs{height:100%;font-size:11px}.config-tabs :deep(.el-tabs__header){margin:0}.config-tabs :deep(.el-tabs__nav){width:100%}.config-tabs :deep(.el-tabs__item){flex:1;height:42px;font-size:12px;font-weight:600}.config-tabs :deep(.el-tabs__content){height:calc(100% - 42px);overflow:auto}.workload-editor :deep(.el-input__inner),.workload-editor :deep(.el-select__placeholder),.workload-editor :deep(.el-button),.workload-editor :deep(.el-tag),.workload-editor :deep(.el-empty__description),.workload-editor :deep(.el-input-number){font-size:11px}.tab-content{display:flex;flex-direction:column;gap:18px;padding:12px}.config-section{border:1px solid var(--line);border-radius:8px;overflow:hidden}.config-section>header{padding:10px 12px;background:#dfb62e;color:#fff}.container-title{display:flex;width:100%;min-width:0;align-items:center;gap:10px;padding-right:10px;font-size:11px}.container-title .el-tag{max-width:55%;overflow:hidden;text-overflow:ellipsis}.container-title .el-button{margin-left:auto}.container-form{display:grid;grid-template-columns:1fr 2fr;gap:10px;padding:12px}.container-form label:nth-child(n+4){grid-column:1/-1}.volume-row{display:grid;grid-template-columns:minmax(0,1fr) 120px 60px;align-items:center;gap:10px;padding:10px 14px;border-top:1px solid var(--line)}.form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));align-content:start;gap:16px}.form-grid label>.el-select,.form-grid label>.el-input{width:100%}:global(.el-overlay-dialog:has(.workload-editor-dialog)){display:flex;align-items:center;justify-content:center;padding:12px;overflow:hidden}:global(.workload-editor-dialog){max-height:calc(100vh - 24px);margin:0!important;transform:none!important;font-size:11px}:global(.workload-editor-dialog .el-dialog__title){font-size:16px;font-weight:650}:global(.workload-editor-dialog .el-dialog__body){padding:10px 14px}:global(.workload-editor-dialog .el-dialog__header){padding:14px 16px 8px}:global(.workload-editor-dialog .el-dialog__footer){padding:8px 16px 14px}:global(.workload-editor-dialog .el-dialog__footer .el-button){font-size:11px}@media(max-width:820px){.workload-editor{grid-template-columns:1fr;height:min(65vh,530px);overflow:auto}.basic-panel{overflow:visible;border-right:0;border-bottom:4px solid #dfb62e}main{overflow:visible}.config-tabs :deep(.el-tabs__content){height:auto}.form-grid,.container-form{grid-template-columns:1fr}}@media(max-height:620px){.workload-editor{height:calc(100vh - 140px);min-height:300px}}
.basic-panel{border-right:0;transition:opacity .16s ease}
.sidebar-collapsed .basic-panel{opacity:0;pointer-events:none}
.panel-splitter{position:relative;z-index:3;display:flex;height:100%;align-items:center;justify-content:center;background:var(--line);cursor:col-resize;touch-action:none}
.panel-splitter:hover,.panel-splitter:active{background:var(--primary-soft);border-color:color-mix(in srgb,var(--primary) 35%,var(--line))}
.panel-splitter button{position:absolute;z-index:2;display:grid;width:18px;height:32px;place-items:center;padding:0;border:1px solid var(--line);border-radius:5px;background:#fff;color:var(--primary);font-size:17px;line-height:1;cursor:pointer;box-shadow:0 2px 6px rgba(31,35,48,.08)}
.panel-splitter button:hover{border-color:var(--primary);background:var(--primary-soft)}
.basic-panel>header{border-bottom:1px solid var(--line);background:#fafafa}
.basic-panel>header span{background:var(--primary);color:#fff;box-shadow:0 2px 6px color-mix(in srgb,var(--primary) 22%,transparent)}
.basic-panel>header strong{color:var(--text)}
.pair-section{background:#fff}
.config-section>header{border-bottom:1px solid color-mix(in srgb,var(--primary) 20%,var(--line));background:var(--primary-soft);color:var(--primary)}
.config-section>header :deep(.el-button){color:var(--primary)}
.config-tabs :deep(.el-tabs__item.is-active){color:var(--primary)}
.config-tabs :deep(.el-tabs__active-bar){background-color:var(--primary)}
.basic-panel,.config-tabs :deep(.el-tabs__content){scrollbar-width:thin;scrollbar-color:#c7ced9 transparent}
.basic-panel::-webkit-scrollbar,.config-tabs :deep(.el-tabs__content::-webkit-scrollbar){width:1px;height:1px}
.basic-panel::-webkit-scrollbar-track,.config-tabs :deep(.el-tabs__content::-webkit-scrollbar-track){background:transparent}
.basic-panel::-webkit-scrollbar-thumb,.config-tabs :deep(.el-tabs__content::-webkit-scrollbar-thumb){border-radius:1px;background:#c7ced9}
.basic-panel::-webkit-scrollbar-thumb:hover,.config-tabs :deep(.el-tabs__content::-webkit-scrollbar-thumb:hover){background:#aeb7c5}
@media(max-width:820px){.workload-editor{grid-template-columns:1fr!important}.basic-panel{border-right:0;border-bottom:1px solid var(--line)}.panel-splitter{display:none}.sidebar-collapsed .basic-panel{display:none}.sidebar-collapsed main{display:block}}
.workload-editor{font-size:10px}
.basic-panel>header span,.basic-panel>header strong,.pair-section strong,.config-section>header strong{font-size:11px}
.basic-content>label,.form-grid>label,.container-form>label,.container-title,.config-tabs{font-size:10px}
.config-tabs :deep(.el-tabs__item){font-size:11px}
.workload-editor :deep(.el-input__inner),.workload-editor :deep(.el-select__placeholder),.workload-editor :deep(.el-button),.workload-editor :deep(.el-tag),.workload-editor :deep(.el-input-number){font-size:10px}
.workload-editor :deep(.el-empty__description){font-size:9px}
:global(.workload-editor-dialog){font-size:10px}
:global(.workload-editor-dialog .el-dialog__title){font-size:11px}
:global(.workload-editor-dialog .el-dialog__footer .el-button){font-size:10px}
.basic-panel>header{padding:9px 10px}
.basic-content{gap:8px;padding:10px}
.pair-section{padding:7px;border-radius:6px}
.pair-row{gap:3px;margin-top:5px}
.tab-content{gap:12px;padding:8px}
.config-section{border-width:1px;border-radius:6px}
.config-section>header{padding:7px 9px}
.container-form{gap:7px;padding:8px}
.container-title{gap:7px;padding-right:6px}
.volume-row{gap:7px;padding:7px 9px}
.volume-title{display:flex;width:100%;align-items:center;justify-content:space-between;padding-right:8px}.volume-form{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px;padding:8px 10px}.volume-form label{display:flex;min-width:0;flex-direction:column;gap:4px;font-size:10px;font-weight:600}.volume-form .inline-field{flex-direction:row;align-items:center;justify-content:space-between;padding:0 4px}.storage-section :deep(.el-collapse-item__header){padding:0 10px}.storage-section :deep(.el-collapse-item__content){padding-bottom:4px}
.form-grid{gap:10px}
:global(.workload-editor-dialog .el-dialog__body){padding:6px 8px}
:global(.workload-editor-dialog .el-dialog__header){padding:10px 12px 6px}
:global(.workload-editor-dialog .el-dialog__footer){padding:6px 12px 10px}
@media(max-width:720px){.volume-form{grid-template-columns:1fr}}
.settings-card{padding:9px;border:1px solid var(--line);border-radius:6px;background:#fff}.settings-card>header{display:flex;align-items:center;justify-content:space-between;margin-bottom:9px;padding-bottom:7px;border-bottom:1px solid var(--line)}.settings-card>header strong{padding-left:7px;border-left:2px solid var(--primary);font-size:11px}.settings-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:9px 14px}.settings-grid label,.single-field label{display:flex;min-width:0;flex-direction:column;gap:4px;font-size:10px;font-weight:600}.settings-grid .inline-field{flex-direction:row;align-items:center;justify-content:space-between}.single-field{margin-top:9px}.single-field label{max-width:520px}.rule-list{margin-top:10px}.switch-setting{display:flex;align-items:center;gap:9px}.switch-setting small{color:var(--text-soft);font-size:9px}.toleration-head,.toleration-row{display:grid;grid-template-columns:1fr 100px 1fr 130px 110px 28px;align-items:center;gap:5px}.toleration-head{padding:5px 0;border-bottom:1px solid var(--line);color:var(--text-soft);font-weight:600}.toleration-row{margin-top:5px}.simple-list-row{display:grid;grid-template-columns:minmax(0,1fr) 26px;gap:5px;margin-bottom:5px}.settings-card :deep(.el-radio-button__inner){font-size:10px}.settings-card :deep(.el-input__inner),.settings-card :deep(.el-select__placeholder),.settings-card :deep(.el-button){font-size:10px}@media(max-width:760px){.settings-grid{grid-template-columns:1fr}.toleration-head{display:none}.toleration-row{grid-template-columns:1fr 1fr}.toleration-row .el-button{justify-self:end}}
</style>
