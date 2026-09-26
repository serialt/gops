<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElTag } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { listMonitorLogs, type MonitorAlertLog } from '@/api/monitor'

const rows = ref<MonitorAlertLog[]>([])
const keyword = ref('')
const status = ref('')
const domainId = ref<number | undefined>()
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function load() {
    const { data } = await listMonitorLogs(page.value, pageSize.value, keyword.value.trim(), status.value, domainId.value || 0)
    rows.value = data.data.list
    total.value = data.data.total
}

function search() {
    page.value = 1
    void load()
}

function reset() {
    keyword.value = ''
    status.value = ''
    domainId.value = undefined
    search()
}

const columns: ColDef[] = [
    { headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node!.rowIndex! + 1, width: 70 },
    { field: 'alert_type', headerName: '告警类型' },
    { field: 'status', headerName: '状态', cellRenderer: (p: any) => h(ElTag, { type: p.value === 'sent' ? 'success' : p.value === 'failed' ? 'danger' : 'warning' }, () => p.value) },
    { field: 'message', headerName: '内容', minWidth: 260, tooltipField: 'message' },
    { field: 'channel_type', headerName: '通道' },
    { field: 'sent_at', headerName: '发送时间', minWidth: 160 },
]

onMounted(load)
</script>
<template>
    <section class="page">
                <div class="toolbar">
                        <el-input v-model="keyword" placeholder="搜索告警类型、内容、通道或错误信息" clearable @keyup.enter="search" />
                        <el-input-number v-model="domainId" :min="1" :precision="0" :controls="false" placeholder="监控 ID" @keyup.enter="search" />
                        <el-select v-model="status" placeholder="全部状态" clearable>
                                <el-option label="待发送" value="pending" />
                                <el-option label="已发送" value="sent" />
                                <el-option label="发送失败" value="failed" />
                        </el-select>
                        <el-button type="primary" @click="search">查询</el-button>
                        <el-button @click="reset">重置</el-button>
                </div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total"
            :current-page="page" :page-size="pageSize" @page-change="value => { page = value; void load() }"
            @page-size-change="value => { pageSize = value; page = 1; void load() }" />
    </section>
</template>
<style scoped>
.page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap
}

.toolbar :deep(.el-input) {
    width: 280px
}

.toolbar :deep(.el-input-number) {
    width: 130px
}

.toolbar :deep(.el-select) {
    width: 140px
}

.grid {
    flex: 1;
    min-height: 0
}
</style>
