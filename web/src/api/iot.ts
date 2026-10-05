import request from '@/utils/request'
import type { PageResponse } from '@/utils/request'
export interface IOTACLRule { id:number; role_id:number; role_name:string; operation:string; effect:string; topic_filter:string; created_at:string }
export interface IOTTopic { id:number; topic:string; created_at:string; updated_at:string }
export interface IOTMessage { id:number; received_at:string; client_id:string; username:string; topic:string; payload:string; qos:number; retained:boolean; duplicate:boolean }
export interface IOTClient { id:number; client_id:string; status:string; username:string; remote_address:string; listener:string; protocol_version:number; clean_session:boolean; connected_at:string; disconnected_at?:string; disconnect_reason:string; session_expired:boolean }
export interface IOTDashboard { topic_total:number; message_total:number; client_total:number; online_total:number; recent_messages:IOTMessage[]; recent_clients:IOTClient[] }
export interface IOTBrokerOption { key:string; name:string; url:string; type:string; connected:boolean; internal:boolean }
export interface IOTExternalBroker { id:number; name:string; url:string; client_id:string; username:string; topic_filter:string; enabled:boolean; connected:boolean; has_password:boolean; insecure_skip_verify:boolean; created_at:string; updated_at:string }
export type IOTListParams={page:number;limit:number;keyword?:string;status?:string;role_id?:number}
export const listIOTACLRules=(data:IOTListParams)=>request.post<PageResponse<IOTACLRule>>('/iot/acl-rules/list',data)
export const createIOTACLRule=(data:any)=>request.post<IOTACLRule>('/iot/acl-rules/create',data);export const updateIOTACLRule=(data:any)=>request.post<null>('/iot/acl-rules/update',data);export const deleteIOTACLRule=(id:number)=>request.post<null>('/iot/acl-rules/delete',{id})
export const listIOTTopics=(data:IOTListParams)=>request.post<PageResponse<IOTTopic>>('/iot/topics/list',data)
export const listIOTMessages=(data:IOTListParams)=>request.post<PageResponse<IOTMessage>>('/iot/messages/list',data)
export const listIOTClients=(data:IOTListParams)=>request.post<PageResponse<IOTClient>>('/iot/clients/list',data)
export const publishIOTMessage=(data:{broker:string;topic:string;payload:string;qos:number;retain:boolean})=>request.post<null>('/iot/messages/publish',data)
export const getIOTBrokerOptions=()=>request.post<IOTBrokerOption[]>('/iot/brokers/options',{})
export const listIOTExternalBrokers=(data:{page:number;limit:number;keyword?:string;enabled?:boolean})=>request.post<PageResponse<IOTExternalBroker>>('/iot/brokers/list',data)
export const createIOTExternalBroker=(data:Record<string,unknown>)=>request.post<null>('/iot/brokers/create',data)
export const updateIOTExternalBroker=(data:Record<string,unknown>)=>request.post<null>('/iot/brokers/update',data)
export const deleteIOTExternalBroker=(id:number)=>request.post<null>('/iot/brokers/delete',{id})
export const getIOTDashboard=()=>request.post<IOTDashboard>('/iot/dashboard/stats',{})
