---
# devilGo-mxda
title: 'ADR-002 note: parity anchor ca91d778f canonical, no re-pin'
status: completed
type: task
priority: high
created_at: 2026-11-13T00:15:38Z
updated_at: 2026-11-13T17:00:47Z
parent: devilGo-ok51
---

---
# devilGo-mxda
title: 'ADR-002 note: parity anchor ca91d778f canonical, no re-pin'
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:38Z
updated_at: 2026-11-13T00:16:10Z
parent: devilGo-ok51
---

---
# devilGo-mxda
title: 'ADR-002 note: parity anchor ca91d778f canonical, no re-pin'
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:38Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# ADR-002 note: ca91d778f stays the canonical parity anchor

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-7).

## Goal
One paragraph in docs/adrs/adr-002.md recording: ca91d778f is the canonical
parity anchor; the checkout HEAD sits on fork branch
blaze-controller-turn-and-aim @ e1ad0c2b6 (additive surface, portable as
beans citing branch+commit); pinned-hunk line drift recorded (plrctrls.cpp
WalkInDir :1747->:1720, TurnToDir :1776->:1749, Interact fake-attack
:557->:533 at e1ad0c2b6) so nobody "fixes" citations against HEAD.

## Acceptance criteria
- [x] ADR-002 carries the no-re-pin paragraph with the drift table
- [x] No citation in docs/ or .beans changes commit (ca91d778f stays)

## Triage

Triage source: brainer session task-techdeptTriage-brainer (2026-11-12,
.hermes/lane/dg-chore/techdept.triage.md) — verdict CONFIRMED with the
scope noted above; parent filed this child from that verdict.

## Libraries

- stdlib only (no third-party deps at this stage)

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md — read fully before acting
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest correct change

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Docs are rules, not narration: ADR edits record decisions with upstream
   citations; requirements.md must never reference paths that do not exist in
   the tree (the dgo-parity lesson). Simple English, active voice, no em
   dashes, labeled links.

## Proof of Work

2026-11-13: ADR-002 anchor section landed (commit 14f3794 on chore/techdebt-fixes): canonical anchor ca91d778f recorded, drift table grep-verified (WalkInDir :1747->:1720, TurnToDir :1776->:1749, Interact :557->:532 at e1ad0c2b6; worker corrected one of the parent's draft numbers by grepping). AC2 verified parent-side: zero removed citation lines in the commit diff (existing citations untouched; the two new mentions are the section's own prose).
