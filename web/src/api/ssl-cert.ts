import request, { type PageResponse } from '@/utils/request'

export interface SSLCertificateChain { position:number; version:number; common_name:string; subject:string; issuer:string; serial_number:string; dns_names:string[]|null; email_addresses:string[]|null; ip_addresses:string[]|null; uris:string[]|null; not_before:string|null; not_after:string|null; fingerprint:string; is_ca:boolean; basic_constraints_valid:boolean; max_path_len:number; max_path_len_zero:boolean; signature_algorithm:string; public_key_algorithm:string; public_key_bits:number; key_usage:string[]|null; ext_key_usage:string[]|null; subject_key_id:string; authority_key_id:string; ocsp_servers:string[]|null; issuer_urls:string[]|null; crl_points:string[]|null; policies:string[]|null }
export interface SSLCertificate { id:number; name:string; domain:string; san_domains:string; acme_email:string; ca_provider:string; key_algorithm:string; source_type:string; cloud_account_id:number; dns_provider_id:number; issuer:string; fingerprint:string; status:string; auto_renew:boolean; renew_days_before:number; not_before:string|null; not_after:string|null; last_renew_at:string|null; last_error:string; certificate?:string; private_key?:string; cert_chain?:string; chain_certificates?:SSLCertificateChain[]; chain_parse_error?:string }
export interface SSLCertificateForm { id?:number; name:string; domain:string; san_domains:string; acme_email:string; ca_provider:string; key_algorithm:string; source_type:string; cloud_account_id:number; dns_provider_id:number; certificate:string; private_key:string; cert_chain:string; auto_renew:boolean; renew_days_before:number }
export interface SSLCertificateApplyForm { name:string; domain:string; san_domains:string; acme_email:string; ca_provider:string; key_algorithm:string; dns_provider_id:number; auto_renew:boolean; renew_days_before:number }
export interface SSLDNSProvider { id:number; name:string; provider:string; config?:string; email:string; phone:string; enabled:boolean; last_test_at:string|null; last_test_ok:boolean }
export interface SSLDeployConfig { id:number; certificate_id:number; certificate_name:string; name:string; deploy_type:string; target_config:string; auto_deploy:boolean; enabled:boolean; last_deploy_at:string|null; last_deploy_ok:boolean; last_error:string }
export interface SSLTask { id:number; certificate_id:number; certificate_name:string; task_type:string; status:string; trigger_type:string; started_at:string|null; finished_at:string|null; error_message:string; result:string; created_at:string }

export const listCertificates = (params:{keyword:string;status:string;source_type:string;page:number;limit:number}) => request.post<PageResponse<SSLCertificate>>('/ssl-cert/certificates/list', params)
export const getCertificate = (id:number) => request.post<SSLCertificate>('/ssl-cert/certificates/get', { id })
export const saveCertificate = (data:SSLCertificateForm) => request.post<SSLCertificate>('/ssl-cert/certificates/save', data)
export const deleteCertificate = (id:number) => request.post('/ssl-cert/certificates/delete', { id })
export const renewCertificate = (id:number) => request.post<SSLTask>('/ssl-cert/certificates/renew', { id })
export const downloadCertificates = (ids:number[]) => request.post<Blob>('/ssl-cert/certificates/download', { ids }, { responseType: 'blob' })
export const applyCertificate = (data:SSLCertificateApplyForm) => request.post<SSLTask>('/ssl-cert/applications/create', data)

export const listDNSProviders = (params:{keyword:string;provider:string;page:number;limit:number}) => request.post<PageResponse<SSLDNSProvider>>('/ssl-cert/dns-providers/list', params)
export const getDNSProvider = (id:number) => request.post<SSLDNSProvider>('/ssl-cert/dns-providers/get', { id })
export const saveDNSProvider = (data:{id?:number;name:string;provider:string;config:string;email:string;phone:string;enabled:boolean}) => request.post<SSLDNSProvider>('/ssl-cert/dns-providers/save', data)
export const deleteDNSProvider = (id:number) => request.post('/ssl-cert/dns-providers/delete', { id })
export const testDNSProvider = (id:number) => request.post('/ssl-cert/dns-providers/test', { id })

export const listDeployConfigs = (params:{keyword:string;certificate_id:number;deploy_type:string;page:number;limit:number}) => request.post<PageResponse<SSLDeployConfig>>('/ssl-cert/deploy-configs/list', params)
export const getDeployConfig = (id:number) => request.post<SSLDeployConfig>('/ssl-cert/deploy-configs/get', { id })
export const saveDeployConfig = (data:{id?:number;certificate_id:number;name:string;deploy_type:string;target_config:string;auto_deploy:boolean;enabled:boolean}) => request.post<SSLDeployConfig>('/ssl-cert/deploy-configs/save', data)
export const deleteDeployConfig = (id:number) => request.post('/ssl-cert/deploy-configs/delete', { id })
export const testDeployConfig = (id:number) => request.post('/ssl-cert/deploy-configs/test', { id })
export const executeDeploy = (id:number) => request.post<SSLTask>('/ssl-cert/deploy-configs/execute', { id })

export const listSSLTasks = (params:{certificate_id:number;task_type:string;status:string;trigger_type:string;page:number;limit:number}) => request.post<PageResponse<SSLTask>>('/ssl-cert/tasks/list', params)
export const getSSLTask = (id:number) => request.post<SSLTask>('/ssl-cert/tasks/get', { id })
