import Fastify from "fastify";
import { describe, expect, test } from "vitest";

import { writeLimit } from "@app/server/config/rateLimiter";
import { serializerCompiler, validatorCompiler, ZodTypeProvider } from "@app/server/plugins/fastify-zod";

import { registerAdminRouter } from "./admin-router";

// Pins the CodeQL js/missing-rate-limiting disposition for the admin router:
// every route in this file (including resend-invite, flagged at
// admin-router.ts:1223-1227) declares a framework-recognized
// `config: { rateLimit }` limit, enforced globally by @fastify/rate-limit
// (see server/app.ts). The analyzer does not model that config-based
// mechanism and flags one route out of 32 pattern-identical siblings.
// This test FAILS if any admin route loses its limit, documenting the keep.
describe("admin router rate limits", () => {
  test("every admin route declares a framework rate limit", async () => {
    const app = Fastify({ logger: false }).withTypeProvider<ZodTypeProvider>();
    app.setValidatorCompiler(validatorCompiler);
    app.setSerializerCompiler(serializerCompiler);

    const seen: Array<{ method: string; url: string; rateLimit: unknown }> = [];
    app.addHook("onRoute", (routeOptions) => {
      seen.push({
        method: String(routeOptions.method),
        url: routeOptions.url,
        rateLimit: (routeOptions.config as { rateLimit?: unknown } | undefined)?.rateLimit
      });
    });

    await app.register(registerAdminRouter);
    await app.ready();

    expect(seen.length).toBeGreaterThan(0);
    const unprotected = seen.filter((r) => r.rateLimit == null);
    expect(unprotected).toEqual([]);

    const resendInvite = seen.find((r) => r.url.includes("resend-invite"));
    expect(resendInvite).toBeDefined();
    expect(resendInvite?.rateLimit).toBe(writeLimit);

    await app.close();
  });
});
