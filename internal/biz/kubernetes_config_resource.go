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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
)

var supportedConfigDetailResources = map[string]struct{}{
	"resourcequotas": {}, "limitranges": {}, "horizontalpodautoscalers": {}, "poddisruptionbudgets": {},
}

func (b *KubernetesBiz) ConfigResourceSupport(req *api.KubernetesConfigResourceSupportRequest) (*api.KubernetesConfigResourceSupportResponse, error) {
	if _, ok := supportedConfigDetailResources[req.Resource]; !ok {
		return nil, fmt.Errorf("unsupported configuration resource %q", req.Resource)
	}
	spec, err := resourceSpec(req.Resource)
	if err != nil {
		return nil, err
	}
	client, err := b.dynamicClient(req.ClusterID)
	if err != nil {
		return nil, err
	}
	coreClient, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resource := client.Resource(spec.gvr)
	item, err := resource.Namespace(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesConfigResourceSupportResponse{
		Kind: spec.kind, Name: item.GetName(), Namespace: item.GetNamespace(),
		CreatedAt: item.GetCreationTimestamp().Format("2006-01-02 15:04:05"), Age: utils.CalculateAge(item.GetCreationTimestamp().Time),
		UID: string(item.GetUID()), ResourceVersion: item.GetResourceVersion(), Status: objectStatus(*item, spec.kind),
		Labels: item.GetLabels(), Annotations: item.GetAnnotations(),
		Properties: make([]*api.KubernetesConfigResourcePropertyResponse, 0),
		Related:    make([]*api.KubernetesRelatedResourceResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	if len(item.GetOwnerReferences()) > 0 {
		owners := make([]string, 0, len(item.GetOwnerReferences()))
		for _, owner := range item.GetOwnerReferences() {
			owners = append(owners, owner.Kind+"/"+owner.Name)
		}
		result.Owner = strings.Join(owners, ", ")
	}
	addProperty := func(name, value string) {
		if value != "" {
			result.Properties = append(result.Properties, &api.KubernetesConfigResourcePropertyResponse{Name: name, Value: value})
		}
	}
	addMap := func(title string, value map[string]any) {
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			addProperty(title+" · "+key, fmt.Sprint(value[key]))
		}
	}
	switch req.Resource {
	case "resourcequotas":
		addMap("Hard", nestedConfigMap(item.Object, "status", "hard"))
		addMap("Used", nestedConfigMap(item.Object, "status", "used"))
	case "limitranges":
		limits, _, _ := unstructured.NestedSlice(item.Object, "spec", "limits")
		for index, rawLimit := range limits {
			limit, ok := rawLimit.(map[string]any)
			if !ok {
				continue
			}
			prefix := fmt.Sprintf("%s #%d", fmt.Sprint(limit["type"]), index+1)
			for _, field := range []string{"max", "min", "default", "defaultRequest", "maxLimitRequestRatio"} {
				if values, ok := limit[field].(map[string]any); ok {
					addMap(prefix+" · "+field, values)
				}
			}
		}
	case "horizontalpodautoscalers":
		if target, _, _ := unstructured.NestedMap(item.Object, "spec", "scaleTargetRef"); target != nil {
			kind, name := fmt.Sprint(target["kind"]), fmt.Sprint(target["name"])
			addProperty("目标", kind+"/"+name)
			if kind != "" && name != "" {
				result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: kind, Name: name, Status: "伸缩目标"})
			}
		}
		for _, field := range []string{"minReplicas", "maxReplicas"} {
			if value, found, _ := unstructured.NestedFieldNoCopy(item.Object, "spec", field); found {
				addProperty(field, fmt.Sprint(value))
			}
		}
		for _, field := range []string{"currentReplicas", "desiredReplicas"} {
			if value, found, _ := unstructured.NestedFieldNoCopy(item.Object, "status", field); found {
				addProperty(field, fmt.Sprint(value))
			}
		}
	case "poddisruptionbudgets":
		for _, field := range []string{"minAvailable", "maxUnavailable"} {
			if value, found, _ := unstructured.NestedFieldNoCopy(item.Object, "spec", field); found {
				addProperty(field, fmt.Sprint(value))
			}
		}
		for _, field := range []string{"currentHealthy", "desiredHealthy", "disruptionsAllowed", "expectedPods"} {
			if value, found, _ := unstructured.NestedFieldNoCopy(item.Object, "status", field); found {
				addProperty(field, fmt.Sprint(value))
			}
		}
	}
	selector := fields.AndSelectors(fields.OneTermEqualSelector("involvedObject.kind", spec.kind), fields.OneTermEqualSelector("involvedObject.name", item.GetName())).String()
	if events, listErr := coreClient.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{FieldSelector: selector}); listErr == nil {
		for _, event := range events.Items {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}
	var description strings.Builder
	fmt.Fprintf(&description, "Name: %s\nNamespace: %s\nStatus: %s\n", result.Name, result.Namespace, result.Status)
	for _, property := range result.Properties {
		fmt.Fprintf(&description, "%s: %s\n", property.Name, property.Value)
	}
	if len(result.Events) == 0 {
		description.WriteString("Events: <none>\n")
	} else {
		description.WriteString("Events:\n")
		for _, event := range result.Events {
			fmt.Fprintf(&description, "  %-8s %-20s %d  %s\n", event.Type, event.Reason, event.Count, event.Message)
		}
	}
	result.Description = description.String()
	return result, nil
}

func nestedConfigMap(object map[string]any, fields ...string) map[string]any {
	value, found, err := unstructured.NestedMap(object, fields...)
	if err != nil || !found {
		return map[string]any{}
	}
	return value
}
