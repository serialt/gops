package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type kubernetesResourceSpec struct {
	gvr        schema.GroupVersionResource
	kind       string
	namespaced bool
}

func k8sSpec(group, version, resource, kind string, namespaced bool) kubernetesResourceSpec {
	return kubernetesResourceSpec{gvr: schema.GroupVersionResource{Group: group, Version: version, Resource: resource}, kind: kind, namespaced: namespaced}
}

var kubernetesResourceCatalog = map[string]kubernetesResourceSpec{
	"namespaces": k8sSpec("", "v1", "namespaces", "Namespace", false), "pods": k8sSpec("", "v1", "pods", "Pod", true),
	"deployments": k8sSpec("apps", "v1", "deployments", "Deployment", true), "statefulsets": k8sSpec("apps", "v1", "statefulsets", "StatefulSet", true), "daemonsets": k8sSpec("apps", "v1", "daemonsets", "DaemonSet", true),
	"jobs": k8sSpec("batch", "v1", "jobs", "Job", true), "cronjobs": k8sSpec("batch", "v1", "cronjobs", "CronJob", true),
	"services": k8sSpec("", "v1", "services", "Service", true), "ingresses": k8sSpec("networking.k8s.io", "v1", "ingresses", "Ingress", true), "endpoints": k8sSpec("", "v1", "endpoints", "Endpoints", true), "networkpolicies": k8sSpec("networking.k8s.io", "v1", "networkpolicies", "NetworkPolicy", true),
	"configmaps": k8sSpec("", "v1", "configmaps", "ConfigMap", true), "secrets": k8sSpec("", "v1", "secrets", "Secret", true), "resourcequotas": k8sSpec("", "v1", "resourcequotas", "ResourceQuota", true), "limitranges": k8sSpec("", "v1", "limitranges", "LimitRange", true), "horizontalpodautoscalers": k8sSpec("autoscaling", "v2", "horizontalpodautoscalers", "HorizontalPodAutoscaler", true), "poddisruptionbudgets": k8sSpec("policy", "v1", "poddisruptionbudgets", "PodDisruptionBudget", true),
	"persistentvolumeclaims": k8sSpec("", "v1", "persistentvolumeclaims", "PersistentVolumeClaim", true), "persistentvolumes": k8sSpec("", "v1", "persistentvolumes", "PersistentVolume", false), "storageclasses": k8sSpec("storage.k8s.io", "v1", "storageclasses", "StorageClass", false),
	"serviceaccounts": k8sSpec("", "v1", "serviceaccounts", "ServiceAccount", true), "roles": k8sSpec("rbac.authorization.k8s.io", "v1", "roles", "Role", true), "rolebindings": k8sSpec("rbac.authorization.k8s.io", "v1", "rolebindings", "RoleBinding", true), "clusterroles": k8sSpec("rbac.authorization.k8s.io", "v1", "clusterroles", "ClusterRole", false), "clusterrolebindings": k8sSpec("rbac.authorization.k8s.io", "v1", "clusterrolebindings", "ClusterRoleBinding", false),
}

func (b *KubernetesBiz) dynamicClient(clusterID int64) (dynamic.Interface, error) {
	var cluster model.KubernetesCluster
	if err := b.db.First(&cluster, clusterID).Error; err != nil {
		return nil, k8sErr(err)
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	return dynamic.NewForConfig(config)
}

func resourceSpec(name string) (kubernetesResourceSpec, error) {
	spec, ok := kubernetesResourceCatalog[name]
	if !ok {
		return spec, fmt.Errorf("unsupported kubernetes resource: %s", name)
	}
	return spec, nil
}

func (b *KubernetesBiz) ListObjects(r *api.KubernetesObjectRequest) ([]*api.KubernetesObjectResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
	}
	client, err := b.dynamicClient(r.ClusterID)
	if err != nil {
		return nil, err
	}
	resource := client.Resource(spec.gvr)
	var rows *unstructured.UnstructuredList
	if spec.namespaced {
		namespace := r.Namespace
		if namespace == "" {
			namespace = metav1.NamespaceAll
		}
		rows, err = resource.Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	} else {
		rows, err = resource.List(context.Background(), metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	out := make([]*api.KubernetesObjectResponse, 0, len(rows.Items))
	for _, item := range rows.Items {
		if r.Keyword != "" && !strings.Contains(strings.ToLower(item.GetName()), strings.ToLower(r.Keyword)) {
			continue
		}
		out = append(out, &api.KubernetesObjectResponse{Name: item.GetName(), Namespace: item.GetNamespace(), Kind: spec.kind, Status: objectStatus(item), Age: utils.CalculateAge(item.GetCreationTimestamp().Time), Labels: item.GetLabels(), CreatedAt: item.GetCreationTimestamp().Format("2006-01-02 15:04:05")})
	}
	return out, nil
}

func objectStatus(item unstructured.Unstructured) string {
	for _, path := range [][]string{{"status", "phase"}, {"status", "conditions", "0", "type"}, {"spec", "type"}} {
		if value, found, _ := unstructured.NestedString(item.Object, path...); found {
			return value
		}
	}
	return "-"
}

func (b *KubernetesBiz) GetObjectYAML(r *api.KubernetesObjectRequest) (*api.KubernetesYAMLResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
	}
	if r.Name == "" {
		return nil, fmt.Errorf("resource name is required")
	}
	client, err := b.dynamicClient(r.ClusterID)
	if err != nil {
		return nil, err
	}
	resource := client.Resource(spec.gvr)
	var item *unstructured.Unstructured
	if spec.namespaced {
		item, err = resource.Namespace(r.Namespace).Get(context.Background(), r.Name, metav1.GetOptions{})
	} else {
		item, err = resource.Get(context.Background(), r.Name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	unstructured.RemoveNestedField(item.Object, "metadata", "managedFields")
	data, err := json.Marshal(item.Object)
	if err != nil {
		return nil, err
	}
	content, err := yaml.JSONToYAML(data)
	if err != nil {
		return nil, err
	}
	return &api.KubernetesYAMLResponse{YAML: string(content)}, nil
}

func (b *KubernetesBiz) ApplyObject(r *api.KubernetesObjectRequest) (*api.KubernetesObjectResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
	}
	raw, err := yaml.YAMLToJSON([]byte(r.YAML))
	if err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	item := new(unstructured.Unstructured)
	if err = json.Unmarshal(raw, &item.Object); err != nil {
		return nil, err
	}
	if item.GetName() == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}
	item.SetGroupVersionKind(schema.GroupVersionKind{Group: spec.gvr.Group, Version: spec.gvr.Version, Kind: spec.kind})
	client, err := b.dynamicClient(r.ClusterID)
	if err != nil {
		return nil, err
	}
	resource := client.Resource(spec.gvr)
	var saved *unstructured.Unstructured
	if spec.namespaced {
		namespace := item.GetNamespace()
		if namespace == "" {
			namespace = r.Namespace
		}
		if namespace == "" {
			namespace = "default"
		}
		item.SetNamespace(namespace)
		existing, getErr := resource.Namespace(namespace).Get(context.Background(), item.GetName(), metav1.GetOptions{})
		if getErr == nil {
			item.SetResourceVersion(existing.GetResourceVersion())
			saved, err = resource.Namespace(namespace).Update(context.Background(), item, metav1.UpdateOptions{})
		} else {
			saved, err = resource.Namespace(namespace).Create(context.Background(), item, metav1.CreateOptions{})
		}
	} else {
		existing, getErr := resource.Get(context.Background(), item.GetName(), metav1.GetOptions{})
		if getErr == nil {
			item.SetResourceVersion(existing.GetResourceVersion())
			saved, err = resource.Update(context.Background(), item, metav1.UpdateOptions{})
		} else {
			saved, err = resource.Create(context.Background(), item, metav1.CreateOptions{})
		}
	}
	if err != nil {
		return nil, err
	}
	return &api.KubernetesObjectResponse{Name: saved.GetName(), Namespace: saved.GetNamespace(), Kind: spec.kind, Status: objectStatus(*saved), Age: utils.CalculateAge(saved.GetCreationTimestamp().Time), Labels: saved.GetLabels(), CreatedAt: saved.GetCreationTimestamp().Format("2006-01-02 15:04:05")}, nil
}

func (b *KubernetesBiz) DeleteObject(r *api.KubernetesObjectRequest) error {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return err
	}
	if r.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	client, err := b.dynamicClient(r.ClusterID)
	if err != nil {
		return err
	}
	resource := client.Resource(spec.gvr)
	if spec.namespaced {
		return resource.Namespace(r.Namespace).Delete(context.Background(), r.Name, metav1.DeleteOptions{})
	}
	return resource.Delete(context.Background(), r.Name, metav1.DeleteOptions{})
}
