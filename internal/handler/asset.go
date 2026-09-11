package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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

func NewAssetHandler(assetBiz *biz.AssetBiz) *AssetHandler { return &AssetHandler{biz: assetBiz} }

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

// ListGroups POST /api/v1/assets/groups/list 查询资产分组。
func (h *AssetHandler) ListGroups(c *gin.Context) {
	req := new(api.AssetGroupListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid asset group query")
		return
	}
	result, err := h.biz.ListGroups(req)
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
	result, err := h.biz.ListCredentials(req)
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
	result, err := h.biz.ListCloudAccounts(req)
	if err != nil {
		response.Error(c, 500, "query cloud accounts failed")
		return
	}
	response.OK(c, result)
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
	result, err := h.biz.ListHosts(req)
	if err != nil {
		response.Error(c, 500, "query hosts failed")
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

// ListDBInstances POST /api/v1/assets/db-instances/list 查询数据库实例。
func (h *AssetHandler) ListDBInstances(c *gin.Context) {
	req := new(api.DbInstanceListRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid database instance query")
		return
	}
	result, err := h.biz.ListDBInstances(req)
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
