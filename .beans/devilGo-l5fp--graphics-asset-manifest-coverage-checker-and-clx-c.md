---
# devilGo-l5fp
title: Graphics asset manifest, coverage checker and CLX conversion tooling
status: todo
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
    - devilGo-n89r
    - devilGo-fpkz
---

# Graphics asset manifest & tooling

## Goal
Machine-checked inventory of every graphics asset the renderer needs (family → archive path → format → frame layout), so 'missing art' is a build-time error, not a runtime surprise, and so retail `.cel/.cl2/.pcx` can be pre-converted to `.clx` for the bundled path.

## Research
- docs/research/dg-graphics.md G9 (the table + extension rules + measured counts/sizes), G3 (formats), G1 §extension rule.
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/ebiten/resources/SKILL.md — embed + asset lifecycle
/home/blaze/.agents/skills/go-dev/SKILL.md — CLI + table-driven codegen
/home/blaze/.agents/skills/go-setup/SKILL.md — where tooling lives in the tree
/home/blaze/.agents/skills/coding-rules/SKILL.md — data over code

## Dependencies & blockers
blocked_by devilGo-vof6, devilGo-n89r. Blocks devilGo-ru6l/devilGo-mlv2/devilGo-zdj7/devilGo-v1hy content checks ('all art present').

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: ui_art PCX enumeration from devilutionX devilutionx.mpq build manifest (one-time, recorded in bean). convert = developer-only tool (just recipe), not release pipeline.
- Where do the `ui_art` PCX names come from for a full enumeration (grep + DevilutionX `devilutionx.mpq` build)?
- Should `convert` be part of a release pipeline or developer-only?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Asset manifest per G9 complete: family→archive→format→frame layout for every sprite family; coverage checker fails CI if any family unimplemented.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC-count-errors + wire-deps — 3 declared blocks edges missing; monstdat count contradicts V12.12/disk; font count off (76 vs 64). (Note: l5fp is grouped in batch-03 per the BATCHES split; batch-05's "if distinct" hedge resolves to "skip there".)


## IMPL NOTES (triage 2026-10-16)
1 root: wave-2b graphics-extras task (devilGo-l5fp): manifest + `cmd/dgo-assets` coverage/convert tooling; draft, blocked_by vof6+n89r (wired ✓).
2 local: G9 exists (`dg-graphics-part2.md:517`) ✓; load_cel.hpp:11-14 / load_cl2.hpp:21-29 / load_pcx.hpp:13-19 UNPACKED `.clx` ext rules ✓; diablo.cpp:1359 LoadLvlGFX ✓; cursor.cpp:443-452 widths.txt ✓ + `assets/data/inv/objcurs-widths.txt` on disk ✓; asset counts verified exact: 100 .clx, 20 .trn, 2 .pal ✓; tools/update_bundled_assets.sh ✓; player.cpp:1528-1532 / monster.cpp:3217 / towners.cpp:110 / DiabloUI/diabloui.cpp:629-634 ✓. Three gaps: (a) AC says "all 113 monstdat rows" — disk = 112 data rows (113 physical lines incl. header); V12.12 (append.md:3738) says 112 — same error family as zvwd; (b) research bullet "76 font .clx" — disk count = 64 (`fonts/` 59 + `fonts/tr/` 5); (c) declared blocks 010/012/013/014 not wired: ru6l(010)/mlv2(012)/v1hy(014) blocked_by [n89r] only, zdj7(013) blocked_by [ru6l].
3 online: n/a (no external claims).
4 solution: fix monstdat count 113→112 in AC (or state "113 physical lines incl. header" explicitly); correct font count to 64 (or cite the G9 counting rule); wire l5fp→ru6l/mlv2/v1hy (zdj7 optional, transitive via ru6l).
5 understanding: engine-layer build-time asset gate; the "convert bundled .clx byte-identical" AC is the strongest decoder cross-validation in the tree and only uses repo assets — good v1 fit with DECIDED #2 (extract-at-build).
6 repro: `awk 'NR>1' dX/assets/txtdata/monsters/monstdat.tsv | wc -l` (=112); `find dX/assets -ipath '*font*' -iname '*.clx' | wc -l` (=64); `for t in devilGo-ru6l devilGo-mlv2 devilGo-v1hy; do beans show $t --json | jq .blocked_by; done`.
7 fix: `for t in devilGo-ru6l devilGo-mlv2 devilGo-v1hy; do beans update devilGo-l5fp --blocking $t; done` + body edit: "113 monstdat rows" → "112 monstdat rows (113 physical lines incl. header)"; "76 font .clx" → "64 font .clx on disk".
8 validation: `beans show <t> --json | jq .blocked_by` contains devilGo-l5fp; grep body for "112 monstdat".
9 related: n89r (decoders used by convert), vof6 (resolution source), ru6l/mlv2/zdj7/v1hy (content-check consumers).
VERDICT: NEEDS-WORK:AC-count-errors + wire-deps — 3 declared blocks edges missing; monstdat count contradicts V12.12/disk; font count off (76 vs 64). (Note: l5fp is grouped in batch-03 per the BATCHES split; batch-05's "if distinct" hedge resolves to "skip there".)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


