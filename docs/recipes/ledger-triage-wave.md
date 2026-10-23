# Full-ledger bean triage wave — recipe (pi-as-brainer)

Class task: triage EVERY bean in a port-planning ledger (not bug reports — implementation tasks) one by one, honestly, via pi sub-agents, with the parent as single writer.

## Contract skeleton (paste per wave)
- Header: GIT JAIL (no git writes), NO `beans create/update/delete` (child investigates and REPORTS ONLY), repo untouched, /tmp writes allowed for batch outputs, cwd = beans repo.
- Method per bean: read body → verify against LIVE state (referenced C++ files exist with file:line spot-checks, spec docs exist, blocked_by ids exist and are logically ordered, skills paths exist, TDD-first AC testable, scope matches recorded DECIDED decisions, duplication scan across ledger).
- Verdict vocabulary: TRIAGED-OK / NEEDS-WORK:<gap> / DUPLICATE-OF:<id> / STALE / DECISION-NEEDED:<question>.
- 9-element TRIAGE block per bean (root / local probes / online if external claims / solution / understanding / repro / fix commands for PARENT to apply / validation / related ids) + VERDICT line.
- Output: one md file per batch ending `BATCH-NN-COMPLETE`; final reply = verdict summary + coverage count; nothing skipped.

## Sizing (binding, both print and RPC drivers)
- **One contract per batch (~4 beans), ≤420 s each, sequential loop, `</dev/null` on every run, gate each on its `BATCH-NN-COMPLETE` marker.** A single whole-ledger contract dies at the wrapper timeout and print mode buffers all output — the kill is a TOTAL loss (0 output bytes), not partial progress: one 15-bean run at 1500 s produced nothing after 25 min, while the same 15 beans in four batched contracts all completed inside the window.
- Triage with a fixed 9-element template is MECHANICAL — route `_glm_flash_first`, not `_glm_first` (hard reasoning is for the fix-design step, not the sweep).
- Milestones judge LAST (their verdicts depend on child-epic scopes); tasks first.

## Batch grouping
- ~10 beans per batch; milestone + all epics first (they gate everything), then children grouped by parent epic; final batch = leftover sweep (ids not covered by any grouping) so coverage = sum check.
- Persist grouping as JSON (`/tmp/<proj>-triage-batches.json`); coverage check = union of batches vs `beans list --json` manifest.

## Driver (resumable)
- One PiRpcClient per batch; done() gate = output file exists AND >500 bytes AND contains `BATCH-NN-COMPLETE`.
- Re-running the driver skips completed batches and retries failed ones — cheap resume.
- Batch 01 is always the slowest (milestone bodies are huge); expect 15–25 min for it, 10–20 for the rest.

## Parent duties while the wave runs
- Spot-check each completed batch immediately (split file on `### `, extract `VERDICT:` lines, check the marker): reject a batch whose blocks cite only prior notes with no live probes — re-dispatch with pointed probes.
- Probe liveness via growing omniroute call_logs (`find ~/.omniroute/call_logs/<day> -mmin -3 | wc -l`), NOT via pgrep on pi (argv is just `pi`).
- Tell the user UPFRONT the wave is read-only and bean mutations come in the write-back phase — empty ledger mid-wave is the invariant working, not a stall.

## Write-back phase (parent only, after ALL batches + spot-checks)
1. Aggregate verdicts across batch files into one table (verdict counts + id lists per verdict).
2. Apply fixes for NEEDS-WORK: `beans update <id> --body-append "TRIAGE <date> …"` with the 9-element block; wire missing blockers; patch AC/scope defects the triage named.
3. Promote TRIAGED-OK beans that are dependency-free to `todo` (draft→todo), respecting the recorded execution-order decision. Every promoted bean gets a `## Libraries` section in the same pass — beans-lint rejects todo beans without one; a stdlib-only placeholder line is enough where no dependency attaches yet (lib substitutions are milestone decisions, never worker choices).
4. Route DECISION-NEEDED items to the user (batch them into one grilling round, not one-by-one).
5. Commit ledger changes in one commit referencing the milestone bean.

## Fix pass (parent applies NEEDS-WORK corrections)
- Fix pass 1 applies corrections parsed from the triage blocks (dep wires, count fixes, decision alignments, seams). Fix pass 2 is a VERIFICATION wave that re-checks every fix against live state (read-only contract, verdicts FIX-VERIFIED / FIX-INCOMPLETE / FIX-WRONG). Expect roughly a third of stamps to come back FIX-WRONG the first time: notes-only appends feel like fixes but the operative body never changed. The rule the failure teaches: a fix note that says what should change is NOT a fix — the operative body text must change in the same pass, and a separate verification wave must confirm it did.
- Parse fix commands from triage prose with fully-qualified id regexes only (`devilGo-[a-z0-9]{4}`); loose token patterns over free text produce mangled ids that fail silently or corrupt wiring.
- When patching bean bodies by file edit, never put replacement text containing `: ` into the YAML frontmatter (title lines) — unquoted colons break the YAML and take the whole ledger CLI down with one bean. After any bulk patch, load the full ledger once (`beans list --json`) as a parse gate.

## Pitfalls
- Weak-reply failure mode: a batch can exit 0 with a placeholder/thinking-only reply and NO output file — the done() gate catches it; retry the batch, don't debug the contract.
- **Triage findings arrive as FIX INSTRUCTIONS, not fixes — the parent applies them as OPERATIVE BODY EDITS before promotion:** one wave returned six NEEDS-WORK items for a single bean (stale function citations, false rendering premise, missing consumer demarcation); the fix was surgical replaces of the premise text in the bean body + an appended TRIAGE verdict line, THEN `beans update -s todo`. Appending only the verdict leaves the operative premises stale and the next worker inherits them. Verify each operative edit landed with a grep for the NEW text before promoting.
- Verdicts drift when milestone DECIDED amendments postdate bean bodies (e.g. MPQ/save/settings decisions). Expect a cluster of NEEDS-WORK "body contradicts decision #N" — bulk-patch those in write-back with one append per bean.
- Blocker wiring declared in bean bodies but never applied to the tracker is a recurring gap class — verify `blocks` edges against the tracker, not the body text.
- Overlap seams: two beans claiming the same golden AC (e.g. light-table generation in both a shader bean and a tables bean) — name the canonical owner in DUPLICATE-OF and strip the claim from the loser in write-back.
- Mirror the verification wave's own launch discipline: batch runs longer than ~4 minutes go through a background terminal, never an inline code cell — the cell timeout kills the driver mid-dispatch and orphans the wave.
