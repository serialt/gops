import request from '@/utils/request'

export interface AuditOperationLog {
  id: number
  username: string
  module: string
  action: string
  description: string
  method: string
  path: string
  status: number
  cost_time: number
  ip: string
  created_at: string
}

export interface AuditLoginLog {
  id: number
  username: string
  login_type: string
  status: string
  ip: string
  message: string
  created_at: string
}

export interface AuditDataLog {
  id: number
  username: string
  table_name: string
  record_id: number
  action: string
  old_data: string
  new_data: string
  created_at: string
}

export interface AuditOperationFilter {
  keyword?: string
  username?: string
  module?: string
  action?: string
  status?: string
  start_time?: string
  end_time?: string
}

export interface AuditLoginFilter {
  keyword?: string
  username?: string
  login_type?: string
  status?: string
  start_time?: string
  end_time?: string
}

export const listAuditOperations = (filter: AuditOperationFilter = {}) =>
  request.post<AuditOperationLog[]>('/audit/operation-logs/list', filter)

export const deleteAuditOperations = (ids: number[]) =>
  request.post<null>('/audit/operation-logs/delete', { ids })

export const listAuditLogins = (filter: AuditLoginFilter = {}) =>
  request.post<AuditLoginLog[]>('/audit/login-logs/list', filter)

export const deleteAuditLogins = (ids: number[]) =>
  request.post<null>('/audit/login-logs/delete', { ids })

export const listAuditData = (keyword = '') =>
  request.post<AuditDataLog[]>('/audit/data-logs/list', { keyword })
