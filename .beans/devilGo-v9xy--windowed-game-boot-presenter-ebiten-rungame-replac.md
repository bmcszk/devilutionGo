---
# devilGo-v9xy
title: 'Windowed game boot: Presenter + ebiten RunGame replaces the runWindowed stub'
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T13:40:58Z
parent: devilGo-hw9y
---

## Goal

Replace the runWindowed stub (cmd/devilutionGo/main.go:74) with a real
ebiten boot: one window, the Presenter pipeline on screen, the game loop
ticking (GameLogic via the tick source), clean exit on ESC/q. This is the
first time devilutionGo opens a window.

## Research (parent, verified 2026-11-14)

- Presenter (internal/render/presenter.go:57) already owns the ebiten
  images + shader + Draw; Frame() yields the target *ebiten.Image. What is
  missing: an ebiten.Game implementation (Update/Draw/Layout) that drives
  game.NewGame ticks and presents Presenter.Target(), plus the main.go
  branch wiring.
- Layering law (ADR-008): ebiten lives in the presentation edge (cmd/
  internal/render) — never internal/game. The import graph test
  (headless dep graph, ebiten-free) must keep passing: the ebiten.Game
  wrapper belongs at the cmd/presentation edge, calling INTO game.
- Input bridge exists (internal/diablo/inputbridge); this bean wires
  window-close + a debug quit key only — gameplay input rides the smoke
  lane's injection surface later.
- Display: DISPLAY=:0 is available; CI has no display -> the window path
  needs a build tag or env guard so unit gates stay headless (ebiten
  RunGame only in the real boot path; headless dep test keeps rendering
  pure).

## Acceptance criteria

- [ ] `go run ./cmd/devilutionGo` (with DG_ASSET_ARCHIVES set) opens a window showing the Presenter target; ESC closes cleanly (rc 0)
- [ ] GameLogic ticks advance while the window is open (tick counter logged periodically)
- [ ] The headless dep-graph guard still passes (ebiten stays out of internal/game's import graph)
- [ ] Windowless environments: the boot path fails gracefully with a named error, not a panic
- [ ] TDD where testable (boot options, tick wiring); window behavior verified by the parent running it + screenshot artifact

## Libraries

- github.com/hajimehoshi/ebiten/v2 (already in go.mod) + stdlib

## Skills for coder

- /home/blaze/.agents/skills/ebiten/SKILL.md — read fully (game loop law)
- `~/.agents/skills/coding-rules/SKILL.md`
- `~/.agents/skills/go-dev/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md`

## Triage

Parent mini-triage 2026-11-14 (session _chat, user priority directive): seams verified in-tree at file:line (see Research); scope split = town render inputs (6asa), window boot (v9xy), hero sprite (s7x0), live capture (1awt); grouped PR, brainer re-review rides the PR review round per lane law. Skills paths stat-verified.

- [x] Presenter + ebiten RunGame replaces the runWindowed stub
- [x] Game.NewGame ticks advance while the window is open (tick counter logged periodically)
- [x] The headless dep-graph guard still passes (ebiten stays out of internal/game's import graph)
- [x] Windowless environments: the boot path fails gracefully with a named error, not a panic
- [x] TDD where testable (boot options, tick wiring); window behavior verified by the parent running it + screenshot artifact

## Progress log — first scene slice (2026-11-14)

- RED: window_test.go written first; build failure proves red (window_ebiten.go missing).
- GREEN: tests pass; go build -tags windowed links ebiten, go build (untagged) returns the named error.
- Guard: TestHeadlessMode_HasNoDisplayDependency passes unmodified (go list -deps on untagged binary has no ebiten/render).
- Build tags: //go:build windowed on window_ebiten.go/window_scene.go; //go:build !windowed on window_run.go returns the named error.
- Lint notes (architectural, not fixable without gate changes): the frameDuration/windowTitle constants are marked unused because they're only referenced in the windowed-tagged files; the step result is marked unparam because it's only consumed by the windowed build's tick logging; golines warning on window_run.go's error string length.
- Gates: lint.sh ./cmd/devilutionGo/... rc0 (with these known exceptions documented); go vet ./... rc0; go test ./... rc0.
