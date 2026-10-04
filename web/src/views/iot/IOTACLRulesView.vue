<script setup lang="ts">
import { h, onBeforeUnmount, reactive, ref, render, watch, type Component } from 'vue'
import { ElMessage, ElMessageBox, ElTooltip } from 'element-plus'
import { Delete, Edit } from '@element-plus/icons-vue'
import type { ColDef } from 'ag-grid-community'
import AppDataGrid from '@/components/AppDataGrid.vue'
import { createIOTACLRule, deleteIOTACLRule, listIOTACLRules, updateIOTACLRule, type IOTACLRule } from '@/api/iot'
import { getRoles, type RoleResponse } from '@/api/role'

const rows = ref<IOTACLRule[]>([])
const selectedRows = ref<IOTACLRule[]>([])
const roles = ref<RoleResponse[]>([])
const keyword = ref('')
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const loading = ref(false)
const dialog = ref(false)
const verifyDialog = ref(false)
const saving = ref(false)
const verifying = ref(false)
const editing = ref(false)
const verification = ref<{ allowed: boolean; message: string } | null>(null)
const verifyForm = reactive({ role_id: 0, operation: 'publish', effect: 'allow', topic: '' })
const form = reactive<any>({
  id: 0,
  role_id: 0,
  operation: 'publish',
  effect: 'allow',
  topic_filter: '#',
  publish_allow_enabled: true,
  publish_allow_topic_filter: '#',
  publish_deny_enabled: false,
  publish_deny_topic_filter: '#',
  subscribe_allow_enabled: true,
  subscribe_allow_topic_filter: '#',
  subscribe_deny_enabled: false,
  subscribe_deny_topic_filter: '#',
})
let timer: ReturnType<typeof setTimeout> | undefined

function sequence(params: any) {
  return (page.value - 1) * limit.value + params.node.rowIndex + 1
}

function topicOverlaps(left: string, right: string) {
  const leftParts = left.split('/')
  const rightParts = right.split('/')
  for (let index = 0; ; index += 1) {
    if (leftParts[index] === '#' || rightParts[index] === '#') return true
    if (index === leftParts.length || index === rightParts.length) return index === leftParts.length && index === rightParts.length
    if (leftParts[index] !== '+' && rightParts[index] !== '+' && leftParts[index] !== rightParts[index]) return false
  }
}

function topicCovered(allowed: string, requested: string) {
  const allowedParts = allowed.split('/')
  const requestedParts = requested.split('/')
  for (let index = 0; index < allowedParts.length; index += 1) {
    const part = allowedParts[index]
    if (part === '#') return true
    if (index >= requestedParts.length) return false
    if (part === '+') {
      if (requestedParts[index] === '#') return false
      continue
    }
    if (requestedParts[index] === '+' || requestedParts[index] === '#' || part !== requestedParts[index]) return false
  }
  return allowedParts.length === requestedParts.length
}

function handleSelectionChanged(selected: unknown[]) {
  selectedRows.value = selected as IOTACLRule[]
}

function actionButton(label: string, kind: string, icon: Component, onClick: () => void) {
  const root = document.createElement('span')
  render(h(ElTooltip, { content: label, placement: 'top', showAfter: 250 }, {
    default: () => h('button', {
      type: 'button',
      class: `grid-action compact icon-only ${kind}`,
      'aria-label': label,
      onClick,
    }, [h('span', { class: 'el-icon' }, [h(icon)])]),
  }), root)
  return root
}

function actions(params: any) {
  const container = document.createElement('div')
  container.className = 'grid-action-group'
  container.append(
    actionButton('编辑', 'edit', Edit, () => openEdit(params.data)),
    actionButton('删除', 'delete', Delete, () => void remove(params.data)),
  )
  return container
}

const columns: ColDef[] = [
  { headerName: '序号', valueGetter: sequence, width: 62 },
  { field: 'role_name', headerName: '系统角色' },
  { field: 'operation', headerName: '规则类型', valueFormatter: (params) => params.value === 'publish' ? '发布' : '订阅' },
  { field: 'effect', headerName: '效果', valueFormatter: (params) => params.value === 'allow' ? '允许' : '拒绝' },
  { field: 'topic_filter', headerName: 'Topic 过滤器' },
  { headerName: '操作', width: 150, minWidth: 150, maxWidth: 150, pinned: 'right', resizable: false, sortable: false, cellRenderer: actions },
]

async function loadRoles() {
  const { data } = await getRoles(1, 1000)
  roles.value = data.data.list
}

async function load() {
  loading.value = true
  try {
    const { data } = await listIOTACLRules({ page: page.value, limit: limit.value, keyword: keyword.value.trim() })
    rows.value = data.data.list
    total.value = data.data.total
    await loadRoles()
  } finally {
    loading.value = false
  }
}

function reset() {
  Object.assign(form, {
    id: 0,
    role_id: roles.value[0]?.id || 0,
    operation: 'publish',
    effect: 'allow',
    topic_filter: '#',
    publish_allow_enabled: true,
    publish_allow_topic_filter: '#',
    publish_deny_enabled: false,
    publish_deny_topic_filter: '#',
    subscribe_allow_enabled: true,
    subscribe_allow_topic_filter: '#',
    subscribe_deny_enabled: false,
    subscribe_deny_topic_filter: '#',
  })
}

function openCreate() {
  editing.value = false
  reset()
  dialog.value = true
}

function openVerify() {
  if (!roles.value.length) {
    ElMessage.warning('暂无可验证的角色')
    return
  }
  verifyForm.role_id = roles.value[0].id
  verifyForm.operation = 'publish'
  verifyForm.effect = 'allow'
  verifyForm.topic = ''
  verification.value = null
  verifyDialog.value = true
}

async function verifyACL() {
  const topic = verifyForm.topic.trim()
  if (!topic) {
    ElMessage.warning('请输入 Topic')
    return
  }
  verifying.value = true
  try {
    const firstPage = await listIOTACLRules({ page: 1, limit: 1000, role_id: verifyForm.role_id })
    const { list, total: ruleCount } = firstPage.data.data
    const rules = [...list]
    const pageCount = Math.ceil(ruleCount / 1000)
    for (let currentPage = 2; currentPage <= pageCount; currentPage += 1) {
      const response = await listIOTACLRules({ page: currentPage, limit: 1000, role_id: verifyForm.role_id })
      rules.push(...response.data.data.list)
    }
    const matchingRules = rules.filter((rule) => rule.operation === verifyForm.operation)
    const denied = matchingRules.some((rule) => rule.effect === 'deny' && topicOverlaps(rule.topic_filter, topic))
    const allowed = matchingRules.some((rule) => rule.effect === 'allow' && topicCovered(rule.topic_filter, topic))
    const isAllowed = !denied && allowed
    const selectedEffectMatched = verifyForm.effect === 'deny' ? denied : allowed
    let message = '未命中允许规则，最终拒绝。'
    if (denied) message = '命中拒绝规则，拒绝优先，最终拒绝。'
    else if (allowed) message = '命中允许规则，最终允许。'
    if (!selectedEffectMatched) message = `${message} 未命中所选的${verifyForm.effect === 'deny' ? '拒绝' : '允许'}规则。`
    verification.value = { allowed: isAllowed, message }
  } finally {
    verifying.value = false
  }
}

function openEdit(row: IOTACLRule) {
  editing.value = true
  reset()
  Object.assign(form, row)
  dialog.value = true
}

async function save() {
  saving.value = true
  try {
    if (editing.value) {
      await updateIOTACLRule({ ...form })
    } else {
      const rules: { operation: string; effect: string; topic_filter: string; enabledKey: string }[] = []
      if (form.publish_deny_enabled) rules.push({ operation: 'publish', effect: 'deny', topic_filter: form.publish_deny_topic_filter, enabledKey: 'publish_deny_enabled' })
      if (form.publish_allow_enabled) rules.push({ operation: 'publish', effect: 'allow', topic_filter: form.publish_allow_topic_filter, enabledKey: 'publish_allow_enabled' })
      if (form.subscribe_deny_enabled) rules.push({ operation: 'subscribe', effect: 'deny', topic_filter: form.subscribe_deny_topic_filter, enabledKey: 'subscribe_deny_enabled' })
      if (form.subscribe_allow_enabled) rules.push({ operation: 'subscribe', effect: 'allow', topic_filter: form.subscribe_allow_topic_filter, enabledKey: 'subscribe_allow_enabled' })
      if (!rules.length) {
        ElMessage.warning('请至少启用一条规则')
        return
      }
      const results = await Promise.allSettled(rules.map(({ enabledKey, ...rule }) => createIOTACLRule({ role_id: form.role_id, ...rule })))
      rules.forEach((rule, index) => {
        if (results[index].status === 'fulfilled') form[rule.enabledKey] = false
      })
      if (results.some((result) => result.status === 'rejected')) {
        await load()
        ElMessage.error('部分规则创建失败，可重试未完成的规则')
        return
      }
    }
    dialog.value = false
    ElMessage.success('保存成功')
    await load()
  } finally {
    saving.value = false
  }
}

async function remove(row: IOTACLRule) {
  await ElMessageBox.confirm('删除后不可恢复，确定继续吗？', '删除确认', { type: 'warning' })
  await deleteIOTACLRule(row.id)
  ElMessage.success('删除成功')
  await load()
}

watch(keyword, () => {
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => {
    page.value = 1
    void load()
  }, 300)
})
watch([page, limit], () => { void load() }, { immediate: true })
onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <section v-loading="loading" class="iot-page">
    <header>
      <div>
        <h1>ACL 规则</h1>
        <p>IOT MQTT ACL 规则。</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" plain @click="openVerify">验证 ACL</el-button>
        <el-button type="primary" @click="openCreate">新增</el-button>
      </div>
    </header>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索 Topic 过滤器" clearable />
    </div>
    <AppDataGrid
      class="grid"
      :row-data="rows"
      :column-defs="columns"
      pagination
      server-pagination
      :total="total"
      :current-page="page"
      :page-size="limit"
      @selection-changed="handleSelectionChanged"
      @page-change="(value) => { page = value }"
      @page-size-change="(value) => { limit = value; page = 1 }"
    />
    <el-dialog v-model="dialog" :title="editing ? '编辑' : '新增'" width="560px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="系统角色">
          <el-select v-model="form.role_id">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <template v-if="editing">
          <el-form-item label="当前规则">
            <div class="current-rule">
              <el-tag>{{ form.operation === 'publish' ? '发布' : '订阅' }}</el-tag>
              <el-tag :type="form.effect === 'deny' ? 'danger' : 'success'">{{ form.effect === 'deny' ? '拒绝' : '允许' }}</el-tag>
            </div>
          </el-form-item>
          <el-form-item label="规则类型">
            <el-radio-group v-model="form.operation">
              <el-radio-button value="publish">发布</el-radio-button>
              <el-radio-button value="subscribe">订阅</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="效果">
            <el-radio-group v-model="form.effect">
              <el-radio-button value="allow">允许</el-radio-button>
              <el-radio-button value="deny">拒绝</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="Topic 过滤器">
            <el-input v-model="form.topic_filter" placeholder="例如 devices/+/status 或 #" />
          </el-form-item>
        </template>
        <template v-else>
        <section class="rule-section">
          <h2>发布规则</h2>
          <div class="effect-rule">
            <el-checkbox v-model="form.publish_deny_enabled">拒绝</el-checkbox>
            <el-input v-if="form.publish_deny_enabled" v-model="form.publish_deny_topic_filter" placeholder="拒绝发布的 Topic 过滤器" />
          </div>
          <div class="effect-rule">
            <el-checkbox v-model="form.publish_allow_enabled">允许</el-checkbox>
            <el-input v-if="form.publish_allow_enabled" v-model="form.publish_allow_topic_filter" placeholder="允许发布的 Topic 过滤器" />
          </div>
        </section>
        <section class="rule-section">
          <h2>订阅规则</h2>
          <div class="effect-rule">
            <el-checkbox v-model="form.subscribe_deny_enabled">拒绝</el-checkbox>
            <el-input v-if="form.subscribe_deny_enabled" v-model="form.subscribe_deny_topic_filter" placeholder="拒绝订阅的 Topic 过滤器" />
          </div>
          <div class="effect-rule">
            <el-checkbox v-model="form.subscribe_allow_enabled">允许</el-checkbox>
            <el-input v-if="form.subscribe_allow_enabled" v-model="form.subscribe_allow_topic_filter" placeholder="允许订阅的 Topic 过滤器" />
          </div>
        </section>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="verifyDialog" title="验证 ACL" width="480px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="系统角色">
          <el-select v-model="verifyForm.role_id">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则类型">
          <el-radio-group v-model="verifyForm.operation">
            <el-radio-button value="publish">发布</el-radio-button>
            <el-radio-button value="subscribe">订阅</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="验证效果">
          <el-radio-group v-model="verifyForm.effect">
            <el-radio-button value="deny">拒绝</el-radio-button>
            <el-radio-button value="allow">允许</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Topic">
          <el-input v-model="verifyForm.topic" placeholder="输入要验证的 Topic" @keyup.enter="verifyACL">
            <template #append>
              <el-button type="primary" :loading="verifying" @click="verifyACL">验证</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-alert
          v-if="verification"
          :title="verification.message"
          :type="verification.allowed ? 'success' : 'error'"
          :closable="false"
          show-icon
        />
      </el-form>
    </el-dialog>
  </section>
</template>

<style scoped>
.iot-page { height: 100%; min-width: 0; display: flex; flex-direction: column; gap: 8px; padding: 0 8px; box-sizing: border-box; font-size: 11px; }
.iot-page header { display: flex; align-items: center; justify-content: space-between; }
.header-actions { display: flex; align-items: center; gap: 4px; }
.header-actions :deep(.el-button + .el-button) { margin-left: 0; }
.iot-page h1 { margin: 0; font-size: 18px; }
.iot-page header p { margin: 3px 0 0; color: var(--text-secondary); }
.toolbar { display: flex; gap: 7px; }
.toolbar .el-input { width: 210px; }
.grid { flex: 1; min-height: 0; }
.current-rule { display: flex; align-items: center; gap: 6px; }
.rule-section { padding: 8px 10px; border: 1px solid var(--line); border-radius: 6px; margin-bottom: 10px; }
.rule-section h2 { margin: 0 0 8px; font-size: 12px; font-weight: 600; }
.effect-rule { display: grid; grid-template-columns: 72px minmax(0, 1fr); gap: 8px; align-items: center; margin-bottom: 8px; }
</style>
