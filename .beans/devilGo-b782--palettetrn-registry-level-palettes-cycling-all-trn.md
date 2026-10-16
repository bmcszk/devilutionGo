---
# devilGo-b782
title: 'Palette/TRN registry: level palettes, cycling, all TRN application points'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-n89r
---

# Palette & TRN registry

## Goal
Own the *data* side of colour: per-level palette selection (incl. random 1..4 / nest 1..5), all `.trn` files and the exact order they are applied, baked-vs-per-draw TRN classification, and the RNG call that picks the palette.

## Research
- /tmp/dg-graphics.md G4 (TRN census table), G9 (asset manifest rows for palettes/TRNs).
- Source/engine/palette.cpp:175 LoadPalette, :188 LoadPaletteAndInitBlending (skip ranges per leveltype), :196-213 LoadRndLvlPal (town/crypt special cases, `RandomIntBetween(1,4)`, nest `alternateNestArt` +1)
- Source/engine/trn.cpp:29 GetClassTRN (`plrgfx\<class>.trn` via `spriteData.trn`), :47 GetPlayerGraphicTRN, :14/19/24 infra/stone/pause
- Source/monster.cpp:173-192 InitMonsterTRN (255→0 replace quirk, applied to list OR sheet, skip index 1 for Councilor/Magistr/Cabalist/Advocate!), :3318 unique TRN
- Source/player.cpp:2167-2176 (graphic TRN then class TRN order)
- Source/lighting.cpp:214 LoadTrns
- Source/engine/render/text_render.cpp:88 ColorTranslations[19] (3 null slots) and :188 lazy load
- assets/txtdata/monsters/monstdat.tsv column `trnFile`; assets/plrgfx/warrior/whu/whu{fm,lm,qm}.trn (repo ships 20 .trn — verified `find assets -iname '*.trn' | wc -l`)
- /tmp/dg-requirements.md §6 (asset pipeline), Appendix D.5

## Scope
internal/assets/palette.go (Load + naming + random pick with the shared RNG), internal/assets/trn.go (registry: name → 256 bytes → application mode baked/per-draw), golden tests. Feeds dg-task-202's variant cache and dg-task-010/011's draw decisions.

## Acceptance criteria
- [ ] Palette file name selection reproduces C++ for every leveltype including `alternateNestArt` and the RNG advance (asserted with seeded RNG vectors)
- [ ] All 19 ColorTranslations slots resolved with the 3 null entries mapped to 'no TRN'
- [ ] `ClxApplyTrans` equivalent applied in the same order as C++ for players (graphic then class) and monsters (type), and NOT applied for the 4 skipped demon animations (Councilor/Magistr/Cabalist/Advocate index 1)
- [ ] Monster TRN 255→0 quirk covered by a test using a real 256-byte fixture
- [ ] Every `.trn` in the bundled asset set loads with size 256 and round-trips through the registry (count assertion)

## Skills for coder
/home/blaze/.agents/skills/ebiten/resources/SKILL.md — embed + lazy load
/home/blaze/.agents/skills/go-dev/SKILL.md — registry/typed errors, no globals in exported API
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded tables

## Dependencies & blockers
NOTE (fix2): body verified against live mpqfs tree — decoder port scope is correct as written; earlier stamp referenced notes-only edits.
blocked_by dg-task-007. Blocks dg-task-202 (variant cache) and dg-epic-05 item/monster colouring.

## Open questions
- Where does `_pgfxnum` (armour/weapon graphic index) get computed — entity or assets? Cross-repo with dg-task-016.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] All 16 level palettes + cycle lists + TRN application points from G4 implemented; palette registry data-driven (no hardcoded colours).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 1 declared blocks edge missing (b782→66o7); content otherwise fully verified.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-2b graphics-extras task (dg-task-204): palette selection data + TRN registry + application order; draft, blocked_by devilGo-n89r (wired ✓).
2 local: citations verify with unusual precision — palette.cpp:175/188/196-213 ✓ incl. `RandomIntBetween(1,4)`, town/crypt special cases, nest `!alternateNestArt → rv++` ✓; trn.cpp GetClassTRN:29/GetPlayerGraphicTRN:47/infra:14/stone:19/pause:24 ✓ exact; monster.cpp:173-192 InitMonsterTRN 255→0 + skip i==1 for COUNSLR/MAGISTR/CABALIST/ADVOCATE ✓, :3317 unique TRN ✓; player.cpp:2165-2176 graphic-then-class order ✓; lighting.cpp:214 LoadTrns ✓; text_render.cpp:87-88 ColorTranslations[19] with exactly 3 nullptr ✓; 20 .trn on disk ✓; whu{fm,lm,qm}.trn ✓; monstdat.tsv `trnFile` column ✓. All 3 skills exist ✓. One wiring gap: declared "Blocks dg-task-202 (variant cache)" but 66o7 blocked_by [n89r] only — missing b782 edge.
3 online: n/a (no external claims).
4 solution: wire b782→66o7; optionally sharpen the seam with n89r (both name `internal/assets/palette.go`/`trn.go`: n89r = file decode to bytes, b782 = selection/registry/application — state that split in one line each).
5 understanding: engine/game boundary data task feeding 66o7's variant cache and epic-05 colouring; "no hardcoded colours" gate is the right 1:1 guard.
6 repro: `beans show devilGo-66o7 --json | jq .blocked_by`; `sed -n '196,213p' dX/Source/engine/palette.cpp`; `sed -n '173,192p' dX/Source/monster.cpp`.
7 fix: `beans update devilGo-b782 --blocking devilGo-66o7` (+ optional one-line seam note in both b782/n89r bodies).
8 validation: 66o7 `blocked_by` contains devilGo-b782.
9 related: n89r (decode), 66o7 (variant cache), ru6l/j9be (draw decisions).
VERDICT: NEEDS-WORK:wire-deps — 1 declared blocks edge missing (b782→66o7); content otherwise fully verified.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
