package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/service"
)

type fakeAuth struct {
	result *service.AuthResult
	err    error
}

func (a fakeAuth) Authenticate(context.Context, string, string) (*service.AuthResult, error) {
	return a.result, a.err
}

func TestAPIKeyAuthStoresContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1", APIKeyAuth(fakeAuth{result: &service.AuthResult{TenantID: 2, APIKeyID: 3}}, "scope"), func(c *gin.Context) {
		auth, ok := AuthResultFromContext(c)
		if !ok || auth.TenantID != 2 || auth.APIKeyID != 3 {
			t.Fatalf("auth context = %+v, %v", auth, ok)
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1", nil)
	req.Header.Set("Authorization", "Bearer sk")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d", w.Code)
	}
}
