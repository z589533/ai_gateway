package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	agcrypto "github.com/z589533/ai_gateway/pkg/crypto"
	"github.com/z589533/ai_gateway/pkg/scope"

	"github.com/z589533/ai_gateway/internal/model"
	"gorm.io/gorm"
)

type SecretGenerator func() (string, error)

type APIKeyService struct {
	tenantRepo TenantRepo
	keyRepo    APIKeyRepo
	cache      KeyCache
	ttl        time.Duration
	generator  SecretGenerator
}

type CreatedAPIKey struct {
	model.APIKey
	SecretKey string `json:"secret_key"`
}

type APIKeyList struct {
	Items    []model.APIKey `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

func NewAPIKeyService(tenantRepo TenantRepo, keyRepo APIKeyRepo, cache KeyCache, ttl time.Duration) *APIKeyService {
	return &APIKeyService{
		tenantRepo: tenantRepo,
		keyRepo:    keyRepo,
		cache:      cache,
		ttl:        ttl,
		generator:  GenerateSecret,
	}
}

func (s *APIKeyService) WithGenerator(generator SecretGenerator) *APIKeyService {
	s.generator = generator
	return s
}

func GenerateSecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sk-ag-" + hex.EncodeToString(buf), nil
}

func (s *APIKeyService) Create(ctx context.Context, tenantID uint64, name string, scopes []string, expiresAt *time.Time) (*CreatedAPIKey, error) {
	if strings.TrimSpace(name) == "" {
		return nil, InvalidInput("key name is required")
	}
	if _, err := s.tenantRepo.FindByID(ctx, tenantID); errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("tenant not found")
	} else if err != nil {
		return nil, err
	}
	if scopes == nil {
		scopes = scope.DefaultScopes()
	}
	secret, err := s.generator()
	if err != nil {
		return nil, err
	}
	key := &model.APIKey{
		TenantID:  tenantID,
		Name:      strings.TrimSpace(name),
		KeyPrefix: agcrypto.Prefix(secret, 12),
		KeyHash:   agcrypto.SHA256Hex(secret),
		Scopes:    model.StringList(scopes),
		Status:    model.APIKeyStatusEnabled,
		ExpiresAt: expiresAt,
	}
	if err := s.keyRepo.Create(ctx, key); err != nil {
		return nil, mapWriteError(err, "api key already exists")
	}
	return &CreatedAPIKey{APIKey: *key, SecretKey: secret}, nil
}

func (s *APIKeyService) List(ctx context.Context, tenantID uint64, page, pageSize int) (*APIKeyList, error) {
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.keyRepo.ListByTenant(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &APIKeyList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *APIKeyService) Get(ctx context.Context, tenantID, keyID uint64) (*model.APIKey, error) {
	key, err := s.keyRepo.FindByID(ctx, tenantID, keyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("api key not found")
	}
	return key, err
}

func (s *APIKeyService) Update(ctx context.Context, tenantID, keyID uint64, scopes *[]string, status *int8, expiresAt **time.Time) (*model.APIKey, error) {
	key, err := s.Get(ctx, tenantID, keyID)
	if err != nil {
		return nil, err
	}
	if scopes != nil {
		key.Scopes = model.StringList(*scopes)
	}
	if status != nil {
		if *status != model.APIKeyStatusEnabled && *status != model.APIKeyStatusDisabled {
			return nil, InvalidInput("invalid api key status")
		}
		key.Status = *status
	}
	if expiresAt != nil {
		key.ExpiresAt = *expiresAt
	}
	if err := s.keyRepo.Update(ctx, key); err != nil {
		return nil, err
	}
	_ = s.invalidate(ctx, key)
	return key, nil
}

func (s *APIKeyService) Delete(ctx context.Context, tenantID, keyID uint64) error {
	key, err := s.Get(ctx, tenantID, keyID)
	if err != nil {
		return err
	}
	if err := s.keyRepo.SoftDelete(ctx, key); err != nil {
		return err
	}
	_ = s.invalidate(ctx, key)
	return nil
}

func (s *APIKeyService) invalidate(ctx context.Context, key *model.APIKey) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Del(ctx, key.KeyHash)
}
