// 租户业务逻辑。
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/z589533/ai_gateway/internal/model"
	"gorm.io/gorm"
)

// TenantService 租户 CRUD 服务。
type TenantService struct {
	repo TenantRepo
}

// TenantList 分页租户列表响应。
type TenantList struct {
	Items    []model.Tenant `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

func NewTenantService(repo TenantRepo) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) Create(ctx context.Context, name string) (*model.Tenant, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, InvalidInput("tenant name is required")
	}
	tenant := &model.Tenant{Name: name, Status: model.TenantStatusActive}
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, mapWriteError(err, "tenant already exists")
	}
	return tenant, nil
}

func (s *TenantService) List(ctx context.Context, page, pageSize int) (*TenantList, error) {
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &TenantList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *TenantService) Get(ctx context.Context, id uint64) (*model.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("tenant not found")
	}
	return tenant, err
}

// Update 支持部分更新：name 和/或 status（active/inactive）。
func (s *TenantService) Update(ctx context.Context, id uint64, name *string, status *int8) (*model.Tenant, error) {
	tenant, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, InvalidInput("tenant name cannot be empty")
		}
		tenant.Name = trimmed
	}
	if status != nil {
		if *status != model.TenantStatusActive && *status != model.TenantStatusInactive {
			return nil, InvalidInput("invalid tenant status")
		}
		tenant.Status = *status
	}
	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, mapWriteError(err, "tenant already exists")
	}
	return tenant, nil
}

// normalizePage 统一分页默认值与上限。
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// mapWriteError 将 MySQL 唯一约束冲突映射为 409。
func mapWriteError(err error, conflictMessage string) error {
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
		return Conflict(conflictMessage)
	}
	return err
}
