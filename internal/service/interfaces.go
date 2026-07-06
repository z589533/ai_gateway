package service

import (
	"context"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/repository"
)

type TenantRepo interface {
	Create(ctx context.Context, tenant *model.Tenant) error
	List(ctx context.Context, page, pageSize int) ([]model.Tenant, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
}

type APIKeyRepo interface {
	Create(ctx context.Context, key *model.APIKey) error
	ListByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]model.APIKey, int64, error)
	FindByID(ctx context.Context, tenantID, keyID uint64) (*model.APIKey, error)
	FindByHash(ctx context.Context, hash string) (*model.APIKey, error)
	Update(ctx context.Context, key *model.APIKey) error
	SoftDelete(ctx context.Context, key *model.APIKey) error
}

type UsageRepo interface {
	Create(ctx context.Context, usage *model.UsageRecord) error
	Query(ctx context.Context, q repository.UsageQuery) ([]model.UsageRecord, int64, model.UsageSummary, error)
}

type KeyCache interface {
	Get(ctx context.Context, hash string) (*CachedKey, error)
	Set(ctx context.Context, hash string, key CachedKey, ttl time.Duration) error
	Del(ctx context.Context, hash string) error
}

type Clock func() time.Time

func realClock() time.Time {
	return time.Now().UTC()
}
