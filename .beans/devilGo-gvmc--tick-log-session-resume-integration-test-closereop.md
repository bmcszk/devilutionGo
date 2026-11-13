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
