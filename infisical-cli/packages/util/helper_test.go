package util

import (
	"testing"

	"github.com/Infisical/infisical-merge/packages/models"
)

func TestGenerateETagFromSecretsIsDeterministic(t *testing.T) {
	secrets := []models.SingleEnvironmentVariable{
		{Key: "B", Value: "2"},
		{Key: "A", Value: "1"},
	}
	a, b := GenerateETagFromSecrets(secrets), GenerateETagFromSecrets(secrets)
	if a != b || a == "" {
		t.Fatalf("ETag not deterministic: %q vs %q", a, b)
	}
}
