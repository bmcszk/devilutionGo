---
# devilGo-gvmc
title: Tick-log session-resume integration test (close/reopen/replay/continue)
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-ok51
---

---
# devilGo-gvmc
title: Tick-log session-resume integration test (close/reopen/replay/continue)
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:16Z
parent: devilGo-ok51
---

---
# devilGo-gvmc
title: Tick-log session-resume integration test (close/reopen/replay/continue)
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:16:11Z
parent: devilGo-ok51
---

---
# devilGo-gvmc
title: Tick-log session-resume integration test (close/reopen/replay/continue)
status: todo
type: task
priority: high
created_at: 2026-11-13T00:15:39Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-ok51
---

# Tick-log session-resume integration test (close -> reopen -> replay -> continue)

Parent: [Tech debt](.beans/devilGo-ok51--tech-debt.md) (TD-8). HIGH: the
ADR-018 law 4 load-bearing promise, gates the M1 save/load AC.

## Goal
One persist integration test on a real temp .db: record 1500 ticks
(snapshots every 500) -> Close -> reopen same path -> replay from last
snapshot -> continue K ticks through a second TickLogWriter -> per-tick
trail tail, FoldChecksums and world hash equal a from-scratch N+K run.
Reuse newSeededRig + requireSnapshotTrailMatchesTail seams.

## Acceptance criteria
- [ ] RED first (fails today on the close/reopen+continue spine; run recorded)
- [ ] Green: go test ./internal/persist with the new test

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

1. Save story (ADR-003 + triage TD-4): SQLite rows (SaveHero/SaveGame/
   SaveLevel) are THE session save; the tick log is the replay/desync oracle
   validated against rows (upstream pfile.cpp:657/:665 model) — never a
   competing loader.
2. Data access (repo law): every DB read/write goes through sqlc-generated
   queries; raw SQL is a contract breach. New schema file => added to
   sqlc.yaml in the same commit; `just sqlc-check` in the gate list.
