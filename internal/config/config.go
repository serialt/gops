package config

import (
	"time"

	"github.com/serialt/sugar/v3"
)

type Config struct {
	Dsn    string `json:"dsn" yaml:"dsn"`
	Log    Log    `json:"log" yaml:"log"`
	Server Server `json:"server" yaml:"server"`
	Redis  Redis  `json:"redis" yaml:"redis"`
	Env    string `json:"-" yaml:"-"`
}

type Redis struct {
	Host        string `json:"host" yaml:"host"`
	Port        int64  `json:"port" yaml:"port"`
	Password    string `json:"password" yaml:"password"`
	Db          int64  `json:"db" yaml:"db"`
	PoolSize    int64  `json:"pool_size" yaml:"pool_size"`
	MinIdleConn int64  `json:"min_idle_conn" yaml:"min_idle_conn"`
}

type Server struct {
	HttpPort   int64  `json:"http_port" yaml:"http_port"`
	JwtSecret  string `json:"jwt_secret" yaml:"jwt_secret"`
	AESKey     string `json:"aes_key" yaml:"aes_key"`
	JwtTimeout time.Duration
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
	cfg.Server.JwtTimeout = 7 * 24 * time.Hour
	return
}
