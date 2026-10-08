<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { listTaskJobs, type TaskHostResult, type TaskJob } from '@/api/task'

const rows = ref<TaskJob[]>([]), keyword = ref(''), taskType = ref(''), status = ref('')
const page = ref(1), pageSize = ref(20), total = ref(0), loading = ref(false)
const detail = ref<TaskJob | null>(null)
async function load() { loading.value = true; try { const { data } = await listTaskJobs(keyword.value, taskType.value, status.value, page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total } finally { loading.value = false } }
function label(value: string) { return ({ manual: '执行任务', file: '文件分发', pending: '等待中', running: '执行中', success: '成功', failed: '失败' } as Record<string,string>)[value] || value }
function parameters(row: TaskJob | null) { try { return JSON.parse(row?.parameters || '{}') as Record<string, unknown> } catch { return {} } }
function executionResults(row: TaskJob | null) { try { const value = JSON.parse(row?.result || '[]'); return Array.isArray(value) ? value as TaskHostResult[] : [] } catch { return [] } }
function formatTime(value: string) { return value ? value.replace('T', ' ').slice(0, 19) : '-' }
function tagRenderer(value: string) { const tag = document.createElement('span'); tag.className = `job-tag ${value}`; tag.textContent = label(value); return tag }
function actionRenderer(params: { data: TaskJob }) { const wrap = document.createElement('div'); wrap.className = 'job-actions'; const button = document.createElement('button'); button.className = 'view-button'; button.textContent = '查看'; button.onclick = () => { detail.value = params.data }; wrap.append(button); return wrap }
const columns = computed<ColDef[]>(() => [
  { headerName: '序号', width: 64, valueGetter: params => (page.value - 1) * pageSize.value + (params.node?.rowIndex || 0) + 1 },
  { field: 'name', headerName: '任务名称', minWidth: 190 },
  { field: 'task_type', headerName: '任务类型', width: 110, cellRenderer: (params: { value: string }) => tagRenderer(params.value) },
  { field: 'status', headerName: '状态', width: 90, cellRenderer: (params: { value: string }) => tagRenderer(params.value) },
  { field: 'target_hosts_display', headerName: '目标主机', minWidth: 210, valueFormatter: params => params.value || '-' },
  { field: 'error_message', headerName: '错误信息', minWidth: 150 },
  { headerName: '操作', pinned: 'right', width: 150, cellStyle: { display: 'flex', alignItems: 'center' }, cellRenderer: actionRenderer },
])
watch([keyword, taskType, status], () => { page.value = 1; void load() })
onMounted(load)
</script>
<template>
  <section class="page"><header><h2>任务作业</h2><p>查看执行任务和文件分发的运行记录。</p></header><div class="toolbar"><el-input v-model="keyword" clearable placeholder="搜索任务名称" /><el-select v-model="taskType" clearable placeholder="选择任务类型"><el-option label="执行任务" value="manual" /><el-option label="文件分发" value="file" /></el-select><el-select v-model="status" clearable placeholder="选择状态"><el-option label="执行中" value="running" /><el-option label="成功" value="success" /><el-option label="失败" value="failed" /></el-select></div><AppDataGrid v-loading="loading" class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total" :current-page="page" :page-size="pageSize" @page-change="value => { page = value; void load() }" @page-size-change="value => { pageSize = value; page = 1; void load() }" />
    <el-dialog :model-value="Boolean(detail)" title="任务详情" width="820px" class="task-detail-dialog" @update:model-value="(value: boolean) => { if (!value) detail = null }">
      <div v-if="detail" class="detail-content">
        <div class="detail-summary">
          <div><span>任务名称</span><strong>{{ detail.name }}</strong></div>
          <div><span>任务类型</span><i class="job-tag" :class="detail.task_type">{{ label(detail.task_type) }}</i></div>
          <div><span>执行状态</span><i class="job-tag" :class="detail.status">{{ label(detail.status) }}</i></div>
          <div><span>创建时间</span><strong>{{ formatTime(detail.created_at) }}</strong></div>
        </div>
        <section class="detail-section"><h3>目标主机</h3><p>{{ detail.target_hosts_display || '-' }}</p></section>
        <section class="detail-section"><h3>执行参数</h3><div class="parameter-list"><span v-if="parameters(detail).script_type">脚本类型：<b>{{ parameters(detail).script_type === 'python' ? 'Python' : 'Shell' }}</b></span><span v-if="parameters(detail).timeout">超时时间：<b>{{ parameters(detail).timeout }} 秒</b></span><span v-if="parameters(detail).target_path">目标路径：<b>{{ parameters(detail).target_path }}</b></span><span v-if="Array.isArray(parameters(detail).files)">文件：<b>{{ (parameters(detail).files as string[]).join('、') }}</b></span><span v-if="!Object.keys(parameters(detail)).length">-</span></div></section>
        <section class="detail-section result-section"><h3>执行结果 <small>共 {{ executionResults(detail).length }} 台主机</small></h3><div v-if="executionResults(detail).length" class="detail-results"><article v-for="item in executionResults(detail)" :key="item.host_id" class="host-result"><div class="host-result-head"><strong>{{ item.host_name }}（{{ item.host_ip }}）</strong><i class="job-tag" :class="item.status">{{ label(item.status) }}</i></div><div v-if="item.error" class="host-error">{{ item.error }}</div><pre v-if="item.output">{{ item.output }}</pre><div v-else-if="!item.error" class="empty-output">执行完成，无输出</div></article></div><div v-else class="empty-output">暂无执行结果</div></section>
        <el-alert v-if="detail.error_message" :title="detail.error_message" type="error" :closable="false" show-icon />
      </div>
      <template #footer><el-button @click="detail = null">关闭</el-button></template>
    </el-dialog>
  </section>
</template>
<style scoped>
.page :deep(.el-input__inner::placeholder),.page :deep(.el-textarea__inner::placeholder),.page :deep(.el-select__placeholder.is-transparent){font-size:10px}
</style>
<style scoped>.page{height:100%;display:flex;flex-direction:column;gap:8px;min-width:0;padding:0 4px;box-sizing:border-box;font-size:11px}header h2{margin:0;font-size:20px;color:var(--text)}header p{margin:3px 0 0;color:var(--text-secondary)}.toolbar{display:flex;gap:6px;align-items:center}.toolbar .el-input{width:160px}.toolbar .el-select{width:140px}.grid{flex:1;min-height:0}:deep(.job-actions){height:100%;display:flex;align-items:center;justify-content:flex-start}:deep(.view-button){width:auto!important;height:26px!important;line-height:16px!important;border:1px solid var(--primary);border-radius:8px;background:#fff;color:var(--primary);padding:4px 10px;cursor:pointer;box-shadow:0 1px 2px rgba(65,80,190,.1)}:deep(.view-button:hover){background:rgba(91,103,239,.06)}.job-tag,:deep(.job-tag){display:inline-block;border:1px solid var(--line);border-radius:7px;padding:2px 7px;line-height:16px;font-style:normal;background:#f7f8fb;color:var(--text-secondary)}.job-tag.manual,:deep(.job-tag.manual){color:var(--primary);background:#f3f4ff;border-color:#d8dbff}.job-tag.file,:deep(.job-tag.file){color:#8a63d2;background:#f8f2ff;border-color:#e4d5f6}.job-tag.success,:deep(.job-tag.success){color:#43b936;background:#f2faef;border-color:#d6edcf}.job-tag.failed,:deep(.job-tag.failed){color:#ef5b5b;background:#fff4f4;border-color:#f4d0d0}.job-tag.running,:deep(.job-tag.running){color:#d99016;background:#fff8e8;border-color:#f4dfb1}.detail-content{display:flex;flex-direction:column;gap:10px}.detail-summary{display:grid;grid-template-columns:1fr 1fr;border:1px solid var(--line);border-radius:9px;overflow:hidden}.detail-summary>div{display:grid;grid-template-columns:80px 1fr;align-items:center;min-height:38px;border-bottom:1px solid var(--line)}.detail-summary>div:nth-child(odd){border-right:1px solid var(--line)}.detail-summary>div:nth-last-child(-n+2){border-bottom:0}.detail-summary span{align-self:stretch;display:flex;align-items:center;padding:0 10px;background:#f7f8fb;color:var(--text-secondary);font-weight:600}.detail-summary strong,.detail-summary i{margin-left:12px}.detail-section{border:1px solid var(--line);border-radius:9px;padding:10px 12px}.detail-section h3{margin:0 0 8px;font-size:12px}.detail-section h3 small{margin-left:6px;color:var(--text-secondary);font-weight:400}.detail-section p{margin:0}.parameter-list{display:flex;gap:18px;flex-wrap:wrap;color:var(--text-secondary)}.parameter-list b{color:var(--text);font-weight:600}.detail-results{display:grid;gap:8px}.host-result{border:1px solid var(--line);border-radius:8px;overflow:hidden}.host-result-head{display:flex;align-items:center;justify-content:space-between;padding:7px 9px;background:#f8f9fc}.host-result pre{margin:0;padding:9px 10px;max-height:220px;overflow:auto;background:#17191e;color:#e7eaf0;white-space:pre-wrap;font:11px/1.55 ui-monospace,SFMono-Regular,Menlo,monospace}.host-error{padding:8px 10px;color:#ef5b5b;background:#fff7f7}.empty-output{padding:10px;color:var(--text-secondary);text-align:center}</style>
