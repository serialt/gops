import request, { type PageResponse } from '@/utils/request'

export type MonitorRequestClient = 'chrome' | 'firefox' | 'edge' | 'safari' | 'curl'
export interface MonitorDomain { id: number; domain: string; port: number; request_client: MonitorRequestClient; status: string; response_time: number; ssl_valid: boolean; ssl_expiry: string; ssl_root_ca: string; check_interval: number; check_jitter: number; enable_ssl: boolean; verify_certificate: boolean; enable_alert: boolean; response_threshold: number; ssl_expiry_days: number; last_check: string; last_error: string; created_at: string; updated_at: string }
export interface MonitorDomainRequest { id?: number; domain: string; port: number; request_client: MonitorRequestClient; check_interval: number; check_jitter: number; enable_ssl: boolean; verify_certificate: boolean; enable_alert: boolean; response_threshold: number; ssl_expiry_days: number }
export interface MonitorCheckHistory { id: number; domain_id: number; status: string; response_time: number; ssl_valid: boolean; ssl_expiry: string; status_code: number; error_message: string; checked_at: string }
export interface MonitorStats { total: number; normal: number; abnormal: number; paused: number }
export interface MonitorDomainExcelImportResult { created: number; updated: number }
export type MonitorAlertChannelType = 'email' | 'dingtalk' | 'feishu' | 'wechat' | 'webhook'
export interface MonitorAlertChannelConfig { smtp_host: string; smtp_port: number; smtp_user: string; smtp_password: string; from_email: string; from_name: string; webhook_url: string; wechat_webhook: string; dingtalk_webhook: string; dingtalk_secret: string; feishu_webhook: string }
export interface MonitorAlertChannel { id: number; name: string; channel_type: MonitorAlertChannelType; enabled: boolean; config: MonitorAlertChannelConfig; created_at: string; updated_at: string }
export interface MonitorAlertChannelRequest { id?: number; name: string; channel_type: MonitorAlertChannelType; enabled: boolean; config: MonitorAlertChannelConfig }
export interface MonitorAlertReceiver { id: number; name: string; email: string; phone: string; webhook: string; wechat_id: string; dingtalk_id: string; feishu_id: string; enable_email: boolean; enable_webhook: boolean; enable_wechat: boolean; enable_dingtalk: boolean; enable_feishu: boolean; enable_system_msg: boolean; enabled: boolean; created_at: string; updated_at: string }
export interface MonitorAlertReceiverRequest { id?: number; name: string; email: string; phone: string; webhook: string; wechat_id: string; dingtalk_id: string; feishu_id: string; enable_email: boolean; enable_webhook: boolean; enable_wechat: boolean; enable_dingtalk: boolean; enable_feishu: boolean; enable_system_msg: boolean; enabled: boolean }
export interface MonitorAlertLog { id: number; domain_id: number; alert_type: string; status: string; message: string; channel_type: string; receiver_names: string; error_message: string; sent_at: string }
export interface MonitorDomainFilters { keyword?: string; domain?: string; port?: number; status?: string; ssl_status?: string; ssl_root_ca?: string; ssl_expiry_start?: string; ssl_expiry_end?: string; last_check_start?: string; last_check_end?: string }
export interface MonitorFilterOption { label: string; value: string }
export interface MonitorDomainFilterOptions { domains: string[]; ports: number[]; statuses: MonitorFilterOption[]; ssl_statuses: MonitorFilterOption[]; root_cas: string[] }
export const listMonitorDomains = (filters: MonitorDomainFilters = {}, page = 1, limit = 20) => request.post<PageResponse<MonitorDomain>>('/monitor/domains/list', { ...filters, page, limit })
export const listMonitorDomainOption = () => request.post<MonitorDomainFilterOptions>('/monitor/domains/list-option', {})
export const createMonitorDomain = (data: MonitorDomainRequest) => request.post<MonitorDomain>('/monitor/domains/create', data)
export const updateMonitorDomain = (data: MonitorDomainRequest & { id: number }) => request.post<MonitorDomain>('/monitor/domains/update', data)
export const deleteMonitorDomain = (id: number) => request.post<null>('/monitor/domains/delete', { id })
export const checkMonitorDomain = (id: number) => request.post<MonitorDomain>('/monitor/domains/check', { id })
export const listMonitorCheckHistory = (domain_id: number, page = 1, limit = 30) => request.post<PageResponse<MonitorCheckHistory>>('/monitor/domains/history', { domain_id, page, limit })
export const monitorStats = (filters: MonitorDomainFilters = {}) => request.post<MonitorStats>('/monitor/domains/stats', filters)
export const exportMonitorDomains = (keyword = '', ids: number[] = []) => request.post<Blob>('/monitor/domains/export', { keyword, status: '', ids }, { responseType: 'blob' })
export const importMonitorDomains = (file: File) => { const form = new FormData(); form.append('file', file, file.name); return request.post<MonitorDomainExcelImportResult>('/monitor/domains/import', form) }
export const listMonitorChannels = (keyword = '', page = 1, limit = 20) => request.post<PageResponse<MonitorAlertChannel>>('/monitor/channels/list', { keyword, page, limit })
export const saveMonitorChannel = (data: MonitorAlertChannelRequest) => request.post<MonitorAlertChannel>('/monitor/channels/save', data)
export const deleteMonitorChannel = (id: number) => request.post<null>('/monitor/channels/delete', { id })
export interface MonitorAlertChannelTestRequest { channel_type: MonitorAlertChannelType; config: MonitorAlertChannelConfig; recipient: string }
export const testMonitorChannel = (data: MonitorAlertChannelTestRequest) => request.post<null>('/monitor/channels/test', data)
export const listMonitorReceivers = (keyword = '', page = 1, limit = 20) => request.post<PageResponse<MonitorAlertReceiver>>('/monitor/receivers/list', { keyword, page, limit })
export const saveMonitorReceiver = (data: MonitorAlertReceiverRequest) => request.post<MonitorAlertReceiver>('/monitor/receivers/save', data)
export const deleteMonitorReceiver = (id: number) => request.post<null>('/monitor/receivers/delete', { id })
export const listMonitorLogs = (page = 1, limit = 20, keyword = '', status = '', domain_id = 0, receiver_name = '') => request.post<PageResponse<MonitorAlertLog>>('/monitor/logs/list', { page, limit, keyword, status, domain_id, receiver_name })
