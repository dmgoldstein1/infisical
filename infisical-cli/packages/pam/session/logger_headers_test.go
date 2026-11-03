package session

import (
	"net/http"
	"testing"
)

func TestRedactSensitiveHeaders(t *testing.T) {
	headers := http.Header{
		"Authorization":        {"Bearer abc123"},
		"Content-Type":         {"application/json"},
		"Cookie":               {"session=xyz"},
		"X-Custom-Secret-Data": {"hello"},
	}

	redacted := redactSensitiveHeaders(headers)

	if got := redacted.Get("Authorization"); got != "[REDACTED]" {
		t.Errorf("Authorization = %q, want redacted", got)
	}
	if got := redacted.Get("Cookie"); got != "[REDACTED]" {
		t.Errorf("Cookie = %q, want redacted", got)
	}
	if got := redacted.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want untouched", got)
	}
	if got := redacted.Get("X-Custom-Secret-Data"); got != "hello" {
		t.Errorf("non-sensitive header = %q, want untouched", got)
	}
	// The caller's map must not be mutated: proxies reuse req.Header.
	if got := headers.Get("Authorization"); got != "Bearer abc123" {
		t.Errorf("input headers mutated: Authorization = %q", got)
	}
}
