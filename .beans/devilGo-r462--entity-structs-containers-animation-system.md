---
# devilGo-r462
title: Entity structs + containers + animation system
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - entity
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:06Z
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


## Libraries
- github.com/yohamta/donburi — ECS containers for entity structs

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — struct layout discipline
/home/blaze/.agents/skills/coding-rules/SKILL.md — no logic in structs, methods separate

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — ECS component design
- /home/blaze/.agents/skills/game-development/2d-games/SKILL.md — sprite/animation principles

## Dependencies & blockers
dg-task-015.

## Open questions
- unsafe.Sizeof parity table maintenance cost vs plain field-count assert — take field-count assert.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 task (dg-task-016): Player/Monster/Missile/Object/Item/Towner structs + active lists + AnimationInfo; draft; wire blockedBy [vti7] ✓ = body's dg-task-015.
2 local: struct spot-checks: Object objects.h:32 ✓, Item items.h:189 ✓, Towner towners.h:48 ✓, Missile/MissilePosition missiles.h:89/:25 ✓ (span "25-200" holds), Monster monster.h:213 (body "181-310" — loose span containing it), Player player.h:211 (body "190-389" — span starts at PlayerAnimationData:190, loose but covers). animationinfo_test.cpp ✓ exists; 128-fraction math verified (baseValueFraction in animationinfo.cpp:74-82) ✓; missiles_test.cpp ✓ exists for later cross-checks. Open question self-answers (field-count assert over Sizeof table) — decision already recorded in body.
3 online: n/a.
4 solution: none — ready. Optional: tighten the two loose struct spans.
5 understanding: entity layer under epic-04; exact field order is the contract for epic-08 saves and epic-09 packets — correctly scoped as struct+container only, no logic.
6 repro: `grep -n "^struct Player " devilutionX/Source/player.h; grep -n "^struct Monster\|^struct Missile\|^struct Object\|^struct Item\|^struct Towner" devilutionX/Source/{monster,missiles,objects,items,towners}.h; ls devilutionX/test/animationinfo_test.cpp`
7 fix: none.
8 validation: n/a.
9 related: vti7 (dep), 7be9/rfrj/xw8k/onlm (epic-05 dependents), foug (population).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
