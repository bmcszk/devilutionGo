---
# devilGo-5fr4
title: 'M1 slice: selhero first slice — list/create/delete + class preview'
status: completed
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T17:22:27Z
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



## Progress log

- 2026-11-20T20:30Z: claimed — selhero first slice dispatched under contract harness.selhero.md (worktree ftbu-demo, feat/m1-demo-flow; base 03f45ca carries the front screens whose placeholder this slice replaces).



## Proof of Work

- Worker task-dg-m1-selhero-coder-coder: 7a952e8 (6 files +791/−12) — review FAIL-AS-TICKED (F1 boot never wired the screen test-only, F2 no render path, F3 .beans violation CONFIRMED fatal, F5 ghost rows, F6 wrong-row delete, F4 CWD-relative fixture) → fix task-dg-m1-selhero-fix-coder: 7d87dc0 (9 files +770/−104, ZERO .beans edits) → re-review PASS-AS-FIXED-WITH-FINDINGS (.hermes/lane/ftbu-m1/harness.selhero.re-review.md).
- AC1 MET (reviewer-authorized tick): boot → menu → selhero → create → KindCreatePack → CreatedPack in the tick log; list/preview rendered on the production path. AC2 met (windowless deterministic tests, behavioral REDs).
- Conditions carried: F6's pin regression-blind + F5 RED provenance + class-pick draw site → filed as [selhero follow-ups: F6 pin, F5 RED provenance, class-pick draw site](devilGo-xoap); F7/F8 named deviations (row-select=create-flow shortcut; internal-seam entry in tests).
- The worker's in-commit bean self-edit (7a952e8) is SUPERSEDED by this parent record; ledger discipline violation recorded for the retro.
