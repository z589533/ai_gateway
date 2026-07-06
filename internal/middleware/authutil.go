// HTTP 鉴权头解析工具。
package middleware

import (
	"strings"
)

// bearerToken 从 "Bearer <token>" 格式中提取 token，格式不合法返回空字符串。
func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
