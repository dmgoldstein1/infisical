import crypto from "crypto";
import { Controller, useForm } from "react-hook-form";
import { faArrowUpRightFromSquare, faBookOpen } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { PageTitle } from "@app/components/PageTitle";
import {
  Badge,
  Button,
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  Field,
  FieldError,
  FieldLabel,
  Input
} from "@app/components/v3";
import { useProject } from "@app/context";
import { localStorageService } from "@app/helpers/localStorage";
import { useGetCloudIntegrations } from "@app/hooks/api";

import { createIntegrationMissingEnvVarsNotification } from "../../IntegrationsListPage/IntegrationsListPage.utils";

const schema = z.object({
  gitLabURL: z.string()
});

type FormData = z.infer<typeof schema>;

export const GitlabAuthorizePage = () => {
  const { control, handleSubmit } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      gitLabURL: ""
    }
  });

  const { data: cloudIntegrations } = useGetCloudIntegrations();
  const { currentProject } = useProject();

  const onFormSubmit = ({ gitLabURL }: FormData) => {
    if (!cloudIntegrations) return;
    const integrationOption = cloudIntegrations.find(
      (integration) => integration.slug === "gitlab"
    );

    if (!integrationOption) return;

    if (!integrationOption.clientId) {
      createIntegrationMissingEnvVarsNotification(integrationOption.slug, "cicd");
      return;
    }

    const rawGitLabUrl = (gitLabURL as string).trim();
    let baseURL = "https://gitlab.com";
    if (rawGitLabUrl !== "") {
      let parsed: URL;
      try {
        parsed = new URL(rawGitLabUrl);
      } catch {
        return;
      }
      if (parsed.protocol !== "https:" && parsed.protocol !== "http:") return;
      baseURL = parsed.origin;
    }

    const csrfToken = crypto.randomBytes(16).toString("hex");
    localStorage.setItem("latestCSRFToken", csrfToken);
    localStorageService.setIntegrationProjectId(currentProject.id);

    const stateSuffix = rawGitLabUrl === "" ? "" : rawGitLabUrl;
    const state = `${csrfToken}|${stateSuffix}`;
    localStorageService.setIntegrationProjectId(currentProject.id);
    const callbackUri = `${window.location.origin}/integrations/gitlab/oauth2/callback`;
    const link =
      `${baseURL}/oauth/authorize` +
      `?client_id=${encodeURIComponent(integrationOption.clientId)}` +
      `&redirect_uri=${encodeURIComponent(callbackUri)}` +
      `&response_type=code` +
      `&state=${encodeURIComponent(state)}`;

    window.location.assign(link);
  };

  return (
    <div className="flex h-full w-full items-center justify-center">
      <PageTitle title="Authorize GitLab Integration" />
      <Card className="mb-12 w-full max-w-lg">
        <CardHeader>
          <CardTitle>
            <img src="/images/integrations/GitLab.png" height={28} width={28} alt="Gitlab logo" />
            <span>GitLab Integration</span>
          </CardTitle>
          <CardDescription>
            Authorize this integration to sync secrets from Infisical to GitLab. If no self-hosted
            GitLab URL is specified, then Infisical will connect you to GitLab Cloud.
          </CardDescription>
          <CardAction>
            <Badge variant="info" asChild>
              <a
                href="https://infisical.com/docs/integrations/cicd/gitlab"
                target="_blank"
                rel="noopener noreferrer"
              >
                <FontAwesomeIcon icon={faBookOpen} />
                Docs
                <FontAwesomeIcon icon={faArrowUpRightFromSquare} />
              </a>
            </Badge>
          </CardAction>
        </CardHeader>
        <form onSubmit={handleSubmit(onFormSubmit)} className="flex flex-col gap-5">
          <CardContent>
            <Controller
              control={control}
              name="gitLabURL"
              render={({ field, fieldState: { error } }) => (
                <Field className="mb-4" data-invalid={Boolean(error)}>
                  <FieldLabel htmlFor="gitlab-gitlaburl">Self-hosted URL (optional)</FieldLabel>
                  <Input
                    {...field}
                    placeholder="https://self-hosted-gitlab.com"
                    id="gitlab-gitlaburl"
                    isError={Boolean(error)}
                    aria-describedby={error ? "gitlab-gitlaburl-error" : undefined}
                  />
                  {Boolean(error) && (
                    <FieldError id="gitlab-gitlaburl-error">{error?.message}</FieldError>
                  )}
                </Field>
              )}
            />
          </CardContent>
          <CardFooter className="justify-end">
            <Button size="sm" type="submit" variant="project">
              Continue with OAuth
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
};
