package gcp

import "testing"

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
func TestRedactedLogHeaders(t *testing.T) {
	h := map[string][]string{
		"Authorization":       {"Bearer secret-token"},
		"Proxy-Authorization": {"Basic c2VjcmV0"},
		"Cookie":              {"session=abc"},
		"Content-Type":        {"application/json"},
	}
	got := redactedLogHeaders(h)
	for _, k := range []string{"Authorization", "Proxy-Authorization", "Cookie"} {
		for _, v := range got[k] {
			if v != "[redacted]" {
				t.Errorf("header %q not redacted: %q", k, v)
			}
		}
	}
	if got["Content-Type"][0] != "application/json" {
		t.Errorf("non-sensitive header was altered: %q", got["Content-Type"])
	}
	// The original must be untouched so forwarding still carries credentials.
	if h["Authorization"][0] != "Bearer secret-token" {
		t.Errorf("original headers mutated")
	}
}
