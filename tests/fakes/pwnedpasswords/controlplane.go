package pwnedpasswords

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Infisical/infisical/tests/infra/fakenet"
)

// pwnedPasswordHash splits a password into the HIBP range-API prefix/suffix.
// The HaveIBeenPwned k-anonymity API is keyed by uppercase SHA-1: the fake
// must hash exactly like the real service (and the backend client), so a
// computationally expensive hash here would desync every lookup. Any other
// hash orients the seeded prefix/suffix away from real queries.
func pwnedPasswordHash(password string) (prefix, suffix string) {
	sum := sha1.Sum([]byte(password))
	hash := strings.ToUpper(hex.EncodeToString(sum[:]))
	return hash[:5], hash[5:]
}

// Breach lists password as seen in count breaches, until the test ends.
func Breach(tt *testing.T, adminURL, password string, count int) {
	tt.Helper()
	prefix, suffix := pwnedPasswordHash(password)
	fakenet.Open[Range](tt, adminURL, Service.Host(), prefix).Seed(tt, map[string]any{
		"suffixes": map[string]int{suffix: count},
	})
}
