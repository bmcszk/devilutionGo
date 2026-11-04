---
# devilGo-76tu
title: DRLG L1 Cathedral + Crypt variant
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-04T02:25:13Z
parent: devilGo-97mn
---

# L1 generator

## Goal
CreateL5Dungeon parity incl. crypt branch.

## Research
drlg_l1.cpp:1298-1312 entry; internals :1200-1294 (Fix*, Substitution, lamps, FillFloor, SelectChamber); crypt :1214,1221; Pass3/DRLG_LPass3 gendung.h:88; fixtures test/fixtures/diablo + levels/l1data.

## Scope
internal/world/drlg_l1.go + crypt.go sharing engine.

## Acceptance criteria
- [ ] All diablo L1 fixtures byte-equal
- [ ] Hellfire crypt fixtures byte-equal
- [ ] Quest chamber selection matches SelectChamber logic for all availability combos

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-tests/SKILL.md — fixture matrix
/home/blaze/.agents/skills/go-dev/SKILL.md — RNG-call-order discipline

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md

## Dependencies & blockers
devilGo-vti7, devilGo-m2x7.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 generator task (devilGo-76tu): CreateL5Dungeon parity incl. crypt; draft; wire blockedBy [vti7] ✓; body's devilGo-m2x7 (m2x7 RNG) is transitively covered (vti7→m2x7 chain verified).
2 local: drlg_l1.cpp:1298-1312 CreateL5Dungeon ✓ exact (sed match, incl. crypt branch `if (leveltype == DTYPE_CRYPT)`); crypt internals :1214 (FixCryptDirtTiles/CryptSubstitution) ✓; SelectChamber at :1270 ✓ (body ":1268-1294" ✓); fixtures live: test/fixtures/diablo 27 .dun + hellfire 17 .dun ✓, levels/l1data ✓, drlg_l1_test.cpp ✓. One cite drift: `DRLG_LPass3 gendung.h:88` — actual gendung.h:95.
3 online: n/a.
4 solution: none blocking — refresh one line number.
5 understanding: world layer core; consumes vti7 grids/miniset + m2x7 RNG call-order; feeds 4ybb/yeia/foug; fixture-byte-equality is exactly the epic's parity gate.
6 repro: `sed -n '1298,1312p' devilutionX/Source/levels/drlg_l1.cpp; grep -n "DRLG_LPass3" devilutionX/Source/levels/gendung.h; ls devilutionX/test/fixtures/diablo/*.dun | wc -l`
7 fix: optional `beans update devilGo-76tu --body-replace-old "Pass3/DRLG_LPass3 gendung.h:88" --body-replace-new "Pass3/DRLG_LPass3 gendung.h:95"`
8 validation: grep shows :95 after edit.
9 related: vti7 (dep), 4ybb/yeia (dependents), foug (population).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

## Progress log

### CLAIM 2026-11-03 (lane devilGo-76tu, worktree dg-lane-76tu @ master 4b8c13e)

**Blocked-by cleanup:** removed `blocked_by: [devilGo-vti7, devilGo-1spo]` from the
frontmatter. Both ids are phantom refs: `beans list --json` has no `devilGo-vti7`
(triaged/closed and archived) and `devilGo-1spo` is likewise archived; neither can
unblock or be waited on. Real dependencies are satisfied on master: engine LCG
(`internal/engine/random.go`, golden vectors from the C++ probe), `internal/game/world`
grid/tile-properties, and quest availability under `internal/diablo/gamedata/quests`.

**Layout decision (Scope repoint):** `internal/world/drlg_l1.go` does not exist and
`internal/world` is not a layer in AGENTS.md. DRLG L1 is Diablo-specific data + algorithm,
so it lands in a new package **`internal/diablo/levels`** (AGENTS.md: "internal/diablo/ …
levels: DRLG, town, dungeon tile data"), importing `internal/engine` (RNG) and
`internal/game/world` (grid/tile props) only. Files: `dungeon.go` (DRLG state, Miniset,
transparency), `drlg_l1.go` (cathedral), `crypt.go` (crypt variant), tests + fixtures.

**Fixture plan:** committed text goldens only, no MPQ reads in tests.
- New committed corpus `internal/testdata/drlg/l<level>-<entry>-<seed>.csv` in the existing
  dun CSV shape (dims / 40x40 tile layer / 80x80 dTransVal window / ViewX,ViewY,leveltype),
  produced ONCE offline by extending `tools/probe/dvg_probe.cpp`'s `dun` variant list with the
  Hellfire crypt levels 21-24 and dumping **before** `CreateThemeRooms()` (theme rooms are
  bean devilGo-foug's scope; the pre-existing `internal/testdata/dun/` corpus stays untouched).
- Levels: l1 + l21..l24, entries main/prev, the 8 pinned seeds.
- Producer run is recorded in the PoW with sha256 + counts; the Go tests read only the
  committed CSVs and fail (never skip) when one is missing.

**Deferrals (recorded, not silent):** `Pass3`/`DRLG_LPass3` needs the `*.min` mega-tile
tables (asset data owned by bean devilGo-gk5e, "dungeon tile pipeline sol mintil"); the
post-Pass3 calls `PlaceCryptLights`/`SetCryptSetPieceRoom` need dPiece and `DoLighting`.
GenerateLevel's own `dRLG_CheckQuests` tail is in scope.

### WIP 2026-11-04 (lane session 2, continuation after wall-timeout)

**Committed:** goldens + producer (`294ea3f`), cathedral L2-4 extension (`c4faad2`),
sibling foundation bean (`b3c57b5`). **Uncommitted:** the whole
`internal/diablo/levels/` package (12 files) — `git status` shows `?? internal/diablo/levels/`
and the pre-commit lint hook refuses to commit it yet (155 `unused` consts because the
generator body is still a stub, plus decorder/funcorder/intrange/errcheck debt).

**Verified this session (commands + output):**
- `sha256sum -c manifest.txt` over `internal/testdata/drlg/` → 0 mismatches (128 fixtures).
- `go build ./...` → ok.
- `go test ./internal/diablo/levels/ -run TestCorpusShape` → ok (128 files, 8 seeds).
- `go test ./internal/diablo/levels/ -run TestL1CathedralParity` → RED on every variant:
  `l1-main-seed-2588: dungeon tile (0,0) = 0, want 22` — `GenerateL1` returns nil.

**Root-cause state of the port (read against /tmp/dvlx-pinned @ ca91d778f):**
- Foundation is ported and compiles: `state.go` (dungeon/pdungeon/dTransVal/Protected/
  Chamber/DungeonMask/ViewPosition/SetPiece), `bitset.go`, `geom.go`, `miniset.go`
  (`Miniset` match/place, `CanReplaceTile` incl. the p2 workaround), `placement.go`
  (`PlaceDunTiles` gendung.cpp:503-521, `PlaceMiniSet` :459-501 with the drlg1 x/y<=12
  quirk, `PlaceMiniSetRandom` drlg_l1.cpp:1251-1268), `trans.go` (DRLG transparency
  window), both tile vocabularies.
- Miniset dimension note from the killed session confirmed against source: `STAIRSUP`
  is `{4,4}` → **16** cells per grid (rows-descend, `search[y][x]`), `STAIRSDOWN`
  `{4,3}` → 12, `LAMPS` `{2,2}` → 4, `PWATERIN` `{6,6}` → 36. The `[6][6]Tile`
  storage is the upstream maximum, not the cell count; fixtures are encoded with
  `len(records[1]) == width*height` (l1_parity_test.go `encodeDun`).
- Missing: `GenerateLevel` body (drlg_l1.cpp:1315-1345) + the whole pass chain
  (`FirstRoom`/`GenerateRoom`/`MapRoom`/`CheckRoom`/`MakeDmt`/`AddWall`/
  `FillChambers`/`FixTilesPatterns`/`Substitution`/`ApplyShadowsPatterns`/
  `FillFloor`/`FixTransparency`/`FixDirtTiles`/`FixCornerTiles`/`PlaceCathedralStairs`/
  `InitSetPiece`/`SelectChamber`/`DRLG_CheckQuests`), the crypt variant (crypt.cpp),
  and `ShadowPatterns`/`BaseTypes`/`TileDecorations` data tables (extracted from the
  pinned source mechanically, not by hand).

**Producer inputs pinned for parity** (tools/probe/dvg_probe.cpp:325-349):
`pOriginalCathedral=true`, `fullQuests=1`, `gbIsMultiplayer=false`,
`randomizeQuests=false`, `LevelSeeds[level]=nullopt`, `currlevel=level`,
`leveltype=GetLevelType(level)`; goldens dumped BEFORE `CreateThemeRooms`.

**Deferral kept honest:** `Pass3`/`DRLG_LPass3` (needs `*.min` mega-tile tables,
bean devilGo-gk5e) and `PlaceCryptLights`/`SetCryptSetPieceRoom` (need dPiece +
DoLighting) stay out — they write `dPiece`/`dLight`, which the committed goldens do
not contain. `DRLG_CheckQuests` stays in scope because `DrawButcher`'s
`DRLG_RectTrans` writes `dTransVal`, which the goldens DO contain.

## PROGRESS session 3 (2026-11-04)

**Salvage landed.** `3df38c6` commits the foundation session 2 left staged (Dungeon
state, Bitset2d, geom, miniset/placement, trans, both tile vocabularies, the L1
parity corpus loader/tests). Commit is honest about the RED state: `GenerateL1`
has no body, so every `TestL1CathedralParity` variant fails on tile (0,0).
Lint debt had to be cleared first — the pre-commit hook runs `scripts/lint.sh`,
which rejected the staged tree with 198 findings (155 of them `unused` on the
crypt vocabulary; block-level `//nolint:stylecheck,unused` with the 1:1-port
exemption, plus reorders/`range n`/phase comments). No assertion changed.
`git push` is BLOCKED by the pre-push gate (it runs the full suite); the branch
can only be pushed once parity is green.

**Offline stage oracle (dev-only, /tmp, NOT committed).** A patched copy of the
pinned C++ tree (`/tmp/dvlx-pinned`, pinned commit ca91d778f) now dumps the
generator state after every pass via `DumpStage`, gated by `DVLX_STAGE_PREFIX`:
`/tmp/drlg-stages/l<level>-<entry>-<seed>-<pass>.txt` for the 128 committed
variant/seed combinations, passes `01-firstroom-a<N>`, `02-predstairs-a<N>`,
`03-stairs` .. `11-quests`, `12-pass3`. Fidelity of the oracle is PROVEN: the
`11-quests` dump of all 128 combinations equals the committed CSV goldens byte
for byte, and the same build still reproduces the committed `.dun` artifacts.
Layout: tiles are flat `dungeon[DMAXX][DMAXY]` memory order (row r, col c =
`dungeon[x=r][y=c]`), trans rows are `dTransVal[y][16..95]`; the loader mapping
is `golden.tiles[x][y] == d.Tiles[x][y]`, `golden.trans[y][x] ==
d.Trans[x+16][y+16]`, so `dump[r][c_tr] == d.Trans[c_tr+16][r+16]`.
No repo file changed for this; `tools/probe/dvg_probe.cpp` was copied to /tmp
before editing, so the committed producer is untouched.

**Units.** A1 = room layout (`MapRoom`/`CheckRoom`/`GenerateRoom`/`FirstRoom`/
`FindArea`, drlg_l1.cpp:434-560) verified against `01-firstroom-a1`; A2 =
`InitDungeonFlags`/`MakeDmt`/`FillChambers`/`FixTilesPatterns`/`AddWall`/
`FloodTransparencyValues`/`PlaceCathedralStairs` + the `GenerateLevel` retry
loop, verified against `02-predstairs`/`03-stairs`; B = `FixTransparency`/
`FixDirtTiles`/`FixCornerTiles`/`Substitution`/`ApplyShadowsPatterns`/lamps/
`FillFloor`, verified against `04..10`; C = `DRLG_CheckQuests` (`11-quests`,
the final golden state) and `ViewPosition`. Pass3/`12-pass3` stays deferred for
the reason already recorded above (it writes dPiece, absent from the goldens).
