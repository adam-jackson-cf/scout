# X1 binding implementation — revised scoped review request

This request is outside execution sources after the round-1 fixes. Authority remains round 9 actions (a)–(e). No acceptance, candidate scan, model CLI or inventory operation has run as part of these fixes. The round-1 response contains issue-to-code/test mappings, full verification outputs and current SHA-256s.

## Process observation and owner prerequisites

One ES NOTIFY exec/fork/exit session is used for the whole acceptance operation, including its required synthetic observer proof and sequentially attached revision trees. Global sequence continuity is checked while idle as well as within each tree. Null script metadata is supported. Only related process payloads are persisted. Logging stubs are first on PATH and are the configured Clawpatch provider paths (`CLAWPATCH_CLAUDE_BIN` and `CLAWPATCH_CODEX_BIN`); contextual provider selection resolves the same stub names. Actual execution outside the attested executable set plus these stubs refuses acceptance.

The frontend has **no outer sandbox-exec wrapper**. Product scanner sandboxing remains unchanged and supplies scanner network/write confinement. The prior denied-exec-argv hard gate and denied absolute-path fixture are removed under the round-1 council interpretation: the signed requirement captures child-process starts and argv; synthetic logging stubs capture model attempts. No log-stream, DTrace, SIP change or AUTH-exec entitlement is needed.

The owner provides:

1. Full Disk Access for the responsible app launching the controller (Terminal or Orca).
2. `/etc/sudoers.d/x1-eslogger`, mode 0440, checked using `visudo -cf`, limited to this exact rule:

   ```text
   adamjackson ALL=(root) NOPASSWD: /usr/bin/eslogger exec fork exit
   ```

The controller invokes only `/usr/bin/sudo -n /usr/bin/eslogger exec fork exit`; a credential-cache refresh is not a prerequisite. Main journals the drop-in's presence and removal after the final acceptance operation without logging any credential. `man eslogger` additionally requires Full Disk Access for remote users when using SSH, or for eslogger itself when using a launch daemon.

Before any scanner, the same session executes a scratch-only observer proof: a shebang logging stub with interpreter/script/argv attribution, an orphan grandchild still observed after its parent exits, a pinned-Node child_process.spawn grandchild, contiguous global sequence and no live descendants. The proof has an absolute byte pin and source/runtime/stub pins. Every later launch verifies that pin and its session identity. Proof and scanner acceptance remain **unexecuted until scoped zero-issue review**.

The exact requested unprivileged nesting probe returned exit 0 with empty stdout/stderr on this host; `binding/unprivileged-probe.json` records that exercised result. No nested-sandbox failure is fabricated. Removing the outer sandbox remains the adopted round-1 fix.

## Policy, identity and parity

Both parent and head policies are resolved before the first scanner by the frontend's `--resolve-only` mode. Later effective records must match those pinned bytes exactly. The frozen policy is parsed separately for each bound revision; the scanner reference never applies that loader or calls Assess/model/correlation stages.

The accepted round-1 reading compares Scout's deterministic occurrence identity rather than independent random durable ULIDs. The key includes family, category, StartLine, StartByte, full `domain.AnchorOf` and rule version, with multiplicity. Bound observations must link to exactly one actual durable issue per key; fresh parent mapping is one-to-one, and head discloses carried IDs. Minting provenance remains `internal/issues/correlate.go:28–29,413` and `internal/domain/issue_anchor.go:77–100`.

Both routes use the same deterministic-scanner health rule; intentional model rows must be observed `skipped (not opted in)` rather than demanding overall RunComplete. Accounted coverage and effective scanner inputs must match before finding parity. Omissions and mapping gaps remain visible and must match. Equal-length route roots preserve argv-budget batching. Synthetic source includes frozen-rule triggers that differ between revisions; real scanner trigger measurement belongs to the separately authorized post-review fixture run, not this source-fix step.

Run-config and preparation.py are not modified by this component. Main uses the existing canonical clock, runtime closure and reviewed runtime/source manifests; source closure is rechecked before each fixture frontend operation. The controller exclusively reserves its canonical receipt before work and writes a terminal refusal for ordinary exceptions. A crash leaves the reservation intact for reviewed recovery.
