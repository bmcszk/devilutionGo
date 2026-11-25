---
# devilGo-uuyh
title: 'M1 slice: front screens minimal — title to main menu to single-player entry'
status: completed
type: task
priority: normal
created_at: 2026-11-23T12:46:49Z
updated_at: 2026-11-23T14:53:04Z
parent: devilGo-xih1
blocked_by:
    - devilGo-njzu
---

# M1 slice: front screens minimal — title → main menu → single-player entry

## Goal
The windowed boot goes straight to the town today. Insert the minimal
front-screen flow: title screen (press any key) → main menu (Single
Player / Exit, driven by the widget kit's MenuList) → Single Player
hands to a selhero PLACEHOLDER screen (the real list/create/delete UI
is the next slice). Presentation-layer only.

## Research
- The widget kit first slice (Button/MenuList/ScreenStack/focus,
  internal/game/ui) — landed in the s6 wave; compose on it, do not
  fork it.
- runWindowed boot path + the live-view presenter seam (the s6b-demo
  wave): the scene stack swaps screens; the town screen stays the
  terminal screen of this slice's stack.
- diablo menu flow is presentation; the sim stays untouched (ADR-017
  layer note).

## Acceptance criteria
- [x] Boot lands on the title screen; any key → main menu; Enter on
      Single Player → the placeholder selhero screen; Escape twice
      walks back; Exit (and Escape on the title) quits with rc 0
- [x] Every transition is a black-box ui/scene test (windowless,
      no mocks); RED captured on the unchanged tree (boot lands on
      town today)
- [x] The widget kit is consumed as-is (no widget forks); screens sit
      behind the window build guard; headless stays scene-free
- [x] Full gates green; trail f1ae95a0… unchanged

## Triage
Filed 2026-11-20 for the demo-flow wave (front screens were in the s6
plan as s6f; the emdz ID from earlier notes does not exist in the
ledger — this is the real filing).

## Libraries
stdlib + the in-repo widget kit.

## Skills for coder
- ~/.agents/skills/ebiten/input/SKILL.md
- ~/.agents/skills/ebiten/rendering/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md



## Progress log

- 2026-11-20T19:40Z: claimed — front-screens slice dispatched under contract harness.frontscreens.md (worktree ftbu-demo, feat/m1-demo-flow; base 4f0362b carries the live view).



## Proof of Work (frontscreens wave, 2026-11-23)

- RED on the unchanged tree (boot landed on the town): TestFrontBootLandsOnTitle
  failed with "boot screen = 0, want FrontTitle"; also tripped the headless
  dep-graph guard once (TestHeadlessMode_HasNoDisplayDependency caught the
  scene import outside the windowed tag) — fixed by moving the boot wiring
  behind //go:build windowed.
- GREEN: front flow in internal/present/scene (title → main menu via the
  widget kit's MenuList as-is → selhero placeholder; Escape walks back; Exit /
  Escape-on-title quit). Black-box windowless tests drive keys and assert the
  active screen + quit paths (frontflow_test.go).
- Screens sit behind the window build guard; the town scene stays the terminal
  render, the flow draws in front of it (frontflow_render.go, palette-index
  band + label bars — the retail 12-05 font sheet carries no ASCII glyphs, so
  string labels wait for the h6wn text lane; recorded as a deferred unit, not
  a silent deviation).
- Gates: gofmt empty; go vet rc0; scripts/lint.sh 0 issues; go test ./... all
  ok (plain AND -tags windowed cmd suite); beans-lint OK; integration trail
  sha256 f1ae95a0… unchanged (M1_LANE_DIR run re-hashed).
- Sim + device lane untouched; scope = cmd/devilutionGo +
  internal/present/scene.



## Proof of Work

- Worker task-dg-m1-frontscreens-coder-coder: commit 97b841c on feat/m1-demo-flow (7 files +500/−6) — title → main menu (widget-kit MenuList, consumed as-is) → selhero placeholder (stub named as such for devilGo-5fr4).
- Review: brainer FAIL (window seam fed KeyUnknown idle ticks as 'any key' — title auto-advanced; single Escape quit instead of back-walking; RED record misdescribed the unchanged tree) → fix task-dg-m1-frontscreens-fix-coder commit 03f45ca (6 files +103/−25: real-key gating at the seam, escape back-walk, honest RED correction + new pins' verbatim RED) → re-review PASS 8/8 (.hermes/lane/ftbu-m1/harness.frontscreens.re-review.md).
- Parent gates at 03f45ca: full suite 54 pkgs ok; trail f1ae95a0… unchanged; ls-remote == HEAD verified.
- The placeholder screen is the seam devilGo-5fr4 (selhero first slice) fills next.
