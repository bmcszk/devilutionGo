---
# devilGo-sf6k
title: Gamepad motion + StickPush + TurnToDir
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-12T12:34:39Z
parent: devilGo-k736
---

# Fork input core

## Goal
Exact port of fork stick semantics.

## Research
commit ca91d778f diff (plrctrls.cpp:557,1776-1782,1861-1882,1903-1920; controller_motion.h); FaceDir table; Direction left/right rotation direction.hpp:26,33.

## Scope
internal/engine/input/{motion,stickpush,turn}.go exposing Movement(player) honoring isWalking/CanChangeAction guards.

## Acceptance criteria
- [x] Synthetic axis table: (x,y,deflection) → {None,Weak,Firm} classifications match C++ thresholds
- [x] TurnToDir no-ops while walking (test)


## Libraries
- github.com/quasilyte/ebitengine-input — action maps/stick handling

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
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
1 root: Input epic (k736) leaf task — fork stick semantics port; no blockers (matches tree devilGo-sf6k unblocked).
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

## Progress log

2026-11-12 parent claim (IN-PROGRESS): stick-decision-core lane on feat/stick-motion (master fd3c2ce). Scope re-point per LAYOUT MAP: upstream Source/controls/ mirrors to internal/diablo/controls/ (Player lives in internal/diablo/entities; engine layer cannot import diablo — the bean's internal/engine/input path predates the layout map). Pure decision core: StickPush classification + 8-way sectoring + FaceDir + Movement/TurnToDir/WalkInDir/IsPathBlocked/IsStickMovementSignificant against a narrow PlayerView + func hooks (house pattern: internal/engine/path.go stepContext canStep/posOk). The quasilyte/ebitengine-input library listed under Libraries belongs to the k736 wiring lane, NOT this decision core (stdlib only this lane; wiring maps real axes onto it later). Lane: .hermes/lane/dg-input/stick.s1.md.

## Proof of Work

Wave S1 (parent verification 2026-11-12, feat/stick-motion @ dc71a55,
ls-remote-verified): `internal/diablo/controls/` decision core — stick.go
(ClassifyStickPush, StickDirection 8-way, FaceDir[3][3], Left/Right
rotations) + movement.go (Movement, TurnToDir, WalkInDir, IsPathBlocked,
IsStickMovementSignificant) over a narrow PlayerView + StickHooks
(NetSendCmdLoc boundary = SendWalk/SendStop hooks; house path.go pattern;
no ebiten, no quasilyte, no internal/game import). RED-first: build-fail
run captured (log:57) before impl; green 99 subtests parent-rerun (worker
log 89; T1 9 push rows w/ exact 0.36 boundary, T2 22 sectoring rows incl.
all 4 cutoff boundaries, T3 9 FaceDir cells, T4 7 TurnToDir rows, T5 11
WalkInDir rows, T6 3+4 dispatch/guard rows, T7 7 significance rows).
Gates: build OK, gofmt clean, vet rc0, lint.sh 0 (sim-purity +
executable-surface), beans-lint 185, full suite 47 pkgs 0 FAIL.
Deviations (recorded per fidelity law): (1) CanChangeAction isWalking &&
isLastFrame branch (player.h:771-772) not ported — unreachable from ported
call sites; (2) sectoring boundary pinned `ratio <= cutoff → cardinal` per
controller_motion.cpp:232 (dispatch phrasing said the inverse; code wins).
Deferred to k736 wiring lane: real axis reads, ScaleJoystickAxes/deadzone,
AxisDirectionRepeater, game-UI handler plumbing; quasilyte
ebitengine-input belongs to that lane, not this one.
