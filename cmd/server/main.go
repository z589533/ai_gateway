// AI Gateway 服务入口：加载配置、初始化依赖并启动 HTTP 服务。
package main

import (
	"context"
	"log"

	"github.com/z589533/ai_gateway/internal/app"
	"github.com/z589533/ai_gateway/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	application, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal("init app", zap.Error(err))
	}
	logger.Info("starting server", zap.String("addr", ":"+cfg.AppPort))
	if err := application.Server().ListenAndServe(); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
	}
}
