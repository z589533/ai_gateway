package app

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/z589533/ai_gateway/internal/config"
	"github.com/z589533/ai_gateway/internal/handler"
	"github.com/z589533/ai_gateway/internal/middleware"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/proxy"
	"github.com/z589533/ai_gateway/internal/repository"
	"github.com/z589533/ai_gateway/internal/service"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type App struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Router *gin.Engine
	Logger *zap.Logger
}

func New(ctx context.Context, cfg config.Config, logger *zap.Logger) (*App, error) {
	gin.SetMode(cfg.GinMode)

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Tenant{}, &model.APIKey{}, &model.UsageRecord{}); err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, DB: cfg.RedisDB})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	tenantRepo := repository.NewTenantRepository(db)
	keyRepo := repository.NewAPIKeyRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	keyCache := service.NewRedisKeyCache(redisClient)

	tenantService := service.NewTenantService(tenantRepo)
	keyService := service.NewAPIKeyService(tenantRepo, keyRepo, keyCache, cfg.KeyCacheTTL)
	authService := service.NewAuthService(keyRepo, keyCache, cfg.KeyCacheTTL)
	usageService := service.NewUsageService(usageRepo)
	mockProxy := proxy.NewMockProxy(cfg.MockLatency, cfg.MockFail)

	rateLimitCfg := middleware.RateLimitConfig{
		GlobalQPS: cfg.RateLimitGlobalQPS,
		KeyQPS:    cfg.RateLimitKeyQPS,
		TenantQPS: cfg.RateLimitTenantQPS,
	}
	if err := middleware.InitSentinel(rateLimitCfg); err != nil {
		return nil, err
	}

	router := NewRouter(RouterDeps{
		Config:        cfg,
		Logger:        logger,
		TenantHandler: handler.NewTenantHandler(tenantService),
		APIKeyHandler: handler.NewAPIKeyHandler(keyService),
		UsageHandler:  handler.NewUsageHandler(usageService),
		ProxyHandler:  handler.NewProxyHandler(mockProxy, usageService, cfg.ProxyTimeout, logger),
		AuthService:   authService,
		RateLimit:     rateLimitCfg,
	})

	return &App{Config: cfg, DB: db, Redis: redisClient, Router: router, Logger: logger}, nil
}

type RouterDeps struct {
	Config        config.Config
	Logger        *zap.Logger
	TenantHandler *handler.TenantHandler
	APIKeyHandler *handler.APIKeyHandler
	UsageHandler  *handler.UsageHandler
	ProxyHandler  *handler.ProxyHandler
	AuthService   middleware.APIKeyAuthenticator
	RateLimit     middleware.RateLimitConfig
}

func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery(deps.Logger))
	router.Use(middleware.RequestLogger(deps.Logger))

	router.GET("/health", handler.Health)
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.File(filepath.Join("api", "openapi.yaml"))
	})

	admin := router.Group("/api/v1", middleware.AdminAuth(deps.Config.AdminToken))
	{
		admin.POST("/tenants", deps.TenantHandler.Create)
		admin.GET("/tenants", deps.TenantHandler.List)
		admin.GET("/tenants/:tenant_id", deps.TenantHandler.Get)
		admin.PATCH("/tenants/:tenant_id", deps.TenantHandler.Update)

		admin.POST("/tenants/:tenant_id/keys", deps.APIKeyHandler.Create)
		admin.GET("/tenants/:tenant_id/keys", deps.APIKeyHandler.List)
		admin.GET("/tenants/:tenant_id/keys/:key_id", deps.APIKeyHandler.Get)
		admin.PATCH("/tenants/:tenant_id/keys/:key_id", deps.APIKeyHandler.Update)
		admin.DELETE("/tenants/:tenant_id/keys/:key_id", deps.APIKeyHandler.Delete)

		admin.GET("/usage", deps.UsageHandler.Query)
	}

	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions",
			middleware.APIKeyAuth(deps.AuthService, "chat:completions"),
			middleware.SentinelRateLimit(deps.RateLimit),
			deps.ProxyHandler.ChatCompletions)
		v1.GET("/models",
			middleware.APIKeyAuth(deps.AuthService, "models:read"),
			deps.ProxyHandler.ListModels)
	}
	return router
}

func (a *App) Server() *http.Server {
	return &http.Server{
		Addr:              ":" + a.Config.AppPort,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
