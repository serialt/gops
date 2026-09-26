<script setup lang="ts">
import { h, onMounted, ref } from 'vue'; import type { ColDef } from 'ag-grid-community'; import { ElTag } from 'element-plus'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { listMonitorLogs, type MonitorAlertLog } from '@/api/monitor'
const rows = ref<MonitorAlertLog[]>([]); const page = ref(1); const pageSize = ref(20); const total = ref(0); async function load() { const { data } = await listMonitorLogs(page.value, pageSize.value); rows.value = data.data.list; total.value = data.data.total }; const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node!.rowIndex! + 1, width: 70 }, { field: 'domain_id', headerName: '监控 ID' }, { field: 'alert_type', headerName: '告警类型' }, { field: 'status', headerName: '状态', cellRenderer: (p: any) => h(ElTag, { type: p.value === 'sent' ? 'success' : p.value === 'failed' ? 'danger' : 'warning' }, () => p.value) }, { field: 'message', headerName: '内容', minWidth: 260 }, { field: 'channel_type', headerName: '通道' }, { field: 'sent_at', headerName: '发送时间', minWidth: 160 }]; onMounted(load)
</script>
<template>
    <section class="page">
        <div class="toolbar"><span>告警记录</span></div>
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
    font-size: 14px;
    font-weight: 600
}

.grid {
    flex: 1;
    min-height: 0
}
</style>
