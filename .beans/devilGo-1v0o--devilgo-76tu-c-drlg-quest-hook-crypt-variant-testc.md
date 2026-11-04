---
# devilGo-1v0o
title: 'devilGo-76tu C: DRLG quest hook, crypt variant, TestCryptParity gate'
status: completed
type: task
created_at: 2026-11-04T04:08:38Z
updated_at: 2026-11-04T04:08:38Z
parent: devilGo-97mn
blocked_by:
    - devilGo-gico
---

Subtask of devilGo-76tu, depends on the fix passes (devilGo-gico). Scope: the quest hook, the crypt variant, and the crypt parity gate.

## Goal
1. Port `DRLG_CheckQuests` so the `11-quests` stage dump equals the port state and the committed cathedral goldens match exactly (this closes `TestL1CathedralParity`).
2. Port the crypt differences of the same generator and wire a `TestCryptParity` gate over the committed l21-l24 fixtures (already in internal/testdata/drlg, currently unreferenced by any test).

## Upstream reference (READ-ONLY)
/tmp/dvlx-pinned = devilutionX @ ca91d778f.
- DRLG_CheckQuests: `grep -rn 'DRLG_CheckQuests' Source/` (quests.cpp) — the cathedral/crypt branches it takes for levels 1-4 and 21-24.
- Crypt-only functions in Source/levels/drlg_l1.cpp: FixCryptDirtTiles, CryptSubstitution, PlaceCryptStairs, PlaceCryptLights, SetCryptSetPieceRoom, plus `DRLG_LPass3` in Source/levels/gendung.cpp (Pass3 is part of CreateL5Dungeon at drlg_l1.cpp:1358-1372, so crypt parity needs it).
- Crypt set-piece data: nlevels/l5data/*.dun, already committed as internal/testdata/drlg/setpieces/*.csv.

## Requirements
- 1:1 port, `file:line` citations; unexported methods on *Dungeon in the landed style.
- `TestCryptParity` mirrors `TestL1CathedralParity`: levels 21-24 x main/prev x 8 seeds, byte equality of tiles/trans/view; it must FAIL (not skip) on a missing fixture.
- Every ported function gets at least one test; table-driven where multi-case.
- No new deps, no go.mod edits, no committed-fixture edits, no .beans writes, no git state changes.

## Acceptance criteria
- [x] AC 1: `go test ./internal/diablo/levels/ -count=1` green: TestL1CathedralParity (l1-l4) AND TestCryptParity (l21-l24) plus the dev-stage oracles 01..11 where /tmp/drlg-stages exists.
- [x] AC 2: `gofmt -l internal/diablo/levels` empty; `go vet ./internal/...` rc0; `./scripts/lint.sh` 0 issues; `go build ./...`.
- [x] AC 3: Report: files, citations, gate tails, and any remaining red cell with got/want.

## Stop rules
Never edit a golden. Never weaken an assertion. If crypt cannot close, deliver the cathedral green and report the crypt cell precisely.

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

## Proof of Work (2026-11-05)

### Landed

- `internal/diablo/levels/crypt.go` - `placeCryptStairs` (crypt.cpp:726-749), the
  level-21 town / Hellfire stairs choice (crypt.cpp:731), `fixCryptDirtTiles`
  (crypt.cpp:708-724), `cryptSubstitution` (crypt.cpp:751-767), `cryptDamage`
  (crypt.cpp:770-797), `fillCryptChamber`/`setCryptRoom`/`setCornerRoom`
  (drlg_l1.cpp:1069-1075, crypt.cpp:676-704), `dungeonSetPiece`.
- `internal/diablo/levels/crypt_sets.go` - the crypt minisets and the
  `ReplaceTile` tables (crypt.cpp:20-198, 357-500).
- `internal/diablo/levels/crypt_shadows.go` - `applyCryptShadowsPatterns`
  (crypt.cpp:486-609) as a rule table plus `applyCryptShadowAt`/`stampShadow`.
- `internal/diablo/levels/crypt_parity_test.go` - `TestCryptParity`, l21-l24 x
  main/prev x 8 committed seeds, tiles/trans/view byte equality, `t.Fatalf` on a
  missing fixture (verified: moving `l23-prev-19770182.csv` away fails the
  subtest with `missing committed fixture for l23-prev seed 19770182`).
- `internal/diablo/levels/l1.go` - `GenerateL1` now branches on `d.LevelType`
  (drlg_l1.cpp:1263-1281); the cathedral-only tail moved into `decorateCathedral`.
- `internal/diablo/levels/l1_stairs.go`, `l1_chambers.go` - crypt branches.
- `internal/diablo/levels/state.go` - `EntryTwarPDn` (gendung_defs.hpp:28-38).

### Defect found and fixed in unit B's code

`CanReplaceTile` used the point-grid diagonals for
`Direction::{NorthWest,SouthEast,SouthWest,NorthEast}`. Those directions are the
isometric displacements `{-1,0}`, `{1,0}`, `{0,1}`, `{0,-1}`
(`Source/engine/displacement.hpp:257-263`), so the cathedral screen probed the
wrong four neighbours and the crypt substitution drew the wrong number of RNG
values. Fixed in `miniset.go`, pinned by the table-driven `TestCanReplaceTile`.
All 64 cathedral fixtures stayed byte-identical after the fix.

### Deviations recorded

- `InitCryptPieces`, `PlaceCryptLights`, `SetCryptSetPieceRoom` and
  `DRLG_LPass3` stay deferred for the reason already recorded in devilGo-76tu:
  they write `dPiece`/`dSpecial`/`dLight`, which the committed goldens
  (`dungeon`, `dTransVal`, `ViewPosition`) do not contain and this package does
  not model.
- Level 24 has no stairs down (crypt.cpp:739-748), so its `ENTRY_PREV` run never
  writes `ViewPosition`; the oracle's process-global carry-over is reproduced in
  `TestCryptParity`, which carries the view between variants like the C++ global.

### Gate tails

```
BUILD OK
GOFMT CLEAN
simpurity: internal/game internal/diablo clean (192 sources)
0 issues.
VET OK
ok  github.com/bmcszk/devilutionGo/internal/diablo/levels  1.9s
ok  github.com/bmcszk/devilutionGo/internal/persist  5.131s
beans-lint: OK (165 beans)
```

RED before the crypt branch landed: 65 failing subtests, first deviation walked
back with the stage oracle to `c02-varch5 tiles [12][25]: port 95, oracle 11`,
which located the `CanReplaceTile` direction defect.
