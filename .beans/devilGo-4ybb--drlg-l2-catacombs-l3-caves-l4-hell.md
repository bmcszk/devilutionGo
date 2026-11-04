---
# devilGo-4ybb
title: DRLG L2 Catacombs + L3 Caves + L4 Hell
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-04T14:57:49Z
parent: devilGo-97mn
blocked_by:
    - devilGo-76tu
---

# Remaining generators

## Goal
L2/L3/L4 parity.

## Research
drlg_l2.cpp:1780-1786 room sizes, :2830/:2839; drlg_l3.cpp:2185; drlg_l4.cpp:1228; fixtures levels/l{2,3,4}data + diablo/hellfire sets; tests drlg_l2..l4_test.cpp.

## Scope
internal/world/drlg_l{2,3,4}.go.

## Acceptance criteria
- [ ] All L2/L3/L4 fixtures byte-equal (diablo + hellfire)
- [ ] River/lava constraints (L3/L4) verified on 20 random seeds vs C++ dumps

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-tests/SKILL.md — fixture matrix

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — procedural generation architecture
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md — level design principles

## Dependencies & blockers
devilGo-76tu (shared engine).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-2 generator task (devilGo-4ybb): L2/L3/L4 parity; draft; wire blockedBy [76tu] ✓ = body's devilGo-76tu shared engine.
2 local: ALL four cites verified exact: drlg_l2.cpp:1780-1786 (RoomMax=10/RoomMin=4) ✓, :2830 CreateL2Dungeon ✓, :2839 LoadPreL2Dungeon ✓, drlg_l3.cpp:2185 CreateL3Dungeon ✓, drlg_l4.cpp:1228 CreateL4Dungeon ✓; fixtures levels/l{2,3,4}data ✓ on disk; drlg_l2..l4_test.cpp ✓ on disk. ACs (byte-equal fixtures diablo+hellfire; 20-seed river/lava vs C++ dumps) are TDD-first and testable.
3 online: n/a.
4 solution: none — ready.
5 understanding: world layer; reuses 76tu's engine; the 20-seed C++-dump AC is the strongest form of parity for non-fixture paths.
6 repro: `sed -n '1780,1786p;2830,2843p' devilutionX/Source/levels/drlg_l2.cpp; sed -n '2185,2187p' devilutionX/Source/levels/drlg_l3.cpp; sed -n '1228,1230p' devilutionX/Source/levels/drlg_l4.cpp; ls devilutionX/test/fixtures/levels`
7 fix: none.
8 validation: n/a.
9 related: 76tu (dep), foug (dependent), vti7/m2x7 (transitive).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.


## Progress log

### CLAIM 2026-11-04 (lane devilGo-4ybb, worktree dg-lane-4ybb @ 3c04ab1)

**Blocked-by cleanup:** removed `blocked_by: [devilGo-1spo]` (archived phantom; the
fixture-storage bean it named is archived, so it can neither unblock nor be waited
on). `blocked_by: [devilGo-76tu]` stays recorded but is satisfied — PR #30 merged the
L1 generator corpus this lane builds on.

**Session scope (this lane, session 1):** golden corpus ONLY. `internal/testdata/drlg`
gains the `DTYPE_CATACOMBS` family (levels 5-8, `CreateL2Dungeon`). The Go port of
`drlg_l2.cpp` is the NEXT session; no `internal/diablo/levels` generator code is
written here.

### Census (disk-first, before touching the probe)

| Check | Result |
|---|---|
| `/tmp/dvlx-pinned` @ `ca91d778f` | present (worktree) |
| `/tmp/drlg-out/drlg` (drlg artifacts) | ABSENT — only `/tmp/dun-out/dun` existed, so catacombs had to be dumped |
| committed corpus | 128 CSVs = 8 variants × 2 entries × 8 seeds (l1-l4 cathedral, l21-l24 crypt) |
| probe `DrlgVariants` | 16 entries; `DTYPE_CATACOMBS` arm present in `GetTileCount` (160) but unexercised |
| levels 5-8 dispatch | `gendung.cpp:326-344` `GetLevelType` → `DTYPE_CATACOMBS` (2); `gendung.cpp:360-362` → `CreateL2Dungeon` |

### Probe: variant table + a broken committed patch

Added `{5..8} x {main, prev}` to `DrlgVariants` (`tools/probe/dvg_probe.cpp`) and
verified the catacomb flow against `drlg_l2.cpp` rather than assuming it:

- `CreateL2Dungeon` = `SetRndSeed` → `GenerateLevel(entry)` → `Pass3()`
  (`drlg_l2.cpp:2830-2837`); `Pass3` = `DRLG_LPass3(11)` + `InitDungeonPieces`, and
  `DRLG_LPass3` (`gendung.cpp:581-600`) writes `dPiece` only — it cannot perturb the
  `dungeon`/`dTransVal`/`ViewPosition` triple the goldens capture.
- `GenerateLevel` (`drlg_l2.cpp:2679-2820`) sets `ViewPosition` in `PlaceStairs`
  (`:2658-2673`) and calls `DRLG_CheckQuests` after `memcpy(pdungeon, ...)`, so the
  dTransVal window is quest-state dependent → all four catacomb levels are pinned,
  not level 5 alone.
- `GetTileCount(DTYPE_CATACOMBS)` = 160 bounds the observed tile vocabulary.

**Defect found and fixed:** `tools/probe/patches/devilutionx-ca91d778f-dvg-probe.patch`
was truncated — its last hunk (`Source/diablo.cpp`, the `ProbeStartGameLevel` /
`ProbeGameEventHandler` shim) ended mid-context, so `git apply` failed with
`corrupt patch at line 75` and no fresh session could rebuild the probe from the
committed patch. The shims were re-applied to a pristine `ca91d778f` checkout,
regenerated with `git diff` (79 lines) and verified with `git apply --check` on the
clean tree.

### Converter: committed, not throwaway

The l1/l2 CSVs were produced by an uncommitted helper, so the refresh path was not
reproducible. This lane adds:

- `testsupport.ParseDunArtifact(dun, meta []byte)` — parses the probe's `.dun` +
  `.meta` pair into the `DunPair` the corpus encoder already consumes.
- `tools/probe/dunconv` — `go run ./tools/probe/dunconv <artifacts> <corpus>`, which
  writes every pair through `testsupport.EncodeDunCSV` (one encoder, so the refresh
  cannot drift from the read path).

Both artifact layers are **x-major**: `dungeon` and `dTransVal` are declared
`[MAXDUNX][MAXDUNY]` (`Source/levels/dun_tile_data.cpp:52-53`) and the probe copies
them linearly.

## Proof of Work

### Corpus regeneration (byte-exactness gate on the EXISTING corpus)

`go run ./tools/probe/dunconv /tmp/drlg-out/drlg internal/testdata/drlg` wrote 192
CSVs. `git status --porcelain internal/testdata/drlg` then reported **64 `??` and
zero `M`**: every one of the 128 previously committed cathedral/crypt CSVs was
reproduced byte-identically by the rebuilt probe, and only the 64 catacomb files are
new. That is the provenance check — a shifted RNG, a perturbed pass order or a
transposed layer would have shown up as modifications.

Manifest check: `diff <(tail -n +5 old) <(tail -n +5 new)` = 64 added rows, 0
changed rows (the 128 existing sha256 + 6 setpiece rows are untouched).

### New goldens (64 files, `l{5,6,7,8}-{main,prev}-<seed>.csv`)

| Property | Observed |
|---|---|
| files | 64 = 4 levels × 2 entries × 8 pinned seeds |
| distinct file sha256 | 64 / 64 |
| distinct (level, entry, seed, view) | 64 / 64 |
| tile vocabulary | 114 distinct ids, min 1, max 160 (= `GetTileCount(DTYPE_CATACOMBS)`) |
| `leveltype` record | 2 (`DTYPE_CATACOMBS`) on all 64 |
| max `dTransVal` id | 21 (generator's transparency regions) |

### Gates

```
$ go build ./...                     → rc 0
$ gofmt -l .                         → (empty)
$ ./scripts/lint.sh                  → 0 issues.  (simpurity: internal/game internal/diablo clean)
$ go vet ./internal/...              → rc 0
$ go test ./... -count=1             → 44 packages ok, 0 FAIL
$ python3 scripts/beans-lint.py .    → 0 error(s)
```

### Scope exception (declared)

`internal/diablo/levels/l1_parity_test.go:TestCorpusShape` pins the committed corpus
size (`l*-*.csv` glob) and failed with `corpus files = 192, want 128` once the new
family landed. That file is outside this lane's stated ownership, but leaving the
gate red was not an option and weakening the assertion was not either: the count was
updated to 192 (24 variants × 2 entries × 8 seeds) with the comment naming the three
generator families. No other file outside
`internal/testdata/drlg/**`, `internal/diablo/testsupport/**`, `tools/probe/**`,
`.beans/` was touched.

### TDD evidence

RED (before `parseDunArtifactGrid` treated the transparency window as x-major):

```
$ go run ./tools/probe/dunconv /tmp/drlg-out/drlg internal/testdata/drlg
dunconv: 192 CSVs in internal/testdata/drlg
$ git status --porcelain internal/testdata/drlg | awk '{print $1}' | sort | uniq -c
     64 ??
    128 M          ← record 2 differed on 3858/6400 cells (l1-main-2588)
```

GREEN: after the fix, 64 `??` and zero `M`.

Residual: `TestCorpusShape` now pins a count, not a per-family matrix, so a future
generator family must bump it consciously (deliberate: a count is the anti-truncation
guard, and each family's own parity test enumerates its variants explicitly).


### Session 8 (2026-11-04, lane devilGo-4ybb): lint-zero then commit

Starting point: port complete and parity-green, `./scripts/lint.sh` red with 26
findings across `l2_layout.go`, `l2_stages.go`, `l2_themerooms.go`, `l2_tables.go`.

Fixes (each verified with `go test ./internal/diablo/levels/ -count=1` before the
next one; parity stayed green at every step):

- `l2_themerooms.go`: `placeCatacombThemeRooms` → per-spot helper; `themeRoomSize`
  split into `themeRoomOriginAllowed` + `themeRoomSweep` + the two ring probes
  (`themeRoomRunWidth`/`themeRoomRunHeight`, gendung.cpp:65-83); `paintCatacombThemeRoom`
  split into border fill / corner stamps / door carve (gendung.cpp:132-147).
- `l2_stages.go`: `FixTilesPatterns` and `FixDirtTiles` six-rule chains became data
  tables (`l2TilePatternRules`, `l2DirtRules`) — rule order is not load bearing, noted
  in the doc comments; `fixCatacombTransparency` per-tile + explicit `default`;
  `FixLockout` split into edges/rows/columns with `lockout*Candidate` + `scanLockout*`;
  `fixCatacombDoors`, `Substitution`, `FixShadows` per-cell helpers.
- `l2_layout.go`: `defineRoom(forceHW bool)` split into `defineRoom` +
  `protectSetPieceRoom` (no bool-flag param); `placeRoom` result values named;
  `hallToParent` explicit `default`; `bounce` receiver dropped (was unused) and split
  into `hallPastEdge`/`hallOpposite`; `TurnaAtQuadrant` switch became the
  `quadrantTurns` table; `placeHallStep(fInRoom bool)` split into
  `placeHallStepInRoom`/`placeHallStepInHall` with `placeHallSides`/`placeHallSide`;
  `knockWalls` split row/column; `DoPatternCheck` cell test became the
  `patternCellAllowed` table; `closeQuadrants` split per-cell + `ringOffsets`;
  stale `//nolint:cyclop,nestif,revive` above `build` removed.
- `l2_tables.go`: the two 161-entry base-type tables are upstream data, not code to
  extract; `dupl` is suppressed on both with that reason.

Parity regression caught and fixed mid-flight: `placeHallSide` first took the
horizontal branch for `Up/Down`, which failed `l5..l8` seeds (the wide-hall flags are
only set on set-piece levels). After correcting the axis (`drlg_l2.cpp:1913-1922`)
all 64 goldens passed again.

### Gates (session 8, post-commit, all real output)

```
$ go build ./...                                   rc=0
$ gofmt -l .                                       (empty)
$ ./scripts/lint.sh                                rc=0, "0 issues."
$ go vet ./internal/...                            rc=0
$ go test ./... -count=1                           44 packages ok, 0 FAIL
$ python3 scripts/beans-lint.py                    "beans-lint: OK (165 beans)"
$ just sqlc-check                                  rc=0, "no raw SQL outside sqlc-generated code"
```

Pushed: `be89ac8` (`devilGo-4ybb: catacombs DRLG port — parity green, lint clean`) on
`dg-4ybb`. The remote lane had moved one docs commit ahead (`fab7034`), so the two
lane commits were rebased onto it — tree verified identical to the pre-rebase commit
(`git diff 97db587 HEAD` empty) before push.
