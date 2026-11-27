package cmd

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamUpstreamBodyPreservesBytesExactly(t *testing.T) {
	payload := "data: {\"secret\":\"<script>alert(1)</script>\"}\n\n" + strings.Repeat("x\x00\xff", 500)
	rec := httptest.NewRecorder()
	streamUpstreamBody(rec, io.NopCloser(strings.NewReader(payload)))
	if got := rec.Body.String(); got != payload {
		t.Fatalf("stream relay mutated bytes: got %q want %q", got, payload)
	}
}
