import request from '@/utils/request'

export type MonitorRequestClient = 'chrome' | 'firefox' | 'edge' | 'safari' | 'curl'
export interface MonitorDomain { id:number; domain:string; port:number; request_client:MonitorRequestClient; status:string; response_time:number; ssl_valid:boolean; ssl_expiry:string; ssl_root_ca:string; check_interval:number; check_jitter:number; enable_ssl:boolean; verify_certificate:boolean; enable_alert:boolean; response_threshold:number; ssl_expiry_days:number; last_check:string; last_error:string; created_at:string; updated_at:string }
export interface MonitorDomainRequest { id?:number; domain:string; port:number; request_client:MonitorRequestClient; check_interval:number; check_jitter:number; enable_ssl:boolean; verify_certificate:boolean; enable_alert:boolean; response_threshold:number; ssl_expiry_days:number }
export interface MonitorCheckHistory { id:number; domain_id:number; status:string; response_time:number; ssl_valid:boolean; ssl_expiry:string; status_code:number; error_message:string; checked_at:string }
export interface MonitorStats { total:number; normal:number; abnormal:number; paused:number }
export interface MonitorDomainExcelExport { file_name:string; content:string }
export interface MonitorDomainExcelImportResult { created:number; updated:number }
export type MonitorAlertChannelType = 'email' | 'dingtalk' | 'feishu' | 'wechat' | 'webhook'
export interface MonitorAlertChannelConfig { smtp_host:string; smtp_port:number; smtp_user:string; smtp_password:string; from_email:string; from_name:string; webhook_url:string; wechat_webhook:string; dingtalk_webhook:string; dingtalk_secret:string; feishu_webhook:string }
export interface MonitorAlertChannel { id:number; name:string; channel_type:MonitorAlertChannelType; enabled:boolean; config:MonitorAlertChannelConfig; created_at:string; updated_at:string }
export interface MonitorAlertChannelRequest { id?:number; name:string; channel_type:MonitorAlertChannelType; enabled:boolean; config:MonitorAlertChannelConfig }
export interface MonitorAlertReceiver { id:number; name:string; email:string; phone:string; webhook:string; wechat_id:string; dingtalk_id:string; feishu_id:string; enable_email:boolean; enable_webhook:boolean; enable_wechat:boolean; enable_dingtalk:boolean; enable_feishu:boolean; enable_system_msg:boolean; enabled:boolean; created_at:string; updated_at:string }
export interface MonitorAlertReceiverRequest { id?:number; name:string; email:string; phone:string; webhook:string; wechat_id:string; dingtalk_id:string; feishu_id:string; enable_email:boolean; enable_webhook:boolean; enable_wechat:boolean; enable_dingtalk:boolean; enable_feishu:boolean; enable_system_msg:boolean; enabled:boolean }
export interface MonitorAlertLog { id:number; domain_id:number; alert_type:string; status:string; message:string; channel_type:string; error_message:string; sent_at:string }
export const listMonitorDomains=(keyword='',status='')=>request.post<MonitorDomain[]>('/monitor/domains/list',{keyword,status})
export const createMonitorDomain=(data:MonitorDomainRequest)=>request.post<MonitorDomain>('/monitor/domains/create',data)
export const updateMonitorDomain=(data:MonitorDomainRequest&{id:number})=>request.post<MonitorDomain>('/monitor/domains/update',data)
export const deleteMonitorDomain=(id:number)=>request.post<null>('/monitor/domains/delete',{id})
export const checkMonitorDomain=(id:number)=>request.post<MonitorDomain>('/monitor/domains/check',{id})
export const listMonitorCheckHistory=(domain_id:number)=>request.post<MonitorCheckHistory[]>('/monitor/domains/history',{domain_id})
export const monitorStats=()=>request.post<MonitorStats>('/monitor/domains/stats',{})
export const exportMonitorDomains=(keyword='')=>request.post<MonitorDomainExcelExport>('/monitor/domains/export',{keyword,status:''})
export const importMonitorDomains=(file_name:string,content:string)=>request.post<MonitorDomainExcelImportResult>('/monitor/domains/import',{file_name,content})
export const listMonitorChannels=(keyword='')=>request.post<MonitorAlertChannel[]>('/monitor/channels/list',{keyword})
export const saveMonitorChannel=(data:MonitorAlertChannelRequest)=>request.post<MonitorAlertChannel>('/monitor/channels/save',data)
export const deleteMonitorChannel=(id:number)=>request.post<null>('/monitor/channels/delete',{id})
export interface MonitorAlertChannelTestRequest { channel_type:MonitorAlertChannelType; config:MonitorAlertChannelConfig; recipient:string }
export const testMonitorChannel=(data:MonitorAlertChannelTestRequest)=>request.post<null>('/monitor/channels/test',data)
export const listMonitorReceivers=(keyword='')=>request.post<MonitorAlertReceiver[]>('/monitor/receivers/list',{keyword})
export const saveMonitorReceiver=(data:MonitorAlertReceiverRequest)=>request.post<MonitorAlertReceiver>('/monitor/receivers/save',data)
export const deleteMonitorReceiver=(id:number)=>request.post<null>('/monitor/receivers/delete',{id})
export const listMonitorLogs=()=>request.post<MonitorAlertLog[]>('/monitor/logs/list',{})
