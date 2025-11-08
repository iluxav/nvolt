package crypto

import (
	"crypto/subtle"
)

// ZeroBytes securely zeros out a byte slice to prevent secrets from remaining in memory
// This is a best-effort approach as Go's GC may have created copies
func ZeroBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	for i := range b {
		b[i] = 0
	}
}

// SecureCompare performs a constant-time comparison of two byte slices
// This prevents timing attacks when comparing secrets
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// SecureCompareString performs a constant-time comparison of two strings
// This prevents timing attacks when comparing secrets
func SecureCompareString(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
