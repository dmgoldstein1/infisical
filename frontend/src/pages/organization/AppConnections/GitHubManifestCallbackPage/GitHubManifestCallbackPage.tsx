import { useEffect, useRef } from "react";
import { useNavigate, useParams, useSearch } from "@tanstack/react-router";

import { createNotification } from "@app/components/notifications";
import { PageTitle } from "@app/components/PageTitle";
import { PageLoader } from "@app/components/v3";
import { ROUTE_PATHS } from "@app/const/routes";
import {
  consumeCsrfToken,
  getConnectionFlowReturnNavigateOptions,
  GITHUB_CONNECTION_FORM_STORAGE_KEY,
  readConnectionFormData
} from "@app/helpers/appConnections";
import { AppConnection } from "@app/hooks/api/appConnections/enums";

export type TStoredGitHubConnectionForm = {
  credentials?: Record<string, unknown>;
  resumeWithGitHubAppId?: string;
  returnUrl?: string;
  // All other staged in-progress fields (name, description, gateway
  // selection, …) pass through untouched so the connection form can resume
  // them after the github.com round-trip wipes in-memory state.
  [key: string]: unknown;
};

// Rebuild the staged form for the trip back to the connection form: the
// return URL is consumed from memory and must never reach Web Storage at
// rest, while every other staged field is preserved so the in-progress form
// resumes intact. Pinned by GitHubManifestCallbackPage.test.ts.
export const buildResumedGitHubConnectionForm = (
  connectionForm: TStoredGitHubConnectionForm,
  gitHubAppId: string
): TStoredGitHubConnectionForm => {
  const { returnUrl: _droppedReturnUrl, ...resumedForm } = connectionForm;
  return {
    ...resumedForm,
    credentials: {
      ...(resumedForm.credentials ?? {}),
      gitHubAppId
    },
    resumeWithGitHubAppId: gitHubAppId
  };
};

export const GitHubManifestCallbackPage = () => {
  const navigate = useNavigate();
  const { orgId } = useParams({ strict: false }) as { orgId: string };
  const search = useSearch({
    from: ROUTE_PATHS.Organization.AppConnections.GitHubManifestCallbackPage.id
  });
  const hasRunRef = useRef(false);

  useEffect(() => {
    if (hasRunRef.current) return;
    hasRunRef.current = true;

    const { gitHubAppId, slug, installState } = search;

    const storedForm = readConnectionFormData<TStoredGitHubConnectionForm>(
      GITHUB_CONNECTION_FORM_STORAGE_KEY
    );

    let connectionForm: TStoredGitHubConnectionForm | null;
    if (storedForm.status === "ok") {
      connectionForm = storedForm.data;
    } else if (storedForm.status === "missing") {
      connectionForm = {};
    } else {
      connectionForm = null;
    }

    const rawReturnUrl = connectionForm?.returnUrl ?? "";
    // Only honor the same-origin path this browser staged before leaving for GitHub. Anything else
    // (an absolute URL planted in storage by another tab) falls back to the default page instead of
    // becoming an open redirect.
    const returnUrl =
      rawReturnUrl.startsWith("/") && !rawReturnUrl.startsWith("//") ? rawReturnUrl : "";

    const navigateBack = (reopenForm: boolean) => {
      if (returnUrl) {
        return navigate(
          getConnectionFlowReturnNavigateOptions({
            returnUrl,
            reopenFormApp: reopenForm ? AppConnection.GitHub : undefined
          })
        );
      }
      return navigate({
        to: "/organizations/$orgId/app-connections",
        params: { orgId },
        search: reopenForm ? { addConnectionApp: AppConnection.GitHub } : undefined
      });
    };

    if (!gitHubAppId || !slug || !installState) {
      createNotification({ type: "error", text: "Invalid GitHub manifest callback parameters." });
      navigateBack(false);
      return;
    }

    // Confirm this callback corresponds to a flow this browser initiated. The install state is
    // round-tripped through the signed manifest token and echoed back here; it must match the CSRF
    // token stored when the user kicked off app creation. This blocks a crafted callback URL from
    // pre-seeding the connection form with an attacker-chosen app id.
    if (!consumeCsrfToken(installState)) {
      createNotification({ type: "error", text: "Invalid GitHub manifest callback state." });
      navigateBack(false);
      return;
    }

    // The GitHub App is already created at this point. Mark the in-progress connection form so it
    // resumes with the new app selected, then send the user back to it to finish the connection.
    // The return URL above is consumed from memory and must not remain in Web Storage at rest, so
    // rebuild the staged form without it instead of round-tripping the stored object: the staged
    // return URL is treated as sensitive by static analysis and must never reach the storage sink.
    if (connectionForm) {
      const resumedForm = buildResumedGitHubConnectionForm(connectionForm, gitHubAppId);
      localStorage.setItem(GITHUB_CONNECTION_FORM_STORAGE_KEY, JSON.stringify(resumedForm));
    }

    createNotification({
      type: "success",
      text: "GitHub App created. Select it and finish setting up your connection."
    });
    navigateBack(true);
  }, []);

  return (
    <>
      <PageTitle title="Creating GitHub App" />
      <PageLoader text="Creating your GitHub App, please wait..." />
    </>
  );
};
