---
# devilGo-6851
title: 'M1 slice: create-pack verb — KindCreatePack + CreatedPack + world create-from-pack'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-19T23:24:38Z
parent: devilGo-6xu2
---

# M1 slice: create-pack verb — KindCreatePack + CreatedPack + world create-from-pack

## Goal
Promote the harness walk's scripted warrior creation to a kernel verb: a
`KindCreatePack` command with integer payloads (class id, spawn tile), a
world create-from-pack handler building the warrior from committed pack
/classdata, and a `CreatedPack` event through the tick log. ADR-015
integers-only payloads; ADR-018 past-tense events; consumer-defined
interface where world needs pack data.

## Research
- Kind inventory verified 2026-11-19 at master 05a792b: no create-pack
  verb exists (`grep Kind internal/game/kernel/command.go`).
- The scripted glue: internal/integration session_test.go:323-327
  raises `warriorPack(spawnTile)` test-side; that stage loses the struct
  mutation.
- Pack/classdata seams exist: persist pack serializers + committed
  classes txtdata (upstream classes/warrior/starting-loadout columns).
- TDD red-first: the verb test fails on the current tree; RED output is
  bean evidence, recorded in the lane log with the observed failure.

## Scope
- internal/game/kernel (command + event kinds/gate), internal/game/world
  (handler), internal/integration (walk switch). internal/persist ONLY
  if a pack-reader seam is missing (small; report in the lane log).
- StateHash: create-through-kernel is state-equivalent; a trail value
  change is allowed and disclosed with the fresh deterministic trail
  across three processes + plain clone (the s3 precedent).

## Acceptance criteria
- [x] KindCreatePack + CreatedPack exist with integer payloads; the walk
      create stage flows through kernel events (harness struct mutation
      retired for create)
- [x] Zero-events pins extended: CreatedPack counted; silent-create
      (no-submit noop) covered by a named test
- [x] Full gates green at HEAD through pre-commit + pre-push (gofmt,
      vet, lint.sh sim purity, go test -count=1, beans-lint); statehash
      trail deterministic across three processes + plain clone
- [x] RED evidence observed and recorded before the fix (failing verb
      event test on the unchanged tree)

## Triage
Planner wave: delegated brainer plan pass + parent verification vs the
live tree, 2026-11-19. Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md.
Parent overrides of plan drift recorded in the milestone body's
disposition section (squash sha corrected to 2f6a10f; pin corrected to
upstream HEAD e1ad0c2b6; D1/D2 rows re-read as keep-as-deviation with
resolution drafts filed).

## Libraries
stdlib only (kernel/world/integration). The repo persist stack stays
untouched unless the pack-reader seam is missing.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md — Go hard rules, layering discipline
- ~/.agents/skills/tdd/SKILL.md — red-first contract
- ~/.agents/skills/coding-rules/SKILL.md — style + TDD law, gate honesty
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — lane, hook, scope laws

## Progress log

- 2026-11-19T23:50Z: claimed — s5a dispatched under contract harness.s5a.md (worktree ftbu-s5, feat/m1-create-pack), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.s5a.md (require_edit harness.s5a.log, sentinel M1-REAL-L1-S5A-DONE). Worker: task-dg-m1-s5a-coder-coder (203 tool calls), branch feat/m1-create-pack @ ac9dbb4 — pushed; parent verified ls-remote == HEAD.
- RED of record: compiling probe on the unchanged tree failed by name ("tick log carries 0 CreatePack commands, want 1"); premise held; probe deleted at green (s3 precedent).
- Parent re-ran gates at ac9dbb4: gofmt clean, go vet rc0, lint.sh 0 issues (sim purity 254 sources), go test ./... -count=1 all pass, beans-lint OK (216 beans). Statehash trail f1ae95a0fd73… byte-identical across three worker runs + plain clone + two parent runs (create is state-equivalent, disclosed as unchanged).
- Cross-review: brainer APPROVE, 8/8 claims PASS (review: .hermes/lane/ftbu-m1/harness.s5a.review.md, sentinel M1-REAL-L1-S5A-REVIEW-DONE). Both OPERATOR-CHECK items executed by the parent at ac9dbb4 (full suite rc0; ls-remote sha match). F1 (PlayerPack.Px/Py retired with warriorPack; upstream pack.h plrx/plry) recorded as a note on the save/load slice bean; F2 non-blocking, no action.
- Residual: none for this slice; the wave continues with save/load verbs.
