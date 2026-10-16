---
# devilGo-gk5e
title: 'Dungeon tile pipeline: SOL/MIN/TIL load + CEL re-encode + micro renderer'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-vti7
---

# Dungeon tile pipeline

## Goal
Load `*.sol`/`*.min`/`*.til`/`*.cel` for a level type, re-encode dungeon CELs exactly like `ReencodeDungeonCels`, and render micro tiles (6 TileTypes × 4 MaskTypes) at the correct screen positions with the tile's light table.

## Research
- /tmp/dg-graphics.md G2 (byte layouts, worked examples, the MIN index scramble, mask ASCII art, foliage split, re-encode sizes 512/784/1024).
- Source/levels/dun_tile.hpp:19 (TileType doc comments), :83 (LevelCelBlock bits), :106 (TileProperties), :119 (MICROS), :131/132 (ReencodedTriangle/TrapezoidFrameSize)
- Source/levels/dun_tile_data.cpp:68 (LoadMinData), :104 (LoadLevelSOLData + hand-patched SOL entries for L1/L3/L4/Crypt — transcribe the exact indices!), :165 (SetDungeonMicros incl. `blocks - 2 + (block & 1) - (block & 0xE)`)
- Source/levels/reencode_dun_cels.cpp:24-180 (every re-encode helper incl. RenderTransparentSquare and ExtractFoliage*)
- Source/engine/render/dun_render.hpp:28 (MaskType), :123 (GetDunFrame/Swap32LE), :147/:162; Source/engine/render/dun_render.cpp:92-104 (prefix increments), :466/:481 (unrolled triangle line widths/src offsets), :1046 (RenderTileFrame), :1085 (world_draw_black_tile)
- Source/engine/render/scrollrt.cpp:549 (DrawCell incl. mask selection), :680 (DrawFloorTile), :955/:992/:1088 (row walk + DrawOOB/DrawDirtTile)
- /tmp/dg-requirements.md §2 and §7; dg-epic-04 dg-task-015 provides the grids.
- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Scope
internal/world/dungeoncel.go (sol/min/til load + micro scramble + re-encode + frame renumbering), internal/render/tile.go (RenderTileFrame for all 6 types × all masks, bottom-up row walking, foliage pass), atlas upload of decoded micros, `DrawFloor`/`DrawCell`/`DrawOOB`/`DrawDirtTile` wired into the row walk. NOT in scope: lighting tables (dg-task-202), entity order (dg-task-010).

## Acceptance criteria
- [ ] `TestMinScramble`: for blocks∈{10,16} the decoded `LevelCelBlock` list equals a C++ dump for the first 200 pieces of l1 and town
- [ ] `TestReencode`: output buffer size equals `GetReencodedSize` and frame sizes are 512 (tri), 784 (trap), 1024 (square); foliage frames extracted count matches C++
- [ ] `TestRenderTileGolden`: per (TileType × MaskType × lightLevel {0,8,15}) a 64×64 index buffer equals the C++ dump (add a dump gtest to devilutionX test/dun_render_benchmark harness)
- [ ] `TestRowWalk`: 5×5 map produces the identical (tile, bufferX, bufferY) sequence as C++ (dg-epic-03 first criterion, owned here)
- [ ] SOL hand-patches (l1 idx 9,15,16,20,21,27,28,51,56,58,61,63,65,72,208,247,253,257,323,403,24,450; l3 170,481,487; l4 210; crypt 142) are encoded as data and asserted
- [ ] `world_draw_black_tile` and `DrawDirtTile` OOB fallbacks implemented and unit-tested
- [ ] Draw path allocates 0 (testing.AllocsPerRun)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — SubImage/atlas rules for micro tiles
/home/blaze/.agents/skills/go-dev/SKILL.md — index math and bounds-checked byte walking
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden-file comparisons
/home/blaze/.agents/skills/coding-rules/SKILL.md — keep the re-encode helpers 1:1 named with C++ for reviewability

## Dependencies & blockers
blocked_by dg-task-007 (CLX decode) and dg-task-015 (grids). Blocks dg-task-010, dg-task-011.

## Open questions
- Endianness of the dungeon CEL frame table: mirror `Swap32LE` (BE-correct) or read LE? See /tmp/dg-graphics.md G2 open questions; decide with the operator and cover with the BE CI job (tools/run_big_endian_tests.sh).
- Ship retail `.cel` fixtures for CI or only derived dumps?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Tile micro-render matches G2 pipeline: per-tile floor+wall micro pairs, arch crawl order from ` dun_render.cpp:1046` — verified by pixel-diff vs C++ dumper on a seed dungeon frame.
- [ ] Tile atlas page ≤4096×4096; atlas budget documented (count of micro-tiles per dungeon type).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
