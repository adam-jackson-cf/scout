#!/usr/bin/env node
// X2c validation wrapper around the FROZEN X2 stub Clawpatch (evaluation/x2-area-stability/validation/stub_clawpatch.js, path in X2C_BASE_STUB). No model call.
// It logs every invocation's argv (X2C_STUB_ARGVLOG, one JSON line each) and rewrites the stub's fixed `agent-start provider=P model=default` progress line to carry the
// `--model` value, as real Clawpatch does (third_party/clawpatch/src/agent-mapper.ts:147-150: `providerOptions.model ?? "default"`). Everything else is the stub.
const { spawnSync } = require("child_process");
const fs = require("fs");

const argv = process.argv.slice(2);
if (process.env.X2C_STUB_ARGVLOG) fs.appendFileSync(process.env.X2C_STUB_ARGVLOG, JSON.stringify({ args: argv }) + "\n");
const r = spawnSync(process.execPath, [process.env.X2C_BASE_STUB, ...argv], { encoding: "utf8", env: process.env });
const mi = argv.indexOf("--model");
const model = mi >= 0 ? argv[mi + 1] : "default";
process.stderr.write((r.stderr || "").replace(/ model=default/g, " model=" + model));
process.stdout.write(r.stdout || "");
process.exit(r.status === null ? 1 : r.status);
