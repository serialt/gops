import request from '@/utils/request'

export interface DashboardStatsResponse {
  hosts_total: number
  hosts_online: number
  hosts_offline: number
  clusters_total: number
  clusters_offline: number
  nodes_total: number
  nodes_offline: number
  pods_total: number
  pods_abnormal: number
  domains_total: number
  domains_normal: number
  domains_abnormal: number
  domains_paused: number
  domain_certificates: Array<{ domain: string; ssl_expiry: string }>
  today_operations: number
  active_alerts: number
}

export const getDashboardStats = () => request.post<DashboardStatsResponse>('/dashboard/stats', {})
