package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

const AuthContextKey = "gateway_auth"

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

func AuthResultFromContext(c *gin.Context) (*service.AuthResult, bool) {
	value, ok := c.Get(AuthContextKey)
	if !ok {
		return nil, false
	}
	result, ok := value.(*service.AuthResult)
	return result, ok
}

func writeOpenAIServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*service.AppError); ok {
		response.OpenAIErrorJSON(c, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	response.OpenAIErrorJSON(c, 500, "internal_error", "internal server error")
}
