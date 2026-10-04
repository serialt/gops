package repo

import (
	"context"
	"errors"
	"time"

	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/query"
	"gops/internal/pkg/utils"

	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type IOTRepo struct{}

func NewIOTRepo() *IOTRepo { return &IOTRepo{} }

func (r *IOTRepo) GetRole(ctx context.Context, db *gorm.DB, id int64) (*model.Role, error) {
	q := query.Use(db).Role
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}
func (r *IOTRepo) RolesByIDs(ctx context.Context, db *gorm.DB, ids []int64) ([]*model.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := query.Use(db).Role
	return q.WithContext(ctx).Where(q.ID.In(ids...)).Find()
}

func (r *IOTRepo) ListACLRules(ctx context.Context, db *gorm.DB, cond *define.IOTListCond) ([]*model.IotAclRule, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).IotAclRule
	d := q.WithContext(ctx)
	if cond.Keyword != "" {
		d = d.Where(q.TopicFilter.Like("%" + cond.Keyword + "%"))
	}
	if cond.RoleID > 0 {
		d = d.Where(q.RoleID.Eq(cond.RoleID))
	}
	if cond.Status != "" {
		d = d.Where(q.Effect.Eq(cond.Status))
	}
	return d.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *IOTRepo) GetACLRule(ctx context.Context, db *gorm.DB, id int64) (*model.IotAclRule, error) {
	q := query.Use(db).IotAclRule
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}
func (r *IOTRepo) CreateACLRule(ctx context.Context, db *gorm.DB, row *model.IotAclRule) error {
	return query.Use(db).IotAclRule.WithContext(ctx).Create(row)
}
func (r *IOTRepo) UpdateACLRule(ctx context.Context, db *gorm.DB, id int64, values map[string]any) error {
	q := query.Use(db).IotAclRule
	_, err := q.WithContext(ctx).Where(q.ID.Eq(id)).Updates(values)
	return err
}
func (r *IOTRepo) DeleteACLRule(ctx context.Context, db *gorm.DB, row *model.IotAclRule) error {
	_, err := query.Use(db).IotAclRule.WithContext(ctx).Delete(row)
	return err
}

func (r *IOTRepo) ListTopics(ctx context.Context, db *gorm.DB, cond *define.IOTListCond) ([]*model.IotTopic, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).IotTopic
	d := q.WithContext(ctx)
	if cond.Keyword != "" {
		d = d.Where(q.Topic.Like("%" + cond.Keyword + "%"))
	}
	return d.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *IOTRepo) ListMessages(ctx context.Context, db *gorm.DB, cond *define.IOTListCond) ([]*model.IotReceivedMessage, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).IotReceivedMessage
	d := q.WithContext(ctx)
	if cond.Keyword != "" {
		v := "%" + cond.Keyword + "%"
		d = d.Where(field.Or(q.Topic.Like(v), q.ClientID.Like(v), q.Username.Like(v), q.Payload.Like(v)))
	}
	return d.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *IOTRepo) ListClients(ctx context.Context, db *gorm.DB, cond *define.IOTListCond) ([]*model.IotClientConnection, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).IotClientConnection
	d := q.WithContext(ctx)
	if cond.Keyword != "" {
		v := "%" + cond.Keyword + "%"
		d = d.Where(field.Or(q.ClientID.Like(v), q.Username.Like(v), q.RemoteAddress.Like(v)))
	}
	if cond.Status != "" {
		d = d.Where(q.Status.Eq(cond.Status))
	}
	return d.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *IOTRepo) SavePublishedMessage(ctx context.Context, db *gorm.DB, topic, payload string, qos int64, retain bool) error {
	q := query.Use(db)
	if err := q.IotReceivedMessage.WithContext(ctx).Create(&model.IotReceivedMessage{Topic: topic, Payload: payload, Qos: qos, Retained: retain}); err != nil {
		return err
	}
	existing, err := q.IotTopic.WithContext(ctx).Where(q.IotTopic.Topic.Eq(topic)).First()
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return q.IotTopic.WithContext(ctx).Create(&model.IotTopic{Topic: topic})
}
func (r *IOTRepo) FindEnabledUser(ctx context.Context, db *gorm.DB, username string) (*model.User, error) {
	q := query.Use(db).User
	return q.WithContext(ctx).Where(q.Username.Eq(username), q.Status.Is(true)).First()
}
func (r *IOTRepo) ACLRulesForUser(ctx context.Context, db *gorm.DB, userID int64) ([]*model.IotAclRule, error) {
	q := query.Use(db)
	links, err := q.UserRole.WithContext(ctx).Where(q.UserRole.UserID.Eq(userID)).Find()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.RoleID)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return q.IotAclRule.WithContext(ctx).Where(q.IotAclRule.RoleID.In(ids...)).Find()
}
func (r *IOTRepo) SaveIncomingMessage(ctx context.Context, db *gorm.DB, row *model.IotReceivedMessage) error {
	if err := query.Use(db).IotReceivedMessage.WithContext(ctx).Create(row); err != nil {
		return err
	}
	q := query.Use(db).IotTopic
	_, err := q.WithContext(ctx).Where(q.Topic.Eq(row.Topic)).First()
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return q.WithContext(ctx).Create(&model.IotTopic{Topic: row.Topic})
}
func (r *IOTRepo) ConnectClient(ctx context.Context, db *gorm.DB, row *model.IotClientConnection) error {
	q := query.Use(db).IotClientConnection
	current, err := q.WithContext(ctx).Where(q.ClientID.Eq(row.ClientID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return q.WithContext(ctx).Create(row)
	}
	if err != nil {
		return err
	}
	row.ID = current.ID
	_, err = q.WithContext(ctx).Where(q.ID.Eq(current.ID)).Updates(map[string]any{"status": row.Status, "username": row.Username, "remote_address": row.RemoteAddress, "listener": row.Listener, "protocol_version": row.ProtocolVersion, "clean_session": row.CleanSession, "connected_at": row.ConnectedAt, "updated_at": time.Now()})
	return err
}
func (r *IOTRepo) DisconnectClient(ctx context.Context, db *gorm.DB, clientID, ip, reason string, expired bool) error {
	q := query.Use(db).IotClientConnection
	now := time.Now()
	_, err := q.WithContext(ctx).Where(q.ClientID.Eq(clientID)).Updates(map[string]any{"status": "disconnected", "disconnected_at": now, "last_disconnected_ip": ip, "disconnect_reason": reason, "session_expired": expired, "updated_at": now})
	return err
}
