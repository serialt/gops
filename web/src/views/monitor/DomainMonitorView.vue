<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Filter, Upload } from '@element-plus/icons-vue'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { checkMonitorDomain, createMonitorDomain, deleteMonitorDomain, exportMonitorDomains, importMonitorDomains, listMonitorCheckHistory, listMonitorDomains, monitorStats, type MonitorCheckHistory, type MonitorDomain, type MonitorDomainRequest, type MonitorRequestClient, type MonitorStats, updateMonitorDomain } from '@/api/monitor'
import { useAuthStore } from '@/stores/auth'
import { downloadAttachment } from '@/utils/download'
const rows = ref<MonitorDomain[]>([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const selectedRows = ref<MonitorDomain[]>([]); const historyRows = ref<MonitorCheckHistory[]>([]); const stats = ref<MonitorStats>({ total: 0, normal: 0, abnormal: 0, paused: 0 }); const keyword = ref(''); const statusFilter = ref(''); const sslStatusFilter = ref(''); const sslExpiryRange = ref<[string, string] | null>(null); const filtersExpanded = ref(false); const dialog = ref(false); const detailDialog = ref(false); const detailMonitor = ref<MonitorDomain | null>(null); const editing = ref<MonitorDomain | null>(null)
const activeFilterCount = () => Number(Boolean(statusFilter.value)) + Number(Boolean(sslStatusFilter.value)) + Number(Boolean(sslExpiryRange.value))
function clearFilters() {
    statusFilter.value = ''
    sslStatusFilter.value = ''
    sslExpiryRange.value = null
}
const importInput = ref<HTMLInputElement>(); const exporting = ref(false); const importing = ref(false)
const auth = useAuthStore()
const requestClientOptions: Array<{ label: string; value: MonitorRequestClient }> = [{ label: 'Google Chrome', value: 'chrome' }, { label: 'Mozilla Firefox', value: 'firefox' }, { label: 'Microsoft Edge', value: 'edge' }, { label: 'Apple Safari', value: 'safari' }, { label: 'curl', value: 'curl' }]
const form = ref<MonitorDomainRequest>({ domain: '', port: 443, request_client: 'chrome', check_interval: 300, check_jitter: 30, enable_ssl: true, verify_certificate: true, enable_alert: true, response_threshold: 3000, ssl_expiry_days: 30 })
async function load() {
    const expiryStart = sslExpiryRange.value?.[0] || ''
    const expiryEnd = sslExpiryRange.value?.[1] || ''
    const [listResponse, statsResponse] = await Promise.all([
        listMonitorDomains(keyword.value, statusFilter.value, page.value, pageSize.value, sslStatusFilter.value, expiryStart, expiryEnd),
        monitorStats(keyword.value, statusFilter.value, sslStatusFilter.value, expiryStart, expiryEnd),
    ])
    rows.value = listResponse.data.data.list
    total.value = listResponse.data.data.total
    stats.value = statsResponse.data.data
}
function open(row?: MonitorDomain) { editing.value = row ?? null; form.value = row ? { id: row.id, domain: row.domain, port: row.port, request_client: row.request_client || 'chrome', check_interval: row.check_interval, check_jitter: row.check_jitter ?? 30, enable_ssl: row.enable_ssl, verify_certificate: row.verify_certificate, enable_alert: row.enable_alert, response_threshold: row.response_threshold, ssl_expiry_days: row.ssl_expiry_days } : { domain: '', port: 443, request_client: 'chrome', check_interval: 300, check_jitter: 30, enable_ssl: true, verify_certificate: true, enable_alert: true, response_threshold: 3000, ssl_expiry_days: 30 }; dialog.value = true }
function handleHTTPSChange(value: boolean | string | number) { if (!value) form.value.verify_certificate = false }
async function save() {
    const isCreate = !editing.value
    const result = editing.value
        ? await updateMonitorDomain({ ...form.value, id: editing.value.id })
        : await createMonitorDomain(form.value)
    await checkMonitorDomain(result.data.data.id)
    dialog.value = false
    if (isCreate) keyword.value = ''
    await load()
    ElMessage.success(isCreate ? '监控已新增并完成首次检测' : '监控已保存并重新检测')
}
async function remove(row: MonitorDomain) { await ElMessageBox.confirm(`删除监控 ${row.domain}？`, '确认删除'); await deleteMonitorDomain(row.id); await load() }
async function batchRemove(selectedRows: unknown[]) { const selected = selectedRows as MonitorDomain[]; if (!selected.length) return; await ElMessageBox.confirm(`确定删除选中的 ${selected.length} 条监控？`, '批量删除确认', { type: 'warning' }); await Promise.all(selected.map(row => deleteMonitorDomain(row.id))); ElMessage.success('批量删除成功'); await load() }
async function check(row: MonitorDomain) { await checkMonitorDomain(row.id); ElMessage.success('检测完成'); await load() }
async function detail(row: MonitorDomain) { const { data } = await listMonitorCheckHistory(row.id); historyRows.value = data.data.list; detailMonitor.value = row; detailDialog.value = true }
async function exportExcel() {
    exporting.value = true
    try {
        const response = await exportMonitorDomains(keyword.value, selectedRows.value.map(row => row.id))
        downloadAttachment(response.data, response.headers['content-disposition'], '域名监控.xlsx')
        ElMessage.success(`已导出 ${rows.value.length} 条监控配置`)
    } finally { exporting.value = false }
}
function chooseImport() { importInput.value?.click() }
async function importExcel(event: Event) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    importing.value = true
    try {
        const { data } = await importMonitorDomains(file)
        await load()
        ElMessage.success(`导入完成：新增 ${data.data.created} 条，更新 ${data.data.updated} 条`)
    } finally { importing.value = false; input.value = '' }
}
watch([keyword, statusFilter, sslStatusFilter, sslExpiryRange], () => { page.value = 1; void load() })
function formatDate(value?: string) { return value && !value.startsWith('0001-') ? value.replace('T', ' ').slice(0, 19) : '—' }
function statusText(status?: string) { return status === 'normal' ? '正常' : status === 'abnormal' ? '异常' : status === 'paused' ? '暂停' : '未知' }
function intervalText(seconds?: number) { if (!seconds) return '—'; return seconds % 60 === 0 ? `${seconds / 60}分钟` : `${seconds}秒` }
function requestClientText(value?: string) { return requestClientOptions.find(option => option.value === value)?.label ?? 'Google Chrome' }
function statusRenderer(text: string, kind: string) {
    const tag = document.createElement('span')
    tag.className = 'monitor-status ' + kind
    tag.textContent = text
    return tag
}
function operationRenderer(p: any) {
    const actions = document.createElement('div')
    actions.className = 'grid-actions'
    const createButton = (text: string, action: () => void, danger = false) => {
        const button = document.createElement('button')
        button.type = 'button'
        button.className = 'monitor-action-text' + (danger ? ' is-danger' : '')
        button.textContent = text
        button.addEventListener('click', event => { event.stopPropagation(); action() })
        return button
    }
    actions.append(
        createButton('详情', () => detail(p.data)),
        createButton('检测', () => check(p.data)),
        createButton('编辑', () => open(p.data)),
        createButton('删除', () => remove(p.data), true),
    )
    return actions
}
const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node!.rowIndex! + 1, width: 70 }, { field: 'domain', headerName: '域名', minWidth: 160 }, { field: 'port', headerName: '端口', width: 80 }, { field: 'status', headerName: '状态', width: 90, cellRenderer: (p: any) => statusRenderer(p.value === 'normal' ? '正常' : p.value === 'abnormal' ? '异常' : p.value === 'paused' ? '暂停' : '未知', p.value === 'normal' ? 'is-success' : p.value === 'abnormal' ? 'is-danger' : 'is-muted') }, { field: 'response_time', headerName: '响应时间', width: 105, valueFormatter: (p: any) => p.value ? `${p.value} ms` : '—' }, { field: 'ssl_valid', headerName: 'SSL 证书', width: 100, cellRenderer: (p: any) => statusRenderer(p.data.enable_ssl ? (p.value ? '有效' : '无效') : '未启用', p.data.enable_ssl ? (p.value ? 'is-success' : 'is-danger') : 'is-muted') }, { field: 'ssl_root_ca', headerName: '根 CA', minWidth: 150, valueFormatter: (p: any) => p.value || '—' }, { field: 'ssl_expiry', headerName: 'SSL 到期时间', minWidth: 155, valueFormatter: (p: any) => formatDate(p.value) }, { field: 'check_interval', headerName: '检查间隔', width: 100, valueFormatter: (p: any) => `${p.value} 秒` }, { field: 'last_check', headerName: '最后检查', minWidth: 155, valueFormatter: (p: any) => formatDate(p.value) }, { headerName: '操作', pinned: 'right', cellClass: 'monitor-operation-cell', cellRenderer: operationRenderer }]
onMounted(load)
</script>
<template>
    <section class="page">
        <div class="monitor-stats">
            <div class="stat-card"><span>监控总数</span><strong>{{ stats.total }}</strong></div>
            <div class="stat-card is-normal"><span>正常</span><strong>{{ stats.normal }}</strong></div>
            <div class="stat-card is-abnormal"><span>异常</span><strong>{{ stats.abnormal }}</strong></div>
            <div class="stat-card is-paused"><span>暂停</span><strong>{{ stats.paused }}</strong></div>
        </div>
        <div class="toolbar">
            <div class="toolbar-filters">
                <el-input v-model="keyword" class="domain-search" placeholder="搜索域名、端口或根 CA" clearable />
                <el-button class="filter-toggle" :type="filtersExpanded ? 'primary' : 'default'"
                    :plain="filtersExpanded" :icon="Filter" @click="filtersExpanded = !filtersExpanded">
                    筛选<span v-if="activeFilterCount()" class="filter-count">{{ activeFilterCount() }}</span>
                    <el-icon class="filter-toggle-chevron" :class="{ expanded: filtersExpanded }">
                        <ArrowDown />
                    </el-icon>
                </el-button>
                <el-button class="clear-filter-button" :disabled="!activeFilterCount()" @click="clearFilters">
                    清除筛选
                </el-button>
            </div>
            <div class="toolbar-actions"><el-dropdown
                    v-if="auth.hasPermission('POST /api/v1/monitor/domains/export') || auth.hasPermission('POST /api/v1/monitor/domains/import')"
                    split-button class="excel-transfer-dropdown" popper-class="excel-transfer-popper"
                    :loading="exporting || importing"
                    :disabled="!auth.hasPermission('POST /api/v1/monitor/domains/export')" @click="exportExcel"
                    @command="chooseImport">导出 Excel
                    <template #dropdown><el-dropdown-menu><el-dropdown-item
                                v-if="auth.hasPermission('POST /api/v1/monitor/domains/import')" command="import"
                                :icon="Upload">从 Excel 导入</el-dropdown-item></el-dropdown-menu></template>
                </el-dropdown><input ref="importInput" type="file" accept=".xlsx" hidden
                    @change="importExcel" /><el-button type="primary" @click="open()">新增监控</el-button></div>
        </div>
        <div v-if="filtersExpanded" class="monitor-filter-panel">
            <el-select v-model="statusFilter" class="status-filter" placeholder="全部状态" clearable>
                <el-option label="正常" value="normal" />
                <el-option label="异常" value="abnormal" />
                <el-option label="暂停" value="paused" />
                <el-option label="未知" value="unknown" />
            </el-select>
            <el-select v-model="sslStatusFilter" class="ssl-filter" placeholder="SSL 证书" clearable>
                <el-option label="有效" value="valid" />
                <el-option label="无效" value="invalid" />
                <el-option label="未启用" value="disabled" />
            </el-select>
            <el-date-picker v-model="sslExpiryRange" class="expiry-filter" type="daterange" value-format="YYYY-MM-DD"
                start-placeholder="到期开始" end-placeholder="到期结束" range-separator="至" clearable />
        </div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total"
            :current-page="page" :page-size="pageSize" :bulk-delete="batchRemove"
            @selection-changed="selectedRows = $event as MonitorDomain[]"
            @page-change="value => { page = value; void load() }"
            @page-size-change="value => { pageSize = value; page = 1; void load() }" />
        <el-dialog v-model="dialog" :title="editing ? '编辑监控' : '新增监控'" width="480px"><el-form
                label-width="100px"><el-form-item label="域名"><el-input
                        v-model="form.domain" /></el-form-item><el-form-item label="端口"><el-input-number
                        v-model="form.port" :min="1" :max="65535" /></el-form-item><el-form-item
                    label="请求客户端"><el-select v-model="form.request_client"><el-option
                            v-for="option in requestClientOptions" :key="option.value" :label="option.label"
                            :value="option.value" /></el-select></el-form-item><el-form-item
                    label="检测间隔"><el-input-number v-model="form.check_interval" :min="30" />
                    秒</el-form-item><el-form-item label="抖动"><el-input-number v-model="form.check_jitter" :min="0"
                        :max="3600" />
                    秒</el-form-item><el-form-item label="连接安全" class="security-options"><el-checkbox
                        v-model="form.enable_ssl" @change="handleHTTPSChange">HTTPS</el-checkbox><el-checkbox
                        v-model="form.verify_certificate"
                        :disabled="!form.enable_ssl">证书校验</el-checkbox></el-form-item><el-form-item
                    label="告警"><el-switch v-model="form.enable_alert" /></el-form-item></el-form><template
                #footer><el-button @click="dialog = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template></el-dialog>
        <el-dialog v-model="detailDialog" title="监控详情" width="720px" class="monitor-detail-dialog">
            <template v-if="detailMonitor">
                <div class="detail-overview">
                    <div class="detail-item"><span>域名：</span><strong class="domain-value">{{ detailMonitor.domain
                            }}</strong></div>
                    <div class="detail-item"><span>状态：</span><strong class="detail-status"
                            :class="`is-${detailMonitor.status}`">{{ statusText(detailMonitor.status) }}</strong></div>
                    <div class="detail-item"><span>端口：</span><strong>{{ detailMonitor.port }}</strong></div>
                    <div class="detail-item"><span>请求客户端：</span><strong>{{
                        requestClientText(detailMonitor.request_client) }}</strong></div>
                    <div class="detail-item"><span>响应时间：</span><strong>{{ detailMonitor.response_time }}ms</strong>
                    </div>
                    <div class="detail-item"><span>SSL证书：</span><strong class="detail-status"
                            :class="detailMonitor.ssl_valid ? 'is-normal' : 'is-abnormal'">{{ detailMonitor.enable_ssl ?
                                (detailMonitor.ssl_valid ? '有效' : '无效') : '未启用' }}</strong></div>
                    <div class="detail-item"><span>根 CA：</span><strong>{{ detailMonitor.ssl_root_ca || '—' }}</strong>
                    </div>
                    <div class="detail-item"><span>SSL到期：</span><strong>{{ formatDate(detailMonitor.ssl_expiry)
                            }}</strong></div>
                    <div class="detail-item"><span>检查间隔：</span><strong>{{ intervalText(detailMonitor.check_interval)
                            }}</strong></div>
                    <div class="detail-item"><span>抖动：</span><strong>0～{{ detailMonitor.check_jitter }}秒</strong></div>
                    <div class="detail-item"><span>最后检查：</span><strong>{{ formatDate(detailMonitor.last_check)
                            }}</strong></div>
                    <div class="detail-item"><span>创建时间：</span><strong>{{ formatDate(detailMonitor.created_at)
                            }}</strong></div>
                    <div class="detail-item"><span>更新时间：</span><strong>{{ formatDate(detailMonitor.updated_at)
                            }}</strong></div>
                </div>
                <div class="history-panel">
                    <h3>检查历史 <small>最近 {{ historyRows.length }} 条</small></h3>
                    <div v-if="historyRows.length" class="history-timeline">
                        <div v-for="item in historyRows" :key="item.id" class="history-entry">
                            <span class="history-dot"></span>
                            <div class="history-time">{{ formatDate(item.checked_at) }}</div>
                            <div class="history-result">
                                <span class="detail-status" :class="`is-${item.status}`">{{ statusText(item.status)
                                    }}</span>
                                <strong>{{ item.response_time }}ms</strong>
                                <span v-if="item.status_code" class="history-code">HTTP {{ item.status_code }}</span>
                            </div>
                            <div v-if="item.error_message" class="history-error">{{ item.error_message }}</div>
                        </div>
                    </div>
                    <el-empty v-else description="暂无检查记录" :image-size="64" />
                </div>
            </template>
        </el-dialog>
    </section>
</template>
<style scoped>
.page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0
}

.toolbar {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    flex-wrap: wrap;
}

.toolbar-filters {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 4px;
}

.toolbar-filters :deep(.el-button + .el-button) {
    margin-left: 0;
}

.toolbar :deep(.domain-search) {
    flex: 0 0 260px;
    width: 260px;
}

.filter-toggle {
    flex: 0 0 auto;
}

.filter-count {
    display: inline-flex;
    min-width: 18px;
    height: 18px;
    align-items: center;
    justify-content: center;
    margin-left: 6px;
    padding: 0 4px;
    border-radius: 9px;
    background: color-mix(in srgb, var(--primary) 12%, white);
    color: var(--primary);
    font-size: 10px;
}

.filter-toggle-chevron {
    margin-left: 4px;
    transition: transform 150ms ease;
}

.filter-toggle-chevron.expanded {
    transform: rotate(180deg);
}

.monitor-filter-panel {
    display: flex;
    width: calc(100% - 1px);
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    margin-right: 1px;
    padding: 1px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-card);
}

.monitor-filter-panel :deep(.status-filter) {
    flex: 0 0 140px;
    width: 140px;
}

.monitor-filter-panel :deep(.ssl-filter) {
    flex: 0 0 150px;
    width: 150px;
}

.monitor-filter-panel :deep(.expiry-filter) {
    flex: 0 0 280px;
    width: 280px;
}

.clear-filter-button {
    flex: 0 0 auto;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: #fff;
    color: var(--text-soft);
}

.clear-filter-button:not(.is-disabled):hover {
    border-color: var(--primary);
    background: color-mix(in srgb, var(--primary) 6%, white);
    color: var(--primary);
}

.toolbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 0 0 auto;
    margin-left: auto;
    white-space: nowrap;
}

.toolbar-actions :deep(.el-button + .el-button) {
    margin-left: 0;
}

.monitor-stats {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 140px));
    gap: 10px;
}

.stat-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 54px;
    padding: 8px 12px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-card);
    color: var(--text-soft);
    font-size: 11px;
}

.stat-card strong {
    color: var(--text);
    font-size: 20px;
    line-height: 1;
}

.stat-card.is-normal strong {
    color: #67c23a;
}

.stat-card.is-abnormal strong {
    color: #ef6262;
}

.stat-card.is-paused strong {
    color: var(--text-muted);
}

.grid {
    flex: 1;
    min-height: 0
}

:deep(.monitor-detail-dialog .el-dialog__body) {
    max-height: 72vh;
    padding: 12px 16px 16px;
    overflow-y: auto;
}

.detail-overview {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px 28px;
    padding: 18px 20px;
    border: 1px solid var(--line);
    border-radius: 9px;
    background: #fafbfc;
}

.detail-item {
    display: grid;
    grid-template-columns: 88px minmax(0, 1fr);
    align-items: center;
    min-width: 0;
}

.detail-item>span {
    color: var(--text-soft);
    font-weight: 600;
}

.detail-item strong {
    min-width: 0;
    color: var(--text);
    font-weight: 500;
    overflow-wrap: anywhere;
}

.detail-item .domain-value {
    color: var(--primary);
}

.detail-status {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    justify-self: start;
    min-width: 44px;
    height: 26px;
    padding: 0 9px;
    border: 1px solid var(--line);
    border-radius: 7px;
    background: #f5f6f8;
    color: var(--text-soft);
    font-weight: 500;
}

.detail-status.is-normal {
    border-color: #ccefc1;
    background: #f4fbf1;
    color: #58b83d;
}

.detail-status.is-abnormal {
    border-color: #ffd5d5;
    background: #fff7f7;
    color: #ef6262;
}

.detail-status.is-paused,
.detail-status.is-unknown {
    color: var(--text-muted);
}

.history-panel {
    margin-top: 16px;
    padding: 16px 20px;
    border: 1px solid var(--line);
    border-radius: 9px;
    background: #fafbfc;
}

.history-panel h3 {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin: 0 0 14px;
    font-size: 15px;
    color: var(--text);
}

.history-panel h3 small {
    color: var(--text-muted);
    font-size: 10px;
    font-weight: 400;
}

.history-timeline {
    padding-left: 8px;
}

.history-entry {
    position: relative;
    min-height: 68px;
    padding: 0 0 14px 28px;
    border-left: 2px solid var(--line);
}

.history-entry:last-child {
    min-height: 48px;
    padding-bottom: 0;
    border-left-color: transparent;
}

.history-dot {
    position: absolute;
    top: 1px;
    left: -7px;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #e5e8ef;
}

.history-time {
    margin-bottom: 7px;
    color: var(--text-muted);
    font-size: 12px;
}

.history-result {
    display: flex;
    align-items: center;
    gap: 10px;
}

.history-result strong {
    color: var(--text-soft);
    font-weight: 500;
}

.history-code {
    color: var(--text-muted);
    font-size: 10px;
}

.history-error {
    margin-top: 6px;
    color: #ef6262;
    font-size: 10px;
    overflow-wrap: anywhere;
}

:deep(.grid-actions) {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    height: 100%;
    line-height: normal
}

:deep(.monitor-operation-cell) {
    display: flex;
    align-items: center;
}

:deep(.monitor-operation-cell .ag-cell-wrapper),
:deep(.monitor-operation-cell .ag-cell-value) {
    display: flex;
    align-items: center;
    width: 100%;
    height: 100%;
}

:deep(.grid-action-btn),
:deep(.grid-action-danger) {
    border-radius: 8px;
    padding: 4px 8px;
    height: 28px
}

:deep(.monitor-action-text) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 26px;
    padding: 0 7px;
    border: 1px solid var(--primary);
    border-radius: 8px;
    color: var(--primary);
    background: #fff;
    font: inherit;
    white-space: nowrap;
    cursor: pointer;
}

:deep(.monitor-action-text:hover) {
    background: var(--primary-soft);
}

:deep(.monitor-action-text.is-danger) {
    border-color: #ffd5d5;
    color: #ef6262;
    background: #fff7f7;
}

:deep(.monitor-action-text.is-danger:hover) {
    border-color: #ef6262;
}

:deep(.monitor-status) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 26px;
    padding: 0 9px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: #f7f7f8;
    color: var(--text-soft);
}

:deep(.monitor-status.is-success) {
    border-color: #ccefc1;
    background: #f4fbf1;
    color: #58b83d;
}

:deep(.monitor-status.is-danger) {
    border-color: #ffd5d5;
    background: #fff7f7;
    color: #ef6262;
}

:deep(.monitor-status.is-muted) {
    color: var(--text-muted);
}

:deep(.grid-action-btn) {
    border-color: var(--primary);
    color: var(--primary)
}

:deep(.grid-action-danger) {
    color: var(--el-color-danger);
    border-color: var(--el-color-danger-light-5);
    background: #fff7f7
}

@media (max-width: 720px) {
    .toolbar {
        align-items: flex-start;
    }

    .toolbar-filters {
        flex: 1 1 100%;
    }

    .toolbar-actions {
        margin-left: 0;
    }

    .toolbar :deep(.domain-search) {
        flex: 1 1 180px;
        width: auto;
        max-width: 260px;
    }

    .detail-overview {
        grid-template-columns: 1fr;
    }
}
</style>
