---
# devilGo-ve6v
title: 'Docs sweep: retire dgo-parity citations'
status: completed
type: task
priority: low
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T17:00:47Z
parent: devilGo-ok51
---

---
# devilGo-ve6v
title: 'Docs sweep: retire dgo-parity citations'
status: todo
type: task
priority: low
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:25Z
parent: devilGo-ok51
---

---
# devilGo-ve6v
title: 'Docs sweep: retire dgo-parity citations'
status: todo
type: task
priority: low
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:11Z
parent: devilGo-ok51
---

---
# devilGo-ve6v
title: 'Docs sweep: retire dgo-parity citations'
status: todo
type: task
priority: low
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# Docs sweep: retire dgo-parity citations

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-9). Low.

## Goal
docs/requirements.md:19170,19201 point at the tools/goldens
parity-record/parity-compare subcommands; dg-graphics-part2.md gets a
one-line superseded note. 7aul bean mentions stay (historical record).

## Acceptance criteria
- [x] grep "cmd/dgo-parity" docs/ returns only historical notes

## Triage

Triage source: brainer session task-techdeptTriage-brainer (2026-11-12,
.hermes/lane/dg-chore/techdept.triage.md) — verdict CONFIRMED with the
scope noted above; parent filed this child from that verdict.

## Libraries

- stdlib only (no third-party deps at this stage)

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md — read fully before acting
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest correct change
/home/blaze/.agents/skills/go-dev/SKILL.md — repo conventions for docs edits
/home/blaze/.agents/skills/go-tests/SKILL.md — gate expectations

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Docs are rules, not narration: ADR edits record decisions with upstream
   citations; requirements.md must never reference paths that do not exist in
   the tree (the dgo-parity lesson). Simple English, active voice, no em
   dashes, labeled links.

## Proof of Work

2026-11-13: docs sweep landed (commit 786ccb7): 6 live citations fixed (3 requirements.md test-plan items incl. the go:generate line; the old flag list was fiction and now describes the real subcommands), superseded notes in requirements.md + research/dg-graphics-part2.md; historical records untouched. AC grep verified: cmd/dgo-parity now appears only inside retired-CLI labels/historical notes.
