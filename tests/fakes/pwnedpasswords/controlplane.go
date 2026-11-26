package pwnedpasswords

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Infisical/infisical/tests/infra/fakenet"
)

// hibpBucket returns the HIBP k-anonymity bucket for password: uppercase
// hex of SHA-1 (protocol-mandated; the API only speaks SHA-1 prefixes).
func hibpBucket(password string) string {
	sum := sha1.Sum([]byte(password))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// Breach lists password as seen in count breaches, until the test ends.
func Breach(tt *testing.T, adminURL, password string, count int) {
	tt.Helper()
	hash := hibpBucket(password)
	fakenet.Open[Range](tt, adminURL, Service.Host(), hash[:5]).Seed(tt, map[string]any{
		"suffixes": map[string]int{hash[5:]: count},
	})
}
