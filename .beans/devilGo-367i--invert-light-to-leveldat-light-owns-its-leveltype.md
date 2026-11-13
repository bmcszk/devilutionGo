---
# devilGo-367i
title: 'Invert light to leveldat: light owns its LevelType'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-ok51
---

---
# devilGo-367i
title: 'Invert light to leveldat: light owns its LevelType'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:16Z
parent: devilGo-ok51
---

---
# devilGo-367i
title: 'Invert light to leveldat: light owns its LevelType'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:11Z
parent: devilGo-ok51
---

---
# devilGo-367i
title: 'Invert light to leveldat: light owns its LevelType'
status: todo
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# Invert light -> leveldat: light owns its LevelType

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-1).

## Goal
internal/light defines LevelType uint8 + seven DTYPE_* constants (ordinals
= leveldat parity contract, cited to loadsave.cpp SaveMonster write);
tables.go drops the diablo/gamedata/leveldat import (the ONLY standing
ADR-008 layer violation). diablo/levels/light.go maps leveldat.DungeonType
-> light.LevelType. Render callers keep the DTYPE_* alias names
(source-compatible; 1-file change per triage Q1).

## Acceptance criteria
- [ ] grep: internal/light imports no internal/diablo/*
- [ ] import-set test pins the layer rule
- [ ] light goldens (falloffs, cone, vision rays) unchanged

## Triage

Triage source: brainer session task-techdeptTriage-brainer (2026-11-12,
.hermes/lane/dg-chore/techdept.triage.md) — verdict CONFIRMED with the
scope noted above; parent filed this child from that verdict.

## Libraries

- stdlib only (no third-party deps at this stage)

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md — read fully before acting
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest correct change
/home/blaze/.agents/skills/go-dev/SKILL.md — Go structure discipline
/home/blaze/.agents/skills/go-tests/SKILL.md — golden-file + test matrix

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Layering (ADR-008): same as above — verify import direction first; generic
   mechanisms go in engine/game, Diablo semantics stay in diablo.
