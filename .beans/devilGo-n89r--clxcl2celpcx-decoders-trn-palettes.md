---
# devilGo-n89r
title: CLX/CL2/CEL/PCX decoders + TRN + palettes
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - sprites
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T21:13:42Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
---

# Sprite decode

## Goal
Decode all sprite formats to paletted index frames; convert to *ebiten.Image with palette/TRN variants.

## Research
Source/engine/clx_sprite.hpp:15-27 (header), CL2 RLE per https://github.com/savagesteel/d1-file-formats/blob/master/PC-Mac/CL2.md; load_cel.cpp (CelToClx path), load_clx.cpp, load_pcx.cpp; trn.cpp; palette.cpp:175 LoadPalette; light_tables (16×256) engine/light_tables.hpp; dun tile re-encode ReencodeDungeonCels (levels/reencode_dun_cels.*).

## Scope
internal/assets/{clx,cl2,cel,pcx,trn,palette}.go; SpriteSet type: frames [][]byte indices + built ebiten images per (palette, trn) variant lazily; atlas builder.

## Acceptance criteria
- [ ] Golden decode tests for ≥20 sampled frames per format (indices byte-equal vs C++ dump)
- [ ] TRN variant image equals reference re-palettized pixels
- [ ] Palette load for l1_2.pal etc. from MPQ
- [ ] Decode benchmarks: full plrgfx warrior sheet < 100ms cold

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — image creation rules (NewImageFromImage at load, SubImage at draw)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden files

## Dependencies & blockers
dg-task-006.

## Open questions
- Lazy vs eager TRN variants: eager for uniques (few), lazy for light (16) — benchmark decides.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Decoders byte-exact: fixture CEL/CL2/CLX/PCX → Go decode vs C++ decode → golden PNG hash match (G3 worked examples).
