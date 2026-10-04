package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (b *KubernetesBiz) JobSupport(req *api.KubernetesJobDetailRequest) (*api.KubernetesStatefulSetSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	job, err := client.BatchV1().Jobs(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	podList, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	eventList, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	desired := int32(1)
	if job.Spec.Completions != nil {
		desired = *job.Spec.Completions
	}
	parallelism := int32(1)
	if job.Spec.Parallelism != nil {
		parallelism = *job.Spec.Parallelism
	}
	result := &api.KubernetesStatefulSetSupportResponse{
		DesiredReplicas: desired,
		CurrentReplicas: job.Status.Active,
		ReadyReplicas:   job.Status.Succeeded,
		UpdatedReplicas: job.Status.Failed,
		ActiveReplicas:  job.Status.Active,
		Succeeded:       job.Status.Succeeded,
		Failed:          job.Status.Failed,
		Completions:     desired,
		Parallelism:     parallelism,
		Pods:            make([]*api.KubernetesObjectResponse, 0),
		Claims:          make([]*api.KubernetesPersistentVolumeClaimResponse, 0),
		Events:          make([]*api.EventInfo, 0),
	}
	result.Status = jobStatusText(job)
	ownedPods := make(map[string]struct{})
	for _, pod := range podList.Items {
		if !hasControllerOwner(pod.OwnerReferences, job.UID, "Job") {
			continue
		}
		ownedPods[pod.Name] = struct{}{}
		result.Pods = append(result.Pods, batchPodResponse(pod))
	}
	for _, event := range eventList.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "job" || event.InvolvedObject.Name != job.Name) && (kind != "pod" || !containsPod(ownedPods, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, batchEventResponse(event))
	}
	var output strings.Builder
	fmt.Fprintf(&output, "Name:                  %s\nNamespace:             %s\n", job.Name, job.Namespace)
	fmt.Fprintf(&output, "Status:                %s\nCompletions:           %d/%d\nParallelism:           %d\nActive:                %d\nSucceeded:             %d\nFailed:                %d\nPods:                  %d\n", result.Status, result.Succeeded, result.Completions, result.Parallelism, result.ActiveReplicas, result.Succeeded, result.Failed, len(result.Pods))
	result.Description = output.String()
	return result, nil
}

func (b *KubernetesBiz) CronJobSupport(req *api.KubernetesCronJobDetailRequest) (*api.KubernetesStatefulSetSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cronJob, err := client.BatchV1().CronJobs(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobList, err := client.BatchV1().Jobs(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	podList, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	eventList, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ownedJobs := make(map[string]struct{})
	jobUIDs := make(map[types.UID]struct{})
	var succeeded, failed int32
	for _, job := range jobList.Items {
		if !hasControllerOwner(job.OwnerReferences, cronJob.UID, "CronJob") {
			continue
		}
		ownedJobs[job.Name] = struct{}{}
		jobUIDs[job.UID] = struct{}{}
		succeeded += job.Status.Succeeded
		failed += job.Status.Failed
	}
	active := int32(len(cronJob.Status.Active))
	lastSchedule := ""
	if cronJob.Status.LastScheduleTime != nil {
		lastSchedule = cronJob.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
	}
	suspended := cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend
	result := &api.KubernetesStatefulSetSupportResponse{
		Status:          cronJobStatusText(suspended, active, cronJob.Status.LastScheduleTime),
		DesiredReplicas: active,
		CurrentReplicas: active,
		ReadyReplicas:   succeeded,
		UpdatedReplicas: failed,
		ActiveReplicas:  active,
		Succeeded:       succeeded,
		Failed:          failed,
		Schedule:        cronJob.Spec.Schedule,
		LastSchedule:    lastSchedule,
		Concurrency:     string(cronJob.Spec.ConcurrencyPolicy),
		Suspended:       suspended,
		Pods:            make([]*api.KubernetesObjectResponse, 0),
		Claims:          make([]*api.KubernetesPersistentVolumeClaimResponse, 0),
		Events:          make([]*api.EventInfo, 0),
	}
	ownedPods := make(map[string]struct{})
	for _, pod := range podList.Items {
		if !ownedByAnyJob(pod.OwnerReferences, jobUIDs) {
			continue
		}
		ownedPods[pod.Name] = struct{}{}
		result.Pods = append(result.Pods, batchPodResponse(pod))
	}
	for _, event := range eventList.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		related := kind == "cronjob" && event.InvolvedObject.Name == cronJob.Name
		if kind == "job" {
			_, related = ownedJobs[event.InvolvedObject.Name]
		}
		if kind == "pod" {
			_, related = ownedPods[event.InvolvedObject.Name]
		}
		if related {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}
	var output strings.Builder
	fmt.Fprintf(&output, "Name:                  %s\nNamespace:             %s\nSchedule:              %s\nConcurrency Policy:    %s\nSuspended:             %t\nActive:                %d\nLast Schedule:         %s\nSucceeded:             %d\nFailed:                %d\nPods:                  %d\n", cronJob.Name, cronJob.Namespace, cronJob.Spec.Schedule, cronJob.Spec.ConcurrencyPolicy, suspended, active, lastSchedule, succeeded, failed, len(result.Pods))
	result.Description = output.String()
	return result, nil
}

func jobStatusText(job *batchv1.Job) string {
	for _, condition := range job.Status.Conditions {
		if condition.Status != corev1.ConditionTrue {
			continue
		}
		if condition.Type == batchv1.JobComplete {
			return "Completed"
		}
		if condition.Type == batchv1.JobFailed {
			return "Failed"
		}
	}
	if job.Status.Active > 0 {
		return "Running"
	}
	return "Pending"
}

func cronJobStatusText(suspended bool, active int32, lastSchedule *metav1.Time) string {
	if suspended {
		return "Suspended"
	}
	if active > 0 {
		return "Running"
	}
	if lastSchedule == nil {
		return "Pending"
	}
	return "Scheduled"
}

func hasControllerOwner(owners []metav1.OwnerReference, uid types.UID, kind string) bool {
	for _, owner := range owners {
		if owner.Controller != nil && *owner.Controller {
			return owner.UID == uid && owner.Kind == kind
		}
	}
	return false
}

func ownedByAnyJob(owners []metav1.OwnerReference, jobUIDs map[types.UID]struct{}) bool {
	for _, owner := range owners {
		if owner.Controller == nil || !*owner.Controller || owner.Kind != "Job" {
			continue
		}
		if _, found := jobUIDs[owner.UID]; found {
			return true
		}
	}
	return false
}

func batchPodResponse(pod corev1.Pod) *api.KubernetesObjectResponse {
	readyContainers := 0
	restarts := int64(0)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			readyContainers++
		}
		restarts += int64(status.RestartCount)
	}
	images := make([]string, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		images = append(images, container.Image)
	}
	return &api.KubernetesObjectResponse{
		Name: pod.Name, Namespace: pod.Namespace, Kind: "Pod", Status: string(pod.Status.Phase),
		Age: utils.CalculateAge(pod.CreationTimestamp.Time), Labels: pod.Labels,
		CreatedAt: pod.CreationTimestamp.Format("2006-01-02 15:04:05"), Ready: fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers)),
		Images: images, Restarts: restarts, PodIP: pod.Status.PodIP, Node: pod.Spec.NodeName,
	}
}

func batchEventResponse(event corev1.Event) *api.EventInfo {
	return &api.EventInfo{
		Type: event.Type, Reason: event.Reason, Message: event.Message, Source: event.Source.Component,
		Count: event.Count, FirstTimestamp: event.FirstTimestamp.Format("2006-01-02 15:04:05"),
		LastTimestamp:  event.LastTimestamp.Format("2006-01-02 15:04:05"),
		InvolvedObject: api.InvolvedObjectInfo{Kind: event.InvolvedObject.Kind, Name: event.InvolvedObject.Name, Namespace: event.InvolvedObject.Namespace},
	}
}
