/* eslint-disable no-param-reassign */
import fp from "fastify-plugin";

import { getConfig } from "@app/lib/config/env";

import { DefaultResponseErrorsSchema } from "../routes/sanitizedSchemas";

const isScimRoutes = (pathname: string) =>
  pathname.startsWith("/api/v1/scim/Users") || pathname.startsWith("/api/v1/scim/Groups");

const isAcmeRoutes = (pathname: string) => pathname.startsWith("/api/v1/cert-manager/acme/");

const isOauthTokenRoute = (pathname: string) => pathname === "/api/v1/oauth/token";

export const registerResponseSchemaHooks = fp(async (server) => {
  const appCfg = getConfig();

  server.addHook("onRoute", (routeOptions) => {
    if (appCfg.isDevelopmentMode && routeOptions.schema) {
      routeOptions.schema.hide = false;
    }

    if (
      routeOptions.schema &&
      routeOptions.schema.response &&
      !isScimRoutes(routeOptions.path) &&
      !isAcmeRoutes(routeOptions.path) &&
      !isOauthTokenRoute(routeOptions.path)
    ) {
      routeOptions.schema.response = {
        ...DefaultResponseErrorsSchema,
        ...routeOptions.schema.response
      };
    }

    // Defense in depth for js/missing-rate-limiting: every route is already
    // capped by the global limiter (register in app.ts), but per-route config
    // is what route-level analysis sees. Backfill the global budget onto
    // routes that do not set their own limit so no handler is structurally
    // unlimited. /healthcheck and /api/status stay on the global allowlist;
    // ACME challenge traffic is a handful of requests per issuance and is
    // unaffected by a 600/min ceiling.
    if (!routeOptions.config?.rateLimit && routeOptions.path !== "/healthcheck" && routeOptions.path !== "/api/status") {
      routeOptions.config = {
        ...routeOptions.config,
        rateLimit: {
          timeWindow: 60 * 1000,
          max: 600,
          keyGenerator: (req) => req.realIp
        }
      };
    }
  });
});
