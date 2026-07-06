package api_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPIValid(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}

	requiredPaths := []string{
		"/health",
		"/openapi.yaml",
		"/api/v1/tenants",
		"/api/v1/tenants/{tenant_id}",
		"/api/v1/tenants/{tenant_id}/keys",
		"/api/v1/tenants/{tenant_id}/keys/{key_id}",
		"/api/v1/usage",
		"/v1/chat/completions",
		"/v1/models",
	}
	for _, path := range requiredPaths {
		if doc.Paths.Value(path) == nil {
			t.Fatalf("missing path %s", path)
		}
	}
}
