package data

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"gops/internal/config"
	"log/slog"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.Config, gslog *slog.Logger) (db *gorm.DB, cleanup func(), err error) {
	logLevel := logger.Warn
	if cfg.Log.Level == "debug" {
		logLevel = logger.Info
	}
	db, err = gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	cleanup = func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	return
}

func GenModel(cfg *config.Config) {
	db, _, err := NewDB(cfg, slog.Default())
	if err != nil {
		return
	}
	g := gen.NewGenerator(gen.Config{
		// 默认会在 OutPath 目录生成CRUD代码，并且同目录下生成 model 包
		// 所以OutPath最终package不能设置为model，在有数据库表同步的情况下会产生冲突
		// 若一定要使用可以通过ModelPkgPath单独指定model package的名称
		OutPath:           "internal/data/query",
		OutFile:           "",
		ModelPkgPath:      "internal/data/model",
		WithUnitTest:      false,
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  false,
		Mode:              gen.WithDefaultQuery,
	})

	// SQLite 的整数类型统一生成 int64，避免不同位宽导致模型字段不一致。
	dataTypeMap := map[string]func(gorm.ColumnType) (dataType string){
		"INTEGER":  func(gorm.ColumnType) string { return "int64" },
		"BIGINT":   func(gorm.ColumnType) string { return "int64" },
		"SMALLINT": func(gorm.ColumnType) string { return "int64" },
	}
	g.UseDB(db)
	g.WithDataTypeMap(dataTypeMap)

	g.ApplyBasic()

	gopsTable(g)
	// 执行并生成代码
	g.Execute()
}

func GenTable(cfg *config.Config, migrations embed.FS, gooseArgs []string, dir string) {
	goose.SetBaseFS(migrations)
	command := gooseArgs[0]
	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("goose: failed to set SQLite dialect", "err", err)
		return
	}
	gooseDB, err := sql.Open("sqlite", cfg.Dsn)
	if err != nil {
		slog.Error("goose: failed to open DB", "err", err)
		return
	}
	defer func() {
		if err := gooseDB.Close(); err != nil {
			slog.Error("goose: failed to close DB", "err", err)
		}
	}()
	arguments := []string{}
	if len(gooseArgs) > 1 {
		arguments = append(arguments, gooseArgs[1:]...)
	}

	if command == "create" {
		dir = fmt.Sprintf("docs/%s", dir)
		if arguments[len(arguments)-1] != "sql" {
			arguments = append(arguments, "sql")
		}
	}

	if err := goose.RunContext(context.Background(), command, gooseDB, dir, arguments...); err != nil {
		slog.Error("goose: failed to run", "cmd", command, "err", err)
		return
	}

}
