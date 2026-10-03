<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listHosts, type Host } from '@/api/asset'
import { executeTask, listTaskTemplates, type TaskHostResult, type TaskTemplate } from '@/api/task'

const hosts = ref<Host[]>([])
const templates = ref<TaskTemplate[]>([])
const results = ref<TaskHostResult[]>([])
const executing = ref(false)
const form = reactive({ name: '', host_ids: [] as number[], template_id: null as number | null, script_type: 'shell' as 'shell' | 'python', content: '', timeout: 300 })

async function loadOptions() {
  const [hostResponse, templateResponse] = await Promise.all([listHosts('', 0, 1, 1000), listTaskTemplates('', '', '', 1, 1, 1000)])
  hosts.value = hostResponse.data.data.list.filter(item => item.port && item.credential_id)
  templates.value = templateResponse.data.data.list
}

function selectTemplate(id: number | null) {
  const template = templates.value.find(item => item.id === id)
  if (!template) return
  form.content = template.content
  form.timeout = template.timeout || 300
  form.script_type = template.category === 'python' ? 'python' : 'shell'
}

async function run() {
  if (!form.host_ids.length) return ElMessage.warning('请选择目标主机')
  if (!form.content.trim()) return ElMessage.warning('请输入执行内容')
  executing.value = true
  results.value = []
  try {
    const { data } = await executeTask({ ...form, template_id: form.template_id || 0 })
    results.value = data.data.results
    ElMessage.success(results.value.every(item => item.status === 'success') ? '任务执行完成' : '任务执行完成，部分主机失败')
  } finally { executing.value = false }
}

onMounted(() => { void loadOptions() })
</script>

<template>
  <section class="task-page">
    <header><div><h2>执行任务</h2><p>在多台主机上执行 Shell 或 Python 脚本。</p></div></header>
    <div class="task-card form-card">
      <el-form label-width="76px">
        <div class="form-row">
          <el-form-item label="任务名称"><el-input v-model="form.name" placeholder="可选，默认按执行时间生成" /></el-form-item>
          <el-form-item label="执行模板">
            <el-select v-model="form.template_id" clearable filterable placeholder="选择执行模板（可选）" popper-class="task-template-popper" @change="selectTemplate">
              <el-option v-for="item in templates" :key="item.id" :label="item.name" :value="item.id" class="template-option-item">
                <div class="template-option-main"><strong>{{ item.name }}</strong><span>{{ item.category === 'python' ? 'Python' : 'Shell' }}</span><span v-if="item.platform">{{ item.platform }}</span></div>
                <div class="template-option-description">{{ item.description || item.code }}</div>
              </el-option>
            </el-select>
          </el-form-item>
        </div>
        <el-form-item label="目标主机" required><el-select v-model="form.host_ids" multiple filterable collapse-tags :max-collapse-tags="4" placeholder="选择已配置 SSH 的主机"><el-option v-for="item in hosts" :key="item.id" :label="`${item.name}（${item.ip}:${item.port}）`" :value="item.id" /></el-select></el-form-item>
        <div class="form-row compact-row">
          <el-form-item label="脚本类型"><el-radio-group v-model="form.script_type"><el-radio-button value="shell">Shell</el-radio-button><el-radio-button value="python">Python</el-radio-button></el-radio-group></el-form-item>
          <el-form-item label="超时时间"><el-input-number v-model="form.timeout" :min="1" :max="3600" /><span class="unit">秒</span></el-form-item>
        </div>
        <el-form-item label="执行内容" required><el-input v-model="form.content" type="textarea" :rows="13" resize="vertical" placeholder="请输入要执行的脚本内容" class="code-input" /></el-form-item>
        <div class="submit-row"><el-button type="primary" :loading="executing" @click="run">开始执行</el-button></div>
      </el-form>
    </div>
    <div class="task-card result-card">
      <div class="card-title">执行结果</div>
      <el-empty v-if="!results.length" description="暂无执行结果" :image-size="52" />
      <div v-else class="results">
        <article v-for="item in results" :key="item.host_id" class="result-item">
          <div class="result-head"><strong>{{ item.host_name }}（{{ item.host_ip }}）</strong><el-tag :type="item.status === 'success' ? 'success' : 'danger'" effect="plain">{{ item.status === 'success' ? '成功' : '失败' }}</el-tag></div>
          <pre>{{ item.error || item.output || '执行完成，无输出' }}</pre>
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped>
.task-page{height:100%;display:flex;flex-direction:column;gap:8px;min-width:0;overflow:auto;padding:0 4px;box-sizing:border-box;font-size:11px}header h2{margin:0;color:var(--text);font-size:20px}header p{margin:3px 0 0;color:var(--text-secondary)}.task-card{border:1px solid var(--line);border-radius:10px;background:#fff;padding:12px}.form-card{flex:none}.form-row{display:grid;grid-template-columns:1fr 1fr;gap:12px}.compact-row{grid-template-columns:minmax(280px,1fr) minmax(220px,1fr)}.el-form-item{margin-bottom:10px}.el-select{width:100%}.unit{margin-left:6px;color:var(--text-secondary)}.code-input :deep(textarea){font-family:ui-monospace,SFMono-Regular,Menlo,monospace}.submit-row{display:flex;justify-content:flex-end}.result-card{min-height:150px}.card-title{font-size:13px;font-weight:700;margin-bottom:8px}.results{display:grid;gap:8px}.result-item{border:1px solid var(--line);border-radius:8px;overflow:hidden}.result-head{display:flex;justify-content:space-between;align-items:center;padding:8px 10px;background:#f8f9fc}.result-item pre{margin:0;padding:10px;min-height:48px;max-height:220px;overflow:auto;background:#16181d;color:#e7eaf0;white-space:pre-wrap}@media(max-width:760px){.form-row{grid-template-columns:1fr}}
:global(.task-template-popper .el-select-dropdown__item.template-option-item){height:auto;min-height:48px;margin:3px 6px;padding:7px 10px;border:1px solid transparent;border-radius:8px;line-height:normal}:global(.task-template-popper .el-select-dropdown__item.template-option-item:hover){border-color:#d8dbff;background:#f7f7ff}:global(.task-template-popper .el-select-dropdown__item.template-option-item.is-selected){border-color:#cbd0ff;background:#f1f2ff}:global(.task-template-popper .template-option-main){display:flex;align-items:center;gap:6px}:global(.task-template-popper .template-option-main strong){margin-right:auto;color:var(--text);font-size:12px}:global(.task-template-popper .template-option-main span){border:1px solid var(--line);border-radius:6px;padding:2px 6px;background:#f6f7fb;color:var(--text-secondary);font-size:10px}:global(.task-template-popper .template-option-description){margin-top:4px;overflow:hidden;color:var(--text-secondary);font-size:10px;text-overflow:ellipsis;white-space:nowrap}:global(.task-template-popper .el-select-dropdown__item.is-selected .template-option-main strong){color:var(--primary)}
.task-page :deep(.el-input__inner::placeholder),.task-page :deep(.el-textarea__inner::placeholder),.task-page :deep(.el-select__placeholder.is-transparent){font-size:10px}
</style>
