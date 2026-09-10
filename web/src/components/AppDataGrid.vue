<script setup lang="ts">
import { computed } from 'vue'
import { AgGridVue } from 'ag-grid-vue3'
import { AllCommunityModule, ModuleRegistry, themeQuartz, type ColDef } from 'ag-grid-community'

ModuleRegistry.registerModules([AllCommunityModule])

const props = defineProps<{
  rowData: unknown[]
  columnDefs: ColDef[]
}>()

const defaultColDef: ColDef = {
  resizable: true,
  sortable: false,
  suppressHeaderMenuButton: true,
}
// 操作列保持固定宽度。仅其左侧最后一个数据列填充余量，其他列继续贴合实际内容。
const normalizedColumnDefs = computed<ColDef[]>(() => {
  const isActionColumn = (column: ColDef) => column.headerName === '操作'
  const fillIndex = props.columnDefs.reduce((lastIndex, column, index) => isActionColumn(column) ? lastIndex : index, -1)
  return props.columnDefs.map((column, index) => index === fillIndex
    ? { ...column, width: undefined, flex: undefined, minWidth: column.minWidth ?? 100, suppressSizeToFit: false }
    : { ...column, pinned: isActionColumn(column) ? undefined : column.pinned, suppressSizeToFit: true },
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
</script>

<template>
  <AgGridVue
    class="app-data-grid"
    :theme="themeQuartz"
    :row-data="rowData"
    :column-defs="normalizedColumnDefs"
    :default-col-def="defaultColDef"
    :animate-rows="false"
    :dom-layout="'autoHeight'"
    @grid-ready="resizeColumns"
    @first-data-rendered="resizeColumns"
    @column-resized="fillRemainingSpace"
    @grid-size-changed="fillRemainingSpace"
  />
</template>

<style scoped>
.app-data-grid { width: 100%; --ag-font-size: 11px; --ag-font-family: inherit; --ag-border-color: var(--line); --ag-header-background-color: #fafafa; --ag-row-border-color: var(--line); --ag-background-color: #fff; --ag-odd-row-background-color: #fafafa; --ag-header-foreground-color: var(--text); --ag-foreground-color: var(--text); --ag-row-hover-color: #f0f0f0; }
</style>
