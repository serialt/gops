<script setup
    lang="ts">    import { computed, onMounted, reactive, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import { ElMessageBox } from 'element-plus'; import AppDataGrid from '@/components/AppDataGrid.vue'; import { deleteKubernetesCluster, listKubernetesClusters, saveKubernetesCluster, type KubernetesCluster } from '@/api/kubernetes'; const rows = ref<KubernetesCluster[]>([]), keyword = ref(''), dialog = ref(false), form = reactive({ id: 0, name: '', alias: '', api_endpoint: '', kube_config: '', region: '', provider: '', description: '', status: 1 }); async function load() { const { data } = await listKubernetesClusters(keyword.value); rows.value = data.data } function open(x?: KubernetesCluster) { Object.assign(form, x ? { ...x, kube_config: '' } : { id: 0, name: '', alias: '', api_endpoint: '', kube_config: '', region: '', provider: '', description: '', status: 1 }); dialog.value = true } async function save() { await saveKubernetesCluster(form); dialog.value = false; load() } async function remove(x: KubernetesCluster) { await ElMessageBox.confirm('确定删除该集群？', '确认删除'); await deleteKubernetesCluster(x.id); load() } watch(keyword, load); const columns = computed<ColDef[]>(() => [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '集群名称', minWidth: 150 }, { field: 'alias', headerName: '别名' }, { field: 'api_endpoint', headerName: 'API 地址', minWidth: 220 }, { field: 'version', headerName: '版本' }, { field: 'node_count', headerName: '节点数' }, { field: 'pod_count', headerName: 'Pod 数' }, { field: 'status', headerName: '状态', valueFormatter: (p: any) => p.value ? '正常' : '未连接' }, { headerName: '操作', pinned: 'right', width: 150, cellRenderer: (p: any) => { const e = document.createElement('button'); e.className = 'btn'; e.textContent = '编辑'; e.onclick = () => open(p.data); const d = document.createElement('button'); d.className = 'btn danger'; d.textContent = '删除'; d.onclick = () => remove(p.data); const w = document.createElement('div'); w.className = 'actions'; w.append(e, d); return w } }]); onMounted(load)</script>
<template>
    <section class="page">
        <div class="toolbar"><el-input v-model="keyword" placeholder="搜索集群" clearable /><el-button type="primary"
                @click="open()">新增集群</el-button></div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination /><el-dialog v-model="dialog"
            title="Kubernetes 集群" width="540px"><el-form label-width="90px"><el-form-item label="名称"><el-input
                        v-model="form.name" /></el-form-item><el-form-item label="别名"><el-input
                        v-model="form.alias" /></el-form-item><el-form-item label="API 地址"><el-input
                        v-model="form.api_endpoint" /></el-form-item><el-form-item label="KubeConfig"><el-input
                        v-model="form.kube_config" type="textarea" /></el-form-item><el-form-item label="区域"><el-input
                        v-model="form.region" /></el-form-item><el-form-item label="启用"><el-switch v-model="form.status"
                        :active-value="1" :inactive-value="0" /></el-form-item></el-form><template #footer><el-button
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

    :deep(.actions) {
        height: 100%;
        display: flex;
        align-items: center;
        gap: 8px
    }

    :deep(.btn) {
        border: 1px solid var(--primary);
        border-radius: 8px;
        background: #fff;
        color: var(--primary);
        padding: 4px 9px
    }

    :deep(.danger) {
        border-color: #ffd5d5;
        color: #ef6262
    }
</style>
