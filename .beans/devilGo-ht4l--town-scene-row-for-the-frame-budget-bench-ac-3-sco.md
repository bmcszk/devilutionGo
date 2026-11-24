---
# devilGo-ht4l
title: Town-scene row for the frame budget bench (AC-3 scope completion)
status: completed
type: task
priority: normal
created_at: 2026-11-20T01:33:46Z
updated_at: 2026-11-20T12:34:29Z
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
- [x] Bench gains a town-scene subtest row measured through the same
      per-stage p95 method
- [x] Milestone AC wording (town+dungeon p95) evidenced by both rows

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


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.ht4l.detail.md (worker task-dg-m1-ht4l-coder-coder) on feat/m1-town-bench @ 7f61376 — pushed; parent verified ls-remote == HEAD (review OPERATOR-CHECK b also verified from disk refs).
- Both rows recorded (parent operator-run, review check a): cathedral total p95 5.618-5.703ms, TOWN row 5.538-5.773ms (world ~4.6-4.8, present ~1.0-1.1) — the milestone AC-3 wording (town+dungeon p95 under 16.67ms mid-hardware) now has two measured rows, both gates real.
- Review: brainer APPROVE 8/8 (harness.ht4l.review.md) — method shared with the cathedral row (no duplication), scene honest (retail records; disclosed CEL remap F1-low pending the town CEL corpus commit), row-table refactor diff-verified non-weakening, RED honest, scope integration-only.
- AC-3 wave gate (the frame-budget bean) scope now evidenced by both rows; F1 disposition: commit the town CEL when the corpus lane lands it, drop the remap.
- Residual: none blocking; town CEL corpus item named above.
