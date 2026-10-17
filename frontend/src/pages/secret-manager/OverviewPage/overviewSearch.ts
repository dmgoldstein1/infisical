import { z } from "zod";

export const SecretOverviewPageQuerySchema = z.object({
  search: z.string().catch(""),
  clearSearch: z.boolean().optional(),
  secretPath: z.string().catch("/"),
  connectionId: z.string().optional(),
  connectionName: z.string().optional(),
  environments: z.array(z.string()).catch([]),
  dynamicSecretId: z.string().optional(),
  honeyTokenId: z.string().optional(),
  tags: z.string().optional(),
  filterBy: z.string().optional()
});

export type SecretOverviewPageSearch = z.infer<typeof SecretOverviewPageQuerySchema>;
