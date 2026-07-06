// API Key scope 常量与校验：控制数据面接口访问权限。
package scope

const (
	ChatCompletions = "chat:completions" // POST /v1/chat/completions
	ModelsRead      = "models:read"      // GET /v1/models
)

// DefaultScopes 新建 Key 时的默认权限，仅允许 chat completions。
func DefaultScopes() []string {
	return []string{ChatCompletions}
}

// Has 检查 scopes 是否包含 required；required 为空则视为无需校验。
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

// RequiredFor 根据 HTTP 方法与路径返回所需 scope，未知路由返回空。
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
