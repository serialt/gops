package biz

import (
	"context"
	"errors"
	"strings"
	"time"

	"gops/internal/config"
	"gops/internal/data/api"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/repo"
	iotbroker "gops/internal/iot"
	"gops/internal/pkg/cryptox"

	"gorm.io/gorm"
)

type IOTBiz struct {
	db     *gorm.DB
	repo   *repo.IOTRepo
	broker *iotbroker.Broker
	aesKey string
}

func NewIOTBiz(db *gorm.DB, r *repo.IOTRepo, broker *iotbroker.Broker, cfg *config.Config) *IOTBiz {
	return &IOTBiz{db: db, repo: r, broker: broker, aesKey: cfg.Server.AESKey}
}
func iotCond(req *api.IOTListRequest) *define.IOTListCond {
	return &define.IOTListCond{PageCond: define.PageCond{Page: int(req.Page), Limit: int(req.Limit)}, Keyword: strings.TrimSpace(req.Keyword), Status: req.Status, RoleID: req.RoleID}
}

func (b *IOTBiz) Dashboard(ctx context.Context) (*api.IOTDashboardResponse, error) {
	data, err := b.repo.Dashboard(ctx, b.db)
	if err != nil {
		return nil, err
	}
	messages := make([]*api.IOTMessageResponse, 0, len(data.RecentMessages))
	for _, row := range data.RecentMessages {
		messages = append(messages, &api.IOTMessageResponse{ID: row.ID, ReceivedAt: row.ReceivedAt, ClientID: row.ClientID, Username: row.Username, Topic: row.Topic, Payload: row.Payload, QoS: row.Qos, Retained: row.Retained, Duplicate: row.Duplicate})
	}
	clients := make([]*api.IOTClientResponse, 0, len(data.RecentClients))
	for _, row := range data.RecentClients {
		var disconnected *time.Time
		if !row.DisconnectedAt.IsZero() {
			value := row.DisconnectedAt
			disconnected = &value
		}
		clients = append(clients, &api.IOTClientResponse{ID: row.ID, ClientID: row.ClientID, Status: row.Status, Username: row.Username, RemoteAddress: row.RemoteAddress, Listener: row.Listener, ProtocolVersion: row.ProtocolVersion, CleanSession: row.CleanSession, ConnectedAt: row.ConnectedAt, DisconnectedAt: disconnected, DisconnectReason: row.DisconnectReason, SessionExpired: row.SessionExpired})
	}
	return &api.IOTDashboardResponse{TopicTotal: data.TopicTotal, MessageTotal: data.MessageTotal, ClientTotal: data.ClientTotal, OnlineTotal: data.OnlineTotal, RecentMessages: messages, RecentClients: clients}, nil
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
	return b.broker.PublishTo(req.Broker, strings.TrimSpace(req.Topic), []byte(req.Payload), req.Retain, byte(req.QoS))
}
func (b *IOTBiz) ReceiveHTTP(ctx context.Context, req *api.IOTHTTPReceiveRequest) error {
	source := strings.TrimSpace(req.Source)
	if source == "" || strings.TrimSpace(req.Topic) == "" {
		return errors.New("source 和 topic 不能为空")
	}
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		clientID = source
	}
	return b.repo.SaveIncomingMessage(ctx, b.db, &model.IotReceivedMessage{
		ReceivedAt: time.Now(),
		ClientID:   "http:" + clientID,
		Username:   source,
		Topic:      strings.TrimSpace(req.Topic),
		Payload:    req.Payload,
		Qos:        req.QoS,
		Retained:   req.Retain,
	})
}
func (b *IOTBiz) BrokerOptions() []*api.IOTBrokerOptionResponse {
	options := b.broker.Options()
	result := make([]*api.IOTBrokerOptionResponse, 0, len(options))
	for _, option := range options {
		result = append(result, &api.IOTBrokerOptionResponse{Key: option.Key, Name: option.Name, URL: option.URL, Type: option.Type, Connected: option.Connected, Internal: option.Internal})
	}
	return result
}

func (b *IOTBiz) ListExternalBrokers(ctx context.Context, req *api.IOTExternalBrokerListRequest) (*api.PageResponse[*api.IOTExternalBrokerResponse], error) {
	rows, total, err := b.repo.ListExternalBrokers(ctx, b.db, &define.IOTBrokerListCond{PageCond: define.PageCond{Page: int(req.Page), Limit: int(req.Limit)}, Keyword: strings.TrimSpace(req.Keyword), Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	result := make([]*api.IOTExternalBrokerResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, externalBrokerResponse(row, b.broker.Connected(row.Name)))
	}
	return &api.PageResponse[*api.IOTExternalBrokerResponse]{List: result, Total: total}, nil
}

func (b *IOTBiz) CreateExternalBroker(ctx context.Context, req *api.IOTExternalBrokerRequest) error {
	item, encryptedPassword, err := b.externalBrokerConfig(req, "")
	if err != nil {
		return err
	}
	row := &model.IotExternalBroker{Name: item.Name, URL: item.URL, ClientID: item.ClientID, Username: item.Username, Password: encryptedPassword, TopicFilter: item.TopicFilter, Enabled: boolInt64(item.Enabled), InsecureSkipVerify: boolInt64(item.InsecureSkipVerify)}
	if err = b.repo.CreateExternalBroker(ctx, b.db, row); err != nil {
		return err
	}
	return b.broker.Apply(item, b.db, b.repo)
}

func (b *IOTBiz) UpdateExternalBroker(ctx context.Context, req *api.IOTExternalBrokerRequest) error {
	row, err := b.repo.GetExternalBroker(ctx, b.db, req.ID)
	if err != nil {
		return mapIOTError(err)
	}
	password := req.Password
	encryptedPassword := row.Password
	if password == "" {
		password, err = cryptox.Decrypt(b.aesKey, row.Password)
		if err != nil {
			return err
		}
	}
	item, nextEncryptedPassword, err := b.externalBrokerConfig(req, password)
	if err != nil {
		return err
	}
	if req.Password != "" {
		encryptedPassword = nextEncryptedPassword
	}
	values := map[string]any{"name": item.Name, "url": item.URL, "client_id": item.ClientID, "username": item.Username, "password": encryptedPassword, "topic_filter": item.TopicFilter, "enabled": boolInt64(item.Enabled), "insecure_skip_verify": boolInt64(item.InsecureSkipVerify), "updated_at": time.Now()}
	if err = b.repo.UpdateExternalBroker(ctx, b.db, row.ID, values); err != nil {
		return err
	}
	if row.Name != item.Name {
		b.broker.Remove(row.Name)
	}
	return b.broker.Apply(item, b.db, b.repo)
}

func (b *IOTBiz) DeleteExternalBroker(ctx context.Context, req *api.IOTIDRequest) error {
	row, err := b.repo.GetExternalBroker(ctx, b.db, req.ID)
	if err != nil {
		return mapIOTError(err)
	}
	if err = b.repo.DeleteExternalBroker(ctx, b.db, row); err != nil {
		return err
	}
	b.broker.Remove(row.Name)
	return nil
}

func (b *IOTBiz) externalBrokerConfig(req *api.IOTExternalBrokerRequest, passwordOverride string) (config.IOTExternalBroker, string, error) {
	item := config.IOTExternalBroker{Name: strings.TrimSpace(req.Name), Enabled: req.Enabled, URL: strings.TrimSpace(req.URL), ClientID: strings.TrimSpace(req.ClientID), Username: strings.TrimSpace(req.Username), Password: req.Password, TopicFilter: strings.TrimSpace(req.TopicFilter), InsecureSkipVerify: req.InsecureSkipVerify}
	if passwordOverride != "" || req.Password == "" {
		item.Password = passwordOverride
	}
	if item.Name == "" || item.Name == "internal" || item.URL == "" || item.ClientID == "" || item.TopicFilter == "" {
		return item, "", errors.New("Broker 名称、地址、Client ID 和 Topic 过滤器不能为空，且名称不能为 internal")
	}
	if err := iotbroker.ValidateExternalBrokerURL(item.URL); err != nil {
		return item, "", err
	}
	encrypted, err := cryptox.Encrypt(b.aesKey, item.Password)
	return item, encrypted, err
}

func externalBrokerResponse(row *model.IotExternalBroker, connected bool) *api.IOTExternalBrokerResponse {
	return &api.IOTExternalBrokerResponse{ID: row.ID, Name: row.Name, URL: row.URL, ClientID: row.ClientID, Username: row.Username, TopicFilter: row.TopicFilter, Enabled: row.Enabled == 1, Connected: connected, HasPassword: row.Password != "", InsecureSkipVerify: row.InsecureSkipVerify == 1, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func boolInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
func mapIOTError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
