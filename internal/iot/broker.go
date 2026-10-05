package iot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gops/internal/config"
	"gops/internal/data/model"
	"gops/internal/data/repo"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Broker struct {
	server          *mqtt.Server
	enabled         bool
	mutex           sync.RWMutex
	externalClients map[string]*externalBrokerClient
}

func NewBroker(cfg *config.Config, db *gorm.DB, repository *repo.IOTRepo) (*Broker, error) {
	server := mqtt.New(&mqtt.Options{InlineClient: true})
	hook := &brokerHook{db: db, repo: repository, filter: cfg.IOT.MessageTopic}
	if err := server.AddHook(hook, nil); err != nil {
		return nil, err
	}
	if cfg.IOT.Enabled {
		listenerSet, err := newListeners(cfg.IOT.Listeners)
		if err != nil {
			return nil, err
		}
		if len(listenerSet) == 0 {
			return nil, fmt.Errorf("IOT MQTT 至少需要启用一个 listener")
		}
		for _, listener := range listenerSet {
			if err := server.AddListener(listener); err != nil {
				return nil, fmt.Errorf("add IOT MQTT listener %q: %w", listener.ID(), err)
			}
		}
	}
	externalClients, err := newExternalBrokerClients(cfg.IOT.ExternalBrokers, db, repository)
	if err != nil {
		return nil, err
	}
	rows, err := repository.EnabledExternalBrokers(context.Background(), db)
	if err != nil {
		return nil, fmt.Errorf("load external MQTT brokers: %w", err)
	}
	for _, row := range rows {
		password, decryptErr := decryptBrokerPassword(cfg.Server.AESKey, row.Password)
		if decryptErr != nil {
			return nil, fmt.Errorf("decrypt external MQTT broker %s password: %w", row.Name, decryptErr)
		}
		item := config.IOTExternalBroker{Name: row.Name, Enabled: true, URL: row.URL, ClientID: row.ClientID, Username: row.Username, Password: password, TopicFilter: row.TopicFilter, InsecureSkipVerify: row.InsecureSkipVerify == 1}
		client, createErr := newExternalBrokerClient(item, db, repository)
		if createErr != nil {
			return nil, createErr
		}
		externalClients[item.Name] = client
	}
	return &Broker{server: server, enabled: cfg.IOT.Enabled, externalClients: externalClients}, nil
}

func newListeners(cfg config.IOTListeners) ([]listeners.Listener, error) {
	result := make([]listeners.Listener, 0, 4)
	if cfg.TCP.IsEnabled() {
		result = append(result, listeners.NewTCP(listeners.Config{ID: "iot-tcp", Address: cfg.TCP.Address}))
	}
	if cfg.TLS.IsEnabled() {
		tlsConfig, err := loadTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("load IOT MQTT TLS certificate: %w", err)
		}
		result = append(result, listeners.NewTCP(listeners.Config{ID: "iot-tls", Address: cfg.TLS.Address, TLSConfig: tlsConfig}))
	}
	if cfg.WebSocket.IsEnabled() {
		result = append(result, listeners.NewWebsocket(listeners.Config{ID: "iot-websocket", Address: cfg.WebSocket.Address}))
	}
	if cfg.WSS.IsEnabled() {
		tlsConfig, err := loadTLSConfig(cfg.WSS)
		if err != nil {
			return nil, fmt.Errorf("load IOT MQTT WSS certificate: %w", err)
		}
		result = append(result, listeners.NewWebsocket(listeners.Config{ID: "iot-wss", Address: cfg.WSS.Address, TLSConfig: tlsConfig}))
	}
	return result, nil
}

func loadTLSConfig(cfg config.IOTListener) (*tls.Config, error) {
	if strings.TrimSpace(cfg.CertFile) == "" || strings.TrimSpace(cfg.KeyFile) == "" {
		return nil, fmt.Errorf("cert_file and key_file are required")
	}
	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certificate}}, nil
}
func (b *Broker) Run() error {
	b.connectExternalBrokers()
	if !b.enabled {
		return nil
	}
	return b.server.Serve()
}
func (b *Broker) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, external := range b.externalClients {
		external.client.Disconnect(250)
	}
	return b.server.Close()
}
func (b *Broker) Publish(topic string, payload []byte, retain bool, qos byte) error {
	return b.PublishTo("internal", topic, payload, retain, qos)
}

type brokerHook struct {
	mqtt.HookBase
	db     *gorm.DB
	repo   *repo.IOTRepo
	filter string
}

func (h *brokerHook) ID() string { return "iot-database" }
func (h *brokerHook) Provides(v byte) bool {
	return v == mqtt.OnConnectAuthenticate || v == mqtt.OnACLCheck || v == mqtt.OnPublished || v == mqtt.OnSessionEstablished || v == mqtt.OnDisconnect
}
func (h *brokerHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	ctx := context.Background()
	username := string(pk.Connect.Username)
	ip := ""
	if cl != nil {
		ip = remoteHost(cl.Net.Remote)
	}
	user, err := h.repo.FindEnabledUser(ctx, h.db, username)
	if err != nil {
		_ = h.repo.RecordLogin(ctx, h.db, 0, username, ip, "failed", "IOT 设备用户名或密码错误")
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), pk.Connect.Password) != nil {
		_ = h.repo.RecordLogin(ctx, h.db, user.ID, user.Username, ip, "failed", "IOT 设备用户名或密码错误")
		return false
	}
	_ = h.repo.RecordLogin(ctx, h.db, user.ID, user.Username, ip, "success", "IOT 设备登录成功")
	return true
}
func (h *brokerHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	user, err := h.repo.FindEnabledUser(context.Background(), h.db, string(cl.Properties.Username))
	if err != nil {
		return false
	}
	rules, err := h.repo.ACLRulesForUser(context.Background(), h.db, user.ID)
	if err != nil {
		return false
	}
	operation := "subscribe"
	if write {
		operation = "publish"
	}
	return aclAllows(rules, operation, topic)
}

func aclAllows(rules []*model.IotAclRule, operation, topic string) bool {
	allowed := false
	for _, rule := range rules {
		if rule.Operation != operation {
			continue
		}
		if rule.Effect == "deny" && topicOverlap(rule.TopicFilter, topic) {
			return false
		}
		if rule.Effect == "allow" && topicCovered(rule.TopicFilter, topic) {
			allowed = true
		}
	}
	return allowed
}
func (h *brokerHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	if strings.HasPrefix(pk.TopicName, "$SYS") || !topicCovered(h.filter, pk.TopicName) {
		return
	}
	row := &model.IotReceivedMessage{ReceivedAt: time.Now(), Topic: pk.TopicName, Payload: string(pk.Payload), Qos: int64(pk.FixedHeader.Qos), Retained: pk.FixedHeader.Retain, Duplicate: pk.FixedHeader.Dup}
	if cl != nil {
		row.ClientID = cl.ID
		row.Username = string(cl.Properties.Username)
	}
	_ = h.repo.SaveIncomingMessage(context.Background(), h.db, row)
}
func (h *brokerHook) OnSessionEstablished(cl *mqtt.Client, _ packets.Packet) {
	_ = h.repo.ConnectClient(context.Background(), h.db, &model.IotClientConnection{ClientID: cl.ID, Status: "connected", Username: string(cl.Properties.Username), RemoteAddress: remoteHost(cl.Net.Remote), Listener: cl.Net.Listener, ProtocolVersion: int64(cl.Properties.ProtocolVersion), CleanSession: cl.Properties.Clean, ConnectedAt: time.Now()})
}
func (h *brokerHook) OnDisconnect(cl *mqtt.Client, disconnectErr error, expire bool) {
	reason := ""
	if disconnectErr != nil {
		reason = disconnectErr.Error()
	}
	_ = h.repo.DisconnectClient(context.Background(), h.db, cl.ID, remoteHost(cl.Net.Remote), reason, expire)
}
func remoteHost(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return host
	}
	return strings.Trim(address, "[]")
}
func topicOverlap(left, right string) bool {
	a, b := strings.Split(left, "/"), strings.Split(right, "/")
	for i := 0; ; i++ {
		if i < len(a) && a[i] == "#" || i < len(b) && b[i] == "#" {
			return true
		}
		if i == len(a) || i == len(b) {
			return i == len(a) && i == len(b)
		}
		if a[i] != "+" && b[i] != "+" && a[i] != b[i] {
			return false
		}
	}
}
func topicCovered(allowed, requested string) bool {
	a, r := strings.Split(allowed, "/"), strings.Split(requested, "/")
	for i, part := range a {
		if part == "#" {
			return true
		}
		if i >= len(r) {
			return false
		}
		if part == "+" {
			if r[i] == "#" {
				return false
			}
			continue
		}
		if r[i] == "+" || r[i] == "#" || part != r[i] {
			return false
		}
	}
	return len(a) == len(r)
}
