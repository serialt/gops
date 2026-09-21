<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { AgGridVue } from 'ag-grid-vue3'
import { AllCommunityModule, ModuleRegistry, themeQuartz, type ColDef, type RowSelectionOptions, type SelectionColumnDef } from 'ag-grid-community'

ModuleRegistry.registerModules([AllCommunityModule])

const props = defineProps<{
  rowData: unknown[]
  columnDefs: ColDef[]
  pagination?: boolean
  rowHeight?: number
}>()
const emit = defineEmits<{
  selectionChanged: [rows: unknown[]]
}>()

const defaultColDef: ColDef = {
  minWidth: 10,
  resizable: true,
  sortable: false,
  suppressHeaderMenuButton: true,
}
const rowSelection: RowSelectionOptions = {
  mode: 'multiRow',
  checkboxes: true,
  headerCheckbox: true,
  enableClickSelection: false,
}
const selectionColumnDef: SelectionColumnDef = {
  width: 42,
  minWidth: 42,
  maxWidth: 42,
  pinned: 'left',
  resizable: false,
  suppressHeaderMenuButton: true,
}
const localeText = {
  noRowsToShow: '暂无数据',
  page: '页',
  more: '更多',
  to: '至',
  of: '共',
  next: '下一页',
  last: '末页',
  first: '首页',
  previous: '上一页',
  loadingOoo: '加载中…',
  pageSizeSelectorLabel: '每页',
}
const currentPage = ref(1)
const pageSize = ref(20)
const pagedRows = computed(() => {
  if (!props.pagination) return props.rowData
  const start = (currentPage.value - 1) * pageSize.value
  return props.rowData.slice(start, start + pageSize.value)
})
const domLayout = computed(() => props.pagination ? 'normal' : 'autoHeight')

function hasTimestampField(field: 'created_at' | 'updated_at') {
  return props.rowData.some((row) => typeof row === 'object' && row !== null && field in row)
}

function formatTimestamp(value: unknown) {
  if (!value) return '-'
  const timestamp = String(value)
  return timestamp.replace('T', ' ').slice(0, 19)
}

// 所有列表在操作列前统一补齐时间字段，页面仅需提供业务列即可。
const columnDefsWithTimestamps = computed<ColDef[]>(() => {
  const columns = [...props.columnDefs]
  const fields = new Set(columns.map((column) => column.field))
  const timestamps: ColDef[] = []
  if (hasTimestampField('created_at') && !fields.has('created_at')) {
    timestamps.push({ field: 'created_at', headerName: '创建时间', minWidth: 150, valueFormatter: (params) => formatTimestamp(params.value) })
  }
  if (hasTimestampField('updated_at') && !fields.has('updated_at')) {
    timestamps.push({ field: 'updated_at', headerName: '编辑时间', minWidth: 150, valueFormatter: (params) => formatTimestamp(params.value) })
  }
  const actionIndex = columns.findIndex((column) => column.headerName === '操作')
  columns.splice(actionIndex < 0 ? columns.length : actionIndex, 0, ...timestamps)
  return columns
})

watch(() => props.rowData, () => { currentPage.value = 1 })
// 操作列保持固定宽度。仅其左侧最后一个数据列填充余量，其他列继续贴合实际内容。
const normalizedColumnDefs = computed<ColDef[]>(() => {
  const isActionColumn = (column: ColDef) => column.headerName === '操作'
  const fillIndex = columnDefsWithTimestamps.value.reduce((lastIndex, column, index) => isActionColumn(column) ? lastIndex : index, -1)
  return columnDefsWithTimestamps.value.map((column, index) => {
    if (index === fillIndex) {
      return { ...column, width: undefined, flex: undefined, minWidth: column.minWidth ?? 100, maxWidth: undefined, suppressSizeToFit: false }
    }
    if (isActionColumn(column) || column.width) return { ...column, suppressSizeToFit: true }
    return { ...column, minWidth: column.minWidth ?? 10, maxWidth: column.maxWidth ?? 320, suppressSizeToFit: true }
  })
})

function resizeColumns(event: any) {
  requestAnimationFrame(() => {
    // 先按内容压缩普通列；再只把余量分配给最后一个数据列。
    const columns = event.api.getColumns().filter((column: any) => !column.isPinned() && column.getColDef().suppressSizeToFit && !column.getColDef().width)
    event.api.autoSizeColumns(columns)
    event.api.sizeColumnsToFit()
  })
}

function fillRemainingSpace(event: any) {
  // 拖拽任意普通列后，由最后一个数据列接收余量，防止固定操作列前出现空区。
  if (event.finished !== false) event.api.sizeColumnsToFit()
}

function handleSelectionChanged(event: any) {
  emit('selectionChanged', event.api.getSelectedRows())
}
</script>

<template>
  <div class="app-data-grid-shell">
    <AgGridVue
      class="app-data-grid"
      :theme="themeQuartz"
      :row-data="pagedRows"
      :column-defs="normalizedColumnDefs"
      :default-col-def="defaultColDef"
      :row-selection="rowSelection"
      :selection-column-def="selectionColumnDef"
      :locale-text="localeText"
      :animate-rows="false"
      :row-height="rowHeight"
      :dom-layout="domLayout"
      :enable-cell-text-selection="true"
      :ensure-dom-order="true"
      :tooltip-show-delay="200"
      :tooltip-hide-delay="10000"
      @grid-ready="resizeColumns"
      @first-data-rendered="resizeColumns"
      @column-resized="fillRemainingSpace"
      @grid-size-changed="fillRemainingSpace"
      @selection-changed="handleSelectionChanged"
    />
    <div v-if="pagination" class="grid-pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100, 500, 1000]"
        :total="rowData.length"
        layout="total, sizes, prev, pager, next, jumper"
      />
    </div>
  </div>
</template>

<style scoped>
.app-data-grid-shell { display: flex; flex-direction: column; width: 100%; min-width: 0; min-height: 0; }
.app-data-grid { flex: 1; width: 100%; min-height: 0; --ag-font-size: 11px; --ag-font-family: inherit; --ag-border-color: var(--line); --ag-header-background-color: #fafafa; --ag-row-border-color: var(--line); --ag-background-color: #fff; --ag-odd-row-background-color: #fafafa; --ag-header-foreground-color: var(--text); --ag-foreground-color: var(--text); --ag-row-hover-color: #f0f0f0; }
.grid-pagination { display: flex; justify-content: flex-end; min-height: 30px; padding: 2px 0 0; overflow-x: auto; }
.grid-pagination :deep(.el-pagination) { flex-wrap: nowrap; min-width: max-content; font-size: 10px; --el-pagination-font-size: 10px; }
.grid-pagination :deep(.el-pagination .el-select .el-input__wrapper), .grid-pagination :deep(.el-pagination .el-input__wrapper), .grid-pagination :deep(.el-pagination .el-input__inner) { font-size: 10px; }
</style>
