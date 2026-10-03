package handler

import (
	"encoding/json"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/middleware"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct{ biz *biz.TaskBiz }

func NewTaskHandler(b *biz.TaskBiz) *TaskHandler { return &TaskHandler{biz: b} }

func (h *TaskHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/tasks/jobs/list", h.ListJobs)
	g.POST("/tasks/execute", h.Execute)
	g.POST("/tasks/distribute", h.Distribute)
	g.POST("/tasks/templates/list", h.ListTemplates)
	g.POST("/tasks/templates/create", h.CreateTemplate)
	g.POST("/tasks/templates/update", h.UpdateTemplate)
	g.POST("/tasks/templates/delete", h.DeleteTemplate)
}

// ListJobs POST /api/v1/tasks/jobs/list 查询任务作业。
func (h *TaskHandler) ListJobs(c *gin.Context) {
	req := new(api.TaskJobListRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "任务查询参数错误")
		return
	}
	data, err := h.biz.ListJobs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// Execute POST /api/v1/tasks/execute 在目标主机执行脚本任务。
func (h *TaskHandler) Execute(c *gin.Context) {
	req := new(api.TaskExecuteRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "任务执行参数错误")
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	data, err := h.biz.Execute(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// Distribute POST /api/v1/tasks/distribute 向目标主机分发文件。
func (h *TaskHandler) Distribute(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(128 << 20); err != nil {
		response.Error(c, 400, "无法解析上传内容")
		return
	}
	hostIDs := make([]int64, 0)
	if err := json.Unmarshal([]byte(c.PostForm("host_ids")), &hostIDs); err != nil {
		response.Error(c, 400, "目标主机参数错误")
		return
	}
	form := c.Request.MultipartForm
	if form == nil {
		response.Error(c, 400, "请上传文件")
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	data, err := h.biz.Distribute(c.Request.Context(), userID, hostIDs, c.PostForm("target_path"), form.File["files"])
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ListTemplates POST /api/v1/tasks/templates/list 查询任务模板。
func (h *TaskHandler) ListTemplates(c *gin.Context) {
	req := new(api.TaskTemplateListRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "模板查询参数错误")
		return
	}
	data, err := h.biz.ListTemplates(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// CreateTemplate POST /api/v1/tasks/templates/create 新建任务模板。
func (h *TaskHandler) CreateTemplate(c *gin.Context) {
	req := new(api.TaskTemplateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "模板参数错误")
		return
	}
	userID, _ := middleware.CurrentUserID(c)
	data, err := h.biz.CreateTemplate(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// UpdateTemplate POST /api/v1/tasks/templates/update 更新任务模板。
func (h *TaskHandler) UpdateTemplate(c *gin.Context) {
	req := new(api.TaskTemplateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "模板参数错误")
		return
	}
	data, err := h.biz.UpdateTemplate(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// DeleteTemplate POST /api/v1/tasks/templates/delete 删除任务模板。
func (h *TaskHandler) DeleteTemplate(c *gin.Context) {
	req := new(api.TaskTemplateIDRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "模板 ID 错误")
		return
	}
	if err := h.biz.DeleteTemplate(c.Request.Context(), req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}
