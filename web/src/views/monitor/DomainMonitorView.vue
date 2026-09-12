<script setup lang="ts">
import { h, onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { checkMonitorDomain, createMonitorDomain, deleteMonitorDomain, listMonitorDomains, type MonitorDomain, type MonitorDomainRequest, updateMonitorDomain } from '@/api/monitor'
const rows = ref<MonitorDomain[]>([]); const keyword = ref(''); const dialog = ref(false); const editing = ref<MonitorDomain | null>(null)
const form = ref<MonitorDomainRequest>({ domain: '', port: 443, check_interval: 300, enable_ssl: true, enable_alert: true, response_threshold: 3000, ssl_expiry_days: 30 })
async function load() { const { data } = await listMonitorDomains(keyword.value); rows.value = data.data }
function open(row?: MonitorDomain) { editing.value = row ?? null; form.value = row ? { id: row.id, domain: row.domain, port: row.port, check_interval: row.check_interval, enable_ssl: row.enable_ssl, enable_alert: row.enable_alert, response_threshold: row.response_threshold, ssl_expiry_days: row.ssl_expiry_days } : { domain: '', port: 443, check_interval: 300, enable_ssl: true, enable_alert: true, response_threshold: 3000, ssl_expiry_days: 30 }; dialog.value = true }
async function save() { if (editing.value) await updateMonitorDomain({ ...form.value, id: editing.value.id }); else await createMonitorDomain(form.value); dialog.value = false; await load() }
async function remove(row: MonitorDomain) { await ElMessageBox.confirm(`删除监控 ${row.domain}？`, '确认删除'); await deleteMonitorDomain(row.id); await load() }
async function check(row: MonitorDomain) { await checkMonitorDomain(row.id); ElMessage.success('检测完成'); await load() }
watch(keyword, load)
const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => p.node!.rowIndex! + 1, width: 70 }, { field: 'domain', headerName: '域名', minWidth: 160 }, { field: 'port', headerName: '端口', width: 80 }, { field: 'status', headerName: '状态', cellRenderer: (p: any) => h(ElTag, { class: `status-tag status-${p.value}`, type: p.value === 'normal' ? 'success' : p.value === 'abnormal' ? 'danger' : 'info' }, () => p.value === 'normal' ? '正常' : p.value === 'abnormal' ? '异常' : p.value === 'paused' ? '暂停' : '未知') }, { field: 'response_time', headerName: '响应时间', valueFormatter: (p: any) => p.value ? `${p.value} ms` : '-' }, { field: 'last_error', headerName: '最近错误', minWidth: 180 }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => h('div', { class: 'grid-actions' }, [h(ElButton, { class: 'grid-action-btn', onClick: () => check(p.data) }, () => '检测'), h(ElButton, { class: 'grid-action-btn', onClick: () => open(p.data) }, () => '编辑'), h(ElButton, { class: 'grid-action-danger', onClick: () => remove(p.data) }, () => '删除')]) }]
onMounted(load)
</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索域名" clearable /><el-button type="primary"
                @click="open()">新增监控</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination /><el-dialog v-model="dialog"
            :title="editing ? '编辑监控' : '新增监控'" width="480px"><el-form label-width="100px"><el-form-item label="域名"><el-input
                        v-model="form.domain" /></el-form-item><el-form-item label="端口"><el-input-number
                        v-model="form.port" :min="1" :max="65535" /></el-form-item><el-form-item
                    label="检测间隔"><el-input-number v-model="form.check_interval" :min="30" />
                    秒</el-form-item><el-form-item label="HTTPS"><el-switch
                        v-model="form.enable_ssl" /></el-form-item><el-form-item label="告警"><el-switch
                        v-model="form.enable_alert" /></el-form-item></el-form><template #footer><el-button
                    @click="dialog = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template></el-dialog>
    </section>
</template>
<style
    scoped>
    .page {
        height: 100%;
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-height: 0
    }

    .toolbar {
        display: flex;
        justify-content: space-between
    }

    .toolbar .el-input {
        width: 260px
    }

    .grid {
        flex: 1;
        min-height: 0
    }

    .grid-actions {
        display: flex;
        align-items: center;
        gap: 6px;
        height: 100%
    }

    :deep(.grid-action-btn),
    :deep(.grid-action-danger) {
        border-radius: 8px;
        padding: 4px 8px;
        height: 28px
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
</style>
