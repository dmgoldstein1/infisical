// Guards for OAuth authorize redirects: tenant ids and custom hosts are
// user-controlled and flow into window.location assignments, so they are
// validated at the sink instead of interpolated raw.
const AZURE_TENANT_RE = /^[A-Za-z0-9._-]+$/;

export const sanitizeAzureTenantId = (value: unknown): string => {
  const tenant = typeof value === "string" ? value.trim() : "";
  if (tenant === "") return "common";
  if (tenant.length > 256 || !AZURE_TENANT_RE.test(tenant)) {
    throw new Error("Invalid tenant ID");
  }
  return tenant;
};

// Hostnames for self-hosted instances (e.g. GHES). Rejects schemes,
// userinfo, ports, paths, and whitespace so the value cannot break out of
// the https:// URL under construction.
const CUSTOM_HOST_RE = /^[A-Za-z0-9]([A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$/;

export const sanitizeCustomHost = (value: unknown): string | undefined => {
  if (value === undefined || value === null) return undefined;
  const host = String(value).trim().toLowerCase();
  if (host === "") return undefined;
  if (!CUSTOM_HOST_RE.test(host)) {
    throw new Error("Invalid host");
  }
  return host;
};

// GitHub App slugs come from the server-side app list, but the install URL
// is built client-side, so constrain the value at the sink.
const APP_SLUG_RE = /^[A-Za-z0-9._-]+$/;

export const sanitizeAppSlug = (value: unknown): string => {
  const slug = typeof value === "string" ? value.trim() : "";
  if (slug === "" || slug.length > 256 || !APP_SLUG_RE.test(slug)) {
    throw new Error("Invalid app identifier");
  }
  return slug;
};

// Allow only http(s) URLs for user-supplied base URLs (self-hosted GitLab).
// http is kept for on-prem deployments without TLS; anything else
// (javascript:, data:, ...) is rejected.
export const sanitizeHttpBaseUrl = (value: unknown, fallback: string): string => {
  const raw = typeof value === "string" && value.trim() !== "" ? value.trim() : fallback;
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error("Invalid URL");
  }
  if (parsed.protocol !== "https:" && parsed.protocol !== "http:") {
    throw new Error("Invalid URL scheme");
  }
  if (parsed.username !== "" || parsed.password !== "") {
    throw new Error("Invalid URL");
  }
  return parsed.toString().replace(/\/$/, "");
};

// Slack incoming-webhook URLs live exactly on hooks.slack.com over https.
export const isSlackWebhookUrl = (value: string): boolean => {
  try {
    const parsed = new URL(value.trim());
    return parsed.protocol === "https:" && parsed.hostname.toLowerCase() === "hooks.slack.com";
  } catch {
    return false;
  }
};

// Azure vault URIs look like https://<vault>.vault.azure.net (no path,
// no credentials, standard port).
export const isAzureVaultUrl = (value: string): boolean => {
  try {
    const parsed = new URL(value.trim());
    if (parsed.protocol !== "https:") return false;
    if (parsed.username !== "" || parsed.password !== "" || parsed.port !== "") return false;
    const host = parsed.hostname.toLowerCase();
    return host.endsWith(".vault.azure.net") && host.length > ".vault.azure.net".length;
  } catch {
    return false;
  }
};
