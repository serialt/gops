package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (b *KubernetesBiz) ServiceSupport(req *api.KubernetesServiceDetailRequest) (*api.KubernetesServiceSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	service, err := client.CoreV1().Services(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesServiceSupportResponse{
		Name: service.Name, Namespace: service.Namespace, Type: string(service.Spec.Type),
		ClusterIP: service.Spec.ClusterIP, ExternalName: service.Spec.ExternalName, ExternalIPs: append([]string(nil), service.Spec.ExternalIPs...),
		Selector: service.Spec.Selector, Labels: service.Labels, Annotations: service.Annotations,
		CreatedAt: service.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Age:       utils.CalculateAge(service.CreationTimestamp.Time), UID: string(service.UID),
		ResourceVersion: service.ResourceVersion,
		Ports:           make([]*api.KubernetesServicePortResponse, 0, len(service.Spec.Ports)),
		Pods:            make([]*api.KubernetesObjectResponse, 0),
		Endpoints:       make([]*api.KubernetesServiceEndpointResponse, 0, 1),
		EndpointSlices:  make([]*api.KubernetesServiceEndpointResponse, 0),
		Events:          make([]*api.EventInfo, 0),
	}
	if len(service.OwnerReferences) > 0 {
		owners := make([]string, 0, len(service.OwnerReferences))
		for _, owner := range service.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for _, port := range service.Spec.Ports {
		protocol := string(port.Protocol)
		if protocol == "" {
			protocol = string(corev1.ProtocolTCP)
		}
		targetPort := port.TargetPort.String()
		if targetPort == "" || targetPort == "0" {
			targetPort = fmt.Sprint(port.Port)
		}
		result.Ports = append(result.Ports, &api.KubernetesServicePortResponse{
			Name: port.Name, Protocol: protocol, Port: port.Port, TargetPort: targetPort, NodePort: port.NodePort,
		})
	}
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			address := ingress.IP
			if address == "" {
				address = ingress.Hostname
			}
			if address != "" && !containsString(result.ExternalIPs, address) {
				result.ExternalIPs = append(result.ExternalIPs, address)
			}
		}
	}

	ownedPodNames := make(map[string]struct{})
	if len(service.Spec.Selector) > 0 {
		selector := labels.Set(service.Spec.Selector).AsSelector().String()
		pods, listErr := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if listErr != nil {
			return nil, listErr
		}
		for _, pod := range pods.Items {
			result.Pods = append(result.Pods, batchPodResponse(pod))
			ownedPodNames[pod.Name] = struct{}{}
		}
	}

	endpoint, endpointErr := client.CoreV1().Endpoints(req.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
	if endpointErr != nil && !apierrors.IsNotFound(endpointErr) {
		return nil, endpointErr
	}
	if endpointErr == nil {
		row := &api.KubernetesServiceEndpointResponse{
			Name: endpoint.Name, CreatedAt: endpoint.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Age: utils.CalculateAge(endpoint.CreationTimestamp.Time), Addresses: make([]string, 0),
			NotReady: make([]string, 0), Ports: make([]string, 0),
		}
		for _, subset := range endpoint.Subsets {
			for _, address := range subset.Addresses {
				row.Addresses = append(row.Addresses, address.IP)
			}
			for _, address := range subset.NotReadyAddresses {
				row.NotReady = append(row.NotReady, address.IP)
			}
			for _, port := range subset.Ports {
				row.Ports = append(row.Ports, fmt.Sprintf("%s/%d", port.Protocol, port.Port))
			}
		}
		row.EndpointCount = len(row.Addresses)
		result.Endpoints = append(result.Endpoints, row)
	}

	slices, err := client.DiscoveryV1().EndpointSlices(req.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{discoveryv1.LabelServiceName: service.Name}.AsSelector().String(),
	})
	if err != nil {
		return nil, err
	}
	for _, slice := range slices.Items {
		row := &api.KubernetesServiceEndpointResponse{
			Name: slice.Name, CreatedAt: slice.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Age: utils.CalculateAge(slice.CreationTimestamp.Time), AddressType: string(slice.AddressType),
			Addresses: make([]string, 0), Ports: make([]string, 0),
		}
		for _, endpoint := range slice.Endpoints {
			row.EndpointCount++
			row.Addresses = append(row.Addresses, endpoint.Addresses...)
		}
		for _, port := range slice.Ports {
			if port.Port == nil {
				continue
			}
			protocol := string(corev1.ProtocolTCP)
			if port.Protocol != nil {
				protocol = string(*port.Protocol)
			}
			row.Ports = append(row.Ports, fmt.Sprintf("%s/%d", protocol, *port.Port))
		}
		result.EndpointSlices = append(result.EndpointSlices, row)
	}

	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "service" || event.InvolvedObject.Name != service.Name) && (kind != "pod" || !containsPod(ownedPodNames, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}

	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nNamespace:             %s\nType:                  %s\nClusterIP:             %s\n", service.Name, service.Namespace, service.Spec.Type, service.Spec.ClusterIP)
	fmt.Fprintf(&description, "Selector:              %s\nEndpoints:             %d\nEndpointSlices:        %d\n", labels.Set(service.Spec.Selector).AsSelector().String(), len(result.Endpoints), len(result.EndpointSlices))
	for _, port := range result.Ports {
		fmt.Fprintf(&description, "Port:                  %s %d/%s → %s\n", port.Name, port.Port, port.Protocol, port.TargetPort)
	}
	result.Description = description.String()
	return result, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
