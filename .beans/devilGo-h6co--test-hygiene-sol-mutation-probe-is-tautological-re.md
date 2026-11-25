---
# devilGo-h6co
title: 'Test hygiene: SOL mutation probe is tautological (re-review R1)'
status: todo
type: task
priority: normal
created_at: 2026-11-20T14:43:26Z
updated_at: 2026-11-25T12:12:30Z
parent: devilGo-97mn
blocked_by:
    - devilGo-9ukg
---

# Test hygiene: the SOL mutation probe is tautological (re-review R1)

## Goal
checkSOLMutationProbe (internal/diablo/levels/townlayout_test.go:199-213 @949c728
at 140bd2b) allocates a LOCAL zeroed slice and asserts byte 0 lacks the
solid bit — true by construction, can never fail, and never re-runs the
rock-border assertion on zeroed served data, while its doc comments and
the lane log claim it "proves the test bites". Either delete the probe
(the bite already lives in checkRockBorderSOL on the served data) or
make it genuinely re-assert checkRockBorderSOL against a zeroed copy of
layout.SOL. Also: checkRockBorderSOL silently returns when maxPiece < 426 —
prefer a hard failure there.

## Research
- Re-review finding R1 (.hermes/lane/ftbu-m1/harness.s6c.re-review.md).

## Acceptance criteria
- [ ] The probe either deleted or made to actually flip the rock-border
      assertion on zeroed served data (the biting version's evidence in
      the lane log)
- [ ] checkRockBorderSOL hard-fails (no silent return) when the corpus
      lacks piece 426

## Triage
Filed 2026-11-20 from the s6c re-review R1 (minor); small, self-contained
test-hygiene leaf.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage hl1 (2026-11-25, brainer task-hl1-triage, parent-verified against 949c728)

TRIAGED-OK; decision made: DELETE the probe and hard-fail both piece guards. The tautology is confirmed live (line numbers drifted: live :199-213 at 949c728, not :186-206 at 140bd2b) — checkSOLMutationProbe allocates make([]byte, len(layout.SOL)) and asserts on its own zeroed scratch, can never fail. The honest bite already exists: checkRockBorderSOL reads the SERVED layout.SOL[426] and fails when the solid bit is missing. Fix: delete the probe + its call (:155); convert both silent returns (checkRockBorderSOL maxPiece<426, checkSpawnPieceSOL spawnPiece>maxPiece) to t.Fatalf naming the seed/entry; fix the enclosing comment (:142-145) to name the two served-data assertions. Evidence-first: the bite run (scratch-zeroed byte, test fails at the rock-border message) pasted before the cleanup; scratch change never committed. Full triage: .hermes/lane/m2-hero-load/result.md.
