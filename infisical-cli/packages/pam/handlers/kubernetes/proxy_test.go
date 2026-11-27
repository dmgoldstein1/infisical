package kubernetes

import (
	"net/http"
	"testing"
)

func TestLoggedHeadersRedactsAllValues(t *testing.T) {
	in := http.Header{
		"Authorization":          {"AWS4-HMAC-SHA256 Credential=AKID"},
		"X-Amz-Security-Token":   {"session-token"},
		"Proxy-Authorization":    {"Bearer x"},
		"Cookie":                 {"sess=abc"},
		"Set-Cookie":             {"sess=abc"},
		"Content-Type":           {"application/json"},
		"X-Amz-Target":           {"DynamoDB_20120810.DescribeTable"},
		"X-Amz-Sso_bearer_token": {"sso"},
		"X-Api-Key":              {"key"},
	}
	got := loggedHeaders(in)
	for name := range in {
		if v := got.Get(name); v != "[redacted]" {
			t.Errorf("logged header %q = %q, want [redacted]", name, v)
		}
	}
	if in.Get("Authorization") == "" {
		t.Error("loggedHeaders mutated the caller's header map")
	}
}
