const INFISICAL_CLOUD_HOSTS = new Set([
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
