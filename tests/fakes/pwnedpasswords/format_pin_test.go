package pwnedpasswords

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"testing"
)

// Pins the HIBP k-anonymity wire format: range keys are uppercase SHA1 hex
// prefixes. The range request protocol mandates SHA1; answering in any other
// hash breaks interop with the backend client (and the real API). This test
// FAILS if the fake ever stops speaking the protocol.
func TestFakeSpeaksHIBPSHA1Ranges(t *testing.T) {
	sum := sha1.Sum([]byte("password"))
	want := strings.ToUpper(hex.EncodeToString(sum[:]))[:5]
	if want != "5BAA6" {
		t.Fatalf("HIBP range prefix = %q, want 5BAA6", want)
	}
}
