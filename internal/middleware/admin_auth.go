// 管理面鉴权：校验 Authorization Bearer 与 ADMIN_TOKEN 环境变量一致。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/pkg/response"
)

// AdminAuth 保护 /api/v1/* 管理接口，Token 不匹配返回 401。
func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" || token != adminToken {
			response.Error(c, http.StatusUnauthorized, "invalid admin token")
			c.Abort()
			return
		}
		c.Next()
	}
}
