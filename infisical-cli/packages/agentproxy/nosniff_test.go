package agentproxy

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// CodeQL go/reflected-xss pin: the proxy relays upstream bytes transparently,
// so it must refuse MIME-sniffing on the client side unless the upstream set
// its own policy. Bodies must stay byte-identical.
func TestPlainForwardSetsNosniffAndPreservesBody(t *testing.T) {
	payload := "<html><script>alert(document.cookie)</script></html>"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, payload)
	}))
	defer upstream.Close()

	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	hostname := u.Hostname()

	jwt := "test.jwt.token"
	scope := agentScope{projectID: "proj", environment: "prod", secretPath: "/"}
	services := []*resolvedService{{
		name:         "internal",
		hostPatterns: parseHostPatterns(hostname),
		isEnabled:    true,
	}}
	client := newTestProxy(t, UnmatchedAllow, jwt, scope, services)
	reader := bufio.NewReader(client)

	_, err = fmt.Fprintf(client, "GET http://%s/xss HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: %s\r\n\r\n",
		u.Host, u.Host, proxyAuthHeader("proj", "prod", "/", jwt))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != payload {
		t.Fatalf("body not byte-identical: got %q", body)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html" {
		t.Fatalf("upstream Content-Type not preserved, got %q", got)
	}
}

// An upstream that sets its own framing policy must win over the default.
func TestPlainForwardPreservesUpstreamFramingPolicy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "custom-policy")
		_, _ = io.WriteString(w, "ok")
	}))
	defer upstream.Close()

	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	hostname := u.Hostname()

	jwt := "test.jwt.token"
	scope := agentScope{projectID: "proj", environment: "prod", secretPath: "/"}
	services := []*resolvedService{{
		name:         "internal",
		hostPatterns: parseHostPatterns(hostname),
		isEnabled:    true,
	}}
	client := newTestProxy(t, UnmatchedAllow, jwt, scope, services)
	reader := bufio.NewReader(client)

	_, err = fmt.Fprintf(client, "GET http://%s/policy HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: %s\r\n\r\n",
		u.Host, u.Host, proxyAuthHeader("proj", "prod", "/", jwt))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "custom-policy" {
		t.Fatalf("upstream policy overwritten, got %q", got)
	}
}
