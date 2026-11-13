---
# devilGo-iq0d
title: 'k736 child: controls decisions to kernel commands (event bridge)'
status: todo
type: task
priority: critical
created_at: 2026-11-13T00:15:58Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-k736
---

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

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
