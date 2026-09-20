package biz

import (
	"errors"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"sync"

	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
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
	}
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
func k8sCluster(x model.KubernetesCluster) *api.KubernetesClusterResponse {
	return &api.KubernetesClusterResponse{ID: x.ID, Name: x.Name, Alias: x.Alias_, APIEndpoint: x.APIEndpoint, Version: x.Version, Status: x.Status, NodeCount: x.NodeCount, PodCount: x.PodCount, Region: x.Region, Provider: x.Provider, Description: x.Description, CreatedAt: x.CreatedAt, UpdatedAt: x.UpdatedAt}
}
func k8sErr(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return e
}
