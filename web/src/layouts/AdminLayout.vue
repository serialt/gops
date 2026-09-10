<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataAnalysis, User, Key, Setting, OfficeBuilding, SwitchButton, ArrowDown, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
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
].filter((item) => auth.hasPermission(item.permission)))

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
  const newWidth = Math.min(maxWidth, Math.max(56, startWidth.value + delta))
  sidebarWidth.value = newWidth
  isCollapsed.value = newWidth <= 70
}

function handlePointerUp() {
  isDragging.value = false
  document.removeEventListener('pointermove', handlePointerMove)
  document.removeEventListener('pointerup', handlePointerUp)
}

onBeforeUnmount(handlePointerUp)

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
  if (!isCollapsed.value && sidebarWidth.value <= 100) {
    sidebarWidth.value = 180
  } else if (isCollapsed.value) {
    sidebarWidth.value = 56
  }
}
</script>

<template>
  <el-container class="layout">
    <el-aside :width="`${sidebarWidth}px`" class="sidebar">
      <div v-if="sidebarWidth > 70" class="brand-wrap">
        <div class="brand-mark">V3</div>
        <div class="brand-name">Admin</div>
      </div>

      <div v-if="sidebarWidth > 70" class="nav-group-label">导航</div>
      <el-menu v-if="sidebarWidth > 70" :default-active="active" router class="menu">
        <el-menu-item v-if="auth.hasPermission(dashboardMenu.permission)" :index="dashboardMenu.path">
          <el-icon><component :is="dashboardMenu.icon" /></el-icon>
          <span>{{ dashboardMenu.title }}</span>
        </el-menu-item>
        <el-sub-menu v-if="systemMenus.length" index="system-config">
          <template #title><el-icon><Setting /></el-icon><span>系统配置</span></template>
          <el-menu-item v-for="item in systemMenus" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>

      <div v-else class="menu-icons">
        <el-menu :default-active="active" :collapse="true" router class="menu-collapsed">
          <el-menu-item v-if="auth.hasPermission(dashboardMenu.permission)" :index="dashboardMenu.path">
            <el-icon><component :is="dashboardMenu.icon" /></el-icon>
          </el-menu-item>
          <el-sub-menu v-if="systemMenus.length" index="system-config">
            <template #title><el-icon><Setting /></el-icon><span>系统配置</span></template>
            <el-menu-item v-for="item in systemMenus" :key="item.path" :index="item.path">
              <el-icon><component :is="item.icon" /></el-icon><span>{{ item.title }}</span>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </div>

      <div class="collapse-btn" @click="toggleCollapse" :title="sidebarWidth > 70 ? '收起' : '展开'">
        <el-icon><component :is="sidebarWidth > 70 ? ArrowLeft : ArrowRight" /></el-icon>
      </div>
    </el-aside>

    <div
      class="divider"
      :class="{ dragging: isDragging }"
      role="separator"
      aria-label="拖动调整侧边栏宽度"
      @pointerdown="handlePointerDown"
    />

    <el-container class="content-shell">
      <el-header class="header">
        <div class="header-title-wrap">
          <span class="header-badge" />
          <div class="header-title">{{ route.meta.title }}</div>
        </div>

        <el-dropdown trigger="click">
          <div class="user-entry">
            <div class="avatar">{{ (auth.user?.nickname || auth.user?.username || 'A').slice(0, 1).toUpperCase() }}</div>
            <div class="user-meta">
              <span class="user-name">{{ auth.user?.nickname || auth.user?.username }}</span>
              <small>管理员</small>
            </div>
            <el-icon class="arrow"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="SwitchButton" @click="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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

.nav-group-label {
  margin: 12px 16px 8px;
  font-size: 12px;
  letter-spacing: 0;
  text-transform: none;
  color: #888;
  font-weight: 500;
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
  padding: 8px !important;
  min-height: 40px;
  width: 40px;
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

.collapse-btn {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #536fe8, #6c5ce7);
  border-radius: 6px;
  cursor: pointer;
  color: #fff;
  box-shadow: 0 6px 14px rgba(83, 111, 232, 0.26);
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  background: linear-gradient(135deg, #405bcc, #5b4bd4);
  color: #fff;
  box-shadow: 0 8px 18px rgba(83, 111, 232, 0.36);
}

.divider {
  width: 6px;
  flex: 0 0 6px;
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
  left: 1px;
  width: 4px;
  height: 42px;
  border-radius: 99px;
  background: #cbd5e1;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.divider:hover,
.divider.dragging {
  background: var(--primary);
}

.divider:hover::after,
.divider.dragging::after {
  opacity: 1;
}

.content-shell {
  --el-font-size-medium: 11px;
  --el-font-size-base: 11px;
  --el-font-size-small: 10px;
  font-size: 11px;
  width: 0;
  flex: 1 1 auto;
  min-width: 0;
  background: #f7f8fc;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 68px !important;
  padding: 0 24px !important;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--line);
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
  gap: 10px;
  height: 40px;
  padding: 6px 12px;
  border-radius: 6px;
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
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: 600;
  font-size: 12px;
  color: #fff;
  background: linear-gradient(135deg, #5b7cfa, #7c5cff);
}

.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
}

.user-meta small {
  color: #64748b;
}

.arrow {
  color: #64748b;
}

.main {
  min-width: 0;
  overflow-x: hidden;
  overflow-y: auto;
  background: var(--bg);
  padding: 0;
}
</style>
