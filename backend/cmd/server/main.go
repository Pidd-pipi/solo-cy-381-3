// Package main 仅负责加载配置、装配依赖、启动服务。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aasplit/aasplit/internal/config"
	"github.com/aasplit/aasplit/internal/constants"
	"github.com/aasplit/aasplit/internal/database"
	"github.com/aasplit/aasplit/internal/router"
	"github.com/aasplit/aasplit/internal/util"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}
	log := util.NewLogger()
	log.Info(fmt.Sprintf(constants.LogConfigLoaded, cfg.AppName, cfg.Env, cfg.Port, cfg.DBHost, cfg.DBName))

	slogLogger := logger.New(
		writerAdapter{log: log},
		logger.Config{SlowThreshold: 500 * time.Millisecond, LogLevel: logger.Warn},
	)
	db, err := database.Connect(cfg.DSN(), slogLogger)
	if err != nil {
		log.Error("database connect failed", "err", err)
		os.Exit(1)
	}
	log.Info(fmt.Sprintf(constants.LogDBConnected, cfg.DBHost, cfg.DBName))

	rdb, err := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Warn("redis unavailable, rate limit falls back to in-memory", "err", err)
		rdb = nil
	}
	if rdb != nil {
		log.Info(fmt.Sprintf(constants.LogRedisConnected, cfg.RedisAddr))
	}

	if err := database.EnsureBootstrapUser(db); err != nil {
		log.Error("bootstrap admin failed", "err", err)
		os.Exit(1)
	}

	engine := router.Setup(cfg, db, rdb, log)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: engine,
	}

	go func() {
		log.Info(fmt.Sprintf(constants.LogServerStart, srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server listen failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf(constants.LogServerShutdown, err))
	}
	log.Info("server stopped gracefully")
}

// writerAdapter 将 slog 输出适配为 GORM logger.Writer。
type writerAdapter struct {
	log *slog.Logger
}

func (w writerAdapter) Printf(format string, args ...interface{}) {
	w.log.Debug(fmt.Sprintf(format, args...))
}
