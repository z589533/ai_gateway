package service

import (
	"context"
	"testing"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/pkg/scope"
)

type fakeTenantRepo struct {
	tenant *model.Tenant
}

func (r fakeTenantRepo) Create(context.Context, *model.Tenant) error { return nil }
func (r fakeTenantRepo) List(context.Context, int, int) ([]model.Tenant, int64, error) {
	return nil, 0, nil
}
func (r fakeTenantRepo) FindByID(context.Context, uint64) (*model.Tenant, error) {
	return r.tenant, nil
}
func (r fakeTenantRepo) Update(context.Context, *model.Tenant) error { return nil }

type captureKeyRepo struct {
	fakeKeyRepo
	created *model.APIKey
}

func (r *captureKeyRepo) Create(_ context.Context, key *model.APIKey) error {
	r.created = key
	key.ID = 1
	return nil
}

func TestAPIKeyServiceCreateDefaultsScopeAndHashes(t *testing.T) {
	keyRepo := &captureKeyRepo{}
	service := NewAPIKeyService(
		fakeTenantRepo{tenant: &model.Tenant{ID: 1, Status: model.TenantStatusActive}},
		keyRepo,
		&memoryCache{},
		time.Minute,
	).WithGenerator(func() (string, error) {
		return "sk-ag-test-secret", nil
	})

	created, err := service.Create(context.Background(), 1, "default", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if created.SecretKey != "sk-ag-test-secret" {
		t.Fatalf("secret = %q", created.SecretKey)
	}
	if len(keyRepo.created.Scopes) != 1 || keyRepo.created.Scopes[0] != scope.ChatCompletions {
		t.Fatalf("scopes = %#v", keyRepo.created.Scopes)
	}
	if keyRepo.created.KeyHash == "" || keyRepo.created.KeyHash == created.SecretKey {
		t.Fatalf("key hash not set correctly")
	}
}
