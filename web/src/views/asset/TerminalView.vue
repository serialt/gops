<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, type ElTree } from 'element-plus'
import { ArrowDown, ArrowLeft, ArrowRight, Close, CollectionTag, Coin, Delete, Document, Download, Folder, FolderOpened, Monitor, Platform, Plus, Refresh, Search, Upload } from '@element-plus/icons-vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import gopsLogoUrl from '@/assets/gops.svg'
import { downloadSftp, getSftpTransferProgress, getTerminalAssetTree, listSftp, uploadSftp, type AssetTreeNode, type DbInstance, type Host, type SftpEntry } from '@/api/asset'
import { useSystemConfigStore } from '@/stores/system-config'
import { downloadAttachment } from '@/utils/download'

type TreeNode = AssetTreeNode
interface Session { key: string; kind: 'host' | 'database'; label: string; host?: Host; database?: DbInstance }
interface Runtime { terminal: Terminal; fitAddon: FitAddon; socket: WebSocket; resizeObserver: ResizeObserver; connectTimer: ReturnType<typeof setTimeout>; heartbeatTimer: ReturnType<typeof setInterval> }
interface SftpTransferRecord {
  id: number
  name: string
  path: string
  createdAt: string
  status: 'processing' | 'success' | 'failed'
  progress: number
}

const systemConfig = useSystemConfigStore()
const treeData = ref<TreeNode[]>([])
const assetCount = computed(() => treeData.value.reduce((total, node) => total + (node.count || 0), 0))
const keyword = ref('')
const treeRef = ref<InstanceType<typeof ElTree>>()
const sessions = ref<Session[]>([])
const activeKey = ref<string | null>(null)
const assetPaneWidth = ref(180)
const assetPaneDragging = ref(false)
const assetPaneStartX = ref(0)
const assetPaneStartWidth = ref(180)
const terminalElements = new Map<string, HTMLElement>()
const runtimes = new Map<string, Runtime>()
const sftpVisible = ref(false)
const sftpOriginKey = ref<number | null>(null)
const sftpPath = ref('/root')
const sftpCommittedPath = ref('/root')
const sftpKeyword = ref('')
const sftpSearchVisible = ref(false)
const sftpSearchInput = ref<HTMLInputElement>()
const sftpUploadMenuVisible = ref(false)
const sftpFileInput = ref<HTMLInputElement>()
const sftpDirectoryInput = ref<HTMLInputElement>()
const sftpUploading = ref(false)
const sftpSelected = ref<string[]>([])
const sftpUploadRecords = ref<SftpTransferRecord[]>([])
const sftpDownloadRecords = ref<SftpTransferRecord[]>([])
const sftpRecordPanel = ref<'upload' | 'download' | null>(null)
const sftpDownloadNames = new Set<string>()
let sftpTransferRecordID = 0
type SftpColumn = 'size' | 'modified' | 'permissions'
const sftpColumnWidths = ref({ size: 144, modified: 210, permissions: 150 })
const sftpResizingColumn = ref<SftpColumn | null>(null)
const sftpResizeStartX = ref(0)
const sftpResizeStartWidth = ref(0)

const sftpEntries = ref<SftpEntry[]>([])
const filteredSftpEntries = computed(() => {
  const value = sftpKeyword.value.trim().toLowerCase()
  return value ? sftpEntries.value.filter(entry => entry.name.toLowerCase().includes(value)) : sftpEntries.value
})
const activeSession = computed(() => sessions.value.find(session => session.key === activeKey.value))

function setTerminalElement(key: string, element: unknown) {
  if (element instanceof HTMLElement) terminalElements.set(key, element)
  else terminalElements.delete(key)
}
function filterTree(value: string, data: TreeNode) {
  if (!value.trim()) return true
  const assetType = data.type === 'host' ? '主机' : data.database ? databaseTypeLabel(data.database.db_type) : ''
  return `${data.label} ${assetType} ${data.host?.ip || ''} ${data.database?.host || ''} ${data.database?.db_type || ''}`.toLowerCase().includes(value.trim().toLowerCase())
}
function databaseTypeLabel(type: string) {
  const normalized = type.trim().toLowerCase().replace(/[\s_-]+/g, '')
  const labels: Record<string, string> = {
    postgresql: 'PostgreSql',
    postgres: 'PostgreSql',
    mysql: 'MySQL',
    oracle: 'Oracle',
    sqlserver: 'SQL Server',
    mssql: 'SQL Server',
    mariadb: 'MariaDB',
  }
  return labels[normalized] || type
}
function startAssetPaneDrag(event: PointerEvent) {
  if (event.button !== 0) return
  assetPaneDragging.value = true
  assetPaneStartX.value = event.clientX
  assetPaneStartWidth.value = assetPaneWidth.value
  document.addEventListener('pointermove', moveAssetPane)
  document.addEventListener('pointerup', stopAssetPaneDrag)
}
function moveAssetPane(event: PointerEvent) {
  if (!assetPaneDragging.value) return
  assetPaneWidth.value = Math.max(0, Math.min(480, assetPaneStartWidth.value + event.clientX - assetPaneStartX.value))
}
function stopAssetPaneDrag() {
  if (!assetPaneDragging.value) return
  assetPaneDragging.value = false
  if (assetPaneWidth.value < 120) assetPaneWidth.value = 0
  document.removeEventListener('pointermove', moveAssetPane)
  document.removeEventListener('pointerup', stopAssetPaneDrag)
  window.getSelection()?.removeAllRanges()
}
function toggleAssetPane() { assetPaneWidth.value = assetPaneWidth.value === 0 ? 260 : 0 }
function socketURL(session: Session, terminal: Terminal) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token') || ''
  const endpoint = session.kind === 'host'
    ? `/api/v1/assets/terminal/connect/${session.host?.id}`
    : `/api/v1/assets/terminal/database/connect/${session.database?.id}`
  return `${protocol}//${window.location.host}${endpoint}?token=${encodeURIComponent(token)}&cols=${terminal.cols}&rows=${terminal.rows}`
}
function sendResize(runtime: Runtime) {
  if (runtime.socket.readyState === WebSocket.OPEN) runtime.socket.send(JSON.stringify({ type: 'resize', cols: runtime.terminal.cols, rows: runtime.terminal.rows }))
}
async function initializeSession(session: Session) {
  await nextTick()
  const element = terminalElements.get(session.key)
  if (!element || runtimes.has(session.key)) return
  const terminal = new Terminal({ cursorBlink: true, fontSize: 11, lineHeight: 1.2, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace', theme: { background: '#1d1e20', foreground: '#d7dae0', cursor: '#f5f5f5', selectionBackground: '#42556d' } })
  const fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(element)
  fitAddon.fit()
  const target = session.host
    ? `${session.host.name}（${session.host.ip}:${session.host.port}）`
    : `${session.database?.name}（${session.database?.host}:${session.database?.port}）`
  terminal.writeln(`正在连接 ${target}...`)
  const socket = new WebSocket(socketURL(session, terminal))
  socket.binaryType = 'arraybuffer'
  const resizeObserver = new ResizeObserver(() => { fitAddon.fit(); const runtime = runtimes.get(session.key); if (runtime) sendResize(runtime) })
  const connectTimer = setTimeout(() => {
    if (socket.readyState !== WebSocket.OPEN) {
      terminal.writeln('\r\n连接超时，请检查后端服务、代理和网络。')
      socket.close()
    }
  }, 15000)
  const heartbeatTimer = setInterval(() => {
    if (socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: 'heartbeat' }))
  }, 25000)
  const runtime = { terminal, fitAddon, socket, resizeObserver, connectTimer, heartbeatTimer }
  runtimes.set(session.key, runtime)
  terminal.onData(data => socket.readyState === WebSocket.OPEN && socket.send(data))
  socket.onopen = () => { clearTimeout(connectTimer); fitAddon.fit(); sendResize(runtime); terminal.focus() }
  socket.onmessage = event => terminal.write(event.data instanceof ArrayBuffer ? new Uint8Array(event.data) : String(event.data))
  socket.onerror = () => terminal.writeln('\r\n连接失败，请检查主机、凭证和网络。')
  socket.onclose = () => terminal.writeln('\r\n连接已关闭。')
  resizeObserver.observe(element)
}
async function openTerminal(host?: Host) {
  if (!host) return
  const key = `host:${host.id}`
  const existing = sessions.value.find(session => session.key === key)
  if (existing) { activateSession(existing.key); return }
  const session: Session = { key, kind: 'host', label: host.name, host }
  sessions.value.push(session)
  activeKey.value = session.key
  await initializeSession(session)
}
async function openDatabaseTerminal(database?: DbInstance) {
  if (!database) return
  const key = `database:${database.id}`
  const existing = sessions.value.find(session => session.key === key)
  if (existing) { activateSession(existing.key); return }
  const session: Session = { key, kind: 'database', label: database.name, database }
  sessions.value.push(session)
  activeKey.value = session.key
  await initializeSession(session)
}
function activateSession(key: string) {
  activeKey.value = key
  void nextTick(() => { runtimes.get(key)?.fitAddon.fit(); runtimes.get(key)?.terminal.focus() })
}
function clearTerminal() {
  const runtime = activeKey.value === null ? undefined : runtimes.get(activeKey.value)
  runtime?.terminal.clear()
  runtime?.terminal.focus()
}
function openSftp() {
  const session = activeSession.value
  if (session?.kind !== 'host' || !session.host) return
  sftpOriginKey.value = session.host.id
  sftpVisible.value = true
  sftpSearchVisible.value = false
  sftpUploadMenuVisible.value = false
  sftpRecordPanel.value = null
  sftpSelected.value = []
  sftpKeyword.value = ''
  sftpPath.value = '/root'
  sftpCommittedPath.value = '/root'
  void loadSftpDirectory()
}
function closeSftp() {
  const originKey = sftpOriginKey.value
  sftpVisible.value = false
  sftpOriginKey.value = null
  sftpSearchVisible.value = false
  sftpUploadMenuVisible.value = false
  sftpRecordPanel.value = null
  sftpSelected.value = []
  const sessionKey = originKey === null ? null : `host:${originKey}`
  if (sessionKey !== null && sessions.value.some(session => session.key === sessionKey)) {
    activeKey.value = sessionKey
    void nextTick(() => { runtimes.get(sessionKey)?.fitAddon.fit(); runtimes.get(sessionKey)?.terminal.focus() })
  }
}
function goSftpParent() {
  if (sftpCommittedPath.value === '/') return
  const parent = sftpCommittedPath.value.replace(/\/+$/, '').split('/').slice(0, -1).join('/')
  void navigateSftpPath(parent || '/')
}
async function openSftpEntry(entry: SftpEntry) {
  if (entry.type !== 'directory') return
  const nextPath = `${sftpPath.value.replace(/\/$/, '')}/${entry.name}`
  sftpSelected.value = []
  await navigateSftpPath(nextPath)
}
function toggleSftpEntry(name: string) {
  sftpSelected.value = sftpSelected.value.includes(name)
    ? sftpSelected.value.filter(item => item !== name)
    : [...sftpSelected.value, name]
}
function startSftpColumnResize(event: PointerEvent, column: SftpColumn) {
  if (event.button !== 0) return
  event.preventDefault()
  sftpResizingColumn.value = column
  sftpResizeStartX.value = event.clientX
  sftpResizeStartWidth.value = sftpColumnWidths.value[column]
  document.addEventListener('pointermove', moveSftpColumnResize)
  document.addEventListener('pointerup', stopSftpColumnResize)
  document.addEventListener('pointercancel', stopSftpColumnResize)
}
function moveSftpColumnResize(event: PointerEvent) {
  const column = sftpResizingColumn.value
  if (!column) return
  const minimum = column === 'size' ? 80 : 100
  const maximum = column === 'modified' ? 480 : 360
  sftpColumnWidths.value[column] = Math.max(minimum, Math.min(maximum, sftpResizeStartWidth.value + event.clientX - sftpResizeStartX.value))
}
function stopSftpColumnResize() {
  if (!sftpResizingColumn.value) return
  sftpResizingColumn.value = null
  document.removeEventListener('pointermove', moveSftpColumnResize)
  document.removeEventListener('pointerup', stopSftpColumnResize)
  document.removeEventListener('pointercancel', stopSftpColumnResize)
  window.getSelection()?.removeAllRanges()
}
function sftpAction(message: string) {
  ElMessage.info(message)
}
function createSftpTransferRecord(name: string): SftpTransferRecord {
  return reactive({
    id: ++sftpTransferRecordID,
    name,
    path: sftpCommittedPath.value,
    createdAt: new Date().toLocaleString('zh-CN', { hour12: false }),
    status: 'processing',
    progress: 0,
  })
}
function nextSftpDownloadName(filename: string) {
  const dotIndex = filename.lastIndexOf('.')
  const stem = dotIndex > 0 ? filename.slice(0, dotIndex) : filename
  const extension = dotIndex > 0 ? filename.slice(dotIndex) : ''
  let candidate = filename
  let index = 1
  while (sftpDownloadNames.has(candidate)) candidate = `${stem}_${index++}${extension}`
  sftpDownloadNames.add(candidate)
  return candidate
}
function toggleSftpRecordPanel(panel: 'upload' | 'download') {
  sftpUploadMenuVisible.value = false
  sftpRecordPanel.value = sftpRecordPanel.value === panel ? null : panel
}
async function trackSftpTransfer(transferID: string, records: SftpTransferRecord[], base = 0, span = 100) {
  while (records.some(record => record.status === 'processing')) {
    try {
      const result = await getSftpTransferProgress(transferID)
      const progress = result.data.data
      const percentage = progress.total > 0 ? Math.min(100, Math.round((progress.transferred / progress.total) * 100)) : 0
      records.forEach(record => { record.progress = Math.max(record.progress, Math.round(base + percentage * span / 100)) })
      if (progress.status === 'success') {
        records.forEach(record => { record.progress = 100; record.status = 'success' })
        return
      }
      if (progress.status === 'failed') {
        records.forEach(record => { record.status = 'failed' })
        return
      }
    } catch {
      // The transfer request owns user-facing error reporting; polling retries quietly.
    }
    await new Promise(resolve => window.setTimeout(resolve, 250))
  }
}
function chooseSftpUpload(kind: 'file' | 'directory') {
  sftpUploadMenuVisible.value = false
  if (kind === 'file') sftpFileInput.value?.click()
  else sftpDirectoryInput.value?.click()
}
async function handleSftpUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (sftpOriginKey.value === null || !files.length) return
  const records = files.map(file => createSftpTransferRecord(file.webkitRelativePath || file.name))
  sftpUploadRecords.value.unshift(...records)
  sftpRecordPanel.value = 'upload'
  sftpUploading.value = true
  const transferID = crypto.randomUUID()
  void trackSftpTransfer(transferID, records, 20, 80)
  try {
    await uploadSftp(sftpOriginKey.value, sftpCommittedPath.value, files, transferID, progress => {
      records.forEach(record => { record.progress = Math.max(record.progress, Math.round(progress * .2)) })
    })
    records.forEach(record => { record.progress = 100; record.status = 'success' })
    ElMessage.success('上传成功')
    await loadSftpDirectory()
  } catch (error) {
    records.forEach(record => { record.status = 'failed' })
    throw error
  } finally {
    sftpUploading.value = false
  }
}
async function downloadSelectedSftp() {
  if (sftpOriginKey.value === null || !sftpSelected.value.length) return
  const names = [...sftpSelected.value]
  const record = createSftpTransferRecord(names.length === 1 ? names[0] : `${names.length} 个文件`)
  sftpDownloadRecords.value.unshift(record)
  sftpRecordPanel.value = 'download'
  const transferID = crypto.randomUUID()
  void trackSftpTransfer(transferID, [record])
  try {
    const response = await downloadSftp(sftpOriginKey.value, sftpCommittedPath.value, names, transferID)
    downloadAttachment(response.data as unknown as Blob, undefined, nextSftpDownloadName('sftp-download.zip'))
    record.progress = 100
    record.status = 'success'
  } catch (error) {
    record.status = 'failed'
    throw error
  }
}
async function loadSftpDirectory(path = sftpPath.value) {
  if (sftpOriginKey.value === null) return
  try {
    const result = await listSftp(sftpOriginKey.value, path)
    sftpEntries.value = result.data.data.list
    return true
  } catch {
    sftpEntries.value = []
    return false
  }
}
async function navigateSftpPath(path: string) {
  const previousPath = sftpCommittedPath.value
  const normalizedPath = path.trim().replace(/\/{2,}/g, '/') || '/'
  sftpPath.value = normalizedPath
  if (!await loadSftpDirectory(normalizedPath)) {
    sftpPath.value = previousPath
    ElMessage.error('目录不存在或没有访问权限')
    return
  }
  sftpCommittedPath.value = normalizedPath
}
function toggleSftpSearch() {
  sftpSearchVisible.value = !sftpSearchVisible.value
  if (sftpSearchVisible.value) void nextTick(() => sftpSearchInput.value?.focus())
}
async function reconnectSession() {
  if (activeKey.value === null) return
  const session = sessions.value.find(item => item.key === activeKey.value)
  const runtime = runtimes.get(activeKey.value)
  if (!session || !runtime) return
  clearTimeout(runtime.connectTimer)
  clearInterval(runtime.heartbeatTimer)
  runtime.resizeObserver.disconnect()
  runtime.socket.close()
  runtime.terminal.dispose()
  runtimes.delete(activeKey.value)
  await initializeSession(session)
}
function disposeSession(key: string) {
  const runtime = runtimes.get(key)
  if (runtime) clearTimeout(runtime.connectTimer)
  if (runtime) clearInterval(runtime.heartbeatTimer)
  runtime?.resizeObserver.disconnect()
  runtime?.socket.close()
  runtime?.terminal.dispose()
  runtimes.delete(key)
  terminalElements.delete(key)
}
function closeSession(key: string) {
  const index = sessions.value.findIndex(session => session.key === key)
  if (index < 0) return
  disposeSession(key)
  sessions.value.splice(index, 1)
  if (activeKey.value === key) activeKey.value = sessions.value[Math.min(index, sessions.value.length - 1)]?.key ?? null
}
async function load() {
  const result = await getTerminalAssetTree()
  treeData.value = result.data.data
}
watch(keyword, value => treeRef.value?.filter(value))
onMounted(() => { void systemConfig.fetchConfig(); void load().catch(() => ElMessage.error('读取资产失败')) })
onBeforeUnmount(() => {
  sessions.value.forEach(session => disposeSession(session.key))
  document.removeEventListener('pointermove', moveAssetPane)
  document.removeEventListener('pointerup', stopAssetPaneDrag)
  stopSftpColumnResize()
})
</script>

<template>
  <div class="web-terminal-page">
    <header class="terminal-header"><a class="terminal-brand" href="/dashboard"><img :src="gopsLogoUrl"
          :alt="systemConfig.systemName" /></a><a href="/dashboard">首页</a><span>/</span><strong>Web终端</strong></header>
    <main class="terminal-workbench" :style="{ gridTemplateColumns: `${assetPaneWidth}px 2px minmax(0, 1fr)` }">
      <aside class="asset-panel">
        <div class="asset-title"><span><el-icon>
              <CollectionTag />
            </el-icon>资产分组</span><small>({{ assetCount }})</small></div>
        <el-input v-model="keyword" class="asset-search" placeholder="搜索" clearable :prefix-icon="Search" />
        <el-tree ref="treeRef" class="asset-tree" :data="treeData" node-key="id" default-expand-all
          :filter-node-method="filterTree" :expand-on-click-node="true">
          <template #default="{ data, node }"><span class="tree-node"
              :class="[`is-${data.type}`, { 'asset-tree-leaf': data.type !== 'group' }]"
              @dblclick.stop="data.type === 'host' ? openTerminal(data.host) : data.type === 'database' && openDatabaseTerminal(data.database)"><el-icon
                v-if="data.type === 'group'" class="group-folder-icon" @click.stop="node.expanded = !node.expanded">
                <FolderOpened v-if="node.expanded" />
                <Folder v-else />
              </el-icon><el-icon v-else-if="data.type === 'host'" class="host-asset-icon" title="主机" aria-label="主机">
                <Platform />
              </el-icon><el-icon v-else class="database-asset-icon"
                :title="databaseTypeLabel(data.database?.db_type || '')"
                :aria-label="databaseTypeLabel(data.database?.db_type || '')">
                <Coin />
              </el-icon><span class="tree-label">{{ data.label }}</span><small v-if="data.type === 'database'"
                class="database-type-badge">{{ databaseTypeLabel(data.database?.db_type || '') }}</small><small
                v-else-if="data.type === 'host'" class="database-type-badge">Host</small><small
                v-if="data.type === 'group'">({{
                  data.count }})</small><i v-else-if="data.type === 'host'"
                :class="{ online: data.host?.status === 1 }" /><i v-else-if="data.type === 'database'"
                :class="{ online: data.database?.status }" /></span></template>
        </el-tree>
      </aside>
      <div class="asset-divider" :class="{ dragging: assetPaneDragging }" role="separator" aria-label="调整资产分组宽度"
        @pointerdown="startAssetPaneDrag" @dblclick="toggleAssetPane">
        <button type="button" class="asset-divider-handle" :aria-label="assetPaneWidth ? '收起资产分组' : '展开资产分组'"
          @pointerdown.stop @click.stop="toggleAssetPane"><el-icon>
            <component :is="assetPaneWidth ? ArrowLeft : ArrowRight" />
          </el-icon></button>
      </div>
      <section class="terminal-stage">
        <nav class="session-tabs">
          <div v-if="!sessions.length" class="session-tab placeholder"><el-icon>
              <Monitor />
            </el-icon>新建终端<span /></div>
          <button v-for="session in sessions" :key="session.key" type="button" class="session-tab"
            :class="{ active: activeKey === session.key }" @click="activateSession(session.key)"><el-icon>
              <component :is="session.kind === 'host' ? Platform : Coin" />
            </el-icon><span>{{ session.label }}</span><i v-if="session.host"
              :class="{ online: session.host.status === 1 }" /><el-icon class="close"
              @click.stop="closeSession(session.key)">
              <Close />
            </el-icon></button>
          <div class="terminal-actions">
            <button type="button" class="terminal-action" :disabled="activeSession?.kind !== 'host'"
              aria-label="SFTP文件管理" title="SFTP文件管理" @click="openSftp"><el-icon>
                <Folder />
              </el-icon></button>
            <button type="button" class="terminal-action" :disabled="activeKey === null" aria-label="重新连接" title="重新连接"
              @click="void reconnectSession()"><el-icon>
                <Refresh />
              </el-icon></button>
            <button type="button" class="terminal-action" :disabled="activeKey === null" aria-label="清空终端" title="清空终端"
              @click="clearTerminal"><el-icon>
                <Delete />
              </el-icon></button>
          </div>
        </nav>
        <div v-show="!sftpVisible" class="session-content">
          <div v-if="!sessions.length" class="terminal-empty"><el-icon>
              <Monitor />
            </el-icon><strong>双击资产打开连接</strong><span>在左侧资产分组中选择主机或数据库</span></div>
          <div v-for="session in sessions" v-show="activeKey === session.key" :key="session.key"
            :ref="element => setTerminalElement(session.key, element)" class="terminal-body" />
        </div>
        <div v-if="sftpVisible" class="sftp-overlay">
          <section class="sftp-panel" aria-label="SFTP 文件管理"
            @click="sftpSearchVisible = false; sftpUploadMenuVisible = false; sftpRecordPanel = null">
            <header class="sftp-toolbar">
              <strong>SFTP</strong>
              <button type="button" class="sftp-icon-button" aria-label="返回上级目录" title="返回上级目录"
                @click="goSftpParent"><el-icon>
                  <ArrowLeft />
                </el-icon></button>
              <input v-model="sftpPath" class="sftp-path" aria-label="当前目录路径" title="输入路径后按回车跳转"
                @keydown.enter.prevent="void navigateSftpPath(sftpPath)" />
              <button type="button" class="sftp-icon-button" aria-label="刷新目录" title="刷新目录"
                @click="void loadSftpDirectory()"><el-icon>
                  <Refresh />
                </el-icon></button>
              <div class="sftp-search-wrap" @click.stop>
                <button type="button" class="sftp-icon-button" aria-label="搜索当前目录" title="搜索当前目录"
                  @click="toggleSftpSearch"><el-icon>
                    <Search />
                  </el-icon></button>
                <label v-if="sftpSearchVisible" class="sftp-search-popup"><el-icon>
                    <Search />
                  </el-icon><input ref="sftpSearchInput" v-model="sftpKeyword" placeholder="搜索当前目录" /></label>
              </div>
              <button type="button" class="sftp-button" @click="sftpAction('新建功能尚未配置')"><el-icon>
                  <Plus />
                </el-icon>新建</button>
              <div class="sftp-split-wrap" @click.stop>
                <div class="sftp-split-button">
                  <button type="button" class="sftp-button sftp-split-main" :disabled="sftpUploading"
                  @click="sftpUploadMenuVisible = !sftpUploadMenuVisible"><el-icon>
                    <Upload />
                  </el-icon>{{
                    sftpUploading ? '上传中' : '上传' }}</button>
                  <button type="button" class="sftp-button sftp-split-toggle" aria-label="查看上传记录" title="上传记录"
                    @click="toggleSftpRecordPanel('upload')"><el-icon><ArrowDown /></el-icon></button>
                </div>
                <div v-if="sftpUploadMenuVisible" class="sftp-upload-menu">
                  <button type="button" @click="chooseSftpUpload('file')"><el-icon>
                      <Upload />
                    </el-icon>选择文件</button>
                  <button type="button" @click="chooseSftpUpload('directory')"><el-icon>
                      <Folder />
                    </el-icon>选择目录</button>
                </div>
                <input ref="sftpFileInput" class="sftp-hidden-input" type="file" multiple @change="handleSftpUpload" />
                <input ref="sftpDirectoryInput" class="sftp-hidden-input" type="file" webkitdirectory multiple
                  @change="handleSftpUpload" />
                <div v-if="sftpRecordPanel === 'upload'" class="sftp-record-panel">
                  <div class="sftp-record-title"><strong>上传记录</strong><button v-if="sftpUploadRecords.length" type="button"
                      @click="sftpUploadRecords = []">清空</button></div>
                  <div v-if="!sftpUploadRecords.length" class="sftp-record-empty">暂无上传记录</div>
                  <div v-else class="sftp-record-list">
                    <div v-for="record in sftpUploadRecords" :key="record.id" class="sftp-record-item">
                      <div><strong :title="record.name">{{ record.name }}</strong><span>{{ record.path }}</span></div>
                      <div><span>{{ record.createdAt }}</span><em :class="`is-${record.status}`">{{ record.status ===
                        'processing' ? `上传中 ${record.progress}%` : record.status === 'success' ? '成功' : '失败' }}</em></div>
                      <div class="sftp-record-progress" :class="{ 'is-failed': record.status === 'failed' }"><i
                          :style="{ width: `${record.progress}%` }" /></div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="sftp-split-wrap" @click.stop>
                <div class="sftp-split-button">
                  <button type="button" class="sftp-button sftp-split-main" :disabled="!sftpSelected.length"
                    @click="void downloadSelectedSftp()"><el-icon>
                      <Download />
                    </el-icon>下载</button>
                  <button type="button" class="sftp-button sftp-split-toggle" aria-label="查看下载记录" title="下载记录"
                    @click="toggleSftpRecordPanel('download')"><el-icon><ArrowDown /></el-icon></button>
                </div>
                <div v-if="sftpRecordPanel === 'download'" class="sftp-record-panel is-download">
                  <div class="sftp-record-title"><strong>下载记录</strong><button v-if="sftpDownloadRecords.length" type="button"
                      @click="sftpDownloadRecords = []">清空</button></div>
                  <div v-if="!sftpDownloadRecords.length" class="sftp-record-empty">暂无下载记录</div>
                  <div v-else class="sftp-record-list">
                    <div v-for="record in sftpDownloadRecords" :key="record.id" class="sftp-record-item">
                      <div><strong :title="record.name">{{ record.name }}</strong><span>{{ record.path }}</span></div>
                      <div><span>{{ record.createdAt }}</span><em :class="`is-${record.status}`">{{ record.status ===
                        'processing' ? (record.progress >= 0 ? `下载中 ${record.progress}%` : '下载中') : record.status ===
                          'success' ? '成功' : '失败' }}</em></div>
                      <div class="sftp-record-progress"
                        :class="{ 'is-failed': record.status === 'failed', 'is-indeterminate': record.progress < 0 }"><i
                          :style="{ width: record.progress < 0 ? '35%' : `${record.progress}%` }" /></div>
                    </div>
                  </div>
                </div>
              </div>
              <button type="button" class="sftp-icon-button sftp-close" aria-label="关闭 SFTP" title="关闭 SFTP"
                @click="closeSftp"><el-icon>
                  <Close />
                </el-icon></button>
            </header>
            <div class="sftp-table-scroll" :class="{ 'is-resizing': sftpResizingColumn }"
              :style="{ '--sftp-size-width': `${sftpColumnWidths.size}px`, '--sftp-modified-width': `${sftpColumnWidths.modified}px`, '--sftp-permissions-width': `${sftpColumnWidths.permissions}px` }">
              <div class="sftp-table-head"><span class="sftp-check-spacer" /><span class="sftp-name-head">名称</span>
                <span class="sftp-column-head">大小<button type="button" class="sftp-resize-handle" aria-label="调整大小列宽度"
                    title="拖动调整列宽" @pointerdown.stop="startSftpColumnResize($event, 'size')" /></span>
                <span class="sftp-column-head">修改时间<button type="button" class="sftp-resize-handle"
                    aria-label="调整修改时间列宽度" title="拖动调整列宽"
                    @pointerdown.stop="startSftpColumnResize($event, 'modified')" /></span>
                <span class="sftp-column-head">权限<button type="button" class="sftp-resize-handle" aria-label="调整权限列宽度"
                    title="拖动调整列宽" @pointerdown.stop="startSftpColumnResize($event, 'permissions')" /></span>
              </div>
              <div class="sftp-list">
                <button v-for="entry in filteredSftpEntries" :key="entry.name" type="button" class="sftp-row"
                  :class="{ selected: sftpSelected.includes(entry.name) }" @dblclick="openSftpEntry(entry)"
                  @click="toggleSftpEntry(entry.name)">
                  <span class="sftp-check" :class="{ checked: sftpSelected.includes(entry.name) }" />
                  <span class="sftp-name"><el-icon>
                      <FolderOpened v-if="entry.type === 'directory'" />
                      <Document v-else />
                    </el-icon>{{ entry.name }}</span>
                  <span>{{ entry.size }}</span><span>{{ entry.modified }}</span><span>{{ entry.permissions }}</span>
                </button>
                <div v-if="!filteredSftpEntries.length" class="sftp-empty">暂无目录数据</div>
              </div>
            </div>
          </section>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.web-terminal-page {
  height: 100vh;
  min-width: 720px;
  overflow: hidden;
  background: #f5f6f8;
  color: #1f2430;
  font-size: 11px
}

.terminal-header {
  height: 48px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  box-sizing: border-box;
  font-size: 13px
}

.terminal-header a {
  color: #242833;
  text-decoration: none;
  font-weight: 700
}

.terminal-header>span {
  color: #a4a9b3
}

.terminal-brand {
  display: inline-flex;
  margin-right: 5px
}

.terminal-brand img {
  width: 32px;
  height: 32px;
  border-radius: 7px
}

.terminal-workbench {
  height: calc(100vh - 48px);
  display: grid;
  padding: 4px;
  box-sizing: border-box;
  background: #f2f3f5
}

.asset-panel {
  min-width: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  color: #c8cbd1;
  background: #252628;
  border-radius: 8px
}

.asset-title {
  position: relative;
  height: 42px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  font-weight: 700
}

.asset-title::after {
  position: absolute;
  right: 8px;
  bottom: 0;
  left: 8px;
  height: 1px;
  border-radius: 2px;
  background: #393a3e;
  content: ''
}

.asset-title>span {
  display: inline-flex;
  align-items: center;
  gap: 8px
}

.asset-title .el-icon {
  color: #55cbb5;
  font-size: 16px
}

.asset-title small {
  color: #8c9098
}

.asset-search {
  padding: 8px 10px 7px;
  box-sizing: border-box;
  background: #252628
}

.asset-search :deep(.el-input__wrapper) {
  background: #343537 !important;
  border: 1px solid #3b3c40;
  box-shadow: none !important;
  border-radius: 6px
}

.asset-search :deep(.el-input__wrapper.is-focus),
.asset-search :deep(.el-input__wrapper:focus-within) {
  border-color: #55cbb5;
  box-shadow: 0 0 0 1px rgb(85 203 181 / 18%) !important
}

.asset-search :deep(.el-input__inner) {
  color: #fff !important;
  caret-color: #fff;
  -webkit-text-fill-color: #fff
}

.asset-search :deep(.el-input__inner::placeholder) {
  color: #92969f
}

.asset-divider {
  position: relative;
  z-index: 4;
  margin: 8px 0;
  background: #3b3c40;
  border-radius: 6px;
  cursor: col-resize;
  touch-action: none
}

.asset-divider:hover,
.asset-divider.dragging {
  background: var(--primary)
}

.asset-divider-handle {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 20px;
  height: 38px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transform: translate(-50%, -50%);
  border: 1px solid #4b4d53;
  border-radius: 6px;
  color: #bfc3ca;
  background: #2b2c2f;
  cursor: pointer;
  opacity: .55;
  transition: opacity .15s, background-color .15s, color .15s
}

.asset-divider:hover .asset-divider-handle,
.asset-divider-handle:focus-visible,
.asset-divider-handle:hover {
  opacity: 1
}

.asset-divider-handle:hover {
  color: #fff;
  background: var(--primary)
}

.asset-tree {
  flex: 1;
  overflow: auto;
  padding: 3px 8px 12px;
  color: #c8cbd1;
  background: transparent;
  --el-tree-node-hover-bg-color: #343537;
  --el-fill-color-light: #343537
}

.asset-tree :deep(.el-tree-node__content) {
  justify-content: flex-start;
  height: 34px;
  border-radius: 5px
}

.asset-tree :deep(.el-tree-node__content:has(.asset-tree-leaf)) {
  padding-left: 36px !important
}

.asset-tree :deep(.el-tree-node__expand-icon) {
  display: none
}

.tree-node {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 7px;
  padding-right: 7px;
  user-select: none
}

.tree-node.is-host {
  justify-content: flex-start;
  padding-left: 4px;
  cursor: default
}

.tree-node.is-database {
  padding-left: 4px;
  cursor: default
}

.tree-node .el-icon {
  flex: none;
  color: #8f959f
}

.tree-node.is-group .group-folder-icon {
  color: #8f959f;
  cursor: pointer
}

.tree-node.is-host .el-icon {
  color: #8f959f
}

.tree-node .host-asset-icon,
.tree-node .group-folder-icon,
.tree-node .database-asset-icon {
  display: inline-flex;
  width: 14px;
  height: 14px;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  font-size: 14px
}

.database-type-badge {
  flex: none;
  max-width: 48px;
  padding: 1px 3px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(104 169 215 / 34%);
  border-radius: 3px;
  color: #9dc9e8;
  background: rgb(104 169 215 / 12%);
  font-size: 6px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap
}

.tree-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap
}

.tree-node.is-database .tree-label {
  flex: 0 1 auto
}

.tree-node.is-host .tree-label {
  flex: 0 1 auto
}

.tree-node small {
  color: #91959d
}

.tree-node i,
.session-tab>i {
  width: 6px;
  height: 6px;
  flex: none;
  border-radius: 50%;
  background: #73767d
}

.tree-node.is-host i {
  margin-left: 0
}

.tree-node i.online,
.session-tab>i.online {
  background: #61c957
}

.terminal-stage {
  min-width: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #1d1e20;
  border-radius: 8px
}

.session-tabs {
  position: relative;
  height: 42px;
  display: flex;
  align-items: flex-end;
  padding-left: 8px;
  background: #303135;
  box-sizing: border-box;
  overflow-x: auto
}

.session-tabs::after {
  position: absolute;
  right: 8px;
  bottom: 0;
  left: 8px;
  height: 2px;
  border-radius: 2px;
  background: #3b3c40;
  content: ''
}

.session-tab {
  height: 34px;
  min-width: 130px;
  max-width: 210px;
  padding: 0 12px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border: 0;
  border-right: 1px solid #3c3d41;
  border-radius: 6px;
  color: #bfc2c8;
  background: #292a2d;
  cursor: pointer;
  font-size: 11px
}

.session-tab>span {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left
}

.session-tab.active,
.session-tab.placeholder {
  color: #eceef2;
  background: #1d1e20
}

.session-tab.placeholder {
  cursor: default
}

.terminal-actions {
  height: 100%;
  margin-left: auto;
  padding: 0 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex: none
}

.terminal-action {
  width: 28px;
  height: 28px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px;
  color: #aeb2ba;
  background: transparent;
  cursor: pointer
}

.terminal-action:hover:not(:disabled) {
  color: #fff;
  background: #4c4d52
}

.terminal-action:disabled {
  color: #666970;
  cursor: not-allowed;
  opacity: .55
}

.session-tab.placeholder>span {
  width: 6px;
  height: 6px;
  flex: none;
  border-radius: 50%;
  background: #7f838a
}

.session-tab .close {
  margin-left: 3px;
  border-radius: 3px
}

.session-tab .close:hover {
  color: #fff;
  background: #4c4d52
}

.session-content {
  position: relative;
  min-height: 0;
  flex: 1;
  background: #1d1e20
}

.terminal-body {
  position: absolute;
  inset: 0;
  padding: 10px 12px;
  box-sizing: border-box;
  overflow: hidden
}

.terminal-body :deep(.xterm) {
  height: 100%
}

.terminal-empty {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 9px;
  color: #696c72
}

.terminal-empty .el-icon {
  margin-bottom: 5px;
  font-size: 54px
}

.terminal-empty strong {
  color: #85888e;
  font-size: 13px
}

.terminal-empty span {
  font-size: 11px
}

.sftp-overlay {
  min-height: 0;
  flex: 1;
  display: flex;
  overflow: hidden;
  background: #1d1e20;
  font-size: 10px
}

.sftp-panel {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: #c8cbd1;
  background: #1d1e20;
  font-size: 10px
}

.sftp-session-bar {
  height: 42px;
  display: flex;
  align-items: stretch;
  border-bottom: 1px solid #393a3e;
  background: #252628
}

.sftp-session-tab,
.sftp-session-add,
.sftp-session-tool {
  min-width: 0;
  padding: 0 20px;
  display: inline-flex;
  align-items: center;
  gap: 9px;
  border: 0;
  border-right: 1px solid #393a3e;
  color: #aeb2ba;
  background: transparent;
  cursor: pointer;
  font-size: 10px
}

.sftp-session-tab {
  min-width: 180px;
  justify-content: flex-start;
  border-top: 2px solid transparent
}

.sftp-session-tab.active {
  border-top-color: #55cbb5;
  color: #eceef2;
  background: #1d1e20
}

.sftp-session-tab i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #9ca6b5
}

.sftp-session-tab i.online {
  background: #42c77b
}

.sftp-session-add {
  width: 54px;
  justify-content: center;
  padding: 0;
  font-size: 10px
}

.sftp-session-spacer {
  flex: 1
}

.sftp-session-tool {
  width: 54px;
  justify-content: center;
  padding: 0;
  border-left: 1px solid #393a3e;
  border-right: 0;
  font-size: 10px
}

.sftp-session-tool.active {
  color: #55cbb5;
  background: #303135
}

.sftp-toolbar {
  min-width: 0;
  min-height: 46px;
  padding: 5px 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  border-bottom: 1px solid #252a34;
  box-sizing: border-box;
  background: #1d1e20
}

.sftp-toolbar strong {
  flex: none;
  color: #55cbb5;
  font-size: 10px;
  font-weight: 700
}

.sftp-path,
.sftp-search,
.sftp-button,
.sftp-icon-button {
  height: 32px;
  border: 1px solid #45464a;
  border-radius: 9px;
  box-sizing: border-box;
  background: #1d1e20;
  color: #d7dae0
}

.sftp-path {
  min-width: 80px;
  flex: 1 1 auto;
  padding: 0 8px;
  display: flex;
  align-items: center;
  overflow: hidden;
  outline: none;
  color: #d7dae0;
  font: inherit;
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap
}

.sftp-icon-button {
  width: 32px;
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer
}

.sftp-icon-button .el-icon {
  font-size: 10px
}

.sftp-close {
  margin-left: auto
}

.sftp-search-wrap {
  position: relative;
  flex: none;
  display: flex;
  align-items: center;
  height: 32px
}

.sftp-search-popup {
  position: absolute;
  top: 42px;
  right: 0;
  z-index: 3;
  width: 190px;
  height: 32px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 6px;
  border: 1px solid #45464a;
  border-radius: 7px;
  box-sizing: border-box;
  background: #252628;
  box-shadow: 0 8px 18px rgb(0 0 0 / 28%)
}

.sftp-search-popup .el-icon {
  flex: none;
  color: #8c9098;
  font-size: 10px
}

.sftp-search-popup input {
  width: 100%;
  border: 0;
  outline: 0;
  color: #d7dae0;
  background: transparent;
  font: inherit;
  font-size: 10px
}

.sftp-search-popup input::placeholder {
  color: #8a95a6
}

.sftp-search {
  min-width: 150px;
  flex: 1;
  padding: 0 16px;
  display: flex;
  align-items: center;
  gap: 6px
}

.sftp-search .el-icon {
  flex: none;
  font-size: 10px
}

.sftp-search input {
  width: 100%;
  border: 0;
  outline: 0;
  color: #d7dae0;
  background: transparent;
  font: inherit;
  font-size: 10px
}

.sftp-search input::placeholder {
  color: #8a95a6
}

.sftp-button {
  padding: 0 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  flex: none;
  color: #c8cbd1;
  cursor: pointer;
  font-size: 10px
}

.sftp-button .el-icon {
  font-size: 10px
}

.sftp-split-wrap {
  position: relative;
  flex: none
}

.sftp-split-button {
  display: inline-flex;
  align-items: stretch
}

.sftp-split-button .sftp-button {
  background: #1d1e20
}

.sftp-split-button .sftp-split-main {
  border-radius: 9px 0 0 9px
}

.sftp-split-button .sftp-split-toggle {
  width: 28px;
  padding: 0;
  margin-left: -1px;
  border-radius: 0 9px 9px 0
}

.sftp-split-button .sftp-split-toggle .el-icon {
  font-size: 9px
}

.sftp-record-panel {
  position: absolute;
  top: 38px;
  right: 0;
  z-index: 5;
  width: 300px;
  overflow: hidden;
  border: 1px solid #45464a;
  border-radius: 8px;
  color: #c8cbd1;
  background: #252628;
  box-shadow: 0 10px 24px rgb(0 0 0 / 32%)
}

.sftp-record-panel.is-download {
  right: 0
}

.sftp-record-title {
  height: 34px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #3a3b3e
}

.sftp-record-title strong {
  color: #f2f4f7;
  font-size: 11px
}

.sftp-record-title button {
  border: 0;
  color: #8a95a6;
  background: transparent;
  cursor: pointer;
  font: inherit;
  font-size: 9px
}

.sftp-record-title button:hover {
  color: #fff
}

.sftp-record-empty {
  padding: 28px 12px;
  color: #8a95a6;
  text-align: center;
  font-size: 10px
}

.sftp-record-list {
  max-height: 260px;
  overflow-y: auto
}

.sftp-record-item {
  padding: 8px 10px;
  display: grid;
  gap: 4px;
  border-bottom: 1px solid #343537;
  font-size: 9px;
  animation: sftp-record-enter .22s ease-out both
}

.sftp-record-item:last-child {
  border-bottom: 0
}

.sftp-record-item>div {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px
}

.sftp-record-item strong {
  overflow: hidden;
  color: #e4e7ec;
  text-overflow: ellipsis;
  white-space: nowrap
}

.sftp-record-item span {
  overflow: hidden;
  color: #8a95a6;
  text-overflow: ellipsis;
  white-space: nowrap
}

.sftp-record-item em {
  flex: none;
  color: #e6a23c;
  font-style: normal
}

.sftp-record-item em.is-success {
  color: #67c23a
}

.sftp-record-item em.is-failed {
  color: #f56c6c
}

.sftp-record-item>.sftp-record-progress {
  width: 100%;
  height: 3px;
  display: block;
  overflow: hidden;
  border-radius: 2px;
  background: #3a3b3e
}

.sftp-record-progress i {
  height: 100%;
  display: block;
  border-radius: inherit;
  background: #55cbb5;
  transition: width .18s ease
}

.sftp-record-progress.is-failed i {
  background: #f56c6c
}

.sftp-record-progress.is-indeterminate i {
  animation: sftp-progress-indeterminate 1s ease-in-out infinite
}

@keyframes sftp-record-enter {
  from {
    opacity: 0;
    transform: translateY(-5px)
  }

  to {
    opacity: 1;
    transform: translateY(0)
  }
}

@keyframes sftp-progress-indeterminate {
  from {
    transform: translateX(-100%)
  }

  to {
    transform: translateX(285%)
  }
}

.sftp-upload-menu {
  position: absolute;
  top: 38px;
  left: 0;
  z-index: 4;
  min-width: 90px;
  padding: 3px;
  border: 1px solid #45464a;
  border-radius: 7px;
  background: #252628;
  box-shadow: 0 8px 18px rgb(0 0 0 / 28%)
}

.sftp-upload-menu button {
  width: 100%;
  padding: 6px 8px;
  display: flex;
  align-items: center;
  gap: 5px;
  border: 0;
  color: #c8cbd1;
  background: transparent;
  cursor: pointer;
  font: inherit;
  font-size: 10px;
  white-space: nowrap
}

.sftp-upload-menu button:hover {
  color: #fff;
  background: #343537
}

.sftp-hidden-input {
  display: none
}

.sftp-icon-button:hover,
.sftp-button:hover:not(:disabled) {
  border-color: #55cbb5;
  background: #252628
}

.sftp-button:disabled {
  color: #777b83;
  border-color: #1b2028;
  cursor: not-allowed
}

.sftp-table-head,
.sftp-row {
  width: max(100%, calc(34px + 50px + var(--sftp-size-width) + var(--sftp-modified-width) + var(--sftp-permissions-width) + 52px));
  display: grid;
  grid-template-columns: 34px minmax(50px, 1fr) var(--sftp-size-width) var(--sftp-modified-width) var(--sftp-permissions-width);
  align-items: center;
  column-gap: 8px;
  box-sizing: border-box
}

.sftp-table-scroll {
  min-height: 0;
  flex: 1;
  overflow: auto
}

.sftp-table-scroll.is-resizing {
  cursor: col-resize;
  user-select: none
}

.sftp-table-head {
  position: sticky;
  top: 0;
  z-index: 3;
  min-height: 38px;
  padding: 0 10px;
  color: #8c9098;
  background: #252628;
  border-bottom: 1px solid #393a3e;
  font-size: 10px
}

.sftp-list {
  min-height: 0
}

.sftp-row {
  min-height: 44px;
  padding: 0 10px;
  border: 0;
  border-bottom: 1px solid #393a3e;
  color: #c8cbd1;
  background: #1d1e20;
  text-align: left;
  cursor: pointer;
  font: inherit;
  font-size: 10px
}

.sftp-row:hover,
.sftp-row.selected {
  background: #343537
}

.sftp-check {
  position: sticky;
  left: 0;
  z-index: 2;
  width: 12px;
  height: 12px;
  justify-self: start;
  border: 1px solid #73767d;
  border-radius: 3px;
  background: #1d1e20;
  box-sizing: border-box
}

.sftp-check-spacer {
  position: sticky;
  left: 0;
  z-index: 4;
  width: 12px;
  height: 1px
}

.sftp-check.checked {
  border-color: #55cbb5;
  background: #55cbb5
}

.sftp-name {
  position: sticky;
  left: 42px;
  z-index: 2;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 14px;
  overflow: hidden;
  color: #d7dae0;
  background: #1d1e20;
  text-overflow: ellipsis;
  white-space: nowrap
}

.sftp-row:hover .sftp-check:not(.checked),
.sftp-row.selected .sftp-check:not(.checked),
.sftp-row:hover .sftp-name,
.sftp-row.selected .sftp-name {
  background: #343537
}

.sftp-name-head {
  position: sticky;
  left: 42px;
  z-index: 4;
  align-self: stretch;
  display: flex;
  align-items: center;
  background: #252628
}

.sftp-column-head {
  position: relative;
  align-self: stretch;
  display: flex;
  align-items: center;
  min-width: 0;
  white-space: nowrap
}

.sftp-resize-handle {
  position: absolute;
  top: 0;
  right: -8px;
  z-index: 5;
  width: 16px;
  height: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: col-resize;
  touch-action: none
}

.sftp-resize-handle::after {
  position: absolute;
  top: 25%;
  left: 7px;
  width: 1px;
  height: 50%;
  background: #56585e;
  content: ''
}

.sftp-resize-handle:hover::after {
  width: 2px;
  background: #55cbb5
}

.sftp-name .el-icon {
  flex: none;
  color: #55cbb5;
  font-size: 10px
}

.sftp-empty {
  padding-top: 50px;
  color: #8a95a6;
  text-align: center;
  font-size: 10px
}

@media(max-width:900px) {
  .terminal-header {
    height: 44px
  }

  .terminal-workbench {
    height: calc(100vh - 44px)
  }

  .sftp-toolbar {
    min-height: 58px;
    gap: 6px
  }

  .sftp-toolbar strong {
    font-size: 10px
  }

  .sftp-path,
  .sftp-search,
  .sftp-button,
  .sftp-icon-button {
    height: 44px;
    border-radius: 8px
  }

  .sftp-path {
    min-width: 110px;
    padding: 0 10px;
    font-size: 10px
  }

  .sftp-icon-button {
    width: 44px
  }

  .sftp-search {
    min-width: 120px;
    padding: 0 8px
  }

  .sftp-search input,
  .sftp-button {
    font-size: 10px
  }

  .sftp-button {
    padding: 0 10px
  }

  .sftp-table-head,
  .sftp-row {
    font-size: 10px
  }
}

.asset-search :deep(.el-input__wrapper) {
  min-height: 32px;
  border-radius: 7px !important;
  background: #343537 !important;
  border: 1px solid #3b3c40 !important;
  box-shadow: none !important
}

.asset-search :deep(.el-input__wrapper.is-focus),
.asset-search :deep(.el-input__wrapper:focus-within) {
  border-color: #55cbb5 !important;
  box-shadow: 0 0 0 1px rgb(85 203 181 / 18%) !important
}

.asset-search :deep(.el-input__inner),
.asset-search :deep(.el-input__inner::placeholder),
.asset-search :deep(.el-input__prefix),
.asset-search :deep(.el-input__suffix) {
  color: #c8cbd1 !important;
  -webkit-text-fill-color: #c8cbd1 !important
}
</style>
