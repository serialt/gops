<script setup
    lang="ts">    import { onMounted, ref, watch } from 'vue'; import type { ColDef } from 'ag-grid-community'; import AppDataGrid from '@/components/AppDataGrid.vue'; import request from '@/utils/request'; interface Certificate { id: number; name: string; domain: string; source_type: string; ca_provider: string; status: string; auto_renew: boolean; renew_days_before: number; not_after: string }; const rows = ref<Certificate[]>([]), keyword = ref(''); async function load() { const { data } = await request.post<Certificate[]>('/ssl-cert/certificates/list', { keyword: keyword.value }); rows.value = data.data } watch(keyword, load); const columns: ColDef[] = [{ headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 }, { field: 'name', headerName: '证书名称', minWidth: 150 }, { field: 'domain', headerName: '域名', minWidth: 180 }, { field: 'ca_provider', headerName: 'CA 提供商' }, { field: 'status', headerName: '状态' }, { field: 'not_after', headerName: '到期时间', minWidth: 160 }]; onMounted(load)</script>
<template>
    <section class="page"><el-input v-model="keyword" placeholder="搜索证书或域名" clearable />
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
