---
# devilGo-367i
title: 'Invert light to leveldat: light owns its LevelType'
status: completed
type: task
priority: normal
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T17:00:46Z
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
- [x] grep: internal/light imports no internal/diablo/*
- [x] import-set test pins the layer rule
- [x] light goldens (falloffs, cone, vision rays) unchanged

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

## Proof of Work

2026-11-13, commit 5c6f84d on chore/techdebt-fixes: internal/light owns
LevelType + DTYPE_* (ordinals cited to upstream SaveMonster DTYPE
writes, verified at e1ad0c2b6); tables.go drops the leveldat import;
TestLightPackageImportsNoDiablo (parse-based, stdlib go/parser) pins the
layer rule; adapter leveldat.DungeonType -> light.LevelType landed in
the EXISTING home internal/diablo/levels/light.go (no parallel package).
Parent re-verified on the worktree: layer test + Falloffs/Cone/
VisionRays goldens PASS; goldens git-diff empty; sim-purity gate clean
(243 sources); full tree 49 pkgs 0 FAIL. The last standing ADR-008
layer violation is gone.
