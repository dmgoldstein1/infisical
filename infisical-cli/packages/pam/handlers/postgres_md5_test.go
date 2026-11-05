package handlers

import (
	"strings"
	"testing"
)

// Pins the CodeQL go/weak-sensitive-data-hashing disposition: the MD5 auth
// response is dictated by the PostgreSQL wire protocol (the server sends
// AuthenticationMD5Password and only accepts md5(md5(password+user)+salt)).
// This test FAILS if the response hash is ever changed, documenting the keep.
func TestMD5PasswordResponseMatchesWireProtocol(t *testing.T) {
	t.Parallel()

	got := md5PasswordResponse("username", "password", [4]byte{1, 2, 3, 4})
	if got != "md57906291157541c1dec603135b36b2d10" {
		t.Fatalf("MD5 auth response = %q, want the protocol-mandated value", got)
	}
	if !strings.HasPrefix(got, "md5") || len(got) != 35 {
		t.Fatalf("MD5 auth response %q lost its wire shape", got)
	}
}
