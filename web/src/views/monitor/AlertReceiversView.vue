<script setup lang="ts">
import { h, onMounted, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import { ElButton, ElMessageBox, ElTag } from 'element-plus'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { deleteMonitorReceiver, listMonitorReceivers, saveMonitorReceiver, type MonitorAlertReceiver, type MonitorAlertReceiverRequest } from '@/api/monitor'
const rows = ref<MonitorAlertReceiver[]>([]), keyword = ref(''), dialog = ref(false), editing = ref<MonitorAlertReceiver | null>(null), form = ref<MonitorAlertReceiverRequest>({ name: '', email: '', phone: '', webhook: '', enabled: true }); async function load() { const { data } = await listMonitorReceivers(keyword.value); rows.value = data.data }; function open(row?: MonitorAlertReceiver) { editing.value = row ?? null; form.value = row ? { ...row } : { name: '', email: '', phone: '', webhook: '', enabled: true }; dialog.value = true }; async function save() { await saveMonitorReceiver(editing.value ? { ...form.value, id: editing.value.id } : form.value); dialog.value = false; load() }; async function remove(row: MonitorAlertReceiver) { await ElMessageBox.confirm('确定删除该接收人？', '确认删除'); await deleteMonitorReceiver(row.id); load() }; watch(keyword, load); onMounted(load)
const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => p.node!.rowIndex! + 1, width: 70 }, { field: 'name', headerName: '名称', minWidth: 140 }, { field: 'email', headerName: '邮箱', minWidth: 180 }, { field: 'phone', headerName: '手机号' }, { field: 'webhook', headerName: 'Webhook', minWidth: 180 }, { field: 'enabled', headerName: '状态', cellRenderer: (p: any) => h(ElTag, { type: p.value ? 'success' : 'info' }, () => p.value ? '启用' : '禁用') }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => h('div', { class: 'actions' }, [h(ElButton, { class: 'edit', onClick: () => open(p.data) }, () => '编辑'), h(ElButton, { class: 'del', onClick: () => remove(p.data) }, () => '删除')]) }]
</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索接收人" clearable /><el-button type="primary"
                @click="open()">新增接收人</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination /><el-dialog v-model="dialog"
            :title="editing ? '编辑接收人' : '新增接收人'" width="480px"><el-form label-width="90px"><el-form-item
                    label="名称"><el-input v-model="form.name" /></el-form-item><el-form-item label="邮箱"><el-input
                        v-model="form.email" /></el-form-item><el-form-item label="手机号"><el-input
                        v-model="form.phone" /></el-form-item><el-form-item label="Webhook"><el-input
                        v-model="form.webhook" /></el-form-item><el-form-item label="启用"><el-switch
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
