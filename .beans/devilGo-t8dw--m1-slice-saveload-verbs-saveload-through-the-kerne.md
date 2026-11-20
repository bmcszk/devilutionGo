---
# devilGo-t8dw
title: 'M1 slice: save/load verbs — save/load through the kernel'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-20T00:14:27Z
parent: devilGo-6xu2
blocked_by:
    - devilGo-6851
---

# M1 slice: save/load verbs — save/load flow through the kernel

## Goal
Promote the harness walk's scripted save/load to kernel verbs: commands
(kind(s) SaveSlot/LoadSlot with integer payloads: slot, seed) wired to
the persist SQLite store through a consumer-defined world seam, and
past-tense `SavedSlot`/`LoadedSlot` events through the tick log.
The walk's test-side saved/loaded flags then derive from kernel events.

## Research
- Scripted glue verified: internal/integration save()/load() call the
  persist surfaces directly and stamp `saved=true` (:391) /
  `loaded=true` (session_stage_test.go:526); GameHeader flag bytes come
  from flagByte(s.saved)/flagByte(s.loaded) (modeWord :236-237).
- Persist has the SQLite store + SaveHero/SaveGame/SaveLevel; the seam
  question is the world-to-store interface (consumer-defined, ADR-008).
- TDD red-first: event tests on the unchanged tree; RED recorded.

## Scope
- internal/game/kernel, internal/game/world, internal/integration.
- Determinism: save/load must not perturb the statehash trail except at
  the disclosed border (event log growth that pins count by design).
- Kernel silent-handler law — a SaveSlot that saves nothing (no
  session) stays silent, no event.

## Acceptance criteria
- [x] Save/Load verb kinds + SavedSlot/LoadedSlot events exist with
      integer payloads; the walk save/load stages flow through the
      kernel (test-side flag stamping retired; flags derive from events)
- [x] Zero-events pins extended (no-submit save stays silent, named
      test); event-count pins count the new kinds
- [x] Full gates green at HEAD through pre-commit + pre-push; statehash
      trail deterministic across three processes + plain clone
- [x] Roundtrip honest: LoadedSlot after SaveSlot restores the walk
      state (a restore-equivalence test names what loaded state equals)

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: the create-pack
slice (the created warrior is the state the save slot freezes).

## Libraries
stdlib + the repo persist/sqlite stack as-is; no new dependencies.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md — Go hard rules, interfaces at consumers
- ~/.agents/skills/go-tests/SKILL.md — test design (black box, no mocks)
- ~/.agents/skills/tdd/SKILL.md — red-first contract
- ~/.agents/skills/coding-rules/SKILL.md — style + gate honesty
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — lane, hook, scope laws



## Notes

- Review F1 from the create-pack cross-review: heroPack stopped populating PlayerPack.Px/Py (upstream pack.h carries plrx/plry). When this slice wires hero position from the pack, decide whether CreatePack's spawn tile lands there; hash currently never folds them.



## Progress log

- 2026-11-20T00:30Z: claimed — s5b dispatched under contract harness.s5b.md (worktree ftbu-s5, feat/m1-create-pack; create-pack slice landed as its base), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.s5b.md (require_edit harness.s5a.log, sentinel M1-REAL-L1-S5B-DONE). Worker: task-dg-m1-s5b-coder-coder (179 tool calls), branch feat/m1-create-pack @ 5ef28dc — pushed; parent verified ls-remote == HEAD.
- RED of record: compiling probe on the unchanged tree — "tick log carries 0 SaveSlot commands" / "0 LoadSlot commands"; premise held; probe deleted at green.
- Parent re-ran gates at 5ef28dc: gofmt clean, go vet rc0, lint.sh 0 issues, go test ./... -count=1 54 pkgs ok / 0 FAIL, beans-lint OK. Determinism OPERATOR-CHECK (review's own command, post-mkdir): three runs f1ae95a0fd73… byte-identical — trail unchanged, save/load state-equivalent.
- Cross-review: brainer APPROVE, 9/9 claims PASS (.hermes/lane/ftbu-m1/harness.s5b.review.md). Findings F1-F4 all non-blocking with named dispositions: F1/F2 pack-parity bytes + cross-connection durability land with the hero-persistence bean; F3 wrap the real error when the slotArchive fill graduates from harness glue; F4 level-seed tail is a recorded deviation, persist lane reconciles the block format before any byte-parity claim.
- Residual: F1-F4 dispositions above (none blocking); D1/D2 registry waves remain filed as drafts.
