---
# devilGo-hi7t
title: Keymapper actions to kernel commands (bind callbacks + window routing)
status: draft
type: task
created_at: 2026-11-26T21:22:36Z
updated_at: 2026-11-26T21:22:36Z
parent: devilGo-k736
---

# Keymapper actions to kernel commands (bind callbacks + window routing)

## Goal
Bind the ~100 keymapper actions to gameplay verbs and route windowFlowKey through the keymapper. Deferred from the Options-slice triage: the verbs do not exist yet (no inventory/character/quest panel, no spell select); the model + defaults land first in the Options slice.

## Research
Defaults live at InitKeymapActions call sites (diablo.cpp:1839-2157). The model lands in internal/engine (Options slice); this bean wires the callbacks as kernel commands where sim-state is touched (kernel-law: sim mutations are commands) and plain UI toggles otherwise.

## Acceptance criteria
- [ ] A representative action set dispatches: sim-touching actions become kernel commands, UI-only actions toggle state; windowFlowKey routes through engine.Keymapper Press/Release; the guard order (unmapped/nil handler/disabled/chat-open) holds end to end.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split out of devilGo-t4fz by the Options-slice combined triage (task-opts-triage-brainer, 2026-11-26); blocked by the inventory/character panel beans it needs (parent resolves ids at claim time).
