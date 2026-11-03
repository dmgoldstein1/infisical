package kubernetes

import "testing"

func TestRedactedLogHeaders(t *testing.T) {
	h := map[string][]string{
		"Authorization":    {"Bearer secret-token"},
		"Impersonate-User": {"system:admin"},
		"Content-Type":     {"application/json"},
	}
	got := redactedLogHeaders(h)
	for _, k := range []string{"Authorization", "Impersonate-User"} {
		for _, v := range got[k] {
			if v != "[redacted]" {
				t.Errorf("header %q not redacted: %q", k, v)
			}
		}
	}
	if got["Content-Type"][0] != "application/json" {
		t.Errorf("non-sensitive header was altered: %q", got["Content-Type"])
	}
	if h["Authorization"][0] != "Bearer secret-token" {
		t.Errorf("original headers mutated")
	}
}
