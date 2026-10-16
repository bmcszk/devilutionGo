---
# devilGo-v1hy
title: Bitmap text renderer (CLX fonts)
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:06:56Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Original-tree task (dg-task-014): DrawString equivalents, fonts, wrap, colors, formatters. Wire: blockedBy [n89r] — matches body "dg-task-007" ✓.
2 local: Cites verified: text_render.cpp:291 CountNewlines ✓ (read :291-300, DrawStringFormatArg follows); fonts priority 9200 = assets.hpp:309-312 (verified in batch-01); test/text_render_integration_test.cpp + fixtures/*.png exist ✓ (basic.png, kerning_fit_spacing__align_center-colors.png, …). Skills ✓. Gates block present (font atlas per size, TRN palette rows per G7, CJK fallback documented).
3 online: n/a.
4 solution: none — ready. ACs testable (integration fixture hash-equal; 20-string wrap-width table).
5 understanding: engine/render layer over n89r's decoded CLX fonts; feeds xih1 UI screens and 7aul's text-golden compose AC.
6 repro: `sed -n '291,295p' Source/engine/render/text_render.cpp; ls test/fixtures/text_render_integration_test/ | head -5; beans query --json '{ bean(id:"devilGo-v1hy") { blockedByIds } }'`
7 fix: none.
8 validation: n/a.
9 related: n89r (dep), 7aul (composes text goldens), h6wn/emdz (UI consumers), glp0 (gmenu text).
VERDICT: TRIAGED-OK
