package biz

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"

	mysql "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
)

// DatabaseTerminalConnection holds one interactive SQL console and its audit data.
type DatabaseTerminalConnection struct {
	db           *sql.DB
	databaseID   int64
	userID       int64
	databaseType string
	startedAt    time.Time
	commands     int64
	recording    string
	mu           sync.Mutex
}

// StartDatabaseTerminal opens a SQL session for a configured database instance.
func (b *AssetBiz) StartDatabaseTerminal(databaseID, userID int64) (*DatabaseTerminalConnection, error) {
	var instance model.DbInstance
	if err := b.db.First(&instance, databaseID).Error; err != nil {
		return nil, ErrNotFound
	}
	credential, err := b.repo.GetCredential(context.Background(), b.db, instance.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("读取数据库凭据失败: %w", err)
	}
	if credential.CredentialType != "password" {
		return nil, fmt.Errorf("数据库终端仅支持用户名密码凭据")
	}
	password, err := cryptox.Decrypt(b.aesKey, credential.Password)
	if err != nil {
		return nil, fmt.Errorf("解密数据库凭据失败: %w", err)
	}
	driver, dsn, err := databaseTerminalDSN(&instance, credential.Username, password)
	if err != nil {
		return nil, err
	}
	database, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	startedAt := time.Now()
	return &DatabaseTerminalConnection{
		db: database, databaseID: instance.ID, userID: userID,
		databaseType: instance.DbType, startedAt: startedAt, recording: newAsciinemaRecording(startedAt, 120, 32),
	}, nil
}

func databaseTerminalDSN(instance *model.DbInstance, username, password string) (string, string, error) {
	address := net.JoinHostPort(instance.Host, fmt.Sprintf("%d", instance.Port))
	switch instance.DbType {
	case "mysql", "mariadb":
		cfg := mysql.NewConfig()
		cfg.User, cfg.Passwd, cfg.Net, cfg.Addr = username, password, "tcp", address
		cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 10*time.Second, 0, 0
		return "mysql", cfg.FormatDSN(), nil
	case "postgresql":
		dsn := &url.URL{Scheme: "postgres", User: url.UserPassword(username, password), Host: address, Path: "/postgres"}
		query := dsn.Query()
		query.Set("sslmode", "disable")
		dsn.RawQuery = query.Encode()
		return "pgx", dsn.String(), nil
	case "sqlserver":
		dsn := &url.URL{Scheme: "sqlserver", User: url.UserPassword(username, password), Host: address}
		query := dsn.Query()
		query.Set("encrypt", "disable")
		dsn.RawQuery = query.Encode()
		return "sqlserver", dsn.String(), nil
	default:
		return "", "", fmt.Errorf("暂不支持 %s 数据库交互终端", databaseTypeLabel(instance.DbType))
	}
}

func (c *DatabaseTerminalConnection) Query(ctx context.Context, statement string) (*sql.Rows, error) {
	c.mu.Lock()
	c.commands++
	c.mu.Unlock()
	return c.db.QueryContext(ctx, statement)
}

func (c *DatabaseTerminalConnection) Prompt() string {
	switch c.databaseType {
	case "postgresql":
		return "postgres=# "
	case "oracle":
		return "SQL> "
	case "mariadb":
		return "MariaDB> "
	case "sqlserver":
		return "sqlserver> "
	case "mysql":
		return "mysql> "
	default:
		return "sql> "
	}
}

func (c *DatabaseTerminalConnection) Record(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	appendAsciinemaEvent(&c.recording, c.startedAt, "o", data)
}

// CloseDatabaseTerminal closes the SQL connection and persists its audit record.
func (b *AssetBiz) CloseDatabaseTerminal(connection *DatabaseTerminalConnection) {
	if connection == nil {
		return
	}
	_ = connection.db.Close()
	connection.mu.Lock()
	recording, commands := connection.recording, connection.commands
	connection.mu.Unlock()
	now := time.Now()
	_ = b.db.Create(&model.TerminalSession{
		UserID: connection.userID, DbInstanceID: connection.databaseID, SessionType: "database",
		StartedAt: connection.startedAt, EndedAt: now, DurationSeconds: int64(now.Sub(connection.startedAt).Seconds()),
		CommandCount: commands, Recording: recording,
	}).Error
}

func databaseTypeLabel(value string) string {
	for _, option := range databaseTypeOptions {
		if option.Value == value {
			return option.Label
		}
	}
	return value
}
