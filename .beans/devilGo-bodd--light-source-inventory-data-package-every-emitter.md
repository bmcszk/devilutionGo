---
# devilGo-bodd
title: Light source inventory data package — every emitter radius/lifetime/flags (L3)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:04Z
updated_at: 2026-10-15T22:42:22Z
parent: devilGo-97mn
blocked_by:
    - devilGo-n89r
    - devilGo-b782
    - devilGo-m18k
---

# Light source inventory data package

## Goal
Create the complete data-driven light emitter catalog (L3 table) as a Go package: every C++ AddLight call site with its radius, lifetime rules, network sync needs, and source file:line. This feeds runtime light creation and the net command payloads.

## Research
Grep-verified callers:
- Player torch: player.cpp:2524 AddLight(pos, _pLightRad=10) + ChangeLightXY each walk; items.cpp:2542 ChangeLightRadius on equip
- Monsters default: monster.cpp:3351 radius 3; diablo.cpp:1506 radius from monstdat (monster light field); dead.cpp:40 corpse
- Diablo: monster.cpp:898 radius 8
- Objects: objects.cpp:1483 objdat lightRadius; trap objects.cpp:1582/1596/1638 anim-driven ChangeLightRadius
- Missiles: portal.cpp:59 radius 15; missiles.cpp:691/1613/1631/1876 radius 8; :1726/1767 radius 5
- Vision (separate): ActivateVision(player.pos, _pLightRad, id) player.cpp:2527
- MAXLIGHTS=32 cap, NO_LIGHT=-1 fallback

## Scope
Package: `internal/light/sources.go`
- struct LightSourceDef { SourceType enum; Radius uint8; Flags (Persistent, Animated, Networked); CodeSite string }
- Registry: map[SourceType]LightSourceDef with all entries above
- Function `GetLightSourceDef(t SourceType) LightSourceDef` for runtime creation
- Cross-ref: objdat loader (devilGo-107), monstdat loader (devilGo-104), missile defs (devilGo-106)

## Acceptance criteria
- [ ] Registry contains ≥12 distinct entries covering all grep callers
- [ ] Row-count parity: source count = grep lines (13 exact) + vision (1) = 14
- [ ] Each entry cites exact file:line from C++ source
- [ ] Integration test: spawn each source type via AddLight → dLight matches radius within 2 tiles

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers
- devilGo-gk5e (dungeon bounds for in-bounds checks)
- devilGo-n89r (decoders)

## Open questions
- Town torches: town.cpp AddLight calls not fully grep'd — add if found
