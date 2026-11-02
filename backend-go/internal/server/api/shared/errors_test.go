package shared

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/doordash-oss/oapi-codegen-dd/v3/pkg/runtime"
)

// Redaction must strip credential-shaped values while keeping the log useful.
func TestRedactSensitiveText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string // must be absent from output
	}{
		{"bearer", "header Authorization: Bearer abcDEF123._-~+/=", []string{"abcDEF123"}},
		{"basic", "auth Basic dXNlcjpwYXNz", []string{"dXNlcjpwYXNz"}},
		{"password kv", `login failed for password=SuperSecret123!`, []string{"SuperSecret123!"}},
		{"api key json", `{"api_key": "ak_live_456"}`, []string{"ak_live_456"}},
	}
	for _, tc := range cases {
		got := redactSensitiveText(tc.in)
		for _, secret := range tc.want {
			if strings.Contains(got, secret) {
				t.Errorf("%s: redactor leaked %q in %q", tc.name, secret, got)
			}
		}
		if !strings.Contains(got, "[REDACTED]") {
			t.Errorf("%s: redactor did not mark output: %q", tc.name, got)
		}
	}
}

// Validation-error logs must not echo rejected input values (which may be secrets).
func TestWriteValidationErrorDoesNotLogValues(t *testing.T) {
	var sb strings.Builder
	h := &ErrorHandler{Logger: slog.New(slog.NewTextHandler(&sb, nil))}
	verr := runtime.ValidationErrors{
		{Field: "password", Message: "too short", Err: errors.New("got SuperSecret123!")},
	}
	w := &captureWriter{}
	h.writeValidationError(context.Background(), w, "req-1", verr)
	out := sb.String()
	if strings.Contains(out, "SuperSecret123!") {
		t.Errorf("validation log leaked input value: %q", out)
	}
	if !strings.Contains(out, "password") {
		t.Errorf("validation log should keep the field name: %q", out)
	}
}

type captureWriter struct {
	header int
	body   strings.Builder
}

func (c *captureWriter) Header() http.Header         { return http.Header{} }
func (c *captureWriter) Write(b []byte) (int, error) { return c.body.Write(b) }
func (c *captureWriter) WriteHeader(statusCode int)  { c.header = statusCode }
