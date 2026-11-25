---
# devilGo-uscq
title: 'persist: sweep retry honesty + ghost-slot file pin (hl3 review follow-ups)'
status: draft
type: task
priority: normal
created_at: 2026-11-25T16:52:39Z
updated_at: 2026-11-25T16:52:39Z
parent: devilGo-p1t8
---

# persist: sweep retry honesty + ghost-slot file pin (hl3 review follow-ups)

## Goal
Two non-blocking review findings on the per-hero store migration (hl3 review, PR #85), filed so they do not rot:
1. rehomeLegacySingleFile's comment claims a mid-failure sweep is retried on the next construction; the code does not implement retry (a half-swept state stays half-swept). Either implement idempotent retry (per-row completion markers make it naturally resumable) or correct the comment.
2. A corrupt NON-slot ghost file in the hero directory is not pinned by a test: ListHeroes probes only single_N.db names, so a ghost file is silently ignored — pin that behaviour (or decide it should error).

## Acceptance criteria
- [ ] The sweep comment matches the code (retry implemented idempotently OR comment corrected to state the non-retry + marker behaviour)
- [ ] A test pins ListHeroes/HasSlot behaviour when a corrupt ghost file shares the hero directory
- [ ] Full gates incl. -tags windowed; TDD RED evidence per law

## Libraries
stdlib + modernc.org/sqlite (existing).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Filed 2026-11-25 from the hl3 independent review (APPROVE-WITH-FINDINGS; findings non-blocking, neither lost data — reviewer's own words). Draft; deep triage owed (decide retry-vs-comment; check the ghost-file decision against upstream ReadHero-failure-silently-skips).
