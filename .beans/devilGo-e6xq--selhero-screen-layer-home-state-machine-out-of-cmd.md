---
# devilGo-e6xq
title: 'selhero screen layer home: state machine out of cmd wiring (ADR-008 ui map)'
status: draft
type: task
priority: normal
created_at: 2026-11-25T14:09:59Z
updated_at: 2026-11-25T14:09:59Z
parent: devilGo-xih1
---

# Home for the selhero screen layer: screen state machine vs cmd wiring

## Goal
Record (then implement or reject) the layer home of the selhero SCREEN. Today `cmd/devilutionGo/selhero.go` (windowed, package main) holds the whole screen state machine — create steps, focus tracking, class-ordinal mapping, slot policy, delete confirm — while ADR-008's LAYOUT MAP maps ui to `game/ui` and the widget kit already lives there (internal/game/ui: widget, belt, flasks, panels).

## Research
- ADR-008 LAYOUT MAP: "ui to game/ui"; the intended-homes note covers stick math (internal/game/input) and animation (engine/animation) but records NO home for front-flow screens.
- dcob's original scope said internal/ui/selhero.go — superseded by the LAYOUT MAP, never re-homed; the 5fr4/#81 slice built it as boot composition (it glues persist.SQLiteStore + kernel command submission + present/scene FrontFlow — all things the boot owns).
- The screen's decision logic is windowless-testable and Diablo-neutral in shape (list/create/delete/confirm over a store seam); only the labels and class data are Diablo.

## Scope
- Triage decides: (a) move the screen state machine to internal/game/ui (or diablo/ui if judged Diablo-specific) leaving cmd with wiring + composition only, or (b) record cmd as the sanctioned home for front-flow screens in ADR-008. Either way the ADR gets the missing sentence.

## Acceptance criteria
- [ ] ADR-008 carries the recorded home for front-flow screens (move or stay)
- [ ] If move: screen state machine lives in the recorded package, cmd keeps wiring only; windowless tests move with it; full gates green incl. -tags windowed

## Libraries
stdlib + repo UI stack.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Filed 2026-11-25 from the owner's question ("why is the selhero thing not in domain packages") during the hl1 wave close. Draft — deep triage (brainer pass) owed: decide move-vs-record against ADR-008 and the 5fr4 composition precedent before any wave.
