package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/service"
	"github.com/z589533/ai_gateway/pkg/response"
)

type TenantService interface {
	Create(ctx context.Context, name string) (*model.Tenant, error)
	List(ctx context.Context, page, pageSize int) (*service.TenantList, error)
	Get(ctx context.Context, id uint64) (*model.Tenant, error)
	Update(ctx context.Context, id uint64, name *string, status *int8) (*model.Tenant, error)
}

type TenantHandler struct {
	service TenantService
}

func NewTenantHandler(service TenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

type createTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateTenantRequest struct {
	Name   *string `json:"name"`
	Status *int8   `json:"status"`
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tenant, err := h.service.Create(c.Request.Context(), req.Name)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.Created(c, tenant)
}

func (h *TenantHandler) List(c *gin.Context) {
	page, pageSize := parsePage(c)
	result, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *TenantHandler) Get(c *gin.Context) {
	id, ok := parseUint64Param(c, "tenant_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid tenant_id")
		return
	}
	tenant, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, tenant)
}

func (h *TenantHandler) Update(c *gin.Context) {
	id, ok := parseUint64Param(c, "tenant_id")
	if !ok {
		response.Error(c, http.StatusBadRequest, "invalid tenant_id")
		return
	}
	var req updateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tenant, err := h.service.Update(c.Request.Context(), id, req.Name, req.Status)
	if err != nil {
		writeManagementError(c, err)
		return
	}
	response.OK(c, tenant)
}
