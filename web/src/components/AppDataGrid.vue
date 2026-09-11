<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { AgGridVue } from 'ag-grid-vue3'
import { AllCommunityModule, ModuleRegistry, themeQuartz, type ColDef, type RowSelectionOptions, type SelectionColumnDef } from 'ag-grid-community'

ModuleRegistry.registerModules([AllCommunityModule])

const props = defineProps<{
  rowData: unknown[]
  columnDefs: ColDef[]
  pagination?: boolean
}>()
const emit = defineEmits<{
  selectionChanged: [rows: unknown[]]
}>()

const defaultColDef: ColDef = {
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
const pageSize = ref(10)
const pagedRows = computed(() => {
  if (!props.pagination) return props.rowData
  const start = (currentPage.value - 1) * pageSize.value
  return props.rowData.slice(start, start + pageSize.value)
})
const domLayout = computed(() => props.pagination ? 'normal' : 'autoHeight')

watch(() => props.rowData, () => { currentPage.value = 1 })
// 操作列保持固定宽度。仅其左侧最后一个数据列填充余量，其他列继续贴合实际内容。
const normalizedColumnDefs = computed<ColDef[]>(() => {
  const isActionColumn = (column: ColDef) => column.headerName === '操作'
  const fillIndex = props.columnDefs.reduce((lastIndex, column, index) => isActionColumn(column) ? lastIndex : index, -1)
  return props.columnDefs.map((column, index) => index === fillIndex
    ? { ...column, width: undefined, flex: undefined, minWidth: column.minWidth ?? 100, suppressSizeToFit: false }
    : { ...column, suppressSizeToFit: true },
  )
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
      :dom-layout="domLayout"
      :enable-cell-text-selection="true"
      :ensure-dom-order="true"
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
        :page-sizes="[10, 20, 50, 100]"
        :total="rowData.length"
        layout="total, sizes, prev, pager, next, jumper"
      />
    </div>
  </div>
</template>

<style scoped>
.app-data-grid-shell { display: flex; flex-direction: column; width: 100%; min-width: 0; min-height: 0; }
.app-data-grid { flex: 1; width: 100%; min-height: 0; --ag-font-size: 11px; --ag-font-family: inherit; --ag-border-color: var(--line); --ag-header-background-color: #fafafa; --ag-row-border-color: var(--line); --ag-background-color: #fff; --ag-odd-row-background-color: #fafafa; --ag-header-foreground-color: var(--text); --ag-foreground-color: var(--text); --ag-row-hover-color: #f0f0f0; }
.grid-pagination { display: flex; justify-content: flex-end; min-height: 48px; padding: 12px 0 0; overflow-x: auto; }
.grid-pagination :deep(.el-pagination) { flex-wrap: nowrap; min-width: max-content; font-size: 10px; --el-pagination-font-size: 10px; }
.grid-pagination :deep(.el-pagination .el-select .el-input__wrapper), .grid-pagination :deep(.el-pagination .el-input__wrapper), .grid-pagination :deep(.el-pagination .el-input__inner) { font-size: 10px; }
</style>
