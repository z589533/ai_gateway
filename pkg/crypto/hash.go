// API Key 哈希与前缀工具：明文 secret 不落库，仅存 SHA-256 与展示前缀。
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex 对输入做 SHA-256 并返回小写十六进制，用于 Key 查库。
func SHA256Hex(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

// Prefix 截取字符串前 n 个字符，用于生成 key_prefix 展示字段。
func Prefix(input string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(input) <= n {
		return input
	}
	return input[:n]
}
