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
	return
}
