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
- [ ] grep "cmd/dgo-parity" docs/ returns only historical notes

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
