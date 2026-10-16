---
# devilGo-n89r
title: CLX/CL2/CEL/PCX decoders + TRN + palettes
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - sprites
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:16:43Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-1 decoder task (dg-task-007); draft, blocked_by devilGo-vof6 (wired ✓); all 6 downstream dependents wired (ru6l, mlv2, v1hy, 66o7, b782, l5fp all blocked_by n89r ✓).
2 local: `dX/Source/engine/clx_sprite.hpp:15-27` CLX 6-byte frame header + CL2 ref URL (cited URL exists inside the C++ header itself, line 22) ✓; load_cel/clx/pcx.cpp, trn.cpp, palette.cpp (`:175 LoadPalette` ✓), reencode_dun_cels.* all exist ✓; light_tables.hpp NumLightRadiuses=16 ✓. Skills ebiten/rendering + go-fluent-testing exist ✓. TDD-first golden-decode AC (≥20 frames/format, byte-equal vs C++ dump) present and testable; G10/G11 gate present.
3 online: n/a (CL2.md URL is quoted from the in-tree header, not an independent claim).
4 solution: none — ready.
5 understanding: engine-layer asset decode producing paletted index frames + lazy (palette,TRN) ebiten variants; foundation for every renderer/UI task.
6 repro: `sed -n '15,27p' dX/Source/engine/clx_sprite.hpp`; `ls dX/Source/engine/load_{cel,clx,pcx}.cpp dX/Source/levels/reencode_dun_cels.cpp`; `beans show devilGo-ru6l --json | jq .blocked_by`.
7 fix: none.
8 validation: n/a.
9 related: vof6 (bytes source), l5fp (manifest), b782 (palette/TRN data), ru6l/66o7 (consumers).
VERDICT: TRIAGED-OK.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
