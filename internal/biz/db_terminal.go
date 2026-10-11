package biz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
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
	driver       string
	dsn          string
	databaseName string
	schema       string
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
	database.SetMaxOpenConns(1)
	startedAt := time.Now()
	return &DatabaseTerminalConnection{
		db: database, databaseID: instance.ID, userID: userID,
		databaseType: instance.DbType, driver: driver, dsn: dsn, databaseName: databaseTerminalDefaultDatabase(instance.DbType),
		startedAt: startedAt, recording: newAsciinemaRecording(startedAt, 120, 32),
	}, nil
}

func databaseTerminalDefaultDatabase(databaseType string) string {
	if databaseType == "postgresql" {
		return "postgres"
	}
	return ""
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

// ResolveMetaCommand handles PostgreSQL session commands and converts query
// meta commands into SQL that can be executed by the database terminal.
func (c *DatabaseTerminalConnection) ResolveMetaCommand(ctx context.Context, statement string) (query, message string, handled bool) {
	if !strings.HasPrefix(statement, "\\") {
		return "", "", false
	}
	if c.databaseType != "postgresql" {
		return "", "Backslash commands are only supported for PostgreSQL connections.\r\n", true
	}
	parts := strings.Fields(statement)
	switch parts[0] {
	case "\\l", "\\list":
		return `SELECT datname AS name, pg_get_userbyid(datdba) AS owner, pg_encoding_to_char(encoding) AS encoding, datcollate AS collate, datctype AS ctype FROM pg_database WHERE datallowconn ORDER BY datname`, "", true
	case "\\dt":
		return `SELECT schemaname AS schema, tablename AS name, tableowner AS owner FROM pg_catalog.pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY schemaname, tablename`, "", true
	case "\\dn":
		return `SELECT nspname AS schema, pg_get_userbyid(nspowner) AS owner FROM pg_catalog.pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema' ORDER BY nspname`, "", true
	case "\\d":
		if len(parts) == 1 {
			return `SELECT schemaname AS schema, tablename AS name, tableowner AS owner FROM pg_catalog.pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY schemaname, tablename`, "", true
		}
		nameParts := strings.SplitN(parts[1], ".", 2)
		condition := fmt.Sprintf("table_name = '%s'", escapeSQLLiteral(nameParts[len(nameParts)-1]))
		if len(nameParts) == 2 {
			condition += fmt.Sprintf(" AND table_schema = '%s'", escapeSQLLiteral(nameParts[0]))
		}
		return fmt.Sprintf(`SELECT table_schema AS schema, table_name AS table, ordinal_position AS position, column_name AS column, data_type AS type, is_nullable AS nullable, column_default AS default_value FROM information_schema.columns WHERE %s ORDER BY table_schema, ordinal_position`, condition), "", true
	case "\\c", "\\connect":
		if len(parts) < 2 || len(parts) > 3 {
			return "", "Usage: \\c database [schema]\r\n", true
		}
		schema := ""
		if len(parts) == 3 {
			schema = parts[2]
		}
		if err := c.SwitchDatabase(ctx, parts[1], schema); err != nil {
			return "", fmt.Sprintf("ERROR: %v\r\n", err), true
		}
		return "", fmt.Sprintf("Connected to database %s%s.\r\n", parts[1], schemaDescription(schema)), true
	case "\\schema":
		if len(parts) != 2 {
			return "", "Usage: \\schema schema\r\n", true
		}
		if err := c.SwitchSchema(ctx, parts[1]); err != nil {
			return "", fmt.Sprintf("ERROR: %v\r\n", err), true
		}
		return "", fmt.Sprintf("Schema changed to %s.\r\n", parts[1]), true
	case "\\?":
		return "", "Supported commands: \\l, \\dn, \\dt, \\d [table], \\c database [schema], \\schema schema, \\q.\r\n", true
	default:
		return "", fmt.Sprintf("Unsupported command: %s. Type \\? for help.\r\n", parts[0]), true
	}
}

func schemaDescription(schema string) string {
	if schema == "" {
		return ""
	}
	return " with schema " + schema
}

// SwitchDatabase reconnects a PostgreSQL terminal to another database.
func (c *DatabaseTerminalConnection) SwitchDatabase(ctx context.Context, databaseName, schema string) error {
	if c.databaseType != "postgresql" {
		return errors.New("switching databases is only supported for PostgreSQL")
	}
	dsn, err := url.Parse(c.dsn)
	if err != nil {
		return fmt.Errorf("parse database connection: %w", err)
	}
	dsn.Path = "/" + databaseName
	database, err := sql.Open(c.driver, dsn.String())
	if err != nil {
		return fmt.Errorf("open database %s: %w", databaseName, err)
	}
	database.SetMaxOpenConns(1)
	if err = database.PingContext(ctx); err != nil {
		_ = database.Close()
		return fmt.Errorf("connect database %s: %w", databaseName, err)
	}
	if schema != "" {
		if _, err = database.ExecContext(ctx, "SET search_path TO "+quotePostgreSQLIdentifier(schema)); err != nil {
			_ = database.Close()
			return fmt.Errorf("switch schema %s: %w", schema, err)
		}
	}

	c.mu.Lock()
	previous := c.db
	c.db, c.dsn, c.databaseName, c.schema = database, dsn.String(), databaseName, schema
	c.mu.Unlock()
	_ = previous.Close()
	return nil
}

// SwitchSchema changes the PostgreSQL search path for the current terminal.
func (c *DatabaseTerminalConnection) SwitchSchema(ctx context.Context, schema string) error {
	if c.databaseType != "postgresql" {
		return errors.New("switching schemas is only supported for PostgreSQL")
	}
	if _, err := c.db.ExecContext(ctx, "SET search_path TO "+quotePostgreSQLIdentifier(schema)); err != nil {
		return fmt.Errorf("switch schema %s: %w", schema, err)
	}
	c.mu.Lock()
	c.schema = schema
	c.mu.Unlock()
	return nil
}

func quotePostgreSQLIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func escapeSQLLiteral(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func (c *DatabaseTerminalConnection) Prompt() string {
	switch c.databaseType {
	case "postgresql":
		c.mu.Lock()
		databaseName, schema := c.databaseName, c.schema
		c.mu.Unlock()
		if schema != "" {
			return databaseName + "/" + schema + "=# "
		}
		return databaseName + "=# "
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
