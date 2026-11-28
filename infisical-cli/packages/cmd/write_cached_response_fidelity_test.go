package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteCachedResponsePreservesBodyBytes(t *testing.T) {
	payload := `<script>alert(1)</script>`
	cached := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/html"}},
		Body:       io.NopCloser(strings.NewReader(payload)),
	}
	rec := httptest.NewRecorder()
	writeCachedResponse(rec, cached)

	if rec.Body.String() != payload {
		t.Fatalf("cached bytes mutated: got %q want %q", rec.Body.String(), payload)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
