/* eslint-disable */
import { execSync } from "child_process";
import { readdirSync } from "fs";
import path from "path";
import promptSync from "prompt-sync";

const prompt = promptSync({ sigint: true });

const migrationName = prompt("Enter name for seedfile: ");
const fileCounter = readdirSync(path.join(__dirname, "../src/db/seeds")).length || 1;
// Allowlist the operator-supplied name: it is interpolated into a shell command below.
const safeMigrationName = migrationName.replace(/[^A-Za-z0-9-_]/g, "").slice(0, 64) || "seed";
execSync(
  `npx knex seed:make --knexfile ${path.join(__dirname, "../src/db/knexfile.ts")} -x ts ${
    fileCounter + 1
  }-${safeMigrationName}`,
  { stdio: "inherit" }
);
