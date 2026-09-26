package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/middleware"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AssetHandler struct{ biz *biz.AssetBiz }

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}

const terminalHeartbeatTimeout = 3 * time.Minute

func NewAssetHandler(assetBiz *biz.AssetBiz) *AssetHandler { return &AssetHandler{biz: assetBiz} }

func (h *AssetHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/assets/groups/list", h.ListGroups)
	g.POST("/assets/groups/create", h.CreateGroup)
	g.POST("/assets/groups/update", h.UpdateGroup)
	g.POST("/assets/groups/delete", h.DeleteGroup)

	g.POST("/assets/credentials/list", h.ListCredentials)
	g.POST("/assets/credentials/create", h.CreateCredential)
	g.POST("/assets/credentials/update", h.UpdateCredential)
	g.POST("/assets/credentials/delete", h.DeleteCredential)

	g.POST("/assets/cloud-accounts/list", h.ListCloudAccounts)
	g.POST("/assets/cloud-accounts/options", h.CloudProviderOptions)
	g.POST("/assets/cloud-accounts/create", h.CreateCloudAccount)
	g.POST("/assets/cloud-accounts/update", h.UpdateCloudAccount)
	g.POST("/assets/cloud-accounts/delete", h.DeleteCloudAccount)

	g.POST("/assets/hosts/list", h.ListHosts)
	g.POST("/assets/hosts/export", h.ExportHosts)
	g.POST("/assets/hosts/template", h.HostImportTemplate)
	g.POST("/assets/hosts/import", h.ImportHosts)
	g.POST("/assets/terminal/tree", h.AssetTree)
	g.POST("/assets/hosts/create", h.CreateHost)
	g.POST("/assets/hosts/update", h.UpdateHost)
	g.POST("/assets/hosts/test", h.TestHostConnection)
	g.POST("/assets/hosts/collect", h.CollectHostInfo)
	g.POST("/assets/hosts/delete", h.DeleteHost)
	g.POST("/assets/terminal/sftp/list", h.ListSFTP)
	g.POST("/assets/terminal/sftp/upload", h.UploadSFTP)
	g.POST("/assets/terminal/sftp/download", h.DownloadSFTP)

	g.POST("/assets/db-instances/options", h.DatabaseTypeOptions)
	g.POST("/assets/db-instances/list", h.ListDBInstances)
	g.POST("/assets/db-instances/create", h.CreateDBInstance)
	g.POST("/assets/db-instances/update", h.UpdateDBInstance)
	g.POST("/assets/db-instances/test", h.TestDBInstanceConnection)
	g.POST("/assets/db-instances/collect", h.CollectDBInstanceInfo)
	g.POST("/assets/db-instances/delete", h.DeleteDBInstance)

	g.POST("/assets/terminal-sessions/list", h.ListTerminalSessions)
	g.POST("/assets/terminal-sessions/recording", h.GetTerminalSessionRecording)
	g.POST("/assets/terminal-sessions/delete", h.DeleteTerminalSession)
}

// AssetTree POST /api/v1/assets/terminal/tree returns groups and their host/database nodes.
func (h *AssetHandler) AssetTree(c *gin.Context) {
	result, err := h.biz.AssetTree(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "query asset tree failed")
		return
	}
	response.OK(c, result)
}

// ListSFTP POST /api/v1/assets/terminal/sftp/list 查询主机远程目录。
func (h *AssetHandler) ListSFTP(c *gin.Context) {
	req := new(api.SFTPListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid SFTP directory request")
		return
	}
	result, err := h.biz.ListSFTP(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// UploadSFTP POST /api/v1/assets/terminal/sftp/upload 上传文件或目录。
func (h *AssetHandler) UploadSFTP(c *gin.Context) {
	hostID, err := strconv.ParseInt(c.PostForm("host_id"), 10, 64)
	if err != nil || hostID <= 0 {
		response.Error(c, 400, "invalid host id")
		return
	}
	form, err := c.MultipartForm()
	if err != nil || len(form.File["files"]) == 0 {
		response.Error(c, 400, "no files selected")
		return
	}
	if err = h.biz.UploadSFTP(hostID, c.PostForm("path"), form.File["files"]); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// DownloadSFTP POST /api/v1/assets/terminal/sftp/download 下载选中的文件。
func (h *AssetHandler) DownloadSFTP(c *gin.Context) {
	req := new(api.SFTPDownloadRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid SFTP download request")
		return
	}
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="sftp-download.zip"`)
	if err := h.biz.DownloadSFTP(req, c.Writer); err != nil {
		return
	}
}

// ConnectTerminal GET /api/v1/assets/terminal/connect/:id 建立 SSH WebSocket 终端连接。
func (h *AssetHandler) ConnectTerminal(c *gin.Context) {
	hostID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || hostID <= 0 {
		response.Error(c, 400, "invalid host id")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "32"))
	connection, err := h.biz.StartSSH(hostID, int64(userID), cols, rows)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	ws, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.biz.CloseSSH(connection)
		return
	}
	defer ws.Close()
	defer h.biz.CloseSSH(connection)
	_ = ws.SetReadDeadline(time.Now().Add(terminalHeartbeatTimeout))

	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 4096)
		for {
			n, readErr := connection.Output().Read(buffer)
			if n > 0 {
				chunk := append([]byte(nil), buffer[:n]...)
				connection.RecordOutput(chunk)
				if ws.WriteMessage(websocket.BinaryMessage, chunk) != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()
	for {
		messageType, payload, readErr := ws.ReadMessage()
		if readErr != nil {
			return
		}
		_ = ws.SetReadDeadline(time.Now().Add(terminalHeartbeatTimeout))
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		var control struct {
			Type string `json:"type"`
			Cols int    `json:"cols"`
			Rows int    `json:"rows"`
		}
		if json.Unmarshal(payload, &control) == nil && control.Type == "resize" {
			_ = connection.Resize(control.Cols, control.Rows)
			continue
		}
		if control.Type == "heartbeat" {
			continue
		}
		if err := connection.Write(payload); err != nil && !errors.Is(err, io.EOF) {
			return
		}
		select {
		case <-done:
			return
		default:
		}
	}
}

// ConnectDatabaseTerminal GET /api/v1/assets/terminal/database/connect/:id opens a SQL console WebSocket.
func (h *AssetHandler) ConnectDatabaseTerminal(c *gin.Context) {
	databaseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || databaseID <= 0 {
		response.Error(c, 400, "invalid database id")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	connection, err := h.biz.StartDatabaseTerminal(databaseID, int64(userID))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	ws, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.biz.CloseDatabaseTerminal(connection)
		return
	}
	defer ws.Close()
	defer h.biz.CloseDatabaseTerminal(connection)
	_ = ws.SetReadDeadline(time.Now().Add(terminalHeartbeatTimeout))

	write := func(text string) error {
		connection.Record([]byte(text))
		return ws.WriteMessage(websocket.TextMessage, []byte(text))
	}
	if err = write("Connected to database SQL console. Type \\q to close.\r\n" + connection.Prompt()); err != nil {
		return
	}
	line := make([]rune, 0, 256)
	lastWasCarriageReturn := false
	for {
		messageType, payload, readErr := ws.ReadMessage()
		if readErr != nil {
			return
		}
		_ = ws.SetReadDeadline(time.Now().Add(terminalHeartbeatTimeout))
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		var control struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(payload, &control) == nil && control.Type == "resize" {
			continue
		}
		if control.Type == "heartbeat" {
			continue
		}
		for _, char := range string(payload) {
			switch char {
			case '\r', '\n':
				if char == '\n' && lastWasCarriageReturn {
					lastWasCarriageReturn = false
					continue
				}
				lastWasCarriageReturn = char == '\r'
				if write("\r\n") != nil {
					return
				}
				statement := strings.TrimSpace(string(line))
				line = line[:0]
				if statement == "\\q" || strings.EqualFold(statement, "quit") || strings.EqualFold(statement, "exit") {
					_ = write("Database session closed.\r\n")
					return
				}
				if statement != "" {
					if writeDatabaseQuery(connection, write, statement) != nil {
						return
					}
				}
				if write(connection.Prompt()) != nil {
					return
				}
			case '\b', 0x7f:
				if len(line) > 0 {
					line = line[:len(line)-1]
					if write("\b \b") != nil {
						return
					}
				}
			case 0x03:
				line = line[:0]
				if write("^C\r\n"+connection.Prompt()) != nil {
					return
				}
			default:
				lastWasCarriageReturn = false
				line = append(line, char)
				if write(string(char)) != nil {
					return
				}
			}
		}
	}
}

func writeDatabaseQuery(connection *biz.DatabaseTerminalConnection, write func(string) error, statement string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	rows, err := connection.Query(ctx, statement)
	if err != nil {
		return write(fmt.Sprintf("ERROR: %v\r\n", err))
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return write(fmt.Sprintf("ERROR: %v\r\n", err))
	}
	if len(columns) == 0 {
		return write("Query completed.\r\n")
	}
	if err = write(strings.Join(columns, " | ") + "\r\n"); err != nil {
		return err
	}
	values := make([]any, len(columns))
	destinations := make([]any, len(columns))
	for index := range values {
		destinations[index] = &values[index]
	}
	rowCount := 0
	for rows.Next() {
		if err = rows.Scan(destinations...); err != nil {
			return write(fmt.Sprintf("ERROR: %v\r\n", err))
		}
		cells := make([]string, len(values))
		for index, value := range values {
			if bytes, ok := value.([]byte); ok {
				cells[index] = string(bytes)
			} else if value == nil {
				cells[index] = "NULL"
			} else {
				cells[index] = fmt.Sprint(value)
			}
		}
		if err = write(strings.Join(cells, " | ") + "\r\n"); err != nil {
			return err
		}
		rowCount++
		if rowCount == 200 {
			if err = write("... output limited to 200 rows ...\r\n"); err != nil {
				return err
			}
			break
		}
	}
	if err = rows.Err(); err != nil {
		return write(fmt.Sprintf("ERROR: %v\r\n", err))
	}
	return write(fmt.Sprintf("(%d rows)\r\n", rowCount))
}

// ListGroups POST /api/v1/assets/groups/list 查询资产分组。
func (h *AssetHandler) ListGroups(c *gin.Context) {
	req := new(api.AssetGroupListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid asset group query")
		return
	}
	result, err := h.biz.ListGroups(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "query asset groups failed")
		return
	}
	response.OK(c, result)
}

// CreateGroup POST /api/v1/assets/groups/create 创建资产分组。
func (h *AssetHandler) CreateGroup(c *gin.Context) {
	req := new(api.AssetGroupRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid asset group data")
		return
	}
	result, err := h.biz.CreateGroup(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// UpdateGroup POST /api/v1/assets/groups/update 更新资产分组。
func (h *AssetHandler) UpdateGroup(c *gin.Context) {
	req := new(api.AssetGroupRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid asset group data")
		return
	}
	result, err := h.biz.UpdateGroup(req)
	if errors.Is(err, biz.ErrNotFound) {
		response.Error(c, 404, "asset group not found")
		return
	}
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteGroup POST /api/v1/assets/groups/delete 删除资产分组。
func (h *AssetHandler) DeleteGroup(c *gin.Context) {
	req := new(api.AssetGroupIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid asset group id")
		return
	}
	if err := h.biz.DeleteGroup(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// ListCredentials POST /api/v1/assets/credentials/list 查询资产凭证。
func (h *AssetHandler) ListCredentials(c *gin.Context) {
	req := new(api.CredentialListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid credential query")
		return
	}
	result, err := h.biz.ListCredentials(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "query credentials failed")
		return
	}
	response.OK(c, result)
}

// CreateCredential POST /api/v1/assets/credentials/create 创建资产凭证。
func (h *AssetHandler) CreateCredential(c *gin.Context) {
	req := new(api.CredentialRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid credential data")
		return
	}
	result, err := h.biz.CreateCredential(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// UpdateCredential POST /api/v1/assets/credentials/update 更新资产凭证。
func (h *AssetHandler) UpdateCredential(c *gin.Context) {
	req := new(api.CredentialRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid credential data")
		return
	}
	result, err := h.biz.UpdateCredential(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteCredential POST /api/v1/assets/credentials/delete 删除资产凭证。
func (h *AssetHandler) DeleteCredential(c *gin.Context) {
	req := new(api.CredentialIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid credential id")
		return
	}
	if err := h.biz.DeleteCredential(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// ListCloudAccounts POST /api/v1/assets/cloud-accounts/list 查询云账号。
func (h *AssetHandler) ListCloudAccounts(c *gin.Context) {
	req := new(api.CloudAccountListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid cloud account query")
		return
	}
	result, err := h.biz.ListCloudAccounts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "query cloud accounts failed")
		return
	}
	response.OK(c, result)
}

// CloudProviderOptions POST /api/v1/assets/cloud-accounts/options 查询云服务商与区域枚举。
func (h *AssetHandler) CloudProviderOptions(c *gin.Context) {
	response.OK(c, h.biz.CloudProviderOptions())
}

// CreateCloudAccount POST /api/v1/assets/cloud-accounts/create 创建云账号。
func (h *AssetHandler) CreateCloudAccount(c *gin.Context) {
	req := new(api.CloudAccountRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid cloud account data")
		return
	}
	result, err := h.biz.CreateCloudAccount(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// UpdateCloudAccount POST /api/v1/assets/cloud-accounts/update 更新云账号。
func (h *AssetHandler) UpdateCloudAccount(c *gin.Context) {
	req := new(api.CloudAccountRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid cloud account data")
		return
	}
	result, err := h.biz.UpdateCloudAccount(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteCloudAccount POST /api/v1/assets/cloud-accounts/delete 删除云账号。
func (h *AssetHandler) DeleteCloudAccount(c *gin.Context) {
	req := new(api.CloudAccountIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid cloud account id")
		return
	}
	if err := h.biz.DeleteCloudAccount(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// ListHosts POST /api/v1/assets/hosts/list 查询主机资产。
func (h *AssetHandler) ListHosts(c *gin.Context) {
	req := new(api.HostListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host query")
		return
	}
	result, err := h.biz.ListHosts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "query hosts failed")
		return
	}
	response.OK(c, result)
}

// ExportHosts POST /api/v1/assets/hosts/export 导出主机 Excel。
func (h *AssetHandler) ExportHosts(c *gin.Context) {
	req := new(api.HostListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host export request")
		return
	}
	result, err := h.biz.ExportHosts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "export hosts failed")
		return
	}
	response.OK(c, result)
}

// HostImportTemplate POST /api/v1/assets/hosts/template 下载主机导入模板。
func (h *AssetHandler) HostImportTemplate(c *gin.Context) {
	result, err := h.biz.HostImportTemplate()
	if err != nil {
		response.Error(c, 500, "create host import template failed")
		return
	}
	response.OK(c, result)
}

// ImportHosts POST /api/v1/assets/hosts/import 导入主机 Excel。
func (h *AssetHandler) ImportHosts(c *gin.Context) {
	req := new(api.HostExcelImportRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host import request")
		return
	}
	result, err := h.biz.ImportHosts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// CreateHost POST /api/v1/assets/hosts/create 创建主机资产。
func (h *AssetHandler) CreateHost(c *gin.Context) {
	req := new(api.HostRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host data")
		return
	}
	result, err := h.biz.CreateHost(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// UpdateHost POST /api/v1/assets/hosts/update 更新主机资产。
func (h *AssetHandler) UpdateHost(c *gin.Context) {
	req := new(api.HostRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host data")
		return
	}
	result, err := h.biz.UpdateHost(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// TestHostConnection POST /api/v1/assets/hosts/test 测试主机 SSH 连接。
func (h *AssetHandler) TestHostConnection(c *gin.Context) {
	req := new(api.HostTestRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host connection data")
		return
	}
	if err := h.biz.TestHostConnection(c.Request.Context(), req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// CollectHostInfo POST /api/v1/assets/hosts/collect 通过 SSH 采集主机运行信息。
func (h *AssetHandler) CollectHostInfo(c *gin.Context) {
	req := new(api.HostIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host id")
		return
	}
	result, err := h.biz.CollectHostInfo(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteHost POST /api/v1/assets/hosts/delete 删除主机资产。
func (h *AssetHandler) DeleteHost(c *gin.Context) {
	req := new(api.HostIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid host id")
		return
	}
	if err := h.biz.DeleteHost(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// DatabaseTypeOptions POST /api/v1/assets/db-instances/options 查询数据库类型枚举。
func (h *AssetHandler) DatabaseTypeOptions(c *gin.Context) {
	response.OK(c, h.biz.DatabaseTypeOptions())
}

// ListDBInstances POST /api/v1/assets/db-instances/list 查询数据库实例。
func (h *AssetHandler) ListDBInstances(c *gin.Context) {
	req := new(api.DbInstanceListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database instance query")
		return
	}
	result, err := h.biz.ListDBInstances(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, "query database instances failed")
		return
	}
	response.OK(c, result)
}

// CreateDBInstance POST /api/v1/assets/db-instances/create 创建数据库实例。
func (h *AssetHandler) CreateDBInstance(c *gin.Context) {
	req := new(api.DbInstanceRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database instance data")
		return
	}
	result, err := h.biz.CreateDBInstance(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// UpdateDBInstance POST /api/v1/assets/db-instances/update 更新数据库实例。
func (h *AssetHandler) UpdateDBInstance(c *gin.Context) {
	req := new(api.DbInstanceRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database instance data")
		return
	}
	result, err := h.biz.UpdateDBInstance(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// TestDBInstanceConnection POST /api/v1/assets/db-instances/test 测试数据库连接。
func (h *AssetHandler) TestDBInstanceConnection(c *gin.Context) {
	req := new(api.DbInstanceTestRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database connection data")
		return
	}
	result, err := h.biz.TestDBInstanceConnection(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// CollectDBInstanceInfo POST /api/v1/assets/db-instances/collect 采集数据库连接状态和版本。
func (h *AssetHandler) CollectDBInstanceInfo(c *gin.Context) {
	req := new(api.DbInstanceIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database id")
		return
	}
	result, err := h.biz.CollectDBInstanceInfo(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// DeleteDBInstance POST /api/v1/assets/db-instances/delete 删除数据库实例。
func (h *AssetHandler) DeleteDBInstance(c *gin.Context) {
	req := new(api.DbInstanceIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database instance id")
		return
	}
	if err := h.biz.DeleteDBInstance(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// ListTerminalSessions POST /api/v1/assets/terminal-sessions/list 查询终端审计会话。
func (h *AssetHandler) ListTerminalSessions(c *gin.Context) {
	req := new(api.TerminalSessionListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid terminal session query")
		return
	}
	result, err := h.biz.ListTerminalSessions(req)
	if err != nil {
		response.Error(c, 500, "query terminal sessions failed")
		return
	}
	response.OK(c, result)
}

// GetTerminalSessionRecording POST /api/v1/assets/terminal-sessions/recording 查看终端会话录制内容。
func (h *AssetHandler) GetTerminalSessionRecording(c *gin.Context) {
	req := new(api.TerminalSessionIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid terminal session id")
		return
	}
	result, err := h.biz.GetTerminalSessionRecording(req)
	if err != nil {
		response.Error(c, 404, "terminal session not found")
		return
	}
	response.OK(c, result)
}

// DeleteTerminalSession POST /api/v1/assets/terminal-sessions/delete 删除终端审计会话。
func (h *AssetHandler) DeleteTerminalSession(c *gin.Context) {
	req := new(api.TerminalSessionIDRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid terminal session id")
		return
	}
	if err := h.biz.DeleteTerminalSession(req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}
