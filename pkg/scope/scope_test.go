package scope

import "testing"

func TestHas(t *testing.T) {
	if !Has([]string{ChatCompletions}, ChatCompletions) {
		t.Fatal("expected scope to be present")
	}
	if Has([]string{}, ChatCompletions) {
		t.Fatal("empty scopes must not grant permission")
	}
	if !Has([]string{}, "") {
		t.Fatal("empty required scope should pass")
	}
}

func TestRequiredFor(t *testing.T) {
	if got := RequiredFor("POST", "/v1/chat/completions"); got != ChatCompletions {
		t.Fatalf("chat route scope = %q", got)
	}
	if got := RequiredFor("GET", "/v1/models"); got != ModelsRead {
		t.Fatalf("models route scope = %q", got)
	}
	if got := RequiredFor("GET", "/unknown"); got != "" {
		t.Fatalf("unknown route scope = %q", got)
	}
}
