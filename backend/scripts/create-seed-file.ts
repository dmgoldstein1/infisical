/* eslint-disable */
import { execFileSync } from "child_process";
import { readdirSync } from "fs";
import path from "path";
import promptSync from "prompt-sync";

const prompt = promptSync({ sigint: true });

const SEED_NAME_RE = /^[A-Za-z0-9_-]+$/;

const migrationName = prompt("Enter name for seedfile: ").trim();
if (!SEED_NAME_RE.test(migrationName)) {
  throw new Error("Seed name may only contain letters, numbers, dashes and underscores");
}
const fileCounter = readdirSync(path.join(__dirname, "../src/db/seeds")).length || 1;
execFileSync(
  "npx",
  [
    "knex",
    "seed:make",
    "--knexfile",
    path.join(__dirname, "../src/db/knexfile.ts"),
    "-x",
    "ts",
    `${fileCounter + 1}-${migrationName}`
  ],
  { stdio: "inherit" }
);
