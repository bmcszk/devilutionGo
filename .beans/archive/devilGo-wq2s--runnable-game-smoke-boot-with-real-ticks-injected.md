---
# devilGo-wq2s
title: 'Runnable game smoke: boot with real ticks, injected input, logs, screenshots'
status: in-progress
type: epic
priority: critical
created_at: 2026-11-14T11:12:57Z
updated_at: 2026-11-14T11:14:21Z
---

## Goal

Get a REAL running game observable this week: boot the game loop with real
ticks, inject input events, read structured logs, capture a screenshot of a
live frame — no menu flow, no character selection, no ebiten window required
to start. "Something works" = an e2e run boots the game, sends a click, the
hero moves, the log proves it, and a screenshot shows the frame.

## Why now (user directive 2026-11-14, verbatim intent)

"Testing running game as early as possible — input events straight to the
game, checking running game logs, taking screenshots — we need this early
to know that we have something working, even without menu/character
selection."

## Verified seams (parent triage 2026-11-14, file:line)

- internal/game/game.go:66 NewGame(TickSource); :109 GameLogic; :217
  runKernelPass — the loop already runs headless (cmd/devilutionGo/headless.go
  drives it today with MockTickSource).
- internal/game/world/movement.go:8 RegisterMovement wires PlayerMove
  commands into the kernel; state.go:30 StepTo moves the hero.
- internal/game/click.go:18 ClickRouter turns demo MouseButtonRecords into
  walk/attack commands; internal/game/demo.go:65 MouseButton records them.
- internal/render/index_blit.go:53 NewSurface — index framebuffer allocates
  headless (no GPU); paritytest.Capture exists for the GPU path (DISPLAY=:0)
  but is NOT required for the first screenshot: dump the index surface
  through the palette to PNG with stdlib.
- Retail MPQs via DG_ASSET_ARCHIVES are available but the world layer does
  not need them for this epic.

## Task children (all critical, one grouped PR)

1. Headless real-world boot with structured run logs (e2e) — u1hd
2. Scripted input injection: mouse -> kernel commands -> hero moves — judg
3. Screenshot from a live frame + smoke recipe — jlrz

## Acceptance criteria (epic)

- [ ] One command (`just smoke` or `go test`) boots Game+kernel+world, injects a scripted click, and finishes green with structured logs
- [ ] The log stream proves the causal chain: command submitted -> gate passed -> event emitted -> hero tile changed
- [ ] A PNG screenshot of a composed live frame exists as a run artifact
- [ ] All three child tasks completed

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-dev/SKILL.md` — Go rules + gates
- `~/.agents/skills/go-tests/SKILL.md` — e2e test design

## Progress log

Claimed 2026-11-14 (parent planning session): seams verified (game.go:66/109/217, movement.go:8, click.go:35, demo.go:65, index_blit.go:53); children u1hd/judg/jlrz filed critical + todo; grouped-PR law applies (one PR, user standing directive). Epic leads the queue ahead of M1's remaining text ACs per user directive.
