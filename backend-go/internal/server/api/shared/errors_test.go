package shared

import "testing"

func TestRedactCredentialsForLog(t *testing.T) {
	in := `field Authorization: Bearer abcDEF123._~+/= rejected`
	want := `field Authorization: Bearer [REDACTED] rejected`
	if got := redactCredentialsForLog(in); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	plain := "connection refused"
	if got := redactCredentialsForLog(plain); got != plain {
		t.Errorf("plain message changed: %q", got)
	}
}
