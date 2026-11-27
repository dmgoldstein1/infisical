package gcp

import (
	"net/http"
	"testing"
)

func TestIsGCPHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"googleapis.com", true},
		{"storage.googleapis.com", true},
		{"compute.googleapis.com", true},
		{"oauth2.googleapis.com:443", true},
		{"storage.googleapis.com:443", true},

		{"evil.com", false},
		{"evil.com#.googleapis.com", false},
		{"evil.com?.googleapis.com", false},
		{"evil.com/.googleapis.com", false},
		{"evil.com\\.googleapis.com", false},
		{"user@googleapis.com", false},
		{"user:pass@googleapis.com", false},
		{"notgoogleapis.com", false},
		{"evil-googleapis.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isGCPHost(tt.host); got != tt.want {
				t.Errorf("isGCPHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

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
