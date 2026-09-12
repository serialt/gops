<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { listKubernetesClusters, listKubernetesNamespaces, listKubernetesNodes, listKubernetesWorkloads, type KubernetesCluster } from '@/api/kubernetes'
const route = useRoute(); const kind = computed(() => String(route.meta.kubernetesResource || 'nodes')); const title = computed(() => kind.value === 'nodes' ? '节点管理' : kind.value === 'namespaces' ? '命名空间' : '工作负载'); const clusters = ref<KubernetesCluster[]>([]), clusterID = ref(0), keyword = ref(''), rows = ref<any[]>([])
async function loadClusters() { const { data } = await listKubernetesClusters(); clusters.value = data.data; if (!clusterID.value && clusters.value[0]) clusterID.value = clusters.value[0].id }
async function load() { if (!clusterID.value) { rows.value = []; return }; const r = kind.value === 'nodes' ? await listKubernetesNodes(clusterID.value, keyword.value) : kind.value === 'namespaces' ? await listKubernetesNamespaces(clusterID.value, keyword.value) : await listKubernetesWorkloads(clusterID.value, keyword.value); rows.value = r.data.data }
watch([clusterID, keyword, kind], load); loadClusters().then(load)
const columns = computed<ColDef[]>(() => kind.value === 'nodes' ? [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '节点名称', minWidth: 180 }, { field: 'status', headerName: '状态' }, { field: 'roles', headerName: '角色' }, { field: 'internal_ip', headerName: '内部 IP' }, { field: 'version', headerName: 'Kubelet 版本', minWidth: 140 }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }] : kind.value === 'namespaces' ? [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '命名空间', minWidth: 180 }, { field: 'status', headerName: '状态' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }] : [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'namespace', headerName: '命名空间' }, { field: 'name', headerName: '工作负载名称', minWidth: 180 }, { field: 'kind', headerName: '类型' }, { field: 'ready', headerName: '就绪副本' }, { field: 'status', headerName: '状态' }, { field: 'created_at', headerName: '创建时间', minWidth: 160 }])
</script>
<template>
    <section class="page">
        <div class="toolbar">
            <div class="filters"><el-select v-model="clusterID" placeholder="选择集群"><el-option v-for="item in clusters"
                        :key="item.id" :label="item.alias || item.name" :value="item.id" /></el-select><el-input
                    v-model="keyword" :placeholder="`搜索${title}`" clearable /></div><strong>{{ title }}</strong>
        </div>
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination />
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
        justify-content: space-between;
        align-items: center
    }

    .filters {
        display: flex;
        gap: 8px
    }

    .filters .el-select {
        width: 180px
    }

    .filters .el-input {
        width: 220px
    }

    .grid {
        flex: 1;
        min-height: 0
    }
</style>
