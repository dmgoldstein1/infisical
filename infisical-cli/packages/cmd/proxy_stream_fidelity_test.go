package cmd

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for the SSE streaming sink:
// streamResponseBody is a transparent byte copy from the upstream response
// body. HTML-escaping at this sink would corrupt SSE framing/binary streams;
// this test FAILS if output encoding is ever applied here.
func TestStreamResponseBodyPreservesBytes(t *testing.T) {
	t.Parallel()

	payload := "data: <script>alert(1)</script>\n\ndata: {\"a\":\"b&c\"}\n\n"
	rec := httptest.NewRecorder()
	streamResponseBody(rec, strings.NewReader(payload))

	if rec.Body.String() != payload {
		t.Fatalf("streamed bytes mutated: got %q want %q", rec.Body.String(), payload)
	}
}
