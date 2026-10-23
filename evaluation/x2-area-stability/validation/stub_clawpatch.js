#!/usr/bin/env node
// Validation stub for the Clawpatch CLI used by collect.py. No model call. Mimics only what collect.py relies on:
//   --version | --root R --state-dir S --json init | ... map --source agent --provider P
// `map` marks every existing record `skipped` (as Clawpatch does for records the latest map did not emit), then writes one
// agent-mapper feature per directory of tracked .py files (IDs churn per call) plus one stable heuristic feature, and prints Clawpatch's
// progress lines on stderr (`clawpatch map agent-start provider=P model=default`).
// X2_STUB_COUNTER: file counting map calls across state dirs; X2_STUB_FAIL_CALLS: comma-separated call numbers that exit 3;
// X2_STUB_PARTIAL_CALLS: call numbers that leave a junk active agent record in the state and then exit 3 (a partial write);
// X2_STUB_SLEEP_CALLS / X2_STUB_SLEEP_SECONDS: calls that sleep (timeout test); X2_STUB_SLEEP_ALL: seconds every map call sleeps (wall-clock test);
// X2_STUB_CHILD_CALLS: calls that spawn a DETACHED sleeping child (own process group, like Clawpatch's provider CLI) and then hang; on SIGTERM the stub
// kills that child and exits (as Clawpatch's handler does) unless X2_STUB_IGNORE_TERM is set; X2_STUB_CHILD_PIDFILE receives the child pid.
// X2_STUB_ENVLOG: file receiving one JSON line per invocation with the environment variable NAMES (and the CLAWPATCH_* values).
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { execFileSync, spawn } = require("child_process");

const argv = process.argv.slice(2);
if (argv[0] === "--version") { console.log("stub-0"); process.exit(0); }
const opt = (n) => argv[argv.indexOf(n) + 1];
const root = opt("--root"), state = opt("--state-dir");
const cmd = argv.filter((a, i) => !a.startsWith("--") && !["--root", "--state-dir", "--source", "--provider"].includes(argv[i - 1]))[0];
const feats = path.join(state, "features");
const list = (v) => (v || "").split(",").filter(Boolean).map(Number);
const sleep = (s) => Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, s * 1000);

const counter = process.env.X2_STUB_COUNTER;
let call = 1;
if (cmd === "map" && counter) {
  call = (fs.existsSync(counter) ? parseInt(fs.readFileSync(counter, "utf8"), 10) : 0) + 1;
  fs.writeFileSync(counter, String(call));
}
if (process.env.X2_STUB_ENVLOG) {
  fs.appendFileSync(process.env.X2_STUB_ENVLOG, JSON.stringify({
    cmd, call, args: argv, names: Object.keys(process.env).sort(),
    claudeAuth: process.env.CLAWPATCH_CLAUDE_AUTH_CONTEXT || null, codexTimeout: process.env.CLAWPATCH_CODEX_TIMEOUT_MS || null,
    claudeTimeout: process.env.CLAWPATCH_CLAUDE_TIMEOUT_MS || null,
  }) + "\n");
}

if (cmd === "init") {
  fs.mkdirSync(feats, { recursive: true });
  console.log(JSON.stringify({ ok: true }));
  process.exit(0);
}
if (cmd !== "map") { console.error("unsupported " + cmd); process.exit(2); }

const provider = opt("--provider");
console.error("clawpatch map inventory-start");
console.error("clawpatch map agent-start provider=" + provider + " model=default");
const rec = (id, source, title, owned) => ({
  schemaVersion: 1, featureId: id, title, summary: "Handles " + title, kind: "module", source, confidence: "medium",
  entrypoints: [], ownedFiles: owned.map((p) => ({ path: p, reason: "stub" })), contextFiles: [], tests: [], status: "mapped",
});
if (list(process.env.X2_STUB_FAIL_CALLS).includes(call)) { console.error("stub failure at call " + call); process.exit(3); }
if (list(process.env.X2_STUB_PARTIAL_CALLS).includes(call)) {
  fs.mkdirSync(feats, { recursive: true });
  fs.writeFileSync(path.join(feats, "feat_partial_junk.json"), JSON.stringify(rec("feat_partial_junk", "agent-mapper", "Junk from a partial write", ["junk.py"])));
  console.error("stub partial write at call " + call);
  process.exit(3);
}
if (list(process.env.X2_STUB_CHILD_CALLS).includes(call)) {
  const child = spawn("sleep", ["300"], { detached: true, stdio: "ignore" });
  child.unref();
  if (process.env.X2_STUB_CHILD_PIDFILE) fs.writeFileSync(process.env.X2_STUB_CHILD_PIDFILE, String(child.pid));
  process.on("SIGTERM", () => {
    if (process.env.X2_STUB_IGNORE_TERM) return;
    try { process.kill(-child.pid, "SIGKILL"); } catch (e) { /* already gone */ }
    process.exit(143);
  });
  setInterval(() => {}, 1000);
  return;
}
if (list(process.env.X2_STUB_SLEEP_CALLS).includes(call)) sleep(parseFloat(process.env.X2_STUB_SLEEP_SECONDS || "30"));
if (process.env.X2_STUB_SLEEP_ALL) sleep(parseFloat(process.env.X2_STUB_SLEEP_ALL));

for (const f of fs.readdirSync(feats)) {
  const p = path.join(feats, f);
  const r = JSON.parse(fs.readFileSync(p, "utf8"));
  r.status = "skipped";
  fs.writeFileSync(p, JSON.stringify(r));
}
const files = execFileSync("git", ["ls-files"], { cwd: root, encoding: "utf8" }).split("\n").filter((f) => f.endsWith(".py"));
const byDir = {};
for (const f of files) (byDir[path.dirname(f)] = byDir[path.dirname(f)] || []).push(f);
for (const [dir, fs_] of Object.entries(byDir)) {
  const id = "feat_stub_" + crypto.createHash("sha1").update(dir + ":" + call).digest("hex").slice(0, 10);
  fs.writeFileSync(path.join(feats, id + ".json"), JSON.stringify(rec(id, "agent-mapper", "Module " + dir, fs_)));
}
fs.writeFileSync(path.join(feats, "feat_heur_all.json"), JSON.stringify(rec("feat_heur_all", "python-source-group", "All python", files)));
console.error("clawpatch map agent-done features=" + Object.keys(byDir).length + " elapsed=0s");
console.log(JSON.stringify({ ok: true, call }));
