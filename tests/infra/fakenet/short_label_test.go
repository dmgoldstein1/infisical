package fakenet

import (
	"strings"
	"testing"
)

// Pins the CodeQL go/weak-sensitive-data-hashing disposition for short(): the
// SHA-256 digest is a truncated, deterministic log label for a fake test
// credential — never stored, never compared as a secret. A salted slow hash
// (bcrypt/scrypt) would give a fresh label per call and break log correlation;
// leaking a credential prefix instead would be worse. This test FAILS if the
// label stops being deterministic 12-hex, documenting the keep.
func TestShortLabelIsDeterministicTruncatedDigest(t *testing.T) {
	t.Parallel()

	a := short("fake-credential-1")
	b := short("fake-credential-1")
	if a != b {
		t.Fatalf("short() not deterministic: %q vs %q", a, b)
	}
	if len(a) != 15 || !strings.HasSuffix(a, "...") {
		t.Fatalf("short() lost its 12-hex-plus-ellipsis shape: %q", a)
	}
	for _, c := range a[:12] {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("short() label %q is not lowercase hex", a)
		}
	}
	if short("fake-credential-2") == a {
		t.Fatal("short() collides across distinct credentials")
	}
}
