---
# devilGo-gk5e
title: 'Dungeon tile pipeline: SOL/MIN/TIL load + CEL re-encode + micro renderer'
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-11-11T13:56:57Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-vti7
    - devilGo-1spo
---

# Dungeon tile pipeline

## Goal
Load `*.sol`/`*.min`/`*.til`/`*.cel` for a level type, re-encode dungeon CELs exactly like `ReencodeDungeonCels`, and render micro tiles (6 TileTypes × 4 MaskTypes) at the correct screen positions with the tile's light table.

## Research
- docs/research/dg-graphics.md G2 (byte layouts, worked examples, the MIN index scramble, mask ASCII art, foliage split, re-encode sizes 512/784/1024).
- Source/levels/dun_tile.hpp:19 (TileType doc comments), :83 (LevelCelBlock bits), :106 (TileProperties), :119 (MICROS), :131/132 (ReencodedTriangle/TrapezoidFrameSize)
- Source/levels/dun_tile_data.cpp:68 (LoadMinData), :104 (LoadLevelSOLData + hand-patched SOL entries for L1/L3/L4/Crypt — transcribe the exact indices!), :165 (SetDungeonMicros incl. `blocks - 2 + (block & 1) - (block & 0xE)`)
- Source/levels/reencode_dun_cels.cpp:24-180 (every re-encode helper incl. RenderTransparentSquare and ExtractFoliage*)
- Source/engine/render/dun_render.hpp:28 (MaskType), :123 (GetDunFrame/Swap32LE), :147/:162; Source/engine/render/dun_render.cpp:92-104 (prefix increments), :466/:481 (unrolled triangle line widths/src offsets), :1046 (RenderTileFrame), :1085 (world_draw_black_tile)
- Source/engine/render/scrollrt.cpp:549 (DrawCell incl. mask selection), :680 (DrawFloorTile), :955/:992/:1088 (row walk + DrawOOB/DrawDirtTile)
- docs/research/dg-requirements.md §2 and §7; dg-epic-04 devilGo-vti7 provides the grids.
- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Scope
internal/world/dungeoncel.go (sol/min/til load + micro scramble + re-encode + frame renumbering), internal/render/tile.go (RenderTileFrame for all 6 types × all masks, bottom-up row walking, foliage pass), atlas upload of decoded micros, `DrawFloor`/`DrawCell`/`DrawOOB`/`DrawDirtTile` wired into the row walk. NOT in scope: lighting tables (devilGo-66o7), entity order (devilGo-ru6l).

## Acceptance criteria
- [ ] `TestMinScramble`: for blocks∈{10,16} the decoded `LevelCelBlock` list equals a C++ dump for the first 200 pieces of l1 and town
- [x] `TestReencode`: output buffer size equals `GetReencodedSize` and frame sizes are 512 (tri), 784 (trap), 1024 (square); foliage frames extracted count matches C++
- [x] `TestRenderTileGolden`: per (TileType × MaskType × lightLevel {0,8,15}) a 64×64 index buffer equals the C++ dump (add a dump gtest to devilutionX test/dun_render_benchmark harness)
- [x] `TestRowWalk`: 5×5 map produces the identical (tile, bufferX, bufferY) sequence as C++ (dg-epic-03 first criterion, owned here)
- [x] SOL hand-patches (l1 idx 9,15,16,20,21,27,28,51,56,58,61,63,65,72,208,247,253,257,323,403,24,450; l3 170,481,487; l4 210; crypt 142) are encoded as data and asserted
- [x] `world_draw_black_tile` and `DrawDirtTile` OOB fallbacks implemented and unit-tested
- [ ] Draw path allocates 0 (testing.AllocsPerRun)

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — SubImage/atlas rules for micro tiles
/home/blaze/.agents/skills/go-dev/SKILL.md — index math and bounds-checked byte walking
/home/blaze/.agents/skills/go-tests/SKILL.md — golden-file comparisons
/home/blaze/.agents/skills/coding-rules/SKILL.md — keep the re-encode helpers 1:1 named with C++ for reviewability

## Dependencies & blockers
blocked_by devilGo-n89r (CLX decode) and devilGo-vti7 (grids). Blocks devilGo-ru6l, devilGo-j9be.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: CEL frame-table endianness read from the C++ loader (Swap32LE calls) as step 1 — port whatever the C++ does, BE CI job covers it.
- Endianness of the dungeon CEL frame table: mirror `Swap32LE` (BE-correct) or read LE? See docs/research/dg-graphics.md G2 open questions; decide with the operator and cover with the BE CI job (tools/run_big_endian_tests.sh).
- Ship retail `.cel` fixtures for CI or only derived dumps?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [x] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [x] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Tile micro-render matches G2 pipeline: per-tile floor+wall micro pairs, arch crawl order from ` dun_render.cpp:1046` — verified by pixel-diff vs C++ dumper on a seed dungeon frame.
- [ ] Tile atlas page ≤4096×4096; atlas budget documented (count of micro-tiles per dungeon type).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Graphics-wave task (devilGo-gk5e): the heaviest renderer bean — tile data load, re-encode, micro render. Wire: blockedBy [n89r, vti7] — matches body (devilGo-n89r + devilGo-vti7) ✓; dependents pdpc/m18k/7aul wired ✓.
2 local: STRONG verification: dun_tile.hpp:83 LevelCelBlock, :106 TileProperties, :119 MICROS, :131/:132 Reencoded{Triangle=544-32, Trapezoid=800-16}FrameSize ✓ (=AC's 512/784; square 1024 ✓); dun_tile_data.cpp LoadLevelSOLData region :104+ with the SOL hand-patch list VERIFIED INDEX-EXACT (l1: 9,15,16,20,21,27,28,51,56,58,61,63,65,72,208,247,253,257,323,403,24,450; l3: 170,481,487; l4: 210=None; crypt/l5: 142=None — every index matches the bean's list); SetDungeonMicros :166 with blocks=10/16 ✓ (TestMinScramble params). reencode_dun_cels.cpp:24+, 322 lines ✓. scrollrt.cpp:549 DrawCell, :680 DrawFloorTile, :1088 DrawOOB/:1047 DrawDirtTile ✓. MINOR DRIFT: RenderTileFrame at :1040 (bean/gate say :1046), world_draw_black_tile at :1080 (bean :1085). Skills 4/4 ✓. tools/run_big_endian_tests.sh exists (open question's CI hook) ✓.
3 online: n/a (G2 doc is local).
4 solution: none blocking — cite refresh only. Open questions (CEL frame-table endianness; retail .cel fixtures vs derived dumps) are genuine operator decisions, correctly parked with a CI hook.
5 understanding: engine/render + world seam; owns TestRowWalk (epic's first criterion) — declared "owned here"; feeds pdpc/m18k/7aul.
6 repro: `sed -n '104,165p' Source/levels/dun_tile_data.cpp | grep -n "SOLData\[" | head -30; grep -n "void RenderTileFrame\|world_draw_black_tile" Source/engine/render/dun_render.cpp; beans query --json '{ bean(id:"devilGo-gk5e") { blockedByIds blockingIds } }'`
7 fix: optional: `beans update devilGo-gk5e --body-replace-old ":1046 (RenderTileFrame), :1085 (world_draw_black_tile)" --body-replace-new ":1040 (RenderTileFrame), :1080 (world_draw_black_tile)"` (and same numbers in the gate block).
8 validation: cites match live lines; SOL-patch AC list still index-exact (it is today).
9 related: n89r, vti7 (deps); pdpc, m18k, 7aul (dependents); ru6l (row-walk seam, declared), j9be.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

## Zero-dependency test constraint\n\nParity tests compare committed Go-owned fixtures only. C++ dumpers/probes may be used once in an offline producer workflow, but must never run in go test, CI, production code, or ordinary fixture verification. No DG_DEVILUTIONX or external source-tree access; missing fixtures are hard failures.

## Progress log

- Re-audited 2026-11-11 (read-only reconciliation wave vs master 2a91b97): STANDS, M1-critical — the head of the remaining render chain. Scope trim adopted (audit item 4): the row-walk TDD AC re-points to asserting the C++ (tile, bufferX, bufferY) golden sequence against the ALREADY-LANDED #53 walk (internal/render/world.go); the OOB-fallback AC is already satisfied (world_test.go:359/:375/:390 — cite them, do not reimplement). Claimed parent: implementation lane pending.

## Proof of Work — wave slice (parent verification, 2026-11-12, feat/tiles @ 7ef7c82)

- Commits: 34fecc4 W1 SOL load+hand-patches, 9d8ccfa W2 MIN scramble, cacea32 W3 re-encode 512/784/1024, ba26775 W4 render-tile micros, a5e4e07 W5 renumber upper-bound fix (review F1), 862d371 W6 black-tile clip guards (F2), 093e3bd W7 InitPrefix (F3), 4eb786b W8 default-no-op UB note (F7), 97d7aa4 W9 sha256 manifest (F5), 7ef7c82 W10 all-49-section asserts (F4). Pushed: origin/feat/tiles == 7ef7c828647cfde8cc6f10cdc48bcec9c69f9aea (ls-remote verified).
- Parent gates at 7ef7c82: go build OK; gofmt empty; lint.sh 0 issues (sim purity clean); `go test ./internal/render/...` 86 pass; `go test ./...` 11,438 pass / 45 pkgs; sqlc no-raw-SQL clean; beans-lint OK (184); porcelain clean.
- Review round: brainer REQUEST-CHANGES (7 findings: F1 blocker renumber no-op, F2 major black-tile clip panic, F3-F7 minor/nit) → fix wave W5-W10 closed all seven; no findings left open.
- The W3 hang (inherited WIP) root-caused — 5 fixes (two in the test helpers' accumulator/loop-var bugs, the rest in the implementation), validated against verbatim upstream reencode_dun_cels.cpp; the interrupted dun_tile_data.go change was adopted (IsFloor/IsFloorLeft per reencode_dun_cels.hpp:23-28).
- Scope-trim verdict for the row-walk AC (audit item): the C++ (tile, bufferX, bufferY) sequence is asserted by the LANDED #53 walk tests (internal/render/world_test.go:253 TestDrawFloorWalksTheIsoDiamond; :359/:375/:390 OOB dirt/black/town-gap) — cited, not reimplemented.

## Deviations recorded (fidelity gate)

- The offline C++ dumper was sed-pinned (per-pixel-option read → false, mirroring the benchmark harness InitOnce) at /tmp/dvlx-pin; the patch is recorded verbatim in .hermes/lane/dg-tiles/tiles.w3.log and the producer source ships committed at internal/testdata/render/producer/.
- Fixture provenance rests on the committed producer + the W9 sha256 manifest test (TestTilesFixtureManifest) — tampering fails the suite.
- MaskType out-of-range default is a documented silent no-op (upstream UB, one-line note in tile.go) instead of a panic — F7 disposition.

## Still open (stays in-progress; units named)

- TestMinScramble C++ fixture extension (needs a 200-piece l1+town dump from the dumper)
- Draw path alloc-0 (testing.AllocsPerRun benchmark)
- G2 arch-crawl pixel-diff on a seed dungeon frame
- Tile-atlas page budget documentation
These have no owner lane; each is pick-up-ready with the dumper/producer chain committed.


ADR-020 executable-count addendum (2026-11-12, user-enforced): W1's incidental `package main` under internal/testdata/render/producer/go-tiles was a fourth executable surface — folded into tools/goldens' `tiles` subcommand (commit 51abbc3, PR #58), the stray main deleted; corpus re-generated 20/20 byte-identical vs the committed sha256 manifest. cmd/ stays two binaries; producers live only under tools/.
