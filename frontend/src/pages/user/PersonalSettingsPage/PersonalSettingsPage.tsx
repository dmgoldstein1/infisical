import { useTranslation } from "react-i18next";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { ArrowLeftIcon } from "lucide-react";

import { WishForm } from "@app/components/features/WishForm";
import { PageTitle } from "@app/components/PageTitle";

import { PersonalTabGroup } from "./components/PersonalTabGroup";

const isCloudHost = () => {
  try {
    const { hostname, protocol } = new URL(window.location.origin);
    const host = hostname.toLowerCase();
    return (
      protocol === "https:" && (host === "app.infisical.com" || host === "gamma.infisical.com")
    );
  } catch {
    return false;
  }
};

export const PersonalSettingsPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate({ from: "/personal-settings/" });
  const { selectedTab } = useSearch({
    from: "/_authenticate/personal-settings/_layout/"
  });

  return (
    <div className="mx-auto w-full max-w-5xl p-4 lg:p-8">
      <PageTitle title={t("settings.personal.title")} />
      <link rel="icon" href="/infisical.ico" />
      <Link
        to="/"
        className="mb-6 inline-flex items-center gap-2 rounded-sm text-sm text-muted transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        <ArrowLeftIcon className="size-4" />
        Back to organization
      </Link>
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="grid gap-1">
          <h1 className="font-alliance text-2xl font-normal text-foreground">Personal Settings</h1>
          <p className="text-sm text-muted">
            Manage your profile, authentication, and active sessions.
          </p>
        </div>
        {!isCloudHost() && <WishForm />}
      </header>
      <PersonalTabGroup
        selectedTab={selectedTab}
        onTabChange={(nextTab) =>
          navigate({
            search: (previous) => ({ ...previous, selectedTab: nextTab })
          })
        }
      />
    </div>
  );
};
