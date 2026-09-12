<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteTerminalSession, getTerminalSessionRecording, listTerminalSessions, type TerminalSession } from '@/api/asset'

const rows = ref<TerminalSession[]>([])
const keyword = ref('')
const sessionType = ref('')
const recording = ref('')
const recordingVisible = ref(false)
function button(text: string, kind: string, click: () => void) { const el = document.createElement('button'); el.textContent = text; el.className = `audit-action ${kind}`; el.onclick = click; return el }
const columns = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: p => String((p.node?.rowIndex ?? 0) + 1), width: 70 },
  { field: 'session_type', headerName: '会话类型', width: 100 },
  { field: 'user_id', headerName: '用户', width: 90 },
  { field: 'host_id', headerName: '主机', width: 90 },
  { field: 'db_instance_id', headerName: '数据库', width: 90 },
  { field: 'started_at', headerName: '开始时间', minWidth: 150 },
  { field: 'duration_seconds', headerName: '时长(秒)', width: 100 },
  { field: 'command_count', headerName: '命令数', width: 90 },
  { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'audit-actions'; wrap.append(button('查看', 'edit', () => viewRecording(p.data)), button('删除', 'delete', () => remove(p.data))); return wrap } },
])
async function load() { const { data } = await listTerminalSessions(keyword.value, sessionType.value); rows.value = data.data }
async function viewRecording(row: TerminalSession) { const { data } = await getTerminalSessionRecording(row.id); recording.value = data.data.recording || '暂无录制内容'; recordingVisible.value = true }
async function remove(row: TerminalSession) { await ElMessageBox.confirm('确定删除该审计会话？', '删除确认', { type: 'warning' }); await deleteTerminalSession(row.id); ElMessage.success('删除成功'); await load() }
onMounted(load)
</script>
<template>
  <div class="page-container table-shell">
    <div class="section-card">
      <div class="page-toolbar">
        <h2 class="page-title">终端审计会话</h2>
      </div>
      <div class="filters"><el-input v-model="keyword" placeholder="搜索用户或主机 ID" clearable @input="load" /><el-select
          v-model="sessionType" placeholder="会话类型" clearable @change="load"><el-option label="SSH"
            value="ssh" /><el-option label="数据库" value="database" /></el-select></div>
      <AppDataGrid :row-data="rows" :column-defs="columns" pagination />
    </div><el-dialog v-model="recordingVisible" title="会话录制" width="680px">
      <pre class="recording">{{ recording }}</pre>
    </el-dialog>
  </div>
</template>
<style
  scoped>
  .filters {
    display: flex;
    gap: 8px;
    margin-bottom: 12px
  }

  .filters .el-input {
    width: 240px
  }

  .filters .el-select {
    width: 120px
  }

  :deep(.audit-actions) {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px
  }

  :deep(.audit-action) {
    padding: 4px 9px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: #fff;
    font-size: 11px;
    cursor: pointer;
    box-shadow: 0 1px 3px rgba(24, 34, 53, .08)
  }

  :deep(.audit-action.edit) {
    border-color: var(--primary);
    color: var(--primary);
    background: #f7f8ff
  }

  :deep(.audit-action.delete) {
    border-color: #ffd5d5;
    color: #ef6262;
    background: #fff7f7
  }

  .recording {
    max-height: 480px;
    overflow: auto;
    white-space: pre-wrap;
    font-size: 11px
  }
</style>
