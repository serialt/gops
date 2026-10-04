package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

func (b *KubernetesBiz) StatefulSetSupport(req *api.KubernetesStatefulSetDetailRequest) (*api.KubernetesStatefulSetSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	statefulSet, err := client.AppsV1().StatefulSets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		return nil, err
	}
	podList, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	eventList, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	claimList, err := client.CoreV1().PersistentVolumeClaims(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := &api.KubernetesStatefulSetSupportResponse{
		DesiredReplicas: valueOrZero(statefulSet.Spec.Replicas),
		CurrentReplicas: statefulSet.Status.Replicas,
		ReadyReplicas:   statefulSet.Status.ReadyReplicas,
		UpdatedReplicas: statefulSet.Status.UpdatedReplicas,
		Pods:            make([]*api.KubernetesObjectResponse, 0, len(podList.Items)),
		Claims:          make([]*api.KubernetesPersistentVolumeClaimResponse, 0),
		Events:          make([]*api.EventInfo, 0),
	}
	if statefulSet.Spec.Replicas == nil {
		result.DesiredReplicas = 1
	}
	switch {
	case result.DesiredReplicas == 0:
		result.Status = "Stopped"
	case statefulSet.Status.ObservedGeneration < statefulSet.Generation:
		result.Status = "Progressing"
	case result.ReadyReplicas >= result.DesiredReplicas && result.UpdatedReplicas >= result.DesiredReplicas:
		result.Status = "Running"
	case result.CurrentReplicas == 0:
		result.Status = "Pending"
	case result.UpdatedReplicas < result.DesiredReplicas:
		result.Status = "Progressing"
	default:
		result.Status = "Degraded"
	}
	ownedPods := make(map[string]struct{}, len(podList.Items))
	for _, pod := range podList.Items {
		if !workloadOwnsPod(statefulSet.UID, "StatefulSet", pod) {
			continue
		}
		ownedPods[pod.Name] = struct{}{}
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
		result.Pods = append(result.Pods, &api.KubernetesObjectResponse{
			Name: pod.Name, Namespace: pod.Namespace, Kind: "Pod", Status: string(pod.Status.Phase),
			Age: utils.CalculateAge(pod.CreationTimestamp.Time), Labels: pod.Labels,
			CreatedAt: pod.CreationTimestamp.Format("2006-01-02 15:04:05"), Ready: fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers)),
			Images: images, Restarts: restarts, PodIP: pod.Status.PodIP, Node: pod.Spec.NodeName,
		})
	}
	claimPrefixes := make([]string, 0, len(statefulSet.Spec.VolumeClaimTemplates))
	for _, template := range statefulSet.Spec.VolumeClaimTemplates {
		claimPrefixes = append(claimPrefixes, template.Name+"-"+statefulSet.Name+"-")
	}
	for _, claim := range claimList.Items {
		if !matchesAnyPrefix(claim.Name, claimPrefixes) {
			continue
		}
		modes := make([]string, 0, len(claim.Spec.AccessModes))
		for _, mode := range claim.Spec.AccessModes {
			modes = append(modes, string(mode))
		}
		capacity := claim.Status.Capacity[corev1.ResourceStorage]
		storageClass := ""
		if claim.Spec.StorageClassName != nil {
			storageClass = *claim.Spec.StorageClassName
		}
		result.Claims = append(result.Claims, &api.KubernetesPersistentVolumeClaimResponse{
			Name: claim.Name, Status: string(claim.Status.Phase), VolumeName: claim.Spec.VolumeName,
			Capacity: capacity.String(), AccessModes: modes, StorageClassName: storageClass,
			CreatedAt: claim.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	for _, event := range eventList.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "statefulset" || event.InvolvedObject.Name != statefulSet.Name) && (kind != "pod" || event.InvolvedObject.Namespace != req.Namespace || !containsPod(ownedPods, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, &api.EventInfo{
			Type: event.Type, Reason: event.Reason, Message: event.Message, Source: event.Source.Component,
			Count: event.Count, FirstTimestamp: event.FirstTimestamp.Format("2006-01-02 15:04:05"),
			LastTimestamp:  event.LastTimestamp.Format("2006-01-02 15:04:05"),
			InvolvedObject: api.InvolvedObjectInfo{Kind: event.InvolvedObject.Kind, Name: event.InvolvedObject.Name, Namespace: event.InvolvedObject.Namespace},
		})
	}
	result.Description = formatStatefulSetDescription(statefulSet, result.Pods, result.Claims)
	return result, nil
}

func (b *KubernetesBiz) DaemonSetSupport(req *api.KubernetesDaemonSetDetailRequest) (*api.KubernetesStatefulSetSupportResponse, error) {
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	daemonSet, err := client.AppsV1().DaemonSets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
	if err != nil {
		return nil, err
	}
	podList, err := client.CoreV1().Pods(req.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	eventList, err := client.CoreV1().Events(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := &api.KubernetesStatefulSetSupportResponse{
		DesiredReplicas: daemonSet.Status.DesiredNumberScheduled,
		CurrentReplicas: daemonSet.Status.CurrentNumberScheduled,
		ReadyReplicas:   daemonSet.Status.NumberReady,
		UpdatedReplicas: daemonSet.Status.UpdatedNumberScheduled,
		Pods:            make([]*api.KubernetesObjectResponse, 0, len(podList.Items)),
		Claims:          make([]*api.KubernetesPersistentVolumeClaimResponse, 0),
		Events:          make([]*api.EventInfo, 0),
	}
	switch {
	case result.DesiredReplicas == 0:
		result.Status = "Pending"
	case daemonSet.Status.ObservedGeneration < daemonSet.Generation:
		result.Status = "Progressing"
	case result.ReadyReplicas >= result.DesiredReplicas && result.UpdatedReplicas >= result.DesiredReplicas:
		result.Status = "Running"
	case result.CurrentReplicas == 0:
		result.Status = "Pending"
	case result.UpdatedReplicas < result.DesiredReplicas:
		result.Status = "Progressing"
	default:
		result.Status = "Degraded"
	}
	ownedPods := make(map[string]struct{}, len(podList.Items))
	for _, pod := range podList.Items {
		if !workloadOwnsPod(daemonSet.UID, "DaemonSet", pod) {
			continue
		}
		ownedPods[pod.Name] = struct{}{}
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
		result.Pods = append(result.Pods, &api.KubernetesObjectResponse{
			Name: pod.Name, Namespace: pod.Namespace, Kind: "Pod", Status: string(pod.Status.Phase),
			Age: utils.CalculateAge(pod.CreationTimestamp.Time), Labels: pod.Labels,
			CreatedAt: pod.CreationTimestamp.Format("2006-01-02 15:04:05"), Ready: fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers)),
			Images: images, Restarts: restarts, PodIP: pod.Status.PodIP, Node: pod.Spec.NodeName,
		})
	}
	for _, event := range eventList.Items {
		kind := strings.ToLower(event.InvolvedObject.Kind)
		if (kind != "daemonset" || event.InvolvedObject.Name != daemonSet.Name) && (kind != "pod" || event.InvolvedObject.Namespace != req.Namespace || !containsPod(ownedPods, event.InvolvedObject.Name)) {
			continue
		}
		result.Events = append(result.Events, &api.EventInfo{
			Type: event.Type, Reason: event.Reason, Message: event.Message, Source: event.Source.Component,
			Count: event.Count, FirstTimestamp: event.FirstTimestamp.Format("2006-01-02 15:04:05"),
			LastTimestamp:  event.LastTimestamp.Format("2006-01-02 15:04:05"),
			InvolvedObject: api.InvolvedObjectInfo{Kind: event.InvolvedObject.Kind, Name: event.InvolvedObject.Name, Namespace: event.InvolvedObject.Namespace},
		})
	}
	var output strings.Builder
	fmt.Fprintf(&output, "Name:                  %s\nNamespace:             %s\n", daemonSet.Name, daemonSet.Namespace)
	fmt.Fprintf(&output, "Desired:               %d\nCurrent:               %d\nReady:                 %d\nUpdated:               %d\n", result.DesiredReplicas, result.CurrentReplicas, result.ReadyReplicas, result.UpdatedReplicas)
	fmt.Fprintf(&output, "Selector:              %s\nUpdate Strategy:       %s\nPods:                  %d\n", selector.String(), daemonSet.Spec.UpdateStrategy.Type, len(result.Pods))
	result.Description = output.String()
	return result, nil
}

func workloadOwnsPod(uid types.UID, kind string, pod corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Controller != nil && *owner.Controller {
			return owner.Kind == kind && owner.UID == uid
		}
	}
	return true
}

func matchesAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func containsPod(pods map[string]struct{}, name string) bool {
	_, found := pods[name]
	return found
}

func formatStatefulSetDescription(statefulSet *appsv1.StatefulSet, pods []*api.KubernetesObjectResponse, claims []*api.KubernetesPersistentVolumeClaimResponse) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Name:                  %s\nNamespace:             %s\n", statefulSet.Name, statefulSet.Namespace)
	fmt.Fprintf(&output, "Service Name:          %s\nPod Management Policy: %s\nUpdate Strategy:       %s\n", statefulSet.Spec.ServiceName, statefulSet.Spec.PodManagementPolicy, statefulSet.Spec.UpdateStrategy.Type)
	fmt.Fprintf(&output, "Replicas:              %d desired | %d current | %d ready | %d updated\n", valueOrZero(statefulSet.Spec.Replicas), statefulSet.Status.Replicas, statefulSet.Status.ReadyReplicas, statefulSet.Status.UpdatedReplicas)
	fmt.Fprintf(&output, "Selector:              %s\nPods:                  %d\nPVCs:                  %d\n", labels.Set(statefulSet.Spec.Selector.MatchLabels).AsSelector().String(), len(pods), len(claims))
	return output.String()
}
