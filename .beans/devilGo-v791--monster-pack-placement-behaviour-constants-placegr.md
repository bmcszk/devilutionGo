---
# devilGo-v791
title: Monster pack placement + behaviour constants (PlaceGroup parity)
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-11-28T14:29:23Z
parent: devilGo-97mn
blocked_by:
    - devilGo-778h
    - devilGo-vti7
---

# Monster pack placement + behaviour constants (PlaceGroup parity)

## Goal
The DATA half (goal2 wave re-scope): port PlaceGroup as a PURE, RNG-injected, state-injected function with the vanilla `yp += deltaX` bug reproduced verbatim (parity honesty - no silent fix), plus IsMonsterAvailable spawn gating over monstdat, tested with a scripted RNG tape oracle (house pattern: levels.Rnd, missiles.ProcessCallbacks). Constants land as data (scatter tables, SkeletonKing level 30 if in scope).

## Research
Upstream anchors pinned e1ad0c2b6 (monster.cpp PlaceGroup/PlaceMonster, monotonicity.c). The combined triage (.hermes/lane/goal2/triage.md) verified: monsters data package (778h) completed; world.CreatePack is the HERO starting pack, not a monster pack - zero overlap with this bean's pure function.

## Acceptance criteria
- [ ] PlaceGroup ported as a pure function (RNG + state injected); the vanilla yp += deltaX bug reproduced VERBATIM with a comment citing the upstream line - a silent fix is a parity defect.
- [ ] Scripted RNG tape oracle: placement outcomes pinned to committed expectations (not vs C++ - no probe exists; zero-dependency rule).
- [ ] IsMonsterAvailable gating over the monstdat table (spawn gating as a pure function over the 138-row table).
- [ ] Audit residual line updated: the PlaceGroup/AI-constants registry counts now owned here land (i2cr hand-off closes).

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/port-parity-testing/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Goal2 combined triage 2026-11-27 (task-g2-triage-brainer): PROMOTE-PARTIAL - data/constants half this wave; sim wiring (Monster entity, InitMonster, dMonster occupancy, PlaceMonster verb + kernel handler, unique placement) deferred to a new draft under the World & entities epic (filed by the parent).
