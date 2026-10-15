---
# devilGo-l5fp
title: Graphics asset manifest, coverage checker and CLX conversion tooling
status: draft
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-15T21:13:42Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
    - devilGo-n89r
---

# Graphics asset manifest & tooling

## Goal
Machine-checked inventory of every graphics asset the renderer needs (family → archive path → format → frame layout), so 'missing art' is a build-time error, not a runtime surprise, and so retail `.cel/.cl2/.pcx` can be pre-converted to `.clx` for the bundled path.

## Research
- /tmp/dg-graphics.md G9 (the table + extension rules + measured counts/sizes), G3 (formats), G1 §extension rule.
- Source/engine/load_cel.hpp:12/14, Source/engine/load_cl2.hpp:23/27, Source/engine/load_pcx.hpp:15/17 (`.clx` vs `.cel/.cl2/.pcx` by build flag)
- Source/diablo.cpp:1359 LoadLvlGFX (all level triples), Source/levels/dun_tile_data.cpp:68/:104 (min/sol), Source/cursor.cpp:445-452 (+ `-widths.txt`), Source/items.cpp:2424, Source/objects.cpp:3700, Source/towners.cpp:110/:780, Source/player.cpp:1530/:2163, Source/monster.cpp:3217, Source/panels/mainpanel.cpp:100, Source/DiabloUI/{button.cpp:34,credits.cpp:189,diabloui.cpp:629-654}.cpp
- Source/engine/assets.cpp:410 LoadMPQ + assets.hpp:309 priorities
- tools/update_bundled_assets.sh (how devilutionX produces the bundled `.clx` set), 3rdParty/mpqfs
- assets/ tree: 100 `.clx`, 20 `.trn`, 2 `.pal`, 76 font `.clx`, `assets/data/inv/objcurs-widths.txt` (verified counts)

## Scope
graphics_manifest.tsv (hand-curated from G9, machine-readable), `cmd/dgo-assets` with subcommands `list`, `coverage`, `convert` (CEL/CL2/PCX → CLX using the Go decoders, with the same width sources as C++), CI job asserting every manifest entry resolves from either the bundled dir or a provided MPQ, and a report of unresolvable names.

## Acceptance criteria
- [ ] Manifest covers all 7 level types (cel/til/min/sol/special/pal), players (6 classes × anims × 8 dirs), monsters (all 113 monstdat rows with a sprite), items (35 vs 43 under spawn/hellfire), objects, towners, missiles, cursors (+widths), fonts (6 sizes × available rows), UI (`ui_art/*`), gendata cutscenes
- [ ] `coverage` exits non-zero on a missing entry and prints the loader file:line that asked for it
- [ ] `convert` output for the 100 bundled `.clx` files is byte-identical to the committed files (strongest possible decoder cross-validation, uses repo assets only)
- [ ] Widths sourcing matches C++ for every family (no hard-coded width where C++ reads a `-widths.txt`)
- [ ] Docs/graphics.md generated from the manifest

## Skills for coder
/home/blaze/.agents/skills/ebiten/resources/SKILL.md — embed + asset lifecycle
/home/blaze/.agents/skills/go-dev/SKILL.md — CLI + table-driven codegen
/home/blaze/.agents/skills/go-setup/SKILL.md — where tooling lives in the tree
/home/blaze/.agents/skills/coding-rules/SKILL.md — data over code

## Dependencies & blockers
blocked_by dg-task-006, dg-task-007. Blocks dg-task-010/dg-task-012/dg-task-013/dg-task-014 content checks ('all art present').

## Open questions
- Where do the `ui_art` PCX names come from for a full enumeration (grep + DevilutionX `devilutionx.mpq` build)?
- Should `convert` be part of a release pipeline or developer-only?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Asset manifest per G9 complete: family→archive→format→frame layout for every sprite family; coverage checker fails CI if any family unimplemented.
