import { createFileRoute, linkOptions, stripSearchParams } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";

import { OverviewPage } from "./OverviewPage";
import { SecretOverviewPageQuerySchema } from "./overviewSearch";

export const Route = createFileRoute(
  "/_authenticate/_inject-org-details/_org-layout/organizations/$orgId/projects/secret-management/$projectId/_secret-manager-layout/overview"
)({
  component: OverviewPage,
  validateSearch: zodValidator(SecretOverviewPageQuerySchema),
  search: {
    middlewares: [
      stripSearchParams({
        secretPath: "/",
        search: "",
        clearSearch: undefined,
        environments: [],
        tags: undefined,
        filterBy: undefined
      })
    ]
  },
  beforeLoad: ({
    context,
    params
  }): { breadcrumbs: { label: string; link: ReturnType<typeof linkOptions> }[] } => ({
    breadcrumbs: [
      ...context.breadcrumbs,
      {
        label: "Secrets",
        link: linkOptions({
          to: "/organizations/$orgId/projects/secret-management/$projectId/overview",
          params
        })
      }
    ]
  })
});
