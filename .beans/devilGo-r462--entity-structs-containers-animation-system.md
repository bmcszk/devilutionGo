---
# devilGo-r462
title: Entity structs + containers + animation system
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - entity
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-97mn
blocked_by:
    - devilGo-vti7
---

# Entities

## Goal
Full Player/Monster/Missile/Object/Item/Towner structs (all fields, even unused) + active-list containers + AnimationInfo.

## Research
player.h:190-389; monster.h:181-310; missiles.h:25-200; objects.h:32-80; items.h:189-230; towners.h:48-63; animationinfo.h/.cpp; ActorPosition engine/actor_position.hpp; direction.hpp:11,26,33.

## Scope
internal/entity/*: structs with exact field order (save/packet serialization depends), ActiveMonsters/Items/Objects swap lists, Missiles pool with free-list emulating std::list splice order, Direction helpers, AnimationInfo with 128-fraction math.

## Acceptance criteria
- [ ] animationinfo_test.cpp port passes
- [ ] Missile iteration order equals std::list simulation for 1000 random add/del sequences (cross-checked against a small C++ probe)
- [ ] Struct field order frozen: reflection-free serializers compile-time match via generated size assertions (unsafe.Sizeof parity table)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — struct layout discipline
/home/blaze/.agents/skills/coding-rules/SKILL.md — no logic in structs, methods separate

## Dependencies & blockers
dg-task-015.

## Open questions
- unsafe.Sizeof parity table maintenance cost vs plain field-count assert — take field-count assert.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
