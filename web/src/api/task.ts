import request from '@/utils/request'
export interface TaskJob{id:number;name:string;template_id:number;task_type:string;status:string;target_hosts:string;parameters:string;created_at:string}
export interface TaskJobRequest{id?:number;name:string;template_id:number;task_type:string;target_hosts:string;parameters:string}
export const listTaskJobs=(keyword='')=>request.post<TaskJob[]>('/tasks/jobs/list',{keyword});export const saveTaskJob=(data:TaskJobRequest)=>request.post<TaskJob>('/tasks/jobs/save',data)
