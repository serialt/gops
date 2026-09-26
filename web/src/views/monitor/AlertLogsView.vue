<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { listMonitorLogs, type MonitorAlertLog } from '@/api/monitor'

const rows = ref<MonitorAlertLog[]>([])
const keyword = ref('')
const receiverName = ref('')
const status = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const toolbarRef = ref<HTMLElement>()
let searchTimer: ReturnType<typeof setTimeout> | undefined

async function load() {
    const { data } = await listMonitorLogs(page.value, pageSize.value, keyword.value.trim(), status.value, 0, receiverName.value.trim())
    rows.value = data.data.list
    total.value = data.data.total
}

function search() {
    if (searchTimer) clearTimeout(searchTimer)
    page.value = 1
    void load()
}

function reset() {
    keyword.value = ''
    receiverName.value = ''
    status.value = ''
}

watch([keyword, receiverName, status], () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
        page.value = 1
        void load()
    }, 300)
})

onBeforeUnmount(() => {
    if (searchTimer) clearTimeout(searchTimer)
})

onMounted(async () => {
    await nextTick()
    const canvas = document.createElement('canvas')
    const context = canvas.getContext('2d')
    if (!context) return
    toolbarRef.value?.querySelectorAll<HTMLInputElement>('.content-sized-filter input').forEach(input => {
        const inputStyle = getComputedStyle(input)
        context.font = inputStyle.font
        const textWidth = context.measureText(input.placeholder).width
        const wrapper = input.closest<HTMLElement>('.el-input__wrapper')
        const wrapperStyle = wrapper ? getComputedStyle(wrapper) : null
        const clearIcon = wrapper?.querySelector<HTMLElement>('.el-input__clear')
        const suffix = wrapper?.querySelector<HTMLElement>('.el-input__suffix')
        const clearIconStyle = clearIcon ? getComputedStyle(clearIcon) : null
        const suffixWidth = suffix?.getBoundingClientRect().width || (clearIconStyle
            ? (clearIcon?.getBoundingClientRect().width || Number.parseFloat(clearIconStyle.fontSize) || 14)
            + (Number.parseFloat(clearIconStyle.marginLeft) || 0)
            + (Number.parseFloat(clearIconStyle.marginRight) || 0)
            : 16)
        const wrapperSpace = wrapperStyle
            ? ['paddingLeft', 'paddingRight', 'borderLeftWidth', 'borderRightWidth']
                .reduce((sum, property) => sum + (Number.parseFloat(wrapperStyle[property as 'paddingLeft']) || 0), 0)
            : 24
        const width = Math.ceil(textWidth + wrapperSpace + suffixWidth)
        input.closest<HTMLElement>('.content-sized-filter')?.style.setProperty('--placeholder-width', `${width}px`)
    })
})

const columns: ColDef[] = [
    { headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node!.rowIndex! + 1, width: 70 },
    { field: 'alert_type', headerName: '告警类型' },
    {
        field: 'status',
        headerName: '发送状态',
        cellRenderer: (params: { value: string }) => {
            const labels: Record<string, string> = { sent: '发送成功', failed: '发送失败', pending: '待发送' }
            const tag = document.createElement('span')
            tag.className = `alert-status-tag is-${params.value || 'unknown'}`
            tag.textContent = labels[params.value] || params.value || '未记录'
            return tag
        },
    },
    { field: 'message', headerName: '内容', minWidth: 260, tooltipField: 'message' },
    { field: 'channel_type', headerName: '通道' },
    { field: 'receiver_names', headerName: '接收人', minWidth: 160, valueFormatter: params => params.value || '-', tooltipField: 'receiver_names' },
    { field: 'error_message', headerName: '失败原因', minWidth: 220, valueFormatter: params => params.value || '-', tooltipField: 'error_message' },
    { field: 'sent_at', headerName: '发送时间', minWidth: 160, valueFormatter: params => params.value ? String(params.value).replace('T', ' ').slice(0, 19) : '-' },
]

onMounted(load)
</script>
<template>
    <section class="page">
        <form ref="toolbarRef" class="toolbar" @submit.prevent="search">
            <el-input v-model="keyword" class="content-sized-filter" placeholder="搜索告警类型、内容、通道或错误信息" clearable />
            <el-input v-model="receiverName" class="content-sized-filter" placeholder="筛选接收人姓名" clearable />
            <el-select v-model="status" placeholder="全部状态" clearable>
                <el-option label="待发送" value="pending" />
                <el-option label="已发送" value="sent" />
                <el-option label="发送失败" value="failed" />
            </el-select>
            <el-button native-type="button" @click="reset">重置</el-button>
        </form>
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

.toolbar :deep(.content-sized-filter.el-input) {
    --el-input-width: var(--placeholder-width);
    flex: 0 0 var(--placeholder-width);
    width: var(--placeholder-width);
}

.toolbar :deep(.el-select) {
    flex: 0 0 120px;
    width: 120px;
}

.grid {
    flex: 1;
    min-height: 0
}

.grid :deep(.alert-status-tag) {
    display: inline-flex;
    align-items: center;
    min-height: 24px;
    padding: 0 10px;
    border: 1px solid;
    border-radius: 6px;
    font-size: 12px;
    line-height: 1;
}

.grid :deep(.alert-status-tag.is-sent) {
    border-color: #b7e4ca;
    background: #edf8f1;
    color: #24834b;
}

.grid :deep(.alert-status-tag.is-failed) {
    border-color: #f2b8b5;
    background: #fff1f0;
    color: #c83d36;
}

.grid :deep(.alert-status-tag.is-pending),
.grid :deep(.alert-status-tag.is-unknown) {
    border-color: #f0d59a;
    background: #fff8e8;
    color: #986b0c;
}
</style>
