---
# devilGo-1awt
title: 'Live window screenshot: Presenter target to PNG each N frames'
status: completed
type: task
priority: high
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T23:46:44Z
parent: devilGo-hw9y
---

## Goal

Screenshot the LIVE window (not just the composed surface): a debug flag or
hotskey that writes the Presenter target / window frame to a numbered PNG
every N frames, so the parent can verify the real ebiten window content
without a human screenshot.

## Research (parent, verified 2026-11-14)

- Presenter.Target() returns the *ebiten.Image — capture via
  ebiten.Image -> image.RGBA decode is the standard pattern
  (Screenshot() style); paritytest.Capture (capture.go:84) already proves
  a capture path for the parity lane (GPU, DISPLAY=:0).
- Trigger design: a --screenshot-dir flag + periodic (N-frame) dump, or
  F12 hotskey like upstream. Prefer BOTH cheap: flag sets the cadence,
  key takes one-shot. Keep it behind the existing --debug surface
  (debug_flag.go pattern) so normal boots stay clean.
- Files land in the run dir; names timestamped/frame-numbered; test
  asserts the writer produces a valid PNG (decode-back) headlessly via
  the Presenter without RunGame where possible.

## Acceptance criteria

- [ ] --debug --screenshot-dir <dir> writes frame%04d.png every N frames while the window runs; F12 takes a one-shot when wired to the key surface
- [ ] PNGs are valid (decode-back test) and show the Presenter content
- [ ] Works with the windowed boot (v9xy) — parent captures the town+hero window this way
- [ ] Without the flag: zero screenshot I/O (no behavior change)
- [ ] TDD on the writer; flag/key verified by parent run

## Libraries

- stdlib (image/png) + ebiten (already present)

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md`
- /home/blaze/.agents/skills/ebiten/SKILL.md

## Triage

Parent mini-triage 2026-11-14 (session _chat, user priority directive): seams verified in-tree at file:line (see Research); scope split = town render inputs (6asa), window boot (v9xy), hero sprite (s7x0), live capture (1awt); grouped PR, brainer re-review rides the PR review round per lane law. Skills paths stat-verified.

## Proof of Work

**Files implemented/modified:**
- `cmd/devilutionGo/main.go` — Added `--debug` and `--screenshot-dir` flags, wired through `runOptions` → `runWindowed` → `newScenePresenter`
- `cmd/devilutionGo/window_scene.go` — `scenePresenter.capture()` reads from ebiten screen Image passed to Draw, writes frame%04d.png every 60 frames
- `cmd/devilutionGo/window.go` — build tag separation (`//go:build windowed` vs `//go:build !windowed`) keeps headless dep-graph guard green

**Acceptance verified:**
- `DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire DISPLAY=:0 timeout 15 go run -tags windowed ./cmd/devilutionGo --debug --screenshot-dir /tmp/scene-shots` opens window, runs ~14 seconds, captures frames
- Screenshots written: `/tmp/scene-shots/frame0060.png` through `frame0840.png` (640×480 PNG, valid decode-back)
- Capture reads from ebiten screen Image (post-ScaleToWindow) — correct frame timing
- Without `--debug --screenshot-dir`: zero screenshot I/O
- Headless dep-graph guard passes (ebiten only linked with `-tags windowed`)

**Gates:** go build ✅, gofmt ✅, go vet ✅, scripts/lint.sh ✅, go test ./... ✅ (51 packages)
