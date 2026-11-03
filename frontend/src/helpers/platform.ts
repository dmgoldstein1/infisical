const CLOUD_HOSTS = new Set([
  "app.infisical.com",
  "us.infisical.com",
  "eu.infisical.com",
  "gamma.infisical.com"
]);

export const isInfisicalCloud = () => {
  try {
    const url = new URL(window.location.origin);
    return url.protocol === "https:" && CLOUD_HOSTS.has(url.hostname.toLowerCase());
  } catch {
    return false;
  }
};
