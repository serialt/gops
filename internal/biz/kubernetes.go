package biz

import (
	"context"
	"errors"
	"fmt"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubernetesBiz struct {
	db             *gorm.DB
	clientsetCache map[string]*kubernetes.Clientset
	metricsCache   map[uint]*metricsv.Clientset
	cacheMutex     sync.RWMutex
}

func NewKubernetesBiz(db *gorm.DB) *KubernetesBiz {
	return &KubernetesBiz{
		db: db,
	}
}

func (b *KubernetesBiz) Test(r *api.KubernetesClusterIDRequest) (*api.KubernetesClusterTestResponse, error) {
	client, err := b.client(r.ID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("connect cluster: %w", err)
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	if err = b.db.Model(&model.KubernetesCluster{}).Where("id = ?", r.ID).Updates(map[string]any{"version": version.GitVersion, "status": 1, "node_count": len(nodes.Items), "pod_count": len(pods.Items)}).Error; err != nil {
		return nil, err
	}
	return &api.KubernetesClusterTestResponse{Status: "connected", Version: version.GitVersion}, nil
}

func (b *KubernetesBiz) SyncAll(ctx context.Context) (*api.KubernetesClusterSyncResponse, error) {
	var clusters []model.KubernetesCluster
	if err := b.db.Select("id", "name").Order("id").Find(&clusters).Error; err != nil {
		return nil, err
	}
	result := &api.KubernetesClusterSyncResponse{Total: len(clusters), Errors: make([]string, 0)}
	var wg sync.WaitGroup
	var mu sync.Mutex
	limit := make(chan struct{}, 5)
	for _, cluster := range clusters {
		cluster := cluster
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case limit <- struct{}{}:
				defer func() { <-limit }()
			case <-ctx.Done():
				mu.Lock()
				result.Failed++
				result.Errors = append(result.Errors, cluster.Name+": sync canceled")
				mu.Unlock()
				return
			}
			_, err := b.Test(&api.KubernetesClusterIDRequest{ID: cluster.ID})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, cluster.Name+": "+err.Error())
				_ = b.db.Model(&model.KubernetesCluster{}).Where("id = ?", cluster.ID).Update("status", 0).Error
				return
			}
			result.Success++
		}()
	}
	wg.Wait()
	return result, nil
}

// StartStatusScheduler starts the periodic Kubernetes cluster status synchronization loop.
func (b *KubernetesBiz) StartStatusScheduler(ctx context.Context, interval time.Duration) {
	sync := func() {
		result, err := b.SyncAll(ctx)
		if err != nil {
			slog.Error("sync kubernetes cluster status", "error", err)
			return
		}
		slog.Info("kubernetes cluster status synchronized", "total", result.Total, "success", result.Success, "failed", result.Failed)
	}
	sync()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sync()
		}
	}
}

func (b *KubernetesBiz) List(r *api.KubernetesClusterListRequest) ([]*api.KubernetesClusterResponse, error) {
	var a []model.KubernetesCluster
	q := b.db.Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("name LIKE ? OR alias LIKE ?", "%"+r.Keyword+"%", "%"+r.Keyword+"%")
	}
	if e := q.Find(&a).Error; e != nil {
		return nil, e
	}
	o := make([]*api.KubernetesClusterResponse, 0, len(a))
	for _, x := range a {
		o = append(o, k8sCluster(x))
	}
	return o, nil
}
func (b *KubernetesBiz) Save(r *api.KubernetesClusterRequest) (*api.KubernetesClusterResponse, error) {
	x := model.KubernetesCluster{ID: r.ID, Name: r.Name, Alias_: r.Alias, APIEndpoint: r.APIEndpoint, KubeConfig: r.KubeConfig, Region: r.Region, Provider: r.Provider, Description: r.Description, Status: r.Status}
	if r.ID > 0 {
		var old model.KubernetesCluster
		if e := b.db.First(&old, r.ID).Error; e != nil {
			return nil, k8sErr(e)
		}
		x.CreatedAt = old.CreatedAt
		x.NodeCount = old.NodeCount
		x.PodCount = old.PodCount
		x.Version = old.Version
		if r.KubeConfig == "" {
			x.KubeConfig = old.KubeConfig
		}
	}
	if x.KubeConfig == "" {
		return nil, fmt.Errorf("kubeconfig is required")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(x.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	config.Timeout = 10 * time.Second
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("connect cluster: %w", err)
	}
	x.APIEndpoint = config.Host
	x.Version = version.GitVersion
	x.Status = 1
	if e := b.db.Save(&x).Error; e != nil {
		return nil, e
	}
	return k8sCluster(x), nil
}
func (b *KubernetesBiz) Delete(r *api.KubernetesClusterIDRequest) error {
	x := b.db.Delete(&model.KubernetesCluster{}, r.ID)
	if x.Error != nil {
		return x.Error
	}
	if x.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *KubernetesBiz) Config(r *api.KubernetesClusterIDRequest) (*api.KubernetesClusterConfigResponse, error) {
	var cluster model.KubernetesCluster
	if err := b.db.Select("kube_config").First(&cluster, r.ID).Error; err != nil {
		return nil, k8sErr(err)
	}
	return &api.KubernetesClusterConfigResponse{KubeConfig: cluster.KubeConfig}, nil
}
func k8sCluster(x model.KubernetesCluster) *api.KubernetesClusterResponse {
	return &api.KubernetesClusterResponse{ID: x.ID, Name: x.Name, Alias: x.Alias_, APIEndpoint: x.APIEndpoint, Version: x.Version, Status: x.Status, NodeCount: x.NodeCount, PodCount: x.PodCount, Region: x.Region, Provider: x.Provider, Description: x.Description, CreatedAt: x.CreatedAt, UpdatedAt: x.UpdatedAt}
}
func k8sErr(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return e
}
