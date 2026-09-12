package handler

import (
	"github.com/gin-gonic/gin"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
)

type TaskHandler struct{ biz *biz.TaskBiz }

func NewTaskHandler(b *biz.TaskBiz) *TaskHandler { return &TaskHandler{b} }

// ListJobs POST /api/v1/tasks/jobs/list 查询任务作业。
func (h *TaskHandler) ListJobs(c *gin.Context) {
	r := new(api.TaskListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid task query")
		return
	}
	d, e := h.biz.ListJobs(r)
	if e != nil {
		response.Error(c, 500, e.Error())
		return
	}
	response.OK(c, d)
}

// SaveJob POST /api/v1/tasks/jobs/save 保存任务作业。
func (h *TaskHandler) SaveJob(c *gin.Context) {
	r := new(api.TaskJobRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid task data")
		return
	}
	d, e := h.biz.SaveJob(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}
