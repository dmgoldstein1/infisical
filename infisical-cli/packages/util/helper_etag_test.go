package util

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Infisical/infisical-merge/packages/models"
)

func etagSecrets() []models.SingleEnvironmentVariable {
	return []models.SingleEnvironmentVariable{
		{Key: "B", Value: "2"},
		{Key: "A", Value: "1"},
	}
}

// Pins the CodeQL go/weak-sensitive-data-hashing remediation: the ETag is an
// opaque quoted-hex change detector that is deterministic within the process,
// order-insensitive, sensitive to content, and NOT a plain SHA-256 of the
// secret material (keyed MAC under a process-lifetime key).
func TestGenerateETagFromSecretsIsKeyed(t *testing.T) {
	a := GenerateETagFromSecrets(etagSecrets())
	if a != GenerateETagFromSecrets(etagSecrets()) {
		t.Fatal("ETag is not deterministic within the process")
	}
	if !strings.HasPrefix(a, `"`) || !strings.HasSuffix(a, `"`) || len(a) != 66 {
		t.Fatalf("ETag %q lost its quoted-hex shape", a)
	}
	if GenerateETagFromSecrets(etagSecrets()) == GenerateETagFromSecrets([]models.SingleEnvironmentVariable{{Key: "A", Value: "changed"}}) {
		t.Fatal("ETag is insensitive to secret content")
	}
	// Order-insensitivity (SortSecretsByKeys) keeps hot-reload quiet.
	shuffled := []models.SingleEnvironmentVariable{{Key: "A", Value: "1"}, {Key: "B", Value: "2"}}
	if GenerateETagFromSecrets(shuffled) != a {
		t.Fatal("ETag depends on input order")
	}
	// Must not be the plain SHA-256 the old implementation emitted.
	content := []byte("A1B2")
	sum := sha256.Sum256(content)
	if a == `"`+hex.EncodeToString(sum[:])+`"` {
		t.Fatal("ETag is a plain SHA-256 of the secret material")
	}
}
