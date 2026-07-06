// AuthService：数据面 API Key 鉴权（哈希查库、Redis 缓存、scope/状态校验）。
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	agcrypto "github.com/z589533/ai_gateway/pkg/crypto"
	"github.com/z589533/ai_gateway/pkg/scope"
	"gorm.io/gorm"
)

// AuthService 负责数据面 API Key 鉴权：哈希查库、Redis 缓存、scope/状态/过期校验。
type AuthService struct {
	keyRepo APIKeyRepo
	cache   KeyCache
	ttl     time.Duration
	now     Clock
}

// AuthResult 鉴权成功后写入 Gin Context，供代理与用量记录使用。
type AuthResult struct {
	TenantID uint64
	APIKeyID uint64
	Scopes   []string
}

func NewAuthService(keyRepo APIKeyRepo, cache KeyCache, ttl time.Duration) *AuthService {
	return &AuthService{keyRepo: keyRepo, cache: cache, ttl: ttl, now: realClock}
}

func (s *AuthService) WithClock(clock Clock) *AuthService {
	s.now = clock
	return s
}

// Authenticate 校验 Bearer Token：SHA-256 查 Key → 校验状态/过期/scope。
func (s *AuthService) Authenticate(ctx context.Context, bearerToken string, requiredScope string) (*AuthResult, error) {
	bearerToken = strings.TrimSpace(bearerToken)
	if bearerToken == "" {
		return nil, InvalidAPIKey()
	}
	hash := agcrypto.SHA256Hex(bearerToken)
	cached, err := s.cachedOrLoad(ctx, hash)
	if err != nil {
		return nil, err
	}
	if err := s.validateCached(cached, requiredScope); err != nil {
		return nil, err
	}
	return &AuthResult{TenantID: cached.TenantID, APIKeyID: cached.APIKeyID, Scopes: cached.Scopes}, nil
}

// cachedOrLoad 优先读 Redis 缓存，miss 时回源 MySQL 并回填缓存。
func (s *AuthService) cachedOrLoad(ctx context.Context, hash string) (*CachedKey, error) {
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, hash)
		if err == nil {
			return cached, nil
		}
		if err != nil && !errors.Is(err, ErrCacheMiss) {
			// Redis 异常时不阻断鉴权，降级走 MySQL
		}
	}
	key, err := s.keyRepo.FindByHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, InvalidAPIKey()
	}
	if err != nil {
		return nil, err
	}
	cached := &CachedKey{
		TenantID:     key.TenantID,
		TenantStatus: key.Tenant.Status,
		APIKeyID:     key.ID,
		Status:       key.Status,
		Scopes:       key.Scopes.Slice(),
		ExpiresAt:    key.ExpiresAt,
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, hash, *cached, s.ttl)
	}
	return cached, nil
}

// validateCached 按顺序校验：禁用 → 删除 → 过期 → 租户状态 → scope。
func (s *AuthService) validateCached(key *CachedKey, requiredScope string) error {
	switch {
	case key.Status == model.APIKeyStatusDisabled:
		return Forbidden("key_disabled", "API key is disabled")
	case key.Status == model.APIKeyStatusDeleted:
		return InvalidAPIKey()
	case key.ExpiresAt != nil && key.ExpiresAt.Before(s.now()):
		return Forbidden("key_expired", "API key is expired")
	case key.TenantStatus == model.TenantStatusInactive:
		return Forbidden("tenant_disabled", "tenant is disabled")
	case !scope.Has(key.Scopes, requiredScope):
		return Forbidden("insufficient_scope", "API key does not have required scope")
	default:
		return nil
	}
}
