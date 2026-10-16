---
# devilGo-b782
title: 'Palette/TRN registry: level palettes, cycling, all TRN application points'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T14:01:03Z
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
