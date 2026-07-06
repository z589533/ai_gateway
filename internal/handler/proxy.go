package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/middleware"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/proxy"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
	"go.uber.org/zap"
)

type ChatProxy interface {
	Chat(ctx context.Context, req proxy.ChatCompletionRequest) (*proxy.ChatCompletionResponse, error)
	Models() proxy.ModelListResponse
}

type UsageRecorder interface {
	Record(ctx context.Context, input service.RecordUsageInput) error
}

type ProxyHandler struct {
	proxy   ChatProxy
	usage   UsageRecorder
	timeout time.Duration
	logger  *zap.Logger
}

func NewProxyHandler(proxy ChatProxy, usage UsageRecorder, timeout time.Duration, logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{proxy: proxy, usage: usage, timeout: timeout, logger: logger}
}

func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	auth, ok := middleware.AuthResultFromContext(c)
	if !ok {
		response.OpenAIErrorJSON(c, http.StatusUnauthorized, "invalid_api_key", "Your API key is invalid")
		return
	}
	var req proxy.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.OpenAIErrorJSON(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()
	result, err := h.proxy.Chat(ctx, req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		h.recordUsage(c.Request.Context(), auth, req.Model, proxy.Usage{}, latencyMs, model.UsageStatusError)
		h.writeProxyError(c, err)
		return
	}
	h.recordUsage(c.Request.Context(), auth, req.Model, result.Usage, latencyMs, model.UsageStatusSuccess)
	c.JSON(http.StatusOK, result)
}

func (h *ProxyHandler) ListModels(c *gin.Context) {
	c.JSON(http.StatusOK, h.proxy.Models())
}

func (h *ProxyHandler) recordUsage(ctx context.Context, auth *service.AuthResult, modelName string, usage proxy.Usage, latencyMs int, status string) {
	if h.usage == nil {
		return
	}
	if err := h.usage.Record(ctx, service.RecordUsageInput{
		TenantID:         auth.TenantID,
		APIKeyID:         auth.APIKeyID,
		Model:            modelName,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		LatencyMs:        latencyMs,
		Status:           status,
		RequestedAt:      time.Now().UTC(),
	}); err != nil {
		h.logger.Warn("failed to write usage", zap.Error(err))
	}
}

func (h *ProxyHandler) writeProxyError(c *gin.Context, err error) {
	var proxyErr *proxy.Error
	if errors.As(err, &proxyErr) {
		response.OpenAIErrorJSON(c, proxyErr.Status, proxyErr.Code, proxyErr.Message)
		return
	}
	response.OpenAIErrorJSON(c, http.StatusBadGateway, "bad_gateway", "proxy request failed")
}
