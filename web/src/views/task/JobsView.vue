<script setup
    lang="ts">    import { onMounted, reactive, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { listTaskJobs, saveTaskJob, type TaskJob } from '@/api/task'; const rows = ref<TaskJob[]>([]), keyword = ref(''), dialog = ref(false), form = reactive({ id: 0, name: '', template_id: 0, task_type: 'manual', target_hosts: '', parameters: '' }); async function load() { const { data } = await listTaskJobs(keyword.value); rows.value = data.data } function open(x?: TaskJob) { Object.assign(form, x ?? { id: 0, name: '', template_id: 0, task_type: 'manual', target_hosts: '', parameters: '' }); dialog.value = true } async function save() { await saveTaskJob(form); dialog.value = false; load() } watch(keyword, load); const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '任务名称', minWidth: 180 }, { field: 'task_type', headerName: '类型' }, { field: 'status', headerName: '状态' }, { field: 'target_hosts', headerName: '目标主机', minWidth: 220 }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const b = document.createElement('button'); b.className = 'btn'; b.textContent = '编辑'; b.onclick = () => open(p.data); return b } }]; onMounted(load)</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索任务" clearable /><el-button type="primary"
                @click="open()">新建任务</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination /><el-dialog v-model="dialog"
            title="任务作业" width="480px"><el-form label-width="90px"><el-form-item label="名称"><el-input
                        v-model="form.name" /></el-form-item><el-form-item label="目标主机"><el-input
                        v-model="form.target_hosts" placeholder="主机 ID，多个以逗号分隔" /></el-form-item><el-form-item
                    label="参数"><el-input v-model="form.parameters" type="textarea" /></el-form-item></el-form><template
                #footer><el-button @click="dialog = false">取消</el-button><el-button type="primary"
                    @click="save">保存</el-button></template></el-dialog>
    </section>
</template>
<style
    scoped>
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

    :deep(.btn) {
        border: 1px solid var(--primary);
        border-radius: 8px;
        background: #fff;
        color: var(--primary);
        padding: 4px 9px
    }
</style>
