// Exact-host allowlist for Infisical Cloud. The previous implementation used
// `origin.includes("https://...")`, which an attacker-controlled origin such as
// `https://app.infisical.com.evil.example` would also satisfy. Hostnames are
// compared exactly after URL parsing so super/substring matches cannot pass.
const INFISICAL_CLOUD_HOSTS: ReadonlySet<string> = new Set([
  "app.infisical.com",
  "us.infisical.com",
  "eu.infisical.com",
  "gamma.infisical.com"
]);

export const isInfisicalCloud = () => {
  try {
    return INFISICAL_CLOUD_HOSTS.has(new URL(window.location.origin).hostname.toLowerCase());
  } catch {
    return false;
  }
};
