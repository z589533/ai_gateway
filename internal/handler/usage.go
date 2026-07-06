// 用量查询 HTTP 处理器：对应 GET /api/v1/usage。
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/repository"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

// UsageService 用量业务接口。
type UsageService interface {
	Query(ctx context.Context, q repository.UsageQuery) (*service.UsageList, error)
}

// UsageHandler 按租户、Key、时间范围查询用量明细与汇总。
type UsageHandler struct {
	service UsageService
}

func NewUsageHandler(service UsageService) *UsageHandler {
	return &UsageHandler{service: service}
}

// Query 支持 tenant_id、api_key_id、from、to 过滤，并返回 summary 聚合。
func (h *UsageHandler) Query(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := repository.UsageQuery{
		TenantID: parseUint64Query(c, "tenant_id"),
		APIKeyID: parseUint64Query(c, "api_key_id"),
		Page:     page,
		PageSize: pageSize,
	}
	if from := c.Query("from"); from != "" {
		parsed, err := time.Parse(time.RFC3339, from)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid from")
			return
		}
		query.From = &parsed
	}
	if to := c.Query("to"); to != "" {
		parsed, err := time.Parse(time.RFC3339, to)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid to")
			return
		}
		query.To = &parsed
	}
	result, err := h.service.Query(c.Request.Context(), query)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, result)
}
