// Browser-native random hex tokens (WebCrypto). Replaces the Node
// `crypto.randomBytes` polyfill so the frontend no longer bundles
// node-stdlib-browser / crypto-browserify / elliptic.
export const randomHex = (bytes: number): string => {
  const arr = new Uint8Array(bytes);
  globalThis.crypto.getRandomValues(arr);
  return Array.from(arr, (b) => b.toString(16).padStart(2, "0")).join("");
};
