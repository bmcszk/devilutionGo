---
# devilGo-bz7q
title: 'Surface C: inline CLX conversion in migrate assets'
status: todo
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T21:41:44Z
parent: devilGo-8csd
blocked_by:
    - devilGo-srb3
---

# Surface C: inline CLX conversion in migrate assets

## Goal
Per ADR-019 (Decision part 1): `migrate assets` conversion. The extraction pass converts extracted `.cel`, `.cl2`, `.pcx` records that carry frame-width inventory data into `.clx` using the existing internal/assets converters; records without width data keep today's copy-as-is behavior (extract.go:249-251). One write = ADR-005 embed parity ("the build embeds exactly the files the tool writes").

## Research (pre-validated)
- Parent review APPROVE (/home/blaze/github/dg-conv.review.md, task-dgconv-review-brainer).
- docs/research/dg-conversion-inventory.md: today the conversion is a SECOND program (dgo-assets convert), so what the build embeds does not equal what migrate writes — the split the user ruled out.
- Converters proven in internal/assets: cel_to_clx.go, cl2_to_clx.go, plus PCX handling; the l5fp craft covers the manifest loader/width fields.
- Widths source: the graphics inventory entries carry frames/width/width_source (internal/assets/manifest/manifest.go:91-96) — the inventory is the width oracle for convert decisions.

## Scope
- assets pass: for each convertible record whose inventory entry has width data -> CLX output replaces the raw file in the assets dir; manifest records updated (final format clx + the conversion inputs hashed, keeping idempotency summary stable across a second run).
- No-width records: copy as-is, format stays cel/cl2/pcx.
- dgo-assets convert remains available until Surface D (folded CLI path wins if both exist).
- Tests: synthetic fixtures only (never retail MPQ); round-trip the proven converter tests; idempotency test re-run hash stable.

## Acceptance criteria
- [ ] `migrate assets` on synthetic fixtures writes .clx for width-carrying records; copy-as-is otherwise (integration test with fixture CSV/TSV inputs, no MPQ)
- [ ] Idempotent: second run produces identical assets dir + identical hash summary (test)
- [ ] ADR-005 embed parity: a test asserts the embed source dir contains exactly what migrate wrote
- [ ] Full gates green

## Libraries

No new library bindings. Everything stays within the pinned catalog (ADR-009); internal/assets converters already committed.

## Skills for coder
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/beans/SKILL.md

## Dependencies & blockers
blockedBy: Surface A. Blocks Surface D. Parent: epic devilGo-8csd.

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.
