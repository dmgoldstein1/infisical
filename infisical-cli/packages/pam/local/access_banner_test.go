package pam

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	f()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDatabaseBannerShowsTheLiteralPlaceholderPassword(t *testing.T) {
	out := captureStdout(t, func() {
		printDatabaseSessionInfo(AccountConnectionDisplay{
			TypeLabel:        "Oracle",
			RequiredPassword: "password",
		}, "", "acct", time.Hour, "scott", "orcl", 1521)
	})
	if !bytes.Contains([]byte(out), []byte("Type password literally")) {
		t.Fatalf("banner must tell the user what to type literally, got:\n%s", out)
	}
}
