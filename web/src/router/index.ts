import { useAuthStore } from '@/stores/auth'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true, guestOnly: true } },
  { path: '/401', name: 'unauthorized', component: () => import('@/views/HttpErrorView.vue'), props: { code: 401 }, meta: { public: true, title: '登录失效' } },
  { path: '/403', name: 'forbidden', component: () => import('@/views/HttpErrorView.vue'), props: { code: 403 }, meta: { public: true, title: '无权限' } },
  { path: '/404', name: 'not-found', component: () => import('@/views/HttpErrorView.vue'), props: { code: 404 }, meta: { public: true, title: '页面不存在' } },
  { path: '/429', name: 'too-many-requests', component: () => import('@/views/HttpErrorView.vue'), props: { code: 429 }, meta: { public: true, title: '请求过于频繁' } },
  { path: '/assets/terminal', name: 'terminal', component: () => import('@/views/asset/TerminalView.vue'), meta: { title: 'Web终端', permission: 'POST /api/v1/assets/hosts/list' } },
  {
    path: '/',
    component: () => import('@/layouts/Layout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { title: '仪表盘', permission: 'POST /api/v1/dashboard/stats' } },
      { path: 'users', name: 'users', component: () => import('@/views/system/UsersView.vue'), meta: { title: '用户管理', permission: 'POST /api/v1/users/list' } },
      { path: 'roles', name: 'roles', component: () => import('@/views/system/RolesView.vue'), meta: { title: '角色权限', permission: 'POST /api/v1/roles/list' } },
      { path: 'organizations', name: 'organizations', component: () => import('@/views/system/OrganizationsView.vue'), meta: { title: '组织管理', permission: 'POST /api/v1/organizations/list' } },
      { path: 'assets/groups', name: 'asset-groups', component: () => import('@/views/asset/GroupsView.vue'), meta: { title: '资产分组', permission: 'POST /api/v1/assets/groups/list' } },
      { path: 'assets/credentials', name: 'credentials', component: () => import('@/views/asset/CredentialsView.vue'), meta: { title: '凭证管理', permission: 'POST /api/v1/assets/credentials/list' } },
      { path: 'assets/cloud-accounts', name: 'cloud-accounts', component: () => import('@/views/asset/CloudAccountsView.vue'), meta: { title: '云账号', permission: 'POST /api/v1/assets/cloud-accounts/list' } },
      { path: 'assets/hosts', name: 'hosts', component: () => import('@/views/asset/HostsView.vue'), meta: { title: '主机管理', permission: 'POST /api/v1/assets/hosts/list' } },
      { path: 'assets/db-instances', name: 'db-instances', component: () => import('@/views/asset/DBInstancesView.vue'), meta: { title: '数据库管理', permission: 'POST /api/v1/assets/db-instances/list' } },
      { path: 'assets/terminal-audit', name: 'terminal-audit', component: () => import('@/views/asset/TerminalAuditView.vue'), meta: { title: '终端审计', permission: 'POST /api/v1/assets/terminal-sessions/list' } },
      { path: 'monitor/domains', name: 'monitor-domains', component: () => import('@/views/monitor/DomainMonitorView.vue'), meta: { title: '域名监控', permission: 'POST /api/v1/monitor/domains/list' } },
      { path: 'monitor/channels', name: 'monitor-channels', component: () => import('@/views/monitor/AlertChannelsView.vue'), meta: { title: '告警通道', permission: 'POST /api/v1/monitor/channels/list' } },
      { path: 'monitor/receivers', name: 'monitor-receivers', component: () => import('@/views/monitor/AlertReceiversView.vue'), meta: { title: '告警接收人', permission: 'POST /api/v1/monitor/receivers/list' } },
      { path: 'monitor/logs', name: 'monitor-logs', component: () => import('@/views/monitor/AlertLogsView.vue'), meta: { title: '告警日志', permission: 'POST /api/v1/monitor/logs/list' } },
      { path: 'tasks/jobs', name: 'task-jobs', component: () => import('@/views/task/JobsView.vue'), meta: { title: '任务作业', permission: 'POST /api/v1/tasks/jobs/list' } },
      { path: 'tasks/execute', name: 'task-execute', component: () => import('@/views/task/ExecuteView.vue'), meta: { title: '执行任务', permission: 'POST /api/v1/tasks/execute' } },
      { path: 'tasks/templates', name: 'task-templates', component: () => import('@/views/task/TemplatesView.vue'), meta: { title: '模板管理', permission: 'POST /api/v1/tasks/templates/list' } },
      { path: 'tasks/distribute', name: 'task-distribute', component: () => import('@/views/task/FileDistributionView.vue'), meta: { title: '文件分发', permission: 'POST /api/v1/tasks/distribute' } },
      { path: 'kubernetes/clusters', name: 'kubernetes-clusters', component: () => import('@/views/kubernetes/ClustersView.vue'), meta: { title: 'Kubernetes 集群', permission: 'POST /api/v1/kubernetes/clusters/list' } },
      { path: 'kubernetes/nodes', name: 'kubernetes-nodes', component: () => import('@/views/kubernetes/ResourceListView.vue'), meta: { title: '节点管理', kubernetesResource: 'nodes', permission: 'POST /api/v1/kubernetes/nodes/list' } },
      { path: 'kubernetes/nodes/detail', name: 'kubernetes-node-detail', component: () => import('@/views/kubernetes/NodeDetailView.vue'), meta: { title: 'Node 详情', permission: 'POST /api/v1/kubernetes/nodes/support' } },
      { path: 'kubernetes/namespaces', name: 'kubernetes-namespaces', component: () => import('@/views/kubernetes/ResourceListView.vue'), meta: { title: '命名空间', kubernetesResource: 'namespaces', permission: 'POST /api/v1/kubernetes/namespaces/list' } },
      { path: 'kubernetes/namespaces/detail', name: 'kubernetes-namespace-detail', component: () => import('@/views/kubernetes/NamespaceDetailView.vue'), meta: { title: 'Namespace 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads', name: 'kubernetes-workloads', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '工作负载', kubernetesGroup: 'workloads', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/workloads/deployments/detail', name: 'kubernetes-deployment-detail', component: () => import('@/views/kubernetes/DeploymentDetailView.vue'), meta: { title: 'Deployment 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads/statefulsets/detail', name: 'kubernetes-statefulset-detail', component: () => import('@/views/kubernetes/StatefulSetDetailView.vue'), meta: { title: 'StatefulSet 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads/daemonsets/detail', name: 'kubernetes-daemonset-detail', component: () => import('@/views/kubernetes/StatefulSetDetailView.vue'), meta: { title: 'DaemonSet 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads/jobs/detail', name: 'kubernetes-job-detail', component: () => import('@/views/kubernetes/StatefulSetDetailView.vue'), meta: { title: 'Job 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads/cronjobs/detail', name: 'kubernetes-cronjob-detail', component: () => import('@/views/kubernetes/StatefulSetDetailView.vue'), meta: { title: 'CronJob 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/workloads/pods/detail', name: 'kubernetes-pod-detail', component: () => import('@/views/kubernetes/PodDetailView.vue'), meta: { title: 'Pod 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/storage/persistentvolumeclaims/detail', name: 'kubernetes-pvc-detail', component: () => import('@/views/kubernetes/StorageResourceDetailView.vue'), meta: { title: 'PVC 详情', kubernetesStorageResource: 'persistentvolumeclaims', permission: 'POST /api/v1/kubernetes/storage/resources/support' } },
      { path: 'kubernetes/storage/persistentvolumes/detail', name: 'kubernetes-pv-detail', component: () => import('@/views/kubernetes/StorageResourceDetailView.vue'), meta: { title: 'PV 详情', kubernetesStorageResource: 'persistentvolumes', permission: 'POST /api/v1/kubernetes/storage/resources/support' } },
      { path: 'kubernetes/storage/storageclasses/detail', name: 'kubernetes-storageclass-detail', component: () => import('@/views/kubernetes/StorageResourceDetailView.vue'), meta: { title: 'StorageClass 详情', kubernetesStorageResource: 'storageclasses', permission: 'POST /api/v1/kubernetes/storage/resources/support' } },
      { path: 'kubernetes/network/services/detail', name: 'kubernetes-service-detail', component: () => import('@/views/kubernetes/ServiceDetailView.vue'), meta: { title: 'Service 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/config/configmaps/detail', name: 'kubernetes-configmap-detail', component: () => import('@/views/kubernetes/ConfigMapDetailView.vue'), meta: { title: 'ConfigMap 详情', permission: 'POST /api/v1/kubernetes/configmaps/support' } },
      { path: 'kubernetes/config/secrets/detail', name: 'kubernetes-secret-detail', component: () => import('@/views/kubernetes/SecretDetailView.vue'), meta: { title: 'Secret 详情', permission: 'POST /api/v1/kubernetes/secrets/support' } },
      { path: 'kubernetes/access/serviceaccounts/detail', name: 'kubernetes-serviceaccount-detail', component: () => import('@/views/kubernetes/ServiceAccountDetailView.vue'), meta: { title: 'ServiceAccount 详情', permission: 'POST /api/v1/kubernetes/serviceaccounts/support' } },
      { path: 'kubernetes/access/roles/detail', name: 'kubernetes-role-detail', component: () => import('@/views/kubernetes/RoleDetailView.vue'), meta: { title: 'Role 详情', permission: 'POST /api/v1/kubernetes/roles/support' } },
      { path: 'kubernetes/access/clusterroles/detail', name: 'kubernetes-clusterrole-detail', component: () => import('@/views/kubernetes/ClusterRoleDetailView.vue'), meta: { title: 'ClusterRole 详情', permission: 'POST /api/v1/kubernetes/clusterroles/support' } },
      { path: 'kubernetes/access/clusterrolebindings/detail', name: 'kubernetes-clusterrolebinding-detail', component: () => import('@/views/kubernetes/ClusterRoleBindingDetailView.vue'), meta: { title: 'ClusterRoleBinding 详情', permission: 'POST /api/v1/kubernetes/clusterrolebindings/support' } },
      { path: 'kubernetes/access/rolebindings/detail', name: 'kubernetes-rolebinding-detail', component: () => import('@/views/kubernetes/RoleBindingDetailView.vue'), meta: { title: 'RoleBinding 详情', permission: 'POST /api/v1/kubernetes/rolebindings/support' } },
      { path: 'kubernetes/network/ingresses/detail', name: 'kubernetes-ingress-detail', component: () => import('@/views/kubernetes/IngressDetailView.vue'), meta: { title: 'Ingress 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/network/networkpolicies/detail', name: 'kubernetes-networkpolicy-detail', component: () => import('@/views/kubernetes/NetworkPolicyDetailView.vue'), meta: { title: 'NetworkPolicy 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/network/endpoints/detail', name: 'kubernetes-endpoints-detail', component: () => import('@/views/kubernetes/EndpointsDetailView.vue'), meta: { title: 'Endpoints 详情', permission: 'POST /api/v1/kubernetes/resources/get' } },
      { path: 'kubernetes/network', name: 'kubernetes-network', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '网络管理', kubernetesGroup: 'network', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/config', name: 'kubernetes-config', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '配置管理', kubernetesGroup: 'config', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/config/resourcequotas/detail', name: 'kubernetes-resourcequota-detail', component: () => import('@/views/kubernetes/ConfigResourceDetailView.vue'), meta: { title: 'ResourceQuota 详情', kubernetesConfigResource: 'resourcequotas', permission: 'POST /api/v1/kubernetes/config/resources/support' } },
      { path: 'kubernetes/config/limitranges/detail', name: 'kubernetes-limitrange-detail', component: () => import('@/views/kubernetes/ConfigResourceDetailView.vue'), meta: { title: 'LimitRange 详情', kubernetesConfigResource: 'limitranges', permission: 'POST /api/v1/kubernetes/config/resources/support' } },
      { path: 'kubernetes/config/horizontalpodautoscalers/detail', name: 'kubernetes-hpa-detail', component: () => import('@/views/kubernetes/ConfigResourceDetailView.vue'), meta: { title: 'HPA 详情', kubernetesConfigResource: 'horizontalpodautoscalers', permission: 'POST /api/v1/kubernetes/config/resources/support' } },
      { path: 'kubernetes/config/poddisruptionbudgets/detail', name: 'kubernetes-pdb-detail', component: () => import('@/views/kubernetes/ConfigResourceDetailView.vue'), meta: { title: 'PDB 详情', kubernetesConfigResource: 'poddisruptionbudgets', permission: 'POST /api/v1/kubernetes/config/resources/support' } },
      { path: 'kubernetes/storage', name: 'kubernetes-storage', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '存储管理', kubernetesGroup: 'storage', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/access', name: 'kubernetes-access', component: () => import('@/views/kubernetes/ResourceManagerView.vue'), meta: { title: '访问控制', kubernetesGroup: 'access', permission: 'POST /api/v1/kubernetes/resources/list' } },
      { path: 'kubernetes/inspections', name: 'kubernetes-inspections', component: () => import('@/views/kubernetes/ClusterInspectionView.vue'), meta: { title: '集群巡检', permission: 'POST /api/v1/kubernetes/inspections/history' } },
      { path: 'iot/users', name: 'iot-users', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: 'IOT 用户管理', iotKind: 'users', permission: 'POST /api/v1/iot/users/list' } },
      { path: 'iot/roles', name: 'iot-roles', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: 'IOT 角色管理', iotKind: 'roles', permission: 'POST /api/v1/iot/roles/list' } },
      { path: 'iot/acl-rules', name: 'iot-acl-rules', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: 'ACL 规则', iotKind: 'acl', permission: 'POST /api/v1/iot/acl-rules/list' } },
      { path: 'iot/messages', name: 'iot-messages', component: () => import('@/views/iot/IOTMessageView.vue'), meta: { title: '消息收发', permission: 'POST /api/v1/iot/messages/publish' } },
      { path: 'iot/topics', name: 'iot-topics', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: 'Topic 列表', iotKind: 'topics', permission: 'POST /api/v1/iot/topics/list' } },
      { path: 'iot/received-messages', name: 'iot-received-messages', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: '接收消息', iotKind: 'messages', permission: 'POST /api/v1/iot/messages/list' } },
      { path: 'iot/clients', name: 'iot-clients', component: () => import('@/views/iot/IOTManagementView.vue'), meta: { title: '客户端记录', iotKind: 'clients', permission: 'POST /api/v1/iot/clients/list' } },
      { path: 'audit/operations', name: 'audit-operations', component: () => import('@/views/audit/OperationLogsView.vue'), meta: { title: '操作日志', permission: 'POST /api/v1/audit/operation-logs/list' } },
      { path: 'audit/logins', name: 'audit-logins', component: () => import('@/views/audit/LoginLogsView.vue'), meta: { title: '登录日志', permission: 'POST /api/v1/audit/login-logs/list' } },
      { path: 'ssl-cert/certificates', name: 'ssl-certificates', component: () => import('@/views/ssl-cert/CertificatesView.vue'), meta: { title: '证书管理', permission: 'POST /api/v1/ssl-cert/certificates/list' } },
      { path: 'ssl-cert/dns-providers', name: 'ssl-dns-providers', component: () => import('@/views/ssl-cert/DNSProvidersView.vue'), meta: { title: 'DNS 配置', permission: 'POST /api/v1/ssl-cert/dns-providers/list' } },
      { path: 'ssl-cert/deploy-configs', name: 'ssl-deploy-configs', component: () => import('@/views/ssl-cert/DeployConfigsView.vue'), meta: { title: '部署配置', permission: 'POST /api/v1/ssl-cert/deploy-configs/list' } },
      { path: 'ssl-cert/tasks', name: 'ssl-tasks', component: () => import('@/views/ssl-cert/TasksView.vue'), meta: { title: '任务记录', permission: 'POST /api/v1/ssl-cert/tasks/list' } },
      { path: 'system-config', name: 'system-config', component: () => import('@/views/system/SystemConfigView.vue'), meta: { title: '系统配置', permission: 'POST /api/v1/system-config/get' } },
    ],
  },
  { path: '/:pathMatch(.*)*', name: 'catch-all', component: () => import('@/views/HttpErrorView.vue'), props: { code: 404 }, meta: { public: true, title: '页面不存在' } },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.guestOnly) {
    if (auth.isLoggedIn) return '/'
    return true
  }
  if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
  if (!auth.loaded) {
    try { await auth.fetchMe() } catch { auth.logout(); return { path: '/login', query: { redirect: to.fullPath } } }
  }
  if (to.meta.public) return true
  const permission = to.meta.permission as string | undefined
  if (!auth.hasPermission(permission)) return '/403'
  return true
})

export default router
