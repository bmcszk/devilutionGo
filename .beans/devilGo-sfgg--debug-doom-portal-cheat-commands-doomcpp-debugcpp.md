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

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers

None hard. Low priority (M2+ filler). Doom sprite needs asset pipeline (landed).
