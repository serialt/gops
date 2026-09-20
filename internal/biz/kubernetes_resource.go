package biz

import (
	"context"
	"fmt"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/utils"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (b *KubernetesBiz) client(clusterID int64) (*kubernetes.Clientset, error) {
	var c model.KubernetesCluster
	if err := b.db.First(&c, clusterID).Error; err != nil {
		return nil, err
	}
	if c.KubeConfig == "" {
		return nil, fmt.Errorf("cluster kubeconfig is empty")
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(c.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}
func (b *KubernetesBiz) Nodes(r *api.KubernetesResourceListRequest) ([]*api.KubernetesNodeResponse, error) {
	c, e := b.client(r.ClusterID)
	if e != nil {
		return nil, e
	}
	rows, e := c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	pods, e := c.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	podCounts := make(map[string]int, len(rows.Items))
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			podCounts[pod.Spec.NodeName]++
		}
	}
	o := make([]*api.KubernetesNodeResponse, 0, len(rows.Items))
	for _, n := range rows.Items {
		item := nodeResponse(n)
		item.PodCount = podCounts[n.Name]
		keyword := strings.ToLower(strings.TrimSpace(r.Keyword))
		if keyword != "" && !strings.Contains(strings.ToLower(item.Name), keyword) && !strings.Contains(strings.ToLower(item.InternalIP), keyword) && !strings.Contains(strings.ToLower(item.ExternalIP), keyword) {
			continue
		}
		if r.Status == "Ready" && item.Status != "就绪" {
			continue
		}
		if r.Status == "NotReady" && item.Status == "就绪" {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(r.Role))
		roles := strings.ToLower(item.Roles)
		if role == "worker" && roles != "" && !strings.Contains(roles, "worker") {
			continue
		}
		if role != "" && role != "worker" && !strings.Contains(roles, role) {
			continue
		}
		o = append(o, item)
	}
	return o, nil
}
func (b *KubernetesBiz) UpdateNode(r *api.KubernetesNodeUpdateRequest) (*api.KubernetesNodeResponse, error) {
	c, err := b.client(r.ClusterID)
	if err != nil {
		return nil, err
	}
	node, err := c.CoreV1().Nodes().Get(context.Background(), r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	node.Labels = r.Labels
	node.Spec.Unschedulable = !r.Schedulable
	if r.Taints != nil {
		node.Spec.Taints = make([]corev1.Taint, 0, len(*r.Taints))
		for _, taint := range *r.Taints {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{Key: taint.Key, Value: taint.Value, Effect: corev1.TaintEffect(taint.Effect)})
		}
	}
	updated, err := c.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return nodeResponse(*updated), nil
}
func nodeResponse(n corev1.Node) *api.KubernetesNodeResponse {
	status := "未知"
	for _, x := range n.Status.Conditions {
		if x.Type == corev1.NodeReady {
			if x.Status == corev1.ConditionTrue {
				status = "就绪"
			} else {
				status = "未就绪"
			}
		}
	}
	roles := make([]string, 0)
	for k := range n.Labels {
		const p = "node-role.kubernetes.io/"
		if strings.HasPrefix(k, p) {
			roles = append(roles, strings.TrimPrefix(k, p))
		}
	}
	ip, externalIP := "", ""
	for _, a := range n.Status.Addresses {
		if a.Type == corev1.NodeInternalIP {
			ip = a.Address
		} else if a.Type == corev1.NodeExternalIP {
			externalIP = a.Address
		}
	}
	taints := make([]api.TaintInfo, 0, len(n.Spec.Taints))
	for _, t := range n.Spec.Taints {
		taints = append(taints, api.TaintInfo{Key: t.Key, Value: t.Value, Effect: string(t.Effect)})
	}
	conditions := make([]api.NodeCondition, 0, len(n.Status.Conditions))
	for _, c := range n.Status.Conditions {
		conditions = append(conditions, api.NodeCondition{Type: string(c.Type), Status: string(c.Status), LastHeartbeatTime: c.LastHeartbeatTime.Format("2006-01-02 15:04:05"), LastTransitionTime: c.LastTransitionTime.Format("2006-01-02 15:04:05"), Reason: c.Reason, Message: c.Message})
	}
	return &api.KubernetesNodeResponse{Name: n.Name, Status: status, Roles: strings.Join(roles, ","), Age: utils.CalculateAge(n.CreationTimestamp.Time), Version: n.Status.NodeInfo.KubeletVersion, InternalIP: ip, ExternalIP: externalIP, OSImage: n.Status.NodeInfo.OSImage, KernelVersion: n.Status.NodeInfo.KernelVersion, ContainerRuntime: n.Status.NodeInfo.ContainerRuntimeVersion, Labels: n.Labels, Annotations: n.Annotations, CPUCapacity: n.Status.Capacity.Cpu().String(), MemoryCapacity: n.Status.Capacity.Memory().String(), PodCapacity: int(n.Status.Capacity.Pods().Value()), Schedulable: !n.Spec.Unschedulable, TaintCount: len(n.Spec.Taints), Taints: taints, PodCIDR: n.Spec.PodCIDR, ProviderID: n.Spec.ProviderID, Conditions: conditions, CreatedAt: n.CreationTimestamp.Format("2006-01-02 15:04:05")}
}

func (b *KubernetesBiz) Pods(r *api.KubernetesResourceListRequest) ([]*api.PodInfo, error) {
	c, err := b.client(r.ClusterID)
	if err != nil {
		return nil, err
	}
	ns := r.Namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	rows, err := c.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]*api.PodInfo, 0, len(rows.Items))
	for _, p := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(r.Keyword)) {
			continue
		}
		ready, restarts := 0, int32(0)
		containers := make([]api.ContainerInfo, 0, len(p.Spec.Containers))
		for _, x := range p.Status.ContainerStatuses {
			if x.Ready {
				ready++
			}
			restarts += x.RestartCount
		}
		for _, x := range p.Spec.Containers {
			containers = append(containers, api.ContainerInfo{Name: x.Name})
		}
		out = append(out, &api.PodInfo{Name: p.Name, Namespace: p.Namespace, Ready: fmt.Sprintf("%d/%d", ready, len(p.Spec.Containers)), Status: string(p.Status.Phase), Phase: string(p.Status.Phase), Restarts: restarts, Age: utils.CalculateAge(p.CreationTimestamp.Time), IP: p.Status.PodIP, Node: p.Spec.NodeName, Labels: p.Labels, Containers: containers})
	}
	return out, nil
}
func (b *KubernetesBiz) Namespaces(r *api.KubernetesResourceListRequest) ([]*api.KubernetesNamespaceResponse, error) {
	c, e := b.client(r.ClusterID)
	if e != nil {
		return nil, e
	}
	rows, e := c.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	o := make([]*api.KubernetesNamespaceResponse, 0, len(rows.Items))
	for _, x := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(x.Name), strings.ToLower(r.Keyword)) {
			continue
		}
		if r.Status != "" && !strings.EqualFold(string(x.Status.Phase), r.Status) {
			continue
		}
		o = append(o, &api.KubernetesNamespaceResponse{
			Name:      x.Name,
			Status:    string(x.Status.Phase),
			Age:       utils.CalculateAge(x.CreationTimestamp.Time),
			Labels:    x.Labels,
			CreatedAt: x.CreationTimestamp.Format("2006-01-02 15:04:05")})
	}
	return o, nil
}
func (b *KubernetesBiz) UpdateNamespace(r *api.KubernetesNamespaceUpdateRequest) (*api.KubernetesNamespaceResponse, error) {
	c, err := b.client(r.ClusterID)
	if err != nil {
		return nil, err
	}
	namespace, err := c.CoreV1().Namespaces().Get(context.Background(), r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	namespace.Labels = r.Labels
	updated, err := c.CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return &api.KubernetesNamespaceResponse{Name: updated.Name, Status: string(updated.Status.Phase), Age: utils.CalculateAge(updated.CreationTimestamp.Time), Labels: updated.Labels, CreatedAt: updated.CreationTimestamp.Format("2006-01-02 15:04:05")}, nil
}
func (b *KubernetesBiz) Workloads(r *api.KubernetesResourceListRequest) ([]*api.KubernetesWorkloadResponse, error) {
	c, e := b.client(r.ClusterID)
	if e != nil {
		return nil, e
	}
	ns := r.Namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	deployments, e := c.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	statefulSets, e := c.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	daemonSets, e := c.AppsV1().DaemonSets(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	jobs, e := c.BatchV1().Jobs(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	cronJobs, e := c.BatchV1().CronJobs(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	o := make([]*api.KubernetesWorkloadResponse, 0, len(deployments.Items)+len(statefulSets.Items)+len(daemonSets.Items)+len(jobs.Items)+len(cronJobs.Items))
	for _, x := range deployments.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(x.Name), strings.ToLower(r.Keyword)) {
			continue
		}
		o = append(o, workloadResponse(x))
	}
	appendWorkload := func(namespace, name, kind, ready, status string, created metav1.Time) {
		if r.Keyword == "" || strings.Contains(strings.ToLower(name), strings.ToLower(r.Keyword)) {
			o = append(o, &api.KubernetesWorkloadResponse{Namespace: namespace, Name: name, Kind: kind, Ready: ready, Status: status, CreatedAt: created.Format("2006-01-02 15:04:05")})
		}
	}
	for _, x := range statefulSets.Items {
		appendWorkload(x.Namespace, x.Name, "StatefulSet", fmt.Sprintf("%d/%d", x.Status.ReadyReplicas, x.Status.Replicas), "Running", x.CreationTimestamp)
	}
	for _, x := range daemonSets.Items {
		appendWorkload(x.Namespace, x.Name, "DaemonSet", fmt.Sprintf("%d/%d", x.Status.NumberReady, x.Status.DesiredNumberScheduled), "Running", x.CreationTimestamp)
	}
	for _, x := range jobs.Items {
		completions := int32(1)
		if x.Spec.Completions != nil {
			completions = *x.Spec.Completions
		}
		appendWorkload(x.Namespace, x.Name, "Job", fmt.Sprintf("%d/%d", x.Status.Succeeded, completions), jobStatus(x), x.CreationTimestamp)
	}
	for _, x := range cronJobs.Items {
		status := "Active"
		if x.Spec.Suspend != nil && *x.Spec.Suspend {
			status = "Suspended"
		}
		appendWorkload(x.Namespace, x.Name, "CronJob", fmt.Sprintf("%d active", len(x.Status.Active)), status, x.CreationTimestamp)
	}
	return o, nil
}
func jobStatus(x batchv1.Job) string {
	if x.Status.Failed > 0 {
		return "Failed"
	}
	if x.Status.Succeeded > 0 {
		return "Completed"
	}
	return "Running"
}

func (b *KubernetesBiz) Stats(r *api.KubernetesClusterStatsRequest) (*api.ClusterStats, error) {
	c, err := b.client(r.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	pods, err := c.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	deployments, _ := c.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	daemonSets, _ := c.AppsV1().DaemonSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	statefulSets, _ := c.AppsV1().StatefulSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	jobs, _ := c.BatchV1().Jobs(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	result := &api.ClusterStats{NodeCount: len(nodes.Items), PodCount: len(pods.Items), WorkloadCount: len(deployments.Items) + len(daemonSets.Items) + len(statefulSets.Items) + len(jobs.Items)}
	for _, n := range nodes.Items {
		result.CPUCapacity += float64(n.Status.Capacity.Cpu().MilliValue()) / 1000
		result.MemoryCapacity += float64(n.Status.Capacity.Memory().Value())
		result.CPUAllocatable += float64(n.Status.Allocatable.Cpu().MilliValue()) / 1000
		result.MemoryAllocatable += float64(n.Status.Allocatable.Memory().Value())
	}
	return result, nil
}

func (b *KubernetesBiz) Events(r *api.KubernetesResourceListRequest) ([]*api.EventInfo, error) {
	c, err := b.client(r.ClusterID)
	if err != nil {
		return nil, err
	}
	ns := r.Namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	rows, err := c.CoreV1().Events(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]*api.EventInfo, 0, len(rows.Items))
	for _, x := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(x.Message), strings.ToLower(r.Keyword)) {
			continue
		}
		out = append(out, &api.EventInfo{Type: x.Type, Reason: x.Reason, Message: x.Message, Source: x.Source.Component, Count: x.Count, FirstTimestamp: x.FirstTimestamp.Format("2006-01-02 15:04:05"), LastTimestamp: x.LastTimestamp.Format("2006-01-02 15:04:05"), InvolvedObject: api.InvolvedObjectInfo{Kind: x.InvolvedObject.Kind, Name: x.InvolvedObject.Name, Namespace: x.InvolvedObject.Namespace}})
	}
	return out, nil
}
func workloadResponse(x appsv1.Deployment) *api.KubernetesWorkloadResponse {
	status := "未知"
	if len(x.Status.Conditions) > 0 {
		status = string(x.Status.Conditions[len(x.Status.Conditions)-1].Type)
	}
	return &api.KubernetesWorkloadResponse{Namespace: x.Namespace, Name: x.Name, Kind: "Deployment", Ready: fmt.Sprintf("%d/%d", x.Status.ReadyReplicas, x.Status.Replicas), Status: status, CreatedAt: x.CreationTimestamp.Format("2006-01-02 15:04:05")}
}
