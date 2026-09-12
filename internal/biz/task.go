package biz

import (
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gorm.io/gorm"
)

type TaskBiz struct{ db *gorm.DB }

func NewTaskBiz(db *gorm.DB) *TaskBiz { return &TaskBiz{db} }
func (b *TaskBiz) ListJobs(r *api.TaskListRequest) ([]*api.TaskJobResponse, error) {
	var a []model.TaskJob
	q := b.db.Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+r.Keyword+"%")
	}
	if e := q.Find(&a).Error; e != nil {
		return nil, e
	}
	o := make([]*api.TaskJobResponse, 0, len(a))
	for _, x := range a {
		o = append(o, taskJob(x))
	}
	return o, nil
}
func (b *TaskBiz) SaveJob(r *api.TaskJobRequest) (*api.TaskJobResponse, error) {
	x := model.TaskJob{ID: r.ID, Name: r.Name, TemplateID: r.TemplateID, TaskType: r.TaskType, TargetHosts: r.TargetHosts, Parameters: r.Parameters, Status: "pending"}
	if x.TaskType == "" {
		x.TaskType = "manual"
	}
	if r.ID > 0 {
		var old model.TaskJob
		if e := b.db.First(&old, r.ID).Error; e != nil {
			return nil, e
		}
		x.CreatedAt = old.CreatedAt
		x.Status = old.Status
	}
	if e := b.db.Save(&x).Error; e != nil {
		return nil, e
	}
	return taskJob(x), nil
}
func taskJob(x model.TaskJob) *api.TaskJobResponse {
	return &api.TaskJobResponse{ID: x.ID, Name: x.Name, TemplateID: x.TemplateID, TaskType: x.TaskType, Status: x.Status, TargetHosts: x.TargetHosts, Parameters: x.Parameters, CreatedAt: x.CreatedAt}
}
