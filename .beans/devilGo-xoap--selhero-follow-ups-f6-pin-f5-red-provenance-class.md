---
# devilGo-xoap
title: 'selhero follow-ups: F6 pin, F5 RED provenance, class-pick draw site'
status: todo
type: task
priority: normal
created_at: 2026-11-23T17:22:13Z
updated_at: 2026-11-25T12:12:30Z
---

# selhero follow-ups: F6 regression-blind pin, F5 RED provenance, class-pick draw site

## Goal
Close the selhero re-review's named conditions (harness.selhero.re-review.md)
before dcob/full-parity builds on the slice:

1. F6 pin: the committed multi-hero delete test passes on the OLD code
   too (the discriminating walk existed only in the RED draft). Re-pin
   with the walk that FAILS on the old focused-delete math (the walk must leave the LAST-highlighted hero NOT the last hero row — up-wrap from row 0; the committed Down-walk highlights Second last, the same hero the old offset math deleted) — RED on the old math first (revert-harness or the
   diff-draft), verbatim in the log.
2. F5 RED provenance: the recorded F5 RED line cites a draft artifact
   not reproducible from the committed test — re-record it from the
   committed artifact.
3. Class pick list draw site: the pick list has no render path yet —
   add it to ComposeFrontFrame's selhero surface (rows + highlight),
   pinned by a compose test.

## Acceptance criteria
- [ ] The F6 pin fails on the old math (evidence verbatim) and passes on
      the fixed code
- [ ] F5's RED re-recorded from the committed artifact
- [ ] The class pick list renders (compose test names it)
- [ ] Full gates green; trail unchanged

## Triage
Filed 2026-11-20 from the selhero re-review conditions (parent must
file before dcob — the reviewer's own recommendation).

## Libraries
stdlib + repo UI stack.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage hl1 (2026-11-25, brainer task-hl1-triage, parent-verified against 949c728)

TRIAGED-OK with one scope correction: all three items verified still open. F6 correction: the committed pin's Down-walk highlights Second last — the same hero the OLD offset math deleted — so the blind walk is ANY walk whose last-highlighted hero is the list's LAST hero row; the minimal discriminating walk is the UP-wrap from row 0 (First last-highlighted; old math: idx = 3-1-1 = Second, fixed code: tracked focused hero). X1 must paste the old-math revert run verbatim. F5: re-record the RED from the committed artifact (the pre-fix rogue pick emits ZERO events — the committed test fails at the ghost-row assertion, not the events one). X3: the class-pick list has no draw site — SelHeroContent exposes no create-mode rows (internal/present/scene/frontflow.go:34-44); X4 (optional, same commit): give lastCreateError a reader. Rides the dcob wave branch FIRST (before any D5/D6 delete work). Full triage: .hermes/lane/m2-hero-load/result.md.
