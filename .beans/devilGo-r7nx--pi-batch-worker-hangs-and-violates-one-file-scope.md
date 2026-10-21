---
# devilGo-r7nx
title: pi batch worker hangs and violates one-file scope
status: in-progress
type: bug
priority: normal
created_at: 2026-10-21T16:43:07Z
updated_at: 2026-10-21T17:45:28Z
---

## Evidence
- Batch `devilGo-iw3r` was intended to process 26 files one by one.
- The launcher used pi client `--timeout 420` but no external process timeout.
- A worker remained active for hours with only introductory output and no exact `FILE: ... VERDICT` marker.
- State contains multiple `RETRY` records and only 3 completed files before manual stop.
- `git diff --stat` showed 52 modified Go files although the active contract allowed one target file.
- Output: `/tmp/dg-kefy-batch1/out/`; state: `/tmp/dg-kefy-batch1/state.tsv`.

## Expected
Each dispatch terminates at a bounded wall-clock deadline, emits an exact verdict, and cannot modify files outside its declared target (unexpected scope stops the batch).

## Root-cause hypothesis to verify
The RPC client's timeout is request/session-level rather than a hard process deadline; the launcher lacks external `timeout` and pre/post scope enforcement. Existing dirty files from prior attempts also made attribution impossible.

## Acceptance criteria
- [ ] Reproduce/verify timeout semantics from client help/source or a bounded smoke run.
- [ ] Launcher wraps each client invocation in external `timeout;` timeout produces explicit FAILED state.
- [ ] Launcher snapshots porcelain before dispatch and stops if files outside target change.
- [ ] Fresh unique session per attempt; exact marker validated only in per-dispatch output.
- [ ] Smoke run completes within the hard deadline and leaves an auditable state record.

## Skills for coder
- `/home/blaze/.agents/skills/pi-delegation/SKILL.md` — delegation lifecycle and failure ladder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — process-script structure


## Triage result

Confirmed in `/home/blaze/.hermes/skills/dev/pi-delegation/scripts/pi_rpc_client.py:134-147,209-224`: `--timeout` is passed to `queue.Queue.get(timeout=...)`; it raises only when no frame arrives, while the wrapper process can remain alive through teardown/provider behavior. The launcher had no external wall-clock `timeout`. The 52-file dirty tree also proves the launcher had no pre/post scope guard; existing dirt makes attribution ambiguous, so the guard compares exact porcelain records before and after each dispatch and permits only the declared target to change.

Sources read: pi RPC client source and `pi-delegation` failure-ladder/verify-before-believe rules. No online source needed; root cause is local and reproducible.

## Revised acceptance
- [x] Verify timeout semantics from source.
- [ ] External hard deadline around every client invocation; timeout is explicit FAILED/RETRY.
- [ ] Pre/post porcelain scope guard; unexpected paths stop the batch.
- [ ] Smoke run proves bounded termination and auditable state.


## Run evidence (batch 1 re-run)

- Runner completed exit 0; state shows 4 DONE (CLEAN), 22 RETRY (no verdict within 480s).
- Verdict files: cmd/devilutionGo/main.go, internal/diablo/doc.go, internal/diablo/enums/cursors.go, internal/diablo/enums/spawn.go.
- Scope guard held: no unexpected paths changed (52-file dirty tree matches baseline + those 4 edits).
- Hard deadline works; workers still fail to emit verdicts within 480s on non-trivial files (they emit only "reading skills" or "gathering facts" and hit capped-lane turn death).

## Updated root cause
Workers on files with actual functions/logic exceed the gateway turn output budget before any tool call; `--thinking low` alone is insufficient. Need: smaller micro-contracts (3-4 lines actionable), or force synchronous edit in contract, or use `_glm_first` lane for style work, or split function bodies into separate helper-only dispatches.
