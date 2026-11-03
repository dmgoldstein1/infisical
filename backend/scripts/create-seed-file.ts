/* eslint-disable */
import { execFileSync } from "child_process";
import { readdirSync } from "fs";
import path from "path";
import promptSync from "prompt-sync";

const prompt = promptSync({ sigint: true });

const rawName = (prompt("Enter name for seedfile: ") ?? "").trim();
if (!/^[A-Za-z0-9_-]+$/.test(rawName)) {
  throw new Error("Invalid seed name: use only letters, numbers, dashes and underscores");
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
    `${fileCounter + 1}-${rawName}`
  ],
  { stdio: "inherit", shell: false }
);
