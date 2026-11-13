---
# devilGo-iq0d
title: 'k736 child: controls decisions to kernel commands (event bridge)'
status: todo
type: task
priority: critical
created_at: 2026-11-13T00:15:58Z
updated_at: 2026-11-13T00:16:16Z
parent: devilGo-k736
---

---
# devilGo-iq0d
title: 'k736 child: controls decisions to kernel commands (event bridge)'
status: todo
type: task
priority: critical
created_at: 2026-11-13T00:15:58Z
updated_at: 2026-11-13T00:16:11Z
parent: devilGo-k736
---

---
# devilGo-iq0d
title: 'k736 child: controls decisions to kernel commands (event bridge)'
status: todo
type: task
priority: critical
created_at: 2026-11-13T00:15:58Z
updated_at: 2026-11-13T00:15:58Z
parent: devilGo-k736
---

# k736 child: controls decisions -> kernel commands (the event bridge)

Parent: [Input epic](.beans/devilGo-k736--input-keyboard-mouse-gamepad-incl-fork-stick-turn-walk.md).
Tech-debt source: TD-2. M1-CRITICAL — the stick core + plrctrls action
machine are unreachable code until this lands.

## Goal
ebiten devices -> diablo/controls decision cores -> kernel Commands.
Define the action-family command/event kinds (spell select, icon cursor,
interact, store release) WITH gate rules, per-wave (no wholesale
kind-definition wave; ADR-018 law 5 fixes the home only). Submit sites
mirror demo/ticklog (the only current production Submit callers).

## Acceptance criteria
- [ ] Stick Motion/Interact produce gate-checked kernel commands replayable from the tick log
- [ ] PerformPrimaryAction dispatch rows map 1:1 to command kinds, each with a gate test (overlay refusal + dead-target drops pinned)
- [ ] Demo replay of a recorded pad session reproduces the same trace hash
- [ ] Ordering note honored: the TD-5 controls package move happens AFTER this wave

## Dependencies
blocked_by: c6wt open slices (Interact internals hooks) — verify at dispatch.

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
