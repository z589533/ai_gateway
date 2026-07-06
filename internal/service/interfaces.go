// 业务层接口定义：仓储与缓存抽象，便于单测 mock。
package service

import (
	"context"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/repository"
)

// TenantRepo 租户持久化接口。
type TenantRepo interface {
	Create(ctx context.Context, tenant *model.Tenant) error
	List(ctx context.Context, page, pageSize int) ([]model.Tenant, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
}

// APIKeyRepo API Key 持久化接口。
type APIKeyRepo interface {
	Create(ctx context.Context, key *model.APIKey) error
	ListByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]model.APIKey, int64, error)
	FindByID(ctx context.Context, tenantID, keyID uint64) (*model.APIKey, error)
	FindByHash(ctx context.Context, hash string) (*model.APIKey, error)
	Update(ctx context.Context, key *model.APIKey) error
	SoftDelete(ctx context.Context, key *model.APIKey) error
}

// UsageRepo 用量记录持久化接口。
type UsageRepo interface {
	Create(ctx context.Context, usage *model.UsageRecord) error
	Query(ctx context.Context, q repository.UsageQuery) ([]model.UsageRecord, int64, model.UsageSummary, error)
}

// KeyCache API Key 鉴权元数据缓存（Redis）。
type KeyCache interface {
	Get(ctx context.Context, hash string) (*CachedKey, error)
	Set(ctx context.Context, hash string, key CachedKey, ttl time.Duration) error
	Del(ctx context.Context, hash string) error
}

// Clock 可注入的时间源，单测中用于模拟过期。
type Clock func() time.Time

func realClock() time.Time {
	return time.Now().UTC()
}
