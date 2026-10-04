package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) NodeSupport(req *api.KubernetesNodeDetailRequest) (*api.KubernetesNodeSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + node.Name})
	if err != nil {
		return nil, err
	}
	events, err := client.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	response := nodeResponse(*node)
	response.PodCount = len(pods.Items)
	result := &api.KubernetesNodeSupportResponse{
		Node: response, Pods: make([]*api.KubernetesObjectResponse, 0, len(pods.Items)),
		Events: make([]*api.EventInfo, 0),
	}
	podNames := make(map[string]struct{}, len(pods.Items))
	for _, pod := range pods.Items {
		result.Pods = append(result.Pods, batchPodResponse(pod))
		podNames[pod.Namespace+"/"+pod.Name] = struct{}{}
	}
	for _, event := range events.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		related := kind == "node" && event.InvolvedObject.Name == node.Name
		if kind == "pod" {
			_, related = podNames[event.InvolvedObject.Namespace+"/"+event.InvolvedObject.Name]
		}
		if related {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}
	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nStatus:                %s\nRoles:                 %s\nInternal IP:           %s\nExternal IP:           %s\nKubelet Version:       %s\nOS Image:              %s\nKernel Version:        %s\nContainer Runtime:     %s\nPod CIDR:              %s\nProvider ID:           %s\nPod Capacity:          %d/%d\nSchedulable:           %t\n", response.Name, response.Status, defaultDescriptionValue(response.Roles, "<none>"), response.InternalIP, defaultDescriptionValue(response.ExternalIP, "<none>"), response.Version, response.OSImage, response.KernelVersion, response.ContainerRuntime, defaultDescriptionValue(response.PodCIDR, "<none>"), defaultDescriptionValue(response.ProviderID, "<none>"), response.PodCount, response.PodCapacity, response.Schedulable)
	result.Description = description.String()
	return result, nil
}
