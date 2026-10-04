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

func (b *KubernetesBiz) ServiceAccountSupport(req *api.KubernetesServiceAccountDetailRequest) (*api.KubernetesServiceAccountSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	account, err := client.CoreV1().ServiceAccounts(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesServiceAccountSupportResponse{
		Name: account.Name, Namespace: account.Namespace,
		CreatedAt: account.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(account.CreationTimestamp.Time),
		UID: string(account.UID), ResourceVersion: account.ResourceVersion, Labels: account.Labels, Annotations: account.Annotations,
		Secrets: make([]string, 0, len(account.Secrets)), ImagePullSecrets: make([]string, 0, len(account.ImagePullSecrets)),
		Related: make([]*api.KubernetesRelatedResourceResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	if account.AutomountServiceAccountToken == nil {
		result.AutomountToken = "默认"
	} else if *account.AutomountServiceAccountToken {
		result.AutomountToken = "启用"
	} else {
		result.AutomountToken = "禁用"
	}
	if len(account.OwnerReferences) > 0 {
		owners := make([]string, 0, len(account.OwnerReferences))
		for _, owner := range account.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for _, secret := range account.Secrets {
		result.Secrets = append(result.Secrets, secret.Name)
	}
	for _, secret := range account.ImagePullSecrets {
		result.ImagePullSecrets = append(result.ImagePullSecrets, secret.Name)
	}

	related := make(map[string]*api.KubernetesRelatedResourceResponse)
	pods, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		serviceAccountName := pod.Spec.ServiceAccountName
		if serviceAccountName == "" {
			serviceAccountName = "default"
		}
		if serviceAccountName == account.Name {
			related["Pod/"+pod.Name] = &api.KubernetesRelatedResourceResponse{Kind: "Pod", Name: pod.Name, Status: string(pod.Status.Phase)}
		}
	}
	roleBindings, err := client.RbacV1().RoleBindings(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, binding := range roleBindings.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == "ServiceAccount" && subject.Name == account.Name && (subject.Namespace == "" || subject.Namespace == req.Namespace) {
				related["RoleBinding/"+binding.Name] = &api.KubernetesRelatedResourceResponse{Kind: "RoleBinding", Name: binding.Name, Status: "已绑定"}
				break
			}
		}
	}
	clusterRoleBindings, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, binding := range clusterRoleBindings.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == "ServiceAccount" && subject.Name == account.Name && subject.Namespace == req.Namespace {
				related["ClusterRoleBinding/"+binding.Name] = &api.KubernetesRelatedResourceResponse{Kind: "ClusterRoleBinding", Name: binding.Name, Status: "已绑定"}
				break
			}
		}
	}
	for _, item := range related {
		result.Related = append(result.Related, item)
	}
	sort.Slice(result.Related, func(i, j int) bool {
		if result.Related[i].Kind == result.Related[j].Kind {
			return result.Related[i].Name < result.Related[j].Name
		}
		return result.Related[i].Kind < result.Related[j].Kind
	})

	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		if !strings.EqualFold(event.InvolvedObject.Kind, "ServiceAccount") || event.InvolvedObject.Name != account.Name {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}

	var description strings.Builder
	fmt.Fprintf(&description, "Name:                      %s\nNamespace:                 %s\nLabels:                    %s\nAnnotations:               %s\nAutomount Service Account Token: %s\n", account.Name, account.Namespace, formatKubernetesMetadata(account.Labels), formatKubernetesMetadata(account.Annotations), result.AutomountToken)
	if len(result.Secrets) == 0 {
		description.WriteString("Mountable secrets:         <none>\n")
	} else {
		fmt.Fprintf(&description, "Mountable secrets:         %s\n", strings.Join(result.Secrets, ", "))
	}
	if len(result.ImagePullSecrets) == 0 {
		description.WriteString("Image pull secrets:        <none>\n")
	} else {
		fmt.Fprintf(&description, "Image pull secrets:        %s\n", strings.Join(result.ImagePullSecrets, ", "))
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
