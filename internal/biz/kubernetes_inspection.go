package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (b *KubernetesBiz) StartInspection(r *api.KubernetesInspectionStartRequest) (*api.KubernetesInspectionStartResponse, error) {
	var cluster model.KubernetesCluster
	if err := b.db.Select("id", "name", "alias").First(&cluster, r.ClusterID).Error; err != nil {
		return nil, k8sErr(err)
	}
	name := cluster.Alias_
	if name == "" {
		name = cluster.Name
	}
	now := time.Now()
	record := &model.KubernetesClusterInspection{ClusterID: cluster.ID, ClusterName: name, Status: "running", UserID: r.UserID, StartedAt: now, CreatedAt: now, UpdatedAt: now}
	if err := b.db.Create(record).Error; err != nil {
		return nil, err
	}
	go b.runInspection(record)
	return &api.KubernetesInspectionStartResponse{ID: record.ID}, nil
}

type inspectionSnapshot struct {
	version      string
	nodes        *corev1.NodeList
	pods         *corev1.PodList
	deployments  *appsv1.DeploymentList
	daemonSets   *appsv1.DaemonSetList
	statefulSets *appsv1.StatefulSetList
	services     *corev1.ServiceList
	endpoints    *corev1.EndpointsList
	pvs          *corev1.PersistentVolumeList
	pvcs         *corev1.PersistentVolumeClaimList
	events       *corev1.EventList
}

func (b *KubernetesBiz) runInspection(record *model.KubernetesClusterInspection) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, err := b.client(record.ClusterID)
	if err != nil {
		b.failInspection(record.ID, err)
		return
	}
	snapshot, err := collectInspectionSnapshot(ctx, client)
	if err != nil {
		b.failInspection(record.ID, err)
		return
	}
	report := buildInspectionReport(record.ClusterID, record.ClusterName, snapshot)
	report.Summary.Duration = int(time.Since(started).Seconds())
	data, err := json.Marshal(report)
	if err != nil {
		b.failInspection(record.ID, err)
		return
	}
	ended := time.Now()
	_ = b.db.Model(&model.KubernetesClusterInspection{}).Where("id = ?", record.ID).Updates(map[string]any{
		"status": "completed", "score": report.Score, "check_count": report.Summary.Total, "pass_count": report.Summary.Passed,
		"warning_count": report.Summary.Warning, "fail_count": report.Summary.Failed, "duration": report.Summary.Duration,
		"report_data": string(data), "ended_at": ended, "updated_at": ended,
	}).Error
}

func collectInspectionSnapshot(ctx context.Context, client *kubernetes.Clientset) (*inspectionSnapshot, error) {
	s := &inspectionSnapshot{}
	type task func() error
	tasks := []task{
		func() error {
			v, e := client.Discovery().ServerVersion()
			if e == nil {
				s.version = v.GitVersion
			}
			return e
		},
		func() error { v, e := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); s.nodes = v; return e },
		func() error {
			v, e := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.pods = v
			return e
		},
		func() error {
			v, e := client.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.deployments = v
			return e
		},
		func() error {
			v, e := client.AppsV1().DaemonSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.daemonSets = v
			return e
		},
		func() error {
			v, e := client.AppsV1().StatefulSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.statefulSets = v
			return e
		},
		func() error {
			v, e := client.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.services = v
			return e
		},
		func() error {
			v, e := client.CoreV1().Endpoints(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.endpoints = v
			return e
		},
		func() error {
			v, e := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
			s.pvs = v
			return e
		},
		func() error {
			v, e := client.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.pvcs = v
			return e
		},
		func() error {
			v, e := client.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			s.events = v
			return e
		},
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))
	for _, run := range tasks {
		wg.Add(1)
		go func(run task) {
			defer wg.Done()
			if err := run(); err != nil {
				errCh <- err
			}
		}(run)
	}
	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		return nil, err
	}
	return s, nil
}

func check(category, name, status, value, expected, detail, suggestion string) *api.KubernetesInspectionCheck {
	return &api.KubernetesInspectionCheck{Category: category, Name: name, Status: status, Value: value, Expected: expected, Detail: detail, Suggestion: suggestion}
}

func buildInspectionReport(clusterID int64, clusterName string, s *inspectionSnapshot) *api.KubernetesInspectionReport {
	sections := []*api.KubernetesInspectionSection{}
	add := func(key, title string, checks ...*api.KubernetesInspectionCheck) {
		sections = append(sections, &api.KubernetesInspectionSection{Key: key, Title: title, Checks: checks})
	}
	add("cluster", "集群信息", check("集群", "API Server 连接", "success", s.version, "可连接", "Kubernetes API Server 连接正常", ""))
	ready, pressure := 0, 0
	for _, node := range s.nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				ready++
			}
			if (condition.Type == corev1.NodeMemoryPressure || condition.Type == corev1.NodeDiskPressure || condition.Type == corev1.NodePIDPressure) && condition.Status == corev1.ConditionTrue {
				pressure++
			}
		}
	}
	nodeStatus := "success"
	if ready < len(s.nodes.Items) {
		nodeStatus = "error"
	} else if pressure > 0 {
		nodeStatus = "warning"
	}
	add("nodes", "节点健康", check("节点", "节点就绪状态", nodeStatus, fmt.Sprintf("%d/%d", ready, len(s.nodes.Items)), "全部 Ready", fmt.Sprintf("资源压力节点：%d", pressure), "检查异常节点状态、资源和 kubelet 日志"))
	running, pending, failed, restarts, noLimits := 0, 0, 0, 0, 0
	for _, pod := range s.pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			running++
		case corev1.PodPending:
			pending++
		case corev1.PodFailed:
			failed++
		}
		limited := true
		for _, c := range pod.Spec.Containers {
			if c.Resources.Requests.Cpu().IsZero() || c.Resources.Requests.Memory().IsZero() || c.Resources.Limits.Cpu().IsZero() || c.Resources.Limits.Memory().IsZero() {
				limited = false
			}
		}
		if !limited {
			noLimits++
		}
		for _, c := range pod.Status.ContainerStatuses {
			if c.RestartCount > 5 {
				restarts++
			}
		}
	}
	workloadStatus := "success"
	if failed > 0 {
		workloadStatus = "error"
	} else if pending > 0 || restarts > 0 {
		workloadStatus = "warning"
	}
	add("workloads", "工作负载", check("工作负载", "Pod 运行状态", workloadStatus, fmt.Sprintf("运行 %d / 总数 %d", running, len(s.pods.Items)), "无失败或长期 Pending", fmt.Sprintf("Pending %d，Failed %d，高重启容器 %d", pending, failed, restarts), "检查 Pod 事件、镜像、探针与资源配额"), check("工作负载", "控制器数量", "success", fmt.Sprintf("Deployment %d / DaemonSet %d / StatefulSet %d", len(s.deployments.Items), len(s.daemonSets.Items), len(s.statefulSets.Items)), "控制器状态正常", "", ""))
	componentTotal, componentUnhealthy := 0, 0
	for _, pod := range s.pods.Items {
		if pod.Namespace != "kube-system" {
			continue
		}
		componentTotal++
		if pod.Status.Phase != corev1.PodRunning {
			componentUnhealthy++
		}
	}
	componentStatus := "success"
	if componentUnhealthy > 0 {
		componentStatus = "error"
	}
	add("components", "系统组件", check("组件", "kube-system Pod", componentStatus, fmt.Sprintf("健康 %d / 总数 %d", componentTotal-componentUnhealthy, componentTotal), "全部 Running", fmt.Sprintf("异常组件 Pod：%d", componentUnhealthy), "检查控制面与集群插件 Pod 日志和事件"))
	endpointNames := map[string]bool{}
	for _, endpoint := range s.endpoints.Items {
		for _, subset := range endpoint.Subsets {
			if len(subset.Addresses) > 0 {
				endpointNames[endpoint.Namespace+"/"+endpoint.Name] = true
			}
		}
	}
	missingEndpoints := 0
	for _, service := range s.services.Items {
		if service.Spec.Type != corev1.ServiceTypeExternalName && !endpointNames[service.Namespace+"/"+service.Name] {
			missingEndpoints++
		}
	}
	networkStatus := "success"
	if missingEndpoints > 0 {
		networkStatus = "warning"
	}
	add("network", "网络", check("网络", "Service 端点", networkStatus, fmt.Sprintf("%d 个 Service", len(s.services.Items)), "业务 Service 有可用端点", fmt.Sprintf("无端点 Service：%d", missingEndpoints), "检查 Service Selector 与后端 Pod 就绪状态"))
	pendingPVC, failedPV := 0, 0
	for _, pvc := range s.pvcs.Items {
		if pvc.Status.Phase == corev1.ClaimPending {
			pendingPVC++
		}
	}
	for _, pv := range s.pvs.Items {
		if pv.Status.Phase == corev1.VolumeFailed {
			failedPV++
		}
	}
	storageStatus := "success"
	if failedPV > 0 {
		storageStatus = "error"
	} else if pendingPVC > 0 {
		storageStatus = "warning"
	}
	add("storage", "存储", check("存储", "卷绑定状态", storageStatus, fmt.Sprintf("PV %d / PVC %d", len(s.pvs.Items), len(s.pvcs.Items)), "无 Pending PVC 或 Failed PV", fmt.Sprintf("Pending PVC %d，Failed PV %d", pendingPVC, failedPV), "检查 StorageClass、Provisioner 与卷容量"))
	configStatus := "success"
	if noLimits > 0 {
		configStatus = "warning"
	}
	add("config", "配置", check("配置", "容器资源约束", configStatus, fmt.Sprintf("%d 个 Pod 未完整设置", noLimits), "全部容器设置 requests/limits", "缺少资源约束可能影响调度和稳定性", "为容器配置 CPU、内存 requests 与 limits"))
	privileged, hostNetwork := 0, 0
	var requestedCPU, requestedMemory, allocatableCPU, allocatableMemory int64
	podCapacity := int64(0)
	for _, node := range s.nodes.Items {
		allocatableCPU += node.Status.Allocatable.Cpu().MilliValue()
		allocatableMemory += node.Status.Allocatable.Memory().Value()
		podCapacity += node.Status.Allocatable.Pods().Value()
	}
	for _, pod := range s.pods.Items {
		if pod.Spec.HostNetwork {
			hostNetwork++
		}
		for _, container := range pod.Spec.Containers {
			requestedCPU += container.Resources.Requests.Cpu().MilliValue()
			requestedMemory += container.Resources.Requests.Memory().Value()
			if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				privileged++
			}
		}
	}
	securityStatus := "success"
	if privileged > 0 {
		securityStatus = "error"
	} else if hostNetwork > 0 {
		securityStatus = "warning"
	}
	add("security", "安全", check("安全", "高权限工作负载", securityStatus, fmt.Sprintf("特权容器 %d / HostNetwork Pod %d", privileged, hostNetwork), "无特权容器，按需使用 HostNetwork", "高权限工作负载会扩大攻击面", "使用最小权限、安全上下文和网络策略"))
	cpuPercent, memoryPercent, podPercent := float64(0), float64(0), float64(0)
	if allocatableCPU > 0 {
		cpuPercent = float64(requestedCPU) * 100 / float64(allocatableCPU)
	}
	if allocatableMemory > 0 {
		memoryPercent = float64(requestedMemory) * 100 / float64(allocatableMemory)
	}
	if podCapacity > 0 {
		podPercent = float64(len(s.pods.Items)) * 100 / float64(podCapacity)
	}
	capacityStatus := "success"
	if cpuPercent > 90 || memoryPercent > 90 || podPercent > 90 {
		capacityStatus = "error"
	} else if cpuPercent > 75 || memoryPercent > 75 || podPercent > 75 {
		capacityStatus = "warning"
	}
	add("capacity", "容量", check("容量", "资源分配率", capacityStatus, fmt.Sprintf("CPU %.1f%% / 内存 %.1f%% / Pod %.1f%%", cpuPercent, memoryPercent, podPercent), "低于 75%", "根据容器 Requests 与节点 Allocatable 计算", "扩容节点或优化工作负载资源请求"))
	warnings := 0
	recent := time.Now().Add(-24 * time.Hour)
	for _, event := range s.events.Items {
		if event.Type == corev1.EventTypeWarning && event.LastTimestamp.Time.After(recent) {
			warnings++
		}
	}
	eventStatus := "success"
	if warnings > 10 {
		eventStatus = "warning"
	}
	add("events", "集群事件", check("事件", "近 24 小时警告事件", eventStatus, fmt.Sprintf("%d", warnings), "不超过 10", "高频警告通常意味着资源或控制器异常", "检查事件详情并处理重复告警原因"))
	metric := func(label, value, tone string) *api.KubernetesInspectionMetric { return &api.KubernetesInspectionMetric{Label: label, Value: value, Tone: tone} }
	sectionByKey := map[string]*api.KubernetesInspectionSection{}
	for _, section := range sections { sectionByKey[section.Key] = section }
	sectionByKey["cluster"].Metrics = []*api.KubernetesInspectionMetric{metric("Kubernetes 版本", s.version, ""), metric("连接状态", "正常", "success")}
	nodeDetails := make([]map[string]any, 0, len(s.nodes.Items))
	for _, node := range s.nodes.Items { status := "NotReady"; for _, condition := range node.Status.Conditions { if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue { status = "Ready" } }; nodeDetails = append(nodeDetails, map[string]any{"name":node.Name,"status":status,"cpu":node.Status.Capacity.Cpu().String(),"memory":node.Status.Capacity.Memory().String(),"pods":node.Status.Capacity.Pods().String()}) }
	sectionByKey["nodes"].Metrics = []*api.KubernetesInspectionMetric{metric("节点总数",fmt.Sprint(len(s.nodes.Items)),""),metric("Ready 节点",fmt.Sprint(ready),"success"),metric("NotReady 节点",fmt.Sprint(len(s.nodes.Items)-ready),nodeStatus),metric("有资源压力",fmt.Sprint(pressure),nodeStatus)}
	sectionByKey["nodes"].Details = nodeDetails
	sectionByKey["workloads"].Metrics = []*api.KubernetesInspectionMetric{metric("Deployment",fmt.Sprint(len(s.deployments.Items)),""),metric("DaemonSet",fmt.Sprint(len(s.daemonSets.Items)),""),metric("StatefulSet",fmt.Sprint(len(s.statefulSets.Items)),""),metric("Pod总数",fmt.Sprintf("%d/%d",running,len(s.pods.Items)),workloadStatus)}
	sectionByKey["network"].Metrics = []*api.KubernetesInspectionMetric{metric("Service总数",fmt.Sprint(len(s.services.Items)),""),metric("Endpoint总数",fmt.Sprint(len(s.endpoints.Items)),""),metric("无Endpoint的Service",fmt.Sprint(missingEndpoints),networkStatus)}
	sectionByKey["storage"].Metrics = []*api.KubernetesInspectionMetric{metric("PV总数",fmt.Sprint(len(s.pvs.Items)),""),metric("PVC总数",fmt.Sprint(len(s.pvcs.Items)),""),metric("Pending PVC",fmt.Sprint(pendingPVC),storageStatus)}
	sectionByKey["security"].Metrics = []*api.KubernetesInspectionMetric{metric("Pod总数",fmt.Sprint(len(s.pods.Items)),""),metric("特权容器",fmt.Sprint(privileged),securityStatus),metric("hostNetwork Pod",fmt.Sprint(hostNetwork),securityStatus)}
	sectionByKey["capacity"].Metrics = []*api.KubernetesInspectionMetric{metric("CPU分配率",fmt.Sprintf("%.1f%%",cpuPercent),capacityStatus),metric("内存分配率",fmt.Sprintf("%.1f%%",memoryPercent),capacityStatus),metric("Pod密度",fmt.Sprintf("%.1f%%",podPercent),capacityStatus),metric("Pod数量",fmt.Sprintf("%d/%d",len(s.pods.Items),podCapacity),"")}
	eventDetails := make([]map[string]any, 0)
	for _, event := range s.events.Items { if event.Type == corev1.EventTypeWarning && event.LastTimestamp.Time.After(recent) { eventDetails = append(eventDetails,map[string]any{"type":event.Type,"reason":event.Reason,"object":event.InvolvedObject.Kind+"/"+event.InvolvedObject.Name,"namespace":event.Namespace,"count":event.Count,"last_seen":event.LastTimestamp.Format("2006-01-02 15:04:05"),"message":event.Message}) } }
	sectionByKey["events"].Metrics = []*api.KubernetesInspectionMetric{metric("Warning事件",fmt.Sprint(warnings),eventStatus)}
	sectionByKey["events"].Details = eventDetails
	summary := &api.KubernetesInspectionSummary{}
	for _, section := range sections {
		for _, item := range section.Checks {
			summary.Total++
			switch item.Status {
			case "success":
				summary.Passed++
			case "warning":
				summary.Warning++
			default:
				summary.Failed++
			}
		}
	}
	score := 100
	if summary.Total > 0 {
		score = (summary.Passed*100 + summary.Warning*60) / summary.Total
	}
	return &api.KubernetesInspectionReport{ClusterID: clusterID, ClusterName: clusterName, Score: score, Summary: summary, Sections: sections}
}

func (b *KubernetesBiz) failInspection(id int64, cause error) {
	now := time.Now()
	data, _ := json.Marshal(map[string]string{"error": cause.Error()})
	_ = b.db.Model(&model.KubernetesClusterInspection{}).Where("id = ?", id).Updates(map[string]any{"status": "failed", "report_data": string(data), "ended_at": now, "updated_at": now}).Error
}

func (b *KubernetesBiz) InspectionProgress(r *api.KubernetesInspectionIDRequest) (*api.KubernetesInspectionProgressResponse, error) {
	var record model.KubernetesClusterInspection
	if err := b.db.Select("id", "status").First(&record, r.ID).Error; err != nil {
		return nil, k8sErr(err)
	}
	progress, step := 45, "正在并发检查集群资源"
	if record.Status == "completed" {
		progress, step = 100, "巡检完成"
	} else if record.Status == "failed" {
		progress, step = 100, "巡检失败"
	}
	return &api.KubernetesInspectionProgressResponse{ID: record.ID, Status: record.Status, Progress: progress, CurrentStep: step}, nil
}

func (b *KubernetesBiz) InspectionResult(r *api.KubernetesInspectionIDRequest) (*api.KubernetesInspectionResultResponse, error) {
	var record model.KubernetesClusterInspection
	if err := b.db.First(&record, r.ID).Error; err != nil {
		return nil, k8sErr(err)
	}
	var report api.KubernetesInspectionReport
	if record.ReportData != "" && record.Status == "completed" {
		if err := json.Unmarshal([]byte(record.ReportData), &report); err != nil {
			return nil, err
		}
	}
	return &api.KubernetesInspectionResultResponse{Inspection: inspectionHistory(record), Report: &report}, nil
}

func (b *KubernetesBiz) InspectionHistory(r *api.KubernetesInspectionHistoryRequest) ([]*api.KubernetesInspectionHistoryResponse, error) {
	var records []model.KubernetesClusterInspection
	query := b.db.Order("id DESC")
	if r.ClusterID > 0 {
		query = query.Where("cluster_id = ?", r.ClusterID)
	}
	if err := query.Limit(500).Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]*api.KubernetesInspectionHistoryResponse, 0, len(records))
	for _, record := range records {
		result = append(result, inspectionHistory(record))
	}
	return result, nil
}

func (b *KubernetesBiz) DeleteInspection(r *api.KubernetesInspectionIDRequest) error {
	result := b.db.Delete(&model.KubernetesClusterInspection{}, r.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func inspectionHistory(record model.KubernetesClusterInspection) *api.KubernetesInspectionHistoryResponse {
	return &api.KubernetesInspectionHistoryResponse{ID: record.ID, ClusterID: record.ClusterID, ClusterName: record.ClusterName, Status: record.Status, Score: record.Score, CheckCount: record.CheckCount, PassCount: record.PassCount, WarningCount: record.WarningCount, FailCount: record.FailCount, Duration: record.Duration, StartedAt: record.StartedAt, EndedAt: record.EndedAt}
}
