---
# devilGo-njzu
title: 'M1 slice: live view — presenter follows the hero'
status: draft
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-20T12:19:16Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-9qzw
---

# M1 slice: live view — presenter follows the hero

## Goal
The presenter composes ONE static first scene and never re-composes
(cmd/devilutionGo/window_scene.go:73-103). Make the view live the legal
way: the presentation layer reads kernel events (PlayerMoved,
CreatedPack) as its ONLY state source (ADR-018 law 2 — it never writes
sim), recenters the town view on the hero (camera as presentation
state, ClickRouter TileAt precedent), draws the hero sprite from the
world's hero state, and caches town records. A screenshot artifact
shows a moved hero.

## Research
- Plan: wave-s6-plan.md; the first-scene pipeline (retail Tristram +
  warrior STAND at heroTile 75,68) is byte-validated against the
  user-approved #73 baseline — the recenter must keep that fidelity
  (pixel-stats evidence per the visual law).

## Scope
- internal/present/** + cmd/devilutionGo view glue; NO sim changes.

## Acceptance criteria
- [ ] View follows the hero during a scripted (test-driven) walk: a
      later frame's screenshot differs from the first by the expected
      recenter (pixel stats recorded); hero sprite follows state
- [ ] The static first frame is UNCHANGED (the validated baseline
      intact before any motion); full gates green; RED first

## Triage
Planner wave (wave-s6-plan.md s6h), filed 2026-11-20; blocked by the
device lane (a live walk to follow).

## Libraries
stdlib + ebiten (existing dep).

## Skills for coder
- ~/.agents/skills/ebiten/rendering/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
