package biz

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) NamespaceSupport(req *api.KubernetesNamespaceDetailRequest) (*api.KubernetesNamespaceSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	namespace, err := client.CoreV1().Namespaces().Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesNamespaceSupportResponse{
		Name: namespace.Name, Status: string(namespace.Status.Phase),
		CreatedAt: namespace.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Age:       utils.CalculateAge(namespace.CreationTimestamp.Time), ResourceVersion: namespace.ResourceVersion,
		UID: string(namespace.UID), Labels: namespace.Labels, Annotations: namespace.Annotations,
		Resources: make([]*api.KubernetesNamespaceResourceResponse, 0),
		Pods:      make([]*api.KubernetesObjectResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	dynamic, err := b.dynamicClient(req.ClusterID)
	if err != nil {
		return nil, err
	}
	for resourceName, spec := range kubernetesResourceCatalog {
		if !spec.namespaced || resourceName == "pods" {
			continue
		}
		rows, listErr := dynamic.Resource(spec.gvr).Namespace(namespace.Name).List(ctx, metav1.ListOptions{})
		if listErr != nil {
			return nil, fmt.Errorf("list %s in namespace %s: %w", resourceName, namespace.Name, listErr)
		}
		for _, item := range rows.Items {
			result.Resources = append(result.Resources, &api.KubernetesNamespaceResourceResponse{
				Kind: spec.kind, Name: item.GetName(), Status: objectStatus(item, spec.kind),
				Age: utils.CalculateAge(item.GetCreationTimestamp().Time),
			})
		}
	}
	pods, err := client.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		result.Pods = append(result.Pods, batchPodResponse(pod))
		result.Resources = append(result.Resources, &api.KubernetesNamespaceResourceResponse{
			Kind: "Pod", Name: pod.Name, Status: string(pod.Status.Phase), Age: utils.CalculateAge(pod.CreationTimestamp.Time),
		})
	}
	events, err := client.CoreV1().Events(namespace.Name).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		result.Events = append(result.Events, batchEventResponse(event))
	}
	sort.Slice(result.Resources, func(i, j int) bool {
		if result.Resources[i].Kind == result.Resources[j].Kind {
			return result.Resources[i].Name < result.Resources[j].Name
		}
		return result.Resources[i].Kind < result.Resources[j].Kind
	})
	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nStatus:                %s\nCreated:               %s\nResource Version:      %s\nUID:                   %s\nResources:             %d\nPods:                  %d\nEvents:                %d\n", result.Name, result.Status, result.CreatedAt, result.ResourceVersion, result.UID, len(result.Resources), len(result.Pods), len(result.Events))
	result.Description = description.String()
	return result, nil
}
