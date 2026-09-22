package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cache "github.com/patrickmn/go-cache"
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
	"nodes": k8sSpec("", "v1", "nodes", "Node", false), "namespaces": k8sSpec("", "v1", "namespaces", "Namespace", false), "pods": k8sSpec("", "v1", "pods", "Pod", true),
	"deployments": k8sSpec("apps", "v1", "deployments", "Deployment", true), "statefulsets": k8sSpec("apps", "v1", "statefulsets", "StatefulSet", true), "daemonsets": k8sSpec("apps", "v1", "daemonsets", "DaemonSet", true),
	"jobs": k8sSpec("batch", "v1", "jobs", "Job", true), "cronjobs": k8sSpec("batch", "v1", "cronjobs", "CronJob", true),
	"services": k8sSpec("", "v1", "services", "Service", true), "ingresses": k8sSpec("networking.k8s.io", "v1", "ingresses", "Ingress", true), "endpoints": k8sSpec("", "v1", "endpoints", "Endpoints", true), "networkpolicies": k8sSpec("networking.k8s.io", "v1", "networkpolicies", "NetworkPolicy", true),
	"configmaps": k8sSpec("", "v1", "configmaps", "ConfigMap", true), "secrets": k8sSpec("", "v1", "secrets", "Secret", true), "resourcequotas": k8sSpec("", "v1", "resourcequotas", "ResourceQuota", true), "limitranges": k8sSpec("", "v1", "limitranges", "LimitRange", true), "horizontalpodautoscalers": k8sSpec("autoscaling", "v2", "horizontalpodautoscalers", "HorizontalPodAutoscaler", true), "poddisruptionbudgets": k8sSpec("policy", "v1", "poddisruptionbudgets", "PodDisruptionBudget", true),
	"persistentvolumeclaims": k8sSpec("", "v1", "persistentvolumeclaims", "PersistentVolumeClaim", true), "persistentvolumes": k8sSpec("", "v1", "persistentvolumes", "PersistentVolume", false), "storageclasses": k8sSpec("storage.k8s.io", "v1", "storageclasses", "StorageClass", false),
	"serviceaccounts": k8sSpec("", "v1", "serviceaccounts", "ServiceAccount", true), "roles": k8sSpec("rbac.authorization.k8s.io", "v1", "roles", "Role", true), "rolebindings": k8sSpec("rbac.authorization.k8s.io", "v1", "rolebindings", "RoleBinding", true), "clusterroles": k8sSpec("rbac.authorization.k8s.io", "v1", "clusterroles", "ClusterRole", false), "clusterrolebindings": k8sSpec("rbac.authorization.k8s.io", "v1", "clusterrolebindings", "ClusterRoleBinding", false),
}

const deploymentYAMLTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: %s
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21.0
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
`

const statefulSetYAMLTemplate = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx-statefulset
  namespace: %s
spec:
  serviceName: nginx-headless
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21.0
        ports:
        - containerPort: 80
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
`

const daemonSetYAMLTemplate = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-daemonset
  namespace: %s
spec:
  selector:
    matchLabels:
      app: fluentd
  template:
    metadata:
      labels:
        app: fluentd
    spec:
      containers:
      - name: fluentd
        image: fluentd:v1.14.0
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
`

const jobYAMLTemplate = `apiVersion: batch/v1
kind: Job
metadata:
  name: pi-job
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl:5.34.0
        command: ["perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
  backoffLimit: 4
`

const cronJobYAMLTemplate = `apiVersion: batch/v1
kind: CronJob
metadata:
  name: hello-cronjob
  namespace: %s
spec:
  schedule: "*/1 * * * *"
  concurrencyPolicy: Allow
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox:1.36
            imagePullPolicy: IfNotPresent
            command:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
`

const podYAMLTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: debug
  namespace: %s
spec:
  containers:
    - name: debug
      image: nicolaka/netshoot
      command:
        - /bin/sh
      args:
        - '-c'
        - sleep 100000
      resources:
        limits:
          cpu: 100m
          memory: 128Mi
        requests:
          cpu: 50m
          memory: 64Mi
`

const serviceYAMLTemplate = `apiVersion: v1
kind: Service
metadata:
  name: example
  namespace: %s
spec:
  type: ClusterIP
  selector:
    app: example
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 80
`

const ingressYAMLTemplate = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example
  namespace: %s
spec:
  rules:
    - host: example.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: example
                port:
                  number: 80
`

const endpointsYAMLTemplate = `apiVersion: v1
kind: Endpoints
metadata:
  name: example
  namespace: %s
subsets:
  - addresses:
      - ip: 10.0.0.10
    ports:
      - name: http
        port: 80
        protocol: TCP
`

const networkPolicyYAMLTemplate = `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: example
  namespace: %s
spec:
  podSelector:
    matchLabels:
      app: example
  policyTypes:
    - Ingress
    - Egress
`

const configMapYAMLTemplate = `apiVersion: v1
kind: ConfigMap
metadata:
  name: example
  namespace: %s
data:
  key1: value1
  key2: value2
`

const secretYAMLTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: %s
type: Opaque
data:
  username: YWRtaW4=
  password: cGFzc3dvcmQ=

`

const resourceQuotaYAMLTemplate = `apiVersion: v1
kind: ResourceQuota
metadata:
  name: example
  namespace: %s
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    requests.storage: 100Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "50"
    persistentvolumeclaims: "10"
    services: "20"
    services.loadbalancers: "2"
    services.nodeports: "5"
    configmaps: "50"
    secrets: "50"
    count/deployments.apps: "20"
    count/statefulsets.apps: "10"
`

const limitRangeYAMLTemplate = `apiVersion: v1
kind: LimitRange
metadata:
  name: example
  namespace: %s
spec:
  limits:
    - type: Container
      max:
        cpu: "2"
        memory: 2Gi
      min:
        cpu: 50m
        memory: 64Mi
      default:
        cpu: 500m
        memory: 512Mi
      defaultRequest:
        cpu: 100m
        memory: 128Mi
      maxLimitRequestRatio:
        cpu: "10"
        memory: "4"
    - type: Pod
      max:
        cpu: "4"
        memory: 4Gi
      min:
        cpu: 50m
        memory: 64Mi
    - type: PersistentVolumeClaim
      max:
        storage: 50Gi
      min:
        storage: 1Gi
`

const horizontalPodAutoscalerYAMLTemplate = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: example
  namespace: %s
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: example
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0
      selectPolicy: Max
      policies:
        - type: Percent
          value: 100
          periodSeconds: 60
        - type: Pods
          value: 4
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      selectPolicy: Max
      policies:
        - type: Percent
          value: 25
          periodSeconds: 60
`

const podDisruptionBudgetYAMLTemplate = `apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: example
  namespace: %s
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: example
  unhealthyPodEvictionPolicy: IfHealthyBudget
`

func (b *KubernetesBiz) dynamicClient(clusterID int64) (dynamic.Interface, error) {
	b.cacheMutex.RLock()
	if cached := b.dynamicCache[clusterID]; cached != nil {
		b.cacheMutex.RUnlock()
		return cached, nil
	}
	b.cacheMutex.RUnlock()
	var cluster model.KubernetesCluster
	if err := b.db.First(&cluster, clusterID).Error; err != nil {
		return nil, k8sErr(err)
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	b.cacheMutex.Lock()
	if cached := b.dynamicCache[clusterID]; cached != nil {
		b.cacheMutex.Unlock()
		return cached, nil
	}
	b.dynamicCache[clusterID] = client
	b.cacheMutex.Unlock()
	return client, nil
}

func resourceSpec(name string) (kubernetesResourceSpec, error) {
	spec, ok := kubernetesResourceCatalog[name]
	if !ok {
		return spec, fmt.Errorf("unsupported kubernetes resource: %s", name)
	}
	return spec, nil
}

func (b *KubernetesBiz) ResourceTemplate(r *api.KubernetesObjectRequest) (*api.KubernetesYAMLResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
	}
	namespace := r.Namespace
	if namespace == "" {
		namespace = "default"
	}
	metadata := "metadata:\n  name: example\n"
	if spec.namespaced {
		metadata += fmt.Sprintf("  namespace: %s\n", namespace)
	}
	apiVersion := spec.gvr.Version
	if spec.gvr.Group != "" {
		apiVersion = spec.gvr.Group + "/" + spec.gvr.Version
	}
	prefix := fmt.Sprintf("apiVersion: %s\nkind: %s\n%s", apiVersion, spec.kind, metadata)
	var content string
	switch r.Resource {
	case "deployments":
		content = fmt.Sprintf(deploymentYAMLTemplate, namespace)
	case "statefulsets":
		content = fmt.Sprintf(statefulSetYAMLTemplate, namespace)
	case "daemonsets":
		content = fmt.Sprintf(daemonSetYAMLTemplate, namespace)
	case "jobs":
		content = fmt.Sprintf(jobYAMLTemplate, namespace)
	case "cronjobs":
		content = fmt.Sprintf(cronJobYAMLTemplate, namespace)
	case "pods":
		content = fmt.Sprintf(podYAMLTemplate, namespace)
	case "services":
		content = fmt.Sprintf(serviceYAMLTemplate, namespace)
	case "ingresses":
		content = fmt.Sprintf(ingressYAMLTemplate, namespace)
	case "endpoints":
		content = fmt.Sprintf(endpointsYAMLTemplate, namespace)
	case "networkpolicies":
		content = fmt.Sprintf(networkPolicyYAMLTemplate, namespace)
	case "configmaps":
		content = fmt.Sprintf(configMapYAMLTemplate, namespace)
	case "secrets":
		content = fmt.Sprintf(secretYAMLTemplate, namespace)
	case "resourcequotas":
		content = fmt.Sprintf(resourceQuotaYAMLTemplate, namespace)
	case "limitranges":
		content = fmt.Sprintf(limitRangeYAMLTemplate, namespace)
	case "horizontalpodautoscalers":
		content = fmt.Sprintf(horizontalPodAutoscalerYAMLTemplate, namespace)
	case "poddisruptionbudgets":
		content = fmt.Sprintf(podDisruptionBudgetYAMLTemplate, namespace)
	default:
		content = prefix
	}
	return &api.KubernetesYAMLResponse{YAML: content}, nil
}

func (b *KubernetesBiz) ListObjects(r *api.KubernetesObjectRequest) ([]*api.KubernetesObjectResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
	}
	cacheable := isCachedKubernetesResource(r.Resource)
	cacheKey := fmt.Sprintf("resource:list:%d:%s:%s", r.ClusterID, r.Resource, r.Namespace)
	if cacheable && !r.Refresh {
		if cached, found := b.workloadCache.Get(cacheKey); found {
			if values, ok := cached.([]*api.KubernetesObjectResponse); ok {
				return filterKubernetesObjects(values, r.Keyword), nil
			}
		}
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
		row := &api.KubernetesObjectResponse{Name: item.GetName(), Namespace: item.GetNamespace(), Kind: spec.kind, Status: objectStatus(item, spec.kind), Age: utils.CalculateAge(item.GetCreationTimestamp().Time), Labels: item.GetLabels(), CreatedAt: item.GetCreationTimestamp().Format("2006-01-02 15:04:05")}
		fillWorkloadObject(row, item)
		fillNetworkObject(row, item)
		out = append(out, row)
	}
	if cacheable {
		b.workloadCache.Set(cacheKey, out, cache.DefaultExpiration)
	}
	return filterKubernetesObjects(out, r.Keyword), nil
}

func isWorkloadResource(resource string) bool {
	switch resource {
	case "deployments", "statefulsets", "daemonsets", "jobs", "cronjobs", "pods":
		return true
	default:
		return false
	}
}

func isConfigResource(resource string) bool {
	switch resource {
	case "configmaps", "secrets", "resourcequotas", "limitranges", "horizontalpodautoscalers", "poddisruptionbudgets":
		return true
	default:
		return false
	}
}

func isStorageResource(resource string) bool {
	switch resource {
	case "persistentvolumeclaims", "persistentvolumes", "storageclasses":
		return true
	default:
		return false
	}
}

func isCachedKubernetesResource(resource string) bool {
	return isWorkloadResource(resource) || isConfigResource(resource) || isStorageResource(resource)
}

func (b *KubernetesBiz) invalidateWorkloadCache(resource string) {
	if isCachedKubernetesResource(resource) {
		b.workloadCache.Flush()
	}
}

func filterKubernetesObjects(rows []*api.KubernetesObjectResponse, keyword string) []*api.KubernetesObjectResponse {
	if keyword == "" {
		return rows
	}
	keyword = strings.ToLower(keyword)
	filtered := make([]*api.KubernetesObjectResponse, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Name), keyword) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func fillNetworkObject(row *api.KubernetesObjectResponse, item unstructured.Unstructured) {
	switch row.Kind {
	case "Service":
		row.ServiceType, _, _ = unstructured.NestedString(item.Object, "spec", "type")
		row.ClusterIP, _, _ = unstructured.NestedString(item.Object, "spec", "clusterIP")
		loadBalancer, _, _ := unstructured.NestedSlice(item.Object, "status", "loadBalancer", "ingress")
		row.ExternalIP = firstNetworkAddress(loadBalancer)
		ports, _, _ := unstructured.NestedSlice(item.Object, "spec", "ports")
		for _, raw := range ports {
			port, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			protocol, _ := port["protocol"].(string)
			if protocol == "" {
				protocol = "TCP"
			}
			value, target := fmt.Sprint(port["port"]), fmt.Sprint(port["targetPort"])
			if target == "" || target == "<nil>" {
				target = value
			}
			row.Ports = append(row.Ports, fmt.Sprintf("%s: %s → %s", protocol, value, target))
		}
	case "Ingress":
		rules, _, _ := unstructured.NestedSlice(item.Object, "spec", "rules")
		for _, raw := range rules {
			if rule, ok := raw.(map[string]interface{}); ok {
				if host, ok := rule["host"].(string); ok && host != "" {
					row.Hosts = append(row.Hosts, host)
				}
			}
		}
		loadBalancer, _, _ := unstructured.NestedSlice(item.Object, "status", "loadBalancer", "ingress")
		for _, raw := range loadBalancer {
			if value := networkAddress(raw); value != "" {
				row.Addresses = append(row.Addresses, value)
			}
		}
	case "Endpoints":
		subsets, _, _ := unstructured.NestedSlice(item.Object, "subsets")
		for _, raw := range subsets {
			subset, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if addresses, ok := subset["addresses"].([]interface{}); ok {
				row.EndpointCount += int64(len(addresses))
				for _, address := range addresses {
					if value := networkAddress(address); value != "" {
						row.Addresses = append(row.Addresses, value)
					}
				}
			}
			if ports, ok := subset["ports"].([]interface{}); ok {
				for _, portRaw := range ports {
					if port, ok := portRaw.(map[string]interface{}); ok {
						row.Ports = append(row.Ports, fmt.Sprintf("%v: %v", port["protocol"], port["port"]))
					}
				}
			}
		}
	case "NetworkPolicy":
		row.PolicyTypes, _, _ = unstructured.NestedStringSlice(item.Object, "spec", "policyTypes")
	}
}

func firstNetworkAddress(values []interface{}) string {
	for _, raw := range values {
		if value := networkAddress(raw); value != "" {
			return value
		}
	}
	return ""
}

func networkAddress(raw interface{}) string {
	value, ok := raw.(map[string]interface{})
	if !ok {
		return ""
	}
	if ip, ok := value["ip"].(string); ok && ip != "" {
		return ip
	}
	if hostname, ok := value["hostname"].(string); ok {
		return hostname
	}
	return ""
}

func fillWorkloadObject(row *api.KubernetesObjectResponse, item unstructured.Unstructured) {
	if desired, found, _ := unstructured.NestedInt64(item.Object, "spec", "replicas"); found {
		ready, _, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas")
		row.Ready = fmt.Sprintf("%d/%d", ready, desired)
	}
	if item.GetKind() == "DaemonSet" || row.Kind == "DaemonSet" {
		desired, _, _ := unstructured.NestedInt64(item.Object, "status", "desiredNumberScheduled")
		ready, _, _ := unstructured.NestedInt64(item.Object, "status", "numberReady")
		row.Ready = fmt.Sprintf("%d/%d", ready, desired)
	}
	if row.Kind == "Job" {
		desired, found, _ := unstructured.NestedInt64(item.Object, "spec", "completions")
		if !found {
			desired = 1
		}
		ready, _, _ := unstructured.NestedInt64(item.Object, "status", "succeeded")
		row.Ready = fmt.Sprintf("%d/%d", ready, desired)
	}
	if row.Kind == "CronJob" {
		row.Schedule, _, _ = unstructured.NestedString(item.Object, "spec", "schedule")
		row.Suspended, _, _ = unstructured.NestedBool(item.Object, "spec", "suspend")
		row.LastScheduleTime, _, _ = unstructured.NestedString(item.Object, "status", "lastScheduleTime")
	}
	containers, found, _ := unstructured.NestedSlice(item.Object, "spec", "template", "spec", "containers")
	if row.Kind == "Pod" {
		containers, found, _ = unstructured.NestedSlice(item.Object, "spec", "containers")
	}
	requests := &api.KubernetesResourceValues{}
	limits := &api.KubernetesResourceValues{}
	for _, raw := range containers {
		container, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if image, ok := container["image"].(string); ok {
			row.Images = append(row.Images, image)
		}
		resources, _ := container["resources"].(map[string]interface{})
		addResourceValues(requests, resources["requests"])
		addResourceValues(limits, resources["limits"])
	}
	if found {
		row.Requests, row.Limits = requests, limits
	}
	if row.Kind == "Pod" {
		row.Ready = podReady(item)
		row.PodIP, _, _ = unstructured.NestedString(item.Object, "status", "podIP")
		row.Node, _, _ = unstructured.NestedString(item.Object, "spec", "nodeName")
		statuses, _, _ := unstructured.NestedSlice(item.Object, "status", "containerStatuses")
		for _, raw := range statuses {
			if status, ok := raw.(map[string]interface{}); ok {
				if value, ok := status["restartCount"].(int64); ok {
					row.Restarts += value
				}
			}
		}
	}
}

func addResourceValues(target *api.KubernetesResourceValues, raw interface{}) {
	values, ok := raw.(map[string]interface{})
	if !ok {
		return
	}
	if value, ok := values["cpu"].(string); ok {
		target.CPU = joinResource(target.CPU, value)
	}
	if value, ok := values["memory"].(string); ok {
		target.Memory = joinResource(target.Memory, value)
	}
}

func joinResource(current, value string) string {
	if current == "" {
		return value
	}
	return current + " + " + value
}

func podReady(item unstructured.Unstructured) string {
	statuses, _, _ := unstructured.NestedSlice(item.Object, "status", "containerStatuses")
	ready := 0
	for _, raw := range statuses {
		if status, ok := raw.(map[string]interface{}); ok {
			if value, _ := status["ready"].(bool); value {
				ready++
			}
		}
	}
	return fmt.Sprintf("%d/%d", ready, len(statuses))
}

func objectStatus(item unstructured.Unstructured, kind string) string {
	for _, path := range [][]string{{"status", "phase"}, {"spec", "type"}} {
		if value, found, _ := unstructured.NestedString(item.Object, path...); found {
			return value
		}
	}
	conditions, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions")
	for _, raw := range conditions {
		condition, ok := raw.(map[string]interface{})
		if !ok || fmt.Sprint(condition["status"]) != "True" {
			continue
		}
		if value, ok := condition["type"].(string); ok && value != "" {
			return value
		}
	}
	switch kind {
	case "ConfigMap", "Secret", "ResourceQuota", "LimitRange":
		return "Active"
	case "HorizontalPodAutoscaler", "PodDisruptionBudget":
		return "Ready"
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
	cleanObjectForYAML(item)
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

func cleanObjectForYAML(item *unstructured.Unstructured) {
	for _, field := range []string{"uid", "resourceVersion", "generation", "creationTimestamp", "managedFields", "selfLink"} {
		unstructured.RemoveNestedField(item.Object, "metadata", field)
	}
	unstructured.RemoveNestedField(item.Object, "status")
	annotations := item.GetAnnotations()
	delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	if len(annotations) == 0 {
		unstructured.RemoveNestedField(item.Object, "metadata", "annotations")
	} else {
		item.SetAnnotations(annotations)
	}
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
	b.invalidateWorkloadCache(r.Resource)
	return &api.KubernetesObjectResponse{Name: saved.GetName(), Namespace: saved.GetNamespace(), Kind: spec.kind, Status: objectStatus(*saved, spec.kind), Age: utils.CalculateAge(saved.GetCreationTimestamp().Time), Labels: saved.GetLabels(), CreatedAt: saved.GetCreationTimestamp().Format("2006-01-02 15:04:05")}, nil
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
	var deleteErr error
	if spec.namespaced {
		deleteErr = resource.Namespace(r.Namespace).Delete(context.Background(), r.Name, metav1.DeleteOptions{})
	} else {
		deleteErr = resource.Delete(context.Background(), r.Name, metav1.DeleteOptions{})
	}
	if deleteErr == nil {
		b.invalidateWorkloadCache(r.Resource)
	}
	return deleteErr
}

func (b *KubernetesBiz) UpdateObjectLabels(r *api.KubernetesLabelUpdateRequest) (*api.KubernetesObjectResponse, error) {
	spec, err := resourceSpec(r.Resource)
	if err != nil {
		return nil, err
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
	item.SetLabels(r.Labels)
	if spec.namespaced {
		item, err = resource.Namespace(r.Namespace).Update(context.Background(), item, metav1.UpdateOptions{})
	} else {
		item, err = resource.Update(context.Background(), item, metav1.UpdateOptions{})
	}
	if err != nil {
		return nil, err
	}
	b.invalidateWorkloadCache(r.Resource)
	return &api.KubernetesObjectResponse{Name: item.GetName(), Namespace: item.GetNamespace(), Kind: spec.kind, Status: objectStatus(*item, spec.kind), Age: utils.CalculateAge(item.GetCreationTimestamp().Time), Labels: item.GetLabels(), CreatedAt: item.GetCreationTimestamp().Format("2006-01-02 15:04:05")}, nil
}
