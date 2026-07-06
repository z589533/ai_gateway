// 请求日志中间件：生成 request_id 并记录延迟、状态码及鉴权上下文。
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// RequestLogger 为每个请求分配 X-Request-ID，请求结束后输出结构化访问日志。
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Set("request_id", requestID)
		c.Next()

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		}
		if auth, ok := AuthResultFromContext(c); ok {
			fields = append(fields, zap.Uint64("tenant_id", auth.TenantID), zap.Uint64("api_key_id", auth.APIKeyID))
		}
		logger.Info("http_request", fields...)
	}
}
