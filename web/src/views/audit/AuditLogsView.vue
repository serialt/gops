<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import type { ColDef } from 'ag-grid-community';
import AppDataGrid from '@/components/AppDataGrid.vue';
import { listAuditData, listAuditLogins, listAuditOperations } from '@/api/audit';
const route = useRoute(), keyword = ref(''), rows = ref<any[]>([]); const mode = computed(() => String(route.meta.audit || 'operation')); async function load() { const r = mode.value === 'login' ? await listAuditLogins(keyword.value) : mode.value === 'data' ? await listAuditData(keyword.value) : await listAuditOperations(keyword.value); rows.value = r.data.data } watch([keyword, mode], load, { immediate: true }); const columns = computed<ColDef[]>(() => mode.value === 'login' ? [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'username', headerName: '用户名' }, { field: 'login_type', headerName: '登录类型' }, { field: 'status', headerName: '状态' }, { field: 'ip', headerName: 'IP 地址' }, { field: 'message', headerName: '说明', minWidth: 180 }, { field: 'created_at', headerName: '时间', minWidth: 160 }] : mode.value === 'data' ? [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'username', headerName: '用户名' }, { field: 'table_name', headerName: '数据表' }, { field: 'record_id', headerName: '记录 ID' }, { field: 'action', headerName: '操作' }, { field: 'created_at', headerName: '时间', minWidth: 160 }] : [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'username', headerName: '用户名' }, { field: 'module', headerName: '模块' }, { field: 'action', headerName: '操作' }, { field: 'description', headerName: '说明', minWidth: 180 }, { field: 'status', headerName: '状态码' }, { field: 'ip', headerName: 'IP' }, { field: 'created_at', headerName: '时间', minWidth: 160 }])
</script>
<template>
    <section class="page"><el-input v-model="keyword" placeholder="搜索审计日志" clearable />
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination />
    </section>
</template>
<style scoped>
.page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px
}

.page>.el-input {
    width: 260px
}

.grid {
    flex: 1;
    min-height: 0
}
</style>
