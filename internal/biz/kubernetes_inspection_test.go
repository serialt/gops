package biz

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildInspectionReportTreatsCompletedPodsAsHealthy(t *testing.T) {
	snapshot := &inspectionSnapshot{
		nodes: &corev1.NodeList{},
		pods: &corev1.PodList{Items: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "running", Namespace: "default"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
			{ObjectMeta: metav1.ObjectMeta{Name: "inspection-job", Namespace: "kube-system"}, Status: corev1.PodStatus{Phase: corev1.PodSucceeded}},
		}},
		deployments:  &appsv1.DeploymentList{},
		daemonSets:   &appsv1.DaemonSetList{},
		statefulSets: &appsv1.StatefulSetList{},
		services:     &corev1.ServiceList{},
		endpoints:    &corev1.EndpointsList{},
		pvs:          &corev1.PersistentVolumeList{},
		pvcs:         &corev1.PersistentVolumeClaimList{},
		events:       &corev1.EventList{},
	}

	report := buildInspectionReport(1, "test", snapshot)
	statuses := make(map[string]string)
	values := make(map[string]string)
	for _, section := range report.Sections {
		for _, item := range section.Checks {
			statuses[item.Name] = item.Status
			values[item.Name] = item.Value
		}
	}

	if statuses["Pod 运行状态"] != "success" {
		t.Fatalf("completed pod should keep workload status healthy, got %q", statuses["Pod 运行状态"])
	}
	if values["Pod 运行状态"] != "健康 2 / 总数 2" {
		t.Fatalf("completed pod should be counted as healthy, got %q", values["Pod 运行状态"])
	}
	if statuses["kube-system Pod"] != "success" {
		t.Fatalf("completed kube-system pod should be healthy, got %q", statuses["kube-system Pod"])
	}
}
