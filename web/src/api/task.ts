import request from '@/utils/request'
import type { PageResponse } from '@/utils/request'

export interface TaskJob {
  id: number; name: string; template_id: number; task_type: string; status: string
  target_hosts: string; target_hosts_display: string; parameters: string; result: string; error_message: string
  created_by: number; created_at: string; updated_at: string
}

export interface TaskTemplate {
  id: number; name: string; code: string; description: string; content: string
  variables: string; category: string; platform: string; timeout: number
  sort_order: number; status: number; created_by: number; created_at: string; updated_at: string
}

export interface TaskTemplateRequest {
  id?: number; name: string; code: string; description: string; content: string
  variables: string; category: string; platform: string; timeout: number
  sort_order: number; status: number
}

export interface TaskHostResult {
  host_id: number; host_name: string; host_ip: string; status: string; output: string; error: string
}

export interface TaskExecuteResponse { task_id: number; results: TaskHostResult[] }

export const listTaskJobs = (keyword = '', task_type = '', status = '', page = 1, limit = 20) =>
  request.post<PageResponse<TaskJob>>('/tasks/jobs/list', { keyword, task_type, status, page, limit })

export const listTaskTemplates = (keyword = '', category = '', platform = '', status: number | null = null, page = 1, limit = 20) =>
  request.post<PageResponse<TaskTemplate>>('/tasks/templates/list', { keyword, category, platform, status, page, limit })
export const createTaskTemplate = (data: TaskTemplateRequest) => request.post<TaskTemplate>('/tasks/templates/create', data)
export const updateTaskTemplate = (data: TaskTemplateRequest & { id: number }) => request.post<TaskTemplate>('/tasks/templates/update', data)
export const deleteTaskTemplate = (id: number) => request.post<null>('/tasks/templates/delete', { id })

export const executeTask = (data: { name: string; host_ids: number[]; template_id: number; script_type: 'shell' | 'python'; content: string; timeout: number }) =>
  request.post<TaskExecuteResponse>('/tasks/execute', data, { timeout: 0 })

export const distributeTaskFiles = (hostIDs: number[], targetPath: string, files: File[]) => {
  const form = new FormData()
  form.append('host_ids', JSON.stringify(hostIDs))
  form.append('target_path', targetPath)
  files.forEach(file => form.append('files', file, file.name))
  return request.post<TaskExecuteResponse>('/tasks/distribute', form, { timeout: 0 })
}
