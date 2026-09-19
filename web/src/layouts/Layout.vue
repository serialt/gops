<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { DataAnalysis, User, Key, Setting, OfficeBuilding, SwitchButton, ArrowDown, ArrowLeft, ArrowRight, Search, Close, Brush, FolderOpened, Monitor } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemConfigStore } from '@/stores/system-config'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const systemConfig = useSystemConfigStore()
const active = computed(() => route.path)

const sidebarWidth = ref(180)
const isCollapsed = ref(false)
const isDragging = ref(false)
const startX = ref(0)
const startWidth = ref(0)

const dashboardMenu = computed(() => ({ path: '/dashboard', title: '仪表盘', icon: DataAnalysis, permission: 'POST /api/v1/dashboard/stats' }))
const systemMenus = computed(() => [
  { path: '/users', title: '用户管理', icon: User, permission: 'POST /api/v1/users/list' },
  { path: '/roles', title: '角色权限', icon: Key, permission: 'POST /api/v1/roles/list' },
  { path: '/organizations', title: '组织管理', icon: OfficeBuilding, permission: 'POST /api/v1/organizations/list' },
  { path: '/system-config', title: '系统配置', icon: Setting },
].filter((item) => auth.hasPermission(item.permission)))
const assetMenus = computed(() => [
  { path: '/assets/groups', title: '资产分组', icon: FolderOpened, permission: 'POST /api/v1/assets/groups/list' },
  { path: '/assets/credentials', title: '凭证管理', icon: Key, permission: 'POST /api/v1/assets/credentials/list' },
  { path: '/assets/cloud-accounts', title: '云账号', icon: OfficeBuilding, permission: 'POST /api/v1/assets/cloud-accounts/list' },
  { path: '/assets/hosts', title: '主机管理', icon: OfficeBuilding, permission: 'POST /api/v1/assets/hosts/list' },
  { path: '/assets/db-instances', title: '数据库管理', icon: OfficeBuilding, permission: 'POST /api/v1/assets/db-instances/list' },
  { path: '/assets/terminal', title: '终端', icon: Monitor, permission: 'POST /api/v1/assets/hosts/list' },
  { path: '/assets/terminal-audit', title: '终端审计', icon: OfficeBuilding, permission: 'POST /api/v1/assets/terminal-sessions/list' },
].filter((item) => auth.hasPermission(item.permission)))
const monitorMenus = computed(() => [
  { path: '/monitor/domains', title: '域名监控', icon: Monitor, permission: 'POST /api/v1/monitor/domains/list' },
  { path: '/monitor/channels', title: '告警通道', icon: Setting, permission: 'POST /api/v1/monitor/channels/list' },
  { path: '/monitor/receivers', title: '告警接收人', icon: User, permission: 'POST /api/v1/monitor/receivers/list' },
  { path: '/monitor/logs', title: '告警日志', icon: DataAnalysis, permission: 'POST /api/v1/monitor/logs/list' },
].filter((item) => auth.hasPermission(item.permission)))
const nginxMenus = computed(() => [
  { path: '/nginx/sources', title: 'Nginx 数据源', icon: Monitor, permission: 'POST /api/v1/nginx/sources/list' },
].filter((item) => auth.hasPermission(item.permission)))
const taskMenus = computed(() => [{ path: '/tasks/jobs', title: '任务作业', icon: Monitor, permission: 'POST /api/v1/tasks/jobs/list' }].filter((item) => auth.hasPermission(item.permission)))
const kubernetesMenus = computed(() => [{ path: '/kubernetes/clusters', title: '集群管理', icon: OfficeBuilding, permission: 'POST /api/v1/kubernetes/clusters/list' }, { path: '/kubernetes/nodes', title: '节点管理', icon: Monitor, permission: 'POST /api/v1/kubernetes/nodes/list' }, { path: '/kubernetes/namespaces', title: '命名空间', icon: FolderOpened, permission: 'POST /api/v1/kubernetes/namespaces/list' }, { path: '/kubernetes/workloads', title: '工作负载', icon: DataAnalysis, permission: 'POST /api/v1/kubernetes/workloads/list' }].filter((item) => auth.hasPermission(item.permission)))
const auditMenus = computed(() => [{ path: '/audit/operations', title: '操作日志', icon: DataAnalysis, permission: 'POST /api/v1/audit/operation-logs/list' }, { path: '/audit/logins', title: '登录日志', icon: User, permission: 'POST /api/v1/audit/login-logs/list' }, { path: '/audit/data', title: '数据日志', icon: FolderOpened, permission: 'POST /api/v1/audit/data-logs/list' }].filter(item => auth.hasPermission(item.permission)))
const sslCertMenus = computed(() => [{ path: '/ssl-cert/certificates', title: '证书管理', icon: Key, permission: 'POST /api/v1/ssl-cert/certificates/list' }].filter(item => auth.hasPermission(item.permission)))

type PageTab = { path: string; title: string }
type MenuSuggestion = { value: string; path: string }

const pageTabs = ref<PageTab[]>([])
const draggedTabPath = ref<string | null>(null)
const menuSearch = ref('')
const menuSearchVisible = ref(false)
const menuItems = computed(() => [
  { path: '/dashboard', title: '仪表盘', permission: 'POST /api/v1/dashboard/stats' },
  ...systemMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...assetMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...monitorMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...nginxMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...taskMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...kubernetesMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...auditMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
  ...sslCertMenus.value.map(({ path, title, permission }) => ({ path, title, permission })),
])
const activeTab = computed(() => route.path)

watch(() => route.path, (path) => {
  const title = (route.meta.title as string) || '页面'
  if (!pageTabs.value.some((tab) => tab.path === path)) pageTabs.value.push({ path, title })
}, { immediate: true })

watch(sidebarWidth, (width) => {
  document.documentElement.style.setProperty('--app-content-offset', `${width + 2}px`)
}, { immediate: true })

function queryMenuSuggestions(query: string, callback: (results: MenuSuggestion[]) => void) {
  const keyword = query.trim().toLowerCase()
  callback(menuItems.value
    .filter((item) => !keyword || item.title.toLowerCase().includes(keyword))
    .map((item) => ({ value: item.title, path: item.path })))
}

function selectMenu(item: MenuSuggestion) {
  menuSearch.value = ''
  menuSearchVisible.value = false
  router.push(item.path)
}

function openTab(path: string) {
  router.push(path)
}

function closeTab(path: string) {
  const index = pageTabs.value.findIndex((tab) => tab.path === path)
  pageTabs.value = pageTabs.value.filter((tab) => tab.path !== path)
  if (activeTab.value === path) {
    const nextTab = pageTabs.value[index] || pageTabs.value[index - 1]
    router.push(nextTab?.path || '/dashboard')
  }
}

async function clearTabs() {
  if (pageTabs.value.length <= 1) return
  try {
    await ElMessageBox.confirm('确定清空除当前页面外的所有已打开标签吗？', '清空其他页面', {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'warning',
    })
    pageTabs.value = pageTabs.value.filter(tab => tab.path === activeTab.value)
  } catch {
    // 用户取消操作。
  }
}

function handleTabDragStart(event: DragEvent, path: string) {
  draggedTabPath.value = path
  event.dataTransfer?.setData('text/plain', path)
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function handleTabDrop(path: string) {
  const sourcePath = draggedTabPath.value
  draggedTabPath.value = null
  if (!sourcePath || sourcePath === path) return
  const sourceIndex = pageTabs.value.findIndex(tab => tab.path === sourcePath)
  const targetIndex = pageTabs.value.findIndex(tab => tab.path === path)
  if (sourceIndex < 0 || targetIndex < 0) return
  const [tab] = pageTabs.value.splice(sourceIndex, 1)
  pageTabs.value.splice(targetIndex, 0, tab)
}

function handleTabDropAtEnd() {
  const sourcePath = draggedTabPath.value
  draggedTabPath.value = null
  if (!sourcePath) return
  const sourceIndex = pageTabs.value.findIndex(tab => tab.path === sourcePath)
  if (sourceIndex < 0 || sourceIndex === pageTabs.value.length - 1) return
  const [tab] = pageTabs.value.splice(sourceIndex, 1)
  pageTabs.value.push(tab)
}

function logout() {
  auth.logout()
  router.push('/login')
}

function handlePointerDown(e: PointerEvent) {
  if (e.button !== 0) return
  e.preventDefault()
  isDragging.value = true
  startX.value = e.clientX
  startWidth.value = sidebarWidth.value
  document.addEventListener('pointermove', handlePointerMove)
  document.addEventListener('pointerup', handlePointerUp)
}

function handlePointerMove(e: PointerEvent) {
  if (!isDragging.value) return
  const delta = e.clientX - startX.value
  const maxWidth = Math.min(360, Math.max(180, window.innerWidth * 0.32))
  const newWidth = Math.min(maxWidth, Math.max(36, startWidth.value + delta))
  sidebarWidth.value = newWidth
  isCollapsed.value = newWidth <= 100
}

function handlePointerUp() {
  if (isDragging.value) {
    if (sidebarWidth.value <= 120) {
      sidebarWidth.value = 36
      isCollapsed.value = true
    } else if (startWidth.value <= 36) {
      sidebarWidth.value = 180
      isCollapsed.value = false
    } else {
      isCollapsed.value = false
    }
  }
  isDragging.value = false
  document.removeEventListener('pointermove', handlePointerMove)
  document.removeEventListener('pointerup', handlePointerUp)
}

onBeforeUnmount(handlePointerUp)
onMounted(() => systemConfig.fetchConfig())

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
  if (!isCollapsed.value && sidebarWidth.value <= 100) {
    sidebarWidth.value = 180
  } else if (isCollapsed.value) {
    sidebarWidth.value = 36
  }
}
</script>

<template>
  <el-container class="layout">
    <el-aside :width="`${sidebarWidth}px`" class="sidebar">
      <div v-if="sidebarWidth > 60" class="brand-wrap">
        <div class="brand-mark">V3</div>
        <div class="brand-name">{{ systemConfig.systemName }}</div>
      </div>

      <div class="sidebar-menu-scroll">
        <el-menu v-if="sidebarWidth > 60" :default-active="active" router class="menu">
          <el-menu-item v-if="auth.hasPermission(dashboardMenu.permission)" :index="dashboardMenu.path">
            <el-icon>
              <component :is="dashboardMenu.icon" />
            </el-icon>
            <span>{{ dashboardMenu.title }}</span>
          </el-menu-item>
          <el-sub-menu v-if="systemMenus.length" index="system-config">
            <template #title><el-icon>
                <Setting />
              </el-icon><span>系统管理</span></template>
            <el-menu-item v-for="item in systemMenus" :key="item.path" :index="item.path">
              <el-icon>
                <component :is="item.icon" />
              </el-icon>
              <span>{{ item.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-sub-menu v-if="assetMenus.length" index="assets">
            <template #title><el-icon>
                <FolderOpened />
              </el-icon><span>资产管理</span></template>
            <el-menu-item v-for="item in assetMenus" :key="item.path" :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item>
          </el-sub-menu>
          <el-sub-menu v-if="monitorMenus.length" index="monitor">
            <template #title><el-icon>
                <Monitor />
              </el-icon><span>监控管理</span></template>
            <el-menu-item v-for="item in monitorMenus" :key="item.path" :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item>
          </el-sub-menu>
          <el-sub-menu v-if="nginxMenus.length" index="nginx"><template #title><el-icon>
                <Monitor />
              </el-icon><span>Nginx 管理</span></template><el-menu-item v-for="item in nginxMenus" :key="item.path"
              :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
          <el-sub-menu v-if="taskMenus.length" index="tasks"><template #title><el-icon>
                <Monitor />
              </el-icon><span>任务中心</span></template><el-menu-item v-for="item in taskMenus" :key="item.path"
              :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
          <el-sub-menu v-if="kubernetesMenus.length" index="kubernetes"><template #title><el-icon>
                <OfficeBuilding />
              </el-icon><span>容器管理</span></template><el-menu-item v-for="item in kubernetesMenus" :key="item.path"
              :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
          <el-sub-menu v-if="auditMenus.length" index="audit"><template #title><el-icon>
                <DataAnalysis />
              </el-icon><span>审计管理</span></template><el-menu-item v-for="item in auditMenus" :key="item.path"
              :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
          <el-sub-menu v-if="sslCertMenus.length" index="ssl-cert"><template #title><el-icon>
                <Key />
              </el-icon><span>SSL 证书</span></template><el-menu-item v-for="item in sslCertMenus" :key="item.path"
              :index="item.path"><el-icon>
                <component :is="item.icon" />
              </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
        </el-menu>

        <div v-else class="menu-icons">
          <el-menu :default-active="active" :collapse="true" router class="menu-collapsed">
            <el-menu-item v-if="auth.hasPermission(dashboardMenu.permission)" :index="dashboardMenu.path">
              <el-icon>
                <component :is="dashboardMenu.icon" />
              </el-icon>
            </el-menu-item>
            <el-sub-menu v-if="systemMenus.length" index="system-config">
              <template #title><el-icon>
                  <Setting />
                </el-icon><span>系统管理</span></template>
              <el-menu-item v-for="item in systemMenus" :key="item.path" :index="item.path">
                <el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-sub-menu v-if="assetMenus.length" index="assets">
              <template #title><el-icon>
                  <FolderOpened />
                </el-icon><span>资产管理</span></template>
              <el-menu-item v-for="item in assetMenus" :key="item.path" :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item>
            </el-sub-menu>
            <el-sub-menu v-if="monitorMenus.length" index="monitor">
              <template #title><el-icon>
                  <Monitor />
                </el-icon><span>监控管理</span></template>
              <el-menu-item v-for="item in monitorMenus" :key="item.path" :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item>
            </el-sub-menu>
            <el-sub-menu v-if="nginxMenus.length" index="nginx"><template #title><el-icon>
                  <Monitor />
                </el-icon><span>Nginx 管理</span></template><el-menu-item v-for="item in nginxMenus" :key="item.path"
                :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
            <el-sub-menu v-if="taskMenus.length" index="tasks"><template #title><el-icon>
                  <Monitor />
                </el-icon><span>任务中心</span></template><el-menu-item v-for="item in taskMenus" :key="item.path"
                :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
            <el-sub-menu v-if="kubernetesMenus.length" index="kubernetes"><template #title><el-icon>
                  <OfficeBuilding />
                </el-icon><span>容器管理</span></template><el-menu-item v-for="item in kubernetesMenus" :key="item.path"
                :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
            <el-sub-menu v-if="auditMenus.length" index="audit"><template #title><el-icon>
                  <DataAnalysis />
                </el-icon><span>审计管理</span></template><el-menu-item v-for="item in auditMenus" :key="item.path"
                :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
            <el-sub-menu v-if="sslCertMenus.length" index="ssl-cert"><template #title><el-icon>
                  <Key />
                </el-icon><span>SSL 证书</span></template><el-menu-item v-for="item in sslCertMenus" :key="item.path"
                :index="item.path"><el-icon>
                  <component :is="item.icon" />
                </el-icon><span>{{ item.title }}</span></el-menu-item></el-sub-menu>
          </el-menu>
        </div>
      </div>

    </el-aside>

    <div class="divider" :class="{ dragging: isDragging }" role="separator" aria-label="拖动调整侧边栏宽度"
      @pointerdown="handlePointerDown" @dblclick="toggleCollapse">
      <button class="divider-handle" type="button" :aria-label="sidebarWidth > 60 ? '收起侧栏' : '展开侧栏'" @pointerdown.stop
        @click.stop="toggleCollapse">
        <el-icon>
          <component :is="sidebarWidth > 60 ? ArrowLeft : ArrowRight" />
        </el-icon>
      </button>
    </div>

    <el-container class="content-shell">
      <el-header class="header">
        <div class="top-navigation">
          <el-popover v-model:visible="menuSearchVisible" placement="bottom-start" :width="240" trigger="click">
            <template #reference>
              <button class="menu-search-button" type="button" title="搜索菜单" aria-label="搜索菜单">
                <el-icon>
                  <Search />
                </el-icon>
              </button>
            </template>
            <el-autocomplete v-model="menuSearch" class="menu-search-input" :fetch-suggestions="queryMenuSuggestions"
              placeholder="搜索菜单" clearable autofocus @select="selectMenu" />
          </el-popover>
          <div class="page-tabs" @dragover.prevent @drop.prevent="handleTabDropAtEnd">
            <button v-for="tab in pageTabs" :key="tab.path" class="page-tab" :class="{ active: activeTab === tab.path }"
              draggable="true" @dragstart="handleTabDragStart($event, tab.path)" @dragover.prevent
              @drop.stop.prevent="handleTabDrop(tab.path)" @dragend="draggedTabPath = null" @click="openTab(tab.path)">
              <span>{{ tab.title }}</span>
              <el-icon class="tab-close" @click.stop="closeTab(tab.path)">
                <Close />
              </el-icon>
            </button>
            <div class="tab-drop-end" aria-label="拖到末尾" @dragover.prevent @drop.stop.prevent="handleTabDropAtEnd" />
          </div>
        </div>

        <div class="header-actions">
          <el-tooltip content="清空其他已打开页面" placement="bottom">
            <button class="header-action" type="button" aria-label="清空其他已打开页面" @click="clearTabs"><el-icon>
                <Brush />
              </el-icon></button>
          </el-tooltip>
          <el-dropdown trigger="click">
            <div class="user-entry">
              <div class="avatar">{{ (auth.user?.nickname || auth.user?.username || 'A').slice(0, 1).toUpperCase() }}
              </div>
              <div class="user-meta">
                <span class="user-name">{{ auth.user?.nickname || auth.user?.username }}</span>
                <small>管理员</small>
              </div>
              <el-icon class="arrow">
                <ArrowDown />
              </el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :icon="SwitchButton" @click="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main"><router-view /></el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100%;
  background: var(--bg);
  display: flex;
}

.sidebar {
  background: #111827;
  color: #fff;
  box-shadow: none;
  padding: 12px 6px 14px;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  overflow: hidden;
  transition: width 0.2s ease;
  position: relative;
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.sidebar:has(+ .divider.dragging) {
  transition: none;
}

.brand-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 4px 10px;
  margin-bottom: 4px;
}

.brand-mark {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #5b7cfa, #7c5cff);
  box-shadow: 0 8px 18px rgba(91, 124, 250, 0.28);
  flex-shrink: 0;
}

.brand-name {
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1;
  color: #fff;
  white-space: nowrap;
}

.sidebar-menu-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  scrollbar-color: rgba(168, 179, 199, .45) transparent;
}

.sidebar-menu-scroll::-webkit-scrollbar {
  width: 4px;
}

.sidebar-menu-scroll::-webkit-scrollbar-thumb {
  border-radius: 4px;
  background: rgba(168, 179, 199, .45);
}

.menu {
  background: transparent !important;
  border-right: 0 !important;
  padding-top: 4px;
}

.menu :deep(.el-menu-item) {
  color: #a8b3c7 !important;
  background: transparent !important;
  font-weight: 500;
  border-radius: 8px;
  font-size: 12px;
  padding: 10px 12px !important;
  margin: 6px 8px !important;
  line-height: 1.5;
  transition: background-color 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}

.menu :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.08) !important;
  color: #fff !important;
}

.menu :deep(.el-menu-item.is-active) {
  color: #fff !important;
  background: linear-gradient(90deg, #536fe8, #6c5ce7) !important;
  box-shadow: 0 6px 16px rgba(83, 111, 232, 0.28);
  border-radius: 8px;
}

.menu :deep(.el-menu-item .el-icon) {
  margin-right: 8px;
  font-size: 18px;
  flex-shrink: 0;
}

.menu :deep(.el-sub-menu__title) {
  color: #a8b3c7 !important;
  height: 46px;
  line-height: 46px;
  margin: 4px 8px;
  padding: 0 12px !important;
  border-radius: 9px;
  font-size: 12px;
  font-weight: 700;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.menu :deep(.el-sub-menu__title:hover) {
  color: #fff !important;
  background: rgba(255, 255, 255, 0.08) !important;
}

.menu :deep(.el-sub-menu .el-menu) {
  background: transparent !important;
  overflow: hidden;
}

.menu :deep(.el-sub-menu .el-menu-item) {
  min-width: 0;
  margin-left: 30px !important;
  margin-right: 8px !important;
}

.menu :deep(.el-sub-menu__icon-arrow) {
  color: #7c3aed;
  scale: 1.22;
}

.menu-icons {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 4px 0;
}

.menu-collapsed {
  background: transparent !important;
  border-right: 0 !important;
  padding: 0 !important;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.menu-collapsed :deep(.el-menu-item) {
  color: #a8b3c7 !important;
  background: transparent !important;
  padding: 6px !important;
  min-height: 34px;
  width: 32px;
  border-radius: 6px;
  margin: 4px auto !important;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-collapsed :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.08) !important;
  color: #fff !important;
}

.menu-collapsed :deep(.el-menu-item.is-active) {
  color: #fff !important;
  background: linear-gradient(135deg, #536fe8, #6c5ce7) !important;
  box-shadow: 0 6px 16px rgba(83, 111, 232, 0.28);
}

.menu-collapsed :deep(.el-menu-item .el-icon) {
  margin-right: 0 !important;
  font-size: 18px;
}

.menu-collapsed :deep(.el-sub-menu__title) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  margin: 4px auto;
  padding: 0 !important;
  color: #a8b3c7 !important;
  border-radius: 6px;
}

.menu-collapsed :deep(.el-sub-menu__title:hover) {
  color: #fff !important;
  background: rgba(255, 255, 255, 0.08) !important;
}

.menu-collapsed :deep(.el-sub-menu__icon-arrow) {
  display: none;
}

.divider {
  width: 2px;
  flex: 0 0 2px;
  background: #e8e8e8;
  cursor: col-resize;
  user-select: none;
  touch-action: none;
  position: relative;
  transition: background 0.2s ease;
}

.divider::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  width: 2px;
  height: 42px;
  border-radius: 99px;
  background: #cbd5e1;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.divider::before {
  content: '';
  position: absolute;
  z-index: 1;
  top: 0;
  bottom: 0;
  left: -6px;
  width: 13px;
}

.divider:hover,
.divider.dragging {
  background: var(--primary);
}

.divider:hover::after,
.divider.dragging::after {
  opacity: 1;
}

.divider-handle {
  position: absolute;
  z-index: 2;
  top: 50%;
  left: 50%;
  width: 15px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transform: translate(-50%, -50%);
  border: 1px solid var(--line);
  border-radius: 10px;
  background: #fff;
  color: #667085;
  box-shadow: 0 3px 10px rgba(31, 42, 68, 0.12);
  cursor: pointer;
}

.divider-handle:hover {
  color: var(--primary);
  border-color: var(--primary);
}

.divider-handle :deep(.el-icon) {
  font-size: 10px;
}

.content-shell {
  --el-font-size-medium: 11px;
  --el-font-size-base: 11px;
  --el-font-size-small: 10px;
  font-size: 11px;
  width: 0;
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  background: #f7f8fc;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  height: 40px !important;
  padding: 0 10px !important;
  gap: 10px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--line);
}

.top-navigation {
  display: flex;
  align-items: center;
  align-self: center;
  gap: 8px;
  flex: 0 1 auto;
  overflow: hidden;
  min-width: 0;
  max-width: calc(100% - 180px);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  flex: 0 0 auto;
}

.menu-search-button {
  width: 30px;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 30px;
  padding: 0;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #f8fafc;
  color: var(--text);
  cursor: pointer;
}

.menu-search-button:hover {
  color: var(--primary);
  background: var(--bg-soft);
}

.menu-search-button :deep(.el-icon) {
  font-size: 16px;
}

.menu-search-input {
  width: 240px !important;
}

.menu-search-input :deep(.el-input) {
  width: 100% !important;
}

.page-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1 1 auto;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.page-tabs::-webkit-scrollbar {
  display: none;
}

.tab-drop-end {
  min-width: 32px;
  height: 30px;
  flex: 1 0 32px;
}

.page-tab {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  width: auto;
  flex: 0 0 auto;
  max-width: 180px;
  overflow: hidden;
  height: 30px;
  padding: 0 10px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #f8fafc;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  cursor: pointer;
}

.page-tab[draggable='true'] {
  cursor: grab;
}

.page-tab[draggable='true']:active {
  cursor: grabbing;
}

.page-tab.active {
  background: #fff;
  color: var(--text);
  box-shadow: 0 2px 6px rgba(31, 42, 68, 0.05);
}

.tab-close {
  color: #aeb5c0;
  font-size: 14px;
}

.tab-close:hover {
  color: var(--text);
}

.header-action {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  padding: 0;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #fff;
  color: var(--text);
  cursor: pointer;
}

.header-action:hover {
  color: var(--primary);
  border-color: var(--primary);
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-badge {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: linear-gradient(135deg, #5b7cfa, #7c5cff);
  box-shadow: 0 0 0 6px rgba(91, 124, 250, 0.12);
}

.header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}

.user-entry {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 3px 8px;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px solid var(--line);
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-entry:hover {
  box-shadow: var(--shadow);
  background: #fff;
}

.avatar {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: 600;
  font-size: 10px;
  color: #fff;
  background: linear-gradient(135deg, #5b7cfa, #7c5cff);
}

.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  font-size: 11px;
  font-weight: 600;
  color: var(--text);
}

.user-meta small {
  color: #64748b;
  font-size: 9px;
}

.arrow {
  color: #64748b;
}

.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  background: var(--bg);
  padding: 0 0 1px;
}

.main :deep(.page-container) {
  flex: 1 0 auto;
  min-height: 0;
}

@media (max-width: 768px) {
  .header {
    padding: 0 12px !important;
  }

  .user-meta,
  .user-entry .arrow {
    display: none;
  }

  .page-tab {
    width: auto;
    padding: 0 12px;
    font-size: 13px;
  }

  .top-navigation {
    max-width: calc(100% - 120px);
  }

  .menu-search-button {
    width: 48px;
    flex-basis: 48px;
  }

  .page-tab {
    padding: 0 10px;
    font-size: 13px;
  }

  .header-action {
    width: 34px;
    height: 34px;
  }
}
</style>
