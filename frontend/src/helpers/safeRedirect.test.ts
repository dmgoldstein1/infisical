import { describe, expect, it } from "vitest";

import {
  isAzureVaultUrl,
  isSlackWebhookUrl,
  sanitizeAppSlug,
  sanitizeAzureTenantId,
  sanitizeCustomHost,
  sanitizeHttpBaseUrl
} from "./safeRedirect";

describe("sanitizeAzureTenantId", () => {
  it("passes through guids, domains and aliases", () => {
    expect(sanitizeAzureTenantId("common")).toBe("common");
    expect(sanitizeAzureTenantId("b7c8d9e0-f1a2-4b3c-8d5e-6f7a8b9c0d1e")).toBe(
      "b7c8d9e0-f1a2-4b3c-8d5e-6f7a8b9c0d1e"
    );
    expect(sanitizeAzureTenantId("contoso.onmicrosoft.com")).toBe("contoso.onmicrosoft.com");
  });

  it("defaults blank input to common", () => {
    expect(sanitizeAzureTenantId("")).toBe("common");
    expect(sanitizeAzureTenantId(undefined)).toBe("common");
  });

  it("rejects url-breaking input", () => {
    expect(() => sanitizeAzureTenantId("tenant/evil?x=1")).toThrow();
    expect(() => sanitizeAzureTenantId("a b")).toThrow();
  });
});

describe("sanitizeCustomHost", () => {
  it("passes through plain hostnames", () => {
    expect(sanitizeCustomHost("ghes.example.com")).toBe("ghes.example.com");
    expect(sanitizeCustomHost("")).toBeUndefined();
  });

  it("rejects schemes, paths and userinfo", () => {
    expect(() => sanitizeCustomHost("https://evil.com")).toThrow();
    expect(() => sanitizeCustomHost("host/path")).toThrow();
    expect(() => sanitizeCustomHost("user@host")).toThrow();
  });
});

describe("sanitizeAppSlug", () => {
  it("passes through slugs and rejects path separators", () => {
    expect(sanitizeAppSlug("my-app")).toBe("my-app");
    expect(() => sanitizeAppSlug("../evil")).toThrow();
  });
});

describe("sanitizeHttpBaseUrl", () => {
  it("keeps https self-hosted urls and falls back when blank", () => {
    expect(sanitizeHttpBaseUrl("https://gitlab.example.com", "https://gitlab.com")).toBe(
      "https://gitlab.example.com"
    );
    expect(sanitizeHttpBaseUrl("", "https://gitlab.com")).toBe("https://gitlab.com");
  });

  it("rejects non-http schemes", () => {
    expect(() => sanitizeHttpBaseUrl(`java${"script"}:alert(1)`, "https://gitlab.com")).toThrow();
  });
});

describe("isSlackWebhookUrl", () => {
  it("accepts exact host and rejects lookalikes", () => {
    expect(isSlackWebhookUrl("https://hooks.slack.com/services/T/B/X")).toBe(true);
    expect(isSlackWebhookUrl("https://hooks.slack.com.evil.com/services")).toBe(false);
    expect(isSlackWebhookUrl("https://evil.com/?hooks.slack.com")).toBe(false);
  });
});

describe("isAzureVaultUrl", () => {
  it("accepts vault uris and rejects lookalikes", () => {
    expect(isAzureVaultUrl("https://myvault.vault.azure.net")).toBe(true);
    expect(isAzureVaultUrl("https://myvault.vault.azure.net.evil.com")).toBe(false);
    expect(isAzureVaultUrl("http://myvault.vault.azure.net")).toBe(false);
    expect(isAzureVaultUrl("https://vault.azure.net")).toBe(false);
  });
});
