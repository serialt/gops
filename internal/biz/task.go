package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/repo"

	"github.com/pkg/sftp"
	"gorm.io/gorm"
)

type TaskBiz struct {
	db     *gorm.DB
	repo   *repo.TaskRepo
	assets *AssetBiz
}

func NewTaskBiz(db *gorm.DB, taskRepo *repo.TaskRepo, assets *AssetBiz) *TaskBiz {
	return &TaskBiz{db: db, repo: taskRepo, assets: assets}
}

func (b *TaskBiz) ListJobs(ctx context.Context, req *api.TaskJobListRequest) (*api.PageResponse[*api.TaskJobResponse], error) {
	rows, total, err := b.repo.ListJobs(ctx, b.db, &define.TaskJobListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, Keyword: req.Keyword, TaskType: req.TaskType, Status: req.Status})
	if err != nil {
		return nil, err
	}
	hostIDs := make([]int64, 0)
	seenHostIDs := make(map[int64]struct{})
	for _, row := range rows {
		var ids []int64
		if json.Unmarshal([]byte(row.TargetHosts), &ids) == nil {
			for _, id := range ids {
				if _, exists := seenHostIDs[id]; !exists {
					seenHostIDs[id] = struct{}{}
					hostIDs = append(hostIDs, id)
				}
			}
		}
	}
	hostNames := make(map[int64]string, len(hostIDs))
	if len(hostIDs) > 0 {
		hosts, _, listErr := b.assets.repo.ListHosts(ctx, b.db, &define.HostListCond{PageCond: define.PageCond{Page: 1, Limit: api.MaxLimit}, IDs: hostIDs})
		if listErr != nil {
			return nil, listErr
		}
		for _, host := range hosts {
			hostNames[host.ID] = fmt.Sprintf("%s（%s）", host.Name, host.IP)
		}
	}
	result := make([]*api.TaskJobResponse, 0, len(rows))
	for _, row := range rows {
		var ids []int64
		_ = json.Unmarshal([]byte(row.TargetHosts), &ids)
		displays := make([]string, 0, len(ids))
		for _, id := range ids {
			if name := hostNames[id]; name != "" {
				displays = append(displays, name)
			} else {
				displays = append(displays, fmt.Sprintf("主机 #%d", id))
			}
		}
		result = append(result, taskJobResponse(row, strings.Join(displays, "、")))
	}
	return &api.PageResponse[*api.TaskJobResponse]{List: result, Total: total}, nil
}

func (b *TaskBiz) ListTemplates(ctx context.Context, req *api.TaskTemplateListRequest) (*api.PageResponse[*api.TaskTemplateResponse], error) {
	rows, total, err := b.repo.ListTemplates(ctx, b.db, &define.TaskTemplateListCond{PageCond: define.PageCond{Page: req.Page, Limit: req.Limit}, Keyword: req.Keyword, Category: req.Category, Platform: req.Platform, Status: req.Status})
	if err != nil {
		return nil, err
	}
	result := make([]*api.TaskTemplateResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, taskTemplateResponse(row))
	}
	return &api.PageResponse[*api.TaskTemplateResponse]{List: result, Total: total}, nil
}

func (b *TaskBiz) CreateTemplate(ctx context.Context, userID uint, req *api.TaskTemplateRequest) (*api.TaskTemplateResponse, error) {
	row := taskTemplateModel(req)
	row.CreatedBy = int64(userID)
	if err := b.repo.CreateTemplate(ctx, b.db, row); err != nil {
		return nil, err
	}
	return taskTemplateResponse(row), nil
}

func (b *TaskBiz) UpdateTemplate(ctx context.Context, req *api.TaskTemplateRequest) (*api.TaskTemplateResponse, error) {
	if req.ID <= 0 {
		return nil, ErrInvalid
	}
	row, err := b.repo.GetTemplate(ctx, b.db, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	updated := taskTemplateModel(req)
	updated.CreatedBy, updated.CreatedAt = row.CreatedBy, row.CreatedAt
	if err := b.repo.SaveTemplate(ctx, b.db, updated); err != nil {
		return nil, err
	}
	return taskTemplateResponse(updated), nil
}

func (b *TaskBiz) DeleteTemplate(ctx context.Context, req *api.TaskTemplateIDRequest) error {
	row, err := b.repo.GetTemplate(ctx, b.db, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return b.repo.DeleteTemplate(ctx, b.db, row)
}

func (b *TaskBiz) Execute(ctx context.Context, userID uint, req *api.TaskExecuteRequest) (*api.TaskExecuteResponse, error) {
	if err := validateTaskContent(req.Content); err != nil {
		return nil, err
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 300
	}
	if timeout > 3600 {
		return nil, fmt.Errorf("执行超时不能超过 3600 秒")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "手动执行任务 - " + time.Now().Format("2006-01-02 15:04:05")
	}
	hostsJSON, _ := json.Marshal(req.HostIDs)
	paramsJSON, _ := json.Marshal(map[string]any{"script_type": req.ScriptType, "timeout": timeout})
	job := &model.TaskJob{Name: name, TemplateID: req.TemplateID, TaskType: "manual", Status: "running", TargetHosts: string(hostsJSON), Parameters: string(paramsJSON), ExecuteTime: time.Now(), CreatedBy: int64(userID)}
	if err := b.repo.CreateJob(ctx, b.db, job); err != nil {
		return nil, err
	}

	results := make([]*api.TaskHostResultResponse, 0, len(req.HostIDs))
	allSuccess := true
	for _, hostID := range req.HostIDs {
		result := b.executeOnHost(ctx, hostID, req.ScriptType, req.Content, time.Duration(timeout)*time.Second)
		results = append(results, result)
		if result.Status != "success" {
			allSuccess = false
		}
	}
	status := "success"
	if !allSuccess {
		status = "failed"
	}
	resultJSON, _ := json.Marshal(results)
	if err := b.repo.UpdateJobResult(ctx, b.db, job.ID, status, string(resultJSON), ""); err != nil {
		return nil, err
	}
	return &api.TaskExecuteResponse{TaskID: job.ID, Results: results}, nil
}

func (b *TaskBiz) Distribute(ctx context.Context, userID uint, hostIDs []int64, targetPath string, files []*multipart.FileHeader) (*api.TaskDistributeResponse, error) {
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("请选择至少一台目标主机")
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("请上传至少一个文件")
	}
	targetPath = path.Clean(strings.TrimSpace(targetPath))
	if !strings.HasPrefix(targetPath, "/") {
		return nil, fmt.Errorf("目标路径必须是绝对路径")
	}
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		fileNames = append(fileNames, path.Base(file.Filename))
	}
	hostsJSON, _ := json.Marshal(hostIDs)
	paramsJSON, _ := json.Marshal(map[string]any{"files": fileNames, "target_path": targetPath})
	job := &model.TaskJob{Name: "文件分发 - " + time.Now().Format("2006-01-02 15:04:05"), TaskType: "file", Status: "running", TargetHosts: string(hostsJSON), Parameters: string(paramsJSON), ExecuteTime: time.Now(), CreatedBy: int64(userID)}
	if err := b.repo.CreateJob(ctx, b.db, job); err != nil {
		return nil, err
	}

	results := make([]*api.TaskHostResultResponse, 0, len(hostIDs))
	allSuccess := true
	for _, hostID := range hostIDs {
		result := b.distributeToHost(ctx, hostID, targetPath, files)
		results = append(results, result)
		if result.Status != "success" {
			allSuccess = false
		}
	}
	status := "success"
	if !allSuccess {
		status = "failed"
	}
	resultJSON, _ := json.Marshal(results)
	if err := b.repo.UpdateJobResult(ctx, b.db, job.ID, status, string(resultJSON), ""); err != nil {
		return nil, err
	}
	return &api.TaskDistributeResponse{TaskID: job.ID, Results: results}, nil
}

type remoteCommandResult struct {
	output []byte
	err    error
}

func (b *TaskBiz) executeOnHost(ctx context.Context, hostID int64, scriptType, content string, timeout time.Duration) *api.TaskHostResultResponse {
	result := &api.TaskHostResultResponse{HostID: hostID, Status: "failed"}
	host, client, err := b.assets.dialHostSSH(hostID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer client.Close()
	result.HostName, result.HostIP = host.Name, host.IP
	session, err := client.NewSession()
	if err != nil {
		result.Error = "创建 SSH 会话: " + err.Error()
		return result
	}
	defer session.Close()
	command := content
	if scriptType == "python" {
		command = "python3 -c " + shellQuote(content)
	}
	done := make(chan remoteCommandResult, 1)
	go func() {
		output, runErr := session.CombinedOutput(command)
		done <- remoteCommandResult{output: output, err: runErr}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case execution := <-done:
		result.Output = string(execution.output)
		if execution.err != nil {
			result.Error = execution.err.Error()
			return result
		}
		result.Status = "success"
	case <-ctx.Done():
		_ = session.Close()
		result.Error = "任务已取消"
	case <-timer.C:
		_ = session.Close()
		result.Error = "执行超时"
	}
	return result
}

func (b *TaskBiz) distributeToHost(ctx context.Context, hostID int64, targetPath string, files []*multipart.FileHeader) *api.TaskHostResultResponse {
	result := &api.TaskHostResultResponse{HostID: hostID, Status: "failed"}
	host, client, err := b.assets.dialHostSSH(hostID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer client.Close()
	result.HostName, result.HostIP = host.Name, host.IP
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		result.Error = "创建 SFTP 客户端: " + err.Error()
		return result
	}
	defer sftpClient.Close()
	if err := sftpClient.MkdirAll(targetPath); err != nil {
		result.Error = "创建目标目录: " + err.Error()
		return result
	}
	for _, header := range files {
		if err := ctx.Err(); err != nil {
			result.Error = "任务已取消"
			return result
		}
		source, err := header.Open()
		if err != nil {
			result.Error = "打开文件: " + err.Error()
			return result
		}
		remotePath := path.Join(targetPath, path.Base(header.Filename))
		target, err := sftpClient.Create(remotePath)
		if err != nil {
			source.Close()
			result.Error = "创建远程文件: " + err.Error()
			return result
		}
		_, copyErr := io.Copy(target, source)
		closeErr := target.Close()
		source.Close()
		if copyErr != nil {
			result.Error = "传输文件: " + copyErr.Error()
			return result
		}
		if closeErr != nil {
			result.Error = "保存远程文件: " + closeErr.Error()
			return result
		}
	}
	result.Status = "success"
	result.Output = fmt.Sprintf("已分发 %d 个文件到 %s", len(files), targetPath)
	return result
}

func validateTaskContent(content string) error {
	value := strings.ToLower(content)
	for _, item := range []string{"rm ", "rm\t", "unlink ", "shred ", "mkfs", "shutdown", "reboot", "poweroff", "dd if=", ":(){:|:&};:"} {
		if strings.Contains(value, item) {
			return fmt.Errorf("任务内容包含禁止执行的危险命令：%s", strings.TrimSpace(item))
		}
	}
	return nil
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }

func taskTemplateModel(req *api.TaskTemplateRequest) *model.TaskTemplate {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 300
	}
	variables := req.Variables
	if variables == "" {
		variables = "[]"
	}
	return &model.TaskTemplate{ID: req.ID, Name: req.Name, Code: req.Code, Description: req.Description, Content: req.Content, Variables: variables, Category: req.Category, Platform: req.Platform, Timeout: timeout, SortOrder: req.SortOrder, Status: *req.Status}
}

func taskTemplateResponse(row *model.TaskTemplate) *api.TaskTemplateResponse {
	return &api.TaskTemplateResponse{ID: row.ID, Name: row.Name, Code: row.Code, Description: row.Description, Content: row.Content, Variables: row.Variables, Category: row.Category, Platform: row.Platform, Timeout: row.Timeout, SortOrder: row.SortOrder, Status: row.Status, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func taskJobResponse(row *model.TaskJob, targetHostsDisplay string) *api.TaskJobResponse {
	return &api.TaskJobResponse{ID: row.ID, Name: row.Name, TemplateID: row.TemplateID, TaskType: row.TaskType, Status: row.Status, TargetHosts: row.TargetHosts, TargetHostsDisplay: targetHostsDisplay, Parameters: row.Parameters, Result: row.Result, ErrorMessage: row.ErrorMessage, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
