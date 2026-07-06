// 健康检查端点，供 compose / 负载均衡探活。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health 返回简单 JSON，表示进程已就绪。
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
