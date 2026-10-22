import { createFileRoute, linkOptions } from "@tanstack/react-router";

import { AccessManagementPage } from "./AccessManagementPage";

export const Route = createFileRoute(
  "/_authenticate/_inject-org-details/admin/_admin-layout/access-management"
)({
  component: AccessManagementPage,
  beforeLoad: async (): Promise<{
    breadcrumbs: { label: string; link: ReturnType<typeof linkOptions> }[];
  }> => {
    return {
      breadcrumbs: [
        {
          label: "Admin",
          link: linkOptions({ to: "/admin" })
        },
        {
          label: "Access Control",
          link: linkOptions({
            to: "/admin/access-management"
          })
        }
      ]
    };
  }
});
