---
# devilGo-v1hy
title: Bitmap text renderer (CLX fonts)
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-11T18:07:18Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-1spo
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — glyph SubImage draws

## Dependencies & blockers
devilGo-n89r.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [x] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [x] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [x] Font atlas per size; colour via TRN palette rows (G7); wrapping/spacing flags match `text_render.cpp` behavior; CJK fallback path documented per G7.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Original-tree task (devilGo-v1hy): DrawString equivalents, fonts, wrap, colors, formatters. Wire: blockedBy [n89r] — matches body "devilGo-n89r" ✓.
2 local: Cites verified: text_render.cpp:291 CountNewlines ✓ (read :291-300, DrawStringFormatArg follows); fonts priority 9200 = assets.hpp:309-312 (verified in batch-01); test/text_render_integration_test.cpp + fixtures/*.png exist ✓ (basic.png, kerning_fit_spacing__align_center-colors.png, …). Skills ✓. Gates block present (font atlas per size, TRN palette rows per G7, CJK fallback documented).
3 online: n/a.
4 solution: none — ready. ACs testable (integration fixture hash-equal; 20-string wrap-width table).
5 understanding: engine/render layer over n89r's decoded CLX fonts; feeds xih1 UI screens and 7aul's text-golden compose AC.
6 repro: `sed -n '291,295p' Source/engine/render/text_render.cpp; ls test/fixtures/text_render_integration_test/ | head -5; beans query --json '{ bean(id:"devilGo-v1hy") { blockedByIds } }'`
7 fix: none.
8 validation: n/a.
9 related: n89r (dep), 7aul (composes text goldens), h6wn/emdz (UI consumers), glp0 (gmenu text).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

## Progress log

2026-11-12 parent claim (see below)

IN-PROGRESS — M1-critical next wave after the tiles head merged (#58 → 94855d5). Scope: internal/render/text.go TextRenderer (DrawString equivalent) + wrap parity from upstream text_render.cpp:291-359, fed by assets.Font (landed) + committed 12-05 goldens. The integration-hash AC may be honestly DEFERRED (named, not silently) if the C++ text oracle pull-in proves unbounded; wrap-width parity via committed glyph-width tables is the binding core. Lane: .hermes/lane/dg-text/text.s1.md. PR serialization: opens after the ledger-only #59 merges (different area, but bean custody lands with the lane).

## Proof of Work — slice (parent verification, 2026-11-12, feat/text @ e8ed94d, PR #60)

- Waves: S1 (research-only, under-delivered — its findings distilled forward), S2 (same failure class, banked the decisive spec facts incl. 99bf0e4e3 hard-split semantics), S3 (implementation: text.go 214 L + text_test.go 491 L, RED-first with 10 failing LineWidth parity cases), fix wave (review F1-F6: rect clip RED-first, breakable-before-fit reorder on a synthetic non-blank-space sheet, NoTable option, per-font/outline-dedup cites). 4 commits pushed: 0a02e0f custody, b138341 unit, e8ed94d head.
- Parent gates at e8ed94d: render+assets 582 pass (43 text subtests); full suite 11,461/39 pkgs; lint.sh 0 issues (sim purity clean); gofmt empty; porcelain clean; ls-remote == head.
- Review round: REQUEST-CHANGES → all six dispositioned; F4 (wrap oracle) tracked as devilGo-od1d (draft).
- Deviations recorded: hard-split wrap = upstream CURRENT tree (99bf0e4e3), not the pre-2023 keep-whole rule; Options.TRN zero value = no remap, NoTable field added for the all-zero-table case (reviewer F2 disposition); gameFontLineHeight stays single-sheet with the per-font TODO (F5).
- Still open (stays in-progress): integration-hash AC (C++ text oracle), wrap-oracle goldens (od1d), atlas-per-size when the second sheet lands.
