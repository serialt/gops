package biz

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) SecretSupport(req *api.KubernetesSecretDetailRequest) (*api.KubernetesSecretSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	secret, err := client.CoreV1().Secrets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesSecretSupportResponse{
		Name: secret.Name, Namespace: secret.Namespace,
		CreatedAt: secret.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(secret.CreationTimestamp.Time),
		Type: string(secret.Type), UID: string(secret.UID), ResourceVersion: secret.ResourceVersion,
		Labels: secret.Labels, Annotations: secret.Annotations,
		Data: make(map[string]string, len(secret.Data)), Related: make([]*api.KubernetesRelatedResourceResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	if result.Type == "" {
		result.Type = "Opaque"
	}
	if len(secret.OwnerReferences) > 0 {
		owners := make([]string, 0, len(secret.OwnerReferences))
		for _, owner := range secret.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for key, value := range secret.Data {
		result.Data[key] = base64.StdEncoding.EncodeToString(value)
		result.Size += int64(len(value))
	}
	pods, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	related := make(map[string]*api.KubernetesRelatedResourceResponse)
	for _, pod := range pods.Items {
		if !podReferencesSecret(pod, secret.Name) {
			continue
		}
		foundOwner := false
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {
				replicaSet, getErr := client.AppsV1().ReplicaSets(req.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
				if getErr == nil {
					for _, replicaSetOwner := range replicaSet.OwnerReferences {
						if replicaSetOwner.Kind != "Deployment" {
							continue
						}
						key := "Deployment/" + replicaSetOwner.Name
						related[key] = &api.KubernetesRelatedResourceResponse{Kind: "Deployment", Name: replicaSetOwner.Name, Status: "引用中"}
						foundOwner = true
					}
				}
				if foundOwner {
					continue
				}
			}
			key := owner.Kind + "/" + owner.Name
			related[key] = &api.KubernetesRelatedResourceResponse{Kind: owner.Kind, Name: owner.Name, Status: "引用中"}
			foundOwner = true
		}
		if !foundOwner {
			key := "Pod/" + pod.Name
			related[key] = &api.KubernetesRelatedResourceResponse{Kind: "Pod", Name: pod.Name, Status: string(pod.Status.Phase)}
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
		if !strings.EqualFold(event.InvolvedObject.Kind, "Secret") || event.InvolvedObject.Name != secret.Name {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}

	var description strings.Builder
	fmt.Fprintf(&description, "Name:         %s\nNamespace:    %s\nLabels:       %s\nAnnotations:  %s\nType:         %s\n", secret.Name, secret.Namespace, formatKubernetesMetadata(secret.Labels), formatKubernetesMetadata(secret.Annotations), result.Type)
	if len(secret.Data) == 0 {
		description.WriteString("Data: <none>\n")
	} else {
		description.WriteString("Data\n====\n")
		keys := make([]string, 0, len(secret.Data))
		for key := range secret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&description, "%s: %d bytes\n", key, len(secret.Data[key]))
		}
	}
	fmt.Fprintf(&description, "Total size: %d bytes\n", result.Size)
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

func podReferencesSecret(pod corev1.Pod, name string) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil && volume.Secret.SecretName == name {
			return true
		}
	}
	for _, pullSecret := range pod.Spec.ImagePullSecrets {
		if pullSecret.Name == name {
			return true
		}
	}
	for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		if containerReferencesSecret(container, name) {
			return true
		}
	}
	return false
}

func containerReferencesSecret(container corev1.Container, name string) bool {
	for _, source := range container.EnvFrom {
		if source.SecretRef != nil && source.SecretRef.Name == name {
			return true
		}
	}
	for _, env := range container.Env {
		if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == name {
			return true
		}
	}
	return false
}
