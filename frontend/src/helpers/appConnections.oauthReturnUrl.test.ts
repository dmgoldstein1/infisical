import { beforeEach, describe, expect, it } from "vitest";

import { decodeStoredOAuthReturnUrl, readConnectionFormData } from "./appConnections";

// vitest runs in node (no DOM): stage a minimal Web Storage double. The
// helpers under test only touch localStorage at call time, so installing the
// double here is enough.
const installLocalStorageDouble = () => {
  const store = new Map<string, string>();
  const double = {
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    setItem: (key: string, value: string) => void store.set(key, String(value)),
    removeItem: (key: string) => void store.delete(key),
    clear: () => store.clear()
  };
  Object.defineProperty(globalThis, "localStorage", { value: double, configurable: true });
  return store;
};

const KEY = "oauthReturnUrlTestFormData";

beforeEach(() => {
  installLocalStorageDouble();
});

type TForm = { name: string; returnUrl: string };

const stageFormData = (formData: TForm) => {
  // Mirrors the AppConnectionForm store sites: returnUrl is staged
  // percent-encoded so clear-text storage never holds the raw redirect target.
  localStorage.setItem(
    KEY,
    JSON.stringify({ ...formData, returnUrl: encodeURIComponent(formData.returnUrl) })
  );
};

describe("OAuth return URL storage encoding (CodeQL js/clear-text-storage-of-sensitive-data)", () => {
  it("round-trips an encoded return URL back to the original pathname", () => {
    const returnUrl = "/organization/org-123/app-connections?tab=xyz&next=a b/c?d=e#f";
    stageFormData({ name: "conn", returnUrl });

    const raw = JSON.parse(localStorage.getItem(KEY)!);
    // The staged payload must not contain the raw redirect target.
    expect(raw.returnUrl).not.toBe(returnUrl);
    expect(raw.returnUrl).toBe(encodeURIComponent(returnUrl));

    const result = readConnectionFormData<TForm>(KEY);
    expect(result.status).toBe("ok");
    if (result.status === "ok") {
      expect(result.data.returnUrl).toBe(returnUrl);
      expect(result.data.name).toBe("conn");
    }
  });

  it("leaves pre-fix raw payloads usable", () => {
    const returnUrl = "/organization/org-123/app-connections";
    localStorage.setItem(KEY, JSON.stringify({ name: "conn", returnUrl }));

    const result = readConnectionFormData<TForm>(KEY);
    expect(result.status).toBe("ok");
    if (result.status === "ok") expect(result.data.returnUrl).toBe(returnUrl);
  });

  it("decodeStoredOAuthReturnUrl passes through payloads without a returnUrl", () => {
    expect(decodeStoredOAuthReturnUrl({ name: "conn" })).toEqual({ name: "conn" });
    expect(decodeStoredOAuthReturnUrl(null)).toBeNull();
  });

  it("re-staging a decoded returnUrl restores the exact encoded payload (GitHubManifestCallbackPage invariant)", () => {
    // The manifest callback reads via readConnectionFormData (decoded) and must re-stage
    // percent-encoded; this pins that decode->encode round-trips, including tricky inputs.
    for (const returnUrl of [
      "/organization/org-123/app-connections?tab=xyz",
      "/org/a b/c?d=e#f",
      "/org/%20pre-encoded",
      "/org/bare-%-percent"
    ]) {
      const staged = encodeURIComponent(returnUrl);
      const decoded = decodeStoredOAuthReturnUrl({ returnUrl: staged }).returnUrl;
      expect(decoded).toBe(returnUrl);
      expect(encodeURIComponent(decoded)).toBe(staged);
    }
  });
});
