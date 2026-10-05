package iot

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"gops/internal/config"
	"gops/internal/data/model"
	"gops/internal/data/repo"
	"gops/internal/pkg/cryptox"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

type externalBrokerClient struct {
	config config.IOTExternalBroker
	client pahomqtt.Client
}

type BrokerOption struct {
	Key       string
	Name      string
	URL       string
	Type      string
	Connected bool
	Internal  bool
}

func newExternalBrokerClients(configs []config.IOTExternalBroker, db *gorm.DB, repository *repo.IOTRepo) (map[string]*externalBrokerClient, error) {
	clients := make(map[string]*externalBrokerClient)
	for _, item := range configs {
		if !item.Enabled {
			continue
		}
		item.Name = strings.TrimSpace(item.Name)
		item.URL = strings.TrimSpace(item.URL)
		if item.Name == "" || item.URL == "" {
			return nil, fmt.Errorf("外部 MQTT Broker 的 name 和 url 不能为空")
		}
		if item.Name == "internal" {
			return nil, fmt.Errorf("外部 MQTT Broker 名称不能使用 internal")
		}
		if strings.TrimSpace(item.ClientID) == "" {
			item.ClientID = "gops-" + item.Name
		}
		if strings.TrimSpace(item.TopicFilter) == "" {
			item.TopicFilter = "#"
		}
		if _, exists := clients[item.Name]; exists {
			return nil, fmt.Errorf("外部 MQTT Broker 名称重复: %s", item.Name)
		}
		client, err := newExternalBrokerClient(item, db, repository)
		if err != nil {
			return nil, err
		}
		clients[item.Name] = client
	}
	return clients, nil
}

func newExternalBrokerClient(item config.IOTExternalBroker, db *gorm.DB, repository *repo.IOTRepo) (*externalBrokerClient, error) {
	brokerURL, err := normalizeExternalBrokerURL(item.URL)
	if err != nil {
		return nil, fmt.Errorf("外部 MQTT Broker %s: %w", item.Name, err)
	}
	options := pahomqtt.NewClientOptions().AddBroker(brokerURL).SetClientID(item.ClientID)
	options.SetUsername(item.Username).SetPassword(item.Password)
	options.SetAutoReconnect(true).SetConnectRetry(true).SetConnectRetryInterval(5 * time.Second)
	if strings.HasPrefix(brokerURL, "ssl://") || strings.HasPrefix(brokerURL, "wss://") {
		options.SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: item.InsecureSkipVerify}) // #nosec G402 -- controlled by explicit administrator configuration.
	}
	configCopy := item
	options.SetOnConnectHandler(func(client pahomqtt.Client) {
		token := client.Subscribe(configCopy.TopicFilter, 0, externalMessageHandler(configCopy, db, repository))
		if token.Wait() && token.Error() != nil {
			slog.Error("subscribe external MQTT broker failed", "broker", configCopy.Name, "error", token.Error())
		}
	})
	options.SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
		slog.Warn("external MQTT broker connection lost", "broker", configCopy.Name, "error", err)
	})
	return &externalBrokerClient{config: item, client: pahomqtt.NewClient(options)}, nil
}

func decryptBrokerPassword(key, value string) (string, error) {
	return cryptox.Decrypt(key, value)
}

func normalizeExternalBrokerURL(value string) (string, error) {
	replacements := map[string]string{"tls://": "ssl://", "websocket://": "ws://", "websockets://": "wss://"}
	for source, target := range replacements {
		if strings.HasPrefix(value, source) {
			value = target + strings.TrimPrefix(value, source)
			break
		}
	}
	for _, prefix := range []string{"tcp://", "ssl://", "ws://", "wss://"} {
		if strings.HasPrefix(value, prefix) {
			return value, nil
		}
	}
	return "", fmt.Errorf("url 仅支持 tcp、tls、websocket、wss 协议")
}

func ValidateExternalBrokerURL(value string) error {
	_, err := normalizeExternalBrokerURL(strings.TrimSpace(value))
	return err
}

func externalMessageHandler(item config.IOTExternalBroker, db *gorm.DB, repository *repo.IOTRepo) pahomqtt.MessageHandler {
	return func(_ pahomqtt.Client, message pahomqtt.Message) {
		row := &model.IotReceivedMessage{
			ReceivedAt: time.Now(),
			ClientID:   "external:" + item.Name,
			Username:   item.Username,
			Topic:      message.Topic(),
			Payload:    string(message.Payload()),
			Qos:        int64(message.Qos()),
			Retained:   message.Retained(),
			Duplicate:  message.Duplicate(),
		}
		if err := repository.SaveIncomingMessage(context.Background(), db, row); err != nil {
			slog.Error("save external MQTT message failed", "broker", item.Name, "error", err)
		}
	}
}

func (b *Broker) connectExternalBrokers() {
	b.mutex.RLock()
	clients := make([]*externalBrokerClient, 0, len(b.externalClients))
	for _, external := range b.externalClients {
		clients = append(clients, external)
	}
	b.mutex.RUnlock()
	for _, external := range clients {
		if token := external.client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
			slog.Error("connect external MQTT broker failed", "broker", external.config.Name, "error", token.Error())
		} else if !external.client.IsConnected() {
			slog.Warn("connect external MQTT broker timed out", "broker", external.config.Name)
		}
	}
}

func (b *Broker) PublishTo(brokerName, topic string, payload []byte, retain bool, qos byte) error {
	brokerName = strings.TrimSpace(brokerName)
	if brokerName == "" || brokerName == "internal" {
		if !b.enabled {
			return fmt.Errorf("内置 MQTT Broker 未启用")
		}
		return b.server.Publish(topic, payload, retain, qos)
	}
	b.mutex.RLock()
	external, exists := b.externalClients[brokerName]
	b.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("外部 MQTT Broker 不存在: %s", brokerName)
	}
	if !external.client.IsConnected() {
		return fmt.Errorf("外部 MQTT Broker 未连接: %s", brokerName)
	}
	token := external.client.Publish(topic, qos, retain, payload)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("向外部 MQTT Broker 发布消息超时: %s", brokerName)
	}
	return token.Error()
}

func (b *Broker) Options() []BrokerOption {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	result := make([]BrokerOption, 0, len(b.externalClients)+1)
	if b.enabled {
		result = append(result, BrokerOption{Key: "internal", Name: "内置 Broker", Type: "internal", Connected: true, Internal: true})
	}
	for key, external := range b.externalClients {
		result = append(result, BrokerOption{Key: key, Name: external.config.Name, URL: external.config.URL, Type: "external", Connected: external.client.IsConnected()})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Internal != result[j].Internal {
			return result[i].Internal
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func (b *Broker) Connected(name string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	external, exists := b.externalClients[name]
	return exists && external.client.IsConnected()
}

func (b *Broker) Apply(item config.IOTExternalBroker, db *gorm.DB, repository *repo.IOTRepo) error {
	if !item.Enabled {
		b.Remove(item.Name)
		return nil
	}
	client, err := newExternalBrokerClient(item, db, repository)
	if err != nil {
		return err
	}
	b.mutex.Lock()
	previous := b.externalClients[item.Name]
	b.externalClients[item.Name] = client
	b.mutex.Unlock()
	if previous != nil {
		previous.client.Disconnect(250)
	}
	go connectExternalBroker(client)
	return nil
}

func (b *Broker) Remove(name string) {
	b.mutex.Lock()
	client := b.externalClients[name]
	delete(b.externalClients, name)
	b.mutex.Unlock()
	if client != nil {
		client.client.Disconnect(250)
	}
}

func connectExternalBroker(external *externalBrokerClient) {
	if token := external.client.Connect(); token.WaitTimeout(10*time.Second) && token.Error() != nil {
		slog.Error("connect external MQTT broker failed", "broker", external.config.Name, "error", token.Error())
	} else if !external.client.IsConnected() {
		slog.Warn("connect external MQTT broker timed out", "broker", external.config.Name)
	}
}
