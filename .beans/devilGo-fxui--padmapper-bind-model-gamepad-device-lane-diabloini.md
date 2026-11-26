---
# devilGo-fxui
title: Padmapper bind model + gamepad device lane (diablo.ini Padmapping)
status: draft
type: task
created_at: 2026-11-26T21:22:36Z
updated_at: 2026-11-26T21:22:36Z
parent: devilGo-k736
---

# Padmapper bind model + gamepad device lane (diablo.ini Padmapping)

## Goal
PadmapperOptions port: 18-button name table (options.cpp:1328-1352), Action LoadFromIni :1388-1433 (splits on +, modifier+button), SaveToIni :1436-1456, enum ordinals controller_buttons.h:12-33, ~60 action defaults (diablo.cpp:2160-2653), dispatch (controls/padmapper.cpp). Deferred from the Options-slice triage: zero Go gamepad code exists (no ebiten.Gamepad/ControllerButton in internal/), so the table has nothing to dispatch to until the device lane lands.

## Research
Anchors pinned at e1ad0c2b6 in the combined triage (.hermes/lane/options-slice/triage.md, E4). Blocked on the generic device layer (ADR-008 internal/game/input) and the stick/turn lane.

## Acceptance criteria
- [ ] Padmapper model + ini section 1:1 with X:options.cpp PadmapperOptions, defaults fixture measured from diablo.cpp:2160-2653, dispatch wired into the device lane.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/port-parity-testing/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split out of devilGo-t4fz by the Options-slice combined triage (task-opts-triage-brainer, 2026-11-26).
