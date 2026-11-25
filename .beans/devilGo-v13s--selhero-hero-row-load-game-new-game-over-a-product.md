---
# devilGo-v13s
title: 'selhero: hero-row Load Game / New Game over a production SessionArchive'
status: completed
type: task
priority: normal
created_at: 2026-11-25T12:12:09Z
updated_at: 2026-11-25T15:51:59Z
parent: devilGo-xih1
---

# selhero: hero-row Load Game / New Game over a production SessionArchive

## Goal
Activating a hero ROW with a saved game offers Load Game / New Game (upstream SelheroListSelect, selhero.cpp:217-236 @e1ad0c2b6); Load submits KindLoadSlot and the next settled tick logs KindLoadedSlot; a hero with no saved game goes straight to new-game (:239). Includes migrating the boot's hero list to the ADR-003 per-hero single_N.db store model (D1 decision recorded in the dcob triage block).

## Research
- Brainer triage hl1 (task-hl1-triage, 2026-11-25; .hermes/lane/m2-hero-load/result.md section dcob 2d): upstream flow verified at e1ad0c2b6 (anchor ca91d778f stays canonical); gbLoadGame from SELHERO_CONTINUE (menu.cpp:130).
- Seams that exist: kernel verbs KindSaveSlot/KindLoadSlot + events KindSavedSlot/KindLoadedSlot (internal/game/kernel/command.go:125-150), world.RegisterSaveLoad handlers (internal/game/world/session.go:233-281), SessionArchive seam (session.go:87-97). Missing: production SessionArchive (only the test-local slotArchive in internal/integration/session_stage_test.go), RegisterSaveLoad not registered in bootDeviceLane (device_lane.go:108-115), a boot LevelSource.
- ADR-003 "one hero = one .db file" (docs/adrs/adr-003.md:12); ADR-008 world owns no storage.

## Scope
- Production SessionArchive adapter over persist.SQLiteStore in the diablo layer (promote the tested slotArchive shape out of internal/integration); NOT internal/game, NOT internal/persist importing world.
- RegisterSaveLoad in the boot; windowed-boot LevelSource disclosure updated.
- Per-hero single_N.db enumeration (first-free-slot = first missing file number; MAX_CHARACTERS 99 cap enforced at enumeration).
- Hero-row load/new-game list in the selhero UI + compose draw site.

## Acceptance criteria
- [x] **S2+S1 — Production `SessionArchive`** in `internal/persist/session_archive.go`: `NewSessionArchive(store *SQLiteStore, packs world.PackSource) world.SessionArchive`; `SaveSlot` = LoadHero(keep Name/CreatedAt)→SaveHero + SaveGame("single", GameHeader) + SaveLevel(CurrLevel, block w/ 4-byte LE seed tail); `LoadSlot` = `ErrNoSlot` iff `errors.Is(err, ErrHeroNotFound)`, else wrapped `%w`; `HasGame(ctx, slot, kind)` via `LoadGame`+`ErrHeroNotFound`. **RED first:** `internal/persist/session_archive_test.go` `TestSessionArchiveRoundtripMatchesRecord` (save `{Slot:3, LevelSeed:0xC0FFEE, Level:0, Loaded:false, Hero:{75,68}, Pack: warrior}` → Close → reopen → LoadSlot(3) → DeepEqual), `TestSessionArchiveEmptySlotIsErrNoSlot`, `TestSessionArchiveHasGameFollowsTheGameRow`. **RED cmd:** `go test -count=1 -run 'TestSessionArchive' ./internal/persist` → `undefined: persist.NewSessionArchive`. Sentinel: `go test -count=1 ./internal/persist` ok; `just sqlc-check` clean (no query added).
- [x] **S4 — Lane registers SaveLoad** (`device_lane.go`, untagged): `attachSaveLoad(archive, levels)` registers `world.RegisterSaveLoad`; default archive refuses (`ErrNoSlot`/error). **RED first:** `TestDeviceLane_AttachSaveLoadLetsLoadSlotRestoreARecord` (SaveSlot → frame → LoadSlot → frame → Trace holds SavedSlot then LoadedSlot) + `TestDeviceLane_LoadSlotWithoutAttachIsSilent` (pin). **RED cmd:** `go test -count=1 -run 'TestDeviceLane_(AttachSaveLoad|LoadSlotWithoutAttach)' ./cmd/devilutionGo` → `lane.attachSaveLoad undefined`.
- [x] **S5 — Load outcome hook** (`device_lane.go`): `pendingLoad`, `onLoadEvent`, `deliverLoadEvents` beside the create pair. **RED first:** assert `onLoadEvent` fires once with `KindLoadedSlot` and `A==slot`. **RED cmd:** same as S4 (`lane.pendingLoad undefined`).
- [x] **S6 — Hero-row flow** (`selhero.go`, windowed): hero row with game → Load/New list (`selheroLoadStep`); without → `startNewGame` (CreatePack for the existing row — the Go stand-in for SELHERO_NEW_DUNGEON). Load peeks `archive.LoadSlot(slot)` for the seed (the handler compares B==LevelSeed) then submits `KindLoadSlot{A: slot, B: seed}`. **RED first:** `TestSelHeroHeroRowWithSavedGameOffersLoadNew`, `TestSelHeroHeroRowWithoutGameStartsNewGame`, `TestSelHeroLoadSubmitsKindLoadSlotAndSettles`. **RED cmd:** `go test -count=1 -tags windowed -run 'TestSelHeroHeroRow|TestSelHeroLoad' ./cmd/devilutionGo` → `s.loadMenu undefined`. The F7-encoding tests (`TestSelHeroClassPickRendersClassRows` hero-row drive, `TestSelHeroFrameRendersHeroRows`) flip RED and are updated in the same commit — Create is driven via the Create row.
- [x] **S7 — Band draws the Load/New rows** (`scene/frontflow.go`, `frontflow_render.go`, windowed): `SelHeroLoadRows/SelHeroLoadFocused` + first branch in `selHeroBand` (X3 precedent). **RED first:** `TestSelHeroLoadStepRendersTwoRows` (band rows == 2). **RED cmd:** `go test -count=1 -tags windowed -run TestSelHeroLoadStepRenders ./cmd/devilutionGo`.
- [x] **S8 — Load/New-Game enters the game** (`scene/frontflow.go` + `window_ebiten.go`, windowed): `FrontFlow.EnterGame()` sets `FrontGame`; the front-screen branch falls through to the movement lane on it. **RED first:** `TestWindowUpdateLoadEntersTheGame` — saved hero + SaveSlot, hero row → Load Game → N Updates → `flow.Screen()==scene.FrontGame` and the hero tile restored. **RED cmd:** `go test -count=1 -tags windowed -run TestWindowUpdateLoadEntersTheGame ./cmd/devilutionGo` → `scene.FrontGame undefined`. Boot-comment disclosure rewrite rides this commit (town-only LevelSource; stale mouse-lane note fixed).
- [x] **PoW:** per-slice RED pasted; `gofmt -l` empty; `go vet ./...` rc 0; `just lint` 0 issues; `go test -count=1 ./...` ok; windowed cmd+scene without assets ok (AC4 env-gate excepted); with assets cmd ok incl. AC4 and scene fails ONLY on the pre-existing TestScreenshotCaptureWriter panic (draft bean, out of scope); `just sqlc-check` clean; headless no-display pin green.
- [x] Deviations recorded in the bean: (1) adapter home = internal/persist not diablo (persist/doc.go import direction); (2) seed in level-blob tail, not a column; (3) LevelSource town-only; (4) hero pack re-derived from classdat on load (row Pack is zeros); (5) `startNewGame` = CreatePack for an existing row (no KindSelectHero verb exists); (6) per-hero stats panel not shown (zero pack).

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/port-parity-testing/SKILL.md — the archive is a ported shape (three blocks, seed tail) with fixture/deviation discipline
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split from dcob by the hl1 brainer triage (task-hl1-triage, 2026-11-25; parent-verified). Blocked-by dcob (D1-D10 land first, RF1 included).

## Triage hl2 (2026-11-25, brainer task-hl2-load-triage, parent-verified at ba13766)

Verdict: READY — todo after this re-scope. Scope decomposed: the per-hero single_N.db enumeration migration (old AC #3) is SPLIT OUT to its own draft bean (blocked-by this one) — it is a UX + storage-handle redesign the load path never touches (upstream builds the list by probing 99 archives, pfile.cpp:695-726 @e1ad0c2b6). Adapter home decided: `internal/persist` (NOT the diablo layer the draft named) — persist/doc.go:13-17 pins "nothing in diablo depends on persist"; persist already imports diablo/save+entities, and persist->world is the legal diablo-tier->game edge (no cycle). Six recorded deviations (see AC list item 8), two load-path facts the draft missed: loadSlotHandler stays silent unless B==LevelSeed (session.go:258-278 — the UI must peek the seed), and there is no game-entry transition past SelHero (frontflow.go:11-15 — S8 adds FrontGame). Slices S1-S8 land on ONE branch `feat/m2-hero-load-v13s`, five commits. Full triage: .hermes/lane/m2-hero-load/v13s-triage-result.md (gitignored lane evidence).

## DECIDED

D1 store-model migration is SPLIT to its own bean (parent override of the dcob-D1 "rides this bean" text, with reasons above); this bean lands on the single-db boot, which remains a recorded simplification.

## Summary of Changes

The load path: production SessionArchive in internal/persist (three-block promote of the tested integration shape, seed in the level-blob tail, ErrNoSlot vs wrapped errors, pack re-derived from classdat); the boot registers the save/load verbs through attachSaveLoad (no call-site churn) with a load outcome hook; the hero row offers Load Game / New Game (upstream SelheroListSelect flow, seed peek before submit); the compose band draws the two rows; FrontGame is the first transition out of the front screens — a settled Load or New-Game enters the game.

## Proof of Work

- RED per slice recorded in the round log (S2+S1 compile-shaped `undefined: persist.NewSessionArchive`; S7 assertion-shaped `band rows = 4, want 2`; S6/S8 as contracted).
- Commits: e4fc442, 129bd8a, 62dabc6, 0ad8aa5, 3aac5e0 + review repair 7e36a61.
- Independent review: round 1 APPROVE-WITH-FINDINGS, all 8 ACs MET; findings F1 (S8 tile non-discrimination), F2 (false townLevelSource invariant), F3 (stale comment + armed hook) + nits repaired in 7e36a61 — the load tests now pin the restored hero tile {76,68} against the bootstrap spawn tile.
- Parent gates at 7e36a61: untagged full suite ok; `go vet` rc 0; gofmt clean; lint 0 issues; windowed cmd ok WITH real archives (AC4 included); windowed cmd without assets fails only on the pre-existing AC4 env gate; windowed scene ok without assets; just sqlc-check clean; headless no-display pin green.
- Residual: dungeon-record restore waits on a production dungeon LevelSource (town-only disclosure); per-hero stats panel not shown (rows carry zero packs); the per-hero single_N.db enumeration lives in the blocked follow-up bean (boot: per-hero single_N.db store model).
