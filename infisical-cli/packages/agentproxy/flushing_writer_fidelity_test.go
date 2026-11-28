package agentproxy

import (
	"net/http/httptest"
	"testing"
)

func TestFlushingWriterPreservesBodyBytes(t *testing.T) {
	payload := []byte("<script>alert(1)</script>\x00\xff binary \xe2\x98\x83")
	rec := httptest.NewRecorder()
	fw := flushingWriter{w: rec}

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
