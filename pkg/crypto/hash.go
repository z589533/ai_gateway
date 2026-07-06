package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256Hex(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func Prefix(input string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(input) <= n {
		return input
	}
	return input[:n]
}
