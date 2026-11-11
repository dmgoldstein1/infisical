package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Pins the CodeQL go/reflected-xss disposition for timeoutWriter.Write: the
// middleware is a byte-transparent passthrough — any content transformation
// here (e.g. HTML-escaping) corrupts every proxied API response. The timeout
// path's own error body is JSON + nosniff (set explicitly above); the
// passthrough must stay untouched. This test FAILS if Write ever transforms
// bytes, documenting the keep.
func TestTimeoutWriterIsByteTransparent(t *testing.T) {
	payloads := [][]byte{
		[]byte(`{"data":{"secret":"s3cr3t"}}`),
		[]byte("<html><script>alert(1)</script></html>"),
		{0x00, 0x01, 0xff, 0xfe, 0x22, 0x3c, 0x3e, 0x26},
	}
	for _, payload := range payloads {
		rec := httptest.NewRecorder()
		tw := &timeoutWriter{ResponseWriter: rec}
		n, err := tw.Write(payload)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(payload) {
			t.Fatalf("short write: %d of %d", n, len(payload))
		}
		if got := rec.Body.Bytes(); string(got) != string(payload) {
			t.Fatalf("passthrough transformed bytes: got %q want %q", got, payload)
		}
	}
}

// The timeout path's own error response stays JSON + non-sniffable.
func TestTimeoutErrorIsJSONNosniff(t *testing.T) {
	h := Timeout(20 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if nosniff := rec.Header().Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", nosniff)
	}
}
