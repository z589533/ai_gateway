// API Key 管理 HTTP 处理器：对应 /api/v1/tenants/:tenant_id/keys 路由。
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

// APIKeyService Key 业务接口。
type APIKeyService interface {
	Create(ctx context.Context, tenantID uint64, name string, scopes []string, expiresAt *time.Time) (*service.CreatedAPIKey, error)
	List(ctx context.Context, tenantID uint64, page, pageSize int) (*service.APIKeyList, error)
	Get(ctx context.Context, tenantID, keyID uint64) (*model.APIKey, error)
	Update(ctx context.Context, tenantID, keyID uint64, scopes *[]string, status *int8, expiresAt **time.Time) (*model.APIKey, error)
	Delete(ctx context.Context, tenantID, keyID uint64) error
}

// APIKeyHandler 处理 Key CRUD；创建时返回一次性明文 secret。
type APIKeyHandler struct {
	service APIKeyService
}

func NewAPIKeyHandler(service APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

type createAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type updateAPIKeyRequest struct {
	Scopes    *[]string
	Status    *int8
	ExpiresAt **time.Time
}

// Create 为指定租户创建 Key，响应含 secret_key（仅此次返回）。
func (h *APIKeyHandler) Create(c *gin.Context) {
	tenantID, ok := parseUint64Param(c, "tenant_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid tenant_id")
		return
	}
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	var scopes []string
	if req.Scopes != nil {
		scopes = req.Scopes
	}
	key, err := h.service.Create(c.Request.Context(), tenantID, req.Name, scopes, req.ExpiresAt)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.Created(c, key)
}

// List 分页列出某租户下的 Key（不含明文 secret）。
func (h *APIKeyHandler) List(c *gin.Context) {
	tenantID, ok := parseUint64Param(c, "tenant_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid tenant_id")
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.List(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, result)
}

// Get 查询单个 Key 详情。
func (h *APIKeyHandler) Get(c *gin.Context) {
	tenantID, keyID, ok := keyParams(c)
	if !ok {
		return
	}
	key, err := h.service.Get(c.Request.Context(), tenantID, keyID)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, key)
}

// Update 更新 scope、状态或过期时间；更新后使 Redis 缓存失效。
func (h *APIKeyHandler) Update(c *gin.Context) {
	tenantID, keyID, ok := keyParams(c)
	if !ok {
		return
	}
	req, err := parseUpdateAPIKeyRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	key, err := h.service.Update(c.Request.Context(), tenantID, keyID, req.Scopes, req.Status, req.ExpiresAt)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, key)
}

// Delete 软删除 Key，保留历史用量引用。
func (h *APIKeyHandler) Delete(c *gin.Context) {
	tenantID, keyID, ok := keyParams(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), tenantID, keyID); err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func keyParams(c *gin.Context) (uint64, uint64, bool) {
	tenantID, ok := parseUint64Param(c, "tenant_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid tenant_id")
		return 0, 0, false
	}
	keyID, ok := parseUint64Param(c, "key_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid key_id")
		return 0, 0, false
	}
	return tenantID, keyID, true
}

// parseUpdateAPIKeyRequest 手动解析 PATCH 体，支持 expires_at 显式传 null 表示永不过期。
func parseUpdateAPIKeyRequest(c *gin.Context) (*updateAPIKeyRequest, error) {
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		return nil, err
	}
	req := &updateAPIKeyRequest{}
	if value, ok := raw["scopes"]; ok {
		var scopes []string
		if err := json.Unmarshal(value, &scopes); err != nil {
			return nil, err
		}
		req.Scopes = &scopes
	}
	if value, ok := raw["status"]; ok {
		var status int8
		if err := json.Unmarshal(value, &status); err != nil {
			return nil, err
		}
		req.Status = &status
	}
	if value, ok := raw["expires_at"]; ok {
		if string(value) == "null" {
			var expiresAt *time.Time
			req.ExpiresAt = &expiresAt
		} else {
			var parsed time.Time
			if err := json.Unmarshal(value, &parsed); err != nil {
				return nil, err
			}
			expiresAt := &parsed
			req.ExpiresAt = &expiresAt
		}
	}
	return req, nil
}
