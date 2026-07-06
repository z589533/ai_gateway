// Panic 恢复中间件：捕获 handler panic 并返回 500，避免进程崩溃。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/pkg/response"
	"go.uber.org/zap"
)

// Recovery 将 panic 转为 JSON 500 响应并记录错误日志。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", zap.Any("panic", recovered))
				response.Error(c, http.StatusInternalServerError, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
