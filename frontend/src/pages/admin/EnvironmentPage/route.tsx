import { createFileRoute, linkOptions } from "@tanstack/react-router";

import { EnvironmentPage } from "./EnvironmentPage";

export const Route = createFileRoute(
  "/_authenticate/_inject-org-details/admin/_admin-layout/environment"
)({
  component: EnvironmentPage,
  beforeLoad: (): {
    breadcrumbs: { label: string; link: ReturnType<typeof linkOptions> }[];
  } => {
    return {
      breadcrumbs: [
        {
          label: "Admin",
          link: linkOptions({ to: "/admin" })
        },
        {
          label: "Environment",
          link: linkOptions({
            to: "/admin/environment"
          })
        }
      ]
    };
  }
});
