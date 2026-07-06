// API Key 鉴权中间件：数据面 Bearer Token 校验与 scope 检查。
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

const AuthContextKey = "gateway_auth"

// APIKeyAuthenticator 鉴权服务接口，便于测试注入 mock。
type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, bearerToken string, requiredScope string) (*service.AuthResult, error)
}

// APIKeyAuth 数据面鉴权中间件：解析 Bearer Token，校验 scope，将 AuthResult 写入 Context。
func APIKeyAuth(auth APIKeyAuthenticator, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		result, err := auth.Authenticate(c.Request.Context(), token, requiredScope)
		if err != nil {
			writeOpenAIServiceError(c, err)
			c.Abort()
			return
		}
		c.Set(AuthContextKey, result)
		c.Next()
	}
}

// AuthResultFromContext 从 Gin Context 读取鉴权结果，供限流、代理、日志使用。
func AuthResultFromContext(c *gin.Context) (*service.AuthResult, bool) {
	value, ok := c.Get(AuthContextKey)
	if !ok {
		return nil, false
	}
	result, ok := value.(*service.AuthResult)
	return result, ok
}

// writeOpenAIServiceError 将 service.AppError 映射为 OpenAI 兼容 error JSON。
func writeOpenAIServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*service.AppError); ok {
		response.OpenAIErrorJSON(c, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	response.OpenAIErrorJSON(c, 500, "internal_error", "internal server error")
}
