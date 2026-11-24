---
# devilGo-ny3c
title: 'M1 slice: bridge walk rows as world verbs — KindMouseWalk + KindStickPush'
status: completed
type: task
priority: normal
created_at: 2026-11-20T13:40:39Z
updated_at: 2026-11-20T13:40:46Z
parent: devilGo-k736
---

# M1 slice: bridge walk rows as world verbs — KindMouseWalk + KindStickPush

## Goal
World handlers for the input bridge's walk decisions: KindMouseWalk and
KindStickPush get real state changes + PlayerMoved events (today zero
production handlers — the harness papers over it with the movementVerb
mapping), gate rules per ADR-018 law 6, and the harness glue retired so
decisions flow native rows.

## Research
- Plan: .hermes/lane/ftbu-m1/wave-s6-plan.md (fact base: zero handlers,
  inputbridge submit contract, movementVerb glue).
- Upstream walk semantics: plrctrls.cpp WalkInDir (pin e1ad0c2b6).

## Acceptance criteria
- [x] World handles KindMouseWalk + KindStickPush: real state changes,
      past-tense PlayerMoved events, silent-when-nothing holds; gate
      rules drop before dispatch (RuleNoStep), ordered-pipeline
      consistent
- [x] Stick semantics follow the controls decision core (nine sectors,
      axis-ordinal, tile-by-tile); the harness movementVerb mapping is
      retired — decisions flow native rows
- [x] Trail disposition recorded: single-walk statehash f1ae95a0…
      unchanged; the walked seed's fold a649c25d04554b50 re-proven
- [x] RED first: 10/12 named failures on the unchanged tree (incl. the
      RuleNoStep compile failure); gates green at HEAD; push verified

## Proof of Work
- Worker: task-dg-m1-s6b-coder-coder (117 tool calls), commit 7a87cdb
  on feat/m1-interactive-demo — pushed; parent verified ls-remote ==
  HEAD; world suite 63 tests ok, integration 36 ok (parent re-runs).
- Review: brainer APPROVE 8/8 (.hermes/lane/ftbu-m1/harness.s6b.review.md).

## Deviations and findings (recorded)
- F1 (recorded deviation): a hand-edited StickPush word A=Firm,B=0,C=0
  would walk South, while upstream WalkInDir intercepts NONE/NONE and
  stops (plrctrls.cpp:1726-1731). Unreachable from the decision core
  (nonzero magnitude always yields a nonzero axis) — no session or
  replay can carry it. Hardening draft filed (the stick-rule guard
  bean under Input) to extend stickRule or restate the cell.
- F3/F4 (notes): gate-pin wording on drops; MouseWalk cooldown not
  separately pinned — reach in later waves.

## Progress log

- 2026-11-20T14:40Z: FILED RETROACTIVELY (review finding F2, process slip — the coder was dispatched against the plan row before this bean existed). Work landed at 7a87cdb, parent-verified and reviewed before this filing; the ledger records reality as of the review. Status set completed with the evidence above.

## Triage
Planner wave (wave-s6-plan.md s6b), landed and reviewed 2026-11-20;
parent-verified at 7a87cdb.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
