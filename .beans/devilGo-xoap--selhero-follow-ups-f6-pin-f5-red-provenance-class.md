---
# devilGo-xoap
title: 'selhero follow-ups: F6 pin, F5 RED provenance, class-pick draw site'
status: draft
type: task
created_at: 2026-11-23T17:22:13Z
updated_at: 2026-11-23T17:22:13Z
---

# selhero follow-ups: F6 regression-blind pin, F5 RED provenance, class-pick draw site

## Goal
Close the selhero re-review's named conditions (harness.selhero.re-review.md)
before dcob/full-parity builds on the slice:

1. F6 pin: the committed multi-hero delete test passes on the OLD code
   too (the discriminating walk existed only in the RED draft). Re-pin
   with the walk that FAILS on the old focused-delete math (create
   three, focus the middle, delete, assert the middle row died and the
   others survive) — RED on the old math first (revert-harness or the
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
