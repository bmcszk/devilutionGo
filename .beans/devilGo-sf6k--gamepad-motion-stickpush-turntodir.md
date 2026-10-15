---
# devilGo-sf6k
title: Gamepad motion + StickPush + TurnToDir
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-k736
---

# Fork input core

## Goal
Exact port of fork stick semantics.

## Research
commit ca91d778f diff (plrctrls.cpp:557,1776-1782,1861-1882,1903-1920; controller_motion.h); FaceDir table; Direction left/right rotation direction.hpp:26,33.

## Scope
internal/input/{motion,stickpush,turn}.go exposing Movement(player) honoring isWalking/CanChangeAction guards.

## Acceptance criteria
- [ ] Synthetic axis table: (x,y,deflection) → {None,Weak,Firm} classifications match C++ thresholds
- [ ] TurnToDir no-ops while walking (test)

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — StandardGamepadAxisValue

## Dependencies & blockers
None (needs direction.go from engine).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
