---
# devilGo-v1hy
title: Bitmap text renderer (CLX fonts)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T21:13:42Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# Text

## Goal
DrawString equivalents: fonts, wrap, colors, formatters.

## Research
Source/engine/render/text_render.cpp:291-359 (CountNewlines/DrawStringFormatArg), fonts from devilutionx.mpq + fonts.mpq priority 9200; test text_render_integration_test.cpp + fixtures.

## Scope
internal/render/text.go: font cache per (font, color TRN), wrap algorithm matching C++ widths.

## Acceptance criteria
- [ ] Integration fixture outputs hash-equal (CI)
- [ ] Wrap widths identical for 20 sample strings

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — glyph SubImage draws

## Dependencies & blockers
dg-task-007.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Font atlas per size; colour via TRN palette rows (G7); wrapping/spacing flags match `text_render.cpp` behavior; CJK fallback path documented per G7.
