package cmd

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

// Pins the CodeQL go/reflected-xss disposition for relayStreamedBody: the SSE
// relay is an opaque byte pipe between the upstream service and the client.
// It must forward bytes exactly; HTML-escaping the relay would corrupt JSON
// event payloads and binary frames. This test FAILS if the relay transforms
// a single byte.
func TestRelayStreamedBodyPreservesBytesExactly(t *testing.T) {
	payload := "data: {\"secret\":\"<value>&\\\"quoted\\\"\"}\n\ndata: binary\x00\xff\xfe\n\n"
	rec := httptest.NewRecorder()
	relayStreamedBody(rec, strings.NewReader(payload))

	if got := rec.Body.String(); got != payload {
		t.Fatalf("relay transformed the stream: got %q, want %q", got, payload)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte{0x00, 0xff, 0xfe}) {
		t.Fatalf("relay dropped non-UTF8 bytes: %q", rec.Body.Bytes())
	}
}
