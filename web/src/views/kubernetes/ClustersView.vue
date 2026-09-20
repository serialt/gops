<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { ColDef } from 'ag-grid-community'
import type { FormInstance, FormRules, UploadFile } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { deleteKubernetesCluster, getKubernetesClusterConfig, listKubernetesClusters, saveKubernetesCluster, syncKubernetesClusters, testKubernetesCluster, type KubernetesCluster } from '@/api/kubernetes'

type AuthMode = 'kubeconfig' | 'serviceAccount'
const rows = ref<KubernetesCluster[]>([]), allRows = ref<KubernetesCluster[]>([]), keyword = ref(''), dialog = ref(false), saving = ref(false), syncing = ref(false)
const router = useRouter(), credentialDialog = ref(false), credentialCluster = ref(''), credentialContent = ref('')
const formRef = ref<FormInstance>()
const form = reactive({ id: 0, name: '', alias: '', authMode: 'kubeconfig' as AuthMode, api_endpoint: '', kube_config: '', token: '', ca_data: '', region: '', provider: 'self', description: '', status: 1 })
const rules: FormRules = {
  name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  kube_config: [{ validator: (_r, _v, done) => form.authMode === 'kubeconfig' && !form.kube_config.trim() && !form.id ? done(new Error('请上传或粘贴 KubeConfig')) : done(), trigger: 'blur' }],
  api_endpoint: [{ validator: (_r, _v, done) => form.authMode === 'serviceAccount' && !form.api_endpoint.trim() ? done(new Error('请输入 API Server 地址')) : done(), trigger: 'blur' }],
  token: [{ validator: (_r, _v, done) => form.authMode === 'serviceAccount' && !form.token.trim() ? done(new Error('请输入 Service Account Token')) : done(), trigger: 'blur' }],
}
async function load() { if (!keyword.value) { const { data } = await listKubernetesClusters(); rows.value = data.data; allRows.value = data.data; return } const [filtered, all] = await Promise.all([listKubernetesClusters(keyword.value), listKubernetesClusters()]); rows.value = filtered.data.data; allRows.value = all.data.data }
function reset() { Object.assign(form, { id: 0, name: '', alias: '', authMode: 'kubeconfig', api_endpoint: '', kube_config: '', token: '', ca_data: '', region: '', provider: 'self', description: '', status: 1 }) }
function open(row?: KubernetesCluster) { reset(); if (row) Object.assign(form, { ...row, authMode: 'kubeconfig', kube_config: '' }); dialog.value = true }
async function readConfig(file: UploadFile) { if (!file.raw) return; form.kube_config = await file.raw.text(); ElMessage.success('KubeConfig 文件已读取') }
function serviceAccountKubeConfig() { const ca = form.ca_data.trim(); return `apiVersion: v1\nkind: Config\nclusters:\n- name: registered-cluster\n  cluster:\n    server: ${form.api_endpoint.trim()}\n${ca ? `    certificate-authority-data: ${ca}\n` : '    insecure-skip-tls-verify: true\n'}contexts:\n- name: registered-context\n  context:\n    cluster: registered-cluster\n    user: service-account\ncurrent-context: registered-context\nusers:\n- name: service-account\n  user:\n    token: ${form.token.trim()}\n` }
async function save() { if (!await formRef.value?.validate().catch(() => false)) return; saving.value = true; try { await saveKubernetesCluster({ id: form.id, name: form.name.trim(), alias: form.alias.trim(), api_endpoint: form.api_endpoint.trim(), kube_config: form.authMode === 'serviceAccount' ? serviceAccountKubeConfig() : form.kube_config, region: form.region.trim(), provider: form.provider, description: form.description.trim(), status: form.status }); dialog.value = false; ElMessage.success(form.id ? '集群更新成功' : '集群注册成功'); await load() } finally { saving.value = false } }
async function test(row: KubernetesCluster) { const { data } = await testKubernetesCluster(row.id); ElMessage.success(`连接成功：${data.data.version}`); load() }
async function showCredential(row: KubernetesCluster) { const { data } = await getKubernetesClusterConfig(row.id); credentialCluster.value = row.alias || row.name; credentialContent.value = data.data.kube_config; credentialDialog.value = true }
function openAuthorization(row: KubernetesCluster) { router.push({ path: '/kubernetes/access', query: { cluster_id: row.id } }) }
async function copyCredential() { await navigator.clipboard.writeText(credentialContent.value); ElMessage.success('凭证已复制') }
async function syncStatus() { syncing.value = true; try { const { data } = await syncKubernetesClusters(); const result = data.data; if (result.failed) ElMessage.warning(`同步完成：成功 ${result.success}，失败 ${result.failed}`); else ElMessage.success(`已同步 ${result.success} 个集群`); await load() } finally { syncing.value = false } }
async function remove(row: KubernetesCluster) { await ElMessageBox.confirm(`确定删除集群“${row.alias || row.name}”？`, '确认删除'); await deleteKubernetesCluster(row.id); load() }
watch(keyword, load)
const stats = computed(() => ({ total: allRows.value.length, healthy: allRows.value.filter(x => x.status === 1).length, nodes: allRows.value.reduce((sum, x) => sum + Number(x.node_count || 0), 0), selfHosted: allRows.value.filter(x => !x.provider || x.provider === 'self').length }))
const providerNames: Record<string, string> = { self: '自建集群', aliyun: '阿里云 ACK', tencent: '腾讯云 TKE', huawei: '华为云 CCE', aws: 'Amazon EKS', gcp: 'Google GKE', azure: 'Microsoft AKS' }
function actionButton(text: string, type: 'credential' | 'permission' | 'sync' | 'edit' | 'delete', onClick: () => void) { const button = document.createElement('button'); button.className = `grid-action compact ${type}`; button.textContent = text; button.onclick = onClick; return button }
const columns = computed<ColDef[]>(() => [
  { headerName: '序号', valueGetter: (p: any) => p.node.rowIndex + 1, width: 70 },
  { field: 'name', headerName: '▰ 集群名称', minWidth: 170 },
  { field: 'alias', headerName: '别名', minWidth: 120, valueFormatter: p => p.value || '-' },
  { field: 'status', headerName: '状态', width: 92, cellRenderer: (p: any) => { const tag = document.createElement('span'); tag.className = `cluster-status ${p.value ? 'is-healthy' : 'is-offline'}`; tag.textContent = p.value ? '正常' : '未连接'; return tag } },
  { field: 'version', headerName: '● 版本', minWidth: 145, valueFormatter: p => p.value || '-' },
  { field: 'node_count', headerName: '节点数', width: 90 },
  { field: 'provider', headerName: '服务商', minWidth: 130, valueFormatter: p => providerNames[p.value] || p.value || '自建集群' },
  { field: 'region', headerName: '区域', minWidth: 120, valueFormatter: p => p.value || '-' },
  { field: 'api_endpoint', headerName: 'API 地址', minWidth: 220, valueFormatter: p => p.value || '-' },
  { field: 'description', headerName: '备注', minWidth: 150, flex: 1, valueFormatter: p => p.value || '-' },
  { headerName: '操作', pinned: 'right', width: 180, cellRenderer: (p: any) => { const wrap = document.createElement('div'); wrap.className = 'grid-action-group compact'; wrap.append(actionButton('凭证', 'credential', () => showCredential(p.data)), actionButton('授权', 'permission', () => openAuthorization(p.data)), actionButton('同步', 'sync', () => test(p.data)), actionButton('编辑', 'edit', () => open(p.data)), actionButton('删除', 'delete', () => remove(p.data))); return wrap } },
])
onMounted(load)
</script>

<template>
  <section class="page">
    <div class="toolbar"><el-input v-model="keyword" placeholder="输入集群名称或别名搜索" clearable/><div class="toolbar-actions"><el-button :loading="syncing" @click="syncStatus">同步状态</el-button><el-button type="primary" @click="open()">注册集群</el-button></div></div>
    <div class="stats">
      <div class="stat-card total"><div class="stat-icon">▰</div><div class="stat-content"><span>集群总数</span><strong>{{ stats.total }}</strong></div></div>
      <div class="stat-card healthy"><div class="stat-icon">✓</div><div class="stat-content"><span>运行正常</span><strong>{{ stats.healthy }}</strong></div></div>
      <div class="stat-card nodes"><div class="stat-icon">◇</div><div class="stat-content"><span>总节点数</span><strong>{{ stats.nodes }}</strong></div></div>
      <div class="stat-card self-hosted"><div class="stat-icon">⌂</div><div class="stat-content"><span>自建集群</span><strong>{{ stats.selfHosted }}</strong></div></div>
    </div>
    <AppDataGrid class="grid" :row-data="rows" :column-defs="columns" pagination/>
    <el-dialog v-model="dialog" class="register-dialog" :title="form.id ? '编辑集群' : '注册集群'" width="60%" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px" status-icon>
        <div class="section-title">基本信息</div>
        <div class="two-columns"><el-form-item label="集群名称" prop="name"><el-input v-model="form.name" placeholder="请输入集群名称"/></el-form-item><el-form-item label="集群别名"><el-input v-model="form.alias" placeholder="可选"/></el-form-item></div>
        <el-divider/><div class="section-title">认证配置</div>
        <el-form-item label="认证方式"><el-radio-group v-model="form.authMode"><el-radio-button value="kubeconfig">KubeConfig 文件</el-radio-button><el-radio-button value="serviceAccount">Service Account Token</el-radio-button></el-radio-group></el-form-item>
        <template v-if="form.authMode === 'kubeconfig'">
          <el-form-item label="配置内容" prop="kube_config"><div class="config-field"><el-upload :auto-upload="false" :show-file-list="false" :on-change="readConfig"><el-button>上传 KubeConfig 文件</el-button></el-upload><el-input v-model="form.kube_config" class="config-textarea" type="textarea" :rows="7" resize="vertical" :spellcheck="false" placeholder="请粘贴 KubeConfig 文件内容"/><el-alert title="如何获取 KubeConfig？通常位于 ~/.kube/config 文件中" type="info" :closable="false" show-icon/></div></el-form-item>
        </template>
        <template v-else>
          <div class="two-columns"><el-form-item label="API Server" prop="api_endpoint"><el-input v-model="form.api_endpoint" placeholder="https://127.0.0.1:6443"/></el-form-item><el-form-item label="CA Data"><el-input v-model="form.ca_data" placeholder="可选，Base64 编码的 CA 证书"/></el-form-item></div>
          <el-form-item label="访问令牌" prop="token"><el-input v-model="form.token" type="textarea" :rows="5" show-password placeholder="请输入 Service Account Token"/></el-form-item>
          <el-alert title="未填写 CA Data 时将跳过 API Server TLS 证书校验，请仅在可信网络中使用。" type="warning" :closable="false" show-icon/>
        </template>
        <el-divider/><div class="section-title">集群信息</div>
        <div class="two-columns"><el-form-item label="服务商"><el-select v-model="form.provider"><el-option label="自建集群" value="self"/><el-option label="阿里云 ACK" value="aliyun"/><el-option label="腾讯云 TKE" value="tencent"/><el-option label="华为云 CCE" value="huawei"/><el-option label="Amazon EKS" value="aws"/><el-option label="Google GKE" value="gcp"/><el-option label="Microsoft AKS" value="azure"/></el-select></el-form-item><el-form-item label="区域"><el-input v-model="form.region" placeholder="例如：cn-beijing"/></el-form-item></div>
        <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入集群备注（可选）"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">{{ form.id ? '保存修改' : '注册集群' }}</el-button></template>
    </el-dialog>
    <el-dialog v-model="credentialDialog" :title="`${credentialCluster} · 集群凭证`" width="60%" destroy-on-close><el-alert title="KubeConfig 包含集群访问凭据，请妥善保管，避免通过不安全渠道传播。" type="warning" :closable="false" show-icon/><el-input v-model="credentialContent" class="credential-textarea" type="textarea" :rows="18" readonly/><template #footer><el-button @click="credentialDialog=false">关闭</el-button><el-button type="primary" @click="copyCredential">复制凭证</el-button></template></el-dialog>
  </section>
</template>

<style scoped>
.page{height:100%;display:flex;flex-direction:column;gap:10px;min-width:0;min-height:0}.toolbar,.toolbar-actions{display:flex;align-items:center;gap:10px}.toolbar{justify-content:space-between}.toolbar .el-input{width:260px}.stats{display:grid;grid-template-columns:repeat(4,minmax(160px,240px));justify-content:start;gap:10px}.stat-card{display:flex;align-items:center;width:100%;max-width:240px;min-width:0;padding:14px 16px;border:1px solid var(--line);border-radius:9px;background:#fff;box-shadow:0 3px 12px rgba(31,35,48,.06)}.stat-icon{display:flex;align-items:center;justify-content:center;width:34px;height:34px;flex:none;margin-right:12px;border-radius:8px;font-size:16px;font-weight:700;color:var(--primary);background:color-mix(in srgb,var(--primary) 10%,white)}.stat-content{display:flex;min-width:0;flex:1;flex-direction:column;gap:5px}.stat-card span{font-size:11px;color:var(--text-secondary,#717784)}.stat-card strong{font-size:22px;line-height:1;color:var(--text)}.stat-card.healthy .stat-icon,.stat-card.healthy strong{color:#20a162}.stat-card.healthy .stat-icon{background:#eaf8ef}.stat-card.nodes .stat-icon{color:#2878c8;background:#edf5ff}.stat-card.self-hosted .stat-icon{color:#b57816;background:#fff6df}.grid{flex:1;min-height:0}.grid :deep(.cluster-status){display:inline-flex;align-items:center;justify-content:center;padding:3px 9px;border-radius:5px;font-size:11px;font-weight:600;line-height:1.2}.grid :deep(.cluster-status.is-healthy){color:#247d31;background:#eaf7e8;border:1px solid #9bd393}.grid :deep(.cluster-status.is-offline){color:#b54747;background:#fff0f0;border:1px solid #efb1b1}.section-title{border-left:3px solid var(--primary);padding-left:8px;margin:2px 0 16px;font-size:14px;font-weight:650;color:var(--text)}.two-columns{display:grid;grid-template-columns:1fr 1fr;gap:0 28px}.config-field{display:flex;flex:1;min-width:0;width:100%;flex-direction:column;align-items:stretch;gap:9px}.config-field :deep(.el-upload){justify-content:flex-start}.config-textarea :deep(.el-textarea__inner){min-height:188px!important;max-height:52vh;padding:14px 16px;border-radius:10px;background:#fff;color:var(--text);font:12px/1.65 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;white-space:pre;overflow:auto;tab-size:2;box-shadow:0 0 0 1px var(--line) inset}.config-textarea :deep(.el-textarea__inner:hover){box-shadow:0 0 0 1px #c6cbd5 inset}.config-textarea :deep(.el-textarea__inner:focus){box-shadow:0 0 0 1px var(--primary) inset}.config-textarea :deep(.el-textarea__inner::placeholder){color:#a7abb4}.credential-textarea{margin-top:12px}.credential-textarea :deep(.el-textarea__inner){padding:14px 16px;border:1px solid var(--line);border-radius:8px;background:#fff;color:var(--text);-webkit-text-fill-color:var(--text);font:12px/1.55 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;white-space:pre;overflow:auto;opacity:1;box-shadow:0 0 0 1px var(--line) inset}.register-dialog :deep(.el-divider){margin:18px 0}.register-dialog :deep(.el-select){width:100%}.register-dialog :deep(.el-alert){width:auto;font-size:11px}@media(max-width:760px){.stats{grid-template-columns:repeat(2,minmax(0,1fr))}.stat-card{max-width:none}.two-columns{grid-template-columns:1fr}.toolbar{flex-wrap:wrap}.toolbar .el-input{width:100%}.config-textarea :deep(.el-textarea__inner){min-height:150px!important}}
</style>
