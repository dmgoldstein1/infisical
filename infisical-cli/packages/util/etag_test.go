package util

import (
	"testing"

	"github.com/Infisical/infisical-merge/packages/models"
)

func TestGenerateETagFromSecretsIsStable(t *testing.T) {
	secrets := []models.SingleEnvironmentVariable{
		{Key: "B", Value: "2"},
		{Key: "A", Value: "1"},
	}
	first := GenerateETagFromSecrets(secrets)
	if first == "" || first == `""` {
		t.Fatalf("etag must be non-empty, got %q", first)
	}
	reordered := []models.SingleEnvironmentVariable{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "2"},
	}
	if got := GenerateETagFromSecrets(reordered); got != first {
		t.Fatalf("etag must be order-independent, got %q want %q", got, first)
	}
	changed := []models.SingleEnvironmentVariable{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "changed"},
	}
	if got := GenerateETagFromSecrets(changed); got == first {
		t.Fatalf("etag must change when a secret value changes, got %q", got)
	}
}
