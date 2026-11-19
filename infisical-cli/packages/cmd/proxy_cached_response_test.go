package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteCachedResponseSetsNosniffAndPreservesBody(t *testing.T) {
	body := "<html>upstream</html>"
	cached := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	rec := httptest.NewRecorder()
	writeCachedResponse(rec, cached)

	res := rec.Result()
	defer res.Body.Close()
	if got := res.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	gotBody, _ := io.ReadAll(res.Body)
	if string(gotBody) != body {
		t.Fatalf("body = %q, want byte-identical relay", gotBody)
	}
}

func TestWriteCachedResponseKeepsUpstreamFramingPolicy(t *testing.T) {
	cached := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"X-Content-Type-Options": []string{"custom"}},
		Body:       io.NopCloser(strings.NewReader("x")),
	}

	rec := httptest.NewRecorder()
	writeCachedResponse(rec, cached)
	res := rec.Result()
	defer res.Body.Close()
	if got := res.Header.Get("X-Content-Type-Options"); got != "custom" {
		t.Fatalf("X-Content-Type-Options = %q, want upstream value preserved", got)
	}
}
