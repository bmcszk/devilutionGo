---
# devilGo-bixn
title: 'ADR-003 note: save-story precedence (rows = save, tick log = replay oracle)'
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:38Z
updated_at: 2026-11-13T00:16:10Z
parent: devilGo-ok51
---

---
# devilGo-bixn
title: 'ADR-003 note: save-story precedence (rows = save, tick log = replay oracle)'
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:38Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# ADR-003 note: save-story precedence (rows = save, tick log = replay oracle)

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-4).

## Goal
One paragraph in ADR-003 + cross-ref sentence in ADR-018 law 4, following
upstream truth: SaveHero/SaveGame/SaveLevel rows are the session save
(upstream LoadGame); the tick log is the replay/desync artifact in the same
store, validated against rows the way upstream validates demo replay against
the reference hero save (pfile.cpp:657/:665, demomode.cpp:958/:970).
Mid-dungeon restore loads rows; snapshot+replay is verification, not a
competing loader.

## Acceptance criteria
- [ ] ADR-003 paragraph with upstream citations
- [ ] ADR-018 law 4 cross-ref sentence

## Triage

Triage source: brainer session task-techdeptTriage-brainer (2026-11-12,
.hermes/lane/dg-chore/techdept.triage.md) — verdict CONFIRMED with the
scope noted above; parent filed this child from that verdict.

## Libraries

- stdlib only (no third-party deps at this stage)

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md — read fully before acting
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest correct change
