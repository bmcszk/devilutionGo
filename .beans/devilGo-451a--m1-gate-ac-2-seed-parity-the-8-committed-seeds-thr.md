---
# devilGo-451a
title: 'M1 gate: AC-2 seed parity — the 8 committed seeds through the walk'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-19T22:33:24Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-6851
    - devilGo-t8dw
---

# M1 gate: AC-2 seed parity — the 8 committed seeds through the walk

## Goal
Run the scripted descend/loot walk on all eight committed DRLG seeds
(the internal integration committedSeeds corpus: 2588, 743271966,
1383137027, 844660068, 19770182, 1522546307, 125121312, 1511478689) and
freeze the per-seed StateHash trail as the AC-2 parity gate: same trail
across three processes + a plain clone, per-seed values recorded in this
bean.

## Research
- The walk (internal/integration runScriptedWalk(seed)) currently pins
  seed 2588 against the committed parity-true cathedral fixture corpus.
- The DRLG golden corpus covers all eight seeds (internal/testdata/drlg,
  committed through the 76tu/4e0x/gico lanes).
- Coverage gap class: where a per-seed fixture set the walk needs is
  missing, the worker discloses the gap in the lane log and files it in
  the report rather than silently reducing the seed set. Silent scope
  reduction = review REQUEST-CHANGES.

## Scope
- internal/integration (+ testsupport if a per-seed fixture set needs
  generating through the existing committed generator; no new tools).

## Acceptance criteria
- [ ] Walk completes on all 8 committed seeds, zero errors, with the
      committed fixture corpus or the DRLG source as the layout source
- [ ] Per-seed trails frozen (values recorded in the bean), cross-process
      determinism proven (three processes + plain clone comparison)
- [ ] Every seed's result visible in the test output matrix (named
      subtests per seed); any coverage gap disclosed, not dropped
- [ ] Full gates green at HEAD through pre-commit + pre-push

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: create-pack and
save/load verb slices (the gate exercises the full verb surface).

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md — subtest matrix design
- ~/.agents/skills/go-dev/SKILL.md — Go hard rules
- ~/.agents/skills/coding-rules/SKILL.md — style + gate honesty
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — lane + oracle laws
