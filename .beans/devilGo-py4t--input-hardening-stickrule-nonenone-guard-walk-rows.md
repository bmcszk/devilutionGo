---
# devilGo-py4t
title: 'Input hardening: stickRule NONE/NONE guard (walk-rows review F1)'
status: draft
type: task
priority: normal
created_at: 2026-11-20T13:40:39Z
updated_at: 2026-11-20T13:40:39Z
parent: devilGo-k736
blocked_by:
    - devilGo-ny3c
---

# Input hardening: stickRule NONE/NONE guard (walk-rows review F1)

## Goal
A hand-edited StickPush word A=Firm,B=0,C=0 currently walks South in
the world handler, while upstream WalkInDir intercepts the empty
direction and stops (plrctrls.cpp:1726-1731 @ e1ad0c2b6). Unreachable
from the decision core today (nonzero magnitude always yields a nonzero
axis), but the kernel should not depend on that: extend stickRule (or
the handler's facing step) to drop/stop on the empty direction, or
restate the faceDir cell with an explicit deviation comment — decide
during triage against the controls core's contract.

## Research
- Review finding F1 on the bridge walk rows wave (commit 7a87cdb,
  review .hermes/lane/ftbu-m1/harness.s6b.review.md).

## Acceptance criteria
- [ ] The empty-direction push either stays silent (rule drop) or is
      pinned as a named, commented deviation — per the triage decision
- [ ] The walk-rows bean's deviation note updated to point here

## Triage
Filed 2026-11-20 from the walk-rows review F1 (minor, non-blocking);
needs its own triage against the controls core before dispatch.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
