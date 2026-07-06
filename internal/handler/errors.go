package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

func writeManagementError(c *gin.Context, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		response.Error(c, appErr.Status, appErr.Message)
		return
	}
	response.Error(c, http.StatusInternalServerError, "internal server error")
}
