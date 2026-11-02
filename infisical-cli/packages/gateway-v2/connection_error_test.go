package gatewayv2

import (
	"encoding/json"
	"strings"
	"testing"
)

// CodeQL go/unsafe-quoting: the message is attacker-influenced; it must be
// JSON-encoded so a double quote cannot break out of the string.
func TestBuildHttpInternalServerErrorQuotesMessage(t *testing.T) {
	out := buildHttpInternalServerError(`bad "quoted" message`)
	body := out[strings.Index(out, "\r\n\r\n")+4:]
	var decoded map[string]string
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("error body is not valid JSON: %v (%q)", err, body)
	}
	if !strings.Contains(decoded["message"], `bad "quoted" message`) {
		t.Fatalf("message was mangled: %q", decoded["message"])
	}
}
