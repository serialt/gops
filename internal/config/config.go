package config

import (
	"time"

	"github.com/serialt/sugar/v3"
)

type Config struct {
	Dsn    string `json:"dsn" yaml:"dsn"`
	Log    Log    `json:"log" yaml:"log"`
	Server Server `json:"server" yaml:"server"`
	IOT    IOT    `json:"iot" yaml:"iot"`
}

type IOT struct {
	Enabled         bool                `json:"enabled" yaml:"enabled"`
	Listeners       IOTListeners        `json:"listeners" yaml:"listeners"`
	MessageTopic    string              `json:"message_topic" yaml:"message_topic"`
	ExternalBrokers []IOTExternalBroker `json:"external_brokers" yaml:"external_brokers"`
}

type IOTExternalBroker struct {
	Name               string `json:"name" yaml:"name"`
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	URL                string `json:"url" yaml:"url"`
	ClientID           string `json:"client_id" yaml:"client_id"`
	Username           string `json:"username" yaml:"username"`
	Password           string `json:"password" yaml:"password"`
	TopicFilter        string `json:"topic_filter" yaml:"topic_filter"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

type IOTListeners struct {
	TCP       IOTListener `json:"tcp" yaml:"tcp"`
	TLS       IOTListener `json:"tls" yaml:"tls"`
	WebSocket IOTListener `json:"websocket" yaml:"websocket"`
	WSS       IOTListener `json:"wss" yaml:"wss"`
}

type IOTListener struct {
	Enabled  *bool  `json:"enabled" yaml:"enabled"`
	Address  string `json:"address" yaml:"address"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

func (c IOTListener) IsEnabled() bool { return c.Enabled != nil && *c.Enabled }

type Server struct {
	HttpPort            int64         `json:"http_port" yaml:"http_port"`
	JwtSecret           string        `json:"jwt_secret" yaml:"jwt_secret"`
	AESKey              string        `json:"aes_key" yaml:"aes_key"`
	JwtTimeout          time.Duration `json:"jwt_timeout" yaml:"jwt_timeout"`
	AccessTokenTimeout  time.Duration `json:"access_token_timeout" yaml:"access_token_timeout"`
	RefreshTokenTimeout time.Duration `json:"refresh_token_timeout" yaml:"refresh_token_timeout"`
}

type Log struct {
	File  any    `json:"file" yaml:"file"`
	Level string `json:"level" yaml:"level"`
}

func NewConfig(cfgPath string) (cfg *Config) {
	cfg = &Config{}
	sugar.LoadConfig(cfgPath, cfg)

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Server.AccessTokenTimeout <= 0 {
		cfg.Server.AccessTokenTimeout = 15 * time.Minute
	}
	if cfg.Server.RefreshTokenTimeout <= 0 {
		if cfg.Server.JwtTimeout > 0 {
			cfg.Server.RefreshTokenTimeout = cfg.Server.JwtTimeout
		} else {
			cfg.Server.RefreshTokenTimeout = 7 * 24 * time.Hour
		}
	}
	cfg.Server.JwtTimeout = cfg.Server.RefreshTokenTimeout
	applyIOTListenerDefaults(&cfg.IOT.Listeners.TCP, true, ":1883")
	applyIOTListenerDefaults(&cfg.IOT.Listeners.TLS, false, ":8883")
	applyIOTListenerDefaults(&cfg.IOT.Listeners.WebSocket, false, ":8083")
	applyIOTListenerDefaults(&cfg.IOT.Listeners.WSS, false, ":8084")
	if cfg.IOT.MessageTopic == "" {
		cfg.IOT.MessageTopic = "#"
	}
	for index := range cfg.IOT.ExternalBrokers {
		broker := &cfg.IOT.ExternalBrokers[index]
		if broker.ClientID == "" {
			broker.ClientID = "gops-" + broker.Name
		}
		if broker.TopicFilter == "" {
			broker.TopicFilter = "#"
		}
	}
	return
}

func applyIOTListenerDefaults(listener *IOTListener, enabled bool, address string) {
	if listener.Enabled == nil {
		listener.Enabled = &enabled
	}
	if listener.Address == "" {
		listener.Address = address
	}
}
