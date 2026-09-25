<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import { ElMessage, ElMessageBox } from 'element-plus'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { deleteNginxSource, listNginxSources, saveNginxSource, type NginxSource } from '@/api/nginx'
const rows = ref<NginxSource[]>([]), keyword = ref(''), dialog = ref(false), form = reactive({ id: 0, name: '', source_type: 'host', description: '', status: 1, host_id: 0, log_path: '/var/log/nginx/access.log', log_format: 'combined', collect_interval: 60, retention_days: 30 }); async function load() { const { data } = await listNginxSources(keyword.value); rows.value = data.data }; function open(x?: NginxSource) { Object.assign(form, x ?? { id: 0, name: '', source_type: 'host', description: '', status: 1, host_id: 0, log_path: '/var/log/nginx/access.log', log_format: 'combined', collect_interval: 60, retention_days: 30 }); dialog.value = true }; async function save() { await saveNginxSource(form); dialog.value = false; load() }; async function remove(x: NginxSource) { await ElMessageBox.confirm('确定删除该数据源？', '确认删除'); await deleteNginxSource(x.id); load() }; async function batchRemove(selectedRows: unknown[]) { const selected = selectedRows as NginxSource[]; if (!selected.length) return; await ElMessageBox.confirm(`确定删除选中的 ${selected.length} 个数据源？`, '批量删除确认', { type: 'warning' }); await Promise.all(selected.map(row => deleteNginxSource(row.id))); ElMessage.success('批量删除成功'); await load() }; watch(keyword, load); const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '名称', minWidth: 140 }, { field: 'source_type', headerName: '类型' }, { field: 'log_path', headerName: '日志路径', minWidth: 220 }, { field: 'collect_interval', headerName: '采集间隔', valueFormatter: (p: any) => `${p.value} 秒` }, { field: 'status', headerName: '状态', valueFormatter: (p: any) => p.value ? '启用' : '禁用' }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const e = document.createElement('button'); e.className = 'asset-action edit'; e.textContent = '编辑'; e.onclick = () => open(p.data); const d = document.createElement('button'); d.className = 'asset-action delete'; d.textContent = '删除'; d.onclick = () => remove(p.data); const w = document.createElement('div'); w.className = 'asset-actions'; w.append(e, d); return w } }]); onMounted(load)
</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索数据源" clearable /><el-button type="primary"
                @click="open()">新增数据源</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination :bulk-delete="batchRemove" />
        <el-dialog v-model="dialog" title="Nginx 数据源" width="520px"><el-form label-width="90px"><el-form-item
                    label="名称"><el-input v-model="form.name" /></el-form-item><el-form-item label="日志路径"><el-input
                        v-model="form.log_path" /></el-form-item><el-form-item label="采集间隔"><el-input-number
                        v-model="form.collect_interval" :min="10" /> 秒</el-form-item><el-form-item
                    label="保留天数"><el-input-number v-model="form.retention_days" :min="1" /></el-form-item><el-form-item
                    label="启用"><el-switch v-model="form.status" :active-value="1"
                        :inactive-value="0" /></el-form-item></el-form><template #footer><el-button
                    @click="dialog = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template></el-dialog>
    </section>
</template>
<style scoped>
.page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px
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

:deep(.asset-actions) {
    height: 100%;
    display: flex;
    align-items: center;
    gap: 8px
}

:deep(.asset-action) {
    padding: 4px 9px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: #fff;
    font-size: 11px
}

:deep(.edit) {
    border-color: var(--primary);
    color: var(--primary)
}

:deep(.delete) {
    border-color: #ffd5d5;
    color: #ef6262
}
</style>
