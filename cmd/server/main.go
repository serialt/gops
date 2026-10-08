package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"gops/docs"
	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/pkg/casbinx"
	"gops/internal/server"

	"github.com/serialt/sugar/v3"
)

var (
	// 版本信息
	appVersion  bool // 控制是否显示版本
	APPVersion  = "v0.0.2"
	BuildTime   = "2006-01-02 15:04:05"
	GoVersion   = "go1.21"
	GitCommit   = "xxxxxxxxxxx"
	ConfigFile  = "config.yaml"
	CheckCfg    bool
	GenModel    bool
	GenTable    bool
	UserCommand string
	GenTableDir = "sql"
	Args        []string
)

func init() {
	flag.BoolVar(&appVersion, "v", false, "Display build and version messages")
	flag.BoolVar(&CheckCfg, "k", true, "check cfg")
	flag.BoolVar(&GenModel, "g", false, "gorm gen model")
	flag.StringVar(&ConfigFile, "c", "config.yaml", "Config file")
	flag.StringVar(&UserCommand, "u", "", "user command: list, create, reset")
	flag.BoolVar(&GenTable, "t", false, "goose cmd")
	flag.Parse()
	Args = flag.Args()
}

func sqliteDatabaseMissing(dsn string) bool {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return false
	}
	if dsn == ":memory:" || strings.Contains(dsn, "mode=memory") {
		return false
	}
	path := strings.SplitN(dsn, "?", 2)[0]
	path = strings.TrimPrefix(path, "file:")
	path = strings.TrimPrefix(path, "//")
	_, err := os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}

func seedDatabase(cfg *config.Config) error {
	db, cleanup, err := data.NewDB(cfg, slog.Default())
	if err != nil {
		return fmt.Errorf("open migrated database: %w", err)
	}
	defer cleanup()
	casbinService, err := casbinx.NewService(db)
	if err != nil {
		return fmt.Errorf("initialize Casbin: %w", err)
	}
	if err := server.Seed(db, casbinService); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	return nil
}

func main() {
	cfg := config.NewConfig(ConfigFile)
	if appVersion {
		fmt.Printf("AppVersion: %v\nGo Version: %v\nBuild Time: %v\nGit Commit: %v\n\n",
			APPVersion,
			GoVersion,
			BuildTime,
			GitCommit,
		)
		return
	}
	if GenModel {
		data.GenModel(cfg)
		return
	}
	if GenTable {
		if len(Args) < 1 {
			flag.Usage()
			return
		}
		if err := data.GenTable(cfg, docs.Migrations, Args, GenTableDir); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
		if Args[0] == "up" {
			if err := seedDatabase(cfg); err != nil {
				log.Fatalf("failed to initialize database: %v", err)
			}
		}
		return
	}
	if sqliteDatabaseMissing(cfg.Dsn) {
		slog.Info("SQLite database does not exist; applying migrations", "dsn", cfg.Dsn)
		if err := data.GenTable(cfg, docs.Migrations, []string{"up"}, GenTableDir); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
		if err := seedDatabase(cfg); err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
	}
	loggger := sugar.New(
		sugar.WithShort(true),
		sugar.WithLevel(cfg.Log.Level),
	)
	slog.SetDefault(loggger)

	if UserCommand != "" {
		if err := server.ManageUsers(context.Background(), UserCommand, Args, cfg); err != nil {
			log.Fatalf("user command failed: %v", err)
		}
		return
	}

	server, cleanup, err := InitializeServer(cfg, loggger)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	defer cleanup()

	log.Printf("Running server at :%v", cfg.Server.HttpPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
