<script setup lang="ts">
import { ref, watch } from 'vue';
import type { ColDef } from 'ag-grid-community';
import AppDataGrid from '@/components/AppDataGrid.vue';
import { listAuditData } from '@/api/audit';
const keyword = ref(''), rows = ref<any[]>([]), page = ref(1), pageSize = ref(20), total = ref(0); async function load() { const r = await listAuditData(keyword.value, page.value, pageSize.value); rows.value = r.data.data.list; total.value = r.data.data.total } watch(keyword, () => { page.value = 1; void load() }, { immediate: true }); const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => (page.value - 1) * pageSize.value + p.node.rowIndex + 1, width: 70 }, { field: 'username', headerName: '用户名' }, { field: 'table_name', headerName: '数据表' }, { field: 'record_id', headerName: '记录 ID' }, { field: 'action', headerName: '操作' }, { field: 'created_at', headerName: '时间', minWidth: 160 }]
</script>
<template>
    <section class="page"><el-input v-model="keyword" placeholder="搜索审计日志" clearable />
        <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total"
            :current-page="page" :page-size="pageSize" @page-change="value => { page = value; void load() }"
            @page-size-change="value => { pageSize = value; page = 1; void load() }" />
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
