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
      { path: 'assets/db-instances', name: 'db-instances', component: () => import('@/views/asset/DBInstancesView.vue'), meta: { title: '数据库管理', permission: 'POST /api/v1/assets/db-instances/list' } },
      { path: 'assets/terminal', name: 'terminal', component: () => import('@/views/asset/TerminalView.vue'), meta: { title: '终端', permission: 'POST /api/v1/assets/hosts/list' } },
      { path: 'assets/terminal-audit', name: 'terminal-audit', component: () => import('@/views/asset/TerminalAuditView.vue'), meta: { title: '终端审计', permission: 'POST /api/v1/assets/terminal-sessions/list' } },
      { path: 'monitor/domains', name: 'monitor-domains', component: () => import('@/views/monitor/DomainMonitorView.vue'), meta: { title: '域名监控', permission: 'POST /api/v1/monitor/domains/list' } },
      { path: 'monitor/channels', name: 'monitor-channels', component: () => import('@/views/monitor/AlertChannelsView.vue'), meta: { title: '告警通道', permission: 'POST /api/v1/monitor/channels/list' } },
      { path: 'monitor/receivers', name: 'monitor-receivers', component: () => import('@/views/monitor/AlertReceiversView.vue'), meta: { title: '告警接收人', permission: 'POST /api/v1/monitor/receivers/list' } },
      { path: 'monitor/logs', name: 'monitor-logs', component: () => import('@/views/monitor/AlertLogsView.vue'), meta: { title: '告警日志', permission: 'POST /api/v1/monitor/logs/list' } },
      { path: 'nginx/sources', name: 'nginx-sources', component: () => import('@/views/nginx/NginxSourcesView.vue'), meta: { title: 'Nginx 数据源', permission: 'POST /api/v1/nginx/sources/list' } },
      { path: 'tasks/jobs', name: 'task-jobs', component: () => import('@/views/task/JobsView.vue'), meta: { title: '任务作业', permission: 'POST /api/v1/tasks/jobs/list' } },
      { path: 'kubernetes/clusters', name: 'kubernetes-clusters', component: () => import('@/views/kubernetes/ClustersView.vue'), meta: { title: 'Kubernetes 集群', permission: 'POST /api/v1/kubernetes/clusters/list' } },
      { path: 'kubernetes/nodes', name: 'kubernetes-nodes', component: () => import('@/views/kubernetes/ResourceListView.vue'), meta: { title: '节点管理', kubernetesResource: 'nodes', permission: 'POST /api/v1/kubernetes/nodes/list' } },
      { path: 'kubernetes/namespaces', name: 'kubernetes-namespaces', component: () => import('@/views/kubernetes/ResourceListView.vue'), meta: { title: '命名空间', kubernetesResource: 'namespaces', permission: 'POST /api/v1/kubernetes/namespaces/list' } },
      { path: 'kubernetes/workloads', name: 'kubernetes-workloads', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '工作负载', kubernetesGroup: 'workloads', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/pods', name: 'kubernetes-pods', component: () => import('@/views/kubernetes/ResourceListView.vue'), meta: { title: 'Pod 管理', kubernetesResource: 'pods', permission: 'POST /api/v1/kubernetes/pods/list' } },
      { path: 'kubernetes/network', name: 'kubernetes-network', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '网络管理', kubernetesGroup: 'network', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/config', name: 'kubernetes-config', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '配置管理', kubernetesGroup: 'config', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/storage', name: 'kubernetes-storage', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '存储管理', kubernetesGroup: 'storage', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/access', name: 'kubernetes-access', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '访问控制', kubernetesGroup: 'access', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'audit/operations', name: 'audit-operations', component: () => import('@/views/audit/AuditLogsView.vue'), meta: { title: '操作日志', audit: 'operation', permission: 'POST /api/v1/audit/operation-logs/list' } },
      { path: 'audit/logins', name: 'audit-logins', component: () => import('@/views/audit/AuditLogsView.vue'), meta: { title: '登录日志', audit: 'login', permission: 'POST /api/v1/audit/login-logs/list' } },
      { path: 'audit/data', name: 'audit-data', component: () => import('@/views/audit/AuditLogsView.vue'), meta: { title: '数据日志', audit: 'data', permission: 'POST /api/v1/audit/data-logs/list' } },
      { path: 'ssl-cert/certificates', name: 'ssl-certificates', component: () => import('@/views/ssl-cert/CertificatesView.vue'), meta: { title: '证书管理', permission: 'POST /api/v1/ssl-cert/certificates/list' } },
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
