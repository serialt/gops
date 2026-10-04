package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) ClusterRoleBindingSupport(req *api.KubernetesClusterRoleBindingDetailRequest) (*api.KubernetesClusterRoleBindingSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	binding, err := client.RbacV1().ClusterRoleBindings().Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesClusterRoleBindingSupportResponse{
		Name:      binding.Name,
		CreatedAt: binding.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(binding.CreationTimestamp.Time),
		UID: string(binding.UID), ResourceVersion: binding.ResourceVersion, Labels: binding.Labels, Annotations: binding.Annotations,
		RoleRef:  &api.KubernetesRoleBindingRoleRefResponse{APIGroup: binding.RoleRef.APIGroup, Kind: binding.RoleRef.Kind, Name: binding.RoleRef.Name},
		Subjects: make([]*api.KubernetesRoleBindingSubjectResponse, 0, len(binding.Subjects)), Events: make([]*api.EventInfo, 0),
	}
	if len(binding.OwnerReferences) > 0 {
		owners := make([]string, 0, len(binding.OwnerReferences))
		for _, owner := range binding.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for _, subject := range binding.Subjects {
		result.Subjects = append(result.Subjects, &api.KubernetesRoleBindingSubjectResponse{Kind: subject.Kind, Name: subject.Name, Namespace: subject.Namespace})
	}
	events, err := client.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		if !strings.EqualFold(event.InvolvedObject.Kind, "ClusterRoleBinding") || event.InvolvedObject.Name != binding.Name {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}

	var description strings.Builder
	fmt.Fprintf(&description, "Name:         %s\nLabels:       %s\nAnnotations:  %s\nRole:         %s/%s\nAPI Group:    %s\n",
		binding.Name, formatKubernetesMetadata(binding.Labels), formatKubernetesMetadata(binding.Annotations), binding.RoleRef.Kind, binding.RoleRef.Name, binding.RoleRef.APIGroup)
	description.WriteString("Subjects:\n")
	if len(result.Subjects) == 0 {
		description.WriteString("  <none>\n")
	} else {
		for _, subject := range result.Subjects {
			if subject.Namespace != "" {
				fmt.Fprintf(&description, "  %s: %s/%s\n", subject.Kind, subject.Namespace, subject.Name)
			} else {
				fmt.Fprintf(&description, "  %s: %s\n", subject.Kind, subject.Name)
			}
		}
	}
	if len(result.Events) == 0 {
		description.WriteString("Events: <none>\n")
	} else {
		description.WriteString("Events:\n  Type     Reason  Count  Source  Message\n")
		for _, event := range result.Events {
			fmt.Fprintf(&description, "  %-8s %-7s %-6d %-7s %s\n", event.Type, event.Reason, event.Count, event.Source, event.Message)
		}
	}
	result.Description = description.String()
	return result, nil
}
