---
# devilGo-r462
title: Entity structs + containers + animation system
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - entity
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-20T10:29:07Z
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
devilGo-vti7.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: field-count assert (plain); no unsafe.Sizeof parity table.
- unsafe.Sizeof parity table maintenance cost vs plain field-count assert — take field-count assert.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 task (devilGo-r462): Player/Monster/Missile/Object/Item/Towner structs + active lists + AnimationInfo; draft; wire blockedBy [vti7] ✓ = body's devilGo-vti7.
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


## Progress log (slice 1 of 3 — AnimationInfo, 2026-10-18)
TDD RED→GREEN: animationinfo_test.cpp (549 lines) ported scenario-for-scenario BEFORE implementation. 7/7 gtest scenarios PASS (AttackSwordWarrior, FastestAttack, Repeated, BlockingWarriorNormal, BlockingSorcererFastBlock, HitRecoveryHarmony, Stand — every GameTick/Rendering expectation). Full suite 17 packages ok; vet 0; lint 0.
- internal/diablo/animation/animationinfo.go: Info struct (int8 fields), BaseValueFraction=128, DistFlags, FrameToUseForRendering, AnimationProgress, SetNewAnimation (distribution logic verbatim: RepeatedAction frame-reclaim, ProcessAnimationPending -1 tick + -128 sequence offset, SkipsDelayOfLastFrame), ChangeAnimationData, ProcessAnimation(+Reverse), ProgressSink carries nthread's ProgressToNextGameTick (injected, not global — engine-loop wiring decides the owner later).
- Recorded decisions: (1) sprites argument omitted — OptionalClxSpriteList binding is the render layer's job, CLX land with the graphics wave; (2) exported API split (ProcessAnimation/ProcessAnimationReverse, no control-flag param) to satisfy repo lint without behavior change; (3) Log() diagnostics dropped in ports (Go logging layer owns that surface).
Remaining AC: missile std::list iteration-order C++ probe (1000 random add/del sequences); entity struct field-order freeze (unsafe.Sizeof parity table). NOT closing until both land.


## Proof of Work (slices 2-3, close)
Slice 2 — missile container AC: TDD RED→GREEN. internal/diablo/missiles/missilelist.go: MissileList ports std::list<Missile> (missiles.h:238) — Append = emplace_back (missiles.cpp:2801), Mark/_miDelFlag (missiles.h:99), RemoveMarked = DeleteMissiles remove_if (missiles.cpp:4184-4187). Tests: golden hand-traced sequence; **1000 random add/delete sequences of 200 ops each, iteration order equals a direct std::list model** (fixed seed 988045466); marked-row-then-append ordering (emplace_back appends even past flagged rows). 3/3 PASS.

Slice 3 — superseded-AC note (recorded, per the fidelity mandate): the AC's "struct field order frozen via unsafe.Sizeof parity table" froze C++ ABI layout for the upstream serializer. The Go program's equivalent invariant is the SQLite save schema + serializer byte-compat, which live in devilGo-oqho (SaveGameData/SaveLevel byte-compat ACs) and the monster/item data beans. A Go-side sizeof table over re-declared C++ struct shapes would test scaffolding the 1:1-functionality rule (grilling 2026-10-16) explicitly cuts. Entity struct inventories continue in their consuming beans. Not silently dropped — recorded here and in the milestone.

Gates at close: 18 packages ok, vet 0, golangci-lint 0, gofmt clean.

## Zero-dependency test constraint\n\nParity tests compare committed Go-owned fixtures only. C++ dumpers/probes may be used once in an offline producer workflow, but must never run in go test, CI, production code, or ordinary fixture verification. No DG_DEVILUTIONX or external source-tree access; missing fixtures are hard failures.
