package biz

import (
	"context"
	"fmt"
	"gops/internal/data/api"
	"gops/internal/data/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
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
	o := make([]*api.KubernetesNodeResponse, 0, len(rows.Items))
	for _, n := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(n.Name), strings.ToLower(r.Keyword)) {
			continue
		}
		o = append(o, nodeResponse(n))
	}
	return o, nil
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
	ip := ""
	for _, a := range n.Status.Addresses {
		if a.Type == corev1.NodeInternalIP {
			ip = a.Address
			break
		}
	}
	return &api.KubernetesNodeResponse{Name: n.Name, Status: status, Roles: strings.Join(roles, ","), Version: n.Status.NodeInfo.KubeletVersion, InternalIP: ip, CreatedAt: n.CreationTimestamp.Format("2006-01-02 15:04:05")}
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
		o = append(o, &api.KubernetesNamespaceResponse{Name: x.Name, Status: string(x.Status.Phase), CreatedAt: x.CreationTimestamp.Format("2006-01-02 15:04:05")})
	}
	return o, nil
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
	rows, e := c.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	o := make([]*api.KubernetesWorkloadResponse, 0, len(rows.Items))
	for _, x := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(x.Name), strings.ToLower(r.Keyword)) {
			continue
		}
		o = append(o, workloadResponse(x))
	}
	return o, nil
}
func workloadResponse(x appsv1.Deployment) *api.KubernetesWorkloadResponse {
	status := "未知"
	if len(x.Status.Conditions) > 0 {
		status = string(x.Status.Conditions[len(x.Status.Conditions)-1].Type)
	}
	return &api.KubernetesWorkloadResponse{Namespace: x.Namespace, Name: x.Name, Kind: "Deployment", Ready: fmt.Sprintf("%d/%d", x.Status.ReadyReplicas, x.Status.Replicas), Status: status, CreatedAt: x.CreationTimestamp.Format("2006-01-02 15:04:05")}
}
