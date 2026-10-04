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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *KubernetesBiz) ConfigMapSupport(req *api.KubernetesConfigMapDetailRequest) (*api.KubernetesConfigMapSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	configMap, err := client.CoreV1().ConfigMaps(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesConfigMapSupportResponse{
		Name: configMap.Name, Namespace: configMap.Namespace,
		CreatedAt: configMap.CreationTimestamp.Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(configMap.CreationTimestamp.Time),
		UID: string(configMap.UID), ResourceVersion: configMap.ResourceVersion,
		Labels: configMap.Labels, Annotations: configMap.Annotations, Data: configMap.Data,
		BinaryData: make(map[string]string, len(configMap.BinaryData)), Events: make([]*api.EventInfo, 0),
	}
	if len(configMap.OwnerReferences) > 0 {
		owners := make([]string, 0, len(configMap.OwnerReferences))
		for _, owner := range configMap.OwnerReferences {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	for key, value := range configMap.BinaryData {
		result.BinaryData[key] = base64.StdEncoding.EncodeToString(value)
	}
	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Items {
		if !strings.EqualFold(event.InvolvedObject.Kind, "ConfigMap") || event.InvolvedObject.Name != configMap.Name {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}
	var description strings.Builder
	fmt.Fprintf(&description, "Name:         %s\nNamespace:    %s\nLabels:       %s\nAnnotations:  %s\n", configMap.Name, configMap.Namespace, formatKubernetesMetadata(configMap.Labels), formatKubernetesMetadata(configMap.Annotations))
	appendConfigMapDescribeSection(&description, "Data", configMap.Data, nil)
	appendConfigMapDescribeSection(&description, "BinaryData", nil, configMap.BinaryData)
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

func formatKubernetesMetadata(values map[string]string) string {
	if len(values) == 0 {
		return "<none>"
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]string, 0, len(keys))
	for _, key := range keys {
		items = append(items, key+"="+values[key])
	}
	return strings.Join(items, ",")
}

func appendConfigMapDescribeSection(output *strings.Builder, title string, values map[string]string, binaryValues map[string][]byte) {
	if len(values) == 0 && len(binaryValues) == 0 {
		return
	}
	fmt.Fprintf(output, "%s\n====\n", title)
	keys := make([]string, 0, len(values)+len(binaryValues))
	for key := range values {
		keys = append(keys, key)
	}
	for key := range binaryValues {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if value, ok := values[key]; ok {
			fmt.Fprintf(output, "%s:\n----\n%s\n", key, value)
			continue
		}
		fmt.Fprintf(output, "%s: %d bytes\n", key, len(binaryValues[key]))
	}
}
