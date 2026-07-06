package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	agcrypto "github.com/z589533/ai_gateway/pkg/crypto"
	"github.com/z589533/ai_gateway/pkg/scope"
	"gorm.io/gorm"
)

type fakeKeyRepo struct {
	key *model.APIKey
	err error
}

func (r fakeKeyRepo) Create(context.Context, *model.APIKey) error { return nil }
func (r fakeKeyRepo) ListByTenant(context.Context, uint64, int, int) ([]model.APIKey, int64, error) {
	return nil, 0, nil
}
func (r fakeKeyRepo) FindByID(context.Context, uint64, uint64) (*model.APIKey, error) {
	return nil, nil
}
func (r fakeKeyRepo) FindByHash(context.Context, string) (*model.APIKey, error) {
	return r.key, r.err
}
func (r fakeKeyRepo) Update(context.Context, *model.APIKey) error     { return nil }
func (r fakeKeyRepo) SoftDelete(context.Context, *model.APIKey) error { return nil }

type memoryCache struct {
	value *CachedKey
}

func (c *memoryCache) Get(context.Context, string) (*CachedKey, error) {
	if c.value == nil {
		return nil, ErrCacheMiss
	}
	return c.value, nil
}
func (c *memoryCache) Set(_ context.Context, _ string, key CachedKey, _ time.Duration) error {
	c.value = &key
	return nil
}
func (c *memoryCache) Del(context.Context, string) error {
	c.value = nil
	return nil
}

func TestAuthServiceAuthenticateSuccess(t *testing.T) {
	secret := "sk-ag-secret"
	repo := fakeKeyRepo{key: &model.APIKey{
		ID:       9,
		TenantID: 7,
		KeyHash:  agcrypto.SHA256Hex(secret),
		Status:   model.APIKeyStatusEnabled,
		Scopes:   model.StringList{scope.ChatCompletions},
		Tenant:   model.Tenant{ID: 7, Status: model.TenantStatusActive},
	}}
	auth := NewAuthService(repo, &memoryCache{}, time.Minute)

	result, err := auth.Authenticate(context.Background(), secret, scope.ChatCompletions)
	if err != nil {
		t.Fatal(err)
	}
	if result.TenantID != 7 || result.APIKeyID != 9 {
		t.Fatalf("unexpected auth result: %+v", result)
	}
}

func TestAuthServiceInvalidKey(t *testing.T) {
	auth := NewAuthService(fakeKeyRepo{err: gorm.ErrRecordNotFound}, nil, time.Minute)
	_, err := auth.Authenticate(context.Background(), "missing", scope.ChatCompletions)
	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != "invalid_api_key" {
		t.Fatalf("err = %#v", err)
	}
}

func TestAuthServiceExpiredKey(t *testing.T) {
	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	auth := NewAuthService(fakeKeyRepo{}, &memoryCache{value: &CachedKey{
		TenantID:     1,
		TenantStatus: model.TenantStatusActive,
		APIKeyID:     1,
		Status:       model.APIKeyStatusEnabled,
		Scopes:       []string{scope.ChatCompletions},
		ExpiresAt:    &past,
	}}, time.Minute).WithClock(func() time.Time {
		return time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	})

	_, err := auth.Authenticate(context.Background(), "secret", scope.ChatCompletions)
	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != "key_expired" {
		t.Fatalf("err = %#v", err)
	}
}
