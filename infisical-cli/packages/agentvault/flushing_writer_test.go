package agentvault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for flushingWriter.Write: the
// writer is an opaque byte relay between the upstream service and the client
// (SSE streams, JSON, binary). It must forward bytes exactly; HTML-escaping
// the relay would corrupt non-HTML bodies and double-escape HTML ones. Any
// escaping belongs where untrusted data is interpolated into markup, never
// in this pipe. This test FAILS if Write transforms a single byte.
func TestFlushingWriterPreservesBytesExactly(t *testing.T) {
	payload := []byte("<div>&\"'broken\x00\xff\xfe\"'</div>")
	rec := httptest.NewRecorder()
	fw := &flushingWriter{
		ResponseWriter: rec,
		rc:             http.NewResponseController(rec),
	}

	n, err := fw.Write(payload)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Write wrote %d bytes, want %d", n, len(payload))
	}
	if got := rec.Body.Bytes(); string(got) != string(payload) {
		t.Fatalf("Write transformed the relay: got %q, want %q", got, payload)
	}
}
