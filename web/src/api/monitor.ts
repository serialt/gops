import request from '@/utils/request'

export interface MonitorDomain { id:number; domain:string; port:number; status:string; response_time:number; ssl_valid:boolean; check_interval:number; enable_ssl:boolean; enable_alert:boolean; response_threshold:number; ssl_expiry_days:number; last_error:string; updated_at:string }
export interface MonitorDomainRequest { id?:number; domain:string; port:number; check_interval:number; enable_ssl:boolean; enable_alert:boolean; response_threshold:number; ssl_expiry_days:number }
export interface MonitorStats { total:number; normal:number; abnormal:number; paused:number }
export interface MonitorAlertChannel { id:number; name:string; channel_type:string; enabled:boolean; config:string; updated_at:string }
export interface MonitorAlertChannelRequest { id?:number; name:string; channel_type:string; enabled:boolean; config:string }
export interface MonitorAlertReceiver { id:number; name:string; email:string; phone:string; webhook:string; enabled:boolean; updated_at:string }
export interface MonitorAlertReceiverRequest { id?:number; name:string; email:string; phone:string; webhook:string; enabled:boolean }
export interface MonitorAlertLog { id:number; domain_id:number; alert_type:string; status:string; message:string; channel_type:string; error_message:string; sent_at:string }
export const listMonitorDomains=(keyword='',status='')=>request.post<MonitorDomain[]>('/monitor/domains/list',{keyword,status})
export const createMonitorDomain=(data:MonitorDomainRequest)=>request.post<MonitorDomain>('/monitor/domains/create',data)
export const updateMonitorDomain=(data:MonitorDomainRequest&{id:number})=>request.post<MonitorDomain>('/monitor/domains/update',data)
export const deleteMonitorDomain=(id:number)=>request.post<null>('/monitor/domains/delete',{id})
export const checkMonitorDomain=(id:number)=>request.post<MonitorDomain>('/monitor/domains/check',{id})
export const monitorStats=()=>request.post<MonitorStats>('/monitor/domains/stats',{})
export const listMonitorChannels=(keyword='')=>request.post<MonitorAlertChannel[]>('/monitor/channels/list',{keyword})
export const saveMonitorChannel=(data:MonitorAlertChannelRequest)=>request.post<MonitorAlertChannel>('/monitor/channels/save',data)
export const deleteMonitorChannel=(id:number)=>request.post<null>('/monitor/channels/delete',{id})
export const listMonitorReceivers=(keyword='')=>request.post<MonitorAlertReceiver[]>('/monitor/receivers/list',{keyword})
export const saveMonitorReceiver=(data:MonitorAlertReceiverRequest)=>request.post<MonitorAlertReceiver>('/monitor/receivers/save',data)
export const deleteMonitorReceiver=(id:number)=>request.post<null>('/monitor/receivers/delete',{id})
export const listMonitorLogs=()=>request.post<MonitorAlertLog[]>('/monitor/logs/list',{})
