---
# devilGo-1a4r
title: 'Game menu: gamemenu.cpp port (pause, options routing, tick-rate slider)'
status: completed
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-26T21:22:51Z
parent: devilGo-xih1
---

---
# devilGo-1a4r
title: 'Game menu: gamemenu.cpp port (pause, options routing, tick-rate slider)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-25T22:53:27Z
parent: devilGo-xih1
---

---
# devilGo-1a4r
title: 'Game menu: gamemenu.cpp port (pause, options routing, tick-rate slider)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-10-29T16:14:42Z
parent: devilGo-xih1
---

## Goal

Port `gamemenu.cpp` (391 lines): in-game pause menu — Esc menu, options routing
(tick-rate slider 20-50 → gnTickDelay, per diablo.cpp:146), new-game/quit flows,
`gamemenu_previous`/`gamemenu_enter` state machine.

## Research

Source/gamemenu.cpp:1-391 (menu tables GameMenuNewGame/GameMenuSave/GameMenuOptions...);
Source/gamemenu.cpp:277 tick-rate; interfac.cpp screen transitions; DiabloUI/SettingsMenu
for the options submenus; menus.cpp routing.

## Scope

internal/ui/gamemenu.go (screen-stack entry per xih1 widget kit) + routing hooks in
internal/game. Options persistence lands later via yjuv SimOptions — this bean only wires
the menu tree + actions that exist (save/load once oqho/xwtt land: gate those two entries
on their presence, stub with "not available" like spawn did).

## Acceptance criteria

- [x] **AC1** (S1) `game/ui` carries the menu-table model: single-player rows `Options, Save Game, Load Game, Exit to Main Menu, Quit Game` in upstream order, disabled rows skipped by Next/Prev with wrap; slider set/get/steps port the `X:gmenu.cpp:375-392` arithmetic — round-trips 20/35/50 -> steps 0/23/46 -> 20/35/50, Right from step 0 -> 21, saturating at both ends. RED: `go test ./internal/game/ui -run 'TestMenuTable|TestMenuSlider' -count=1`
- [x] **AC2** (S2) `Game.ProcessInput` ports `X:diablo.cpp:229-232`: with the menu open in single-player `Update` runs zero ticks (`TickCount` unchanged); in multiplayer it runs them; a menu-open game with no menu set behaves as today (20 ticks/s). RED: `go test ./internal/game -run 'TestGameLogicSkipsWhileMenuOpen|TestGameLogicRunsWhileMenuOpenInMultiplayer' -count=1` (behavioural: fails on the current always-true stub)
- [x] **AC3** (S2) `KindSetTickRate`/`KindTickRateSet` exist in `kernel`, the gate drops 19 and 51, the handler changes the live tick delay only on a real change and emits exactly one event; `Update(50ms)` then yields 2 ticks at rate 50. RED: `go test ./internal/game ./internal/game/world -run 'TestSetTickRate|TestGate.*TickRate' -count=1`
- [x] **AC4** (S3) `FrontFlow` routes FrontGame keys to an installed game screen; with none, Esc on FrontGame stays inert and never quits; `LeaveGame` lands on FrontMainMenu; `RequestQuit` sets the quit the window already honours. RED: `go test ./internal/present/scene -run 'TestFrontGame' -count=1` (behavioural RED for the Esc leg)
- [x] **AC5** (S4, **W**) Esc on FrontGame opens the menu focused on the first ENABLED row; while open the REAL `Update` advances `TickCount` by zero, a queued stick push does not move the hero, and the frame still runs so a pending save settles. RED: `go test -tags windowed ./cmd/devilutionGo -run 'TestGameMenu_PausesSim' -count=1`
- [x] **AC6** (S4, **W**) Load Game is disabled when the slot has no game (skipped by arrows, Enter inert) and enabled otherwise. RED: `... -run 'TestGameMenu_LoadRowEnabled'`
- [x] **AC7** (S4, **W**) Save Game submits `KindSaveSlot{A: hero slot}` and the archive holds the session under THAT slot after the settled pass; Load Game submits `KindLoadSlot` and the world restores. RED: `... -run 'TestGameMenu_Save|TestGameMenu_Load'`
- [x] **AC8** (S4, **W**) Exit to Main Menu returns the flow to FrontMainMenu and stops the movement lane; Quit Game makes `Update` return `ebiten.Termination`; Esc on the open menu closes it and ticks resume (the round-trip). RED: `... -run 'TestGameMenu_ExitToMainMenu|TestGameMenu_Quit|TestGameMenu_EscRoundTrip'`
- [x] **AC9** (S5, **W**) The Options row opens the entry-for-entry submenu `Music, Sound, Gamma, Speed, Previous Menu` with the first three DISABLED (follow-up bean); Previous Menu returns to the root; Esc in the submenu closes the whole menu (upstream, not root). RED: `... -run 'TestGameMenuOptions'`
- [x] **AC10** (S5, **W**) Speed row: Right at 20 submits `KindSetTickRate{A:21}` and `TickDelay()` becomes `1000/21 ms`; Enter toggles 20<->50; the boot's `SimOptions.TickRate` follows (in-memory; INI persistence is a recorded deviation owned by the runtime-ini bean). With `Multiplayer` set the row is disabled. RED: `... -run 'TestGameMenuSpeed'`
- [x] **AC11** (S5, **W+A**) Render: menu open draws one label bar per row in the band (5 root / 5 submenu) and a slider fill that grows by `fill/46` per step; menu-closed / root / submenu captures have DISTINCT `.stats`. RED: `go test -tags windowed ./internal/present/scene -run 'TestGameMenuRender' -count=1`
- [x] **AC12** (all) Deviation ledger in the bean's PoW: no spawn/hellfire gate exists in gamemenu.cpp (bean clause dropped); dead/cursor enabled-flags are constant-true stand-ins until death/cursor land; `PauseMode` arm deferred; INI persistence deferred to the runtime-ini bean; three options rows disabled.
- [x] **AC13** (all) Upstream parity oracle: `test/game_menu_test.cpp` (`X:test/game_menu_test.cpp:63-157`, 8 open/close/toggle tests; `:164-245`, 6 slider-math tests) is state-only with NO rendering assertions — the Go tests cover the same assertions by name; the PoW maps each upstream TEST_F to a Go test (14 rows).

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ebiten/input/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md — table-driven sub-tests
- ~/.agents/skills/port-parity-testing/SKILL.md — the slider math is a 1:1 port with a C++ oracle
- /home/blaze/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — worktree/branch/idle-frame law/visual PoW

## Dependencies & blockers

blocked_by devilGo-vf8q (widget kit + screen stack) — menus render through it.
Can start the menu-TREE data + routing layer before vf8q lands (pure logic + tests),
UI wiring after.

## Libraries

Binds ebiten v2 for rendering (owner flag in Scope); everything else stdlib.
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

## Triage hl4 DELTA (2026-11-25, brainer task-hl4-gm-triage, parent-verified at 47662de)

Verdict: READY after re-scope — 6 slices (S1-S5 + S6 follow-up bean) on ONE branch `feat/1a4r-game-menu`. Honest corrections vs the Oct draft: (1) the in-game Options row routes to the IN-ENGINE slider table (gamemenu.cpp:62-69), NOT DiabloUI SettingsMenu (that has exactly one caller: the MAIN menu, menu.cpp:201); (2) 20-50 is nTickRate (ticks/s), gnTickDelay = 1000/rate (50ms..20ms, default 50); (3) gamemenu_previous/enter do not exist — the state machine is gamemenu_handle_previous (gmenu_is_active ? off : on); (4) engine.SimOptions with TickRate/GnTickDelay/ClampTickRate ALREADY EXISTS (config.go:22-78) with zero consumers; (5) kernel KindQuit/KindGameQuit exist with no production handler — reused, no new Kind for quit; (6) upstream has TWO pause mechanisms and the menu uses the weaker gmenu_active one (PauseMode deferred, recorded). blocked_by vf8q: SATISFIED for S1-S4 (MenuList/ScreenStack on tree), still REAL for S5's drawn slider — edge kept, do not wait. Save/Load entries are REAL actions now (oqho/xwtt landed — the stub clause is void). Options INI persistence stays with the runtime-ini owner (the boot opens no INI today — recorded deviation). Upstream parity oracle: test/game_menu_test.cpp (14 TEST_Fs) mapped by name in AC13. Full triage: .hermes/lane/m2-hero-load/gm-triage-result.md (gitignored).

## Summary of Changes

The gamemenu.cpp port as 6 slices: the menu table as data + gmenu slider math (game/ui, no callbacks); the event-sourced pause rule + tick rate (ProcessInput ports diablo.cpp:229-232; KindSetTickRate/KindTickRateSet with a 20-50 gate rule; the handler changes the live delay only on a real change and the replay carries the rate); FrontGame key routing + LeaveGame/RequestQuit; the in-game Esc menu with REAL actions (Save Game submits KindSaveSlot, Load Game peeks the seed and submits KindLoadSlot, Exit/Quit route the flow); the entry-for-entry Options submenu with Speed live and Music/Sound/Gamma disabled (their follow-up bean: In-game Options: Music / Sound / Gamma sliders (gamemenu.cpp:193-262)).

## Proof of Work

- RED per slice recorded; the behavioural REDs the contract demanded are real (menu-open zero-tick through the REAL Update loop; FrontGame Esc reaching the stub; Speed applying while open).
- Commits: bf9588f, 754e94c, 29c03a4, 06aa401, 08bb7ab, 99a6c10 + review repairs 5b9262f, bc35107/0fe2793, 38fe948, 1fa553e.
- Independent review round 1: REQUEST-CHANGES — F1 the pause rule starved the menu's own commands (Save/Load/Speed queued but GameLogic bailed before the kernel pass — a circular wait; upstream calls gamemenu_off first); F2 the save proof was vacuous (menu closed during the test, slot pre-seeded); F3 slot/seed never threaded (every save wrote slot 0, seed 0); F4 the AC13 map missing. All four repaired; round 2 confirmed the four fixed and pinned, but caught two regressions the repair introduced — F10 UpdateWhilePaused double-advanced lastTick (paused pass fired every other frame) and F11 the window frame paused the WORLD in multiplayer (reversing AC2's MP leg). Round-2 repair 38fe948/1fa553e: single advance per due tick (tests: N frames -> N ticks, resume runs exactly one pass) and the ProcessInput decision kept at the window layer (MP frame vs framePaused).
- Upstream parity oracle (AC13): test/game_menu_test.cpp's 14 TEST_Fs map by name — GameMenuOpenAndClose twins (DoubleOpenStaysOpen, MenuStateConsistency, IdleFrameNeverChangesAnything, EscRoundTrip, ExitToMainMenu, Quit), slider twins (Speed_RightStepsTheRate/EnterToggles/AppliesWhileTheMenuIsOpen/LabelFollowsTheLiveValue/MultiplayerDisablesTheRow + menu_table_test steps/granularity cases); row 8's sgpCurrentMenu==nullptr leg is an owned deviation (recorded in the test comment).
- Parent gates at 1fa553e: gofmt empty; both builds rc 0; go vet rc 0 with and without the tag; scripts/lint.sh 0 issues; untagged full suite ok; windowed cmd+scene ok without assets (AC4 env-gate excepted); with assets cmd ok incl. AC4 (all 20 menu tests green, 109 in the package); sqlc-check clean; movement-gate regression net (TestWindowUpdateLoadEntersTheGame + AC4) green.
- Residual: PauseMode focus-loss arm deferred; INI persistence deferred to the runtime-ini owner (SimOptions in-memory only); dead/cursor enabled-flags constant-true stand-ins; Music/Sound/Gamma rows disabled (owned bean); the sgpCurrentMenu nullptr leg deviation recorded.



## Progress log - 2026-11-26 (Options-slice triage)
- INI persistence deferred to "the runtime-ini bean": that bean never existed; the Options-slice wave (feat/options-slice, triage .hermes/lane/options-slice/triage.md E3) is its home - window_ebiten.go's in-memory simOpts deviation closes with that PR.
- Defect found during the triage: GameMenuSlider divides by speedSliderSteps=46 for EVERY slider row (gamemenu.go:156-164) - wrong for 2 of 4 sliders once three rows go live. Unreachable today (one slider row), fixed by the Options-slice wave per the upstream per-row steps (64 volumes, 21 gamma); triaged in the same file (E5), no separate bug bean.
