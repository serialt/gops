<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, CopyDocument, Delete, Document, Download, EditPen, FullScreen, Refresh, RefreshRight, Setting, SetUp } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import KubernetesYamlEditor from '@/components/KubernetesYamlEditor.vue'
import { applyKubernetesObject, deleteKubernetesObject, execKubernetesPodCommand, getKubernetesCronJobSupport, getKubernetesDaemonSetSupport, getKubernetesJobSupport, getKubernetesObject, getKubernetesPodLogs, getKubernetesStatefulSetSupport, listKubernetesObjects, type KubernetesObject, type KubernetesStatefulSetSupport } from '@/api/kubernetes'

const route = useRoute(), router = useRouter()
const clusterID = computed(() => Number(route.query.cluster_id || 0))
const namespace = computed(() => String(route.query.namespace || 'default'))
const name = computed(() => String(route.query.name || ''))
const isDaemonSet = computed(() => route.name === 'kubernetes-daemonset-detail')
const isJob = computed(() => route.name === 'kubernetes-job-detail')
const isCronJob = computed(() => route.name === 'kubernetes-cronjob-detail')
const isStatefulSet = computed(() => !isDaemonSet.value && !isJob.value && !isCronJob.value)
const resource = computed(() => isDaemonSet.value ? 'daemonsets' : isJob.value ? 'jobs' : isCronJob.value ? 'cronjobs' : 'statefulsets')
const kindLabel = computed(() => isDaemonSet.value ? 'DaemonSet' : isJob.value ? 'Job' : isCronJob.value ? 'CronJob' : 'StatefulSet')
const loading = ref(false), activeTab = ref('overview'), yaml = ref(''), manifest = ref<Record<string, any>>({})
const row = ref<KubernetesObject | null>(null), statefulSetSupport = ref<KubernetesStatefulSetSupport>({ status: 'Pending', desired_replicas: 0, current_replicas: 0, ready_replicas: 0, updated_replicas: 0, pods: [], claims: [], events: [], description: '' })
const dialogVisible = ref(false), dialogMode = ref<'describe' | 'yaml' | 'scale' | 'delete'>('describe')
const editorYaml = ref(''), yamlMode = ref<'edit' | 'clone'>('edit'), saving = ref(false), scaleReplicas = ref(1), deleting = ref(false), deleteConfirm = ref('')
const logPod = ref('__all__'), logContainer = ref('__all__'), logs = ref(''), logsLoading = ref(false), logFilter = ref(''), logTailLines = ref(100), logTimestamps = ref(false), logFontSize = ref(12), logWrap = ref(false), logPanel = ref<HTMLElement>()
const terminalPod = ref(''), terminalContainer = ref(''), terminalFontSize = ref(12), terminalUp = ref(0), terminalDown = ref(0), commandRunning = ref(false), terminalHost = ref<HTMLElement>(), terminalPanel = ref<HTMLElement>()
let xterm: Terminal | undefined, xtermFit: FitAddon | undefined, terminalBuffer = '', terminalSessionKey = '', terminalResizeObserver: ResizeObserver | undefined

const metadata = computed(() => manifest.value.metadata || {})
const spec = computed(() => manifest.value.spec || {})
const templateSpec = computed(() => spec.value.template?.spec || {})
const status = computed(() => manifest.value.status || {})
const selector = computed<Record<string, string>>(() => spec.value.selector?.matchLabels || {})
const pods = computed(() => statefulSetSupport.value.pods)
const claims = computed(() => statefulSetSupport.value.claims)
const events = computed(() => statefulSetSupport.value.events)
const containers = computed<any[]>(() => templateSpec.value.containers || [])
const initContainers = computed<any[]>(() => templateSpec.value.initContainers || [])
const allContainers = computed<any[]>(() => [...containers.value, ...initContainers.value])
const volumes = computed<any[]>(() => templateSpec.value.volumes || [])
const claimTemplates = computed<any[]>(() => spec.value.volumeClaimTemplates || [])
const labels = computed<Record<string, string>>(() => metadata.value.labels || {})
const annotations = computed<Record<string, string>>(() => metadata.value.annotations || {})
const desired = computed(() => statefulSetSupport.value.desired_replicas)
const ready = computed(() => statefulSetSupport.value.ready_replicas)
const current = computed(() => statefulSetSupport.value.current_replicas)
const updated = computed(() => statefulSetSupport.value.updated_replicas)
const statefulSetStatus = computed(() => statusText(statefulSetSupport.value.status))
const statefulSetHealthy = computed(() => ['Running', 'Completed'].includes(statefulSetSupport.value.status))
const volumeCount = computed(() => volumes.value.length + (isStatefulSet.value ? claims.value.length + claimTemplates.value.length : 0))
const statefulSetEvents = computed(() => events.value)
const description = computed(() => statefulSetSupport.value.description || '暂无描述信息')
const filteredLogs = computed(() => logs.value.split('\n').filter(line => !logFilter.value || line.toLowerCase().includes(logFilter.value.toLowerCase())).join('\n'))
const logLines = computed(() => filteredLogs.value ? filteredLogs.value.split('\n').filter(Boolean).length : 0)
const volumeMounts = computed(() => allContainers.value.flatMap(container => (container.volumeMounts || []).map((mount: any) => ({ container: container.name, ...mount }))))

async function load() {
  if (!clusterID.value || !name.value) return
  loading.value = true
  try {
    const supportRequest = isDaemonSet.value
      ? getKubernetesDaemonSetSupport(clusterID.value, namespace.value, name.value)
      : isJob.value
        ? getKubernetesJobSupport(clusterID.value, namespace.value, name.value)
        : isCronJob.value
          ? getKubernetesCronJobSupport(clusterID.value, namespace.value, name.value)
          : getKubernetesStatefulSetSupport(clusterID.value, namespace.value, name.value)
    const [detail, objects, support] = await Promise.all([
      getKubernetesObject({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, name: name.value }),
      listKubernetesObjects({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, keyword: name.value, page: 1, limit: 100 }),
      supportRequest,
    ])
    yaml.value = detail.data.data.yaml
    manifest.value = parse(yaml.value) || {}
    row.value = objects.data.data.list.find(item => item.name === name.value) || null
    statefulSetSupport.value = support.data.data
    if (!pods.value.some(pod => pod.name === terminalPod.value)) terminalPod.value = pods.value[0]?.name || ''
    if (!containers.value.some(container => container.name === terminalContainer.value)) terminalContainer.value = containers.value[0]?.name || ''
  } finally { loading.value = false }
}

async function loadLogs() {
  if (!pods.value.length || !allContainers.value.length) { logs.value = ''; return }
  logsLoading.value = true
  try {
    const podTargets = logPod.value === '__all__' ? pods.value.map(pod => pod.name) : [logPod.value]
    const containerTargets = logContainer.value === '__all__' ? allContainers.value.map(container => container.name) : [logContainer.value]
    const targets = podTargets.flatMap(pod => containerTargets.map(container => ({ pod, container })))
    const results = await Promise.all(targets.map(async target => {
      try {
        const result = await getKubernetesPodLogs({ cluster_id: clusterID.value, namespace: namespace.value, pod: target.pod, container: target.container, tail_lines: logTailLines.value, timestamps: logTimestamps.value })
        return { ...target, logs: result.data.data.logs || '' }
      } catch (error: any) {
        return { ...target, logs: `[日志读取失败: ${error?.response?.data?.message || error?.message || '未知错误'}]` }
      }
    }))
    logs.value = results.map(result => targets.length > 1 ? `[${result.pod}/${result.container}]\n${result.logs}` : result.logs).join('\n')
  } finally { logsLoading.value = false }
}

async function runTerminalCommand(command: string) {
  if (!command.trim() || !terminalPod.value || !terminalContainer.value || !xterm) { xterm?.write('\r\n$ '); return }
  commandRunning.value = true
  terminalUp.value += new TextEncoder().encode(command).length
  try {
    const { data } = await execKubernetesPodCommand({ cluster_id: clusterID.value, namespace: namespace.value, pod: terminalPod.value, container: terminalContainer.value, command })
    const output = `${data.data.output || ''}${data.data.error || ''}`
    terminalDown.value += new TextEncoder().encode(output).length
    xterm.write(`\r\n${output.replace(/\n/g, '\r\n')}${output && !output.endsWith('\n') ? '\r\n' : ''}$ `)
  } catch (error: any) {
    xterm.write(`\r\n${error?.response?.data?.message || error?.message || '命令执行失败'}\r\n$ `)
  } finally { commandRunning.value = false }
}

async function initializeTerminal() {
  await nextTick()
  if (!terminalHost.value || !terminalPod.value || !terminalContainer.value) return
  terminalResizeObserver?.disconnect()
  xterm?.dispose()
  terminalHost.value.replaceChildren()
  terminalBuffer = ''
  terminalSessionKey = `${terminalPod.value}/${terminalContainer.value}`
  xterm = new Terminal({ cursorBlink: true, fontSize: terminalFontSize.value, lineHeight: 1.25, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace', theme: { background: '#050505', foreground: '#e7eaf0', cursor: '#f5f5f5', green: '#62e785' }, scrollback: 5000 })
  xtermFit = new FitAddon()
  xterm.loadAddon(xtermFit)
  xterm.open(terminalHost.value)
  xtermFit.fit()
  xterm.write('\x1b[32m$ \x1b[0m')
  xterm.onData(data => {
    if (commandRunning.value) return
    if (data === '\r') { const command = terminalBuffer; terminalBuffer = ''; void runTerminalCommand(command); return }
    if (data === '\u007f') { if (terminalBuffer) { terminalBuffer = terminalBuffer.slice(0, -1); xterm?.write('\b \b') }; return }
    if (data >= ' ') { terminalBuffer += data; xterm?.write(data) }
  })
  terminalResizeObserver = new ResizeObserver(() => xtermFit?.fit())
  terminalResizeObserver.observe(terminalHost.value)
  xterm.focus()
}

async function activateTerminal() {
  await nextTick()
  if (!xterm) { await initializeTerminal(); return }
  xtermFit?.fit()
  xterm.focus()
}

function clearConsole(target: 'logs' | 'terminal') {
  if (target === 'logs') { logs.value = ''; return }
  xterm?.clear()
  xterm?.write('\x1b[32m$ \x1b[0m')
}
function formatTraffic(value: number) { return value < 1024 ? `${value} B/s` : `${(value / 1024).toFixed(1)} KB/s` }
function downloadLogs() { const link = document.createElement('a'); link.href = URL.createObjectURL(new Blob([filteredLogs.value], { type: 'text/plain;charset=utf-8' })); link.download = `${name.value}-${logPod.value === '__all__' ? 'all-pods' : logPod.value}.log`; link.click(); URL.revokeObjectURL(link.href) }
async function fullscreen(target: HTMLElement | undefined) { if (!target) return; if (document.fullscreenElement) await document.exitFullscreen(); else await target.requestFullscreen(); await nextTick(); xtermFit?.fit() }

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
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: resource.value, namespace: object?.metadata?.namespace || namespace.value, yaml: editorYaml.value })
    dialogVisible.value = false
    ElMessage.success(yamlMode.value === 'clone' ? `${kindLabel.value} 已克隆` : `${kindLabel.value} 已保存`)
    if (yamlMode.value === 'clone') await router.replace({ name: `${isJob.value ? 'kubernetes-job' : isCronJob.value ? 'kubernetes-cronjob' : isDaemonSet.value ? 'kubernetes-daemonset' : 'kubernetes-statefulset'}-detail`, query: { cluster_id: clusterID.value, namespace: object?.metadata?.namespace || namespace.value, name: object?.metadata?.name } })
    await load()
  } finally { saving.value = false }
}
function scale() { scaleReplicas.value = desired.value; openDialog('scale') }
async function submitScale() {
  saving.value = true
  try {
    const object = parse(yaml.value)
    object.spec = object.spec || {}
    object.spec.replicas = scaleReplicas.value
    await applyKubernetesObject({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, yaml: stringify(object) })
    dialogVisible.value = false
    ElMessage.success('扩缩容成功')
    await load()
  } finally { saving.value = false }
}
async function restart() {
  await ElMessageBox.confirm(`确认重启 ${name.value}？`, `重启 ${kindLabel.value}`)
  const object = parse(yaml.value)
  object.spec.template.metadata = object.spec.template.metadata || {}
  object.spec.template.metadata.annotations = object.spec.template.metadata.annotations || {}
  object.spec.template.metadata.annotations['kubectl.kubernetes.io/restartedAt'] = new Date().toISOString()
  await applyKubernetesObject({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, yaml: stringify(object) })
  ElMessage.success('已触发滚动重启')
  await load()
}
async function remove() {
  deleting.value = true
  try {
    await deleteKubernetesObject({ cluster_id: clusterID.value, resource: resource.value, namespace: namespace.value, name: name.value, wait: true })
    ElMessage.success('删除成功')
    await router.push('/kubernetes/workloads')
  } finally { deleting.value = false }
}
function openPod(pod: KubernetesObject) { void router.push({ name: 'kubernetes-pod-detail', query: { cluster_id: clusterID.value, namespace: namespace.value, name: pod.name } }) }
function statusText(value?: string) { return ({ Running: '运行中', Progressing: '更新中', Pending: '未就绪', Degraded: '部分就绪', Stopped: '已停止', Completed: '已完成', Failed: '失败', Scheduled: '已调度', Suspended: '已暂停', Available: '可用', Ready: '可用' } as Record<string, string>)[value || ''] || value || '未知' }
function volumeSource(volume: Record<string, any>) {
  const type = Object.keys(volume).find(key => key !== 'name') || ''
  const source = volume[type]
  return source?.claimName || source?.name || source?.secretName || source?.path || (type ? JSON.stringify(source) : '-')
}
watch([clusterID, namespace, name, resource], () => { activeTab.value = 'overview'; void load() }, { immediate: true })
watch(activeTab, value => { if (value === 'logs' && !logs.value) void loadLogs(); if (value === 'terminal') void activateTerminal() })
watch([logPod, logContainer], () => { if (activeTab.value === 'logs') void loadLogs() })
watch([logTailLines, logTimestamps], () => { if (activeTab.value === 'logs') void loadLogs() })
watch([terminalPod, terminalContainer], ([pod, container]) => {
  const sessionKey = `${pod}/${container}`
  if (!xterm || !pod || !container || terminalSessionKey === sessionKey) return
  terminalBuffer = ''
  terminalSessionKey = sessionKey
  xterm.write(`\r\n\x1b[90m已切换至 ${sessionKey}\x1b[0m\r\n\x1b[32m$ \x1b[0m`)
  if (activeTab.value === 'terminal') void nextTick(() => { xtermFit?.fit(); xterm?.focus() })
})
watch(terminalFontSize, value => { if (xterm) xterm.options.fontSize = value; xtermFit?.fit() })
onBeforeUnmount(() => { terminalResizeObserver?.disconnect(); xterm?.dispose() })
</script>

<template>
  <section v-loading="loading" class="statefulset-detail">
    <header class="detail-header"><div class="identity"><el-button text :icon="ArrowLeft" @click="router.back()"/><div><h1>{{ name }}</h1><p>{{ kindLabel }} · Namespace: <b>{{ namespace }}</b></p></div></div><div class="actions"><el-button :icon="Refresh" @click="load">刷新</el-button><el-button :icon="Document" @click="openDialog('describe')">描述</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="CopyDocument" @click="clone">克隆</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="EditPen" @click="editYaml">编辑 YAML</el-button><el-button v-if="isStatefulSet" v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="SetUp" @click="scale">扩缩容</el-button><el-button v-if="isStatefulSet || isDaemonSet" v-permission="'POST /api/v1/kubernetes/resources/apply'" :icon="RefreshRight" @click="restart">重启</el-button><el-button v-permission="'POST /api/v1/kubernetes/resources/delete'" type="danger" :icon="Delete" @click="deleteConfirm=''; openDialog('delete')">删除</el-button></div></header>
    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="概览" name="overview">
        <div v-if="isJob" class="metric-grid"><article><span>状态</span><strong :class="statefulSetHealthy ? 'healthy' : 'unhealthy'">● {{ statefulSetStatus }}</strong></article><article><span>完成</span><strong>{{ statefulSetSupport.succeeded }}/{{ statefulSetSupport.completions }}</strong><small>目标完成数</small></article><article><span>运行中</span><strong>{{ statefulSetSupport.active_replicas }}</strong><small>Pod</small></article><article><span>失败</span><strong>{{ statefulSetSupport.failed }}</strong><small>Pod</small></article><article><span>并行数</span><strong>{{ statefulSetSupport.parallelism }}</strong><small>上限</small></article><article><span>创建时间</span><strong>{{ row?.age || '-' }}</strong><small>{{ row?.created_at || '-' }}</small></article></div>
        <div v-else-if="isCronJob" class="metric-grid"><article><span>状态</span><strong :class="statefulSetHealthy ? 'healthy' : 'unhealthy'">● {{ statefulSetStatus }}</strong></article><article><span>调度规则</span><strong>{{ statefulSetSupport.schedule || '-' }}</strong><small>标准 Cron 表达式</small></article><article><span>活动任务</span><strong>{{ statefulSetSupport.active_replicas }}</strong><small>Job</small></article><article><span>上次调度</span><strong>{{ statefulSetSupport.last_schedule || '-' }}</strong><small>最近一次</small></article><article><span>暂停</span><strong>{{ statefulSetSupport.suspended ? '是' : '否' }}</strong><small>{{ statefulSetSupport.concurrency || '-' }}</small></article><article><span>创建时间</span><strong>{{ row?.age || '-' }}</strong><small>{{ row?.created_at || '-' }}</small></article></div>
        <div v-else class="metric-grid"><article><span>状态</span><strong :class="statefulSetHealthy ? 'healthy' : 'unhealthy'">● {{ statefulSetStatus }}</strong></article><article><span>期望</span><strong>{{ desired }}</strong><small>{{ isDaemonSet ? '目标节点 Pod' : '副本数' }}</small></article><article><span>就绪</span><strong>{{ ready }}/{{ desired }}</strong><small>{{ isDaemonSet ? '目标节点 Pod' : '副本数' }}</small></article><article><span>当前</span><strong>{{ current }}</strong><small>Pod</small></article><article><span>已更新</span><strong>{{ updated }}</strong><small>Pod</small></article><article><span>创建时间</span><strong>{{ row?.age || '-' }}</strong><small>{{ row?.created_at || '-' }}</small></article></div>
        <div class="overview-grid"><main><section class="panel"><h2>Pods ({{ pods.length }})</h2><div class="pod-table"><div class="pod-head"><b>Pods</b><b>状态</b><b>就绪</b><b>重启</b><b>节点</b><b>运行时间</b></div><div v-for="pod in pods" :key="pod.name" class="pod-row"><span><b class="pod-link" @click="openPod(pod)">{{ pod.name }}</b><small>{{ pod.pod_ip || '-' }}</small></span><span>{{ statusText(pod.status) }}</span><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><span>{{ pod.node || '-' }}</span><span>{{ pod.age || '-' }}</span></div><el-empty v-if="!pods.length" description="暂无 Pod" :image-size="48"/></div></section>
          <section class="panel"><h2>信息</h2><div class="info-grid">
            <template v-if="isStatefulSet"><span>服务名称</span><b>{{ spec.serviceName || '-' }}</b><span>Pod 管理策略</span><b>{{ spec.podManagementPolicy || 'OrderedReady' }}</b><span>分区</span><b>{{ spec.updateStrategy?.rollingUpdate?.partition ?? 0 }}</b><span>当前版本</span><b>{{ status.currentRevision || '-' }}</b><span>更新版本</span><b>{{ status.updateRevision || '-' }}</b><span>卷声明模板</span><b>{{ claimTemplates.length }} 个</b></template>
            <template v-else-if="isDaemonSet"><span>更新策略</span><b>{{ spec.updateStrategy?.type || 'RollingUpdate' }}</b><span>最大不可用</span><b>{{ spec.updateStrategy?.rollingUpdate?.maxUnavailable ?? '-' }}</b><span>最大扩容</span><b>{{ spec.updateStrategy?.rollingUpdate?.maxSurge ?? '-' }}</b><span>版本历史限制</span><b>{{ spec.revisionHistoryLimit ?? '-' }}</b><span>服务账号</span><b>{{ templateSpec.serviceAccountName || 'default' }}</b><span>主机网络</span><b>{{ templateSpec.hostNetwork ? '是' : '否' }}</b></template>
            <template v-else-if="isJob"><span>并行数</span><b>{{ spec.parallelism ?? 1 }}</b><span>完成数</span><b>{{ spec.completions ?? 1 }}</b><span>失败重试</span><b>{{ spec.backoffLimit ?? 6 }}</b><span>重启策略</span><b>{{ templateSpec.restartPolicy || 'Never' }}</b><span>TTL 清理</span><b>{{ spec.ttlSecondsAfterFinished == null ? '-' : `${spec.ttlSecondsAfterFinished}s` }}</b></template>
            <template v-else><span>调度规则</span><b>{{ spec.schedule || '-' }}</b><span>并发策略</span><b>{{ spec.concurrencyPolicy || 'Allow' }}</b><span>暂停调度</span><b>{{ spec.suspend ? '是' : '否' }}</b><span>保留成功历史</span><b>{{ spec.successfulJobsHistoryLimit ?? 3 }}</b><span>保留失败历史</span><b>{{ spec.failedJobsHistoryLimit ?? 1 }}</b><span>时区</span><b>{{ spec.timeZone || '-' }}</b></template>
            <span>容器</span><b>{{ containers.length }} 个</b><span>初始化容器</span><b>{{ initContainers.length }} 个</b><span>数据卷</span><b>{{ volumes.length }} 个</b>
          </div><div class="selector"><b>Selector</b><code>{{ Object.entries(selector).map(([key,value]) => `${key}=${value}`).join(', ') || '-' }}</code></div><div class="uid-row"><span>UID</span><code>{{ metadata.uid || '-' }}</code></div></section></main>
          <aside><section class="panel"><h2>事件 ({{ statefulSetEvents.length }})</h2><div v-for="item in statefulSetEvents.slice(0,5)" :key="`${item.reason}-${item.lastTimestamp}`" class="event"><b :class="{ warning: item.type==='Warning' }">{{ item.reason }}</b><span>{{ item.message }}</span><small>{{ item.lastTimestamp }}</small></div><p v-if="!statefulSetEvents.length" class="empty">暂无最近事件</p></section><section class="panel"><h2>标签 ({{ Object.keys(labels).length }})</h2><code v-for="(value,key) in labels" :key="key">{{ key }}={{ value }}</code><p v-if="!Object.keys(labels).length" class="empty">暂无标签</p></section><section class="panel"><h2>注解 ({{ Object.keys(annotations).length }})</h2><code v-for="(value,key) in annotations" :key="key">{{ key }}={{ value }}</code><p v-if="!Object.keys(annotations).length" class="empty">暂无注解</p></section></aside></div>
      </el-tab-pane>
      <el-tab-pane :label="`Pods (${pods.length})`" name="pods"><section class="panel pods-panel"><h2>Pods</h2><div class="pods-table"><div class="pods-table-head"><b>Name</b><b>Ready</b><b>Restart</b><b>Status</b><b>IP</b><b>Node</b><b>Created</b></div><div v-for="pod in pods" :key="pod.name" class="pods-table-row"><b class="pod-link" @click="openPod(pod)">{{ pod.name }}</b><span>{{ pod.ready || '-' }}</span><span>{{ pod.restarts ?? 0 }}</span><span>{{ pod.status }}</span><code>{{ pod.pod_ip || '-' }}</code><span>{{ pod.node || '-' }}</span><span>{{ pod.created_at || '-' }} ({{ pod.age || '-' }})</span></div></div><el-empty v-if="!pods.length" description="暂无 Pod"/></section></el-tab-pane>
      <el-tab-pane :label="`容器 (${containers.length})`" name="containers"><section class="panel container-list"><article v-for="container in containers" :key="container.name" class="container-row"><b>{{ container.name }}</b><span>{{ container.image }}</span><small>{{ container.imagePullPolicy || 'IfNotPresent' }}</small></article><el-empty v-if="!containers.length" description="暂无容器"/></section></el-tab-pane>
      <el-tab-pane :label="`卷 (${volumeCount})`" name="volumes"><div class="volume-sections">
        <section v-if="isStatefulSet" class="panel"><h2>持久卷声明 ({{ claims.length }})</h2><div class="volume-table"><div class="volume-table-head"><b>名称</b><b>状态</b><b>容量</b><b>访问模式</b><b>存储类</b><b>绑定卷</b><b>创建时间</b></div><div v-for="claim in claims" :key="claim.name" class="volume-table-row"><b>{{ claim.name }}</b><span>{{ claim.status || '-' }}</span><span>{{ claim.capacity || '-' }}</span><span>{{ claim.access_modes.join(', ') || '-' }}</span><span>{{ claim.storage_class_name || '-' }}</span><code>{{ claim.volume_name || '-' }}</code><span>{{ claim.created_at || '-' }}</span></div></div><el-empty v-if="!claims.length" description="暂无 PVC" :image-size="48"/></section>
        <section v-if="isStatefulSet" class="panel"><h2>卷声明模板 ({{ claimTemplates.length }})</h2><div class="volume-table"><div class="claim-template" v-for="template in claimTemplates" :key="template.metadata?.name"><b>{{ template.metadata?.name || '-' }}</b><span>容量 {{ template.spec?.resources?.requests?.storage || '-' }}</span><span>访问模式 {{ (template.spec?.accessModes || []).join(', ') || '-' }}</span><span>存储类 {{ template.spec?.storageClassName || '-' }}</span></div></div><el-empty v-if="!claimTemplates.length" description="未配置卷声明模板" :image-size="48"/></section>
        <section class="panel"><h2>Pod 数据卷 ({{ volumes.length }})</h2><div class="volume-table"><div class="volume-table-head"><b>名称</b><b>类型</b><b>引用</b></div><div v-for="volume in volumes" :key="volume.name" class="volume-source-row"><b>{{ volume.name }}</b><span>{{ Object.keys(volume).find(key => key !== 'name') || '-' }}</span><code>{{ volumeSource(volume) }}</code></div></div><el-empty v-if="!volumes.length" description="未配置 Pod 数据卷" :image-size="48"/></section>
        <section class="panel"><h2>容器挂载 ({{ volumeMounts.length }})</h2><div class="volume-table"><div class="volume-table-head"><b>容器</b><b>卷</b><b>挂载路径</b><b>子路径</b><b>模式</b></div><div v-for="mount in volumeMounts" :key="`${mount.container}-${mount.name}-${mount.mountPath}`" class="mount-row"><b>{{ mount.container }}</b><span>{{ mount.name }}</span><code>{{ mount.mountPath }}</code><code>{{ mount.subPath || '-' }}</code><span>{{ mount.readOnly ? '只读' : '读写' }}</span></div></div><el-empty v-if="!volumeMounts.length" description="暂无容器挂载" :image-size="48"/></section>
      </div></el-tab-pane>
      <el-tab-pane label="YAML" name="yaml"><KubernetesYamlEditor :model-value="yaml" min-height="calc(100vh - 210px)" readonly/></el-tab-pane>
      <el-tab-pane label="日志" name="logs"><section ref="logPanel" class="console-panel"><header><div class="console-title"><b>Logs</b><span>{{ logLines }} 行</span><i :class="{ loading: logsLoading }"/></div><div class="console-actions"><el-input v-model="logFilter" placeholder="过滤日志…" clearable/><el-select v-model="logPod" class="console-select" filterable placeholder="选择 Pod"><el-option label="全部 Pod" value="__all__"/><el-option v-for="pod in pods" :key="pod.name" :label="pod.name" :value="pod.name"/></el-select><el-select v-model="logContainer" class="console-select" filterable placeholder="选择容器"><el-option label="全部容器" value="__all__"/><el-option v-for="container in allContainers" :key="container.name" :label="container.name" :value="container.name"/></el-select><el-popover placement="bottom-end" :width="250" trigger="click"><template #reference><el-button :icon="Setting" title="日志设置"/></template><div class="console-settings"><label><span>Tail Lines</span><el-select v-model="logTailLines"><el-option label="100" :value="100"/><el-option label="200" :value="200"/><el-option label="500" :value="500"/><el-option label="1000" :value="1000"/><el-option label="全部" :value="0"/></el-select></label><label><span>显示时间戳</span><el-switch v-model="logTimestamps"/></label><label><span>字体大小</span><el-input-number v-model="logFontSize" :min="10" :max="24" controls-position="right"/></label></div></el-popover><el-button title="自动换行" :class="{ active: logWrap }" @click="logWrap=!logWrap">↵</el-button><el-button title="清屏" @click="clearConsole('logs')">≡</el-button><el-button :icon="Download" title="下载日志" @click="downloadLogs"/><el-button :icon="FullScreen" title="全屏" @click="fullscreen(logPanel)"/></div></header><pre :class="['console-output',{ wrap:logWrap }]" :style="{ fontSize: `${logFontSize}px` }">{{ filteredLogs || (logsLoading ? '加载中…' : '暂无日志') }}</pre></section></el-tab-pane>
      <el-tab-pane label="终端" name="terminal"><section ref="terminalPanel" class="console-panel terminal-panel"><header><div class="console-title"><b>&gt;_ Terminal</b><i/><span>↑ {{ formatTraffic(terminalUp) }}</span><span>↓ {{ formatTraffic(terminalDown) }}</span></div><div class="console-actions"><el-select v-model="terminalPod" class="console-select" filterable placeholder="选择 Pod"><el-option v-for="pod in pods" :key="pod.name" :label="pod.name" :value="pod.name"/></el-select><el-select v-model="terminalContainer" class="console-select" filterable placeholder="选择容器"><el-option v-for="container in containers" :key="container.name" :label="container.name" :value="container.name"/></el-select><el-popover placement="bottom-end" :width="220" trigger="click"><template #reference><el-button :icon="Setting" title="终端设置"/></template><div class="console-settings"><label><span>字体大小</span><el-input-number v-model="terminalFontSize" :min="10" :max="24" controls-position="right"/></label></div></el-popover><el-button title="清屏" @click="clearConsole('terminal')">≡</el-button><el-button :icon="FullScreen" title="全屏" @click="fullscreen(terminalPanel)"/></div></header><div ref="terminalHost" class="xterm-host"/><el-empty v-if="!pods.length || !containers.length" description="没有可连接的 Pod 容器" :image-size="48"/></section></el-tab-pane>
      <el-tab-pane :label="`事件 (${statefulSetEvents.length})`" name="events"><section class="panel event-list"><div v-for="item in statefulSetEvents" :key="`${item.reason}-${item.lastTimestamp}`" class="event-row"><b :class="{ warning: item.type==='Warning' }">{{ item.type }}</b><span>{{ item.reason }}</span><span>{{ item.message }}</span><span>{{ item.lastTimestamp }}</span></div><el-empty v-if="!statefulSetEvents.length" description="暂无事件"/></section></el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialogVisible" :title="dialogMode==='describe'?`${kindLabel} 描述`:dialogMode==='yaml'?(yamlMode==='clone'?`克隆 ${kindLabel}`:`编辑 ${kindLabel} YAML`):dialogMode==='scale'?`扩缩容 ${kindLabel}`:`删除 ${kindLabel}`" width="min(800px,calc(100vw - 32px))" align-center destroy-on-close>
      <template v-if="dialogMode==='describe'"><div class="describe-command">kubectl describe {{ kindLabel.toLowerCase() }} -n {{ namespace }} {{ name }}</div><pre class="describe-output">{{ description }}</pre></template>
      <KubernetesYamlEditor v-else-if="dialogMode==='yaml'" v-model="editorYaml"/>
      <div v-else-if="dialogMode==='scale'" class="scale-form"><label>副本数</label><el-input-number v-model="scaleReplicas" :min="0" :max="1000"/></div>
      <div v-else class="delete-form"><p>删除后不可恢复，请输入 <b>{{ name }}</b> 确认。</p><el-input v-model="deleteConfirm" :placeholder="name"/></div>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button v-if="dialogMode==='yaml'" type="primary" :loading="saving" @click="saveYaml">{{ yamlMode==='clone'?'创建副本':'保存' }}</el-button><el-button v-else-if="dialogMode==='scale'" type="primary" :loading="saving" @click="submitScale">扩缩容</el-button><el-button v-else-if="dialogMode==='delete'" type="danger" :loading="deleting" :disabled="deleteConfirm!==name" @click="remove">删除</el-button></template>
    </el-dialog>
  </section>
</template>

<style scoped>
.statefulset-detail{height:100%;min-width:0;overflow:auto;padding:10px 12px;box-sizing:border-box;color:var(--text);font-size:11px;background:#fff}.detail-header,.identity,.actions{display:flex;align-items:center}.detail-header{justify-content:space-between;gap:10px}.identity{gap:4px}.identity h1{margin:0;font-size:20px}.identity p{margin:3px 0 0;color:var(--text-secondary);font-size:12px}.actions{gap:4px;flex-wrap:wrap}.actions .el-button{height:32px;margin:0;padding:0 10px;font-size:11px}.detail-tabs{margin-top:6px}.detail-tabs :deep(.el-tabs__header){margin-bottom:10px}.detail-tabs :deep(.el-tabs__item){height:34px;padding:0 12px;font-size:11px}.metric-grid{display:grid;grid-template-columns:repeat(6,minmax(120px,1fr));gap:9px}.metric-grid article,.panel{min-width:0;border:1px solid var(--line);border-radius:8px;background:#fff}.metric-grid article{min-height:82px;padding:12px;display:flex;flex-direction:column;gap:6px}.metric-grid span,.metric-grid small,.event small,.empty{color:var(--text-secondary)}.metric-grid strong{font-size:18px}.metric-grid .healthy{font-size:15px}.healthy{color:#12b76a}.overview-grid{display:grid;grid-template-columns:minmax(0,2fr) minmax(230px,1fr);gap:9px;margin-top:9px}.overview-grid main,.overview-grid aside{display:flex;min-width:0;flex-direction:column;gap:9px}.panel{padding:10px 12px}.panel h2{margin:0 0 9px;font-size:13px}.pod-head,.pod-row{display:grid;grid-template-columns:minmax(190px,2fr) .8fr .5fr .5fr .8fr .7fr;align-items:center;gap:8px;padding:7px 5px;border-bottom:1px solid var(--line)}.pod-row>span:first-child{display:flex;flex-direction:column}.pod-row small{color:var(--text-secondary)}.pod-link{color:#1683d8;cursor:pointer}.info-grid{display:grid;grid-template-columns:120px 1fr 120px 1fr;gap:10px;padding:5px}.info-grid>span{font-weight:600}.selector,.uid-row{display:grid;grid-template-columns:90px 1fr;align-items:center;gap:10px;padding:10px 5px;border-top:1px solid var(--line)}.selector code,.uid-row code,.panel>code{overflow-wrap:anywhere}.panel>code{display:inline-block;margin:2px;padding:3px 7px;border:1px solid var(--line);border-radius:6px;background:#f7f8fb}.event{display:flex;flex-direction:column;gap:3px;padding:7px 0;border-bottom:1px solid var(--line)}.warning{color:#e5484d}.event-list{min-height:100px}.event-row{display:grid;grid-template-columns:90px 120px minmax(180px,1fr) 170px;gap:10px;padding:9px 5px;border-bottom:1px solid var(--line)}.pods-panel{padding:14px}.pods-table{overflow:auto}.pods-table-head,.pods-table-row{display:grid;grid-template-columns:minmax(220px,2fr) 65px 70px 110px 110px 130px minmax(190px,1.4fr);align-items:center;gap:10px;min-width:1000px;padding:9px 7px;border-bottom:1px solid var(--line)}.pods-table-row{min-height:36px}.container-list{display:flex;flex-direction:column;gap:0}.container-row{display:grid;grid-template-columns:minmax(160px,1fr) minmax(240px,2fr) 130px;gap:12px;padding:11px;border-bottom:1px solid var(--line)}.container-row span{overflow-wrap:anywhere}.container-row small{color:var(--text-secondary)}.describe-command{margin-bottom:10px;font-weight:650}.describe-output{max-height:60vh;margin:0;padding:14px;overflow:auto;border:1px solid var(--line);border-radius:8px;background:#fff;white-space:pre-wrap;word-break:break-word;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,monospace}.scale-form,.delete-form{display:flex;flex-direction:column;gap:10px}.scale-form .el-input-number{width:100%}.delete-form p{margin:0}@media(max-width:900px){.metric-grid{grid-template-columns:repeat(3,minmax(110px,1fr))}.overview-grid{grid-template-columns:1fr}}@media(max-width:600px){.detail-header{align-items:flex-start;flex-direction:column}.actions{overflow:auto;max-width:100%}.metric-grid{grid-template-columns:repeat(2,minmax(100px,1fr))}.info-grid{grid-template-columns:110px 1fr}.event-row{grid-template-columns:70px 100px minmax(180px,1fr)}.event-row span:last-child{grid-column:2/-1}.container-row{grid-template-columns:1fr}.selector,.uid-row{grid-template-columns:1fr}}
.volume-sections{display:flex;flex-direction:column;gap:9px}.volume-table{min-width:0;overflow:auto}.volume-table-head,.volume-table-row,.volume-source-row,.mount-row{display:grid;align-items:center;gap:10px;min-width:780px;padding:9px 7px;border-bottom:1px solid var(--line)}.volume-table-head,.volume-table-row{grid-template-columns:minmax(180px,1.5fr) 90px 90px minmax(120px,1fr) minmax(120px,1fr) minmax(120px,1fr) 160px}.volume-table-head,.claim-template{font-weight:600}.volume-table-row code,.volume-source-row code,.mount-row code{overflow-wrap:anywhere}.claim-template{display:flex;align-items:center;flex-wrap:wrap;gap:10px 24px;min-width:620px;padding:10px 7px;border-bottom:1px solid var(--line)}.volume-source-row{grid-template-columns:minmax(180px,1fr) minmax(150px,1fr) minmax(220px,2fr)}.mount-row{grid-template-columns:minmax(150px,1fr) minmax(150px,1fr) minmax(180px,2fr) minmax(140px,1fr) 80px}.volume-source-row code,.mount-row code{font-size:10px}.console-panel{overflow:hidden;border:1px solid var(--line);border-radius:8px;background:#fff}.console-panel>header{display:flex;align-items:center;justify-content:space-between;gap:10px;padding:9px 12px}.console-title,.console-actions{display:flex;align-items:center;gap:7px}.console-title{white-space:nowrap}.console-title b{font-size:13px}.console-title span{color:var(--text-secondary)}.console-title i{width:8px;height:8px;border-radius:50%;background:#58d68d}.console-title i.loading{animation:pulse 1s infinite}.console-actions{min-width:0;overflow:auto}.console-actions .el-input{width:170px}.console-select{width:190px;flex:none}.console-actions .el-button{margin-left:0}.console-actions .el-button.active{color:var(--primary);border-color:var(--primary);background:var(--primary-soft)}.console-output{min-height:calc(100vh - 270px);max-height:calc(100vh - 270px);margin:0;padding:12px;overflow:auto;background:#17191e;color:#d8f3df;white-space:pre;line-height:1.55;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}.console-output.wrap{white-space:pre-wrap;word-break:break-all}.console-panel:fullscreen{display:flex;flex-direction:column;background:#fff}.console-panel:fullscreen .console-output,.console-panel:fullscreen .xterm-host{flex:1;max-height:none}.terminal-panel{background:#050505}.terminal-panel>header{background:#fff}.terminal-panel .console-title i{background:#58d68d}.xterm-host{height:calc(100vh - 270px);min-height:360px;padding:6px 8px;box-sizing:border-box;background:#050505}.terminal-panel:fullscreen .xterm-host{height:auto}.console-settings{display:flex;flex-direction:column;gap:12px;padding:4px}.console-settings label{display:flex;align-items:center;justify-content:space-between;gap:8px}.console-settings label>span{font-size:11px;font-weight:600}.console-settings .el-select,.console-settings .el-input-number{width:120px}@keyframes pulse{50%{opacity:.35}}@media(max-width:850px){.console-panel>header{align-items:flex-start;flex-direction:column}.console-actions{width:100%}.console-select{width:160px}}@media(max-width:600px){.volume-sections .panel{padding:9px}.console-output{min-height:calc(100vh - 300px);max-height:calc(100vh - 300px)}.xterm-host{height:calc(100vh - 300px)}}
</style>