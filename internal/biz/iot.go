package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/repo"
	iotbroker "gops/internal/iot"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IOTBiz struct {
	db     *gorm.DB
	repo   *repo.IOTRepo
	broker *iotbroker.Broker
}

func NewIOTBiz(db *gorm.DB, r *repo.IOTRepo, broker *iotbroker.Broker) *IOTBiz {
	return &IOTBiz{db: db, repo: r, broker: broker}
}
func iotCond(req *api.IOTListRequest) *define.IOTListCond {
	return &define.IOTListCond{PageCond: define.PageCond{Page: int(req.Page), Limit: int(req.Limit)}, Keyword: strings.TrimSpace(req.Keyword), Status: req.Status, RoleID: req.RoleID}
}

func (b *IOTBiz) ListUsers(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTUserResponse], error) {
	rows, total, err := b.repo.ListUsers(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	links, err := b.repo.UserRoles(ctx, b.db, ids)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int64, 0, len(links))
	for _, link := range links {
		roleIDs = append(roleIDs, link.RoleID)
	}
	roles, err := b.repo.RolesByIDs(ctx, b.db, roleIDs)
	if err != nil {
		return nil, err
	}
	roleMap := map[int64]*api.IOTRoleResponse{}
	for _, role := range roles {
		roleMap[role.ID] = iotRoleResponse(role)
	}
	userRoles := map[int64][]*api.IOTRoleResponse{}
	for _, link := range links {
		if role := roleMap[link.RoleID]; role != nil {
			userRoles[link.UserID] = append(userRoles[link.UserID], role)
		}
	}
	result := make([]*api.IOTUserResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, &api.IOTUserResponse{ID: row.ID, Username: row.Username, Enabled: row.Enabled, Roles: userRoles[row.ID], CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return &api.PageResponse[*api.IOTUserResponse]{List: result, Total: total}, nil
}
func (b *IOTBiz) CreateUser(ctx context.Context, req *api.IOTUserRequest) (*api.IOTUserResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("密码至少 6 位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	row := &model.IotUser{Username: req.Username, PasswordHash: string(hash), Enabled: req.Enabled}
	err = b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := b.repo.CreateUser(ctx, tx, row); err != nil {
			return err
		}
		return b.repo.ReplaceUserRoles(ctx, tx, row.ID, req.RoleIDs)
	})
	if err != nil {
		return nil, err
	}
	return &api.IOTUserResponse{ID: row.ID, Username: row.Username, Enabled: row.Enabled, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
func (b *IOTBiz) UpdateUser(ctx context.Context, req *api.IOTUserRequest) error {
	if req.ID <= 0 {
		return ErrInvalid
	}
	if _, err := b.repo.GetUser(ctx, b.db, req.ID); err != nil {
		return mapIOTError(err)
	}
	values := map[string]any{"username": strings.TrimSpace(req.Username), "enabled": req.Enabled}
	if req.Password != "" {
		if len(req.Password) < 6 {
			return fmt.Errorf("密码至少 6 位")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		values["password_hash"] = string(hash)
	}
	return b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := b.repo.UpdateUser(ctx, tx, req.ID, values); err != nil {
			return err
		}
		return b.repo.ReplaceUserRoles(ctx, tx, req.ID, req.RoleIDs)
	})
}
func (b *IOTBiz) DeleteUser(ctx context.Context, req *api.IOTIDRequest) error {
	row, err := b.repo.GetUser(ctx, b.db, req.ID)
	if err != nil {
		return mapIOTError(err)
	}
	return b.repo.DeleteUser(ctx, b.db, row)
}

func (b *IOTBiz) ListRoles(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTRoleResponse], error) {
	rows, total, err := b.repo.ListRoles(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	result := make([]*api.IOTRoleResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, iotRoleResponse(row))
	}
	return &api.PageResponse[*api.IOTRoleResponse]{List: result, Total: total}, nil
}
func (b *IOTBiz) CreateRole(ctx context.Context, req *api.IOTRoleRequest) (*api.IOTRoleResponse, error) {
	row := &model.IotRole{Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description)}
	if err := b.repo.CreateRole(ctx, b.db, row); err != nil {
		return nil, err
	}
	return iotRoleResponse(row), nil
}
func (b *IOTBiz) UpdateRole(ctx context.Context, req *api.IOTRoleRequest) error {
	if _, err := b.repo.GetRole(ctx, b.db, req.ID); err != nil {
		return mapIOTError(err)
	}
	return b.repo.UpdateRole(ctx, b.db, req.ID, map[string]any{"name": strings.TrimSpace(req.Name), "description": strings.TrimSpace(req.Description)})
}
func (b *IOTBiz) DeleteRole(ctx context.Context, req *api.IOTIDRequest) error {
	row, err := b.repo.GetRole(ctx, b.db, req.ID)
	if err != nil {
		return mapIOTError(err)
	}
	return b.repo.DeleteRole(ctx, b.db, row)
}
func iotRoleResponse(row *model.IotRole) *api.IOTRoleResponse {
	return &api.IOTRoleResponse{ID: row.ID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func (b *IOTBiz) ListACLRules(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTACLRuleResponse], error) {
	rows, total, err := b.repo.ListACLRules(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RoleID)
	}
	roles, err := b.repo.RolesByIDs(ctx, b.db, ids)
	if err != nil {
		return nil, err
	}
	names := map[int64]string{}
	for _, role := range roles {
		names[role.ID] = role.Name
	}
	result := make([]*api.IOTACLRuleResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, &api.IOTACLRuleResponse{ID: row.ID, RoleID: row.RoleID, RoleName: names[row.RoleID], Operation: row.Operation, Effect: row.Effect, TopicFilter: row.TopicFilter, CreatedAt: row.CreatedAt})
	}
	return &api.PageResponse[*api.IOTACLRuleResponse]{List: result, Total: total}, nil
}
func (b *IOTBiz) CreateACLRule(ctx context.Context, req *api.IOTACLRuleRequest) (*api.IOTACLRuleResponse, error) {
	if _, err := b.repo.GetRole(ctx, b.db, req.RoleID); err != nil {
		return nil, mapIOTError(err)
	}
	row := &model.IotAclRule{RoleID: req.RoleID, Operation: req.Operation, Effect: req.Effect, TopicFilter: strings.TrimSpace(req.TopicFilter)}
	if err := b.repo.CreateACLRule(ctx, b.db, row); err != nil {
		return nil, err
	}
	return &api.IOTACLRuleResponse{ID: row.ID, RoleID: row.RoleID, Operation: row.Operation, Effect: row.Effect, TopicFilter: row.TopicFilter, CreatedAt: row.CreatedAt}, nil
}
func (b *IOTBiz) UpdateACLRule(ctx context.Context, req *api.IOTACLRuleRequest) error {
	if _, err := b.repo.GetACLRule(ctx, b.db, req.ID); err != nil {
		return mapIOTError(err)
	}
	return b.repo.UpdateACLRule(ctx, b.db, req.ID, map[string]any{"role_id": req.RoleID, "operation": req.Operation, "effect": req.Effect, "topic_filter": strings.TrimSpace(req.TopicFilter)})
}
func (b *IOTBiz) DeleteACLRule(ctx context.Context, req *api.IOTIDRequest) error {
	row, err := b.repo.GetACLRule(ctx, b.db, req.ID)
	if err != nil {
		return mapIOTError(err)
	}
	return b.repo.DeleteACLRule(ctx, b.db, row)
}

func (b *IOTBiz) ListTopics(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTTopicResponse], error) {
	rows, total, err := b.repo.ListTopics(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	out := make([]*api.IOTTopicResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, &api.IOTTopicResponse{ID: r.ID, Topic: r.Topic, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt})
	}
	return &api.PageResponse[*api.IOTTopicResponse]{List: out, Total: total}, nil
}
func (b *IOTBiz) ListMessages(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTMessageResponse], error) {
	rows, total, err := b.repo.ListMessages(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	out := make([]*api.IOTMessageResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, &api.IOTMessageResponse{ID: r.ID, ReceivedAt: r.ReceivedAt, ClientID: r.ClientID, Username: r.Username, Topic: r.Topic, Payload: r.Payload, QoS: r.Qos, Retained: r.Retained, Duplicate: r.Duplicate})
	}
	return &api.PageResponse[*api.IOTMessageResponse]{List: out, Total: total}, nil
}
func (b *IOTBiz) ListClients(ctx context.Context, req *api.IOTListRequest) (*api.PageResponse[*api.IOTClientResponse], error) {
	rows, total, err := b.repo.ListClients(ctx, b.db, iotCond(req))
	if err != nil {
		return nil, err
	}
	out := make([]*api.IOTClientResponse, 0, len(rows))
	for _, r := range rows {
		var disconnected *time.Time
		if !r.DisconnectedAt.IsZero() {
			v := r.DisconnectedAt
			disconnected = &v
		}
		out = append(out, &api.IOTClientResponse{ID: r.ID, ClientID: r.ClientID, Status: r.Status, Username: r.Username, RemoteAddress: r.RemoteAddress, Listener: r.Listener, ProtocolVersion: r.ProtocolVersion, CleanSession: r.CleanSession, ConnectedAt: r.ConnectedAt, DisconnectedAt: disconnected, DisconnectReason: r.DisconnectReason, SessionExpired: r.SessionExpired})
	}
	return &api.PageResponse[*api.IOTClientResponse]{List: out, Total: total}, nil
}
func (b *IOTBiz) Publish(_ context.Context, req *api.IOTPublishRequest) error {
	return b.broker.Publish(strings.TrimSpace(req.Topic), []byte(req.Payload), req.Retain, byte(req.QoS))
}
func mapIOTError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
