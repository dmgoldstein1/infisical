package aws

import (
	"bufio"
	"io"
	"net/http"
	"testing"
)

// Pins the CodeQL go/request-forgery disposition for handleRequest: a plain
// (untunnelled) proxy request may only reach AWS API endpoints. targetHost
// enforces the suffix allowlist (IsAWSHost) plus port restrictions, so a
// request naming any other host is rejected before any upstream dial. This
// test FAILS if that gate is removed, documenting the keep.
func TestProxyRefusesPlainRequestToNonAWSEndpoint(t *testing.T) {
	server, _ := fakeAWS(t)
	conn, _, _ := startProxy(t, server)

	if _, err := io.WriteString(conn, "GET https://evil.com/ HTTP/1.1\r\nHost: evil.com\r\nConnection: close\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("plain-proxy request to a non-AWS host was forwarded")
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d with the not-an-AWS-endpoint refusal", resp.StatusCode, http.StatusBadRequest)
	}
}
