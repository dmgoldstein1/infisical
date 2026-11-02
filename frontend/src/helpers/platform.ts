const INFISICAL_CLOUD_HOSTS = new Set([
  "app.infisical.com",
  "us.infisical.com",
  "eu.infisical.com",
  "gamma.infisical.com"
]);

export const isInfisicalCloud = () => {
  try {
    const { hostname, protocol } = new URL(window.location.origin);
    return protocol === "https:" && INFISICAL_CLOUD_HOSTS.has(hostname.toLowerCase());
  } catch {
    return false;
  }
};
