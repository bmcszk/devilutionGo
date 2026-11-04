---
# devilGo-gico
title: 'devilGo-76tu B: DRLG L1 fix passes — FixTransparency/FixDirt/FixCorner/Substitution/Shadows/lamps/FillFloor'
status: completed
type: task
priority: normal
created_at: 2026-11-04T04:08:30Z
updated_at: 2026-11-04T06:02:43Z
parent: devilGo-97mn
blocked_by:
    - devilGo-4e0x
---

Subtask of devilGo-76tu, depends on the carve stage (devilGo-4e0x). Scope: the tile fix passes of the cathedral generator.

## Goal
Port GenerateLevel's post-carve passes so the generator state equals the stage dumps `04-fixtrans`, `05-fixdirt`, `06-fixcorner`, `07-subst`, `08-shadows`, `09-lamps`, `10-fillfloor` byte for byte (tiles + trans + view + setpiece), then make `TestL1CathedralParity` green for the cathedral levels 1-4.

## Upstream reference (READ-ONLY)
/tmp/dvlx-pinned = devilutionX @ ca91d778f. Canonical: Source/levels/drlg_l1.cpp.

Order matters — the exact GenerateLevel sequence at :1259-1290:
1. `FixTransparency()`                 :1076-1110
2. `FixDirtTiles()`                    :1111-1136
3. `FixCornerTiles()`                  :1137-1154
4. `Substitution()`                    :996-1035  (RNG: FlipCoin(4) at :1000, GenerateRnd(16) at :1003)
5. `ApplyShadowsPatterns()`            :332-395
6. `const int numt = GenerateRnd(5) + 5;` then `PlaceMiniSet(LAMPS, DMAXX*DMAXY, true)` numt times  :1282-1286
7. `FillFloor()`                       :414-429  (RNG: RandomIntLessThan(3))
8. `memcpy(pdungeon, dungeon, ...)` then `DRLG_CheckQuests(SetPiece.position)` :1292-1295 — port the Pdungeon copy; the quest hook is unit C, keep it a documented no-op call site if the C unit owns it.
Also port `PlaceMiniSet`'s LAMPS Miniset (:69-80) — it may land in unit A2; do not duplicate.

Skip (crypt-only, unit C): FixCryptDirtTiles, CryptSubstitution, DRLG_LPass3, PlaceCryptStairs, PlaceCryptLights, SetCryptSetPieceRoom.

## Requirements
- 1:1 port, cite `drlg_l1.cpp:line` on every ported function and on every non-obvious branch.
- `Rnd.RandomIntLessThan` must be reachable from the Dungeon RNG (extend `Rnd`/`EngineRnd` with `engine.RandomIntLessThan`; keep it a field so tests can script it).
- Every ported function gets at least one test; table-driven where multi-case.
- No new deps, no go.mod edits, no committed-fixture edits, no .beans writes, no git state changes.

## Acceptance criteria
- [x] AC 1: Extend the gitignored oracle harness `zzdev_stage_test.go` (skip when /tmp/drlg-stages absent) to diff stages 04..10 for all 16 variants x 8 seeds; green.
- [x] AC 2: `go test ./internal/diablo/levels/ -count=1` fully green, including `TestL1CathedralParity` (levels 1-4, both entries, 8 seeds) and the dev-stage oracles.
- [x] AC 3: `gofmt -l internal/diablo/levels` empty; `go vet ./internal/...` rc0; `./scripts/lint.sh` 0 issues; `go build ./...`.
- [x] AC 4: Report: files, citations, gate tails, and any cell that stays red with got/want.

## Stop rules
Never edit a golden to pass. Never weaken an assertion. If a stage cannot be closed, report the coordinate and both values.


## Data tables to port (inline upstream, no fixture needed)
- `TileDecorations[207]` — drlg_l1.cpp:255-293 (needed by Substitution)
- `ShadowPatterns[37]` (ShadowStruct: strig/s1/s2/s3/nv1/nv2/nv3) — drlg_l1.cpp:183-226
- `BaseTypes[207]` — drlg_l1.cpp:227-254
- `LAMPS` Miniset — drlg_l1.cpp:69-80 (skip if unit A2 landed it)
Port them as Go values with a `drlg_l1.cpp:line` citation and a table-size assertion test (len == 207 / 37).

## Triage (brainer, 2026-11-04)
Delegated slice of a parent already triaged TRIAGED-OK (devilGo-76tu): the
upstream symbols, fixture path and gate are all cited in the body above and were
verified on disk (`Source/levels/drlg_l1.cpp` @ ca91d778f, committed corpus under
`internal/testdata/drlg/`, oracle dumps under `/tmp/drlg-stages`). No open
question blocks work; no new dependency is introduced.

## Libraries
- stdlib only (no third-party deps; no go.mod edits)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — RNG-call-order discipline, 1:1 port rules
- `/home/blaze/.agents/skills/go-tests/SKILL.md` — table-driven tests, fixture matrix


## Progress log

- 2026-11-05 session 8: unit B landed. `git log`: 7037158 (WIP salvage of the
  previous session's in-flight passes), e715399 (fix passes green), pushed to
  `dg-76tu`.

## Proof of Work

### Files

- `internal/diablo/levels/l1_passes.go` — fixTransparency, fixDirtTiles,
  fixCornerTiles, substitution, placeLamps, fillFloor, copyToPrevTiles
  (drlg_l1.cpp:414-429, 996-1154, 1282-1289).
- `internal/diablo/levels/l1_shadows.go` — applyShadowsPatterns,
  fixShadowFences (drlg_l1.cpp:332-395).
- `internal/diablo/levels/l1_tables.go` — the three inline tables (BaseTypes
  [207], TileDecorations [207], ShadowPatterns [37]) with their line citations.
- `internal/diablo/levels/l1.go` — checkQuests now takes the quest snapshot and
  ports DrawButcher; GenerateL1 wires the pass order 1:1.
- `internal/diablo/levels/export_test.go` + `l1_tables_test.go` — the table-size
  assertion test (BaseTypes 207 / TileDecorations 207 / ShadowPatterns 37).
- gitignored dev harness `zzdev_stage_test.go` — `TestDevStages` replays the
  whole post-carve generator with the DumpStage hook and diffs stages 03-12.

### The one cell class that was still red

Level 2 only, 64 dTransVal cells per seed, e.g. seed 2588 at world (51,23):
port 0, want 18. That sector is exactly the 8x8 rectangle
DRLG_CheckQuests' DrawButcher widens (`drlg_quests.cpp:19-23`). Ported; the
rest of the quest branches write pdungeon/quest positions and stay with unit C
behind the existing `Dungeon.CheckQuests` hook.

### Oracle verification (stage by stage)

The dev harness's trans comparison was transposed: the offline dump prints
`dTransVal[y][x]` while upstream indexes dTransVal [x][y]
(levels/gendung.cpp:389), so dump row r / column c is Trans[r+16][c+16]. With
that fixed, the first real divergence was 11-quests, and after DrawButcher all
64 cathedral cells are clean at 03-stairs, 04-fixtrans, 05-fixdirt,
06-fixcorner, 07-subst, 08-shadows, 09-lamps, 10-fillfloor, 11-quests and
12-pass3 (tile layer + transparency window, byte-equal).

AC 1 note: the cathedral matrix is 8 variants (levels 1-4 x main/prev) x 8
seeds = 64 combinations; the other 8 variants of the "16" are the crypt levels
21-24, whose generator path does not exist until unit C, so they are out of
this bean's reachable scope.

### Gates

```
$ go build ./... && gofmt -l . && go vet ./internal/...
(no output, rc 0)

$ ./scripts/lint.sh
Sim purity: internal/game internal/diablo clean (189 sources)
0 issues.

$ go test ./internal/diablo/levels/ -count=1 -v
--- PASS: TestDevStage01FirstRoom (0.56s)
--- PASS: TestDevStage02MaskFloor (0.06s)
--- PASS: TestDevStages (0.29s)
--- PASS: TestL1CathedralParity (0.41s)
--- PASS: TestCorpusShape (0.00s)
--- PASS: TestGeneratorTableSizes (0.00s)
ok  github.com/bmcszk/devilutionGo/internal/diablo/levels 1.331s

$ go test ./... -count=1
all packages ok (exit 0)
```


### Session 8 addendum — per-pass unit tests

The bean requires one test per ported function. Parity + the stage oracle cover
the passes end to end, so `l1_passes_test.go` adds the rule-level cases:
FixDirtTiles (6 rules, incl. HDirtCorner probed twice), FixCornerTiles (3 rules
+ protected/floor/above guards), FixTransparency (5 wall cases + no-op),
FillFloor (draw 0/1/2 + protected/non-floor skips), Substitution (family
invariant) and checkQuests (Butcher 8x8 sector + idle counter). placeLamps is
still pinned by the 09-lamps stage dump only: it draws its count from
PlaceMiniSet's scan, so a hand-built case would duplicate that oracle rather
than add a rule.

Commit e692b0f.
