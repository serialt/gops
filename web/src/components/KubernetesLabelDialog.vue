<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = withDefaults(defineProps<{ modelValue: boolean; title?: string; labels: Record<string,string>; saving?: boolean }>(), { title: '资源标签', saving: false })
const emit = defineEmits<{ 'update:modelValue': [value:boolean]; save: [labels:Record<string,string>] }>()
const editing = ref(false)
const rows = ref<{key:string;value:string}[]>([])
watch(() => props.modelValue, value => { if (value) { editing.value = false; rows.value = Object.entries(props.labels || {}).map(([key,value]) => ({key,value})) } })
watch(() => props.labels, value => { if (props.modelValue && !editing.value) rows.value = Object.entries(value || {}).map(([key,item]) => ({key,value:item})) }, { deep:true })
function close(){emit('update:modelValue',false)}
function startEdit(){editing.value=true;if(!rows.value.length)rows.value.push({key:'',value:''})}
function add(){rows.value.push({key:'',value:''})}
function remove(index:number){rows.value.splice(index,1)}
function save(){const labels:Record<string,string>={};for(const row of rows.value){const key=row.key.trim();if(!key){ElMessage.warning('标签 Key 不能为空');return}if(key in labels){ElMessage.warning(`标签 Key 重复：${key}`);return}labels[key]=row.value}emit('save',labels)}
</script>

<template>
  <el-dialog :model-value="modelValue" class="kubernetes-label-dialog" :title="editing ? `编辑${title}` : title" width="60%" destroy-on-close @close="close">
    <div v-if="!editing" class="label-view">
      <div class="label-head"><strong>Key</strong><strong>Value</strong></div>
      <div v-for="row in rows" :key="row.key" class="label-view-row"><code>{{row.key}}</code><span>{{row.value||'-'}}</span></div>
      <el-empty v-if="!rows.length" description="暂无标签" :image-size="54"/>
    </div>
    <div v-else class="label-editor">
      <div class="editor-title"><span>编辑资源标签</span><strong>共 {{rows.length}} 个标签</strong></div>
      <div v-for="(row,index) in rows" :key="index" class="label-edit-row"><span class="row-index">{{index+1}}</span><el-input v-model="row.key" placeholder="Key"/><span>=</span><el-input v-model="row.value" placeholder="Value"/><el-button type="danger" plain @click="remove(index)">删除</el-button></div>
      <el-button class="add-button" plain @click="add">＋ 添加标签</el-button>
    </div>
    <template #footer><el-button @click="editing ? editing=false : close()">{{editing?'取消':'关闭'}}</el-button><el-button v-if="!editing" type="primary" @click="startEdit">编辑标签</el-button><el-button v-else type="primary" :loading="saving" @click="save">保存更改</el-button></template>
  </el-dialog>
</template>

<style scoped>
.kubernetes-label-dialog :deep(.el-dialog),.kubernetes-label-dialog :deep(.el-dialog__header),.kubernetes-label-dialog :deep(.el-dialog__body),.kubernetes-label-dialog :deep(.el-dialog__footer){background:#fff}.label-view{overflow:hidden;border:1px solid var(--line);border-radius:9px}.label-head,.label-view-row{display:grid;grid-template-columns:1fr 1fr;align-items:center;gap:20px;padding:10px 14px;border-bottom:1px solid var(--line)}.label-view-row:last-child{border-bottom:0}.label-head{color:var(--text-soft)}.label-view-row code{width:max-content;max-width:100%;padding:5px 9px;overflow-wrap:anywhere;border:1px solid #ead79b;border-radius:6px;background:#fff;color:#9a7312}.label-editor{display:flex;flex-direction:column;gap:12px}.editor-title{display:flex;align-items:center;justify-content:space-between;padding:14px 16px;border:1px solid var(--line);border-radius:9px;background:#fff}.editor-title strong{color:#a77910}.label-edit-row{display:grid;grid-template-columns:30px minmax(0,1fr) auto minmax(0,1fr) auto;align-items:center;gap:10px;padding:12px;border:1px solid var(--line);border-radius:8px;background:#fff}.row-index{display:flex;align-items:center;justify-content:center;width:26px;height:26px;border-radius:6px;background:#f5f6fa;font-weight:700}.add-button{width:100%;background:#fff}@media(max-width:720px){.kubernetes-label-dialog:deep(.el-dialog){width:90%!important}.label-edit-row{grid-template-columns:28px 1fr}.label-edit-row>span:nth-of-type(2){display:none}.label-edit-row .el-button{grid-column:2}}
</style>
