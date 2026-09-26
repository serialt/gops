<script setup lang="ts">
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, render, type Component } from 'vue'
import { ElMessage, ElMessageBox, ElTooltip } from 'element-plus'
import { Delete, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteTerminalSession, getTerminalSessionRecording, listTerminalSessions, type TerminalSession } from '@/api/asset'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

interface CastHeader { version: number; width: number; height: number; timestamp?: number }
interface CastEvent { time: number; type: string; data: string }

const rows = ref<TerminalSession[]>([])
const keyword = ref('')
const sessionType = ref('')
const recording = ref('')
const recordingVisible = ref(false)
const castHeader = ref<CastHeader | null>(null)
const castEvents = ref<CastEvent[]>([])
const playbackElement = ref<HTMLElement>()
const playbackDuration = computed(() => castEvents.value.at(-1)?.time || 0)
const playbackTime = ref(0)
const playbackSpeed = ref(1)
const playbackPlaying = ref(false)
let playbackTerminal: Terminal | undefined
let playbackFrame = 0
let playbackEventIndex = 0
let playbackLastFrame = 0
function formatDuration(seconds: number) {
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return remainder ? `${minutes}m ${remainder}s` : `${minutes}m 0s`
}
function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  const kilobytes = bytes / 1024
  return kilobytes < 1024 ? `${kilobytes.toFixed(kilobytes < 10 ? 1 : 0)} KB` : `${(kilobytes / 1024).toFixed(1)} MB`
}
function iconButton(title: string, kind: string, icon: Component, click: () => void) {
  const root = document.createElement('span')
  render(h(ElTooltip, { content: title, placement: 'top', showAfter: 250 }, {
    default: () => h('button', { type: 'button', 'aria-label': title, class: `audit-icon-action ${kind}`, onClick: click }, [
      h('span', { class: 'el-icon' }, [h(icon)]),
    ]),
  }), root)
  return root
}
function assetCell(row: TerminalSession) {
  const cell = document.createElement('div')
  cell.className = 'audit-asset'
  const name = document.createElement('strong')
  const address = document.createElement('small')
  name.textContent = row.session_type === 'database' ? row.database_name || `数据库 #${row.db_instance_id}` : row.host_name || `主机 #${row.host_id}`
  address.textContent = row.session_type === 'database' ? row.database_address : row.host_address
  cell.append(name, address)
  return cell
}
const columns = computed<ColDef[]>(() => [
  { field: 'id', headerName: 'ID', width: 64 },
  { field: 'session_type', headerName: '会话类型', width: 112, cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = `audit-type ${p.value === 'database' ? 'database' : 'ssh'}`; tag.textContent = p.value === 'database' ? '数据库' : 'SSH'; return tag } },
  { headerName: '资产信息', minWidth: 170, cellRenderer: (p: any) => assetCell(p.data) },
  { field: 'username', headerName: '操作用户', minWidth: 112, valueFormatter: (p: any) => p.value || `用户 #${p.data.user_id}` },
  { field: 'duration_seconds', headerName: '时长', width: 92, valueFormatter: (p: any) => formatDuration(p.value || 0) },
  { field: 'recording_size', headerName: '录制大小', width: 100, valueFormatter: (p: any) => formatBytes(p.value || 0) },
  { field: 'command_count', headerName: '命令数', width: 86 },
  { field: 'completed', headerName: '状态', width: 94, cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = `audit-status${p.value ? '' : ' pending'}`; tag.textContent = p.value ? '已完成' : '进行中'; return tag } },
  { field: 'started_at', headerName: '创建时间', minWidth: 180, valueFormatter: (p: any) => p.value ? String(p.value).replace('T', ' ').replace(/\+.*$/, '').slice(0, 19) : '—' },
  { headerName: '操作', pinned: 'right', width: 96, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'audit-actions'; wrap.append(iconButton('查看录制', 'view', VideoPlay, () => viewRecording(p.data)), iconButton('删除会话', 'delete', Delete, () => remove(p.data))); return wrap } },
])
async function load() { const { data } = await listTerminalSessions(keyword.value, sessionType.value); rows.value = data.data }
async function viewRecording(row: TerminalSession) {
  stopPlayback()
  const { data } = await getTerminalSessionRecording(row.id)
  recording.value = data.data.recording || ''
  parseRecording(recording.value)
  recordingVisible.value = true
  if (castHeader.value) {
    await nextTick()
    setupPlaybackTerminal()
  }
}
function parseRecording(value: string) {
  castHeader.value = null
  castEvents.value = []
  const lines = value.split(/\r?\n/).filter(Boolean)
  if (!lines.length) return
  try {
    const header = JSON.parse(lines[0]) as CastHeader
    if (header.version !== 2 || !header.width || !header.height) return
    const events = lines.slice(1).flatMap(line => {
      try {
        const event = JSON.parse(line) as [number, string, string]
        return Array.isArray(event) && typeof event[0] === 'number' && typeof event[2] === 'string'
          ? [{ time: Math.max(0, event[0]), type: event[1], data: event[2] }]
          : []
      } catch { return [] }
    })
    castHeader.value = header
    castEvents.value = events
  } catch { /* Existing plain-text recordings use the fallback view. */ }
}
function setupPlaybackTerminal() {
  if (!playbackElement.value || !castHeader.value) return
  playbackTerminal?.dispose()
  playbackTerminal = new Terminal({
    cursorBlink: false,
    disableStdin: true,
    fontSize: 12,
    lineHeight: 1.2,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
    theme: { background: '#1d1e20', foreground: '#d7dae0', cursor: '#f5f5f5' },
  })
  playbackTerminal.open(playbackElement.value)
  playbackTerminal.resize(castHeader.value.width, castHeader.value.height)
  playbackTerminal.reset()
  playbackEventIndex = 0
  playbackTime.value = 0
}
function seekPlayback(value: number) {
  const wasPlaying = playbackPlaying.value
  playbackPlaying.value = false
  cancelAnimationFrame(playbackFrame)
  playbackTerminal?.reset()
  playbackEventIndex = 0
  playbackTime.value = value
  while (playbackEventIndex < castEvents.value.length && castEvents.value[playbackEventIndex].time <= value) {
    const event = castEvents.value[playbackEventIndex++]
    if (event.type === 'o') playbackTerminal?.write(event.data)
  }
  if (wasPlaying && value < playbackDuration.value) startPlayback()
}
function startPlayback() {
  if (!playbackTerminal || !castEvents.value.length) return
  if (playbackTime.value >= playbackDuration.value) seekPlayback(0)
  playbackPlaying.value = true
  playbackLastFrame = performance.now()
  playbackFrame = requestAnimationFrame(playbackTick)
}
function togglePlayback() {
  if (playbackPlaying.value) {
    playbackPlaying.value = false
    cancelAnimationFrame(playbackFrame)
  } else startPlayback()
}
function playbackTick(now: number) {
  if (!playbackPlaying.value) return
  playbackTime.value = Math.min(playbackDuration.value, playbackTime.value + (now - playbackLastFrame) / 1000 * playbackSpeed.value)
  playbackLastFrame = now
  while (playbackEventIndex < castEvents.value.length && castEvents.value[playbackEventIndex].time <= playbackTime.value) {
    const event = castEvents.value[playbackEventIndex++]
    if (event.type === 'o') playbackTerminal?.write(event.data)
  }
  if (playbackTime.value >= playbackDuration.value) {
    playbackPlaying.value = false
    return
  }
  playbackFrame = requestAnimationFrame(playbackTick)
}
function formatPlaybackTime(seconds: number) {
  const safeSeconds = Math.floor(seconds)
  return `${Math.floor(safeSeconds / 60)}:${String(safeSeconds % 60).padStart(2, '0')}`
}
function stopPlayback() {
  playbackPlaying.value = false
  cancelAnimationFrame(playbackFrame)
  playbackTerminal?.dispose()
  playbackTerminal = undefined
}
function closePlayback() {
  stopPlayback()
  castHeader.value = null
  castEvents.value = []
  recording.value = ''
}
async function remove(row: TerminalSession) { await ElMessageBox.confirm('确定删除该审计会话？', '删除确认', { type: 'warning' }); await deleteTerminalSession(row.id); ElMessage.success('删除成功'); await load() }
async function batchRemove(selectedRows: unknown[]) {
  const selectedSessions = selectedRows as TerminalSession[]
  if (!selectedSessions.length) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedSessions.length} 条终端会话？`, '批量删除确认', { type: 'warning' })
  await Promise.all(selectedSessions.map(session => deleteTerminalSession(session.id)))
  ElMessage.success('批量删除成功')
  await load()
}
onMounted(load)
onBeforeUnmount(stopPlayback)
</script>
<template>
  <div class="page-container table-shell">
    <div class="section-card">
      <div class="page-toolbar">
        <h2 class="page-title">终端录制</h2>
      </div>
      <div class="filters"><el-input v-model="keyword" placeholder="搜索用户、主机或数据库名称/IP..." clearable
          @input="load" /><el-select v-model="sessionType" placeholder="会话类型" clearable @change="load"><el-option
            label="SSH" value="ssh" /><el-option label="数据库" value="database" /></el-select></div>
      <AppDataGrid :row-data="rows" :column-defs="columns" pagination :bulk-delete="batchRemove" />
    </div><el-dialog v-model="recordingVisible" class="recording-dialog" title="会话录制" width="80%" align-center
      @closed="closePlayback">
      <div v-if="castHeader" class="recording-player">
        <div ref="playbackElement" class="recording-terminal" />
        <div class="playback-controls">
          <el-button circle size="small" :aria-label="playbackPlaying ? '暂停' : '播放'" @click="togglePlayback">
            <el-icon>
              <VideoPause v-if="playbackPlaying" />
              <VideoPlay v-else />
            </el-icon>
          </el-button>
          <span class="playback-time">{{ formatPlaybackTime(playbackTime) }}</span>
          <input class="playback-seek" type="range" min="0" :max="playbackDuration || 1" step="0.05"
            :value="playbackTime" @input="seekPlayback(Number(($event.target as HTMLInputElement).value))" />
          <span class="playback-time">{{ formatPlaybackTime(playbackDuration) }}</span>
          <el-select v-model="playbackSpeed" class="playback-speed" size="small" aria-label="播放速度">
            <el-option v-for="speed in [0.5, 1, 1.5, 2]" :key="speed" :label="`${speed}x`" :value="speed" />
          </el-select>
        </div>
      </div>
      <pre v-else class="recording recording-plain">{{ recording || '暂无录制内容' }}</pre>
    </el-dialog>
  </div>
</template>
<style scoped>
.table-shell {
  background: var(--bg);
  min-height: 100%;
}

.section-card {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding: 12px 16px 16px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-card);
  box-shadow: var(--shadow);
}

.page-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
  margin-bottom: 8px;
}

.filters {
  display: flex;
  gap: 8px;
  margin-bottom: 8px
}

.filters .el-input {
  width: 240px
}

.filters .el-select {
  width: 120px
}

.filters :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--line) inset;
}

:deep(.audit-actions) {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px
}

:deep(.audit-icon-action) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 7px;
  background: transparent;
  color: var(--text-soft);
  cursor: pointer
}

:deep(.audit-icon-action .el-icon),
:deep(.audit-icon-action svg) {
  width: 16px;
  height: 16px
}

:deep(.audit-icon-action:hover) {
  border-color: var(--line);
  background: #f6f7fa;
  color: var(--primary)
}

:deep(.audit-icon-action.delete:hover) {
  border-color: #ffd5d5;
  background: #fff7f7;
  color: #ef6262
}

:deep(.audit-type),
:deep(.audit-status) {
  display: inline-flex;
  align-items: center;
  padding: 3px 7px;
  border-radius: 4px;
  background: #edf5ff;
  color: #5081c8;
  font-size: 10px;
  line-height: 1.2
}

:deep(.audit-type.database) {
  background: #f1edff;
  color: #7959b8
}

:deep(.audit-status) {
  background: #edf8eb;
  color: #4e9b43
}

:deep(.audit-status.pending) {
  background: #fff5df;
  color: #b88220
}

:deep(.audit-asset) {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  min-width: 0;
  line-height: 1.25
}

:deep(.audit-asset strong) {
  overflow: hidden;
  color: var(--text);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap
}

:deep(.audit-asset small) {
  overflow: hidden;
  color: var(--text-muted);
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap
}

.recording-player {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 96px);
  gap: 6px
}

:deep(.recording-dialog.el-dialog) {
  max-height: calc(100vh - 24px);
  overflow: hidden;
  border-radius: 8px
}

:deep(.recording-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 4px 10px 5px
}

:deep(.recording-dialog .el-dialog__body) {
  max-height: calc(100vh - 82px);
  padding: 0 4px 4px;
  overflow: auto
}

.recording-terminal {
  height: min(62vh, 500px);
  min-height: 200px;
  padding: 4px;
  overflow: auto;
  border: 1px solid #30343a;
  border-radius: 6px;
  background: #1d1e20
}

.playback-controls {
  display: flex;
  align-items: center;
  gap: 10px
}

.playback-time {
  flex: none;
  color: var(--text-muted);
  font-size: 11px;
  font-variant-numeric: tabular-nums
}

.playback-seek {
  min-width: 0;
  flex: 1;
  accent-color: var(--primary)
}

.playback-speed {
  width: 82px;
  flex: none
}

.recording-plain {
  max-height: 480px;
  overflow: auto;
  white-space: pre-wrap;
  font-size: 11px
}
</style>
