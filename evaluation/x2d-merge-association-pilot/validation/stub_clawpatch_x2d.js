#!/usr/bin/env node
// X2d validation wrapper around the FROZEN X2 stub Clawpatch (evaluation/x2-area-stability/validation/stub_clawpatch.js, path in X2D_BASE_STUB). No model call.
// What it adds to the base stub (the X2c wrapper's behaviour plus the two things the X2d collector needs):
//   * logs every invocation's argv to X2D_STUB_ARGVLOG (one JSON line each: {args, include}), `include` being the `include` array of <state>/config.json at the time of
//     a `map` call (null for other commands); and rewrites the base stub's fixed `agent-start provider=P model=default` progress line to carry the `--model` value;
//   * `init` (delegated to the base stub) is followed by writing <state>/config.json = {"include":["**"],"exclude":[],"codexConfig":{"x2d_marker":"must-stay-untouched"}}
//     unless it exists, so the collector's rewrite of `include` (and its leaving every other key alone) is observable;
//   * after a successful `map` (delegated to the base stub, which keeps every fault-injection variable X2_STUB_*: FAIL_CALLS, PARTIAL_CALLS, SLEEP_*, CHILD_CALLS, COUNTER),
//     the base stub's per-directory agent features are replaced by X2D_STUB_FEATURES (default 3) agent features (source "agent-mapper") that partition the
//     tracked files of the workspace matching the config `include` (an entry matches a file equal to it or below it as a path prefix; `**`, `**/*` and `<dir>/**` are
//     accepted) into contiguous chunks of the sorted file list. The stub's heuristic feature is kept. No matching file -> no agent feature.
// X2D_STUB_FEATURES=0 writes none. The feature ids are `feat_x2d_<sha1(include|chunk index|call counter or 0)>`.
const { spawnSync, execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const argv = process.argv.slice(2);
const opt = (n) => { const i = argv.indexOf(n); return i >= 0 ? argv[i + 1] : undefined; };
const state = opt("--state-dir"), root = opt("--root");
const cmd = argv.filter((a, i) => !a.startsWith("--") && !["--root", "--state-dir", "--source", "--provider", "--model", "--reasoning-effort"].includes(argv[i - 1]))[0];
const cfgPath = state ? path.join(state, "config.json") : null;

const readInclude = () => {
  try { return JSON.parse(fs.readFileSync(cfgPath, "utf8")).include; } catch (e) { return null; }
};
if (process.env.X2D_STUB_ARGVLOG) {
  fs.appendFileSync(process.env.X2D_STUB_ARGVLOG, JSON.stringify({ args: argv, include: cmd === "map" && cfgPath ? readInclude() : null }) + "\n");
}

const r = spawnSync(process.execPath, [process.env.X2D_BASE_STUB, ...argv], { encoding: "utf8", env: process.env });
const mi = argv.indexOf("--model");
const model = mi >= 0 ? argv[mi + 1] : "default";
process.stderr.write((r.stderr || "").replace(/ model=default/g, " model=" + model));
process.stdout.write(r.stdout || "");
const status = r.status === null ? 1 : r.status;
if (status !== 0) process.exit(status);

if (cmd === "init" && cfgPath && !fs.existsSync(cfgPath)) {
  fs.mkdirSync(state, { recursive: true });
  fs.writeFileSync(cfgPath, JSON.stringify({ include: ["**"], exclude: [], codexConfig: { x2d_marker: "must-stay-untouched" } }, null, 2));
}

if (cmd === "map") {
  const include = readInclude() || ["**"];
  const matches = (f) => include.some((e) => {
    if (e === "**" || e === "**/*") return true;
    const p = e.replace(/\/\*\*$/, "").replace(/\/$/, "");
    return f === p || f.startsWith(p + "/");
  });
  const files = execFileSync("git", ["ls-files"], { cwd: root, encoding: "utf8" }).split("\n").filter(Boolean).filter(matches).sort();
  const feats = path.join(state, "features");
  for (const f of fs.readdirSync(feats)) {  // drop the base stub's own agent features of this call; keep stale records and the heuristic feature
    const p = path.join(feats, f);
    const rec = JSON.parse(fs.readFileSync(p, "utf8"));
    if (rec.source === "agent-mapper" && rec.status !== "skipped") fs.unlinkSync(p);
  }
  const n = Math.min(parseInt(process.env.X2D_STUB_FEATURES || "3", 10), files.length);
  const size = n > 0 ? Math.ceil(files.length / n) : 0;
  const call = process.env.X2_STUB_COUNTER && fs.existsSync(process.env.X2_STUB_COUNTER) ? fs.readFileSync(process.env.X2_STUB_COUNTER, "utf8").trim() : "0";
  for (let i = 0; i * size < files.length && n > 0; i++) {
    const chunk = files.slice(i * size, (i + 1) * size);
    const id = "feat_x2d_" + crypto.createHash("sha1").update(include.join(",") + "|" + i + "|" + call).digest("hex").slice(0, 10);
    const title = "Area " + (i + 1) + " of " + path.dirname(chunk[0]);
    fs.writeFileSync(path.join(feats, id + ".json"), JSON.stringify({
      schemaVersion: 1, featureId: id, title, summary: "Stub area covering " + chunk.length + " files", kind: "module", source: "agent-mapper", confidence: "medium",
      entrypoints: [], ownedFiles: chunk.map((p) => ({ path: p, reason: "stub" })), contextFiles: [], tests: [], status: "mapped",
    }));
  }
}
process.exit(0);
