import request from '@/utils/request'
import type { PageResponse } from '@/utils/request'
export interface TaskJob { id: number; name: string; template_id: number; task_type: string; status: string; target_hosts: string; parameters: string; created_at: string }
export interface TaskJobRequest { id?: number; name: string; template_id: number; task_type: string; target_hosts: string; parameters: string }
export const listTaskJobs = (keyword = '', page = 1, limit = 20) => request.post<PageResponse<TaskJob>>('/tasks/jobs/list', { keyword, page, limit }); export const saveTaskJob = (data: TaskJobRequest) => request.post<TaskJob>('/tasks/jobs/save', data)
