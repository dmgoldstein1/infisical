package azure

import (
	"net/http"
	"testing"
)

func TestLoggedHeadersDropsCredentials(t *testing.T) {
	in := http.Header{
		"Authorization":       {"Bearer real-token"},
		"Proxy-Authorization": {"Bearer x"},
		"Cookie":              {"sess=abc"},
		"Set-Cookie":          {"sess=abc"},
		"X-Api-Key":           {"key"},
		"Content-Type":        {"application/json"},
	}
	got := loggedHeaders(in)
	for _, name := range credentialHeaders {
		if v := got.Get(name); v != "" {
			t.Errorf("logged header %q = %q, want redacted", name, v)
		}
	}
	if got.Get("Content-Type") != "application/json" {
		t.Errorf("non-credential headers lost: %v", got)
	}
	if in.Get("Authorization") == "" {
		t.Error("loggedHeaders mutated the caller's header map")
	}
}
