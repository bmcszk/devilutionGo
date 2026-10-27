---
# devilGo-zdj7
title: Automap renderer
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-21T11:07:45Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-ru6l
---

# Automap

## Goal
Automap tiles, player arrow, text labels, modes.

## Research
Source/automap.cpp:980 DrawAutomapTile, :1341 Plr, :1427 Text, :1752-1874 DrawAutomap; modes automap.h:83,91; automap_render.cpp:168-171; test automap_test.cpp.

## Scope
internal/render/automap.go reusing world SOL data; toggle key; minimap/transparent variants.

## Acceptance criteria
- [ ] automap_test cases pass
- [ ] Player arrow rotates with _pdir

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — line/stamp drawing

## Dependencies & blockers
Seam note (fix2): automap_test port AC overlaps devilGo-qeyu — canonical automap test = qeyu; zdj7 carries extension-only automap checks.
devilGo-ru6l.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:seam-undeclared (automap_test port AC duplicated with qeyu; one-line body note fixes it)


## IMPL NOTES (triage 2026-10-16)
1 root: Original-tree task (devilGo-zdj7): automap layout, player arrow, text labels, modes. Wire: blockedBy [ru6l] — matches body "devilGo-ru6l" ✓.
2 local: Cites verified: automap.cpp:980 DrawAutomapTile ✓, :1341 DrawAutomapPlr ✓, :1427 DrawAutomapText ✓, :1752 DrawAutomap ✓; automap.h AutomapType enum + CurrentAutomapType in :66-91 region ✓; automap_render.cpp:168-171 ✓ (clip/half-blend); test/automap_test.cpp exists — and its gtests are STATE-level (InitAutomapOnce/StartAutomap/AutomapUp/AmLine — read :1-30), no raster asserts. Skills ✓.
3 online: n/a.
4 solution: Thin pre-graphics-wave body; seam with qeyu declared only from qeyu's side. Both beans list "automap_test cases pass" as an AC — duplicate ownership risk for the same test file. Add one seam note (state/layout cases here; stamps + render cases in qeyu).
5 understanding: engine/render layer; achievable without qeyu (state-level tests + arrow rotation) once seam is written down.
6 repro: `head -30 test/automap_test.cpp; beans show devilGo-zdj7 --body-only | grep -n "automap_test"; beans show devilGo-qeyu --body-only | grep -n "devilGo-zdj7"`
7 fix: `beans update devilGo-zdj7 --body-append "## Seam note: raster line/stamp primitives, blend modes and automap_test *render* cases are owned by devilGo-qeyu (devilGo-qeyu). This bean owns layout, player arrow (_pdir rotation), text labels, modes, toggle; ports the automap_test.cpp state-level cases only."`
8 validation: body carries seam note; qeyu+zdj7 ACs no longer both claim the whole automap_test port.
9 related: qeyu (primitives twin), ru6l (blocker), 7aul (automap-on golden), yd48 (QoL automap items).
VERDICT: NEEDS-WORK:seam-undeclared (automap_test port AC duplicated with qeyu; one-line body note fixes it)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
