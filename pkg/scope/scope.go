package scope

const (
	ChatCompletions = "chat:completions"
	ModelsRead      = "models:read"
)

func DefaultScopes() []string {
	return []string{ChatCompletions}
}

func Has(scopes []string, required string) bool {
	if required == "" {
		return true
	}
	for _, item := range scopes {
		if item == required {
			return true
		}
	}
	return false
}

func RequiredFor(method, path string) string {
	switch {
	case method == "POST" && path == "/v1/chat/completions":
		return ChatCompletions
	case method == "GET" && path == "/v1/models":
		return ModelsRead
	default:
		return ""
	}
}
