package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/z589533/ai_gateway/internal/config"
	"github.com/z589533/ai_gateway/internal/handler"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/proxy"
	"github.com/z589533/ai_gateway/internal/repository"
	"github.com/z589533/ai_gateway/internal/service"
	"go.uber.org/zap"
)

type appTenantService struct{}

func (appTenantService) Create(context.Context, string) (*model.Tenant, error) {
	return &model.Tenant{}, nil
}
func (appTenantService) List(context.Context, int, int) (*service.TenantList, error) {
	return &service.TenantList{}, nil
}
func (appTenantService) Get(context.Context, uint64) (*model.Tenant, error) {
	return &model.Tenant{}, nil
}
func (appTenantService) Update(context.Context, uint64, *string, *int8) (*model.Tenant, error) {
	return &model.Tenant{}, nil
}

type appKeyService struct{}

func (appKeyService) Create(context.Context, uint64, string, []string, *time.Time) (*service.CreatedAPIKey, error) {
	return &service.CreatedAPIKey{}, nil
}
func (appKeyService) List(context.Context, uint64, int, int) (*service.APIKeyList, error) {
	return &service.APIKeyList{}, nil
}
func (appKeyService) Get(context.Context, uint64, uint64) (*model.APIKey, error) {
	return &model.APIKey{}, nil
}
func (appKeyService) Update(context.Context, uint64, uint64, *[]string, *int8, **time.Time) (*model.APIKey, error) {
	return &model.APIKey{}, nil
}
func (appKeyService) Delete(context.Context, uint64, uint64) error { return nil }

type appUsageService struct{}

func (appUsageService) Query(context.Context, repository.UsageQuery) (*service.UsageList, error) {
	return &service.UsageList{}, nil
}
func (appUsageService) Record(context.Context, service.RecordUsageInput) error { return nil }

type appAuthService struct{}

func (appAuthService) Authenticate(context.Context, string, string) (*service.AuthResult, error) {
	return &service.AuthResult{TenantID: 1, APIKeyID: 1}, nil
}

type appProxy struct{}

func (appProxy) Chat(context.Context, proxy.ChatCompletionRequest) (*proxy.ChatCompletionResponse, error) {
	return &proxy.ChatCompletionResponse{}, nil
}
func (appProxy) Models() proxy.ModelListResponse {
	return proxy.ModelListResponse{Object: "list"}
}

func TestNewRouterHealth(t *testing.T) {
	router := NewRouter(RouterDeps{
		Config:        config.Config{AdminToken: "admin-dev-token"},
		Logger:        zap.NewNop(),
		TenantHandler: handler.NewTenantHandler(appTenantService{}),
		APIKeyHandler: handler.NewAPIKeyHandler(appKeyService{}),
		UsageHandler:  handler.NewUsageHandler(appUsageService{}),
		ProxyHandler:  handler.NewProxyHandler(appProxy{}, appUsageService{}, time.Second, zap.NewNop()),
		AuthService:   appAuthService{},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}
