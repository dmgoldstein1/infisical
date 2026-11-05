import bcrypt from "bcrypt";

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

// Pins the CodeQL js/insufficient-password-hash disposition: operator-supplied
// bcrypt costs below 10 are clamped to the minimum in createHash, so stored
// hashes never fall to brute force on modern hardware. Stored hashes carry
// their own cost, so verification of older hashes is unaffected.
describe("password hashing (CodeQL js/insufficient-password-hash disposition)", () => {
  test("operator-supplied cost below 10 is clamped to at least 10", async () => {
    const hash = await crypto.hashing().createHash("s3cret", 4);
    expect(bcrypt.getRounds(hash)).toBeGreaterThanOrEqual(10);
    expect(await crypto.hashing().compareHash("s3cret", hash)).toBe(true);
  });

  test("default configured cost meets the minimum", async () => {
    const hash = await crypto.hashing().createHash("s3cret", 12);
    expect(bcrypt.getRounds(hash)).toBeGreaterThanOrEqual(10);
    expect(await crypto.hashing().compareHash("s3cret", hash)).toBe(true);
  });
});
