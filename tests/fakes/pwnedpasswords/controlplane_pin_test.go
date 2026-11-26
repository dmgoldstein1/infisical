package pwnedpasswords

import "testing"

// Pins the CodeQL go/weak-sensitive-data-hashing disposition: the HIBP
// k-anonymity API mandates SHA-1 (5-char prefix lookup, suffix match). Any
// other hash seeds buckets the backend (SHA-1, breached-password.ts) never
// queries. FAILS if the bucket hash ever changes, documenting the keep.
func TestHIBPBucketIsSHA1(t *testing.T) {
	t.Parallel()
	// SHA1("password") = 5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8
	if got := hibpBucket("password"); got != "5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8" {
		t.Fatalf("hibpBucket = %q, want the SHA-1 HIBP vector", got)
	}
	if got := hibpBucket("password"); len(got) != 40 {
		t.Fatalf("hibpBucket length = %d, want 40 hex chars", len(got))
	}
}
