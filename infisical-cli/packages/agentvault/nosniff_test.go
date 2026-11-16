package agentvault

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// CodeQL go/reflected-xss pin: the proxy relays upstream bytes transparently,
// so it must refuse MIME-sniffing on the client side unless the upstream set
// its own policy. Bodies must stay byte-identical.
func TestDispatchSetsNosniffAndPreservesBody(t *testing.T) {
	payload := "<html><script>alert(document.cookie)</script></html>"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, payload)
	}))
	t.Cleanup(upstream.Close)

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	host := "127.0.0.1:" + upstreamURL.Port()

	ps := &proxyServer{transport: newUpstreamTransport()}
	ps.setConfig(ProxyConfig{TrafficPolicy: TrafficPolicyAnyHost})
	ps.cache = newSessionCache(fixedResolver{services: []*resolvedService{policyService(host, nil, nil, nil, nil)}}, ps.pollInterval)

	front := httptest.NewServer(http.HandlerFunc(ps.dispatch))
	t.Cleanup(front.Close)

	proxyURL, _ := url.Parse(front.URL)
	proxyURL.User = url.UserPassword(ProxyAuthUsername, "agv_tok")
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}

	resp, err := client.Get("http://" + host + "/xss")
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
