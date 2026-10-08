<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, type UploadFile, type UploadFiles, type UploadRawFile, type UploadUserFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { listHosts, type Host } from '@/api/asset'
import { distributeTaskFiles, type TaskHostResult } from '@/api/task'

const hosts = ref<Host[]>([])
const hostIDs = ref<number[]>([])
const targetPath = ref('/tmp')
const fileList = ref<UploadUserFile[]>([])
const results = ref<TaskHostResult[]>([])
const distributing = ref(false)

async function loadHosts() {
  const { data } = await listHosts('', 0, 1, 1000)
  hosts.value = data.data.list.filter(item => item.port && item.credential_id)
}
function filesChanged(_file: UploadFile, files: UploadFiles) { fileList.value = files }
function formatSize(size = 0) { if (size < 1024) return `${size} B`; if (size < 1048576) return `${(size / 1024).toFixed(1)} KB`; return `${(size / 1048576).toFixed(1)} MB` }
async function distribute() {
  const files = fileList.value.map(item => item.raw).filter((item): item is UploadRawFile => item !== undefined)
  if (!files.length) return ElMessage.warning('请选择要分发的文件')
  if (!hostIDs.value.length) return ElMessage.warning('请选择目标主机')
  if (!targetPath.value.trim()) return ElMessage.warning('请输入目标路径')
  distributing.value = true
  results.value = []
  try {
    const { data } = await distributeTaskFiles(hostIDs.value, targetPath.value, files)
    results.value = data.data.results
    ElMessage.success(results.value.every(item => item.status === 'success') ? '文件分发完成' : '文件分发完成，部分主机失败')
  } finally { distributing.value = false }
}
onMounted(() => { void loadHosts() })
</script>

<template>
  <section class="task-page">
    <header><h2>文件分发</h2><p>通过 SFTP 将文件批量分发到主机指定目录。</p></header>
    <div class="content-grid">
      <div class="task-card">
        <div class="card-title">上传文件</div>
        <el-upload drag multiple :auto-upload="false" :file-list="fileList" :on-change="filesChanged" :on-remove="filesChanged">
          <el-icon class="upload-icon"><UploadFilled /></el-icon><div>拖拽文件到此处，或点击选择文件</div>
        </el-upload>
        <div v-if="fileList.length" class="file-summary"><span v-for="item in fileList" :key="item.uid">{{ item.name }} · {{ formatSize(item.size) }}</span></div>
      </div>
      <div class="task-card target-card">
        <div class="card-title">分发目标</div>
        <el-form label-width="70px">
          <el-form-item label="目标路径" required><el-input v-model="targetPath" placeholder="例如 /tmp/packages" /></el-form-item>
          <el-form-item label="目标主机" required><el-select v-model="hostIDs" multiple filterable collapse-tags :max-collapse-tags="4" placeholder="选择已配置 SSH 的主机"><el-option v-for="item in hosts" :key="item.id" :label="`${item.name}（${item.ip}:${item.port}）`" :value="item.id" /></el-select></el-form-item>
        </el-form>
        <div class="submit-row"><el-button type="primary" :loading="distributing" @click="distribute">开始分发</el-button></div>
      </div>
    </div>
    <div class="task-card result-card"><div class="card-title">分发结果</div><el-empty v-if="!results.length" description="暂无分发结果" :image-size="52" /><div v-else class="results"><div v-for="item in results" :key="item.host_id" class="result-row"><span>{{ item.host_name }}（{{ item.host_ip }}）</span><span class="message">{{ item.error || item.output }}</span><el-tag :type="item.status === 'success' ? 'success' : 'danger'" effect="plain">{{ item.status === 'success' ? '成功' : '失败' }}</el-tag></div></div></div>
  </section>
</template>

<style scoped>
.task-page{height:100%;display:flex;flex-direction:column;gap:8px;min-width:0;overflow:auto;padding:0 4px;box-sizing:border-box;font-size:11px}header h2{margin:0;color:var(--text);font-size:20px}header p{margin:3px 0 0;color:var(--text-secondary)}.content-grid{display:grid;grid-template-columns:1fr 1fr;gap:8px}.task-card{border:1px solid var(--line);border-radius:10px;background:#fff;padding:12px;min-width:0}.card-title{font-size:13px;font-weight:700;margin-bottom:10px}.upload-icon{font-size:34px;color:var(--primary)}.file-summary{display:flex;flex-direction:column;gap:4px;margin-top:8px;color:var(--text-secondary)}.target-card{display:flex;flex-direction:column}.el-select{width:100%}.submit-row{display:flex;justify-content:flex-end;margin-top:auto}.result-card{min-height:150px}.results{display:grid;gap:6px}.result-row{display:grid;grid-template-columns:minmax(160px,240px) 1fr auto;align-items:center;gap:10px;border-bottom:1px solid var(--line);padding:7px 2px}.message{color:var(--text-secondary);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}@media(max-width:800px){.content-grid{grid-template-columns:1fr}.result-row{grid-template-columns:1fr auto}.message{grid-column:1/-1}}
.task-page :deep(.el-input__inner::placeholder),.task-page :deep(.el-textarea__inner::placeholder),.task-page :deep(.el-select__placeholder.is-transparent){font-size:10px}
</style>
