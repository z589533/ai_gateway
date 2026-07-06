package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/pkg/response"
)

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
