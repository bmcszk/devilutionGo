---
# devilGo-5fr4
title: 'M1 slice: selhero first slice — list/create/delete + class preview'
status: in-progress
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T15:21:06Z
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
- [x] Hero list renders from the real store; create (name + class) →
      KindCreatePack command → CreatedPack in the tick log; delete
      hits DeleteHero; a stats preview panel reads the typed loaders
      (Claimed 2026-11-20; PoW in ## Proof of Work)
- [x] Windowless/deterministic test path for the screen logic (no
      live-window dependency in CI); full gates green; RED first
      (Claimed 2026-11-20; PoW in ## Proof of Work)

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
- ~/.agents/skills/ponytail/SKILL.md

## Progress log

- 2026-11-20: RED recorded on the unchanged tree at 03f45ca (build-failure
  RED: selHeroScreen / newSelHeroScreen / FrontFlow.SetSelHero undefined,
  go test -tags windowed exit 1). Implemented to green; all five selhero
  tests pass windowless.

## Proof of Work

- RED (verbatim, unchanged tree 03f45ca):
  cmd/devilutionGo/selhero_test.go:239:11: undefined: selHeroScreen
  cmd/devilutionGo/selhero_test.go:261:17: undefined: newSelHeroScreen
  cmd/devilutionGo/selhero_test.go:262:12: boot.flow.SetSelHero undefined (type *scene.FrontFlow has no field or method SetSelHero)
  FAIL github.com/bmcszk/devilutionGo/cmd/devilutionGo [build failed]
- GREEN: TestSelHeroListFromStore, TestSelHeroCreateSubmitsCreatePack,
  TestSelHeroDeleteHitsDeleteHero, TestSelHeroEscapeWalksBack,
  TestSelHeroClassPreviewFromCommittedLoaders — all pass (windowed tag,
  windowless execution over a temp-dir SQLite store + MockTickSource).
- Surfaces consumed, nothing invented: persist ListHeroes/SaveHero/
  DeleteHero/OpenSQLiteStore; kernel KindCreatePack (world.
  RegisterCreatePack, createPackHandler -> KindCreatedPack); committed
  txtdata fixtures via txtdata.LoadClassAttributes / LoadClasses
  (internal/testdata corpus); ui.MenuList consumed as-is.
- The device lane now registers the CreatePack verb over bootPackSource
  (the scripted walk's frozen warrior pack); the bootstrap direct raise
  stays so the live-view baseline tile is byte-identical until the front
  flow reaches selhero.
- Trail check: M1_LANE_DIR statehash-run.json sha256
  f1ae95a0fd732040b3e5fafb17ab23d292f0abd4402d4939ec4bc63d98fc1120 —
  unchanged (pin f1ae95a0…).
- Gates at HEAD: gofmt -l empty; go vet rc0; ./scripts/lint.sh ./...
  0 issues; go test ./... -count=1 all ok; beans-lint OK (225 beans).

## Deviations

- Name entry accepts only confirm keys through the widget-kit key set;
  raw per-character text input belongs to the window layer's text seam
  (not in this slice's scope). typeName() is the test seam for it.
- The screen persists the created hero's row (SaveHero) alongside the
  kernel verb; the bean's create AC names both the KindCreatePack
  command and a hero the store lists. Slot policy (first free slot,
  99-slot cap) stays with the dcob umbrella bean as scoped.
