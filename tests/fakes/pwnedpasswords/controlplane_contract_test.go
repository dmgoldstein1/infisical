package pwnedpasswords

import (
	"testing"
)

// Pins the CodeQL go/weak-sensitive-data-hashing disposition for the HIBP
// fake: the HaveIBeenPwned range API is keyed by uppercase SHA-1 (k-anonymity
// contract shared with the real service and the backend client). Any other
// hash orients the fake's prefix/suffix split away from every real lookup.
// This test FAILS if the fake ever stops hashing with SHA-1, documenting the
// keep. Vectors: sha1("password") =
// 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8 (the HIBP docs' own example).
func TestPwnedPasswordHashMatchesHibpContract(t *testing.T) {
	t.Parallel()

	prefix, suffix := pwnedPasswordHash("password")
	if prefix != "5BAA6" {
		t.Fatalf("range prefix = %q, want HIBP SHA-1 prefix %q", prefix, "5BAA6")
	}
	if suffix != "1E4C9B93F3F0682250B6CF8331B7EE68FD8" {
		t.Fatalf("range suffix = %q, want HIBP SHA-1 suffix", suffix)
	}
	if prefix2, _ := pwnedPasswordHash("password"); prefix2 != prefix {
		t.Fatalf("hash not deterministic: %q vs %q", prefix, prefix2)
	}
}
