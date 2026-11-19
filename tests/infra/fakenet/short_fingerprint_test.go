package fakenet

import (
	"strings"
	"testing"
)

// Pins the CodeQL go/weak-sensitive-data-hashing disposition for short(): it
// is a stable, compact log fingerprint of a credential (never the credential
// itself, never used for auth). An expensive salted hash (bcrypt/scrypt) would
// be non-deterministic and ~100ms per call in this hot logging path,
// destroying same-credential log correlation. This test FAILS if short() ever
// becomes non-deterministic or non-compact, documenting the keep.
func TestShortFingerprintIsStableAndCompact(t *testing.T) {
	t.Parallel()

	a, b := short("sk-live-abc123"), short("sk-live-abc123")
	if a != b {
		t.Fatalf("short() not stable: %q vs %q", a, b)
	}
	if len(a) > 20 || !strings.HasSuffix(a, "...") {
		t.Fatalf("short() not compact: %q", a)
	}
	if c := short("sk-live-different"); c == a {
		t.Fatalf("short() collides on distinct keys: %q", c)
	}
}
