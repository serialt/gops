import request from '@/utils/request'
import type { PageResponse } from '@/utils/request'

export interface AssetGroup {
  id: number
  name: string
  code: string
  parent_id: number
  description: string
  sort_order: number
  status: boolean
  created_at: string
  updated_at: string
}

export interface AssetGroupListRequest { keyword: string }
export interface AssetGroupRequest {
  id?: number
  name: string
  code: string
  parent_id: number
  description: string
  sort_order: number
  status: boolean
}
export interface AssetGroupIDRequest { id: number }

export const listAssetGroups = (keyword = '', page = 1, limit = 1000) => request.post<PageResponse<AssetGroup>>('/assets/groups/list', { keyword, page, limit })
export const createAssetGroup = (data: AssetGroupRequest) => request.post<AssetGroup>('/assets/groups/create', data)
export const updateAssetGroup = (data: Required<Pick<AssetGroupRequest, 'id'>> & AssetGroupRequest) => request.post<AssetGroup>('/assets/groups/update', data)
export const deleteAssetGroup = (id: number) => request.post<null>('/assets/groups/delete', { id } satisfies AssetGroupIDRequest)

export interface Credential { id: number; name: string; credential_type: string; username: string; description: string; created_at: string; updated_at: string }
export interface CredentialRequest { id?: number; name: string; credential_type: string; username: string; password: string; private_key: string; passphrase: string; description: string }
export const listCredentials = (keyword = '', page = 1, limit = 20) => request.post<PageResponse<Credential>>('/assets/credentials/list', { keyword, page, limit })
export const createCredential = (data: CredentialRequest) => request.post<Credential>('/assets/credentials/create', data)
export const updateCredential = (data: Required<Pick<CredentialRequest, 'id'>> & CredentialRequest) => request.post<Credential>('/assets/credentials/update', data)
export const deleteCredential = (id: number) => request.post<null>('/assets/credentials/delete', { id })

export interface CloudAccount { id: number; name: string; provider: string; region: string; description: string; status: boolean; created_at: string; updated_at: string }
export interface CloudRegionOption { value: string; label: string }
export interface CloudProviderOption { value: string; label: string; regions: CloudRegionOption[] }
export interface CloudAccountRequest { id?: number; name: string; provider: string; access_key: string; secret_key: string; region: string; description: string; status: boolean }
export const listCloudAccounts = (keyword = '', page = 1, limit = 20) => request.post<PageResponse<CloudAccount>>('/assets/cloud-accounts/list', { keyword, page, limit })
export const getCloudProviderOptions = () => request.post<CloudProviderOption[]>('/assets/cloud-accounts/options', {})
export const createCloudAccount = (data: CloudAccountRequest) => request.post<CloudAccount>('/assets/cloud-accounts/create', data)
export const updateCloudAccount = (data: Required<Pick<CloudAccountRequest, 'id'>> & CloudAccountRequest) => request.post<CloudAccount>('/assets/cloud-accounts/update', data)
export const deleteCloudAccount = (id: number) => request.post<null>('/assets/cloud-accounts/delete', { id })

export interface Host { id: number; name: string; group_id: number; host_type: string; ip: string; port: number; credential_id: number; cloud_account_id: number; tags: string; description: string; status: number; last_seen_at: string; os: string; kernel: string; arch: string; cpu_info: string; cpu_cores: number; cpu_usage: number; memory_total: number; memory_used: number; memory_usage: number; disk_total: number; disk_used: number; disk_usage: number; uptime: string; hostname: string; created_at: string; updated_at: string }
export interface HostRequest { id?: number; name: string; group_id: number; host_type: string; ip: string; port: number; credential_id: number; cloud_account_id: number; tags: string; description: string; status: number }
export const listHosts = (keyword = '', group_id = 0, page = 1, limit = 20) => request.post<PageResponse<Host>>('/assets/hosts/list', { keyword, group_id, page, limit })
export const createHost = (data: HostRequest) => request.post<Host>('/assets/hosts/create', data)
export const updateHost = (data: Required<Pick<HostRequest, 'id'>> & HostRequest) => request.post<Host>('/assets/hosts/update', data)
export const testHostConnection = (data: { ip: string; port: number; credential_id: number }) => request.post<null>('/assets/hosts/test', data)
export const collectHostInfo = (id: number) => request.post<Host>('/assets/hosts/collect', { id })
export const deleteHost = (id: number) => request.post<null>('/assets/hosts/delete', { id })
export interface SftpEntry { name: string; type: 'directory' | 'file'; size: number; modified: string; permissions: string }
export const listSftp = (host_id: number, path: string) => request.post<SftpEntry[]>('/assets/terminal/sftp/list', { host_id, path })
export const uploadSftp = (host_id: number, path: string, files: File[]) => {
  const form = new FormData()
  form.append('host_id', String(host_id))
  form.append('path', path)
  files.forEach(file => form.append('files', file, file.webkitRelativePath || file.name))
  return request.post<null>('/assets/terminal/sftp/upload', form)
}
export const downloadSftp = (host_id: number, path: string, names: string[]) =>
  request.post<Blob>('/assets/terminal/sftp/download', { host_id, path, names }, { responseType: 'blob' })

export interface DbInstance { id: number; group_id: number; name: string; db_type: string; host: string; port: number; version: string; credential_id: number; tags: string; description: string; status: boolean; created_at: string; updated_at: string }
export interface AssetTreeNode { id: string; label: string; type: 'group' | 'host' | 'database'; count?: number; host?: Host; database?: DbInstance; children?: AssetTreeNode[] }
export const getTerminalAssetTree = () => request.post<AssetTreeNode[]>('/assets/terminal/tree', {})
export interface DatabaseTypeOption { value: string; label: string; default_port: number }
export interface DbInstanceRequest { id?: number; group_id: number; name: string; db_type: string; host: string; port: number; version: string; credential_id: number; tags: string; description: string; status: boolean }
export interface DbInstanceTestRequest { db_type: string; host: string; port: number; credential_id: number }
export interface DbInstanceTestResponse { version: string }
export const listDBInstances = (keyword = '', group_id = 0, page = 1, limit = 20) => request.post<PageResponse<DbInstance>>('/assets/db-instances/list', { keyword, group_id, page, limit })
export const getDatabaseTypeOptions = () => request.post<DatabaseTypeOption[]>('/assets/db-instances/options', {})
export const createDBInstance = (data: DbInstanceRequest) => request.post<DbInstance>('/assets/db-instances/create', data)
export const updateDBInstance = (data: Required<Pick<DbInstanceRequest, 'id'>> & DbInstanceRequest) => request.post<DbInstance>('/assets/db-instances/update', data)
export const testDBInstanceConnection = (data: DbInstanceTestRequest) => request.post<DbInstanceTestResponse>('/assets/db-instances/test', data)
export const collectDBInstanceInfo = (id: number) => request.post<DbInstance>('/assets/db-instances/collect', { id })
export const deleteDBInstance = (id: number) => request.post<null>('/assets/db-instances/delete', { id })

export interface TerminalSession { id: number; user_id: number; username: string; host_id: number; host_name: string; host_address: string; db_instance_id: number; database_name: string; database_address: string; session_type: string; started_at: string; ended_at: string; duration_seconds: number; command_count: number; recording_size: number; completed: boolean }
export interface TerminalSessionRecording { id: number; recording: string }
export const listTerminalSessions = (keyword = '', session_type = '') => request.post<TerminalSession[]>('/assets/terminal-sessions/list', { keyword, session_type })
export const getTerminalSessionRecording = (id: number) => request.post<TerminalSessionRecording>('/assets/terminal-sessions/recording', { id })
export const deleteTerminalSession = (id: number) => request.post<null>('/assets/terminal-sessions/delete', { id })
