import { describe, expect, it } from "vitest";

import {
  buildResumedGitHubConnectionForm,
  TStoredGitHubConnectionForm
} from "./GitHubManifestCallbackPage";

// Pins the CodeQL js/clear-text-storage-of-sensitive-data disposition for the
// manifest callback restage: the staged return URL must never reach Web
// Storage at rest, while every other in-progress field must survive so the
// connection form resumes intact after the github.com round-trip wipes
// in-memory state. These tests FAIL if the restage drops user fields
// (breaking resume) or leaks the return URL back into storage.
describe("buildResumedGitHubConnectionForm", () => {
  it("drops the return URL but preserves every other staged field", () => {
    const staged: TStoredGitHubConnectionForm = {
      method: "app",
      name: "my-conn",
      description: "typed before leaving",
      gatewayId: "gw-1",
      credentials: { instanceType: "server", host: "gh.example.com" },
      returnUrl: "/organizations/org-123/app-connections"
    };

    const resumed = buildResumedGitHubConnectionForm(staged, "app-42");

    expect(resumed).not.toHaveProperty("returnUrl");
    expect(resumed.method).toBe("app");
    expect(resumed.name).toBe("my-conn");
    expect(resumed.description).toBe("typed before leaving");
    expect(resumed.gatewayId).toBe("gw-1");
    expect(resumed.credentials).toMatchObject({
      instanceType: "server",
      host: "gh.example.com",
      gitHubAppId: "app-42"
    });
    expect(resumed.resumeWithGitHubAppId).toBe("app-42");
  });

  it("marks an empty staged form with the new app id", () => {
    const resumed = buildResumedGitHubConnectionForm({}, "app-42");
    expect(resumed).not.toHaveProperty("returnUrl");
    expect(resumed.resumeWithGitHubAppId).toBe("app-42");
    expect(resumed.credentials).toMatchObject({ gitHubAppId: "app-42" });
  });
});
