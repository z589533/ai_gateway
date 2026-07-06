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

// App 聚合网关运行所需的配置、数据库、缓存、路由与日志组件。
type App struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Router *gin.Engine
	Logger *zap.Logger
}

// New 初始化 AI Gateway：连接 MySQL/Redis、自动迁移、装配业务服务与路由。
func New(ctx context.Context, cfg config.Config, logger *zap.Logger) (*App, error) {
	gin.SetMode(cfg.GinMode)

	// 1. 连接 MySQL 并自动建表
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Tenant{}, &model.APIKey{}, &model.UsageRecord{}); err != nil {
		return nil, err
	}

	// 2. 连接 Redis，用于 API Key 元数据缓存
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, DB: cfg.RedisDB})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	// 3. 装配仓储层与服务层
	tenantRepo := repository.NewTenantRepository(db)
	keyRepo := repository.NewAPIKeyRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	keyCache := service.NewRedisKeyCache(redisClient)

	tenantService := service.NewTenantService(tenantRepo)
	keyService := service.NewAPIKeyService(tenantRepo, keyRepo, keyCache, cfg.KeyCacheTTL)
	authService := service.NewAuthService(keyRepo, keyCache, cfg.KeyCacheTTL)
	usageService := service.NewUsageService(usageRepo)
	mockProxy := proxy.NewMockProxy(cfg.MockLatency, cfg.MockFail)

	// 4. 初始化 Sentinel 限流规则
	rateLimitCfg := middleware.RateLimitConfig{
		GlobalQPS: cfg.RateLimitGlobalQPS,
		KeyQPS:    cfg.RateLimitKeyQPS,
		TenantQPS: cfg.RateLimitTenantQPS,
	}
	if err := middleware.InitSentinel(rateLimitCfg); err != nil {
		return nil, err
	}

	// 5. 注册 HTTP 路由（管理面 + OpenAI 兼容数据面）
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

// RouterDeps 是注册路由所需的 handler 与中间件依赖。
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

// NewRouter 注册全部 HTTP 路由：健康检查、OpenAPI、管理 API、OpenAI 兼容代理。
func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery(deps.Logger))
	router.Use(middleware.RequestLogger(deps.Logger))

	router.GET("/health", handler.Health)
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.File(filepath.Join("api", "openapi.yaml"))
	})

	// 管理面：租户 / Key / 用量，使用 Admin Token 鉴权
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

	// 数据面：OpenAI 兼容接口，使用租户 API Key + scope + 限流
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
