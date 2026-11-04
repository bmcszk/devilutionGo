---
# devilGo-gico
title: 'devilGo-76tu B: DRLG L1 fix passes — FixTransparency/FixDirt/FixCorner/Substitution/Shadows/lamps/FillFloor'
status: todo
type: task
priority: normal
created_at: 2026-11-04T04:08:30Z
updated_at: 2026-11-04T04:09:07Z
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
- [ ] AC 1: Extend the gitignored oracle harness `zzdev_stage_test.go` (skip when /tmp/drlg-stages absent) to diff stages 04..10 for all 16 variants x 8 seeds; green.
- [ ] AC 2: `go test ./internal/diablo/levels/ -count=1` fully green, including `TestL1CathedralParity` (levels 1-4, both entries, 8 seeds) and the dev-stage oracles.
- [ ] AC 3: `gofmt -l internal/diablo/levels` empty; `go vet ./internal/...` rc0; `./scripts/lint.sh` 0 issues; `go build ./...`.
- [ ] AC 4: Report: files, citations, gate tails, and any cell that stays red with got/want.

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
