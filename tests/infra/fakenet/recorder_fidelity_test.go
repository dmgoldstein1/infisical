package fakenet

import (
	"net/http/httptest"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for recorder.Write: it forwards
// handler bytes untouched for the call log. The test-only fakenet server never
// serves browsers, and HTML-escaping here would corrupt recorded bodies; this
// test FAILS if output encoding is ever applied here.
func TestRecorderPreservesBodyBytes(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"message":"<script>alert(1)</script>"}`)
	rec := httptest.NewRecorder()
	r := &recorder{ResponseWriter: rec}

	n, err := r.Write(payload)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Write returned %d, want %d", n, len(payload))
	}
	if rec.Body.String() != string(payload) {
		t.Fatalf("recorded bytes mutated: got %q want %q", rec.Body.String(), payload)
	}
}
