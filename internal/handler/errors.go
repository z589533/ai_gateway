// 管理面统一错误响应封装。
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

// writeManagementError 将 service.AppError 映射为 {code,message,data} envelope。
func writeManagementError(c *gin.Context, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		response.Error(c, appErr.Status, appErr.Message)
		return
	}
	response.Error(c, http.StatusInternalServerError, "internal server error")
}
