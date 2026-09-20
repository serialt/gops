import request from '@/utils/request'
export interface KubernetesCluster{id:number;name:string;alias:string;api_endpoint:string;version:string;status:number;node_count:number;pod_count:number;region:string;provider:string;description:string;created_at:string;updated_at:string}
export interface KubernetesClusterRequest{id?:number;name:string;alias:string;api_endpoint:string;kube_config:string;region:string;provider:string;description:string;status:number}
export const listKubernetesClusters=(keyword='')=>request.post<KubernetesCluster[]>('/kubernetes/clusters/list',{keyword});export const saveKubernetesCluster=(data:KubernetesClusterRequest)=>request.post<KubernetesCluster>('/kubernetes/clusters/save',data);export const deleteKubernetesCluster=(id:number)=>request.post<null>('/kubernetes/clusters/delete',{id})
export const testKubernetesCluster=(id:number)=>request.post<{status:string;version:string}>('/kubernetes/clusters/test',{id})
export const syncKubernetesClusters=()=>request.post<{total:number;success:number;failed:number;errors:string[]}>('/kubernetes/clusters/sync',{})
export const getKubernetesClusterConfig=(id:number)=>request.post<{kube_config:string}>('/kubernetes/clusters/config',{id})
export interface KubernetesNode{name:string;status:string;roles:string;age:string;version:string;internalIP:string;externalIP?:string;osImage:string;kernelVersion:string;containerRuntime:string;cpuCapacity:string;memoryCapacity:string;podCapacity:number;schedulable:boolean;taintCount:number;created_at:string};export interface KubernetesNamespace{name:string;status:string;age:string;created_at:string};export interface KubernetesWorkload{namespace:string;name:string;kind:string;ready:string;status:string;created_at:string}
export interface KubernetesPod{name:string;namespace:string;ready:string;status:string;restarts:number;age:string;ip:string;node:string}
export const listKubernetesNodes=(cluster_id:number,keyword='')=>request.post<KubernetesNode[]>('/kubernetes/nodes/list',{cluster_id,keyword});export const listKubernetesNamespaces=(cluster_id:number,keyword='')=>request.post<KubernetesNamespace[]>('/kubernetes/namespaces/list',{cluster_id,keyword});export const listKubernetesWorkloads=(cluster_id:number,keyword='',namespace='')=>request.post<KubernetesWorkload[]>('/kubernetes/workloads/list',{cluster_id,keyword,namespace})
export const listKubernetesPods=(cluster_id:number,keyword='',namespace='')=>request.post<KubernetesPod[]>('/kubernetes/pods/list',{cluster_id,keyword,namespace})
export interface KubernetesObject {name:string;namespace:string;kind:string;status:string;age:string;labels:Record<string,string>;created_at:string}
export interface KubernetesObjectRequest {cluster_id:number;resource:string;namespace?:string;name?:string;keyword?:string;yaml?:string}
export const listKubernetesObjects=(data:KubernetesObjectRequest)=>request.post<KubernetesObject[]>('/kubernetes/resources/list',data)
export const getKubernetesObject=(data:KubernetesObjectRequest)=>request.post<{yaml:string}>('/kubernetes/resources/get',data)
export const applyKubernetesObject=(data:KubernetesObjectRequest)=>request.post<KubernetesObject>('/kubernetes/resources/apply',data)
export const deleteKubernetesObject=(data:KubernetesObjectRequest)=>request.post<null>('/kubernetes/resources/delete',data)
