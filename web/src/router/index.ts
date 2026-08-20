import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { title: '仪表盘', permission: 'GET /api/v1/dashboard/stats' } },
      { path: 'users', name: 'users', component: () => import('@/views/UsersView.vue'), meta: { title: '用户管理', permission: 'GET /api/v1/users' } },
      { path: 'roles', name: 'roles', component: () => import('@/views/RolesView.vue'), meta: { title: '角色权限', permission: 'GET /api/v1/roles' } },
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
