---
# devilGo-t8dw
title: 'M1 slice: save/load verbs — save/load through the kernel'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-19T22:33:06Z
parent: devilGo-6xu2
blocked_by:
    - devilGo-6851
---

# M1 slice: save/load verbs — save/load flow through the kernel

## Goal
Promote the harness walk's scripted save/load to kernel verbs: commands
(kind(s) SaveSlot/LoadSlot with integer payloads: slot, seed) wired to
the persist SQLite store through a consumer-defined world seam, and
past-tense `SavedSlot`/`LoadedSlot` events through the tick log.
The walk's test-side saved/loaded flags then derive from kernel events.

## Research
- Scripted glue verified: internal/integration save()/load() call the
  persist surfaces directly and stamp `saved=true` (:391) /
  `loaded=true` (session_stage_test.go:526); GameHeader flag bytes come
  from flagByte(s.saved)/flagByte(s.loaded) (modeWord :236-237).
- Persist has the SQLite store + SaveHero/SaveGame/SaveLevel; the seam
  question is the world-to-store interface (consumer-defined, ADR-008).
- TDD red-first: event tests on the unchanged tree; RED recorded.

## Scope
- internal/game/kernel, internal/game/world, internal/integration.
- Determinism: save/load must not perturb the statehash trail except at
  the disclosed border (event log growth that pins count by design).
- Kernel silent-handler law — a SaveSlot that saves nothing (no
  session) stays silent, no event.

## Acceptance criteria
- [ ] Save/Load verb kinds + SavedSlot/LoadedSlot events exist with
      integer payloads; the walk save/load stages flow through the
      kernel (test-side flag stamping retired; flags derive from events)
- [ ] Zero-events pins extended (no-submit save stays silent, named
      test); event-count pins count the new kinds
- [ ] Full gates green at HEAD through pre-commit + pre-push; statehash
      trail deterministic across three processes + plain clone
- [ ] Roundtrip honest: LoadedSlot after SaveSlot restores the walk
      state (a restore-equivalence test names what loaded state equals)

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: the create-pack
slice (the created warrior is the state the save slot freezes).

## Libraries
stdlib + the repo persist/sqlite stack as-is; no new dependencies.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md — Go hard rules, interfaces at consumers
- ~/.agents/skills/go-tests/SKILL.md — test design (black box, no mocks)
- ~/.agents/skills/tdd/SKILL.md — red-first contract
- ~/.agents/skills/coding-rules/SKILL.md — style + gate honesty
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — lane, hook, scope laws
