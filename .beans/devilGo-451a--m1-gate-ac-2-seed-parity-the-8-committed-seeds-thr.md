---
# devilGo-451a
title: 'M1 gate: AC-2 seed parity — the 8 committed seeds through the walk'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-20T00:34:08Z
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
- [x] Walk completes on all 8 committed seeds, zero errors, with the
      committed fixture corpus or the DRLG source as the layout source
- [x] Per-seed trails frozen (values recorded in the bean), cross-process
      determinism proven (three processes + plain clone comparison)
- [x] Every seed's result visible in the test output matrix (named
      subtests per seed); any coverage gap disclosed, not dropped
- [x] Full gates green at HEAD through pre-commit + pre-push

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



## Progress log

- 2026-11-20T01:10Z: promoted (triage complete: planner plan + parent dep verification) and claimed — AC-2 dispatched under contract harness.ac2.md (worktree ftbu-s5, feat/m1-create-pack), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.ac2.md (require_edit harness.s5a.log, sentinel M1-REAL-L1-AC2-DONE). Worker: task-dg-m1-ac2-coder-coder (87 tool calls), branch feat/m1-create-pack @ 0260460 — pushed; parent verified ls-remote == HEAD.
- Defect found by the gate: newLevelImages hardcoded scenarioSeed 2588 for every walk (7/8 corpus walks ran 2588's cathedral); save/load seed words likewise fixed. RED of record: the matrix failed 7/8 seeds, each naming its own seed. Fix is harness-layer (integration only).
- Per-seed first-stage-divergent fold table (parent run, TestM1SeedParityFreezesThePerSeedTrails): 2588→a649c25d04554b50, 743271966→710f84bad15a4b10, 1383137027→da9eae7046a7f2d3, 844660068→b525c24896866ee6, 19770182→9bb8a566daca26be, 1522546307→7c486faf98d1de46, 125121312→963b8e0053535830, 1511478689→66a9f3162c8829e0; stages 1-3 shared, stage 4+ diverge per seed (trails differ as designed). Digest over the 8 folds df3cf04743d5e791… identical across two parent runs; worker's own six-run table (3 worktree + 3 plain clone at head) recorded in the lane log.
- Parent re-ran gates at 0260460: gofmt clean, vet rc0 (worktree), full suite ok (worker tail 54 pkgs; parent reran integration), beans-lint OK, push verified.
- Cross-review: brainer PASS/8 claims (review file harness.ac2.review.md) with one MEDIUM bookkeeping item — this table + AC ticks; both land in THIS update.
- Residual: none for the gate; AC-4 smoke still queued behind AC-3.
