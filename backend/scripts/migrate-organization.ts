/* eslint-disable */
import promptSync from "prompt-sync";
import { execFileSync } from "child_process";
import path from "path";
import { existsSync, writeFileSync } from "fs";

const prompt = promptSync({
  sigint: true
});

// Raw operator input: never interpolated into a shell command. Values reach
// pg_dump/pg_restore/psql via argv arrays and the PG* environment, never a shell.
const ask = (question: string, fallback?: string) => {
  const answer = (prompt(question) ?? "").trim();
  if (!answer && fallback !== undefined) return fallback;
  return answer;
};

const assertPort = (port: string) => {
  if (!/^\d{1,5}$/.test(port)) throw new Error("Invalid port: must be numeric");
};

const assertUuid = (id: string) => {
  // UUID charset ([0-9a-f-]) carries no shell or SQL metacharacters, so the
  // validated value below is safe to embed in a fixed SQL template.
  if (!/^[0-9a-f-]{36}$/i.test(id)) throw new Error("Organization ID must be a UUID");
};

const pgEnv = (host: string, port: string, user: string, password: string, database: string) => ({
  ...process.env,
  PGHOST: host,
  PGPORT: port,
  PGUSER: user,
  PGPASSWORD: password,
  PGDATABASE: database
});

const exportDb = () => {
  const exportHost = ask("Enter your Postgres Host to migrate from: ");
  const exportPort = ask("Enter your Postgres Port to migrate from [Default = 5432]: ", "5432");
  const exportUser = ask("Enter your Postgres User to migrate from: [Default = infisical]: ", "infisical");
  const exportPassword = ask("Enter your Postgres Password to migrate from: ");
  const exportDatabase = ask("Enter your Postgres Database to migrate from [Default = infisical]: ", "infisical");
  assertPort(exportPort);

  const dump = execFileSync(
    "pg_dump",
    ["-Fc", "infisical", "--exclude-table-data=secret_sharing", "--exclude-table-data=audit_log*"],
    {
      stdio: ["inherit", "pipe", "inherit"],
      env: pgEnv(exportHost, exportPort, exportUser, exportPassword, exportDatabase),
      maxBuffer: 1024 * 1024 * 4096
    }
  );
  writeFileSync(path.join(__dirname, "../src/db/backup.dump"), dump);
};

const importDbForOrg = () => {
  const importHost = ask("Enter your Postgres Host to migrate to: ");
  const importPort = ask("Enter your Postgres Port to migrate to [Default = 5432]: ", "5432");
  const importUser = ask("Enter your Postgres User to migrate to: [Default = infisical]: ", "infisical");
  const importPassword = ask("Enter your Postgres Password to migrate to: ");
  const importDatabase = ask("Enter your Postgres Database to migrate to [Default = infisical]: ", "infisical");
  const orgId = ask("Enter the organization ID to migrate: ");
  assertPort(importPort);
  assertUuid(orgId);

  if (!existsSync(path.join(__dirname, "../src/db/backup.dump"))) {
    console.log("File not found, please export the database first.");
    return;
  }

  const env = pgEnv(importHost, importPort, importUser, importPassword, importDatabase);
  const dumpPath = path.join(__dirname, "../src/db/backup.dump");
  execFileSync("pg_restore", ["-d", importDatabase, "--verbose", dumpPath], {
    maxBuffer: 1024 * 1024 * 4096,
    stdio: "inherit",
    env
  });

  // Fixed SQL templates; the only variable (orgId) is UUID-validated above.
  const statements = [
    `DELETE FROM public.organizations WHERE id != '${orgId}'`,
    'DELETE FROM users WHERE users.id NOT IN (SELECT org_memberships."userId" FROM org_memberships)',
    'DELETE FROM identities WHERE id NOT IN (SELECT "identityId" FROM identity_org_memberships)',
    'UPDATE super_admin SET "encryptedSlackClientId" = null, "encryptedSlackClientSecret" = null'
  ];
  for (const statement of statements) {
    execFileSync("psql", ["-c", statement], { stdio: "inherit", env });
  }

  console.log("Organization migrated successfully.");
};

const main = () => {
  const action = prompt(
    "Enter the action to perform\n 1. Export from existing instance.\n 2. Import org to instance.\n \n Action: "
  );
  if (action === "1") {
    exportDb();
  } else if (action === "2") {
    importDbForOrg();
  } else {
    console.log("Invalid action");
  }
};

main();
