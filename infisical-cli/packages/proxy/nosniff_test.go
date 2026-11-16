package proxy

import (
	"net/http"
	"testing"
)

// CodeQL go/reflected-xss pin for the transparent proxy relays.
func TestSetNosniffIfAbsent(t *testing.T) {
	empty := http.Header{}
	SetNosniffIfAbsent(empty)
	if got := empty.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff default, got %q", got)
	}

	custom := http.Header{"X-Content-Type-Options": []string{"custom-policy"}}
	SetNosniffIfAbsent(custom)
	if got := custom.Get("X-Content-Type-Options"); got != "custom-policy" {
		t.Fatalf("upstream policy overwritten, got %q", got)
	}
	if vals := custom["X-Content-Type-Options"]; len(vals) != 1 {
		t.Fatalf("expected single header value, got %v", vals)
	}
}
