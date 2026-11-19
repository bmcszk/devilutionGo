---
# devilGo-sj5a
title: 'M1 slice: real kernel descend/loot verbs + parity-true L1 in the harness walk'
status: in-progress
type: task
priority: normal
tags:
    - m1
    - input
    - render
created_at: 2026-11-19T00:00:03Z
updated_at: 2026-11-19T01:18:21Z
parent: devilGo-k736
blocked_by:
    - devilGo-ftbu
---

## Goal

Promote the harness's scripted dungeon fiction to parity-true structure,
per the seam inventory left by the AC-1 harness (PR #76 / #76 lane log):

1. Kernel verbs: promote descend + loot from scripted harness state
   mutations to real internal/game/kernel verbs (KindDescend, KindLoot),
   with their events through the tick log and Gate rules; create already
   has its verb.
2. Real L1: load the committed parity-true cathedral L1 fixture
   (dg-tiles/76tu corpus) into the walk's world so the scripted session
   runs on the REAL layout; movement respects walls.
3. Harness assertions extended: StateHash folds the layout; descend
   stage freezes the load trail; no synthetic-grid residue.

## Acceptance criteria

- [ ] KindDescend/KindLoot verbs + events exist; walk uses them (no
      harness-struct mutation for descend/loot anymore)
- [ ] Walk world loads parity-true L1; pathing respects walls; hash
      folds layout
- [ ] Full gates green through hooks (gofmt/vet/lint/go test/beans-lint)

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md

## Notes

- Evidence base: internal/integration session tests + harness.s1.log
  seam-inventory addendum; l1_parity_test.go is the fixture oracle.
- Parent composes PoW at close; workers never touch .beans.

## Progress log

- 2026-11-19T01:38Z: claimed (in-progress) — s3 dispatched under contract harness.s3.md (dg-m1-real-l1-s3), status bean flip.

## Proof of Work

- Contract: `.hermes/lane/ftbu-m1/harness.s3.md` (require_edit lane log, sentinel M1-REAL-L1-S3-DONE).
- Worker: dg-m1-real-l1-s3-coder (session task-dg-m1-real-l1-s3-coder, 263 tool calls), branch feat/m1-real-l1 @ 045b8f7 — pushed.
- Lane log: .hermes/lane/ftbu-m1/harness.s3.log (sentinel present, parent-grepped).

# Progress log (cont.)

- Phase 77 draft: cross-review 2 rounds (APPROVE/APPROVE), CI green on 44342504412, ready for review + mergeable at head 6b44235. Bean stays in-progress until the user merges; post-merge drill + AC-2/3/4 remaining checks next wave per the seam inventory.
