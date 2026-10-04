package biz

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func (b *KubernetesBiz) RoleSupport(req *api.KubernetesRoleDetailRequest) (*api.KubernetesRoleSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	role, err := client.RbacV1().Roles(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesRoleSupportResponse{
		Name: role.Name, Namespace: role.Namespace,
		CreatedAt: role.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(role.CreationTimestamp.Time),
		UID: string(role.UID), ResourceVersion: role.ResourceVersion, Labels: role.Labels, Annotations: role.Annotations,
		Rules: make([]*api.KubernetesRoleRuleResponse, 0, len(role.Rules)), Related: make([]*api.KubernetesRelatedResourceResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	if len(role.OwnerReferences) > 0 {
		owners := make([]string, 0, len(role.OwnerReferences))
		for _, owner := range role.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for _, rule := range role.Rules {
		result.Rules = append(result.Rules, &api.KubernetesRoleRuleResponse{
			APIGroups: append([]string(nil), rule.APIGroups...), Resources: append([]string(nil), rule.Resources...),
			ResourceNames: append([]string(nil), rule.ResourceNames...), Verbs: append([]string(nil), rule.Verbs...),
			NonResourceURLs: append([]string(nil), rule.NonResourceURLs...),
		})
	}
	var ancillary sync.WaitGroup
	ancillary.Add(2)
	go func() {
		defer ancillary.Done()
		listCtx, listCancel := context.WithTimeout(ctx, 4*time.Second)
		defer listCancel()
		bindings, listErr := client.RbacV1().RoleBindings(req.Namespace).List(listCtx, metav1.ListOptions{})
		if listErr != nil {
			return
		}
		for _, binding := range bindings.Items {
			if binding.RoleRef.Kind == "Role" && binding.RoleRef.Name == role.Name {
				result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "RoleBinding", Name: binding.Name, Status: "已绑定"})
			}
		}
	}()
	go func() {
		defer ancillary.Done()
		listCtx, listCancel := context.WithTimeout(ctx, 4*time.Second)
		defer listCancel()
		selector := fields.AndSelectors(fields.OneTermEqualSelector("involvedObject.kind", "Role"), fields.OneTermEqualSelector("involvedObject.name", role.Name)).String()
		events, listErr := client.CoreV1().Events(req.Namespace).List(listCtx, metav1.ListOptions{FieldSelector: selector})
		if listErr != nil {
			return
		}
		for _, event := range events.Items {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}()
	ancillary.Wait()

	var description strings.Builder
	fmt.Fprintf(&description, "Name:         %s\nNamespace:    %s\nLabels:       %s\nAnnotations:  %s\n", role.Name, role.Namespace, formatKubernetesMetadata(role.Labels), formatKubernetesMetadata(role.Annotations))
	for index, rule := range result.Rules {
		fmt.Fprintf(&description, "\nPolicyRule %d:\n  Resources: %s\n  Resource Names: %s\n  Verbs: %s\n  API Groups: %s\n  Non-Resource URLs: %s\n",
			index+1, strings.Join(rule.Resources, ","), strings.Join(rule.ResourceNames, ","), strings.Join(rule.Verbs, ","), strings.Join(rule.APIGroups, ","), strings.Join(rule.NonResourceURLs, ","))
	}
	if len(result.Events) == 0 {
		description.WriteString("\nEvents: <none>\n")
	} else {
		description.WriteString("\nEvents:\n  Type     Reason  Count  Source  Message\n")
		for _, event := range result.Events {
			fmt.Fprintf(&description, "  %-8s %-7s %-6d %-7s %s\n", event.Type, event.Reason, event.Count, event.Source, event.Message)
		}
	}
	result.Description = description.String()
	return result, nil
}
