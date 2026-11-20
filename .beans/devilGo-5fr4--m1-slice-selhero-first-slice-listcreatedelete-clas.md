---
# devilGo-5fr4
title: 'M1 slice: selhero first slice — list/create/delete + class preview'
status: draft
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-20T12:19:16Z
parent: devilGo-xih1
blocked_by:
    - devilGo-9qzw
    - devilGo-bh4m
    - devilGo-vf8q
---

# M1 slice: selhero first slice — list/create/delete + class preview

## Goal
First slice of the hero screens (dcob umbrella keeps full parity scope):
a selhero screen over the persist store (ListHeroes/SaveHero/DeleteHero,
internal/persist/store_sqlite.go:107-154) + class stat preview from the
typed classdat loaders + name entry. The CREATE submits KindCreatePack
through the live kernel (s6e) — the hero is born as a kernel verb, the
tick log carries CreatedPack.

## Research
- Plan: wave-s6-plan.md (facts cited; dcob's fix2 ACs name the full
  scope: Bard/Barb entries, row-exact class-stats parity,
  UiValidPlayerName, 99-slot cap — those stay with dcob, not this
  slice).
- emdz front screens land first (s6f); typed loaders from the loader
  slice; persist surface complete.

## Scope
- internal/game/ui screens + cmd wiring through the widget kit; sim
  untouched (create flows as the existing verb).

## Acceptance criteria
- [ ] Hero list renders from the real store; create (name + class) →
      KindCreatePack command → CreatedPack in the tick log; delete
      hits DeleteHero; a stats preview panel reads the typed loaders
- [ ] Windowless/deterministic test path for the screen logic (no
      live-window dependency in CI); full gates green; RED first

## Triage
Planner wave (wave-s6-plan.md s6g), filed 2026-11-20; blocked by the
front screens slice, the typed loaders, and the device lane.

## Libraries
stdlib + repo UI stack.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ebiten/input/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
