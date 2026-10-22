// Package provision boots a throwaway Infisical backend on the test runner,
// creates the exact fixtures the CLI ./test suite asserts, and exports the
// CLI_TESTS_* environment (plus INFISICAL_DOMAIN) for a later
// `go test ./...` step to consume.
//
// It deliberately reuses the e2e compose harness and generated API client
// instead of growing a second backend-boot path: one way to boot the backend,
// one provision flow. Run it as:
//
//	TESTCONTAINERS_RYUK_DISABLED=true \
//	INFISICAL_BACKEND_DIR=/path/to/infisical/backend \
//	INFISICAL_BACKEND_DOCKERFILE=Dockerfile.dev \
//	PROVISION_ENV_OUT=$GITHUB_ENV \
//	go test -count=1 -run TestProvisionCLIFixtures ./provision/
//
// Ryuk must stay disabled here so the stack survives this test process;
// TestTeardownCLIFixtures removes it afterwards. Nothing in this package is
// an enterprise feature, so no license key is required.
//
// The backend keeps the harness's random host port: the ./test snapshots mask
// the instance domain (see FilterRequestID), so a fixed port would only buy a
// bind conflict on occupied machines.
package provision

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/infisical/cli/e2e-tests/packages/client"
	"github.com/infisical/cli/e2e-tests/packages/infisical"
	"github.com/infisical/cli/e2e-tests/util"
)

const (
	envSlug        = "dev"
	stagingEnvSlug = "staging"
	folderName     = "test-folder"
)

// fixtureSecrets are the exact (name, value) pairs the CLI ./test snapshots
// assert. They mirror the historical Cloud fixtures so the snapshots keep
// testing the same contract against the throwaway backend.
var fixtureSecrets = []struct {
	key         string
	value       string
	environment string
	secretPath  string
}{
	{key: "TEST-SECRET-1", value: "test-value-1", environment: envSlug, secretPath: "/"},
	{key: "TEST-SECRET-2", value: "test-value-2", environment: envSlug, secretPath: "/"},
	{key: "TEST-SECRET-3", value: "test-value-3", environment: envSlug, secretPath: "/"},
	{key: "FOLDER-SECRET-1", value: "folder-value-1", environment: envSlug, secretPath: "/" + folderName},
	{key: "STAGING-SECRET-1", value: "staging-value-1", environment: stagingEnvSlug, secretPath: "/"},
	{key: "STAGING-SECRET-2", value: "staging-value-2", environment: stagingEnvSlug, secretPath: "/"},
}

// newFixtureStack mirrors util.NewInfisicalService's default (non-ACME) stack.
// The definition must stay byte-identical between the provision and teardown
// tests: the compose project name is a hash of it, and the teardown relies on
// that to find the running stack again.
func newFixtureStack() *util.InfisicalService {
	backendOpts := infisical.BackendOptionsFromEnv()
	backendOpts.Dockerfile = "Dockerfile.dev"
	return &util.InfisicalService{Stack: infisical.NewStack(
		infisical.WithDbService(),
		infisical.WithRedisService(),
		infisical.WithBackendService(backendOpts),
	)}
}

// postJSON performs an authenticated raw API call for endpoints the generated
// client does not cover (folders, imports, project memberships, service
// tokens). The UpdateMachineIdentityRole helper in e2e/util follows the same
// pattern.
func postJSON(t *testing.T, ctx context.Context, apiURL, token, path string, body any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+path, bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST %s: %s", path, string(respBody))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(respBody, &decoded))
	return decoded
}

func randomHex(t *testing.T, nBytes int) string {
	t.Helper()
	b := make([]byte, nBytes)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}

// TestProvisionCLIFixtures boots the stack, resets it to a clean database,
// provisions every fixture, and exports the CLI_TESTS_* environment. It leaves
// the stack running (Ryuk disabled); TestTeardownCLIFixtures removes it.
func TestProvisionCLIFixtures(t *testing.T) {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != "true" {
		t.Fatal("TESTCONTAINERS_RYUK_DISABLED=true is required: Ryuk would delete the stack when this test process exits, before the CLI suite runs")
	}

	ctx := t.Context()
	svc := newFixtureStack()
	require.NoError(t, svc.Stack.Up(ctx))

	// A reused stack from an earlier crashed run may hold stale fixtures.
	svc.Reset(ctx, t)
	svc.Bootstrap(ctx, t)

	c := svc.ApiClient()
	provisioned := svc.ProvisionResult()
	apiURL, err := svc.Stack.ApiUrl(ctx)
	require.NoError(t, err)
	t.Logf("Backend API URL: %s", apiURL)

	// 1. Project (default envs: dev/staging/prod).
	projectType := client.SecretManager
	defaultEnvs := true
	projectName := "cli-ci-tests"
	projectDesc := "Throwaway project for the self-hosted CLI test suite"
	projectResp, err := c.CreateProjectWithResponse(ctx, client.CreateProjectJSONRequestBody{
		ProjectName:             projectName,
		ProjectDescription:      &projectDesc,
		Type:                    &projectType,
		ShouldCreateDefaultEnvs: &defaultEnvs,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, projectResp.StatusCode(), "create project: %s", string(projectResp.Body))
	require.NotNil(t, projectResp.JSON200)
	projectID := projectResp.JSON200.Project.Id
	t.Logf("Project ID: %s", projectID)

	// 2. Folder for the recursive-listing fixture.
	postJSON(t, ctx, apiURL, provisioned.Token, "/api/v1/folders", map[string]any{
		"workspaceId": projectID,
		"environment": envSlug,
		"name":        folderName,
	})

	// 3. Secrets (shared type, matching the snapshots).
	sharedType := client.CreateSecretV4JSONBodyTypeShared
	for _, s := range fixtureSecrets {
		secretPath := s.secretPath
		resp, err := c.CreateSecretV4WithResponse(ctx, s.key, client.CreateSecretV4JSONRequestBody{
			ProjectId:   projectID,
			Environment: s.environment,
			SecretValue: s.value,
			SecretPath:  &secretPath,
			Type:        &sharedType,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), "create secret %s: %s", s.key, string(resp.Body))
	}

	// 4. Import staging into dev, mirroring the Cloud fixtures' WithImports pairs.
	postJSON(t, ctx, apiURL, provisioned.Token, "/api/v1/secret-imports", map[string]any{
		"workspaceId": projectID,
		"environment": envSlug,
		"path":        "/",
		"import": map[string]any{
			"environment": stagingEnvSlug,
			"path":        "/",
		},
	})

	// 5. Machine identity with universal auth, granted admin on the project.
	// The role is not under test (the token auth path is); admin removes an
	// RBAC variable from what is meant to be a CLI-behavior suite.
	identity := svc.CreateMachineIdentity(t, ctx, util.WithUniversalAuth())
	require.NotNil(t, identity.UniversalAuthClientId)
	require.NotNil(t, identity.UniversalAuthClientSecret)
	postJSON(t, ctx, apiURL, provisioned.Token,
		fmt.Sprintf("/api/v1/projects/%s/memberships/identities/%s", projectID, identity.Id),
		map[string]any{
			"roles": []map[string]any{{"role": "admin"}},
		})

	// 6. Legacy service token, created exactly the way the UI does
	// (empty key envelope; the v4 reads used by the CLI are server-decrypted).
	// The CLI requires a 4-segment token, so a random 4th segment is appended;
	// the server only verifies the first three. Scopes use /** globs: a bare
	// "/" scope only matches the root path, which would hide the folder and
	// imported fixtures from the recursive/import assertions.
	tokenResp := postJSON(t, ctx, apiURL, provisioned.Token, "/api/v2/service-token/", map[string]any{
		"name":        "cli-ci-tests",
		"workspaceId": projectID,
		"scopes": []map[string]any{
			{"environment": envSlug, "secretPath": "/**"},
			{"environment": stagingEnvSlug, "secretPath": "/**"},
		},
		"encryptedKey": "",
		"iv":           "",
		"tag":          "",
		"expiresIn":    nil,
		"permissions":  []string{"read"},
	})
	baseToken, ok := tokenResp["serviceToken"].(string)
	require.True(t, ok && baseToken != "", "service token response missing serviceToken: %v", tokenResp)
	serviceToken := baseToken + "." + randomHex(t, 16)

	// 7. Export everything the CLI ./test suite consumes. The domain carries
	// the harness's random mapped port; the suite masks it in snapshots.
	domain := apiURL
	envOut := map[string]string{
		"INFISICAL_DOMAIN":           domain,
		"CLI_TESTS_UA_CLIENT_ID":     *identity.UniversalAuthClientId,
		"CLI_TESTS_UA_CLIENT_SECRET": *identity.UniversalAuthClientSecret,
		"CLI_TESTS_SERVICE_TOKEN":    serviceToken,
		"CLI_TESTS_PROJECT_ID":       projectID,
		"CLI_TESTS_ENV_SLUG":         envSlug,
		"CLI_TESTS_USER_EMAIL":       provisioned.Email,
		"CLI_TESTS_USER_PASSWORD":    provisioned.Password,
	}
	var lines []string
	for k, v := range envOut {
		require.NotEmpty(t, v, "provisioned value for %s is empty", k)
		lines = append(lines, k+"="+v+"\n")
	}
	outPath := os.Getenv("PROVISION_ENV_OUT")
	if outPath == "" {
		for _, l := range lines {
			fmt.Print(l)
		}
		t.Log("PROVISION_ENV_OUT not set; printed environment instead of writing it")
		return
	}
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)
	defer f.Close()
	for _, l := range lines {
		_, err := f.WriteString(l)
		require.NoError(t, err)
	}
	t.Logf("Wrote %d variables to %s; backend at %s", len(lines), outPath, domain)
}

// TestTeardownCLIFixtures rebuilds the identical stack definition (same hash,
// same project name) and force-removes it with volumes.
func TestTeardownCLIFixtures(t *testing.T) {
	ctx := t.Context()
	svc := newFixtureStack()
	require.NoError(t, svc.Stack.Up(ctx))
	require.NoError(t, svc.Stack.DownWithForce(ctx, true))
}
