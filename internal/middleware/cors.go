// 跨域中间件：允许管理后台与代理测试页跨域访问 API。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 设置 Access-Control 响应头，OPTIONS 预检直接返回 204。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
