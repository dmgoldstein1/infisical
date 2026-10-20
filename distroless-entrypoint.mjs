// Shell-free replacement for standalone-entrypoint.sh:
//
//   #!/bin/sh
//   update-ca-certificates
//   exec node --enable-source-maps dist/main.mjs
//
// A distroless image has no /bin/sh and no `update-ca-certificates`, so this script does
// the equivalent in Node and then re-execs the server. It runs as uid 65532, which cannot
// rewrite the root-owned system bundle at /etc/ssl/certs/ca-certificates.crt, so extra
// certificates are merged into a copy under /tmp and exported through the environment
// instead: NODE_EXTRA_CA_CERTS for Node itself, and SSL_CERT_FILE / GIT_SSL_CAINFO for the
// OpenSSL-based tools the server spawns (git over HTTPS, curl, the infisical CLI).
//
// Extra certs are read recursively from EXTRA_CA_CERT_DIR, defaulting to Debian's
// /usr/local/share/ca-certificates. With none present this is a no-op and the runtime uses
// the distroless CA bundle as-is, exactly as `update-ca-certificates` does when no local
// certificates are installed.
//
// ENTRYPOINT ["/nodejs/bin/node", "/usr/local/bin/distroless-entrypoint.mjs"]
// CMD        ["--enable-source-maps", "dist/main.mjs"]   (the args below become its args)

import { spawn } from "node:child_process";
import { mkdtemp, readFile, readdir, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

const SYSTEM_BUNDLE = "/etc/ssl/certs/ca-certificates.crt";
const EXTRA_CA_DIR = process.env.EXTRA_CA_CERT_DIR || "/usr/local/share/ca-certificates";

const log = (msg) => process.stderr.write(`[distroless-entrypoint] ${msg}\n`);

const listCerts = async (dir) => {
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return [];
  }

  const found = [];
  for (const entry of entries) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      found.push(...(await listCerts(path)));
    } else if (/\.crt$/i.test(entry.name)) {
      // Only .crt, matching Debian's update-ca-certificates; a .pem here is often a key.
      found.push(path);
    }
  }
  return found;
};

const prepareCaBundle = async () => {
  const extraCerts = await listCerts(EXTRA_CA_DIR);
  if (extraCerts.length === 0) return;

  const chunks = [];
  try {
    chunks.push(await readFile(SYSTEM_BUNDLE));
  } catch {
    // No system bundle in the base image: fall back to the extra certs alone.
  }

  let merged = 0;
  for (const certPath of extraCerts) {
    try {
      chunks.push(await readFile(certPath), Buffer.from("\n"));
      merged += 1;
    } catch (err) {
      // One unreadable cert must not discard the rest.
      log(`warning: skipping ${certPath}: ${err.message}`);
    }
  }
  if (merged === 0) return;

  const bundlePath = join(await mkdtemp(join(tmpdir(), "ca-")), "ca-certificates.crt");
  await writeFile(bundlePath, Buffer.concat(chunks), { mode: 0o444 });

  // Node ADDITIONS: keep an operator's value, only add the merged bundle as an extra.
  process.env.NODE_EXTRA_CA_CERTS ||= bundlePath;
  // For the OpenSSL-based tools the server spawns (git, curl, the CLI) the file is the whole
  // trust store, so it must point at the merged bundle. The distroless base already sets these
  // to the stock system bundle, so replace that default but keep a deliberate override.
  for (const name of ["SSL_CERT_FILE", "GIT_SSL_CAINFO"]) {
    if (!process.env[name] || process.env[name] === SYSTEM_BUNDLE) {
      process.env[name] = bundlePath;
    }
  }
  log(`merged ${merged} extra CA certificate(s) from ${EXTRA_CA_DIR} into ${bundlePath}`);
};

try {
  await prepareCaBundle();
} catch (err) {
  // A refresh failure must never stop the server from starting.
  log(`warning: CA bundle refresh skipped: ${err.message}`);
}

const args = process.argv.slice(2);
if (args.length === 0) {
  log("no command supplied: nothing to run (expected CMD args to forward to node)");
  process.exit(1);
}

// Re-exec the server as a child so this process stays the signal target (pid 1 in the
// container) and forwards SIGTERM/SIGINT to it for a graceful shutdown.
const child = spawn(process.execPath, args, { stdio: "inherit", env: process.env });
const SIGNAL_NUMBERS = { SIGHUP: 1, SIGINT: 2, SIGQUIT: 3, SIGTERM: 15 };

for (const signal of Object.keys(SIGNAL_NUMBERS)) {
  process.on(signal, () => {
    if (!child.killed) child.kill(signal);
  });
}

child.on("error", (err) => {
  log(`failed to launch ${process.execPath}: ${err.message}`);
  process.exit(127);
});

child.on("exit", (code, signal) => {
  process.exit(signal ? 128 + (SIGNAL_NUMBERS[signal] ?? 0) : (code ?? 0));
});
