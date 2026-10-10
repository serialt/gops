<script setup lang="ts">
import { computed, getCurrentInstance, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AgGridVue } from 'ag-grid-vue3'
import { AllCommunityModule, ModuleRegistry, themeQuartz, type ColDef, type RowSelectionOptions, type SelectionColumnDef } from 'ag-grid-community'
import { ArrowDown, Delete, Download, Refresh } from '@element-plus/icons-vue'

ModuleRegistry.registerModules([AllCommunityModule])

const props = defineProps<{
  rowData: unknown[]
  columnDefs: ColDef[]
  pagination?: boolean
  fillHeight?: boolean
  serverPagination?: boolean
  total?: number
  currentPage?: number
  pageSize?: number
  rowHeight?: number
  getRowId?: (row: unknown) => string
  bulkDelete?: (rows: unknown[]) => Promise<void>
  bulkCollect?: (rows: unknown[]) => Promise<void>
  bulkDownload?: (rows: unknown[]) => Promise<void>
  bulkDownloadLabel?: string
}>()
const emit = defineEmits<{
  selectionChanged: [rows: unknown[]]
  pageChange: [page: number]
  pageSizeChange: [pageSize: number]
}>()
const instance = getCurrentInstance()
const selectable = computed(() => Boolean(props.bulkDelete || props.bulkCollect || props.bulkDownload || instance?.vnode.props?.onSelectionChanged))

const defaultColDef: ColDef = {
  minWidth: 40,
  resizable: true,
  sortable: true,
  unSortIcon: true,
  suppressHeaderMenuButton: true,
  tooltipValueGetter: (params) => {
    const value = params.valueFormatted ?? params.value
    if (value == null) return ''
    return typeof value === 'string' ? value : String(value)
  },
}
const rowSelection: RowSelectionOptions = {
  mode: 'multiRow',
  checkboxes: true,
  headerCheckbox: true,
  selectAll: 'all',
  enableClickSelection: false,
}
const selectionColumnDef: SelectionColumnDef = {
  width: 36,
  minWidth: 36,
  maxWidth: 36,
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
const localCurrentPage = ref(1)
const localPageSize = ref(20)
const selectedRows = ref<unknown[]>([])
const collectingRows = ref(false)
const bulkOperationCount = computed(() => Number(Boolean(props.bulkCollect)) + Number(Boolean(props.bulkDownload)) + Number(Boolean(props.bulkDelete)))
const shellRef = ref<HTMLElement>()
const bulkActionTarget = ref<HTMLElement>()
let gridApi: any
let bulkActionHost: HTMLElement | undefined
let filterWidthObserver: MutationObserver | undefined
let filterWidthFrame = 0
const originalFilterWidths = new Map<HTMLElement, { width: string; flex: string }>()
const originalFilterMargins = new Map<HTMLElement, string>()
const pagedRows = computed(() => {
  if (!props.pagination || props.serverPagination) return props.rowData
  const start = (localCurrentPage.value - 1) * localPageSize.value
  return props.rowData.slice(start, start + localPageSize.value)
})
const domLayout = computed(() => props.pagination || props.fillHeight ? 'normal' : 'autoHeight')

const firstDataRow = computed<Record<string, unknown> | undefined>(() =>
  props.rowData.find((row): row is Record<string, unknown> => typeof row === 'object' && row !== null),
)

function hasTimestampField(field: 'created_at' | 'updated_at') {
  return Boolean(firstDataRow.value && field in firstDataRow.value)
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
    timestamps.push({ field: 'updated_at', headerName: '更新时间', minWidth: 150, valueFormatter: (params) => formatTimestamp(params.value) })
  }
  const actionIndex = columns.findIndex((column) => column.headerName === '操作')
  columns.splice(actionIndex < 0 ? columns.length : actionIndex, 0, ...timestamps)
  return columns
})

watch(() => props.rowData, () => {
  if (!props.serverPagination) localCurrentPage.value = 1
  selectedRows.value = []
  gridApi?.deselectAll()
})
const paginationPage = computed({ get: () => props.serverPagination ? (props.currentPage || 1) : localCurrentPage.value, set: value => { if (props.serverPagination) emit('pageChange', value); else localCurrentPage.value = value } })
const paginationSize = computed({ get: () => props.serverPagination ? (props.pageSize || 20) : localPageSize.value, set: value => { if (props.serverPagination) emit('pageSizeChange', value); else localPageSize.value = value } })
// 操作列保持固定宽度。仅其左侧最后一个数据列填充余量，其他列继续贴合实际内容。
const normalizedColumnDefs = computed<ColDef[]>(() => {
  const isActionColumn = (column: ColDef) => column.headerName === '操作'
  const isIndexColumn = (column: ColDef) => column.headerName === '序号'
  const fillIndex = columnDefsWithTimestamps.value.reduce(
    (lastIndex, column, index) => isActionColumn(column) ? lastIndex : index,
    -1,
  )
  return columnDefsWithTimestamps.value.map((column, index) => {
    if (isActionColumn(column)) return { ...column, sortable: false, suppressSizeToFit: true }
    if (isIndexColumn(column)) {
      return {
        ...column,
        sortable: false,
        headerClass: 'app-data-grid-index-header',
        cellClass: 'app-data-grid-index-cell',
      }
    }
    if (index === fillIndex) {
      return {
        ...column,
        width: undefined,
        flex: 1,
        minWidth: Math.min(column.minWidth ?? 40, 250),
        maxWidth: undefined,
        suppressSizeToFit: false,
        tooltipField: undefined,
        tooltipValueGetter: column.tooltipValueGetter ?? ((params) => {
          const value = params.valueFormatted ?? params.value
          if (value == null) return ''
          return typeof value === 'string' ? value : String(value)
        }),
      }
    }
    return {
      ...column,
      width: undefined,
      flex: undefined,
      minWidth: Math.min(column.minWidth ?? 40, 250),
      maxWidth: 250,
      suppressSizeToFit: true,
      tooltipField: undefined,
      tooltipValueGetter: column.tooltipValueGetter ?? ((params) => {
        const value = params.valueFormatted ?? params.value
        if (value == null) return ''
        return typeof value === 'string' ? value : String(value)
      }),
    }
  })
})

let resizeFrame = 0
function resizeColumns(event: any) {
  cancelAnimationFrame(resizeFrame)
  resizeFrame = requestAnimationFrame(() => {
    // Size content columns from the current page; the final data column flexes into remaining space.
    const columns = event.api.getColumns().filter((column: any) => {
      const definition = column.getColDef()
      return definition.headerName !== '操作' && !definition.flex && !column.isPinned()
    })
    event.api.autoSizeColumns(columns)
  })
}

function fitFilterInputWidths() {
  const shell = shellRef.value
  if (!shell) return
  const pageRoot = shell.closest<HTMLElement>('.page-container, .page, .section-card, .page-content') ?? shell.parentElement
  if (!pageRoot) return

  const canvas = document.createElement('canvas')
  const context = canvas.getContext('2d')
  if (!context) return

  for (const input of pageRoot.querySelectorAll<HTMLInputElement>('input[placeholder]')) {
    if (input.closest('.el-form, .el-form-item, .el-dialog, [role="dialog"]')) continue
    if (input.type === 'password' || input.type === 'hidden' || input.disabled) continue
    const inputRoot = input.closest<HTMLElement>('.el-input')
    const wrapper = input.closest<HTMLElement>('.el-input__wrapper')
    if (!inputRoot || !wrapper || !input.placeholder) continue

    const inputStyle = getComputedStyle(input)
    context.font = inputStyle.font
    const textWidth = context.measureText(input.placeholder).width
    const wrapperStyle = getComputedStyle(wrapper)
    const inputSpace = ['paddingLeft', 'paddingRight'].reduce(
      (total, property) => total + (Number.parseFloat(inputStyle[property as 'paddingLeft']) || 0),
      0,
    )
    const wrapperSpace = ['paddingLeft', 'paddingRight', 'borderLeftWidth', 'borderRightWidth'].reduce(
      (total, property) => total + (Number.parseFloat(wrapperStyle[property as 'paddingLeft']) || 0),
      0,
    )
    const prefixWidth = wrapper.querySelector<HTMLElement>('.el-input__prefix')?.getBoundingClientRect().width ?? 0
    const suffixWidth = wrapper.querySelector<HTMLElement>('.el-input__suffix')?.getBoundingClientRect().width ?? 0
    const width = `${Math.ceil(textWidth + inputSpace + wrapperSpace + prefixWidth + suffixWidth)}px`

    if (!originalFilterWidths.has(inputRoot)) {
      originalFilterWidths.set(inputRoot, {
        width: inputRoot.style.width,
        flex: inputRoot.style.flex,
      })
    }
    if (inputRoot.style.width !== width) inputRoot.style.setProperty('width', width, 'important')
    if (inputRoot.style.flex !== '0 0 auto') inputRoot.style.setProperty('flex', '0 0 auto', 'important')
  }

  const filterBar = shell.previousElementSibling as HTMLElement | null
  if (!filterBar?.querySelector('input[placeholder]')) return
  const parent = filterBar.parentElement
  if (!parent) return
  if (!originalFilterMargins.has(filterBar)) originalFilterMargins.set(filterBar, filterBar.style.marginBottom)
  const parentStyle = getComputedStyle(parent)
  const rowGap = Number.parseFloat(parentStyle.rowGap) || 0
  const originalMargin = Number.parseFloat(originalFilterMargins.get(filterBar) || '0') || 0
  const adjustedMargin = originalMargin + 2 - rowGap
  filterBar.style.setProperty('margin-bottom', `${adjustedMargin}px`, 'important')
}

function scheduleFilterWidthFit() {
  cancelAnimationFrame(filterWidthFrame)
  filterWidthFrame = requestAnimationFrame(fitFilterInputWidths)
}

onMounted(() => {
  scheduleFilterWidthFit()
  const pageRoot = shellRef.value?.closest<HTMLElement>('.page-container, .page, .section-card, .page-content')
  if (pageRoot) {
    filterWidthObserver = new MutationObserver(scheduleFilterWidthFit)
    filterWidthObserver.observe(pageRoot, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['placeholder'],
    })
  }
  if (!props.bulkDelete && !props.bulkCollect && !props.bulkDownload) return
  let container = shellRef.value?.parentElement
  let toolbar: HTMLElement | null = null
  while (container && container !== document.body) {
    toolbar = container.querySelector<HTMLElement>(':scope > .toolbar-actions, :scope > .header-actions, :scope > .node-header-actions, :scope > .page-toolbar, :scope > .toolbar, :scope > .el-dialog__header')
    if (toolbar) break
    container = container.parentElement
  }
  if (!toolbar) return
  const actionArea = toolbar.matches('.toolbar-actions, .header-actions, .node-header-actions, .el-dialog__header')
    ? toolbar
    : toolbar.querySelector<HTMLElement>(':scope > .toolbar-actions, :scope > .header-actions, :scope > .node-header-actions') || toolbar
  const explicitTarget = actionArea.querySelector<HTMLElement>(':scope > .app-data-grid-bulk-action-anchor')
  if (explicitTarget) {
    bulkActionTarget.value = explicitTarget
    return
  }
  bulkActionHost = document.createElement('span')
  bulkActionHost.className = 'app-data-grid-bulk-action-host'
  bulkActionHost.style.display = 'contents'
  const firstButton = actionArea.querySelector<HTMLElement>(':scope > .el-button')
  actionArea.insertBefore(bulkActionHost, firstButton)
  bulkActionTarget.value = bulkActionHost
})

onBeforeUnmount(() => {
  cancelAnimationFrame(resizeFrame)
  cancelAnimationFrame(filterWidthFrame)
  filterWidthObserver?.disconnect()
  for (const [element, original] of originalFilterWidths) {
    if (original.width) element.style.setProperty('width', original.width)
    else element.style.removeProperty('width')
    if (original.flex) element.style.setProperty('flex', original.flex)
    else element.style.removeProperty('flex')
  }
  originalFilterWidths.clear()
  for (const [element, marginBottom] of originalFilterMargins) {
    if (marginBottom) element.style.setProperty('margin-bottom', marginBottom)
    else element.style.removeProperty('margin-bottom')
  }
  originalFilterMargins.clear()
  bulkActionHost?.remove()
})

function resolveRowId(params: any) {
  return props.getRowId!(params.data)
}

function handleSelectionChanged(event: any) {
  selectedRows.value = event.api.getSelectedRows()
  emit('selectionChanged', selectedRows.value)
}

async function handleBulkDelete() {
  if (!props.bulkDelete || !selectedRows.value.length) return
  await props.bulkDelete([...selectedRows.value])
  gridApi?.deselectAll()
  selectedRows.value = []
  emit('selectionChanged', [])
}

async function handleBulkCollect() {
  if (!props.bulkCollect || !selectedRows.value.length) return
  collectingRows.value = true
  try {
    await props.bulkCollect([...selectedRows.value])
    gridApi?.deselectAll()
    selectedRows.value = []
    emit('selectionChanged', [])
  } finally {
    collectingRows.value = false
  }
}

async function handleBulkDownload() {
  if (!props.bulkDownload || !selectedRows.value.length) return
  collectingRows.value = true
  try {
    await props.bulkDownload([...selectedRows.value])
  } finally {
    collectingRows.value = false
  }
}

function handleBulkCommand(command: string) {
  if (command === 'collect') void handleBulkCollect()
  if (command === 'download') void handleBulkDownload()
  if (command === 'delete') void handleBulkDelete()
}

function handleGridReady(event: any) {
  gridApi = event.api
  resizeColumns(event)
}
</script>

<template>
  <div ref="shellRef" class="app-data-grid-shell">
    <Teleport v-if="bulkActionTarget" :to="bulkActionTarget">
      <template v-if="selectedRows.length && bulkOperationCount === 1">
        <el-button v-if="bulkCollect" class="app-data-grid-bulk-collect-button" type="primary" plain :icon="Refresh"
          :loading="collectingRows" @click="handleBulkCollect">{{ selectedRows.length > 1 ?
            `批量采集（${selectedRows.length}）` : '采集信息' }}</el-button>
        <el-button v-if="bulkDownload" class="app-data-grid-bulk-collect-button" type="primary" plain :icon="Download"
          :loading="collectingRows" @click="handleBulkDownload">{{ bulkDownloadLabel || (selectedRows.length > 1 ?
            `批量下载（${selectedRows.length}）` : '下载') }}</el-button>
        <el-button v-if="bulkDelete" class="app-data-grid-bulk-button" type="danger" plain @click="handleBulkDelete">{{
          selectedRows.length > 1 ? `批量删除（${selectedRows.length}）` : '删除' }}</el-button>
      </template>
      <el-dropdown v-else-if="selectedRows.length && bulkOperationCount > 1" trigger="click"
        popper-class="app-data-grid-bulk-popper" @command="handleBulkCommand">
        <el-button type="primary" plain>批量操作（{{ selectedRows.length }}）<el-icon class="el-icon--right">
            <ArrowDown />
          </el-icon></el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-if="bulkCollect" class="app-data-grid-bulk-menu-item" command="collect">
              <el-icon>
                <Refresh />
              </el-icon>批量采集
            </el-dropdown-item>
            <el-dropdown-item v-if="bulkDownload" class="app-data-grid-bulk-menu-item" command="download">
              <el-icon>
                <Download />
              </el-icon>{{ bulkDownloadLabel || '批量下载' }}
            </el-dropdown-item>
            <el-dropdown-item v-if="bulkDelete" class="app-data-grid-bulk-menu-item is-danger" command="delete">
              <el-icon>
                <Delete />
              </el-icon>批量删除
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </Teleport>
    <AgGridVue class="app-data-grid" :theme="themeQuartz" :row-data="pagedRows" :column-defs="normalizedColumnDefs"
      :default-col-def="defaultColDef" :row-selection="selectable ? rowSelection : undefined"
      :selection-column-def="selectable ? selectionColumnDef : undefined" :locale-text="localeText"
      :animate-rows="false" :get-row-id="getRowId ? resolveRowId : undefined" :row-height="rowHeight" :row-buffer="5"
      :dom-layout="domLayout" :enable-cell-text-selection="true" :ensure-dom-order="false" :tooltip-show-delay="200"
      tooltip-show-mode="whenTruncated" :tooltip-hide-delay="10000" @grid-ready="handleGridReady"
      @first-data-rendered="resizeColumns" @row-data-updated="resizeColumns"
      @selection-changed="handleSelectionChanged" />
    <div v-if="pagination" class="grid-pagination">
      <el-pagination v-model:current-page="paginationPage" v-model:page-size="paginationSize"
        :page-sizes="[20, 50, 100, 500, 1000]" :total="serverPagination ? (total || 0) : rowData.length"
        layout="total, sizes, prev, pager, next, jumper" />
    </div>
  </div>
</template>

<style scoped>
.app-data-grid-shell {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
  min-height: 0;
}

.app-data-grid :deep(.ag-header-cell-label) {
  gap: 0;
}

.app-data-grid :deep(.ag-sort-indicator-container) {
  flex: 0 0 auto;
  width: max-content;
  min-width: 0;
  padding: 0;
  margin: 0;
}

.app-data-grid :deep(.ag-sort-indicator-icon) {
  padding: 0;
}

.app-data-grid :deep(.ag-cell) {
  padding-inline: 4px;
}

.app-data-grid :deep(.ag-header-cell[col-id='ag-Grid-SelectionColumn']) {
  position: relative;
  padding-left: 0;
  padding-right: 0;
}

.app-data-grid :deep(.ag-header-cell[col-id='ag-Grid-SelectionColumn'] .ag-header-cell-comp-wrapper),
.app-data-grid :deep(.ag-header-cell[col-id='ag-Grid-SelectionColumn'] .ag-header-select-all),
.app-data-grid :deep(.app-data-grid-index-header .ag-header-cell-comp-wrapper),
.app-data-grid :deep(.app-data-grid-index-header .ag-header-cell-label) {
  display: flex;
  justify-content: center;
  width: 100%;
}

.app-data-grid :deep(.ag-header-cell[col-id='ag-Grid-SelectionColumn'] .ag-header-select-all) {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 !important;
  transform: none;
  cursor: pointer;
}

.app-data-grid :deep(.ag-cell[col-id='ag-Grid-SelectionColumn']) {
  position: relative;
  display: flex;
  justify-content: center;
  padding: 0;
}

.app-data-grid :deep(.ag-cell[col-id='ag-Grid-SelectionColumn'] .ag-cell-wrapper) {
  display: flex;
  justify-content: center;
  width: 100%;
}

.app-data-grid :deep(.ag-cell[col-id='ag-Grid-SelectionColumn'] .ag-cell-value) {
  flex: 0 0 0;
  width: 0;
  min-width: 0;
}

.app-data-grid :deep(.ag-cell[col-id='ag-Grid-SelectionColumn'] .ag-selection-checkbox) {
  position: static;
  margin: 0 !important;
  transform: none;
  cursor: pointer;
}

.app-data-grid :deep(.app-data-grid-index-cell),
.app-data-grid :deep(.app-data-grid-index-cell .ag-cell-wrapper),
.app-data-grid :deep(.app-data-grid-index-cell .ag-cell-value) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.app-data-grid :deep(.ag-cell-value) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-data-grid-bulk-button {
  border-radius: 8px;
  background: #fff3f1;
  font-weight: 700;
}

.app-data-grid-bulk-collect-button {
  border-radius: 8px;
  font-weight: 700;
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu) {
  display: flex;
  width: max-content;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  border-radius: 8px;
}

:global(.app-data-grid-bulk-popper.el-popper) {
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 8px;
}

:global(.app-data-grid-bulk-popper.el-popper .el-popper__arrow) {
  display: none;
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu__item) {
  height: auto;
  padding: 0;
  line-height: normal;
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu__item.app-data-grid-bulk-menu-item) {
  display: flex;
  width: 100%;
  box-sizing: border-box;
  min-height: 42px;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border: 1px solid var(--line) !important;
  border-radius: 8px !important;
  background: #fff;
  color: var(--text);
  font-size: 10px;
  line-height: 1.2;
}

:global(.app-data-grid-bulk-popper .app-data-grid-bulk-menu-item .el-icon) {
  font-size: 10px;
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu__item.app-data-grid-bulk-menu-item:not(.is-disabled):hover) {
  border-color: var(--primary);
  background: color-mix(in srgb, var(--primary) 6%, white);
  color: var(--primary);
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu__item.app-data-grid-bulk-menu-item.is-danger) {
  color: #d94848 !important;
}

:global(.app-data-grid-bulk-popper .el-dropdown-menu__item.app-data-grid-bulk-menu-item.is-danger:not(.is-disabled):hover) {
  border-color: #efb1b1 !important;
  background: #fff5f5;
  color: #d94848 !important;
}

.app-data-grid {
  flex: 1;
  width: 100%;
  min-height: 0;
  --ag-font-size: 11px;
  --ag-font-family: inherit;
  --ag-border-color: var(--line);
  --ag-header-background-color: #fafafa;
  --ag-row-border-color: var(--line);
  --ag-background-color: #fff;
  --ag-odd-row-background-color: #fafafa;
  --ag-header-foreground-color: var(--text);
  --ag-foreground-color: var(--text);
  --ag-row-hover-color: #f0f0f0;
}

.app-data-grid :deep(.ag-cell-focus),
.app-data-grid :deep(.ag-cell-range-selected) {
  border-radius: 6px;
}

.app-data-grid :deep(.ag-checkbox-input-wrapper),
.app-data-grid :deep(.ag-checkbox-input-wrapper::after) {
  border-radius: 5px;
}

.grid-pagination {
  display: flex;
  justify-content: flex-end;
  min-height: 30px;
  padding: 2px 0 0;
  overflow-x: auto;
}

.grid-pagination :deep(.el-pagination) {
  flex-wrap: nowrap;
  min-width: max-content;
  font-size: 10px;
  --el-pagination-font-size: 10px;
}

.grid-pagination :deep(.el-pagination .el-select .el-input__wrapper),
.grid-pagination :deep(.el-pagination .el-input__wrapper),
.grid-pagination :deep(.el-pagination .el-input__inner) {
  font-size: 10px;
}
</style>
