---
# devilGo-ht4l
title: Town-scene row for the frame budget bench (AC-3 scope completion)
status: in-progress
type: task
priority: normal
created_at: 2026-11-20T01:33:46Z
updated_at: 2026-11-20T12:19:21Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-prcc
---

# Town-scene row for the frame budget bench (AC-3 scope completion)

## Goal
The AC-3 bench measures the cathedral parity scene only (Cathedral L1,
seed 2588, hero at entry 77,46, radius-10 cone). The milestone AC reads
"60fps rendering of town+dungeon": this bean adds the TOWN scene row
(town compose chain, panes + towner sprites per the town render lane)
so the budget covers both surfaces it names.

## Research
- Review finding on the M1 frame budget gate bean (cathedral-only scene): bench measures
  cathedral only; town covered by pipeline identity, not a measured
  town scene. Suggested follow-up by the reviewer.

## Acceptance criteria
- [ ] Bench gains a town-scene subtest row measured through the same
      per-stage p95 method
- [ ] Milestone AC wording (town+dungeon p95) evidenced by both rows

## Triage
Filed 2026-11-20 from the AC-3 brainer review (finding 1, medium);
needs its own triage pass when the town render lane opens the scene
surfaces it needs.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md



## Progress log

- 2026-11-20T03:20Z: promoted post-merge (blocker completed; triage from the AC-3 review finding) and claimed — dispatched under contract harness.ht4l.md (fresh worktree, branch feat/m1-town-bench), status flip.
