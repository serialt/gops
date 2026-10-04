package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func (b *KubernetesBiz) StorageResourceSupport(req *api.KubernetesStorageDetailRequest) (*api.KubernetesStorageSupportResponse, error) {
	if req.Resource != "persistentvolumes" && req.Resource != "persistentvolumeclaims" && req.Resource != "storageclasses" {
		return nil, fmt.Errorf("unsupported storage resource %q", req.Resource)
	}
	client, err := b.client(req.ClusterID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	kind := map[string]string{"persistentvolumes": "PersistentVolume", "persistentvolumeclaims": "PersistentVolumeClaim", "storageclasses": "StorageClass"}[req.Resource]
	result := &api.KubernetesStorageSupportResponse{
		Kind: kind, Name: req.Name, Namespace: req.Namespace,
		AccessModes: make([]string, 0), MountOptions: make([]string, 0), Parameters: make(map[string]string),
		Related: make([]*api.KubernetesRelatedResourceResponse, 0), Events: make([]*api.EventInfo, 0),
	}
	eventNamespace := metav1.NamespaceAll
	switch req.Resource {
	case "persistentvolumeclaims":
		claim, getErr := client.CoreV1().PersistentVolumeClaims(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
		if getErr != nil {
			return nil, getErr
		}
		result.Namespace, result.CreatedAt = claim.Namespace, claim.CreationTimestamp.Format("2006-01-02 15:04:05")
		result.Age, result.UID, result.ResourceVersion = utils.CalculateAge(claim.CreationTimestamp.Time), string(claim.UID), claim.ResourceVersion
		result.Status, result.VolumeName, eventNamespace = string(claim.Status.Phase), claim.Spec.VolumeName, claim.Namespace
		if claim.Spec.StorageClassName != nil {
			result.StorageClassName = *claim.Spec.StorageClassName
		}
		if quantity, ok := claim.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			result.RequestedCapacity = quantity.String()
		}
		if quantity, ok := claim.Status.Capacity[corev1.ResourceStorage]; ok {
			result.Capacity = quantity.String()
		}
		for _, mode := range claim.Status.AccessModes {
			result.AccessModes = append(result.AccessModes, string(mode))
		}
		if len(result.AccessModes) == 0 {
			for _, mode := range claim.Spec.AccessModes {
				result.AccessModes = append(result.AccessModes, string(mode))
			}
		}
		if claim.Spec.VolumeMode != nil {
			result.VolumeMode = string(*claim.Spec.VolumeMode)
		} else {
			result.VolumeMode = string(corev1.PersistentVolumeFilesystem)
		}
		result.Labels, result.Annotations = claim.Labels, claim.Annotations
		if result.StorageClassName != "" {
			result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "StorageClass", Name: result.StorageClassName, Status: "使用中"})
		}
		if result.VolumeName != "" {
			result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "PersistentVolume", Name: result.VolumeName, Status: "已绑定"})
		}
	case "persistentvolumes":
		volume, getErr := client.CoreV1().PersistentVolumes().Get(ctx, req.Name, metav1.GetOptions{})
		if getErr != nil {
			return nil, getErr
		}
		result.Namespace = ""
		result.CreatedAt = volume.CreationTimestamp.Format("2006-01-02 15:04:05")
		result.Age, result.UID, result.ResourceVersion = utils.CalculateAge(volume.CreationTimestamp.Time), string(volume.UID), volume.ResourceVersion
		result.Status, result.StorageClassName, result.ReclaimPolicy = string(volume.Status.Phase), volume.Spec.StorageClassName, string(volume.Spec.PersistentVolumeReclaimPolicy)
		if quantity, ok := volume.Spec.Capacity[corev1.ResourceStorage]; ok {
			result.Capacity = quantity.String()
		}
		for _, mode := range volume.Spec.AccessModes {
			result.AccessModes = append(result.AccessModes, string(mode))
		}
		if volume.Spec.VolumeMode != nil {
			result.VolumeMode = string(*volume.Spec.VolumeMode)
		} else {
			result.VolumeMode = string(corev1.PersistentVolumeFilesystem)
		}
		result.MountOptions = append(result.MountOptions, volume.Spec.MountOptions...)
		result.Labels, result.Annotations = volume.Labels, volume.Annotations
		if claim := volume.Spec.ClaimRef; claim != nil {
			result.ClaimRef = &api.KubernetesStorageClaimReference{Namespace: claim.Namespace, Name: claim.Name}
			result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "PersistentVolumeClaim", Name: claim.Namespace + "/" + claim.Name, Status: "已绑定"})
		}
		if result.StorageClassName != "" {
			result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "StorageClass", Name: result.StorageClassName, Status: "配置"})
		}
	case "storageclasses":
		storageClass, getErr := client.StorageV1().StorageClasses().Get(ctx, req.Name, metav1.GetOptions{})
		if getErr != nil {
			return nil, getErr
		}
		result.Namespace = ""
		result.CreatedAt = storageClass.CreationTimestamp.Format("2006-01-02 15:04:05")
		result.Age, result.UID, result.ResourceVersion = utils.CalculateAge(storageClass.CreationTimestamp.Time), string(storageClass.UID), storageClass.ResourceVersion
		result.Provisioner, result.Parameters = storageClass.Provisioner, storageClass.Parameters
		result.MountOptions = append(result.MountOptions, storageClass.MountOptions...)
		result.ReclaimPolicy = string(corev1.PersistentVolumeReclaimDelete)
		if storageClass.ReclaimPolicy != nil {
			result.ReclaimPolicy = string(*storageClass.ReclaimPolicy)
		}
		result.VolumeBindingMode = "Immediate"
		if storageClass.VolumeBindingMode != nil {
			result.VolumeBindingMode = string(*storageClass.VolumeBindingMode)
		}
		result.AllowVolumeExpansion = storageClass.AllowVolumeExpansion
		result.Labels, result.Annotations = storageClass.Labels, storageClass.Annotations
		if claims, listErr := client.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(ctx, metav1.ListOptions{}); listErr == nil {
			for _, claim := range claims.Items {
				if claim.Spec.StorageClassName != nil && *claim.Spec.StorageClassName == storageClass.Name {
					result.Related = append(result.Related, &api.KubernetesRelatedResourceResponse{Kind: "PersistentVolumeClaim", Name: claim.Namespace + "/" + claim.Name, Status: string(claim.Status.Phase)})
				}
			}
		}
	}

	selector := fields.AndSelectors(fields.OneTermEqualSelector("involvedObject.kind", kind), fields.OneTermEqualSelector("involvedObject.name", result.Name)).String()
	if events, listErr := client.CoreV1().Events(eventNamespace).List(ctx, metav1.ListOptions{FieldSelector: selector}); listErr == nil {
		for _, event := range events.Items {
			result.Events = append(result.Events, batchEventResponse(event))
		}
	}
	result.Description = formatStorageDescription(result)
	return result, nil
}

func formatStorageDescription(result *api.KubernetesStorageSupportResponse) string {
	var description strings.Builder
	fmt.Fprintf(&description, "Name:          %s\nKind:          %s\n", result.Name, result.Kind)
	if result.Namespace != "" {
		fmt.Fprintf(&description, "Namespace:     %s\n", result.Namespace)
	}
	fmt.Fprintf(&description, "Status:        %s\nCapacity:      %s\nRequested:     %s\nStorageClass:  %s\nVolume:        %s\nProvisioner:   %s\nReclaimPolicy: %s\nVolumeMode:    %s\nAccessModes:   %s\n",
		result.Status, result.Capacity, result.RequestedCapacity, result.StorageClassName, result.VolumeName, result.Provisioner, result.ReclaimPolicy, result.VolumeMode, strings.Join(result.AccessModes, ", "))
	if result.ClaimRef != nil {
		fmt.Fprintf(&description, "Claim:         %s/%s\n", result.ClaimRef.Namespace, result.ClaimRef.Name)
	}
	if len(result.Parameters) > 0 {
		fmt.Fprintf(&description, "Parameters:    %s\n", formatKubernetesMetadata(result.Parameters))
	}
	if len(result.Events) == 0 {
		description.WriteString("Events: <none>\n")
	} else {
		description.WriteString("Events:\n")
		for _, event := range result.Events {
			fmt.Fprintf(&description, "  %-8s %-20s %d  %s\n", event.Type, event.Reason, event.Count, event.Message)
		}
	}
	return description.String()
}
