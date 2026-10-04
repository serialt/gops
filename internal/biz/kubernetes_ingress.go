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

func (b *KubernetesBiz) IngressSupport(req *api.KubernetesIngressDetailRequest) (*api.KubernetesIngressSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ingress, err := client.NetworkingV1().Ingresses(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	className := ""
	if ingress.Spec.IngressClassName != nil {
		className = *ingress.Spec.IngressClassName
	}
	result := &api.KubernetesIngressSupportResponse{
		Name: ingress.Name, Namespace: ingress.Namespace, IngressClass: className,
		Hosts: make([]string, 0), Addresses: make([]string, 0),
		Rules:    make([]*api.KubernetesIngressRuleResponse, 0, len(ingress.Spec.Rules)),
		TLS:      make([]*api.KubernetesIngressTLSResponse, 0, len(ingress.Spec.TLS)),
		Backends: make([]*api.KubernetesIngressServiceResponse, 0),
		Events:   make([]*api.EventInfo, 0), Labels: ingress.Labels, Annotations: ingress.Annotations,
		UID: string(ingress.UID), ResourceVersion: ingress.ResourceVersion,
		CreatedAt: ingress.CreationTimestamp.Format("2006-01-02 15:04:05"),
		Age:       utils.CalculateAge(ingress.CreationTimestamp.Time),
	}
	if len(ingress.OwnerReferences) > 0 {
		owners := make([]string, 0, len(ingress.OwnerReferences))
		for _, owner := range ingress.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for _, address := range ingress.Status.LoadBalancer.Ingress {
		value := address.IP
		if value == "" {
			value = address.Hostname
		}
		if value != "" && !containsString(result.Addresses, value) {
			result.Addresses = append(result.Addresses, value)
		}
	}

	serviceNames := make(map[string]struct{})
	addBackend := func(backend *networkingv1.IngressBackend) *api.KubernetesIngressBackendResponse {
		if backend == nil || backend.Service == nil {
			return nil
		}
		port := ""
		if backend.Service.Port.Name != "" {
			port = backend.Service.Port.Name
		} else if backend.Service.Port.Number != 0 {
			port = fmt.Sprint(backend.Service.Port.Number)
		}
		serviceNames[backend.Service.Name] = struct{}{}
		return &api.KubernetesIngressBackendResponse{ServiceName: backend.Service.Name, ServicePort: port}
	}
	result.DefaultBackend = addBackend(ingress.Spec.DefaultBackend)
	for _, rule := range ingress.Spec.Rules {
		row := &api.KubernetesIngressRuleResponse{Host: rule.Host, Paths: make([]*api.KubernetesIngressPathResponse, 0)}
		if rule.Host != "" && !containsString(result.Hosts, rule.Host) {
			result.Hosts = append(result.Hosts, rule.Host)
		}
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				backend := addBackend(&path.Backend)
				entry := &api.KubernetesIngressPathResponse{Path: path.Path}
				if path.PathType != nil {
					entry.PathType = string(*path.PathType)
				}
				if backend != nil {
					entry.ServiceName = backend.ServiceName
					entry.ServicePort = backend.ServicePort
				}
				row.Paths = append(row.Paths, entry)
			}
		}
		result.Rules = append(result.Rules, row)
	}
	for _, tls := range ingress.Spec.TLS {
		result.TLS = append(result.TLS, &api.KubernetesIngressTLSResponse{Hosts: append([]string(nil), tls.Hosts...), SecretName: tls.SecretName})
		for _, host := range tls.Hosts {
			if host != "" && !containsString(result.Hosts, host) {
				result.Hosts = append(result.Hosts, host)
			}
		}
	}
	for serviceName := range serviceNames {
		service, getErr := client.CoreV1().Services(req.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				result.Backends = append(result.Backends, &api.KubernetesIngressServiceResponse{Name: serviceName, Type: "Service 不存在"})
				continue
			}
			return nil, getErr
		}
		ports := make([]string, 0, len(service.Spec.Ports))
		for _, port := range service.Spec.Ports {
			protocol := string(port.Protocol)
			if protocol == "" {
				protocol = string(corev1.ProtocolTCP)
			}
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, protocol))
		}
		result.Backends = append(result.Backends, &api.KubernetesIngressServiceResponse{Name: service.Name, Type: string(service.Spec.Type), ClusterIP: service.Spec.ClusterIP, Ports: strings.Join(ports, ", ")})
	}
	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		if strings.EqualFold(event.InvolvedObject.Kind, "Ingress") && event.InvolvedObject.Name == ingress.Name {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}

	var description strings.Builder
	fmt.Fprintf(&description, "Name:                  %s\nNamespace:             %s\nIngress Class:         %s\n", ingress.Name, ingress.Namespace, className)
	fmt.Fprintf(&description, "Hosts:                 %s\nAddresses:             %s\nRules:                 %d\nTLS Configurations:    %d\n", strings.Join(result.Hosts, ", "), strings.Join(result.Addresses, ", "), len(result.Rules), len(result.TLS))
	for _, rule := range result.Rules {
		for _, path := range rule.Paths {
			fmt.Fprintf(&description, "Rule:                  %s%s → %s:%s\n", rule.Host, path.Path, path.ServiceName, path.ServicePort)
		}
	}
	result.Description = description.String()
	return result, nil
}
