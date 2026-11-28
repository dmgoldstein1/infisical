import Fastify from "fastify";
import { describe, expect, test, vi } from "vitest";
import { z } from "zod";

vi.mock("@app/lib/config/env", () => ({
  getConfig: () => ({ isDevelopmentMode: false })
}));

import { serializerCompiler, validatorCompiler, ZodTypeProvider } from "@app/server/plugins/fastify-zod";

import { registerResponseSchemaHooks } from "./response-schema-hooks";

// Pins the CodeQL js/missing-rate-limiting disposition for
// response-schema-hooks.ts: this module is a schema-decoration plugin, not a
// route handler — it registers zero routes and performs no authorization, so
// there is no handler sink to rate-limit. These tests FAIL if the module ever
// starts serving requests, documenting the keep.
describe("response schema hooks", () => {
  test("registers no routes, only an onRoute schema hook", async () => {
    const app = Fastify({ logger: false });
    await app.register(registerResponseSchemaHooks);
    await app.ready();

    const tree = app.printRoutes();
    expect(tree).not.toContain("/api/");

    await app.close();
  });

  test("still decorates response schemas for subsequently registered routes", async () => {
    const app = Fastify({ logger: false }).withTypeProvider<ZodTypeProvider>();
    app.setValidatorCompiler(validatorCompiler);
    app.setSerializerCompiler(serializerCompiler);
    await app.register(registerResponseSchemaHooks);
    app.route({
      method: "GET",
      url: "/api/v1/things",
      schema: {
        response: {
          200: z.object({ id: z.string() })
        }
      },
      handler: async () => ({ id: "x" })
    });
    await app.ready();

    const res = await app.inject({ method: "GET", url: "/api/v1/things" });
    expect(res.statusCode).toBe(200);

    await app.close();
  });

  test("backfills a default rate limit onto routes that declare none", async () => {
    const app = Fastify({ logger: false }).withTypeProvider<ZodTypeProvider>();
    app.setValidatorCompiler(validatorCompiler);
    app.setSerializerCompiler(serializerCompiler);
    await app.register(registerResponseSchemaHooks);

    const seen: Array<{ url: string; rateLimit: unknown }> = [];
    app.addHook("onRoute", (routeOptions) => {
      seen.push({
        url: routeOptions.url,
        rateLimit: (routeOptions.config as { rateLimit?: unknown } | undefined)?.rateLimit
      });
    });

    app.route({
      method: "GET",
      url: "/api/v1/unlimited",
      handler: async () => ({ ok: true })
    });
    app.route({
      method: "GET",
      url: "/healthcheck",
      handler: async () => ({ ok: true })
    });
    await app.ready();

    const limited = seen.find((r) => r.url === "/api/v1/unlimited");
    expect(limited?.rateLimit).toMatchObject({ max: 600 });
    const exempt = seen.find((r) => r.url === "/healthcheck");
    expect(exempt?.rateLimit).toBeUndefined();

    await app.close();
  });
});
