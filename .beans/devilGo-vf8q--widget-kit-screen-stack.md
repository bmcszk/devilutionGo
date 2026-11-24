---
# devilGo-vf8q
title: Widget kit + screen stack
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-20T16:36:54Z
parent: devilGo-xih1
blocked_by:
    - devilGo-v1hy
---

---
# devilGo-vf8q
title: Widget kit + screen stack
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-21T11:07:45Z
parent: devilGo-xih1
blocked_by:
    - devilGo-v1hy
---

# Widgets

## Goal
Six primitives + stack + focus model.

## Research
DiabloUI/button.cpp, scrollbar.cpp, text_input.hpp, ui_item.h; gmenu item/sliders.

## Scope
internal/ui/widget.go, screen.go; art loading via assets; input routing.

## Acceptance criteria
- [x] Keyboard nav (arrows+enter+esc) on every primitive
- [ ] Slider emits int steps like gmenu_slider_get semantics


## Libraries
- github.com/ebitenui/ebitenui — widget kit (optional; eval vs hand-rolled at claim, record decision)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/input/SKILL.md — key/edge detection

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — UI state machines
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md — menu/UX flow

## Dependencies & blockers
devilGo-v1hy (text).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic foundation task — six widget primitives + screen stack + focus; blocked_by v1hy (bitmap text renderer, exists in qbiy) matching tree devilGo-vf8q←devilGo-v1hy.
2 local: DiabloUI/button.cpp, scrollbar.cpp, text_input.hpp, ui_item.h (UiItemBase:30) all exist; gmenu_slider_get gmenu.h:101 + gmenu_slider_steps :106 verified.
3 online: n/a.
4 solution: none — AC unit-testable (keyboard nav per primitive; slider int-step semantics vs gmenu_slider_get). Nit: prose "devilGo-v1hy (text)" stale → v1hy (structured already correct); Goal says "six primitives" — enumerate them in AC to make "every primitive" checkable.
5 understanding: UI layer core; correct dependency (text renderer before widget kit).
6 repro: `beans show devilGo-vf8q`; `grep -n "gmenu_slider_get" .../Source/gmenu.h`; `ls .../Source/DiabloUI/{button,scrollbar}.cpp`.
7 fix: none required (optional prose id remap).
8 validation: n/a.
9 related: v1hy, emdz, dcob, glp0, xih1.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.



## Progress log

- 2026-11-20T15:50Z: claimed — s6d dispatched under contract harness.s6d.md (worktree ftbu-s6, feat/m1-interactive-demo; base 140bd2b), status flip.


## Proof of Work (slice 1 — buttons/lists/focus/keyboard nav)

- Worker: task-dg-m1-s6d-coder-coder (147 tool calls), commit 80ae264 on feat/m1-interactive-demo — pushed; parent verified ls-remote == HEAD; ui suite 64 tests ok, full suite 54 pkgs ok, single-walk trail f1ae95a0… unchanged (presentation-only).
- Delivered: Button (exclusive focus, Enter/Space activation), MenuList (wrap both ends, Enter/Escape), WidgetScreen + ScreenStack, focus model, keyboard nav — 377 lines + 509 black-box test lines (package ui_test, no mocks).
- Review: brainer FAIL→fix→PASS (harness.s6d.review.md + harness.s6d.re-review.md) — F1 (keys-inert-without-focus unpinned) resolved by the prescribed named test (fix commit 59287ba, test-only +44, honest unchanged-tree falsification recorded); F2 (Button self-focus check) + F4 (KeyUnknown nit) named for later slices; F3 record corrections landed.
- Scope statement: THIS slice covers the four first-slice behaviors only; the remaining ACs (slider semantics, keyboard nav on every primitive, parity/deviation gates) stay OPEN for the later widget waves — the bean remains in-progress by design (beast-size slicing).
- Residual: slider row, per-primitive nav sweep, F2/F4 nits.
