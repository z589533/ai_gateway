// API Key GORM 仓储：按 hash 鉴权、按租户列表、软删除。
package repository

import (
	"context"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// ListByTenant 排除已软删除的 Key。
func (r *APIKeyRepository) ListByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]model.APIKey, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.APIKey{}).
		Where("tenant_id = ? AND status <> ?", tenantID, model.APIKeyStatusDeleted).
		Where("deleted_at IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.APIKey
	err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *APIKeyRepository) FindByID(ctx context.Context, tenantID, keyID uint64) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND status <> ? AND deleted_at IS NULL", tenantID, keyID, model.APIKeyStatusDeleted).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// FindByHash 鉴权入口：预加载租户状态，排除软删除。
func (r *APIKeyRepository) FindByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("key_hash = ? AND status <> ? AND deleted_at IS NULL", hash, model.APIKeyStatusDeleted).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) Update(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

// SoftDelete 标记 deleted 并写入 deleted_at，保留 usage_records 外键引用。
func (r *APIKeyRepository) SoftDelete(ctx context.Context, key *model.APIKey) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(key).Updates(map[string]interface{}{
		"status":     model.APIKeyStatusDeleted,
		"deleted_at": &now,
	}).Error
}
