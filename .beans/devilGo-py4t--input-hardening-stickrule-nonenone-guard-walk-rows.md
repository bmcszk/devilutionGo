---
# devilGo-py4t
title: 'Input hardening: stickRule NONE/NONE guard (walk-rows review F1)'
status: completed
type: task
priority: normal
created_at: 2026-11-20T13:40:39Z
updated_at: 2026-11-25T13:41:19Z
parent: devilGo-k736
blocked_by:
    - devilGo-ny3c
---

---
# devilGo-py4t
title: 'Input hardening: stickRule NONE/NONE guard (walk-rows review F1)'
status: completed
type: task
priority: normal
created_at: 2026-11-20T13:40:39Z
updated_at: 2026-11-25T13:40:48Z
parent: devilGo-k736
blocked_by:
    - devilGo-ny3c
---

---
# devilGo-py4t
title: 'Input hardening: stickRule NONE/NONE guard (walk-rows review F1)'
status: todo
type: task
priority: normal
created_at: 2026-11-20T13:40:39Z
updated_at: 2026-11-25T12:12:30Z
parent: devilGo-k736
blocked_by:
    - devilGo-ny3c
---

# Input hardening: stickRule NONE/NONE guard (walk-rows review F1)

## Goal
A hand-edited StickPush word A=Firm,B=0,C=0 currently walks South in
the world handler, while upstream WalkInDir intercepts the empty
direction and stops (Source/controls/plrctrls.cpp WalkInDir :1747 @ca91d778f; :1720-1726 @e1ad0c2b6). Unreachable
from the decision core today (nonzero magnitude always yields a nonzero
axis), but the kernel should not depend on that: extend stickRule (or
the handler's facing step) to drop/stop on the empty direction, or
restate the faceDir cell with an explicit deviation comment — decide
during triage against the controls core's contract.

## Research
- Review finding F1 on the bridge walk rows wave (commit 7a87cdb,
  review .hermes/lane/ftbu-m1/harness.s6b.review.md).

## Acceptance criteria
- [x] The empty-direction push either stays silent (rule drop) or is
      pinned as a named, commented deviation — per the triage decision
- [x] The walk-rows bean's deviation note updated to point here

## Triage
Filed 2026-11-20 from the walk-rows review F1 (minor, non-blocking);
needs its own triage against the controls core before dispatch.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff

## Triage hl1 (2026-11-25, brainer task-hl1-triage, parent-verified against 949c728)

TRIAGED-OK; decision made: DROP via the gate, not a commented deviation. Reasons: upstream's observable NONE/NONE behavior is "no walk" (WalkInDir returns before reading FaceDir), so the Go South step is a behavior deviation; a stop-walk touches the walkpath this slice does not carry, so silence is the faithful reduced behavior; two-line change in stickFacing. Guard shape: in stickFacing after the range checks, `if c.B == 0 && c.C == 0 { return 0, false }`; faceDir table bytes stay VERBATIM (upstream table; TurnToDir could read row/col 0 later); reword the table NOTE; the gate's RuleNoStep already names the case. Cite fix: Source/controls/plrctrls.cpp WalkInDir :1747 @ca91d778f (:1720-1726 @e1ad0c2b6). Independent of the hero lane; its own commit on the wave branch. Full triage: .hermes/lane/m2-hero-load/result.md.


## Summary of Changes

stickFacing drops a Firm push with B=0,C=0 (upstream WalkInDir returns before reading FaceDir — Source/controls/plrctrls.cpp :1747 @ca91d778f); faceDir table bytes verbatim; NOTE reworded; the gate (RuleNoStep) and handler both stay silent.

## Proof of Work

- RED (gate): `TestGate_DropsANoneNoneStickPush` — judged "", want "no-step" (pre-fix, faceDir[0][0]=0 South returned with ok=true). RED (handler): `TestRegisterMovement_NoneNoneFirmPushChangesNothing` — hero moved to {41 41} + 1 event, want spawn + 0 events. Both pasted verbatim in the round log before the fix.
- Commits: f29279e (guard + tests). Review round 1: MET; the removed pre-existing "NONE/NONE walks south" assertion row is ratified in the ledger note above (superseded assertion replaced by its successor).
- Parent gates at 42cc541: untagged full suite ok; internal/game/world green.
- Residual: none beyond the ratified assertion-row replacement (recorded in this body).
