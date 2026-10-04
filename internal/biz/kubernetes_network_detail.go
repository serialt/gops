package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) NetworkPolicySupport(req *api.KubernetesNetworkDetailRequest) (*api.KubernetesNetworkPolicySupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	policy, err := client.NetworkingV1().NetworkPolicies(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(&policy.Spec.PodSelector)
	if err != nil {
		return nil, err
	}
	pods, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesNetworkPolicySupportResponse{
		Name: policy.Name, Namespace: policy.Namespace, PodSelector: policy.Spec.PodSelector.MatchLabels,
		PodSelectorText: selector.String(),
		PolicyTypes:     make([]string, 0, len(policy.Spec.PolicyTypes)),
		Ingress:         make([]*api.KubernetesNetworkPolicyRuleResponse, 0, len(policy.Spec.Ingress)),
		Egress:          make([]*api.KubernetesNetworkPolicyRuleResponse, 0, len(policy.Spec.Egress)),
		Pods:            make([]*api.KubernetesObjectResponse, 0, len(pods.Items)), Events: make([]*api.EventInfo, 0),
		Labels: policy.Labels, Annotations: policy.Annotations, UID: string(policy.UID),
		CreatedAt: policy.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(policy.CreationTimestamp.Time),
	}
	for _, policyType := range policy.Spec.PolicyTypes {
		result.PolicyTypes = append(result.PolicyTypes, string(policyType))
	}
	for _, rule := range policy.Spec.Ingress {
		result.Ingress = append(result.Ingress, networkPolicyRuleResponse(rule.From, rule.Ports))
	}
	for _, rule := range policy.Spec.Egress {
		result.Egress = append(result.Egress, networkPolicyRuleResponse(rule.To, rule.Ports))
	}
	selectedPodNames := make(map[string]struct{}, len(pods.Items))
	for _, pod := range pods.Items {
		result.Pods = append(result.Pods, batchPodResponse(pod))
		selectedPodNames[pod.Name] = struct{}{}
	}
	for _, event := range events.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "networkpolicy" || event.InvolvedObject.Name != policy.Name) && (kind != "pod" || !containsPod(selectedPodNames, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}
	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nNamespace:             %s\nPod Selector:          %s\nPolicy Types:          %s\nSelected Pods:         %d\nIngress Rules:         %d\nEgress Rules:          %d\n", policy.Name, policy.Namespace, selector.String(), strings.Join(result.PolicyTypes, ", "), len(result.Pods), len(result.Ingress), len(result.Egress))
	result.Description = description.String()
	return result, nil
}

func networkPolicyRuleResponse(peers []networkingv1.NetworkPolicyPeer, ports []networkingv1.NetworkPolicyPort) *api.KubernetesNetworkPolicyRuleResponse {
	result := &api.KubernetesNetworkPolicyRuleResponse{
		Peers: make([]*api.KubernetesNetworkPolicyPeerResponse, 0, len(peers)),
		Ports: make([]*api.KubernetesNetworkPolicyPortResponse, 0, len(ports)),
	}
	for _, peer := range peers {
		entry := &api.KubernetesNetworkPolicyPeerResponse{}
		if peer.PodSelector != nil {
			entry.PodSelector = peer.PodSelector.MatchLabels
			if selector, err := metav1.LabelSelectorAsSelector(peer.PodSelector); err == nil {
				entry.PodSelectorText = selector.String()
			}
		}
		if peer.NamespaceSelector != nil {
			entry.NamespaceSelector = peer.NamespaceSelector.MatchLabels
			if selector, err := metav1.LabelSelectorAsSelector(peer.NamespaceSelector); err == nil {
				entry.NamespaceSelectorText = selector.String()
			}
		}
		if peer.IPBlock != nil {
			entry.IPBlockCIDR = peer.IPBlock.CIDR
			entry.IPBlockExcept = append([]string(nil), peer.IPBlock.Except...)
		}
		result.Peers = append(result.Peers, entry)
	}
	for _, port := range ports {
		protocol := string(corev1.ProtocolTCP)
		if port.Protocol != nil {
			protocol = string(*port.Protocol)
		}
		portValue := "全部"
		if port.Port != nil {
			portValue = port.Port.String()
		}
		result.Ports = append(result.Ports, &api.KubernetesNetworkPolicyPortResponse{Protocol: protocol, Port: portValue, EndPort: port.EndPort})
	}
	return result
}

func (b *KubernetesBiz) EndpointsSupport(req *api.KubernetesNetworkDetailRequest) (*api.KubernetesEndpointsSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	endpoints, err := client.CoreV1().Endpoints(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesEndpointsSupportResponse{
		Name: endpoints.Name, Namespace: endpoints.Namespace,
		Addresses: make([]*api.KubernetesEndpointAddressResponse, 0),
		NotReady:  make([]*api.KubernetesEndpointAddressResponse, 0),
		Ports:     make([]*api.KubernetesServicePortResponse, 0),
		Pods:      make([]*api.KubernetesObjectResponse, 0), Events: make([]*api.EventInfo, 0),
		Labels: endpoints.Labels, Annotations: endpoints.Annotations, UID: string(endpoints.UID),
		CreatedAt: endpoints.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(endpoints.CreationTimestamp.Time),
	}
	podNames := make(map[string]struct{})
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			result.Addresses = append(result.Addresses, endpointAddressResponse(address))
			if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
				podNames[address.TargetRef.Name] = struct{}{}
			}
		}
		for _, address := range subset.NotReadyAddresses {
			result.NotReady = append(result.NotReady, endpointAddressResponse(address))
			if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
				podNames[address.TargetRef.Name] = struct{}{}
			}
		}
		for _, port := range subset.Ports {
			protocol := string(port.Protocol)
			if protocol == "" {
				protocol = string(corev1.ProtocolTCP)
			}
			result.Ports = append(result.Ports, &api.KubernetesServicePortResponse{Name: port.Name, Protocol: protocol, Port: port.Port, TargetPort: port.Name})
		}
	}
	for podName := range podNames {
		pod, getErr := client.CoreV1().Pods(req.Namespace).Get(ctx, podName, metav1.GetOptions{})
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				continue
			}
			return nil, getErr
		}
		result.Pods = append(result.Pods, batchPodResponse(*pod))
	}
	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "endpoints" || event.InvolvedObject.Name != endpoints.Name) && (kind != "pod" || !containsPod(podNames, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}
	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nNamespace:             %s\nReady Addresses:       %d\nNot Ready Addresses:   %d\nPorts:                 %d\n", endpoints.Name, endpoints.Namespace, len(result.Addresses), len(result.NotReady), len(result.Ports))
	result.Description = description.String()
	return result, nil
}

func endpointAddressResponse(address corev1.EndpointAddress) *api.KubernetesEndpointAddressResponse {
	result := &api.KubernetesEndpointAddressResponse{IP: address.IP, Hostname: address.Hostname}
	if address.NodeName != nil {
		result.NodeName = *address.NodeName
	}
	if address.TargetRef != nil {
		result.TargetKind = address.TargetRef.Kind
		result.TargetName = address.TargetRef.Name
	}
	return result
}
