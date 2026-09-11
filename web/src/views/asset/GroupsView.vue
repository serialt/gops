<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createAssetGroup, deleteAssetGroup, listAssetGroups, updateAssetGroup, type AssetGroup } from '@/api/asset'

const rows = ref<AssetGroup[]>([])
const keyword = ref('')
const dialogVisible = ref(false)
const editingID = ref<number | null>(null)
const form = reactive({ name: '', code: '', parent_id: 0, description: '', sort_order: 0, status: true })

function actionButton(text: string, kind: string, onClick: () => void) {
  const button = document.createElement('button')
  button.className = `asset-action ${kind}`
  button.textContent = text
  button.onclick = onClick
  return button
}
const columnDefs = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: params => String((params.node?.rowIndex ?? 0) + 1), width: 70 },
  { field: 'name', headerName: '分组名称', minWidth: 120 },
  { field: 'code', headerName: '分组编码', minWidth: 120 },
  { field: 'parent_id', headerName: '上级分组', valueFormatter: params => params.value ? String(params.value) : '顶级分组', minWidth: 110 },
  { field: 'description', headerName: '说明', minWidth: 160 },
  { field: 'status', headerName: '状态', width: 90, cellRenderer: (params: any) => `<span class="grid-status${params.value ? '' : ' is-disabled'}">${params.value ? '启用' : '禁用'}</span>` },
  { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (params: any) => { const wrap = document.createElement('div'); wrap.className = 'asset-actions'; wrap.append(actionButton('编辑', 'edit', () => openEdit(params.data)), actionButton('删除', 'delete', () => remove(params.data))); return wrap } },
])

async function load() { const { data } = await listAssetGroups(keyword.value); rows.value = data.data }
function openCreate() { editingID.value = null; Object.assign(form, { name: '', code: '', parent_id: 0, description: '', sort_order: 0, status: true }); dialogVisible.value = true }
function openEdit(row: AssetGroup) { editingID.value = row.id; Object.assign(form, row); dialogVisible.value = true }
async function save() { if (!form.name || !form.code) return ElMessage.warning('请填写分组名称和编码'); if (editingID.value) await updateAssetGroup({ ...form, id: editingID.value }); else await createAssetGroup(form); dialogVisible.value = false; ElMessage.success('保存成功'); await load() }
async function remove(row: AssetGroup) { await ElMessageBox.confirm(`确定删除资产分组 ${row.name}？`, '删除确认', { type: 'warning' }); await deleteAssetGroup(row.id); ElMessage.success('删除成功'); await load() }
onMounted(load)
</script>

<template>
  <div class="page-container table-shell">
    <div class="section-card">
      <div class="page-toolbar"><h2 class="page-title">资产分组</h2><el-button v-permission="'POST /api/v1/assets/groups/create'" type="primary" class="primary-btn" @click="openCreate">新增分组</el-button></div>
      <el-input v-model="keyword" class="asset-search" placeholder="搜索分组名称或编码" clearable @input="load" />
      <AppDataGrid :row-data="rows" :column-defs="columnDefs" pagination />
    </div>
    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑资产分组' : '新增资产分组'" width="460px">
      <el-form :model="form" label-width="80px"><el-form-item label="名称"><el-input v-model="form.name" /></el-form-item><el-form-item label="编码"><el-input v-model="form.code" /></el-form-item><el-form-item label="上级 ID"><el-input-number v-model="form.parent_id" :min="0" /></el-form-item><el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item><el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item><el-form-item label="说明"><el-input v-model="form.description" type="textarea" /></el-form-item></el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.asset-search { width: 260px; margin-bottom: 12px; }
:deep(.asset-actions) { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; gap: 8px; }
:deep(.asset-action) { padding: 4px 9px; border: 1px solid transparent; border-radius: 8px; background: #fff; font-size: 11px; cursor: pointer; box-shadow: 0 1px 3px rgba(24, 34, 53, .08); }
:deep(.asset-action.edit) { border-color: var(--primary); color: var(--primary); background: #f7f8ff; }
:deep(.asset-action.delete) { border-color: #ffd5d5; color: #ef6262; background: #fff7f7; }
:deep(.grid-status) { display: inline-flex; padding: 3px 8px; border: 1px solid #d9f0d0; border-radius: 8px; color: #67c23a; background: #f0f9eb; }
:deep(.grid-status.is-disabled) { border-color: #e4e7ed; color: #909399; background: #f4f4f5; }
</style>
