<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage, type ElTree } from 'element-plus'
import { ArrowLeft, ArrowRight, Close, CollectionTag, Monitor, Search } from '@element-plus/icons-vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import gopsLogoUrl from '@/assets/gops.svg'
import { listAssetGroups, listHosts, type AssetGroup, type Host } from '@/api/asset'
import { useSystemConfigStore } from '@/stores/system-config'

interface TreeNode { id: string; label: string; type: 'group' | 'host'; count?: number; host?: Host; children?: TreeNode[] }
interface Session { key: number; host: Host }
interface Runtime { terminal: Terminal; fitAddon: FitAddon; socket: WebSocket; resizeObserver: ResizeObserver }

const systemConfig = useSystemConfigStore()
const hosts = ref<Host[]>([])
const groups = ref<AssetGroup[]>([])
const keyword = ref('')
const treeRef = ref<InstanceType<typeof ElTree>>()
const sessions = ref<Session[]>([])
const activeKey = ref<number | null>(null)
const assetPaneWidth = ref(180)
const assetPaneDragging = ref(false)
const assetPaneStartX = ref(0)
const assetPaneStartWidth = ref(180)
const terminalElements = new Map<number, HTMLElement>()
const runtimes = new Map<number, Runtime>()

const treeData = computed<TreeNode[]>(() => {
  const nodes = new Map<number, TreeNode>()
  const roots: TreeNode[] = []
  groups.value.forEach(group => nodes.set(group.id, { id: `group-${group.id}`, label: group.name, type: 'group', count: 0, children: [] }))
  groups.value.forEach(group => {
    const node = nodes.get(group.id)
    if (!node) return
    const parent = nodes.get(group.parent_id)
    if (parent) parent.children?.push(node)
    else roots.push(node)
  })
  const ungrouped: TreeNode = { id: 'group-0', label: '未分组', type: 'group', count: 0, children: [] }
  hosts.value.forEach(host => {
    const parent = nodes.get(host.group_id) || ungrouped
    parent.children?.push({ id: `host-${host.id}`, label: host.name, type: 'host', host })
    parent.count = (parent.count || 0) + 1
  })
  const sum = (node: TreeNode): number => {
    const children = (node.children || []).filter(child => child.type === 'group').reduce((total, child) => total + sum(child), 0)
    node.count = (node.count || 0) + children
    return node.count
  }
  roots.forEach(sum)
  if (ungrouped.children?.length) roots.push(ungrouped)
  return roots
})

function setTerminalElement(key: number, element: unknown) {
  if (element instanceof HTMLElement) terminalElements.set(key, element)
  else terminalElements.delete(key)
}
function filterTree(value: string, data: TreeNode) {
  if (!value.trim()) return true
  return `${data.label} ${data.host?.ip || ''}`.toLowerCase().includes(value.trim().toLowerCase())
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
function socketURL(host: Host, terminal: Terminal) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token') || ''
  return `${protocol}//${window.location.host}/api/v1/assets/terminal/connect/${host.id}?token=${encodeURIComponent(token)}&cols=${terminal.cols}&rows=${terminal.rows}`
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
  terminal.writeln(`正在连接 ${session.host.name}（${session.host.ip}:${session.host.port}）...`)
  const socket = new WebSocket(socketURL(session.host, terminal))
  socket.binaryType = 'arraybuffer'
  const resizeObserver = new ResizeObserver(() => { fitAddon.fit(); const runtime = runtimes.get(session.key); if (runtime) sendResize(runtime) })
  const runtime = { terminal, fitAddon, socket, resizeObserver }
  runtimes.set(session.key, runtime)
  terminal.onData(data => socket.readyState === WebSocket.OPEN && socket.send(data))
  socket.onopen = () => { fitAddon.fit(); sendResize(runtime); terminal.focus() }
  socket.onmessage = event => terminal.write(event.data instanceof ArrayBuffer ? new Uint8Array(event.data) : String(event.data))
  socket.onerror = () => terminal.writeln('\r\n连接失败，请检查主机、凭证和网络。')
  socket.onclose = () => terminal.writeln('\r\n连接已关闭。')
  resizeObserver.observe(element)
}
async function openTerminal(host?: Host) {
  if (!host) return
  const existing = sessions.value.find(session => session.key === host.id)
  if (existing) { activateSession(existing.key); return }
  const session = { key: host.id, host }
  sessions.value.push(session)
  activeKey.value = session.key
  await initializeSession(session)
}
function activateSession(key: number) {
  activeKey.value = key
  void nextTick(() => { runtimes.get(key)?.fitAddon.fit(); runtimes.get(key)?.terminal.focus() })
}
function disposeSession(key: number) {
  const runtime = runtimes.get(key)
  runtime?.resizeObserver.disconnect()
  runtime?.socket.close()
  runtime?.terminal.dispose()
  runtimes.delete(key)
  terminalElements.delete(key)
}
function closeSession(key: number) {
  const index = sessions.value.findIndex(session => session.key === key)
  if (index < 0) return
  disposeSession(key)
  sessions.value.splice(index, 1)
  if (activeKey.value === key) activeKey.value = sessions.value[Math.min(index, sessions.value.length - 1)]?.key ?? null
}
async function load() {
  const [hostResult, groupResult] = await Promise.all([listHosts('', 0, 1, 1000), listAssetGroups('', 1, 1000)])
  hosts.value = hostResult.data.data.list
  groups.value = groupResult.data.data.list
}
watch(keyword, value => treeRef.value?.filter(value))
onMounted(() => { void systemConfig.fetchConfig(); void load().catch(() => ElMessage.error('读取资产失败')) })
onBeforeUnmount(() => {
  sessions.value.forEach(session => disposeSession(session.key))
  document.removeEventListener('pointermove', moveAssetPane)
  document.removeEventListener('pointerup', stopAssetPaneDrag)
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
            </el-icon>资产分组</span><small>({{ hosts.length }})</small></div>
        <el-input v-model="keyword" class="asset-search" placeholder="搜索主机..." clearable :prefix-icon="Search" />
        <el-tree ref="treeRef" class="asset-tree" :data="treeData" node-key="id" default-expand-all
          :filter-node-method="filterTree" :expand-on-click-node="true">
          <template #default="{ data }"><span class="tree-node" :class="`is-${data.type}`"
              @dblclick.stop="openTerminal(data.host)"><el-icon v-if="data.type === 'group'">
                <ArrowRight />
              </el-icon><el-icon v-else>
                <Monitor />
              </el-icon><span class="tree-label">{{ data.label }}</span><small v-if="data.type === 'group'">({{
                data.count }})</small><i v-else :class="{ online: data.host?.status === 1 }" /></span></template>
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
              <Monitor />
            </el-icon><span>{{ session.host.name }}</span><i :class="{ online: session.host.status === 1 }" /><el-icon
              class="close" @click.stop="closeSession(session.key)">
              <Close />
            </el-icon></button>
        </nav>
        <div class="session-content">
          <div v-if="!sessions.length" class="terminal-empty"><el-icon>
              <Monitor />
            </el-icon><strong>双击主机打开终端</strong><span>在左侧资产分组中选择主机</span></div>
          <div v-for="session in sessions" v-show="activeKey === session.key" :key="session.key"
            :ref="element => setTerminalElement(session.key, element)" class="terminal-body" />
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
  width: 28px;
  height: 28px;
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
  background: #252628
}

.asset-title {
  height: 42px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  font-weight: 700;
  border-bottom: 1px solid #393a3e
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
  border: 1px solid #45464a;
  box-shadow: none !important;
  border-radius: 6px
}

.asset-search :deep(.el-input__inner) {
  color: #e1e3e8
}

.asset-search :deep(.el-input__inner::placeholder) {
  color: #92969f
}

.asset-divider {
  position: relative;
  z-index: 4;
  background: #3b3c40;
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
  height: 34px;
  border-radius: 5px
}

.asset-tree :deep(.el-tree-node__expand-icon) {
  display: none
}

.tree-node {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 7px;
  padding-right: 7px;
  user-select: none
}

.tree-node.is-host {
  padding-left: 4px;
  cursor: default
}

.tree-node .el-icon {
  flex: none;
  color: #8f959f
}

.tree-node.is-host .el-icon {
  color: #61c957
}

.tree-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap
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

.tree-node i.online,
.session-tab>i.online {
  background: #61c957
}

.terminal-stage {
  min-width: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #1d1e20
}

.session-tabs {
  height: 42px;
  display: flex;
  align-items: flex-end;
  padding-left: 8px;
  background: #303135;
  border-bottom: 1px solid #3b3c40;
  box-sizing: border-box;
  overflow-x: auto
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
  border-radius: 5px 5px 0 0;
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

@media(max-width:900px) {
  .terminal-header {
    height: 44px
  }

  .terminal-workbench {
    height: calc(100vh - 44px)
  }
}

.asset-search :deep(.el-input__wrapper) {
  min-height: 32px;
  border-radius: 7px !important
}
</style>
