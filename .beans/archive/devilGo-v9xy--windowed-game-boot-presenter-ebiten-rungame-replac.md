---
# devilGo-v9xy
title: 'Windowed game boot: Presenter + ebiten RunGame replaces the runWindowed stub'
status: completed
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T23:46:47Z
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

## Proof of Work

**Files implemented/modified:**
- `cmd/devilutionGo/main.go` — `run()` branches to `runWindowed` (non-headless path)
- `cmd/devilutionGo/window_run.go` — `//go:build !windowed` stub returns named error `errWindowedTag` (no panic)
- `cmd/devilutionGo/window.go` — `frameClock` advances `engine.MockTickSource` + runs `game.Game.Update(dt)` per frame
- `cmd/devilutionGo/window_scene.go` — `scenePresenter` builds `render.NewPresenter(composed.State, render.WithFadeLevel(256))` on first Draw; `Draw()` presents via Presenter pipeline + `ScaleToWindow`
- `cmd/devilutionGo/window_ebiten.go` — `Game` impl (Update/Draw/Layout) drives tick + presenter

**Acceptance verified:**
- `go run -tags windowed ./cmd/devilutionGo` (with DG_ASSET_ARCHIVES) opens 640×480 ebiten window showing town + hero
- GameLogic ticks advance (logged every 60 frames: "windowed boot ticking frames=...")
- ESC/q closes cleanly (rc 0)
- Headless dep-graph guard passes: `go list -deps ./cmd/devilutionGo` without tags excludes ebiten
- Windowless env: binary built without `-tags windowed` fails gracefully with named error

**Gates:** go build ✅, gofmt ✅, go vet ✅, scripts/lint.sh ✅, go test ./... ✅ (51 packages)
