---
# devilGo-w0ky
title: 'In-game Options: Music / Sound / Gamma sliders (gamemenu.cpp:193-262)'
status: draft
type: task
priority: normal
created_at: 2026-11-25T22:53:27Z
updated_at: 2026-11-25T22:53:36Z
parent: devilGo-xih1
---

# In-game Options: Music / Sound / Gamma sliders (gamemenu.cpp:193-262)

## Goal
The three DISABLED rows the game-menu bean ships become real: Music and Sound volume through sound_get_or_set_*_volume, Gamma through UpdateBrightness(-1) read / UpdateBrightness(v) write+apply (gamemenu.cpp:193-262, :227-233 @e1ad0c2b6).

## Research
- hl4 triage delta (task-hl4-gm-triage, 2026-11-25; .hermes/lane/m2-hero-load/gm-triage-result.md S6): the tone-map port EXISTS and is tested (render.ApplyBrightness / PaletteState.WithBrightness, internal/render/lightvariants.go:51-78); the town compose applies a fixed boot brightness of 60 (present/scene/warrior.go). Missing: persisted setting + live re-apply (no UpdateBrightness Go counterpart), and no audio engine (ProcessSound is a stub, game.go:167).
- Blocked legs: Gamma by the settings/runtime-ini owner (the window boot opens NO engine.Ini today); Music/Sound by the audio bean.

## Scope
- Gamma: live re-apply through the existing brightness seam + persisted setting (runtime-ini).
- Music/Sound: volume knobs through the audio engine once wired.
- The game menu's three rows flip from DISABLED to live.

## Acceptance criteria
- [ ] Gamma slider live-adjusts scene brightness (before/after compose stats differ; setting persists across boot)
- [ ] Music/Sound sliders adjust volumes through the audio engine
- [ ] Full gates incl. -tags windowed + visual PoW; TDD RED evidence per law

## Libraries
stdlib + ebiten audio (existing).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/ebiten/audio/SKILL.md — the audio legs
- /home/blaze/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split from the game-menu bean by the hl4 triage delta (task-hl4-gm-triage, 2026-11-25, parent-verified). Blocked legs: Gamma by the runtime-ini owner, Music/Sound by the audio bean.
