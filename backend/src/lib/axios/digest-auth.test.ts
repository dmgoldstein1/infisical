import { createHash } from "node:crypto";

import { createDigestAuthRequestInterceptor } from "@app/lib/axios/digest-auth";
import { crypto } from "@app/lib/crypto";

vi.mock("@app/lib/logger", () => ({
  logger: { info: vi.fn(), warn: vi.fn(), error: vi.fn(), debug: vi.fn() },
  sanitizeUrlForLog: (url: string) => url
}));

beforeAll(async () => {
  process.env.FIPS_ENABLED = "false";
  await crypto.initialize({} as never, {} as never, {} as never);
});

afterAll(() => {
  delete process.env.FIPS_ENABLED;
});

// The api-tests runner image enforces OpenSSL FIPS at the provider level,
// where MD5 is unavailable even when the app license flag above is off. The
// conformance test below can only execute where the runtime offers MD5, so it
// is skipped (not failed) there — matching the primitive's own FIPS block.
const isRuntimeMd5Available = (() => {
  try {
    createHash("md5").update("probe").digest("hex");
    return true;
  } catch {
    return false;
  }
})();

// Pins the CodeQL js/insufficient-password-hash + js/weak-cryptographic-algorithm
// disposition for crypto.ts `md5` (sole caller: digest-auth.ts): HTTP Digest
// access authentication (RFC 2617 / RFC 2069) mandates MD5 at the protocol
// level — HA1/HA2/response are md5(...) by spec and the server rejects any
// other hash. This test recomputes the expected response with an independent
// MD5 and fails if the construction ever stops being RFC 2617 MD5,
// documenting the keep.
describe("HTTP digest auth (CodeQL md5 disposition)", () => {
  test.skipIf(!isRuntimeMd5Available)("answers a 401 challenge with an RFC 2617 MD5 digest", async () => {
    const username = "Mufasa";
    const password = "Circle Of Life";
    const realm = "testrealm@host.com";
    const nonce = "dcd98b7102dd2f0e8b11d0f600bfb0c093";
    const path = "/dir/index.html";

    const seen: unknown[] = [];
    const axiosInstance = {
      request: vi.fn(async (opts: unknown) => {
        seen.push(opts);
        if (seen.length === 1) {
          const err = {
            response: {
              status: 401,
              headers: { "www-authenticate": `Digest realm="${realm}", nonce="${nonce}", qop="auth"` }
            },
            config: { url: path, method: "GET", headers: {} }
          };
          // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
          return await Promise.reject(err);
        }
        return { data: "ok" };
      })
    };

    const intercept = createDigestAuthRequestInterceptor(axiosInstance as never, username, password);
    const result = await intercept({ url: path, method: "GET", headers: {} } as never);
    expect(result).toEqual({ data: "ok" });
    expect(axiosInstance.request).toHaveBeenCalledTimes(2);

    const retried = (seen[1] as { headers: Record<string, string> }).headers;
    const authorization = retried.authorization ?? retried.Authorization;
    expect(authorization).toContain('algorithm="MD5"');
    expect(authorization).toContain(`username="${username}"`);

    const response = authorization.match(/response="([0-9a-f]{32})"/)?.[1];
    expect(response).toBeDefined();

    // Independently recompute the RFC 2617 expected response from the
    // challenge parameters echoed back in the header.
    const cnonce = authorization.match(/cnonce="([^"]+)"/)?.[1] as string;
    const nc = authorization.match(/nc=([0-9a-f]{8})/)?.[1] as string;
    const md5 = (s: string) => createHash("md5").update(s).digest("hex");
    const ha1 = md5(`${username}:${realm}:${password}`);
    const ha2 = md5(`GET:${path}`);
    expect(response).toBe(md5(`${ha1}:${nonce}:${nc}:${cnonce}:auth:${ha2}`));
  });

  test("non-digest 401s are rejected without hashing", async () => {
    const axiosInstance = {
      request: vi.fn(async () => {
        // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
        return await Promise.reject({
          response: { status: 401, headers: {} },
          message: "basic?"
        });
      })
    };
    const intercept = createDigestAuthRequestInterceptor(axiosInstance as never, "u", "p");
    await expect(intercept({ url: "/x", method: "GET" } as never)).rejects.toBe("basic?");
    expect(axiosInstance.request).toHaveBeenCalledTimes(1);
  });
});
