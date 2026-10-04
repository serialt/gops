package biz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	pathpkg "path"
	"sort"
	"strconv"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func (b *KubernetesBiz) DeploymentSupport(req *api.KubernetesDeploymentDetailRequest) (*api.KubernetesDeploymentSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	deployment, err := client.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	selector := labels.SelectorFromSet(deployment.Spec.Selector.MatchLabels).String()
	replicaSets, err := client.AppsV1().ReplicaSets(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	services, err := client.CoreV1().Services(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	events, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.kind=Deployment,involvedObject.name=" + deployment.Name})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesDeploymentSupportResponse{Related: make([]*api.KubernetesRelatedResourceResponse, 0), History: make([]*api.KubernetesReplicaSetResponse, 0), Metrics: make([]*api.KubernetesPodMetricResponse, 0)}
	result.UID = string(deployment.UID)
	result.Revision = deployment.Annotations["deployment.kubernetes.io/revision"]
	result.MinReadySeconds = deployment.Spec.MinReadySeconds
	result.ProgressDeadlineSeconds = deployment.Spec.ProgressDeadlineSeconds
	if len(deployment.OwnerReferences) > 0 {
		result.Owner = deployment.OwnerReferences[0].Kind + "/" + deployment.OwnerReferences[0].Name
	}
	for _, service := range services.Items {
		if selectorMatches(service.Spec.Selector, deployment.Spec.Template.Labels) {
			result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "Service", Name: service.Name, Status: string(service.Spec.Type)})
		}
	}
	for _, replicaSet := range replicaSets.Items {
		images := make([]string, 0, len(replicaSet.Spec.Template.Spec.Containers))
		for _, container := range replicaSet.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
		result.History = append(result.History, &api.KubernetesReplicaSetResponse{Name: replicaSet.Name, Revision: replicaSet.Annotations["deployment.kubernetes.io/revision"], Replicas: replicaSet.Status.Replicas, ReadyReplicas: replicaSet.Status.ReadyReplicas, Images: images, CreatedAt: replicaSet.CreationTimestamp.Format("2006-01-02 15:04:05"), ChangeReason: replicaSet.Annotations["kubernetes.io/change-cause"]})
	}
	result.Description = formatDeploymentDescription(deployment, replicaSets.Items, events.Items)
	metricsClient, metricsErr := b.deploymentMetricsClient(req.ClusterID)
	if metricsErr == nil {
		metrics, getErr := metricsClient.MetricsV1beta1().PodMetricses(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if getErr == nil {
			for _, pod := range metrics.Items {
				var cpu, memory int64
				for _, container := range pod.Containers {
					cpu += container.Usage.Cpu().MilliValue()
					memory += container.Usage.Memory().Value()
				}
				result.Metrics = append(result.Metrics, &api.KubernetesPodMetricResponse{Pod: pod.Name, CPU: formatDeploymentCPU(cpu), Memory: formatDeploymentMemory(memory)})
			}
		}
	}
	return result, nil
}

func formatDeploymentDescription(deployment *appsv1.Deployment, replicaSets []appsv1.ReplicaSet, events []corev1.Event) string {
	var output strings.Builder
	writeDescriptionField := func(label string, value any) { fmt.Fprintf(&output, "%-24s%v\n", label+":", value) }
	writeDescriptionField("Name", deployment.Name)
	writeDescriptionField("Namespace", deployment.Namespace)
	writeDescriptionField("CreationTimestamp", deployment.CreationTimestamp.Format(time.RFC1123Z))
	writeDescriptionField("Labels", formatDescriptionMap(deployment.Labels))
	writeDescriptionField("Annotations", formatDescriptionMap(deployment.Annotations))
	writeDescriptionField("Selector", labels.SelectorFromSet(deployment.Spec.Selector.MatchLabels).String())
	writeDescriptionField("Replicas", fmt.Sprintf("%d desired | %d updated | %d total | %d available | %d unavailable", valueOrZero(deployment.Spec.Replicas), deployment.Status.UpdatedReplicas, deployment.Status.Replicas, deployment.Status.AvailableReplicas, deployment.Status.UnavailableReplicas))
	writeDescriptionField("StrategyType", deployment.Spec.Strategy.Type)
	writeDescriptionField("MinReadySeconds", deployment.Spec.MinReadySeconds)
	if deployment.Spec.Strategy.RollingUpdate != nil {
		writeDescriptionField("RollingUpdateStrategy", fmt.Sprintf("%s max unavailable, %s max surge", deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.String(), deployment.Spec.Strategy.RollingUpdate.MaxSurge.String()))
	}
	output.WriteString("Pod Template:\n")
	fmt.Fprintf(&output, "  Labels:               %s\n", formatDescriptionMap(deployment.Spec.Template.Labels))
	fmt.Fprintf(&output, "  Annotations:          %s\n", formatDescriptionMap(deployment.Spec.Template.Annotations))
	fmt.Fprintf(&output, "  Service Account:      %s\n", defaultDescriptionValue(deployment.Spec.Template.Spec.ServiceAccountName, "default"))
	output.WriteString("  Containers:\n")
	for _, container := range deployment.Spec.Template.Spec.Containers {
		fmt.Fprintf(&output, "   %s:\n", container.Name)
		fmt.Fprintf(&output, "    Image:              %s\n", container.Image)
		ports := make([]string, 0, len(container.Ports))
		for _, port := range container.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol))
		}
		fmt.Fprintf(&output, "    Ports:              %s\n", defaultDescriptionValue(strings.Join(ports, ", "), "<none>"))
		fmt.Fprintf(&output, "    Command:            %s\n", defaultDescriptionValue(strings.Join(container.Command, " "), "<none>"))
		fmt.Fprintf(&output, "    Args:               %s\n", defaultDescriptionValue(strings.Join(container.Args, " "), "<none>"))
		fmt.Fprintf(&output, "    Requests:           CPU %s, Memory %s\n", defaultDescriptionValue(container.Resources.Requests.Cpu().String(), "0"), defaultDescriptionValue(container.Resources.Requests.Memory().String(), "0"))
		fmt.Fprintf(&output, "    Limits:             CPU %s, Memory %s\n", defaultDescriptionValue(container.Resources.Limits.Cpu().String(), "0"), defaultDescriptionValue(container.Resources.Limits.Memory().String(), "0"))
	}
	output.WriteString("Conditions:\n  Type                  Status  Reason\n")
	for _, condition := range deployment.Status.Conditions {
		fmt.Fprintf(&output, "  %-21s %-7s %s\n", condition.Type, condition.Status, condition.Reason)
		if condition.Message != "" {
			fmt.Fprintf(&output, "    Message: %s\n", condition.Message)
		}
	}
	oldReplicaSets := make([]string, 0)
	newReplicaSet := "<none>"
	currentRevision := deployment.Annotations["deployment.kubernetes.io/revision"]
	for _, replicaSet := range replicaSets {
		entry := fmt.Sprintf("%s (%d/%d replicas)", replicaSet.Name, replicaSet.Status.ReadyReplicas, replicaSet.Status.Replicas)
		if replicaSet.Annotations["deployment.kubernetes.io/revision"] == currentRevision {
			newReplicaSet = entry
		} else {
			oldReplicaSets = append(oldReplicaSets, entry)
		}
	}
	writeDescriptionField("OldReplicaSets", defaultDescriptionValue(strings.Join(oldReplicaSets, ", "), "<none>"))
	writeDescriptionField("NewReplicaSet", newReplicaSet)
	output.WriteString("Events:\n")
	if len(events) == 0 {
		output.WriteString("  <none>\n")
		return output.String()
	}
	output.WriteString("  Type     Reason              Age                  From                 Message\n")
	for _, event := range events {
		stamp := event.LastTimestamp.Time
		if stamp.IsZero() {
			stamp = event.EventTime.Time
		}
		fmt.Fprintf(&output, "  %-8s %-19s %-20s %-20s %s\n", event.Type, event.Reason, stamp.Format("2006-01-02 15:04:05"), event.Source.Component, event.Message)
	}
	return output.String()
}

func formatDescriptionMap(values map[string]string) string {
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
	return strings.Join(items, ", ")
}

func defaultDescriptionValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func valueOrZero(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func (b *KubernetesBiz) PodLogs(req *api.KubernetesPodLogRequest) (*api.KubernetesPodLogResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	tail := req.TailLines
	var tailLines *int64
	if tail > 0 {
		tailLines = &tail
	}
	stream, err := client.CoreV1().Pods(req.Namespace).GetLogs(req.Pod, &corev1.PodLogOptions{Container: req.Container, Timestamps: req.Timestamps, TailLines: tailLines}).Stream(context.Background())
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	content, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	return &api.KubernetesPodLogResponse{Logs: string(content)}, nil
}

func (b *KubernetesBiz) PodExec(req *api.KubernetesPodExecRequest) (*api.KubernetesPodExecResponse, error) {
	config, err := b.kubernetesRESTConfig(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	request := client.CoreV1().RESTClient().Post().Resource("pods").Name(req.Pod).Namespace(req.Namespace).SubResource("exec").VersionedParams(&corev1.PodExecOptions{Container: req.Container, Command: []string{"sh", "-c", req.Command}, Stdout: true, Stderr: true}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr})
	response := &api.KubernetesPodExecResponse{Output: stdout.String(), Error: stderr.String()}
	if err != nil && response.Error == "" {
		response.Error = err.Error()
	}
	return response, nil
}

func (b *KubernetesBiz) PodFiles(req *api.KubernetesPodFileRequest) (*api.KubernetesPodFileListResponse, error) {
	directory, err := cleanPodPath(req.Path)
	if err != nil {
		return nil, err
	}
	command := fmt.Sprintf("for f in %s/* %s/.[!.]* %s/..?*; do [ -e \"$f\" ] || continue; stat -c '%%n|%%U|%%G|%%s|%%A|%%Y|%%F' \"$f\"; done", podShellQuote(directory), podShellQuote(directory), podShellQuote(directory))
	stdout, stderr, err := b.podExecIO(req.ClusterID, req.Namespace, req.Pod, req.Container, command, nil)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr.String())
	}
	result := &api.KubernetesPodFileListResponse{List: make([]*api.KubernetesPodFileResponse, 0)}
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}
		size, _ := strconv.ParseInt(fields[3], 10, 64)
		modified, _ := strconv.ParseInt(fields[5], 10, 64)
		result.List = append(result.List, &api.KubernetesPodFileResponse{Name: pathpkg.Base(fields[0]), Path: fields[0], UID: fields[1], GID: fields[2], Size: size, Mode: fields[4], ModifiedAt: modified, Directory: strings.Contains(strings.Join(fields[6:], " "), "directory")})
	}
	return result, nil
}

func (b *KubernetesBiz) UploadPodFile(req *api.KubernetesPodFileRequest, filename string, content io.Reader) error {
	directory, err := cleanPodPath(req.Path)
	if err != nil {
		return err
	}
	target := pathpkg.Join(directory, pathpkg.Base(filename))
	_, stderr, err := b.podExecIO(req.ClusterID, req.Namespace, req.Pod, req.Container, "cat > "+podShellQuote(target), content)
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}

func (b *KubernetesBiz) DownloadPodFile(req *api.KubernetesPodFileRequest) ([]byte, error) {
	filename, err := cleanPodPath(req.Path)
	if err != nil {
		return nil, err
	}
	stdout, stderr, err := b.podExecIO(req.ClusterID, req.Namespace, req.Pod, req.Container, "cat -- "+podShellQuote(filename), nil)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func (b *KubernetesBiz) podExecIO(clusterID int64, namespace, pod, container, command string, stdin io.Reader) (*bytes.Buffer, *bytes.Buffer, error) {
	config, err := b.kubernetesRESTConfig(clusterID)
	if err != nil {
		return nil, nil, err
	}
	client, err := b.client(clusterID)
	if err != nil {
		return nil, nil, err
	}
	request := client.CoreV1().RESTClient().Post().Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").VersionedParams(&corev1.PodExecOptions{Container: container, Command: []string{"sh", "-c", command}, Stdin: stdin != nil, Stdout: true, Stderr: true}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return nil, nil, err
	}
	var stdout, stderr bytes.Buffer
	err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{Stdin: stdin, Stdout: &stdout, Stderr: &stderr})
	return &stdout, &stderr, err
}

func cleanPodPath(value string) (string, error) {
	cleaned := pathpkg.Clean(value)
	if !strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("path must be absolute")
	}
	return cleaned, nil
}
func podShellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

func (b *KubernetesBiz) deploymentMetricsClient(clusterID int64) (*metricsv.Clientset, error) {
	config, err := b.kubernetesRESTConfig(clusterID)
	if err != nil {
		return nil, err
	}
	return metricsv.NewForConfig(config)
}

func (b *KubernetesBiz) kubernetesRESTConfig(clusterID int64) (*rest.Config, error) {
	var cluster model.KubernetesCluster
	if err := b.db.Select("kube_config").First(&cluster, clusterID).Error; err != nil {
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
}

func selectorMatches(selector, values map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for key, value := range selector {
		if values[key] != value {
			return false
		}
	}
	return true
}

func formatDeploymentCPU(value int64) string {
	if value >= 1000 {
		return fmt.Sprintf("%.2f Core", float64(value)/1000)
	}
	return fmt.Sprintf("%dm", value)
}
func formatDeploymentMemory(value int64) string {
	const mib = 1024 * 1024
	if value >= 1024*mib {
		return fmt.Sprintf("%.2f GiB", float64(value)/(1024*mib))
	}
	return fmt.Sprintf("%.1f MiB", float64(value)/mib)
}
