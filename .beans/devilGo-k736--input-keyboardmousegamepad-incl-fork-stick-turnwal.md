---
# devilGo-k736
title: 'Input: keyboard/mouse/gamepad incl. fork stick turn/walk'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-i56j
blocked_by:
    - devilGo-qna9
---

# Input layer

## Goal
Kb/mouse first-class, gamepad with virtual cursor, keymapper/padmapper, and the fork's headline stick behavior.

## Research
controls/controller_motion.cpp:35 ScaleJoystickAxes (radial deadzone), :129-146, :192-205; plrctrls.cpp fork HEAD ca91d778f: Movement() Firm→WalkInDir / Weak→TurnToDir (:1861-1882), TurnToDir (:1776-1782), IsStickMovementSignificant (:1916-1920), Interact fake-attack via _pdir (:557-563); controller_motion.h GetLeftStickDirection unification; keymapper.hpp:17-18; per-frame hooks plrctrls_every_frame:2305 / after_game_logic:2312 (diablo.cpp:1568); ControlTypes scrollrt.cpp:228.

## Scope
internal/input: devices, deadzone scaler, StickPush classifier, plrctrls action machine, keymapper/padmapper with JSON config, auto control-mode switch.

## Acceptance criteria
- [ ] ScaleJoystickAxes port passes curve tests (deadzone, magnitude remap)
- [ ] StickPush thresholds: Weak turns in place (StartStand with new _pdir), Firm walks — behavior tests with synthetic axes
- [ ] Interact fake-attack uses current facing when idle (fork semantics)
- [ ] Keymapper defaults match C++ ini defaults; rebinding persists
- [ ] Mouse+kb and gamepad auto-switch active device (ControlMode)

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — axis/button/touch APIs
/home/blaze/.agents/skills/go-dev/SKILL.md — config persistence

## Dependencies & blockers
blocked_by dg-epic-01 only. Integrates with gameplay in epic-05 hooks.

## Open questions
- Touch virtual gamepad deferred to mobile milestone (H12).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
