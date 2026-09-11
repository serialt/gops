import request from '@/utils/request'

export interface DashboardStatsResponse { users: number; enabled_users: number; roles: number; policies: number }

export const getDashboardStats = () => request.post<DashboardStatsResponse>('/dashboard/stats', {})
