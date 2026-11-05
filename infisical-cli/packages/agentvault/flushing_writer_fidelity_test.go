package agentvault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for flushingWriter.Write: it is
// a transparent byte copy toward the downstream client. HTML-escaping at this
// sink would corrupt proxied SSE/binary streams; this test FAILS if output
// encoding is ever applied here.
func TestFlushingWriterPreservesResponseBytes(t *testing.T) {
	t.Parallel()

	payload := []byte("data: <script>alert(1)</script>\n\ndata: {\"a\":\"b&c\"}\n\n")
	rec := httptest.NewRecorder()
	fw := &flushingWriter{ResponseWriter: rec, rc: http.NewResponseController(rec)}

	n, err := fw.Write(payload)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Write returned %d, want %d", n, len(payload))
	}
	if rec.Body.String() != string(payload) {
		t.Fatalf("proxied bytes mutated: got %q want %q", rec.Body.String(), payload)
	}
}
