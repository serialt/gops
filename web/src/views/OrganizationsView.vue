<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createOrganization, deleteOrganization, getOrganizations, updateOrganization } from '@/api'
import type { Organization } from '@/types'

const keyword = ref('')
const selectedId = ref<number | null>(null)
const organizations = ref<Organization[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ parent_id: null as number | null, name: '', code: '', manager: '', status: true })

const treeData = computed(() => {
  const map = new Map<number, any>()
  const roots: any[] = []
  organizations.value.forEach(item => map.set(item.id, { id: item.id, label: item.name, children: [] }))
  organizations.value.forEach(item => {
    const node = map.get(item.id)
    if (item.parent_id) map.get(item.parent_id)?.children.push(node)
    else roots.push(node)
  })
  return roots
})

const selectedName = computed(() => organizations.value.find(item => item.id === selectedId.value)?.name || '全部组织')
const tableData = computed(() => organizations.value.filter(item => {
  const matchesNode = !selectedId.value || item.parent_id === selectedId.value || item.id === selectedId.value
  const term = keyword.value.trim().toLowerCase()
  return matchesNode && (!term || item.name.toLowerCase().includes(term) || item.code.toLowerCase().includes(term))
}))

function selectNode(node: { id: number }) {
  selectedId.value = node.id
}

async function load() { const { data } = await getOrganizations(); organizations.value = data.data }
function openCreate() { editingId.value = null; Object.assign(form, { parent_id: selectedId.value, name: '', code: '', manager: '', status: true }); dialogVisible.value = true }
function openEdit(row: Organization) { editingId.value = row.id; Object.assign(form, { parent_id: row.parent_id, name: row.name, code: row.code, manager: row.manager, status: row.status }); dialogVisible.value = true }
async function save() { if (!form.name || !form.code) return ElMessage.warning('请填写组织名称和编码'); if (editingId.value) await updateOrganization(editingId.value, form); else await createOrganization(form); ElMessage.success('保存成功'); dialogVisible.value = false; await load() }
async function remove(row: Organization) { await ElMessageBox.confirm(`确定删除组织 ${row.name}？`, '提示', { type: 'warning' }); await deleteOrganization(row.id); ElMessage.success('删除成功'); if (selectedId.value === row.id) selectedId.value = null; await load() }
onMounted(load)
</script>

<template>
  <div class="page-container organization-page">
    <div class="organization-layout">
      <aside class="tree-panel">
        <div class="panel-title"><span>组织架构</span><el-button v-permission="'POST /api/v1/organizations/create'" :icon="Plus" circle text title="新增组织" @click="openCreate" /></div>
        <el-tree :data="treeData" node-key="id" highlight-current default-expand-all :current-node-key="selectedId" :expand-on-click-node="false" @node-click="selectNode" />
      </aside>

      <section class="organization-content">
        <div class="section-card">
          <div class="page-toolbar">
            <div><h2 class="page-title">{{ selectedName }}</h2><p class="page-subtitle">查看和维护该组织下的部门信息</p></div>
            <el-button v-permission="'POST /api/v1/organizations/create'" type="primary" :icon="Plus" @click="openCreate">新增组织</el-button>
          </div>
          <div class="list-toolbar">
            <el-input v-model="keyword" placeholder="搜索组织名称或编码" clearable :prefix-icon="Search" />
          </div>
          <el-table :data="tableData" class="data-table">
            <el-table-column prop="name" label="组织名称" min-width="150" />
            <el-table-column prop="code" label="组织编码" min-width="120" />
            <el-table-column prop="manager" label="负责人" min-width="100" />
            <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status ? 'success' : 'info'">{{ row.status ? '正常' : '停用' }}</el-tag></template></el-table-column>
            <el-table-column prop="updated_at" label="更新时间" min-width="150" />
            <el-table-column label="操作" width="120" fixed="right"><template #default="{ row }"><el-button v-permission="'POST /api/v1/organizations/update'" link type="primary" @click="openEdit(row)">编辑</el-button><el-button v-permission="'POST /api/v1/organizations/delete'" link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
          </el-table>
        </div>
      </section>
    </div>
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑组织' : '新增组织'" width="480px">
      <el-form :model="form" label-width="82px"><el-form-item label="上级组织"><el-select v-model="form.parent_id" clearable style="width:100%"><el-option v-for="item in organizations.filter(item => item.id !== editingId)" :key="item.id" :label="item.name" :value="item.id" /></el-select></el-form-item><el-form-item label="组织名称" required><el-input v-model="form.name" /></el-form-item><el-form-item label="组织编码" required><el-input v-model="form.code" /></el-form-item><el-form-item label="负责人"><el-input v-model="form.manager" /></el-form-item><el-form-item label="状态"><el-switch v-model="form.status" /></el-form-item></el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.organization-layout { display: grid; grid-template-columns: 220px minmax(0, 1fr); gap: 16px; min-height: 520px; }
.tree-panel, .section-card { border: 1px solid var(--line); border-radius: 12px; background: var(--bg-card); box-shadow: var(--shadow); }
.tree-panel { padding: 12px 8px; }
.panel-title { display: flex; align-items: center; justify-content: space-between; padding: 2px 8px 10px; color: var(--text); font-size: 13px; font-weight: 700; border-bottom: 1px solid var(--line); }
.panel-title :deep(.el-button) { color: var(--primary); }
.tree-panel :deep(.el-tree) { --el-tree-node-hover-bg-color: var(--primary-soft); background: transparent; padding-top: 8px; font-size: 12px; }
.tree-panel :deep(.el-tree-node__content) { height: 34px; border-radius: 6px; margin: 1px 0; }
.tree-panel :deep(.is-current > .el-tree-node__content) { color: var(--primary); font-weight: 700; background: var(--primary-soft); }
.section-card { padding: 16px; min-width: 0; }
.page-toolbar { margin-bottom: 14px; }
.page-subtitle { margin: 5px 0 0; color: var(--text-soft); font-size: 11px; }
.list-toolbar { display: flex; justify-content: flex-end; margin-bottom: 14px; }
.list-toolbar :deep(.el-input) { width: 230px; }
.data-table { width: 100%; }
@media (max-width: 900px) { .organization-layout { grid-template-columns: 1fr; } .tree-panel { max-height: 220px; overflow: auto; } }
</style>
