---
# devilGo-7bij
title: 'scene: ScreenshotCaptureWriter panics (ReadPixels before game start) under windowed tests with assets'
status: draft
type: bug
priority: normal
created_at: 2026-11-25T13:04:19Z
updated_at: 2026-11-25T13:04:58Z
parent: devilGo-p1t8
---

# TestScreenshotCaptureWriter panics under `-tags windowed` with assets: ReadPixels before game start

## Goal
Stop `TestScreenshotCaptureWriter` (internal/present/scene/screenshot_test.go:40) from panicking
`ui: ReadPixels cannot be called before the game starts` when the windowed-tagged suite runs with
DG_ASSET_ARCHIVES present. The panic kills the whole scene package run, masking any later test
results in that package.

## Research
- Found 2026-11-25 during the hl1 wave's parent gate re-run (not wave-introduced).
- Classification evidence: the identical panic reproduces on master 770df80 in a throwaway
  worktree (`go test -count=1 -tags windowed ./internal/present/scene` with assets set) — FAIL
  at master too; the wave branch behaves identically.
- Root cause shape: the test calls image.ReadPixels outside a running ebiten game loop; the
  ebiten UI forbids it before the game starts (hajimehoshi/ebiten v2.10.6 internal/ui/ui.go:166).
  The capture path needs the game-loop presenter or a headless-safe capture seam.
- CI is unaffected today: CI runs the untagged suite (windowed files never compile there), so the
  panic is local-gate-only — but the fuller local gate is the pre-push hook's contract and the
  parity oracle's home, so it must not stay broken.
- Assets note: without DG_ASSET_ARCHIVES the run SKIPS (env-gated), which is why plain
  untagged/asset-less runs stay green.

## Acceptance criteria
- [ ] `DG_ASSET_ARCHIVES=<real archives> go test -count=1 -tags windowed ./internal/present/scene` passes (no panic), or the capture test is re-scoped to a headless-safe seam with the game-loop constraint documented
- [ ] The full windowed-tagged suite (cmd/devilutionGo + internal/present/scene) passes with assets set on master
- [ ] Fix lands via the proper triaged flow (TDD: the panicking run captured as RED evidence first)

## Libraries
stdlib + ebiten v2.10.6.

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/ebiten/rendering/SKILL.md — the ReadPixels/game-loop constraint lives here

## Triage
Filed 2026-11-25 by the parent session from live reproduction at master 770df80 (evidence above).
Deep triage (brainer pass) owed before promotion to todo: decide fix shape (game-loop-driven
capture vs headless seam) against the ADR-008 layer map and the pi-notes' documented
ReadPixels/game-loop constraint.
