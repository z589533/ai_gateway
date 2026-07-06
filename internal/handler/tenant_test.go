package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/service"
)

type fakeTenantService struct{}

func (fakeTenantService) Create(context.Context, string) (*model.Tenant, error) {
	return &model.Tenant{ID: 1, Name: "demo", Status: model.TenantStatusActive}, nil
}
func (fakeTenantService) List(context.Context, int, int) (*service.TenantList, error) {
	return &service.TenantList{}, nil
}
func (fakeTenantService) Get(context.Context, uint64) (*model.Tenant, error) {
	return &model.Tenant{ID: 1, Name: "demo"}, nil
}
func (fakeTenantService) Update(context.Context, uint64, *string, *int8) (*model.Tenant, error) {
	return &model.Tenant{ID: 1, Name: "demo"}, nil
}

func TestTenantHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tenants", NewTenantHandler(fakeTenantService{}).Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tenants", strings.NewReader(`{"name":"demo"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}
