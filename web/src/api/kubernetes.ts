import request from '@/utils/request'
export interface KubernetesCluster{id:number;name:string;alias:string;api_endpoint:string;version:string;status:number;node_count:number;pod_count:number;region:string;provider:string;description:string;created_at:string;updated_at:string}
export interface KubernetesClusterRequest{id?:number;name:string;alias:string;api_endpoint:string;kube_config:string;region:string;provider:string;description:string;status:number}
export const listKubernetesClusters=(keyword='')=>request.post<KubernetesCluster[]>('/kubernetes/clusters/list',{keyword});export const saveKubernetesCluster=(data:KubernetesClusterRequest)=>request.post<KubernetesCluster>('/kubernetes/clusters/save',data);export const deleteKubernetesCluster=(id:number)=>request.post<null>('/kubernetes/clusters/delete',{id})
export const testKubernetesCluster=(id:number)=>request.post<{status:string;version:string}>('/kubernetes/clusters/test',{id})
export const syncKubernetesClusters=()=>request.post<{total:number;success:number;failed:number;errors:string[]}>('/kubernetes/clusters/sync',{})
export const getKubernetesClusterConfig=(id:number)=>request.post<{kube_config:string}>('/kubernetes/clusters/config',{id})
export interface KubernetesTaint{key:string;value:string;effect:string};export interface KubernetesNode{name:string;status:string;roles:string;age:string;version:string;internalIP:string;externalIP?:string;osImage:string;kernelVersion:string;containerRuntime:string;labels:Record<string,string>;cpuCapacity:string;memoryCapacity:string;podCount:number;podCapacity:number;schedulable:boolean;taintCount:number;taints:KubernetesTaint[];created_at:string};export interface KubernetesNamespace{name:string;status:string;age:string;labels:Record<string,string>;created_at:string};export interface KubernetesWorkload{namespace:string;name:string;kind:string;ready:string;status:string;created_at:string}
export interface KubernetesWorkloadStats{deployment:number;stateful_set:number;daemon_set:number;job:number;cron_job:number;pod:number}
export interface KubernetesPod{name:string;namespace:string;ready:string;status:string;restarts:number;age:string;ip:string;node:string}
export const listKubernetesNodes=(cluster_id:number,keyword='',status='',role='')=>request.post<KubernetesNode[]>('/kubernetes/nodes/list',{cluster_id,keyword,status,role});export const listKubernetesNamespaces=(cluster_id:number,keyword='',status='')=>request.post<KubernetesNamespace[]>('/kubernetes/namespaces/list',{cluster_id,keyword,status});export const listKubernetesWorkloads=(cluster_id:number,keyword='',namespace='')=>request.post<KubernetesWorkload[]>('/kubernetes/workloads/list',{cluster_id,keyword,namespace})
export const getKubernetesWorkloadStats=(cluster_id:number,namespace='')=>request.post<KubernetesWorkloadStats>('/kubernetes/workloads/stats',{cluster_id,namespace})
export const updateKubernetesNode=(data:{cluster_id:number;name:string;labels:Record<string,string>;schedulable:boolean;taints?:KubernetesTaint[]})=>request.post<KubernetesNode>('/kubernetes/nodes/update',data)
export const deleteKubernetesNode=(cluster_id:number,name:string)=>request.post<null>('/kubernetes/nodes/delete',{cluster_id,name})
export const drainKubernetesNode=(cluster_id:number,name:string)=>request.post<null>('/kubernetes/nodes/drain',{cluster_id,name})
export const updateKubernetesNamespace=(data:{cluster_id:number;name:string;labels:Record<string,string>})=>request.post<KubernetesNamespace>('/kubernetes/namespaces/update',data)
export const listKubernetesPods=(cluster_id:number,keyword='',namespace='')=>request.post<KubernetesPod[]>('/kubernetes/pods/list',{cluster_id,keyword,namespace})
export interface KubernetesClusterStats {nodeCount:number;workloadCount:number;podCount:number;cpuUsage:number;memoryUsage:number;cpuCapacity:number;memoryCapacity:number;cpuAllocatable:number;memoryAllocatable:number;cpuUsed:number;memoryUsed:number}
export const getKubernetesClusterStats=(cluster_id:number)=>request.post<KubernetesClusterStats>('/kubernetes/stats/get',{cluster_id})
export interface KubernetesResourceValues {cpu:string;memory:string}
export interface KubernetesObject {name:string;namespace:string;kind:string;status:string;age:string;labels:Record<string,string>;created_at:string;ready?:string;images?:string[];requests?:KubernetesResourceValues;limits?:KubernetesResourceValues;schedule?:string;last_schedule_time?:string;suspended?:boolean;restarts?:number;pod_ip?:string;node?:string}
export interface KubernetesObjectRequest {cluster_id:number;resource:string;namespace?:string;name?:string;keyword?:string;yaml?:string}
export const listKubernetesObjects=(data:KubernetesObjectRequest)=>request.post<KubernetesObject[]>('/kubernetes/resources/list',data)
export const getKubernetesObject=(data:KubernetesObjectRequest)=>request.post<{yaml:string}>('/kubernetes/resources/get',data)
export const getKubernetesResourceTemplate=(data:KubernetesObjectRequest)=>request.post<{yaml:string}>('/kubernetes/resources/template',data)
export const applyKubernetesObject=(data:KubernetesObjectRequest)=>request.post<KubernetesObject>('/kubernetes/resources/apply',data)
export const deleteKubernetesObject=(data:KubernetesObjectRequest)=>request.post<null>('/kubernetes/resources/delete',data)
