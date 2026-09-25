<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { listKubernetesClusters, type KubernetesCluster } from '@/api/kubernetes'
import { deleteInspection, getInspectionProgress, getInspectionResult, listInspectionHistory, startInspection, type InspectionCheck, type InspectionHistory, type InspectionReport } from '@/api/kubernetes-inspection'

const clusters = ref<KubernetesCluster[]>([]), clusterID = ref(0), running = ref(false), progress = ref(0), step = ref(''), report = ref<InspectionReport | null>(null), history = ref<InspectionHistory[]>([]), historyVisible = ref(false)
const activeSection = ref('all')
let timer = 0
const summary = computed(() => report.value?.summary)
const checks = computed(() => report.value?.sections.flatMap(section => section.checks.map(item => ({ ...item, section: section.title }))) || [])
const reportSections = computed(() => { const order = ['cluster', 'nodes', 'workloads', 'network', 'storage', 'security', 'capacity', 'events']; return order.map(key => report.value?.sections.find(section => section.key === key)).filter(Boolean) as NonNullable<InspectionReport['sections']> })
const activeSectionData = computed(() => reportSections.value.find(section => section.key === activeSection.value))
const visibleChecks = computed(() => activeSection.value === 'all' ? checks.value : activeSectionData.value?.checks.map(item => ({ ...item, section: activeSectionData.value?.title })) || [])
const statusColumn: ColDef = { field: 'status', headerName: '状态', width: 86, minWidth: 86, maxWidth: 86, headerClass: 'inspection-status-header', cellClass: 'inspection-status-cell', cellRenderer: statusRenderer }
const checkColumns: ColDef[] = [{ field: 'section', headerName: '分类', width: 100 }, { field: 'name', headerName: '检查项', minWidth: 130 }, statusColumn, { field: 'value', headerName: '当前值', minWidth: 140 }, { field: 'expected', headerName: '期望值', minWidth: 120 }, { field: 'detail', headerName: '说明', minWidth: 180 }, { field: 'suggestion', headerName: '建议', minWidth: 180 }]
const historyColumns = computed<ColDef[]>(() => [{ headerName: '序号', width: 60, valueGetter: (p: any) => p.node.rowIndex + 1 }, { field: 'cluster_name', headerName: '集群', minWidth: 130 }, statusColumn, { field: 'score', headerName: '评分', width: 80 }, { field: 'check_count', headerName: '检查项', width: 80 }, { field: 'warning_count', headerName: '警告', width: 70 }, { field: 'fail_count', headerName: '异常', width: 70 }, { field: 'duration', headerName: '耗时', width: 75, valueFormatter: (p: any) => `${p.value || 0}s` }, { field: 'started_at', headerName: '巡检时间', minWidth: 155 }, { headerName: '操作', pinned: 'right', width: 150, minWidth: 150, maxWidth: 150, headerClass: 'inspection-action-header', cellClass: 'inspection-action-cell', cellRenderer: historyActions }])
const detailColumns = computed<ColDef[]>(() => activeSection.value === 'nodes' ? [{ field: 'name', headerName: '节点名称', minWidth: 160 }, { field: 'status', headerName: '状态', width: 90 }, { field: 'cpu', headerName: 'CPU容量', width: 100 }, { field: 'memory', headerName: '内存容量', width: 120 }, { field: 'pods', headerName: 'Pod容量', width: 100 }] : [{ field: 'type', headerName: '类型', width: 90 }, { field: 'reason', headerName: '原因', minWidth: 130 }, { field: 'object', headerName: '对象', minWidth: 160 }, { field: 'namespace', headerName: '命名空间', width: 110 }, { field: 'count', headerName: '次数', width: 70 }, { field: 'last_seen', headerName: '最后发生', minWidth: 150 }, { field: 'message', headerName: '消息', minWidth: 260 }])
function statusRenderer(p: any) { const tag = document.createElement('span'), dot = document.createElement('i'), text = document.createElement('span'), value = String(p.value || ''), labels: Record<string, string> = { success: '通过', warning: '警告', error: '异常', running: '运行中', completed: '已完成', failed: '失败' }; tag.className = `inspection-status ${value}`; text.textContent = labels[value] || value; tag.append(dot, text); return tag }
function historyActions(p: any) { const wrap = document.createElement('div'); wrap.className = 'inspection-actions'; const view = document.createElement('button'); view.textContent = '查看'; view.className = 'view'; view.onclick = () => void openResult(p.data.id); const remove = document.createElement('button'); remove.textContent = '删除'; remove.className = 'remove'; remove.onclick = () => void removeHistory(p.data); wrap.append(view, remove); return wrap }
async function loadClusters() { const { data } = await listKubernetesClusters(); clusters.value = data.data; clusterID.value = clusters.value[0]?.id || 0 }
async function start() { if (!clusterID.value) return; running.value = true; report.value = null; progress.value = 5; step.value = '正在初始化巡检'; try { const { data } = await startInspection(clusterID.value); poll(data.data.id) } catch { running.value = false } }
function poll(id: number) { window.clearTimeout(timer); timer = window.setTimeout(async () => { try { const { data } = await getInspectionProgress(id), value = data.data; progress.value = value.progress; step.value = value.current_step; if (value.status === 'running') { poll(id); return } running.value = false; if (value.status === 'completed') { await openResult(id); ElMessage.success('集群巡检完成') } else ElMessage.error('集群巡检失败') } catch { running.value = false } }, 1000) }
async function openResult(id: number) { const { data } = await getInspectionResult(id); report.value = data.data.report; activeSection.value = 'all'; historyVisible.value = false }
async function openHistory() { const { data } = await listInspectionHistory(clusterID.value); history.value = data.data; historyVisible.value = true }
async function removeHistory(row: InspectionHistory) { await ElMessageBox.confirm(`确定删除 ${row.cluster_name} 的巡检记录？`, '删除巡检记录'); await deleteInspection(row.id); history.value = history.value.filter(item => item.id !== row.id); ElMessage.success('删除成功') }
async function batchRemoveHistory(selectedRows: unknown[]) { const selected = selectedRows as InspectionHistory[]; if (!selected.length) return; await ElMessageBox.confirm(`确定删除选中的 ${selected.length} 条巡检记录？`, '批量删除巡检记录'); await Promise.all(selected.map(row => deleteInspection(row.id))); history.value = history.value.filter(row => !selected.some(item => item.id === row.id)); ElMessage.success('批量删除成功') }
onBeforeUnmount(() => window.clearTimeout(timer))
loadClusters()
</script>

<template>
  <section class="inspection-page">
    <header class="inspection-header">
      <div>
        <h2>集群巡检</h2>
        <p>检查 Kubernetes 集群健康状态并生成巡检报告</p>
      </div>
      <div class="header-actions"><el-select v-model="clusterID" placeholder="选择集群" :disabled="running"><el-option
            v-for="item in clusters" :key="item.id" :label="item.alias || item.name"
            :value="item.id" /></el-select><el-button @click="openHistory">巡检历史</el-button><el-button type="primary"
          :icon="Refresh" :loading="running" @click="start">{{ running ? '巡检中' : '开始巡检' }}</el-button></div>
    </header>
    <div v-if="running" class="progress-card">
      <div><strong>{{ step }}</strong><span>{{ progress }}%</span></div><el-progress :percentage="progress"
        :stroke-width="8" :show-text="false" />
    </div>
    <div v-if="!report && !running" class="empty-card">
      <div class="empty-icon">✓</div>
      <h3>开始集群健康巡检</h3>
      <p>将并发检查集群连接、节点、工作负载、网络、存储、配置和事件。</p><el-button type="primary" :disabled="!clusterID"
        @click="start">立即巡检</el-button>
    </div>
    <template v-if="report && !running">
      <div class="summary-grid">
        <article class="score-card">
          <div class="score" :class="report.score >= 90 ? 'good' : report.score >= 70 ? 'warn' : 'bad'">{{ report.score }}</div>
          <span>健康评分</span><strong>{{ report.cluster_name }}</strong>
        </article>
        <article><strong>{{ summary?.total }}</strong><span>检查项</span></article>
        <article><strong class="success">{{ summary?.passed }}</strong><span>通过</span></article>
        <article><strong class="warning">{{ summary?.warning }}</strong><span>警告</span></article>
        <article><strong class="danger">{{ summary?.failed }}</strong><span>异常</span></article>
        <article><strong>{{ summary?.duration }}s</strong><span>巡检耗时</span></article>
      </div>
      <div class="report-card">
        <nav class="section-tabs"><button :class="{ active: activeSection === 'all' }"
            @click="activeSection = 'all'">概览</button><button v-for="section in reportSections" :key="section.key"
            :class="{ active: activeSection === section.key }" @click="activeSection = section.key">{{ section.title }}</button>
        </nav>
        <div v-if="activeSection === 'all'" class="status-overview"><span><i class="ok">✓</i>正常
            <strong>{{ summary?.passed }}</strong></span><span><i class="warn">!</i>警告
            <strong>{{ summary?.warning }}</strong></span><span><i class="bad">×</i>异常
            <strong>{{ summary?.failed }}</strong></span></div>
        <div v-if="activeSectionData?.metrics?.length" class="section-metrics">
          <article v-for="item in activeSectionData.metrics" :key="item.label"><span>{{ item.label }}</span><strong
              :class="item.tone">{{ item.value }}</strong></article>
        </div>
        <div v-if="activeSectionData?.details?.length" class="detail-block">
          <h3>{{ activeSection === 'nodes' ? '节点资源概况' : '最近告警事件' }}</h3>
          <AppDataGrid class="detail-grid" :row-data="activeSectionData.details" :column-defs="detailColumns" />
        </div>
        <h3 v-if="activeSection !== 'all'" class="checks-title">检查项</h3>
        <AppDataGrid class="checks-grid" :row-data="visibleChecks" :column-defs="checkColumns"
          :get-row-id="(row: unknown) => `${(row as InspectionCheck).category}/${(row as InspectionCheck).name}`"
          pagination />
      </div>
    </template>
    <el-dialog v-model="historyVisible" title="巡检历史" width="min(1000px,calc(100vw - 20px))" align-center>
      <AppDataGrid class="history-grid" :row-data="history" :column-defs="historyColumns"
        :get-row-id="(row: unknown) => String((row as InspectionHistory).id)" :bulk-delete="batchRemoveHistory"
        pagination />
    </el-dialog>
  </section>
</template>

<style scoped>
.inspection-page {
  display: flex;
  height: 100%;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: 8px;
  font-size: 11px
}

.inspection-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 9px;
  background: #fff
}

.inspection-header h2 {
  margin: 0;
  color: var(--text);
  font-size: 17px
}

.inspection-header p {
  margin: 3px 0 0;
  color: #9298a5
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 6px
}

.header-actions .el-select {
  width: 190px
}

.progress-card,
.empty-card,
.report-card {
  border: 1px solid var(--line);
  border-radius: 9px;
  background: #fff
}

.progress-card {
  padding: 14px
}

.progress-card>div {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px
}

.empty-card {
  display: grid;
  flex: 1;
  place-content: center;
  justify-items: center;
  color: #8f96a3;
  text-align: center
}

.empty-card h3 {
  margin: 8px 0 2px;
  color: var(--text);
  font-size: 14px
}

.empty-card p {
  margin: 0 0 12px
}

.empty-icon {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 50%;
  background: var(--primary-soft);
  color: var(--primary);
  font-size: 25px
}

.summary-grid {
  display: grid;
  grid-template-columns: 1.5fr repeat(5, 1fr);
  gap: 7px
}

.summary-grid article {
  display: flex;
  min-height: 72px;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #fff
}

.summary-grid article strong {
  font-size: 18px
}

.summary-grid article span {
  color: #9298a5
}

.score-card {
  display: grid !important;
  grid-template-columns: 52px 1fr !important;
  grid-template-rows: 1fr 1fr !important;
  column-gap: 8px !important
}

.score-card .score {
  display: grid;
  grid-row: 1/3;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 50%;
  background: #ecf8f1;
  color: #22955b;
  font-size: 20px;
  font-weight: 800
}

.score-card .score.warn {
  background: #fff7df;
  color: #c28b00
}

.score-card .score.bad {
  background: #fff0f0;
  color: #e45656
}

.score-card span,
.score-card strong {
  justify-self: start
}

.score-card strong {
  font-size: 11px !important
}

.success {
  color: #24965b
}

.warning {
  color: #c28b00
}

.danger {
  color: #e45656
}

.report-card {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  padding: 8px
}

.section-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding-bottom: 7px
}

.section-tabs button {
  padding: 5px 9px;
  border: 1px solid var(--line);
  border-radius: 7px;
  background: #fff;
  color: var(--text);
  cursor: pointer
}

.section-tabs button.active {
  border-color: var(--primary);
  background: var(--primary-soft);
  color: var(--primary)
}

.section-tabs small {
  margin-left: 3px
}

.checks-grid {
  flex: 1;
  min-height: 250px
}

.history-grid {
  height: 500px
}

.inspection-page :deep(.inspection-status-header .ag-header-cell-label) {
  justify-content: center
}

.inspection-page :deep(.inspection-status-cell) {
  display: flex;
  align-items: center;
  justify-content: center
}

.inspection-page :deep(.inspection-status) {
  display: inline-flex;
  min-width: 54px;
  height: 23px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 0 7px;
  border: 1px solid #bfe7d0;
  border-radius: 6px;
  background: #f0faf4;
  color: #24965b;
  font-size: 10px;
  font-weight: 600;
  line-height: 21px;
  white-space: nowrap
}

.inspection-page :deep(.inspection-status i) {
  width: 5px;
  height: 5px;
  flex: none;
  border-radius: 50%;
  background: currentColor
}

.inspection-page :deep(.inspection-status.warning),
.inspection-page :deep(.inspection-status.running) {
  border-color: #ead89b;
  background: #fff9e8;
  color: #ad7c00
}

.inspection-page :deep(.inspection-status.error),
.inspection-page :deep(.inspection-status.failed) {
  border-color: #f2bcbc;
  background: #fff2f2;
  color: #dd5050
}

.inspection-page :deep(.inspection-status.completed) {
  border-color: #bfcaf6;
  background: #f2f4ff;
  color: var(--primary)
}

.inspection-page :deep(.inspection-actions) {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 100%
}

.inspection-page :deep(.inspection-actions button) {
  padding: 3px 8px;
  border-radius: 7px;
  background: #fff;
  cursor: pointer
}

.inspection-page :deep(.inspection-actions .view) {
  border: 1px solid var(--primary);
  color: var(--primary)
}

.inspection-page :deep(.inspection-actions .remove) {
  border: 1px solid #efb5b5;
  color: #df5252
}

@media(max-width:800px) {
  .inspection-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap
  }

  .summary-grid {
    grid-template-columns: repeat(3, 1fr)
  }

  .score-card {
    grid-column: span 3
  }
}

.inspection-page :deep(.inspection-action-header .ag-header-cell-label) {
  justify-content: flex-start
}

.inspection-page :deep(.inspection-action-cell) {
  display: flex;
  align-items: center;
  justify-content: flex-start
}

.inspection-page :deep(.inspection-actions) {
  justify-content: flex-start;
  gap: 6px;
  width: 100%
}

.inspection-page :deep(.inspection-actions button) {
  display: inline-flex;
  height: 23px;
  align-items: center;
  justify-content: center;
  padding: 0 9px;
  border-radius: 7px;
  font-size: 10px;
  line-height: 21px
}

.inspection-page :deep(.inspection-actions .view) {
  box-shadow: 0 1px 3px color-mix(in srgb, var(--primary) 12%, transparent)
}

.inspection-page :deep(.inspection-actions .view:hover) {
  background: var(--primary-soft)
}

.inspection-page :deep(.inspection-actions .remove) {
  background: #fff7f7
}

.inspection-page :deep(.inspection-actions .remove:hover) {
  background: #fff0f0
}

.status-overview {
  display: flex;
  gap: 28px;
  margin-bottom: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f6f7f9
}

.status-overview span {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #777e89
}

.status-overview i {
  display: grid;
  width: 17px;
  height: 17px;
  place-items: center;
  border-radius: 50%;
  background: #52c41a;
  color: #fff;
  font-style: normal
}

.status-overview i.warn {
  background: #e7a520
}

.status-overview i.bad {
  background: #f05b64
}

.status-overview strong {
  color: var(--text);
  font-size: 16px
}

.section-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0;
  margin-bottom: 10px;
  padding: 11px 12px;
  border-radius: 8px;
  background: #f6f7f9
}

.section-metrics article {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px
}

.section-metrics span {
  color: #9298a5
}

.section-metrics strong {
  overflow: hidden;
  color: var(--text);
  font-size: 16px;
  text-overflow: ellipsis;
  white-space: nowrap
}

.section-metrics strong.success {
  color: #52c41a
}

.section-metrics strong.warning {
  color: #e7a520
}

.section-metrics strong.error {
  color: #f05b64
}

.detail-block {
  margin-bottom: 8px
}

.detail-block h3,
.checks-title {
  margin: 5px 0 6px;
  font-size: 11px
}

.detail-grid {
  height: 150px
}

.checks-title+.checks-grid {
  min-height: 180px
}

@media(max-width:800px) {
  .section-metrics {
    grid-template-columns: repeat(2, 1fr)
  }
}
</style>
