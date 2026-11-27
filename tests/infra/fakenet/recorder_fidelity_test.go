package fakenet

import (
	"net/http/httptest"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for recorder.Write: the
// recorder is an opaque byte relay that captures the status for the call log
// without buffering the body. It must forward bytes exactly; HTML-escaping
// the relay would corrupt fake API responses under test. This test FAILS if
// Write transforms a single byte.
func TestRecorderPreservesBytesExactly(t *testing.T) {
	payload := []byte("<div>&\"'broken\x00\xff\xfe\"'</div>")
	rec := &recorder{ResponseWriter: httptest.NewRecorder()}

	n, err := rec.Write(payload)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Write wrote %d bytes, want %d", n, len(payload))
	}
	got := rec.ResponseWriter.(*httptest.ResponseRecorder).Body.Bytes()
	if string(got) != string(payload) {
		t.Fatalf("Write transformed the relay: got %q, want %q", got, payload)
	}
}
