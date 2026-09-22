<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Delete, Plus, Upload } from '@element-plus/icons-vue'
import { parse, stringify } from 'yaml'

type Pair = { key: string; value: string }
const props = defineProps<{ modelValue: boolean; yaml: string; resource: 'configmaps'|'secrets'; namespaces: string[]; saving?: boolean; createMode?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; save: [yaml: string] }>()
const visible = computed({ get: () => props.modelValue, set: value => emit('update:modelValue', value) })
const isSecret = computed(() => props.resource === 'secrets')
const activeTab = ref('data')
const name = ref('')
const namespace = ref('default')
const secretType = ref('Opaque')
const data = ref<Pair[]>([])
const binaryData = ref<Pair[]>([])
const labels = ref<Pair[]>([])
const annotations = ref<Pair[]>([])
const fileInput = ref<HTMLInputElement>()
let manifest: Record<string, any> = {}

function pairs(value: Record<string, unknown> | undefined, decode = false): Pair[] {
  return Object.entries(value || {}).map(([key, item]) => ({ key, value: decode ? decodeBase64(String(item)) : String(item) }))
}
function objectFromPairs(items: Pair[]) { return Object.fromEntries(items.filter(item => item.key.trim()).map(item => [item.key.trim(), item.value])) }
function decodeBase64(value: string) { try { return decodeURIComponent(Array.from(atob(value), char => `%${char.charCodeAt(0).toString(16).padStart(2, '0')}`).join('')) } catch { return value } }
function initialize() {
  try { manifest = parse(props.yaml) || {} } catch { manifest = {} }
  name.value = String(manifest.metadata?.name || '')
  namespace.value = String(manifest.metadata?.namespace || props.namespaces[0] || 'default')
  secretType.value = String(manifest.type || 'Opaque')
  data.value = pairs(isSecret.value ? (manifest.stringData || manifest.data) : manifest.data, isSecret.value && !manifest.stringData)
  binaryData.value = pairs(manifest.binaryData, true)
  labels.value = pairs(manifest.metadata?.labels)
  annotations.value = pairs(manifest.metadata?.annotations)
  activeTab.value = 'data'
}
function add(target: Pair[]) { target.push({ key: '', value: '' }) }
function encodeBase64(value: string) { return btoa(unescape(encodeURIComponent(value))) }
function save() {
  const metadata: Record<string, any> = { ...(manifest.metadata || {}), name: name.value.trim(), namespace: namespace.value }
  const labelValues = objectFromPairs(labels.value), annotationValues = objectFromPairs(annotations.value)
  if (Object.keys(labelValues).length) metadata.labels = labelValues; else delete metadata.labels
  if (Object.keys(annotationValues).length) metadata.annotations = annotationValues; else delete metadata.annotations
  manifest = { ...manifest, apiVersion: 'v1', kind: isSecret.value ? 'Secret' : 'ConfigMap', metadata }
  const dataValues = objectFromPairs(data.value)
  if (isSecret.value) {
    manifest.type = secretType.value
    manifest.stringData = dataValues
    delete manifest.data
  } else {
    manifest.data = dataValues
    const binaryValues = objectFromPairs(binaryData.value)
    if (Object.keys(binaryValues).length) manifest.binaryData = Object.fromEntries(Object.entries(binaryValues).map(([key,value]) => [key, encodeBase64(value)])); else delete manifest.binaryData
  }
  emit('save', stringify(manifest, { lineWidth: 0 }))
}
function chooseFile() { fileInput.value?.click() }
async function importFiles(event: Event) {
  const files = Array.from((event.target as HTMLInputElement).files || [])
  for (const file of files) data.value.push({ key: file.name, value: await file.text() })
  ;(event.target as HTMLInputElement).value = ''
}
watch(() => props.modelValue, value => { if (value) initialize() }, { immediate: true })
</script>

<template>
  <el-dialog v-model="visible" :title="`${createMode?'新增':'编辑'} ${isSecret?'Secret':'ConfigMap'}`" width="min(980px,calc(100vw - 20px))" align-center destroy-on-close class="config-editor-dialog">
    <div class="identity-grid" :class="{secret:isSecret}"><label><span><i>*</i> 名称</span><el-input v-model="name" :placeholder="`请输入 ${isSecret?'Secret':'ConfigMap'} 名称`"/></label><label><span><i>*</i> 命名空间</span><el-select v-model="namespace" filterable allow-create><el-option v-for="item in namespaces" :key="item" :label="item" :value="item"/></el-select></label><label v-if="isSecret"><span><i>*</i> 类型</span><el-select v-model="secretType" filterable allow-create><el-option label="Opaque" value="Opaque"/><el-option label="kubernetes.io/tls" value="kubernetes.io/tls"/><el-option label="kubernetes.io/dockerconfigjson" value="kubernetes.io/dockerconfigjson"/><el-option label="kubernetes.io/basic-auth" value="kubernetes.io/basic-auth"/><el-option label="kubernetes.io/ssh-auth" value="kubernetes.io/ssh-auth"/></el-select></label></div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="数据" name="data"><section class="pair-section"><header><strong>Data</strong><div><input ref="fileInput" type="file" multiple hidden @change="importFiles"/><el-button v-if="isSecret" :icon="Upload" @click="chooseFile">上传文件</el-button><el-button type="primary" :icon="Plus" @click="add(data)">添加数据</el-button></div></header><div class="pair-table"><div class="pair-head"><span>Key</span><span>Value</span><span>操作</span></div><div v-if="!data.length" class="empty">暂无数据</div><div v-for="(item,index) in data" :key="index" class="pair-row"><el-input v-model="item.key" placeholder="Key"/><el-input v-model="item.value" type="textarea" autosize placeholder="Value"/><el-button link type="danger" :icon="Delete" @click="data.splice(index,1)"/></div></div></section><section v-if="!isSecret" class="pair-section"><header><strong>BinaryData</strong><el-button type="primary" :icon="Plus" @click="add(binaryData)">添加二进制数据</el-button></header><div class="pair-table"><div class="pair-head"><span>Key</span><span>Value</span><span>操作</span></div><div v-if="!binaryData.length" class="empty">暂无数据</div><div v-for="(item,index) in binaryData" :key="index" class="pair-row"><el-input v-model="item.key"/><el-input v-model="item.value" type="textarea" autosize/><el-button link type="danger" :icon="Delete" @click="binaryData.splice(index,1)"/></div></div></section></el-tab-pane>
      <el-tab-pane label="标签/注解" name="metadata"><section class="pair-section metadata-section"><header><strong>标签</strong><el-button :icon="Plus" @click="add(labels)">添加</el-button></header><div v-if="!labels.length" class="empty bordered">暂无数据</div><div v-for="(item,index) in labels" :key="index" class="metadata-row"><el-input v-model="item.key" placeholder="Key"/><el-input v-model="item.value" placeholder="Value"/><el-button link type="danger" :icon="Delete" @click="labels.splice(index,1)"/></div></section><section class="pair-section metadata-section"><header><strong>注解</strong><el-button :icon="Plus" @click="add(annotations)">添加</el-button></header><div v-if="!annotations.length" class="empty bordered">暂无数据</div><div v-for="(item,index) in annotations" :key="index" class="metadata-row"><el-input v-model="item.key" placeholder="Key"/><el-input v-model="item.value" placeholder="Value"/><el-button link type="danger" :icon="Delete" @click="annotations.splice(index,1)"/></div></section></el-tab-pane>
    </el-tabs>
    <template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" :loading="saving" :disabled="!name.trim()||!namespace" @click="save">{{createMode?'创建':'保存'}}</el-button></template>
  </el-dialog>
</template>

<style scoped>
:global(.config-editor-dialog){font-size:10px}:global(.config-editor-dialog .el-dialog__title){font-size:11px;line-height:16px}:global(.config-editor-dialog .el-dialog__body),:global(.config-editor-dialog .el-tabs__item),:global(.config-editor-dialog .el-button),:global(.config-editor-dialog .el-input__inner),:global(.config-editor-dialog .el-textarea__inner),:global(.config-editor-dialog .el-select__placeholder),:global(.config-editor-dialog .el-select__selected-item){font-size:10px}:global(.config-editor-dialog .el-dialog__footer){font-size:10px}
.identity-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px;margin-bottom:8px}.identity-grid.secret{grid-template-columns:repeat(3,1fr)}.identity-grid label{display:grid;grid-template-columns:78px 1fr;align-items:center;gap:7px}.identity-grid label>span{font-weight:600}.identity-grid i{color:var(--danger);font-style:normal}.identity-grid .el-select{width:100%}.pair-section{margin:10px 0;padding:12px;border:1px solid var(--line);border-radius:8px;background:#fafbfc}.pair-section header{display:flex;align-items:center;justify-content:space-between;margin-bottom:9px}.pair-section header>div{display:flex;gap:7px}.pair-table{border:1px solid var(--line);background:#fff}.pair-head,.pair-row{display:grid;grid-template-columns:210px 1fr 64px}.pair-head{height:34px;align-items:center;border-bottom:1px solid var(--line);color:var(--muted);font-weight:700}.pair-head span,.pair-row>*{padding:0 10px}.pair-row{align-items:center;padding:5px 0;border-top:1px solid var(--line)}.pair-row:first-of-type{border-top:0}.empty{display:grid;height:52px;place-items:center;color:var(--muted)}.empty.bordered{border:1px solid var(--line);background:#fff}.metadata-section{background:#fff;border:0;padding:2px 0}.metadata-row{display:grid;grid-template-columns:220px 1fr 50px;gap:7px;margin-top:7px}@media(max-width:760px){.identity-grid,.identity-grid.secret{grid-template-columns:1fr}.pair-head,.pair-row{grid-template-columns:120px 1fr 50px}}
</style>
