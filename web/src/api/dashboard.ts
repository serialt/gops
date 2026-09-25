import request from '@/utils/request'

export interface DashboardStatsResponse {
  hosts_total: number
  hosts_online: number
  hosts_offline: number
  clusters_total: number
  nodes_total: number
  pods_total: number
  today_operations: number
  active_alerts: number
}

export const getDashboardStats = () => request.post<DashboardStatsResponse>('/dashboard/stats', {})
