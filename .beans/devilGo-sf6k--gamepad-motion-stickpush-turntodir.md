---
# devilGo-sf6k
title: Gamepad motion + StickPush + TurnToDir
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:16:44Z
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


## Libraries
- github.com/quasilyte/ebitengine-input — action maps/stick handling

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — StandardGamepadAxisValue

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — input/controller patterns
- /home/blaze/.agents/skills/ebiten/input/SKILL.md

## Dependencies & blockers
None (needs direction.go from engine).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Input epic (k736) leaf task — fork stick semantics port; no blockers (matches tree dg-task-031 unblocked).
2 local: fork commit ca91d778f exists in repo git log ("Stick turn/walk + GetLeftStickDirection unification"); plrctrls.cpp:557 Interact fake-attack region verified; :1776-1782 TurnToDir exact (isWalking/CanChangeAction guards verified — AC "no-ops while walking" is directly testable); :1861-1882 Movement stick logic verified (PadMenu/Hotspell guards + FaceDir); :1903-1920 IsStickMovementSignificant verified (~:1916); FaceDir table plrctrls.cpp:100-105 verified; Left/Right rotation verified at Source/engine/direction.hpp:24-33 (bean cites "direction.hpp:26,33" — file is under engine/, lines correct); controller_motion.h exists with GetAnalogStickPush/GetAnalogStickDirection (:39-40 — the commit's GetLeftStickDirection name was unified/renamed at HEAD; bean anchors to the commit diff so OK, implementer should port HEAD names).
3 online: n/a (commit is local fork history).
4 solution: none — synthetic axis-table AC + TurnToDir guard AC are directly testable; thresholds portable from controller_motion.cpp.
5 understanding: input layer leaf depending only on engine direction type — correctly unblocked; k736 epic itself gated on qna9 scaffold.
6 repro: `beans show devilGo-sf6k`; `git -C ~/github/bmcszk/devilutionX show --stat ca91d778f`; `sed -n '1776,1782p' .../Source/controls/plrctrls.cpp`; `sed -n '100,105p' .../controls/plrctrls.cpp`.
7 fix: none.
8 validation: n/a.
9 related: c6wt, t4fz, k736.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
