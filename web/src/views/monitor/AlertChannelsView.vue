<script setup lang="ts">
import { h, onMounted, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import { ElButton, ElMessageBox, ElTag } from 'element-plus'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { deleteMonitorChannel, listMonitorChannels, saveMonitorChannel, type MonitorAlertChannel, type MonitorAlertChannelRequest } from '@/api/monitor'
const rows = ref<MonitorAlertChannel[]>([]), keyword = ref(''), dialog = ref(false), editing = ref<MonitorAlertChannel | null>(null), form = ref<MonitorAlertChannelRequest>({ name: '', channel_type: 'email', enabled: true, config: '' }); async function load() { const { data } = await listMonitorChannels(keyword.value); rows.value = data.data }; function open(row?: MonitorAlertChannel) { editing.value = row ?? null; form.value = row ? { ...row } : { name: '', channel_type: 'email', enabled: true, config: '' }; dialog.value = true }; async function save() { await saveMonitorChannel(editing.value ? { ...form.value, id: editing.value.id } : form.value); dialog.value = false; load() }; async function remove(row: MonitorAlertChannel) { await ElMessageBox.confirm('确定删除该告警通道？', '确认删除'); await deleteMonitorChannel(row.id); load() }; watch(keyword, load); onMounted(load)
const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => p.node!.rowIndex! + 1, width: 70 }, { field: 'name', headerName: '通道名称', minWidth: 150 }, { field: 'channel_type', headerName: '类型' }, { field: 'enabled', headerName: '状态', cellRenderer: (p: any) => h(ElTag, { type: p.value ? 'success' : 'info' }, () => p.value ? '启用' : '禁用') }, { field: 'config', headerName: '配置', minWidth: 220 }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => h('div', { class: 'actions' }, [h(ElButton, { class: 'edit', onClick: () => open(p.data) }, () => '编辑'), h(ElButton, { class: 'del', onClick: () => remove(p.data) }, () => '删除')]) }]
</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索通道" clearable /><el-button type="primary"
                @click="open()">新增通道</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination /><el-dialog v-model="dialog"
            :title="editing ? '编辑通道' : '新增通道'" width="480px"><el-form label-width="90px"><el-form-item label="名称"><el-input
                        v-model="form.name" /></el-form-item><el-form-item label="类型"><el-select
                        v-model="form.channel_type"><el-option label="邮件" value="email" /><el-option label="Webhook"
                            value="webhook" /></el-select></el-form-item><el-form-item label="配置"><el-input
                        v-model="form.config" type="textarea" /></el-form-item><el-form-item label="启用"><el-switch
                        v-model="form.enabled" /></el-form-item></el-form><template #footer><el-button
                    @click="dialog = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template></el-dialog>
    </section>
</template>
<style
    scoped>
    .page {
        height: 100%;
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-height: 0
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

    .actions {
        height: 100%;
        display: flex;
        align-items: center;
        gap: 8px
    }

    :deep(.edit),
    :deep(.del) {
        height: 28px;
        border-radius: 8px;
        padding: 4px 8px
    }

    :deep(.edit) {
        color: var(--primary);
        border-color: var(--primary)
    }

    :deep(.del) {
        color: var(--el-color-danger);
        border-color: var(--el-color-danger-light-5)
    }
</style>
