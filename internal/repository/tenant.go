package repository

import (
	"context"

	"github.com/z589533/ai_gateway/internal/model"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *TenantRepository) List(ctx context.Context, page, pageSize int) ([]model.Tenant, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Tenant{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Tenant
	err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *TenantRepository) FindByID(ctx context.Context, id uint64) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}
