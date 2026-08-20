<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataAnalysis, User, Key, SwitchButton, ArrowDown, ArrowLeft } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const active = computed(() => route.path)

const sidebarWidth = ref(56)
const isCollapsed = ref(false)
const isDragging = ref(false)
const startX = ref(0)
const startWidth = ref(0)

const menus = computed(() => [
  { path: '/dashboard', title: '仪表盘', icon: DataAnalysis, permission: 'GET /api/v1/dashboard/stats' },
  { path: '/users', title: '用户管理', icon: User, permission: 'GET /api/v1/users' },
  { path: '/roles', title: '角色权限', icon: Key, permission: 'GET /api/v1/roles' },
].filter((item) => auth.hasPermission(item.permission)))

function logout() {
  auth.logout()
  router.push('/login')
}

function handleMouseDown(e: MouseEvent) {
  isDragging.value = true
  startX.value = e.clientX
  startWidth.value = sidebarWidth.value
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

function handleMouseMove(e: MouseEvent) {
  if (!isDragging.value) return
  const delta = e.clientX - startX.value
  const newWidth = Math.max(80, startWidth.value + delta)
  sidebarWidth.value = newWidth
  if (newWidth > 100) {
    isCollapsed.value = false
  }
}

function handleMouseUp() {
  isDragging.value = false
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
}

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
  if (!isCollapsed.value && sidebarWidth.value <= 56) {
    sidebarWidth.value = 240
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
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>

      <div v-else class="menu-icons">
        <el-menu :default-active="active" router class="menu-collapsed">
          <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
          </el-menu-item>
        </el-menu>
      </div>

      <div class="collapse-btn" @click="toggleCollapse" :title="sidebarWidth > 70 ? '收起' : '展开'">
        <el-icon><ArrowLeft /></el-icon>
      </div>
    </el-aside>

    <div class="divider" @mousedown="handleMouseDown" :class="{ dragging: isDragging }" />

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
  background: #1a1a1a;
  color: #fff;
  box-shadow: none;
  padding: 12px 6px 14px;
  border-right: 1px solid #333;
  overflow: hidden;
  transition: width 0.2s ease;
  position: relative;
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
  background: var(--primary);
  box-shadow: none;
  flex-shrink: 0;
}

.brand-name {
  font-size: 16px;
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
  color: #999 !important;
  background: transparent !important;
  font-weight: 500;
  border-radius: 8px;
  font-size: 14px;
  padding: 10px 12px !important;
  margin: 6px 8px !important;
  line-height: 1.5;
  transition: all 0.2s ease;
}

.menu :deep(.el-menu-item:hover) {
  background: #f0f0f0 !important;
  color: #333 !important;
}

.menu :deep(.el-menu-item.is-active) {
  color: #fff !important;
  background: #0ABAB5 !important;
  box-shadow: 0 2px 8px rgba(10, 186, 181, 0.3);
  border-radius: 8px;
}

.menu :deep(.el-menu-item .el-icon) {
  margin-right: 8px;
  font-size: 18px;
  flex-shrink: 0;
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
  color: #999 !important;
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
  background: #f0f0f0 !important;
  color: #333 !important;
}

.menu-collapsed :deep(.el-menu-item.is-active) {
  color: #fff !important;
  background: #0ABAB5 !important;
  box-shadow: 0 2px 8px rgba(10, 186, 181, 0.3);
}

.menu-collapsed :deep(.el-menu-item .el-icon) {
  margin-right: 0 !important;
  font-size: 18px;
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
  background: #333;
  border-radius: 6px;
  cursor: pointer;
  color: #888;
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  background: #444;
  color: #fff;
}

.divider {
  width: 4px;
  background: #e8e8e8;
  cursor: col-resize;
  user-select: none;
  transition: background 0.2s ease;
  hover: background #0ABAB5;
}

.divider:hover,
.divider.dragging {
  background: #0ABAB5;
}

.content-shell {
  min-width: 0;
  background: #ffffff;
  flex: 1;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px !important;
  padding: 0 24px !important;
  background: #ffffff;
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
  background: linear-gradient(135deg, #4ade80, #16a34a);
  box-shadow: 0 0 0 6px rgba(74, 222, 128, 0.12);
}

.header-title {
  font-size: 18px;
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
  background: #f0f0f0;
  border: 1px solid var(--line);
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-entry:hover {
  box-shadow: var(--shadow);
  background: #e8e8e8;
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
  background: var(--primary);
}

.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  font-size: 13px;
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
  background: var(--bg);
  padding: 20px 24px;
}
</style>
