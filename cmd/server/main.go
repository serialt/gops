package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"gops/docs"
	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/data/define"
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
	DBPath      string
	DevMode     bool
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
	flag.StringVar(&DBPath, "db", "", "sqlite db path")
	flag.BoolVar(&DevMode, "dev", false, "enable dev env")
	flag.Parse()
	Args = flag.Args()
}

func main() {
	cfg := config.NewConfig(ConfigFile)
	if DevMode {
		cfg.Env = define.EnvDev
	}

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
			db, cleanup, err := data.NewDB(cfg, slog.Default())
			if err != nil {
				log.Fatalf("failed to open migrated database: %v", err)
			}
			defer cleanup()
			casbinService, err := casbinx.NewService(db)
			if err != nil {
				log.Fatalf("failed to initialize Casbin: %v", err)
			}
			if err := server.Seed(db, casbinService); err != nil {
				log.Fatalf("failed to seed database: %v", err)
			}
		}
		return
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
