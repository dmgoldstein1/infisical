package handlers

import "testing"

func TestPostgresMD5PasswordMatchesWireFormat(t *testing.T) {
	got := postgresMD5Password("secret", "user", [4]byte{1, 2, 3, 4})
	want := "md5fccef98e4f1cf6cbe96b743fad4e8bd0"
	if got != want {
		t.Fatalf("md5 wire value = %q, want %q", got, want)
	}
}
