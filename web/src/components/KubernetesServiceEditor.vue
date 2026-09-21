<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'

type Port = { name: string; protocol: string; port: number; targetPort: string | number; nodePort?: number }
const props = defineProps<{ modelValue: boolean; yaml: string; saving?: boolean; createMode?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; save: [yaml: string] }>()
const visible = computed({ get: () => props.modelValue, set: value => emit('update:modelValue', value) })
const form = reactive({ name: '', namespace: 'default', type: 'ClusterIP', selectorKey: 'app', selectorValue: '', ports: [] as Port[] })
let manifest: Record<string, any> = {}

function initialize() {
  try { manifest = parse(props.yaml) || {} } catch { manifest = {} }
  const metadata = manifest.metadata || {}
  const spec = manifest.spec || {}
  const selector = Object.entries(spec.selector || {})[0] || ['app', '']
  form.name = String(metadata.name || 'example')
  form.namespace = String(metadata.namespace || 'default')
  form.type = String(spec.type || 'ClusterIP')
  form.selectorKey = String(selector[0])
  form.selectorValue = String(selector[1] || form.name)
  form.ports = (spec.ports || []).map((item: any) => ({ name: String(item.name || ''), protocol: String(item.protocol || 'TCP'), port: Number(item.port || 80), targetPort: item.targetPort ?? item.port ?? 80, nodePort: item.nodePort }))
  if (!form.ports.length) addPort()
}
function addPort() { form.ports.push({ name: 'http', protocol: 'TCP', port: 80, targetPort: 80 }) }
function save() {
  manifest.apiVersion = 'v1'
  manifest.kind = 'Service'
  manifest.metadata = { ...(manifest.metadata || {}), name: form.name, namespace: form.namespace }
  const ports = form.ports.map(port => {
    const value: Record<string, unknown> = { name: port.name || undefined, protocol: port.protocol, port: Number(port.port), targetPort: /^\d+$/.test(String(port.targetPort)) ? Number(port.targetPort) : port.targetPort }
    if (form.type === 'NodePort' && port.nodePort) value.nodePort = Number(port.nodePort)
    return Object.fromEntries(Object.entries(value).filter(([, item]) => item !== undefined && item !== ''))
  })
  manifest.spec = { ...(manifest.spec || {}), type: form.type, selector: { [form.selectorKey || 'app']: form.selectorValue }, ports }
  if (form.type === 'ExternalName') delete manifest.spec.selector
  emit('save', stringify(manifest, { lineWidth: 0 }))
}
watch(() => props.modelValue, value => { if (value) initialize() }, { immediate: true })
</script>

<template>
  <el-dialog v-model="visible" :title="createMode ? '创建服务' : '编辑服务'" width="min(760px,calc(100vw - 20px))" align-center destroy-on-close class="service-editor-dialog">
    <el-form label-position="top" class="service-form">
      <div class="form-grid">
        <el-form-item label="服务名称"><el-input v-model="form.name" placeholder="请输入服务名称"/></el-form-item>
        <el-form-item label="命名空间"><el-input v-model="form.namespace"/></el-form-item>
        <el-form-item label="服务类型"><el-select v-model="form.type"><el-option label="ClusterIP" value="ClusterIP"/><el-option label="NodePort" value="NodePort"/><el-option label="LoadBalancer" value="LoadBalancer"/><el-option label="ExternalName" value="ExternalName"/></el-select></el-form-item>
      </div>
      <section v-if="form.type!=='ExternalName'" class="service-section"><header><strong>Pod 选择器</strong></header><div class="selector-row"><el-input v-model="form.selectorKey" placeholder="键"/><span>=</span><el-input v-model="form.selectorValue" placeholder="值"/></div></section>
      <section class="service-section"><header><strong>端口配置</strong><el-button link :icon="Plus" @click="addPort">添加端口</el-button></header><div v-for="(port,index) in form.ports" :key="index" class="port-row"><el-input v-model="port.name" placeholder="名称"/><el-select v-model="port.protocol"><el-option label="TCP" value="TCP"/><el-option label="UDP" value="UDP"/><el-option label="SCTP" value="SCTP"/></el-select><el-input-number v-model="port.port" :min="1" :max="65535" controls-position="right"/><el-input v-model="port.targetPort" placeholder="目标端口"/><el-input-number v-if="form.type==='NodePort'" v-model="port.nodePort" :min="30000" :max="32767" controls-position="right"/><el-button link type="danger" :icon="Delete" @click="form.ports.splice(index,1)"/></div></section>
    </el-form>
    <template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
  </el-dialog>
</template>

<style scoped>
.service-form{font-size:11px}.form-grid{display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px}.service-form :deep(.el-form-item){margin-bottom:10px}.service-form :deep(.el-form-item__label){font-size:11px;font-weight:600}.service-section{margin-top:8px;padding:12px;border:1px solid var(--line);border-radius:8px}.service-section header{display:flex;align-items:center;justify-content:space-between;margin-bottom:10px}.selector-row{display:grid;grid-template-columns:1fr 18px 1fr;align-items:center;text-align:center}.port-row{display:grid;grid-template-columns:1fr 90px 120px 1fr auto;gap:7px;align-items:center;margin-top:7px}.port-row:has(.el-input-number:nth-last-child(2)){grid-template-columns:1fr 80px 105px 1fr 115px auto}@media(max-width:700px){.form-grid{grid-template-columns:1fr}.port-row,.port-row:has(.el-input-number:nth-last-child(2)){grid-template-columns:1fr 1fr}.port-row .el-button{justify-self:end}}
</style>
