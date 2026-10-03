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

export const listAssetGroups = (keyword = '', page = 1, limit = 1000, parent_id = 0) => request.post<PageResponse<AssetGroup>>('/assets/groups/list', { keyword, page, limit, parent_id })
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

export interface Host { id: number; name: string; group_id: number; host_type: string; ip: string; port: number | null; credential_id: number | null; vnc_credential_id: number | null; vnc_port: number | null; cloud_account_id: number; tags: string; description: string; status: number; last_seen_at: string; os: string; kernel: string; arch: string; cpu_info: string; cpu_cores: number; cpu_usage: number; memory_total: number; memory_used: number; memory_usage: number; disk_total: number; disk_used: number; disk_usage: number; uptime: string; hostname: string; created_at: string; updated_at: string }
export interface HostRequest { id?: number; name: string; group_id: number; host_type: string; ip: string; port: number | null; credential_id: number | null; vnc_credential_id: number | null; vnc_port: number | null; cloud_account_id: number; tags: string; description: string; status: number }
export interface HostExcelImportResult { created: number; updated: number }
export const listHosts = (keyword = '', group_id = 0, page = 1, limit = 20, status: number | null = null) => request.post<PageResponse<Host>>('/assets/hosts/list', { keyword, group_id, page, limit, status })
export const createHost = (data: HostRequest) => request.post<Host>('/assets/hosts/create', data)
export const updateHost = (data: Required<Pick<HostRequest, 'id'>> & HostRequest) => request.post<Host>('/assets/hosts/update', data)
export const testHostConnection = (data: { ip: string; port: number; credential_id: number }) => request.post<null>('/assets/hosts/test', data)
export const collectHostInfo = (id: number) => request.post<Host>('/assets/hosts/collect', { id })
export const deleteHost = (id: number) => request.post<null>('/assets/hosts/delete', { id })
export const exportHosts = (keyword = '', group_id = 0, status: number | null = null, ids: number[] = []) => request.post<Blob>('/assets/hosts/export', { keyword, group_id, status, ids }, { responseType: 'blob' })
export const downloadHostImportTemplate = () => request.post<Blob>('/assets/hosts/template', {}, { responseType: 'blob' })
export const importHosts = (file: File) => {
  const form = new FormData()
  form.append('file', file, file.name)
  return request.post<HostExcelImportResult>('/assets/hosts/import', form)
}
export interface SftpEntry { name: string; type: 'directory' | 'file'; size: number; modified: string; permissions: string }
export interface SftpTransferProgress { transferred: number; total: number; status: 'processing' | 'success' | 'failed'; message: string }
export const listSftp = (host_id: number, path: string, page = 1, limit = 1000) => request.post<PageResponse<SftpEntry>>('/assets/terminal/sftp/list', { host_id, path, page, limit })
export const uploadSftp = (host_id: number, path: string, files: File[], transfer_id: string, onProgress?: (progress: number) => void, relativePaths?: string[]) => {
  const form = new FormData()
  const fileBytes = files.reduce((total, file) => total + file.size, 0)
  form.append('host_id', String(host_id))
  form.append('path', path)
  form.append('transfer_id', transfer_id)
  files.forEach((file, index) => {
    form.append('files', file, file.name)
    form.append('relative_paths', relativePaths?.[index] || file.webkitRelativePath || file.name)
  })
  return request.post<null>('/assets/terminal/sftp/upload', form, {
    timeout: 0,
    onUploadProgress: event => {
      const total = event.total || fileBytes
      if (total > 0) onProgress?.(Math.min(100, Math.round((event.loaded / total) * 100)))
    },
  })
}
export const downloadSftp = (host_id: number, path: string, names: string[], transfer_id: string) =>
  request.post<Blob>('/assets/terminal/sftp/download', { host_id, path, names, transfer_id }, {
    responseType: 'blob',
    timeout: 0,
  })
export const getSftpTransferProgress = (transfer_id: string) =>
  request.post<SftpTransferProgress>('/assets/terminal/sftp/progress', { transfer_id })

export interface DbInstance { id: number; group_id: number; name: string; db_type: string; host: string; port: number; version: string; credential_id: number; tags: string; description: string; status: boolean; created_at: string; updated_at: string }
export interface AssetTreeNode { id: string; label: string; type: 'group' | 'host' | 'database'; count?: number; host?: Host; database?: DbInstance; children?: AssetTreeNode[] }
export const getTerminalAssetTree = () => request.post<AssetTreeNode[]>('/assets/terminal/tree', {})
export interface DatabaseTypeOption { value: string; label: string; default_port: number }
export interface DbInstanceRequest { id?: number; group_id: number; name: string; db_type: string; host: string; port: number; version: string; credential_id: number; tags: string; description: string; status: boolean }
export interface DbInstanceTestRequest { db_type: string; host: string; port: number; credential_id: number }
export interface DbInstanceTestResponse { version: string }
export interface DbInstanceExcelImportResult { created: number; updated: number }
export const listDBInstances = (keyword = '', group_id = 0, page = 1, limit = 20, db_type = '') => request.post<PageResponse<DbInstance>>('/assets/db-instances/list', { keyword, group_id, page, limit, db_type })
export const getDatabaseTypeOptions = () => request.post<DatabaseTypeOption[]>('/assets/db-instances/options', {})
export const createDBInstance = (data: DbInstanceRequest) => request.post<DbInstance>('/assets/db-instances/create', data)
export const updateDBInstance = (data: Required<Pick<DbInstanceRequest, 'id'>> & DbInstanceRequest) => request.post<DbInstance>('/assets/db-instances/update', data)
export const testDBInstanceConnection = (data: DbInstanceTestRequest) => request.post<DbInstanceTestResponse>('/assets/db-instances/test', data)
export const collectDBInstanceInfo = (id: number) => request.post<DbInstance>('/assets/db-instances/collect', { id })
export const deleteDBInstance = (id: number) => request.post<null>('/assets/db-instances/delete', { id })
export const exportDBInstances = (keyword = '', group_id = 0, db_type = '', ids: number[] = []) => request.post<Blob>('/assets/db-instances/export', { keyword, group_id, db_type, ids }, { responseType: 'blob' })
export const downloadDBInstanceImportTemplate = () => request.post<Blob>('/assets/db-instances/template', {}, { responseType: 'blob' })
export const importDBInstances = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<DbInstanceExcelImportResult>('/assets/db-instances/import', form)
}

export interface TerminalSession { id: number; user_id: number; username: string; host_id: number; host_name: string; host_address: string; db_instance_id: number; database_name: string; database_address: string; session_type: string; started_at: string; ended_at: string; duration_seconds: number; command_count: number; recording_size: number; completed: boolean }
export interface TerminalSessionRecording { id: number; recording: string }
export const listTerminalSessions = (keyword = '', session_type = '', page = 1, limit = 20) => request.post<PageResponse<TerminalSession>>('/assets/terminal-sessions/list', { keyword, session_type, page, limit })
export const getTerminalSessionRecording = (id: number) => request.post<TerminalSessionRecording>('/assets/terminal-sessions/recording', { id })
export const deleteTerminalSession = (id: number) => request.post<null>('/assets/terminal-sessions/delete', { id })
