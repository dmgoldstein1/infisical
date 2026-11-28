package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Pins the CodeQL go/insecure-hostkeycallback disposition for
// connectToTargetServer: the PAM SSH proxy dials operator-configured targets
// for which no known_hosts registry exists (see the TODO at the callback),
// so an ephemeral test-server key must connect. A strict callback breaks
// every proxied SSH login. This test FAILS under that change, documenting the
// keep.
func startEphemeralSSHServer(t *testing.T) string {
	t.Helper()

	_, hostKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatal(err)
	}
	config := &ssh.ServerConfig{
		PasswordCallback: func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) { return nil, nil },
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_, chans, reqs, err := ssh.NewServerConn(conn, config)
			if err != nil {
				_ = conn.Close()
				continue
			}
			go ssh.DiscardRequests(reqs)
			go func() {
				for ch := range chans {
					_ = ch.Reject(ssh.UnknownChannelType, "no channels in this test")
				}
			}()
		}
	}()

	return listener.Addr().String()
}

func TestConnectToTargetServerAcceptsOperatorConfiguredHost(t *testing.T) {
	target := startEphemeralSSHServer(t)

	p := &SSHProxy{config: SSHProxyConfig{
		SessionID:      "test-session",
		TargetAddr:     target,
		AuthMethod:     "password",
		InjectUsername: "operator",
		InjectPassword: "operator-password",
	}}

	client, err := p.connectToTargetServer()
	if err != nil {
		t.Fatalf("connectToTargetServer with ephemeral host key: %v", err)
	}
	_ = client.Close()
}
