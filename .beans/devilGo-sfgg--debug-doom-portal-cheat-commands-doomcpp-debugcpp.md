---
# devilGo-sfgg
title: Debug doom portal + cheat commands (doom.cpp + debug.cpp subset)
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-13T00:33:35Z
parent: devilGo-yjuv
---

---
# devilGo-sfgg
title: Debug doom portal + cheat commands (doom.cpp + debug.cpp subset)
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-10-29T16:14:42Z
parent: devilGo-yjuv
---

## Goal

Debug/cheat layer: `doom.cpp` (44 lines — doom portal easter egg) + the
`debug.cpp` command subset worth porting (spawn item/monster cheats, level
jump) gated behind a build tag or `--debug` flag, matching devilutionX's
`#ifdef _DEBUG` spirit.

## Research

Source/doom.cpp:1-44 (QuestNoSpan?, doom sprite load/destroy); Source/debug.cpp
debug command table (in devilutionX behind compile flags); diablo.cpp
`CheckDebugCommand` call sites; decide parity policy: cheats that mutate state
MUST use the same seeded RNG calls as C++ or be excluded from parity tests.

## Scope

internal/game/debug.go (command router) + cmd wiring (flag). Doom portal =
pure easter egg, tiny; port for completeness. Excluded from StateHash parity
runs unless commands are no-ops.

## Acceptance criteria

- [ ] Doom portal: load sprite, quest trigger, remove — matches C++ (44-line file, full port)
- [ ] Debug commands routed behind flag/build tag; default builds unaffected
- [ ] RNG-using cheats call the same GenerateRnd sequence as C++ (documented per command)
- [ ] README/dev docs: how to enable, command list
- [ ] Zero behavior change in non-debug builds (CI test asserts flag-off parity)

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers

None hard. Low priority (M2+ filler). Doom sprite needs asset pipeline (landed).

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).

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
