---
# devilGo-49ph
title: 'refresh-gamedata-goldens is not idempotent: committed goldens are scaffold-era row slices'
status: completed
type: bug
priority: normal
created_at: 2026-11-05T23:21:53Z
updated_at: 2026-11-05T23:54:17Z
---

## Finding (lane dg-binaries s1, during the ADR-020 gamedatagoldens move)

`just refresh-gamedata-goldens` does not reproduce the committed fixtures.

Evidence, collected without touching the repo (`git archive HEAD` export to
`/tmp`, moved generator copied in, `go run ./tools/gamedatagoldens`):

- generator output stamps `probeCommit: refresh-recipe-v1` and emits the full
  table (spelldat 36 rows/15 cols, misdat 68/8, monstdat 112/41, objdat 109/12,
  textdat 338/4, questdat 24/9, towners 12/11);
- committed `internal/testdata/gamedata/goldens/*.json` stamp
  `probeCommit: unit-fixture-slice-a` and hold a row slice (spelldat-base.json
  is 631 lines vs the generated full table), last touched in d6ef2dc (scaffold).

So the first clean-tree run of the recipe rewrites all seven goldens, breaking
the recipe's own contract (print a clean diffstat) and the bean AC "git diff on
the goldens is empty after one run".

## Why it was not fixed in that lane

The lane's scope guard forbids regenerating any committed golden (a move, not a
refresh), and the consumers (`internal/diablo/gamedata/audit`) currently pass
against the slices — so the decision to regenerate (and re-pin the row counts)
belongs to the parent, not to the move.

## Acceptance criteria

- [x] Recipe `just refresh-gamedata-goldens` run on HEAD produces exactly one changed file: internal/testdata/gamedata/goldens/spelldat-base.json (597-line diff, per the triage measurement)
- [x] Second run produces an empty diff (idempotency, already proved in triage)
- [x] All other goldens byte-identical to a9961d6 state
- [x] Full gates green: go build, gofmt, lint.sh, go vet, go test ./... -count=1, beans-lint OK, just sqlc-check
- [x] Push dg-goldens-drift for a fixture-only PR

## Libraries

No new library bindings; ADR-009 catalog untouched. Generator CLI + testdata only.

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md


## Triage needed

Decide: regenerate the seven goldens from the committed TSVs (and update the
audit row-count expectations/manifest), or keep the slices and make the
generator emit them (e.g. a target row limit). Then re-run
`just refresh-gamedata-goldens` on a clean tree and require an empty diff.

## Triage (lane goldens-drift s1, 2026-11-06) — trial run + options, no decision

Delegated-triage evidence: shell trial by goldens-drift s1 (dg-goldens-drift @ d231d3e); the drift was first surfaced by reviewer session task-dgbinaries-review-brainer as F1 (see .hermes/lane/dg-binaries/dg-binaries.review.md).

Trial on a clean tree at 19ab83a (`git status --porcelain` empty, DG_ASSET_ARCHIVES
unset — the generator is zero-dependency: it reads only committed TSVs,
`tools/gamedatagoldens/main.go:22,27-33`). Ran `just refresh-gamedata-goldens`
(recipe `justfile:53-62`) and then `go run ./tools/gamedatagoldens` a second time.

Exit code 0, recipe stdout (abridged, all seven fixtures written):

    fixture written tsv=internal/testdata/gamedata/spells/spelldat.tsv rows=36 columns=15
    fixture written tsv=internal/testdata/gamedata/misdat/misdat.tsv rows=68 columns=8
    fixture written tsv=internal/testdata/gamedata/monstdat/monstdat.tsv rows=112 columns=41
    fixture written tsv=internal/testdata/gamedata/objdat/objdat.tsv rows=109 columns=12
    fixture written tsv=internal/testdata/gamedata/textdat/textdat.tsv rows=338 columns=4
    fixture written tsv=internal/testdata/gamedata/quests/questdat.tsv rows=24 columns=9
    fixture written tsv=internal/testdata/gamedata/towners/towners.tsv rows=12 columns=11
     .../testdata/gamedata/goldens/spelldat-base.json   | 597 ++++++++++++++++++++-
     1 file changed, 596 insertions(+), 1 deletion(-)
     M internal/testdata/gamedata/goldens/spelldat-base.json

### (a) Measured delta — ONE file, not seven

sha256 before vs after the run (`internal/testdata/gamedata/goldens/`):

| golden | rows committed / generated | probeCommit committed / generated | sha256 |
|---|---|---|---|
| spelldat-base.json | 1 / 36 | `unit-fixture-slice-a` / `refresh-recipe-v1` | ec6b2e0… → 57f9759… |
| misdat-base.json | 68 / 68 | `refresh-recipe-v1` both | unchanged 9fd5ef8… |
| monstdat-base.json | 112 / 112 | both | unchanged 85e7da7… |
| objdat-base.json | 109 / 109 | both | unchanged 46a4c30… |
| textdat-base.json | 338 / 338 | both | unchanged 476d6e2… |
| questdat-base.json | 24 / 24 | both | unchanged 04d7cdf… |
| towners-base.json | 12 / 12 | both | unchanged 1fee315… |

`git diff --numstat` = `596 1 spelldat-base.json`; the other six produce no diff
bytes. The whole diff is `probeCommit` line 19 plus 35 appended rows — the
committed spelldat row 0 (`Firebolt`) is a strict prefix of the generated table
(`nrows[:1] == orows`).

Direction: the generator emits MORE rows (36 vs 1) for spelldat; for six tables
there is no row/column/probeCommit difference at all. So the draft finding's
"all seven goldens were rewritten, committed files hold a row slice" is
overstated: only `spelldat-base.json` is a slice, the other six already match
the generator byte for byte. Second run reproduces 57f9759… exactly — the
generator is idempotent after the first run.

Why the single slice: `git log -- internal/testdata/gamedata/goldens/spelldat-base.json`
shows only d6ef2dc (scaffold), where it was already `unit-fixture-slice-a`,
1 row. The branch commit 0a566c2 "full-table goldens for 7 tables" added six
files (9615 lines) and is NOT an ancestor of HEAD; no commit ever regenerated
spelldat. `unit-fixture-slice-a` occurs nowhere in the tree outside this JSON
line — the slice has no reproducible provenance.

### (b) The move (9f22112 / 19ab83a) is NOT the cause — no new bug from it

- `git diff a9961d6 HEAD -- internal/testdata/gamedata/goldens` → empty; the
  goldens are identical before and after the move.
- `git diff a9961d6 HEAD --stat -- cmd/gamedatagoldens tools/gamedatagoldens` →
  `{cmd => tools}/gamedatagoldens/main.go | 0` (pure rename, 0 changed lines).
- Path constants identical at both ends: `probeCommit = "refresh-recipe-v1"`
  and `goldenDir = "internal/testdata/gamedata/goldens"` plus all seven TSV
  paths (`tools/gamedatagoldens/main.go:20,22,27-33` == a9961d6
  `cmd/gamedatagoldens/main.go:20,22,27-33`).

Conclusion: the drift predates the move; the move changed no path constant and
does not break the recipe. No separate move-bug to file.

### (c) Side finding (doc drift, not the cause): the claimed consumer does not exist

`grep -rn "testdata/gamedata/goldens" --include=*.go` returns only the generator
itself (`tools/gamedatagoldens/main.go:22`). No test reads `goldens/*.json`.
`internal/diablo/gamedata/audit/audit_test.go:24-41` pins TSV paths and
`audit_manifest.json` row counts, and has no spelldat entry. Both
`docs/research/dg-conversion-inventory.md:24` and `tools/probe/README.md:96`
state that the audit tests consume the golden JSON — stale. Practical effect:
regeneration touches no test expectation. `go test -count=1
./internal/diablo/gamedata/... ./internal/testsupport/...` passes with the
regenerated file (all ok).

### (d) Options (this lane picks none; status stays draft)

**Option 1 — regenerate, generator output authoritative.** Run the recipe once
on a clean tree and commit `spelldat-base.json`; six files do not change.
- Cost: one recipe run + one fixture commit; no code. Then
  `just refresh-gamedata-goldens` own diffstat is empty (proved: second run
  byte-identical), satisfying 48fe's AC and this bean's AC.
- PR shape: 1 fixture file (`internal/testdata/gamedata/goldens/spelldat-base.json`,
  +596/−1 = 597 lines) + optionally 2 doc lines fixing (c).
- Golden-net: 1 of 7 files, 597 lines; committed testdata grows ~596 lines.

**Option 2 — restore slice semantics in the generator.** Add a row-limit
constant (e.g. `rows[:1]`) so the generator emits what is committed.
- Cost: generator code change (`tools/gamedatagoldens/main.go:buildFixture/
  readTSV`, plus a comment/doc line), and it rewrites the six full-table goldens
  down to one row each — deleting ~9615 lines of already-verified fixture data,
  contradicting the audit manifest which pins full TSV row counts, and making
  this bean's "full table" wording false.
- PR shape: 1 code file + 7 rewritten fixtures + docs.
- Golden-net: 7 files, ≈9615 deletions — ~16× the diff of option 1 and loss of
  coverage the other six tables already have.

**Option 3 — keep the committed slice, add an xfail/known-lean gate.** Leave
the fixture as is; make the drift explicit instead of silent.
- Cost: new `-check` mode in the generator (regenerate in memory, compare,
  exit nonzero on drift) or a Go test of the same shape with a spelldat
  allowlist; roughly 40-60 lines of new code plus its test, and the recipe gains
  a flag/branch.
- PR shape: 1 code file + 1 test file (+ justfile line); no fixture change.
- Golden-net: 0 lines of fixture diff, but the one known-slice fixture stays
  unverified forever and the recipe keeps a documented carve-out contradicting
  its own "empty diff" contract.

Not decided here: option 1 is the smallest complete fix on the measurements
above (zero code, 597-line fixture diff, makes the recipe idempotent), but the
choice belongs to the parent. Worktree restored after the trial
(`git checkout -- internal/testdata/gamedata/goldens`, `git status` clean,
all seven sha256 back to the HEAD values); no generator or fixture change from
this lane.



## DECISION (parent, 2026-11-06)

Option 1 chosen: regenerate the fixture, generator output authoritative. Zero code change; idempotency already proved (second run = empty diffstat). Execution: next session runs the recipe, commits internal/testdata/gamedata/goldens/spelldat-base.json, reruns gates + beans-lint, pushes dg-goldens-drift for a fixture-only PR.

## Proof of Work

Lane dg-goldens-drift (branch `dg-goldens-drift`, parent d231d3e). Option 1
(regenerate, generator output authoritative) executed; one fixture touched.

### Commits

- `714e628` beans: 49ph promoted to todo with option-1 decision (parent)
- `952615c` gamedata: regenerate spelldat-base.json from HEAD generator (49ph option 1)
- pushed: `d231d3e..952615c  dg-goldens-drift -> dg-goldens-drift`;
  `git ls-remote origin refs/heads/dg-goldens-drift` = `952615c6c4b36c06a5c7336b2646b2c78c1bdd4d` = local HEAD

### Run 1 - `just refresh-gamedata-goldens` (clean tree, exit 0)

    fixture written tsv=internal/testdata/gamedata/spells/spelldat.tsv rows=36 columns=15
    ... (7 tables generated)
     .../testdata/gamedata/goldens/spelldat-base.json   | 597 ++++++++++++++++++++-
     1 file changed, 596 insertions(+), 1 deletion(-)
     M internal/testdata/gamedata/goldens/spelldat-base.json

Exactly one changed file, 597-line diff - matches the triage measurement.
Content delta: `probeCommit: unit-fixture-slice-a` -> `refresh-recipe-v1`,
rows extended to the full 36x15 table.

### Run 2 - idempotency

- literal second recipe run refuses by design on a dirty tree:
  `refresh-gamedata-goldens: working tree is dirty; commit or stash first` (exit 1)
- the recipe body it invokes (`go run ./tools/gamedatagoldens`, exit 0) is
  byte-identical: sha256 of spelldat-base.json
  `57f97590137b3688c1c59ecf0e0a83acfb4bb71c6130856c1cee50ad050b202d`
  before and after; `cmp` against the run-1 copy = identical; `git diff --stat`
  unchanged at 597.

### Other six goldens byte-identical (sha256 unchanged from pre-run baseline)

    misdat    9fd5ef8a6143a303ebe0c843b0d6ebec8e22d92d4f67fe5bfa017c73a7badcf8
    monstdat  85e7da7c75ee0569cf27ea25f9ce6017f79ca4ee1de41279167b79ce7dbae364
    objdat    46a4c30caa4cece3ee5c05fc32fcebb85e00d9c76c3771db25356614efb1a7c5
    questdat  04d7cdf6596fe004b5eb1c51d2a8e62da3b77e4d37051b852b83e55e1fa151a3
    textdat   476d6e22672f70cc2424a3a94cd35ea2dbd1214ba7dba8feeede632321bbc08e
    towners   1fee3150a20222c3ddc34bba7ed9311a699aade577fb3af89fd8a107c7757ddc

All six already carried full tables (`probeCommit: refresh-recipe-v1`) on HEAD,
so on current HEAD only spelldat was still a scaffold slice - the original
"rewrites all seven" observation no longer reproduces.

### Gates (green, on 952615c)

- `go build ./...` -> exit 0
- `gofmt -l .` -> empty
- `./scripts/lint.sh` -> `simpurity: internal/game internal/diablo clean (226 sources)` / `0 issues.`
- `go vet ./internal/...` -> exit 0
- `go test ./... -count=1` -> exit 0 (all packages ok)
- `python3 scripts/beans-lint.py` -> `beans-lint: OK (173 beans)`
- `just sqlc-check` -> `no raw SQL outside sqlc-generated code` (exit 0)

Residual: none. Note for future runs: the recipe's dirty-tree guard means a
literal re-run of the recipe can never serve as the idempotency check;
idempotency is proved on the generator body the recipe invokes.
