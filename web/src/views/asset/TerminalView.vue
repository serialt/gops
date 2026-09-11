<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import { listHosts, type Host } from '@/api/asset'

const hosts = ref<Host[]>([])
const keyword = ref('')
const selectedHost = ref<Host | null>(null)
const terminalElement = ref<HTMLElement>()
let terminal: Terminal | undefined
let fitAddon: FitAddon | undefined
let socket: WebSocket | undefined
let resizeObserver: ResizeObserver | undefined

const filteredHosts = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  return value ? hosts.value.filter((host) => `${host.name} ${host.ip}`.toLowerCase().includes(value)) : hosts.value
})

async function loadHosts() {
  const { data } = await listHosts()
  hosts.value = data.data
}

function cleanup() {
  resizeObserver?.disconnect()
  resizeObserver = undefined
  socket?.close()
  socket = undefined
  terminal?.dispose()
  terminal = undefined
  fitAddon = undefined
}

function sendResize() {
  if (!socket || socket.readyState !== WebSocket.OPEN || !terminal) return
  socket.send(JSON.stringify({ type: 'resize', cols: terminal.cols, rows: terminal.rows }))
}

async function open(host: Host) {
  cleanup()
  selectedHost.value = host
  await nextTick()
  if (!terminalElement.value) return
  terminal = new Terminal({ cursorBlink: true, fontSize: 13, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace', theme: { background: '#101827', foreground: '#e5e7eb' } })
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(terminalElement.value)
  fitAddon.fit()
  terminal.writeln(`正在连接 ${host.name}（${host.ip}:${host.port}）...`)
  terminal.onData((data) => socket?.readyState === WebSocket.OPEN && socket.send(data))
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token') || ''
  socket = new WebSocket(`${protocol}//${window.location.host}/api/v1/assets/terminal/connect/${host.id}?token=${encodeURIComponent(token)}&cols=${terminal.cols}&rows=${terminal.rows}`)
  socket.binaryType = 'arraybuffer'
  socket.onopen = () => { fitAddon?.fit(); sendResize(); terminal?.focus() }
  socket.onmessage = (event) => {
    if (event.data instanceof ArrayBuffer) terminal?.write(new Uint8Array(event.data))
    else terminal?.write(String(event.data))
  }
  socket.onerror = () => terminal?.writeln('\r\n连接失败，请检查主机、凭证和网络。')
  socket.onclose = () => terminal?.writeln('\r\n连接已关闭。')
  resizeObserver = new ResizeObserver(() => { fitAddon?.fit(); sendResize() })
  resizeObserver.observe(terminalElement.value)
}

onMounted(() => loadHosts().catch(() => ElMessage.error('读取主机失败')))
onBeforeUnmount(cleanup)
</script>

<template>
  <div class="terminal-page">
    <aside class="host-panel">
      <div class="host-panel-title">主机终端</div>
      <el-input v-model="keyword" placeholder="搜索主机" clearable />
      <div class="host-list">
        <button v-for="host in filteredHosts" :key="host.id" class="host-item" :class="{ active: selectedHost?.id === host.id }" @click="open(host)">
          <strong>{{ host.name }}</strong><span>{{ host.ip }}:{{ host.port }}</span>
        </button>
        <el-empty v-if="!filteredHosts.length" description="暂无主机" :image-size="48" />
      </div>
    </aside>
    <section class="terminal-panel">
      <header>{{ selectedHost ? `${selectedHost.name} · ${selectedHost.ssh_user || 'SSH'}` : '选择左侧主机以打开终端' }}</header>
      <div ref="terminalElement" class="terminal-body" />
    </section>
  </div>
</template>

<style scoped>
.terminal-page{display:flex;height:100%;min-height:520px;gap:10px;padding:10px;box-sizing:border-box}.host-panel,.terminal-panel{border:1px solid var(--line);border-radius:10px;background:#fff;min-width:0}.host-panel{width:220px;padding:12px;display:flex;flex-direction:column;gap:10px}.host-panel-title{font-size:15px;font-weight:700}.host-list{overflow:auto;display:flex;flex-direction:column;gap:4px}.host-item{border:0;border-radius:7px;background:transparent;padding:8px;text-align:left;cursor:pointer;color:var(--text);font-size:11px}.host-item:hover,.host-item.active{background:#f1f3ff}.host-item strong,.host-item span{display:block}.host-item span{margin-top:3px;color:#7a8498;font-size:10px}.terminal-panel{flex:1;display:flex;flex-direction:column;overflow:hidden;background:#101827}.terminal-panel header{padding:9px 12px;background:#fff;color:var(--text);font-size:11px}.terminal-body{flex:1;min-height:0;padding:10px;overflow:hidden}@media(max-width:720px){.terminal-page{flex-direction:column}.host-panel{width:auto;max-height:180px}.host-list{display:grid;grid-template-columns:repeat(2,minmax(0,1fr))}}
</style>
