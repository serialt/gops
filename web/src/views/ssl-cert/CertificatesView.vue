<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { ColDef } from 'ag-grid-community'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppDataGrid from '@/components/AppDataGrid.vue'
import CertificateApplyView from '@/views/ssl-cert/CertificateApplyView.vue'
import { deleteCertificate, getCertificate, listCertificates, listDNSProviders, renewCertificate, saveCertificate, type SSLCertificate, type SSLCertificateForm, type SSLDNSProvider } from '@/api/ssl-cert'
import { useAuthStore } from '@/stores/auth'
const auth=useAuthStore(),rows=ref<SSLCertificate[]>([]),providers=ref<SSLDNSProvider[]>([]),keyword=ref(''),status=ref(''),sourceType=ref(''),page=ref(1),pageSize=ref(20),total=ref(0),loading=ref(false),dialog=ref(false),createTab=ref('existing'),detailDialog=ref(false),detail=ref<SSLCertificate|null>(null)
const detailLoading=false
const empty=():SSLCertificateForm=>({name:'',domain:'',san_domains:'',acme_email:'',ca_provider:'letsencrypt',key_algorithm:'rsa2048',source_type:'manual',cloud_account_id:0,dns_provider_id:0,certificate:'',private_key:'',cert_chain:'',auto_renew:true,renew_days_before:30});const form=reactive<SSLCertificateForm>(empty())
async function load(){loading.value=true;try{const {data}=await listCertificates({keyword:keyword.value,status:status.value,source_type:sourceType.value,page:page.value,limit:pageSize.value});rows.value=data.data.list;total.value=data.data.total}finally{loading.value=false}}
async function open(row?:SSLCertificate){Object.assign(form,empty());createTab.value=auth.hasPermission('POST /api/v1/ssl-cert/certificates/save')?'existing':'apply';if(row){const {data}=await getCertificate(row.id);Object.assign(form,data.data)}dialog.value=true}
async function openDetail(row:SSLCertificate){detail.value={...row};detailDialog.value=true;try{const{data}=await getCertificate(row.id);detail.value=data.data}catch{ElMessage.error('证书详情加载失败，请稍后重试')}}
async function save(){if(form.id&&(!form.name.trim()||!form.domain.trim()))return ElMessage.warning('请填写证书名称和域名');if(!form.id&&(!form.certificate.trim()||!form.private_key.trim()))return ElMessage.warning('请输入 CRT 证书和私钥 Key');await saveCertificate(form);dialog.value=false;ElMessage.success('保存成功');await load()}
async function remove(row:SSLCertificate,confirm=true){if(confirm)await ElMessageBox.confirm(`确认删除证书“${row.name}”吗？`,'删除确认',{type:'warning'});await deleteCertificate(row.id)}
async function bulkDelete(selected:unknown[]){const items=selected as SSLCertificate[];await ElMessageBox.confirm(`确认删除选中的 ${items.length} 个证书吗？`,'批量删除',{type:'warning'});for(const item of items)await remove(item,false);await load()}
async function renew(row:SSLCertificate){await renewCertificate(row.id);ElMessage.success('续期任务已创建')}
function actions(p:{data:SSLCertificate}){const w=document.createElement('div');w.className='ssl-grid-actions';const add=(text:string,kind:string,fn:()=>void,disabled=false)=>{const b=document.createElement('button');b.type='button';b.textContent=text;b.className=`grid-action compact ${kind}`;b.disabled=disabled;if(disabled)b.title='手动导入的证书不支持续期';b.addEventListener('click',event=>{event.stopPropagation();if(!disabled)fn()});w.append(b)};if(auth.hasPermission('POST /api/v1/ssl-cert/certificates/renew'))add('续期','permission',()=>void renew(p.data),p.data.source_type==='manual');if(auth.hasPermission('POST /api/v1/ssl-cert/certificates/save'))add('编辑','edit',()=>void open(p.data));if(auth.hasPermission('POST /api/v1/ssl-cert/certificates/delete'))add('删除','delete',()=>void remove(p.data).then(load));return w}
function nameRenderer(p:{data:SSLCertificate;value:string}){const button=document.createElement('button');button.type='button';button.className='certificate-name-link';button.textContent=p.value||'-';button.addEventListener('click',event=>{event.stopPropagation();void openDetail(p.data)});return button}
const time=(value:string|null|undefined)=>value?String(value).replace('T',' ').slice(0,19):'-'
const sourceLabel=(value:string)=>({manual:'手动导入',acme:'ACME',aliyun:'阿里云'}[value]||value||'-')
const statusLabel=(value:string)=>({pending:'待处理',active:'有效',expiring:'即将到期',expired:'已过期',error:'错误'}[value]||value||'-')
const columns=computed<ColDef[]>(()=>[{headerName:'序号',width:64,valueGetter:p=>(page.value-1)*pageSize.value+(p.node?.rowIndex||0)+1},{field:'name',headerName:'证书名称',minWidth:140,cellRenderer:nameRenderer},{field:'domain',headerName:'域名',minWidth:170},{field:'source_type',headerName:'来源',width:90,valueFormatter:p=>sourceLabel(p.value)},{field:'ca_provider',headerName:'CA',width:100},{field:'status',headerName:'状态',width:90,valueFormatter:p=>statusLabel(p.value)},{field:'not_after',headerName:'到期时间',minWidth:145,valueFormatter:p=>time(p.value)},{headerName:'操作',pinned:'right',width:150,minWidth:150,maxWidth:150,cellClass:'ssl-operation-cell',cellRenderer:actions}])
watch([keyword,status,sourceType],()=>{page.value=1;void load()});onMounted(async()=>{await load();const {data}=await listDNSProviders({keyword:'',provider:'',page:1,limit:1000});providers.value=data.data.list})
</script>
<template><section class="ssl-page"><header><div><h2>证书管理</h2><p>管理证书申请、导入、续期和有效期。</p></div></header><div class="toolbar"><div class="filters"><el-input v-model="keyword" placeholder="搜索证书或域名" clearable/><el-select v-model="sourceType" placeholder="选择证书来源" clearable><el-option label="手动导入" value="manual"/><el-option label="ACME" value="acme"/><el-option label="阿里云" value="aliyun"/></el-select><el-select v-model="status" placeholder="选择状态" clearable><el-option label="待处理" value="pending"/><el-option label="有效" value="active"/><el-option label="即将到期" value="expiring"/><el-option label="已过期" value="expired"/><el-option label="错误" value="error"/></el-select></div><div class="toolbar-actions"><el-button v-if="auth.hasPermission('POST /api/v1/ssl-cert/certificates/save')||auth.hasPermission('POST /api/v1/ssl-cert/applications/create')" type="primary" @click="open()">新增证书</el-button></div></div><AppDataGrid v-loading="loading" class="grid" :row-data="rows" :column-defs="columns" pagination server-pagination :total="total" :current-page="page" :page-size="pageSize" :bulk-delete="auth.hasPermission('POST /api/v1/ssl-cert/certificates/delete')?bulkDelete:undefined" @page-change="v=>{page=v;void load()}" @page-size-change="v=>{pageSize=v;page=1;void load()}"/>
<el-dialog v-model="dialog" :title="form.id?'编辑证书':'新增证书'" width="760px" destroy-on-close>
  <template v-if="form.id">
    <el-form label-width="90px"><div class="form-row"><el-form-item label="证书名称" required><el-input v-model="form.name"/></el-form-item><el-form-item label="主域名" required><el-input v-model="form.domain"/></el-form-item></div><el-form-item label="SAN 域名"><el-input v-model="form.san_domains" placeholder="多个域名用逗号分隔"/></el-form-item><div class="form-row"><el-form-item label="证书来源"><el-select v-model="form.source_type"><el-option label="手动导入" value="manual"/><el-option label="ACME" value="acme"/><el-option label="阿里云" value="aliyun"/></el-select></el-form-item><el-form-item label="CA 提供商"><el-select v-model="form.ca_provider"><el-option label="Let's Encrypt" value="letsencrypt"/><el-option label="ZeroSSL" value="zerossl"/><el-option label="Google" value="google"/><el-option label="Buypass" value="buypass"/></el-select></el-form-item></div><div class="form-row"><el-form-item label="密钥算法"><el-select v-model="form.key_algorithm"><el-option label="RSA 2048" value="rsa2048"/><el-option label="RSA 4096" value="rsa4096"/><el-option label="EC 256" value="ec256"/></el-select></el-form-item><el-form-item label="DNS 配置"><el-select v-model="form.dns_provider_id" clearable><el-option v-for="item in providers" :key="item.id" :label="item.name" :value="item.id"/></el-select></el-form-item></div><div class="form-row"><el-form-item label="ACME 邮箱"><el-input v-model="form.acme_email"/></el-form-item><el-form-item label="提前续期"><el-input-number v-model="form.renew_days_before" :min="1" :max="90"/><span class="unit">天</span></el-form-item></div><el-form-item label="自动续期"><el-switch v-model="form.auto_renew"/></el-form-item><template v-if="form.source_type==='manual'"><el-form-item label="证书 CRT"><el-input v-model="form.certificate" type="textarea" :rows="5"/></el-form-item><el-form-item label="证书链"><el-input v-model="form.cert_chain" type="textarea" :rows="3"/></el-form-item><el-form-item label="私钥 Key"><el-input v-model="form.private_key" type="textarea" :rows="5"/></el-form-item></template></el-form>
    <div class="dialog-actions"><el-button @click="dialog=false">取消</el-button><el-button type="primary" @click="save">保存</el-button></div>
  </template>
  <el-tabs v-else v-model="createTab" class="create-certificate-tabs">
    <el-tab-pane v-if="auth.hasPermission('POST /api/v1/ssl-cert/certificates/save')" label="添加已存在的证书" name="existing">
      <el-form label-width="82px">
        <el-form-item label="证书 CRT" required><el-input v-model="form.certificate" type="textarea" :rows="7" placeholder="粘贴 PEM 格式 CRT；可包含完整证书链"/></el-form-item>
        <el-form-item label="证书链"><el-input v-model="form.cert_chain" type="textarea" :rows="5" placeholder="选填；未填写时自动从 CRT 中拆分证书链"/></el-form-item>
        <el-form-item label="私钥 Key" required><el-input v-model="form.private_key" type="textarea" :rows="7" placeholder="粘贴与证书匹配的 PEM 格式私钥"/></el-form-item>
      </el-form>
      <div class="dialog-actions"><el-button @click="dialog=false">取消</el-button><el-button type="primary" @click="save">添加证书</el-button></div>
    </el-tab-pane>
    <el-tab-pane v-if="auth.hasPermission('POST /api/v1/ssl-cert/applications/create')" label="申请证书" name="apply">
      <CertificateApplyView embedded @cancel="dialog=false" @submitted="dialog=false;load()"/>
    </el-tab-pane>
  </el-tabs>
</el-dialog>
<el-dialog v-model="detailDialog" title="证书详情" width="820px" destroy-on-close class="certificate-detail-dialog">
  <div class="certificate-detail">
    <template v-if="detail">
      <div class="detail-summary"><div><b>{{ detail.name }}</b><span>{{ detail.domain }}</span></div><span class="detail-status" :class="`is-${detail.status}`">{{ statusLabel(detail.status) }}</span></div>
      <div class="detail-grid"><div><span>SAN 域名</span><b>{{ detail.san_domains || '-' }}</b></div><div><span>证书来源</span><b>{{ sourceLabel(detail.source_type) }}</b></div><div><span>CA 提供商</span><b>{{ detail.ca_provider || '-' }}</b></div><div><span>密钥算法</span><b>{{ detail.key_algorithm || '-' }}</b></div><div><span>生效时间</span><b>{{ time(detail.not_before) }}</b></div><div><span>到期时间</span><b>{{ time(detail.not_after) }}</b></div><div><span>自动续期</span><b>{{ detail.auto_renew ? `是，提前 ${detail.renew_days_before} 天` : '否' }}</b></div><div><span>ACME 邮箱</span><b>{{ detail.acme_email || '-' }}</b></div><div class="detail-wide"><span>签发者</span><b>{{ detail.issuer || '-' }}</b></div><div class="detail-wide"><span>证书指纹</span><code>{{ detail.fingerprint || '-' }}</code></div></div>
      <el-alert v-if="detail.last_error" :title="detail.last_error" type="error" :closable="false" show-icon/>
      <el-tabs class="certificate-material">
        <el-tab-pane label="当前证书"><pre>{{ detail.certificate || '-' }}</pre></el-tab-pane>
        <el-tab-pane label="原始证书链 PEM"><pre>{{ detail.cert_chain || '-' }}</pre></el-tab-pane>
      </el-tabs>
      <section v-if="detail.chain_certificates?.length || detail.cert_chain" class="certificate-chain">
        <header><h3>证书链</h3><span>{{ detail.chain_certificates?.length || 0 }} 级</span></header>
        <el-alert v-if="detail.chain_parse_error" :title="detail.chain_parse_error" type="warning" :closable="false" show-icon/>
        <div v-if="detail.chain_certificates?.length" class="chain-list">
          <article v-for="item in detail.chain_certificates" :key="item.position">
            <header><b>第 {{ (detail.chain_certificates?.length || 0) - item.position + 1 }} 级证书</b><span>{{ item.is_ca ? 'CA 证书' : '终端证书' }}</span></header>
            <div>
              <span>版本</span><b>V{{ item.version }}</b>
              <span>通用名称</span><b>{{ item.common_name || '-' }}</b>
              <span>主体</span><b>{{ item.subject || '-' }}</b>
              <span>签发者</span><b>{{ item.issuer || '-' }}</b>
              <span>序列号</span><code>{{ item.serial_number || '-' }}</code>
              <span>签名算法</span><b>{{ item.signature_algorithm || '-' }}</b>
              <span>公钥</span><b>{{ item.public_key_algorithm || '-' }}{{ item.public_key_bits ? ` / ${item.public_key_bits} bit` : '' }}</b>
              <span>生效时间</span><b>{{ time(item.not_before) }}</b>
              <span>到期时间</span><b>{{ time(item.not_after) }}</b>
              <span>DNS 名称</span><b>{{ (item.dns_names || []).join('、') || '-' }}</b>
              <span>IP 地址</span><b>{{ (item.ip_addresses || []).join('、') || '-' }}</b>
              <span>邮箱地址</span><b>{{ (item.email_addresses || []).join('、') || '-' }}</b>
              <span>URI</span><b>{{ (item.uris || []).join('、') || '-' }}</b>
              <span>密钥用途</span><b>{{ (item.key_usage || []).join('、') || '-' }}</b>
              <span>扩展用途</span><b>{{ (item.ext_key_usage || []).join('、') || '-' }}</b>
              <span>基本约束</span><b>{{ item.basic_constraints_valid ? (item.is_ca ? `CA，路径长度 ${item.max_path_len_zero ? 0 : (item.max_path_len < 0 ? '不限' : item.max_path_len)}` : '终端证书') : '-' }}</b>
              <span>Subject Key ID</span><code>{{ item.subject_key_id || '-' }}</code>
              <span>Authority Key ID</span><code>{{ item.authority_key_id || '-' }}</code>
              <span>OCSP</span><b>{{ (item.ocsp_servers || []).join('、') || '-' }}</b>
              <span>颁发者地址</span><b>{{ (item.issuer_urls || []).join('、') || '-' }}</b>
              <span>CRL 地址</span><b>{{ (item.crl_points || []).join('、') || '-' }}</b>
              <span>策略 OID</span><b>{{ (item.policies || []).join('、') || '-' }}</b>
              <span>SHA-256</span><code>{{ item.fingerprint || '-' }}</code>
            </div>
          </article>
        </div>
      </section>
    </template>
  </div>
  <template #footer><el-button @click="detailDialog=false">关闭</el-button></template>
</el-dialog></section></template>
<style scoped>.dialog-actions{display:flex;justify-content:flex-end;gap:6px;padding-top:8px;border-top:1px solid var(--line)}.create-certificate-tabs :deep(.el-tabs__header){margin-bottom:14px}</style>
<style scoped>.ssl-page{height:100%;display:flex;flex-direction:column;gap:8px;min-width:0;padding:0 4px;font-size:11px}header h2{margin:0;font-size:20px}header p{margin:3px 0 0;color:var(--text-soft)}.toolbar{display:flex;justify-content:space-between;gap:8px}.toolbar-actions{display:flex;gap:6px}.filters{display:flex;gap:6px}.filters .el-input{width:170px}.filters .el-select{width:130px}.grid{flex:1;min-height:0}.form-row{display:grid;grid-template-columns:1fr 1fr;gap:10px}.form-row .el-select{width:100%}.unit{margin-left:5px}:deep(.certificate-name-link){padding:0;border:0;background:transparent;color:var(--primary);font:inherit;cursor:pointer;text-align:left}:deep(.certificate-name-link:hover){text-decoration:underline}.certificate-detail{min-height:180px;font-size:10px}.detail-summary{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:8px;padding:10px 12px;border:1px solid var(--line);border-radius:8px;background:#fafbfc}.detail-summary>div{display:flex;min-width:0;flex-direction:column;gap:3px}.detail-summary b{overflow:hidden;font-size:14px;text-overflow:ellipsis;white-space:nowrap}.detail-summary div>span{color:var(--text-soft);font-size:10px}.detail-status{flex:none;padding:3px 8px;border:1px solid #d7dbe3;border-radius:5px;background:#f3f4f6;color:var(--text-soft);font-weight:600}.detail-status.is-active{border-color:#a7e3c5;background:#eefaf4;color:#128455}.detail-status.is-expiring,.detail-status.is-pending{border-color:#f4d58d;background:#fff8e7;color:#a56708}.detail-status.is-expired,.detail-status.is-error{border-color:#f5b7b1;background:#fff2f1;color:#d92d20}.detail-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));border:1px solid var(--line);border-radius:8px;overflow:hidden}.detail-grid>div{display:grid;grid-template-columns:82px minmax(0,1fr);min-width:0;border-bottom:1px solid var(--line)}.detail-grid>div:nth-child(odd){border-right:1px solid var(--line)}.detail-grid>div:nth-last-child(-n+2){border-bottom:0}.detail-grid .detail-wide{grid-column:1/-1;border-right:0}.detail-grid .detail-wide:nth-last-child(2){border-bottom:1px solid var(--line)}.detail-grid span,.detail-grid b,.detail-grid code{min-width:0;padding:7px 9px;overflow-wrap:anywhere}.detail-grid span{background:#f7f8fb;color:var(--text-soft);font-weight:600}.detail-grid b,.detail-grid code{font-weight:500}.detail-grid code{font:10px/1.45 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;word-break:break-all}.certificate-detail .el-alert{margin-top:8px}.pem-tabs{margin-top:8px}.pem-tabs :deep(.el-tabs__header){margin-bottom:6px}.pem-tabs :deep(.el-tabs__item){height:30px;padding:0 10px;font-size:10px}.pem-tabs pre{height:190px;margin:0;padding:9px;overflow:auto;border:1px solid var(--line);border-radius:7px;background:#f7f8fb;white-space:pre;font:9px/1.5 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}@media(max-width:760px){.form-row{grid-template-columns:1fr}.toolbar{flex-wrap:wrap}.toolbar-actions{flex-wrap:wrap}.detail-grid{grid-template-columns:1fr}.detail-grid>div:nth-child(odd){border-right:0}.detail-grid>div:nth-last-child(-n+2){border-bottom:1px solid var(--line)}.detail-grid>div:last-child{border-bottom:0}}</style>
