import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/Layout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { title: '仪表盘', permission: 'POST /api/v1/dashboard/stats' } },
      { path: 'users', name: 'users', component: () => import('@/views/UsersView.vue'), meta: { title: '用户管理', permission: 'POST /api/v1/users/list' } },
      { path: 'roles', name: 'roles', component: () => import('@/views/RolesView.vue'), meta: { title: '角色权限', permission: 'POST /api/v1/roles/list' } },
      { path: 'organizations', name: 'organizations', component: () => import('@/views/OrganizationsView.vue'), meta: { title: '组织管理', permission: 'POST /api/v1/organizations/list' } },
      { path: 'assets/groups', name: 'asset-groups', component: () => import('@/views/asset/GroupsView.vue'), meta: { title: '资产分组', permission: 'POST /api/v1/assets/groups/list' } },
      { path: 'assets/credentials', name: 'credentials', component: () => import('@/views/asset/CredentialsView.vue'), meta: { title: '凭证管理', permission: 'POST /api/v1/assets/credentials/list' } },
      { path: 'assets/cloud-accounts', name: 'cloud-accounts', component: () => import('@/views/asset/CloudAccountsView.vue'), meta: { title: '云账号', permission: 'POST /api/v1/assets/cloud-accounts/list' } },
      { path: 'assets/hosts', name: 'hosts', component: () => import('@/views/asset/HostsView.vue'), meta: { title: '主机管理', permission: 'POST /api/v1/assets/hosts/list' } },
      { path: 'assets/db-instances', name: 'db-instances', component: () => import('@/views/asset/DBInstancesView.vue'), meta: { title: '数据库实例', permission: 'POST /api/v1/assets/db-instances/list' } },
      { path: 'assets/terminal', name: 'terminal', component: () => import('@/views/asset/TerminalView.vue'), meta: { title: '终端', permission: 'POST /api/v1/assets/hosts/list' } },
      { path: 'assets/terminal-audit', name: 'terminal-audit', component: () => import('@/views/asset/TerminalAuditView.vue'), meta: { title: '终端审计', permission: 'POST /api/v1/assets/terminal-sessions/list' } },
      { path: 'system-config', name: 'system-config', component: () => import('@/views/SystemConfigView.vue'), meta: { title: '系统配置' } },
      { path: '403', name: 'forbidden', component: () => import('@/views/ForbiddenView.vue'), meta: { title: '无权限' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) return auth.isLoggedIn ? '/' : true
  if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
  if (!auth.loaded) {
    try { await auth.fetchMe() } catch { auth.logout(); return '/login' }
  }
  const permission = to.meta.permission as string | undefined
  if (!auth.hasPermission(permission)) return '/403'
  return true
})

export default router
