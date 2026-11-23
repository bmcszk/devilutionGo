# M1 manual smoke checklist

This page is the human pass over the M1 slice: boot to a window, create a
warrior, walk the town, descend into the cathedral, kill, loot, save, load,
quit. Run every step from the repo root. Each step names the screen state you
should see and the exact command that produces it.

**AC-4 refresh (devilGo-jjd0): the INTERACTIVE demo path is now wired and
gated.** The front screens (title → main menu → selhero) take real keys, a
warrior is created through the selhero UI and raised by the live kernel
(CreatedPack in the tick log), and after the front screens the device lane
reads the arrow keys and walks the hero. The deterministic proof is
`TestAC4InteractiveDemoSmoke` (cmd/devilutionGo/ac4_demo_gate_test.go), which
drives the whole beat sequence headless-deterministically through the boot's
production construction surface and captures a PNG + pixel stats per beat.

Steps 4 to 8 of the original nine-step session (descend, kill, loot, save,
load) are still headless-only: they have no window UI yet. The scripted
section below stays as the CI oracle for those.

## The interactive demo (what a human can do in the window today)

The beat list below is exactly what `TestAC4InteractiveDemoSmoke` drives; run
it with an artifact dir and you get the same captures this page references:

```sh
export DISPLAY=:0
export DG_ASSET_ARCHIVES="$HOME/Games/Heroic/Diablo"
AC4_DEMO_ARTIFACT_DIR=/tmp/ac4-caps go test -tags windowed ./cmd/devilutionGo \
  -count=1 -run TestAC4InteractiveDemoSmoke -v
```

Each beat's capture lives under `docs/screenshots/m1-smoke/` with a `.stats`
sidecar; the pixel profile of every capture is recorded inline.

### Beat 1: boot lands on the title

A key press on the title opens the main menu; Escape on the title quits
(diablo.cpp's quit surface). The boot screen is `FrontTitle`
(`scene.NewFrontFlow`, frontflow.go).

- Capture: `ac4-beat1-title.png` — mean 66.45, min 0, max 255, 91 distinct
colours, 42.02% non-black. The town render sits behind the title band, so the
frame is the lit town diamond with the front band overlaid from y 96.

### Beat 2: any key → main menu

Single Player / Exit, arrow-key navigation, Enter selects (mainmenu.cpp:72-79
order).

- Capture: `ac4-beat2-mainmenu.png` — mean 66.12, min 0, max 255, 92 distinct
colours, 42.02% non-black. The +1 colour and −0.33 mean are the focus bar the
menu draws on the band.

### Beat 3: Enter on Single Player → selhero

The real hero list/create/delete screen (devilGo-5fr4): the list refreshes
from the persist store at entry time, the actions are Create/Delete/Quit
(selhero.cpp:60-75 order).

- Capture: `ac4-beat3-selhero.png` — mean 68.58, min 0, max 255, 92 distinct
colours, 43.27% non-black. The extra non-black row is the selhero action rows
plus the focused-row marker.

### Beat 4: create a warrior (name + class) → CreatedPack

Enter on Create Hero opens the create flow (name first, then the class pick,
selhero.cpp's own order). The name step accepts Enter/Space to confirm — raw
text input is not wired yet (no rune reader in the window layer):
the gate drives the name white-box, and a human can currently
create an empty-name hero (Enter, Enter). Text-input wiring is a
k736 (input epic) follow-up. The class step navigates with arrows and
confirms with Enter. The pick submits `KindCreatePack` through the live
kernel, and the NEXT settled tick's log must carry exactly one
`KindCreatedPack` for the warrior class — the gate asserts the event and the
persisted hero row ("Gate Hero", class 0).

- Capture: `ac4-beat4-created.png` — mean 66.12, min 0, max 255, 92 distinct
colours, 42.02% non-black. After the create closes, the flow walks back to
the main menu, so this frame equals beat 2 plus the store row (invisible —
the row is data, not pixels; the beat's proof is the tick log + store, not
the frame).

Note: the front-screen/menu bars render as highlight bars without
glyphs — font/text rendering is not wired into the front screens
yet (a named deferral; the dg bitmap-text renderer exists but is
not consumed here).

### Beat 5: walk the town with the arrow keys

After the front screens release, the device lane reads the arrows
(`windowMovement`, window_scene.go): a firm arrow push queues a stick
decision, the frame's settled tick dispatches it, and the kernel raises one
`PlayerMoved` per step. Each east step moves the hero (+1, −1) and the live
view recomposes at the moved tile (devilGo-njzu) — the gate asserts the pixel
content of consecutive walk frames differs and the presenter's tile equals
the state's hero tile.

- Captures: `ac4-beat5-walk01/02/03.png` — hero spawn (75, 68), three east
steps to (78, 65).
  - walk01: mean 62.49, min 0, max 255, 126 distinct colours, 60.39% non-black
  - walk02: mean 62.73, min 0, max 255, 123 distinct colours, 60.24% non-black
  - walk03: mean 63.72, min 0, max 255, 122 distinct colours, 60.23% non-black
  - Each frame is pixel-different from the previous one (asserted by the
gate); the means climb because the hero walks south-east into the brighter
town tiles.

#### Baseline comparison (honest delta)

The validated first-town-frame baseline profile is mean 37.46 (all-pixel),
60.37% non-black, 131 distinct colours (`step2-window-town-hero.png`,
byte-identical to `frame0540.png`). The interactive first town frame
(`ac4-beat5-walk01.png`, hero at (76, 67) after one step) measures 37.74 /
60.39% / 127.

- Non-black matches (60.39 vs 60.37%): the same town diamond, same lit area.
- All-pixel mean 37.74 vs 37.46: within one brightness step — the view moved
one tile, so the light profile shifts a fraction.
- 127 vs 131 distinct colours: the four-colour delta is tiles leaving the
viewport at the walked tile, not a palette change — min 0 and max 255 are
identical, and the per-beat profile above shows the count drifting
continuously (126 → 123 → 122) as the view moves.

### Honest gap: how the walk beat is driven

The gate drives beat 5 through the lane seam directly (`lane.PushStick` +
`lane.frame`, the same seam `windowMovement` feeds and the landed
live-view test uses). The window flow itself today never transitions to a
live-town state: `windowGame.Update` gates movement on the flow leaving the
three front screens, and `FrontFlow` has no enter-town state — the flow sits
on `FrontSelHero` until Escape walks back or a create closes it back to the
menu. A human in the real window therefore reaches the selhero screen and
creates a hero, but the arrow keys reach the movement lane only once the
flow gains an explicit enter-town transition. That transition is the demo
flow's own follow-up slice, not this gate's; the gate pins the beats the
landed surfaces already serve and the doc says exactly which seam each beat
runs through.

## What the window shows today (pre-refresh status, kept for the record)

The windowed binary boots to the title (any key → menu → selhero), creates a
hero through the selhero UI and ticks at 20 TPS; ESC on the title or Exit on
the menu quits cleanly. Movement keys reach the lane only through the gate's
lane seam (the honest gap above). See
[Windowed boot wiring](#windowed-boot-wiring) for the code paths.

The nine-step session itself is real and covered end to end: it runs headless
through the landed kernel verbs in `internal/integration`, and the AC-1, AC-2
and AC-3 gates pin it. So the checklist splits each step into what a human
can do in the window and what the same step proves headless.

| Step | Window (human) | Headless proof |
|------|----------------|----------------|
| 1. Boot to window | yes (title screen) | `TestAC4InteractiveDemoSmoke` beat 1 |
| 2. Create warrior | yes (selhero UI) | `TestAC4InteractiveDemoSmoke` beat 4 + `TestM1ScriptedSessionCreateIsAKernelVerb` |
| 3. Walk town | arrow keys via the lane seam (window flow transition pending) | `TestAC4InteractiveDemoSmoke` beat 5 + `TestM1ScriptedSessionNineStages` |
| 4. Descend to cathedral | no input path | `TestM1ScriptedSessionDungeonRunsTheCommittedLayout` |
| 5. Kill | no input path | `TestM1ScriptedSessionKillsAreKernelDriven` |
| 6. Loot | no input path | `TestM1ScriptedSessionDescendAndLootAreKernelVerbs` |
| 7. Save | no input path | `TestM1ScriptedSessionSaveAndLoadAreKernelVerbs` |
| 8. Load | no input path | `TestM1ScriptedSessionRoundtripRestoresTheWalk` |
| 9. Quit | yes | `TestWindowBootQuitKeys` + the front flow's quit paths |

The dungeon, kill, loot and save/load input paths are tracked in the
milestone's demo-flow beans; until they land, steps 4 to 8 are headless-only.

## Prerequisites

1. Go 1.27 or newer, and `just` ([casey/just](https://github.com/casey/just)).
2. An X display. The windowed build needs one; this box runs X on `:0`
   under gamescope. Export it once per shell:

   ```sh
   export DISPLAY=:0
   ```

   Without `DISPLAY` the binary exits 64 with
   `windowed boot: no display (DISPLAY unset); headless mode is -headless`.
   That named error is the expected guard, not a crash.
3. The retail archives, named by `DG_ASSET_ARCHIVES` as a colon-separated
   path list:

   ```sh
   export DG_ASSET_ARCHIVES="$HOME/Games/Heroic/Diablo"
   ```

   `DIABDAT.MPQ` must be in that directory. A `hellfire` directory is optional;
   when it is absent the loader logs one `WARN mpq archive failed to open` and
   carries on with the retail town data. With `DG_ASSET_ARCHIVES` unset the
   window still opens and ticks, logs
   `ERROR windowed boot: compose first scene err="DG_ASSET_ARCHIVES not set"`,
   and shows a black frame.

## Step 1: boot to the window

Build the game and start it.

```sh
just build
go build -tags windowed -o bin/devilutionGo-windowed ./cmd/devilutionGo
DISPLAY=:0 DG_ASSET_ARCHIVES="$DG_ASSET_ARCHIVES" ./bin/devilutionGo-windowed
```

Expected screen state: a 640x480 window titled `devilutionGo` showing the
Tristram town ground with the warrior standing at the centre, about 60 percent
of the frame non-black and the lower half brighter than the upper half.

Expected log: one `INFO devilutionGo started mode=windowed` line, then one
`INFO windowed boot ticking frames=N` line per second as the frame counter
climbs.

Two build notes, both measured:

- `just build` produces the untagged binary. Running it prints
  `windowed mode requires a build with -tags windowed (this binary links no
  display stack)` and exits 64. That is the ADR-020 two-binary law, not a
  failure: the default binary carries no display stack so the headless
  dependency graph stays clean.
- The `-tags windowed` build is the one that opens a window.

## Step 2: create a warrior

Window: reachable — the selhero UI (beat 4 above). The capture and the pixel
profile live there; the proof that the create is kernel-driven (not a
test-side write) is the same beat's tick-log assertion.

Headless proof: run the create stage.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionCreateIsAKernelVerb -v
```

Expected: `--- PASS`. The pin asserts exactly one `KindCreatePack` command and
one `KindCreatedPack` event whose payload carries the warrior class and the
spawn tile.

## Step 3: walk the town

Window: reachable through the lane seam (beat 5 above); the window flow's
explicit enter-town transition is pending (the honest gap). The window layer's
raw-key read (`windowMovement`) already feeds the lane, so the moment the flow
releases, arrows walk.

Headless proof: the walk's town-enter stage.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionNineStages -v
```

Expected: `--- PASS`. The pin asserts the walk froze exactly the nine stages,
in order, starting with `boot`, `create`, `town-enter`.

To watch the injected-click walk on its own, run the smoke story and look at
the two PNG artifacts it writes:

```sh
SMOKE_ARTIFACT_DIR=/tmp/m1-smoke go test ./internal/game \
  -run TestSmokeStoryFrames -count=1 -v
```

Expected: the test prints `smoke frame: /tmp/m1-smoke/frame-1-before.png` and
the same for `frame-2-after.png`. The before frame shows the hero marker at
the start tile; the after frame shows it one tile right with the clicked cell
tinted by the target shade. These are the smoke harness's own schematic grid,
not the retail town render. Run `just smoke` instead and the before file is
overwritten by `TestSmokeFramePNGArtifact`, which clicks first; see
[Screenshots](#screenshots) for that trap.

## Step 4: descend into the cathedral

Window: not reachable (steps 4 to 8 stay headless-only; see the split at the
top of this page).

Headless proof: the walk loads the committed cathedral L1 layout at its own
seed and freezes it.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionDungeonRunsTheCommittedLayout -v
```

Expected: `--- PASS`. The pin asserts the dungeon stage runs the committed
layout fixture, not a stand-in grid.

## Step 5: kill

Window: not reachable (headless-only).

Headless proof: ten kills, all kernel-driven.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionKillsAreKernelDriven -v
```

Expected: `--- PASS`. The pin asserts the kill stages advance the frozen state
and carry kernel commands, so a kill cannot be a test-side no-op.

## Step 6: loot

Window: not reachable (headless-only).

Headless proof: descend and loot both run as kernel verbs.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionDescendAndLootAreKernelVerbs -v
```

Expected: `--- PASS`, with both transitions appearing exactly once.

## Step 7: save

Window: not reachable (headless-only).

Headless proof: the save stage freezes the session through the kernel.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionSaveAndLoadAreKernelVerbs -v
```

Expected: `--- PASS`. The pin asserts one `KindSaveSlot` command and one
`KindSavedSlot` event, with payloads checked against the walk's slot and level.

## Step 8: load

Window: not reachable (headless-only).

Headless proof: the roundtrip restores exactly what the save froze.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionRoundtripRestoresTheWalk -v
```

Expected: `--- PASS`. The pin asserts the loaded level identity, hero tile,
warrior pack and live-monster count equal the snapshot the save took.

## Step 9: quit

Window: reachable, and this is the one interactive step to run by hand.

With the window focused, press `ESC` on the title (or select Exit on the main
menu).

Expected log: a final `INFO windowed boot exit frames=N` line, then the
process exits 0. Closing the window with the title-bar button exits the same
way.

If the keys do nothing, the window does not have focus. The X input has to
land on the game window: on a desktop with several windows, click the game
window once first. Measured on this box under gamescope, an unfocused
synthetic key press is dropped and the window keeps ticking.

## Determinism check (optional)

The walk is frozen at every stage, so a repeat run must reproduce the same nine
hashes. Run it twice and compare.

```sh
mkdir -p /tmp/m1-walk
M1_LANE_DIR=/tmp/m1-walk go test ./internal/integration -count=1 -run TestM1ScriptedSessionRunReplay -v
cat /tmp/m1-walk/statehash-run.json
```

Expected first stage hashes: `boot` `4e645f013c56620d`, `create`
`4f9b29731d9af0d5`, `town-enter` `3bb0f5f6ee811ecf`. A second run writes the
same values.

## Screenshots

The captures below were taken on this box at the head this page was written on.
Every file lives under `docs/screenshots/m1-smoke/`, and every one carries its
pixel stats and a root-cause note for anything that looks off.

### `step2-window-town-hero.png`

Taken from the live window, so this is the visual oracle path: the binary ran
under `DISPLAY=:0` with `--debug --screenshot-dir`, and the capture is frame
60, the first cadence frame.

- Size 640x480, mean 37.46, min 0, max 255, 131 distinct colours, 60.37 percent
  non-black.
- Corners all `(0,0,0)`: the town view is a diamond inside a black field, so
  the four corners are outside the drawn region.
- Top-half mean 26.57 against bottom-half mean 48.34. That gradient is the town
  light table, not fog: the upper tiles are the darker northern ground.
- The image is byte-identical to `docs/screenshots/frame0540.png`, the capture
  committed with the first-scene epic, and it reproduced byte-identically on a
  second boot (`md5 e7af5dc7122072e389a749096e240632`).
- Visible content: Tristram ground tiles with the warrior standing at the
  centre of the frame.

### `step2-town-without-hero.png`

The same compose with the hero draw skipped, kept as the control that proves
the hero pixels are real and not background.

- Size 640x480, mean 18.42, min 0, max 255, 100 distinct colours, 60.37 percent
  non-black.
- Against `step2-window-town-hero.png` the two frames differ in 185,456 pixels,
  bounding box x 0-639, y 82-464. The whole lit region shifts, because the
  hero compose also applies the boot brightness tone map
  (`sceneBrightness = 60`, `internal/present/scene/warrior.go`).
- Centre 96x128 rect mean 23.55 against the hero frame's 50.36: the warrior
  occupies that rect and lifts its mean.

### `step4-smoke-walk-before.png` and `step4-smoke-walk-after.png`

The smoke story's schematic walk frames, from `just smoke`.

- Both 1280x960 (20x15 tiles at 64 px, which is the smoke harness's own
  camera), mean 210.79 and 210.80, min 72, max 255, 3 and 4 distinct colours,
  100 percent non-black.
- These are deliberately flat: the smoke artifact expands five fixed palette
  roles (floor, hero, target, blocked, grid line) through the committed `l1_1`
  palette, so 3 to 4 colours is the design, not a render failure.
- In the before frame the hero shade covers a 63x63 cell at x 0-62, y 0-62
  (the hero starts on tile 0,0) and the target shade is absent. In the after
  frame the hero shade sits one tile right at x 64-126, y 0-62 with 125 px
  left visible, and the target shade covers 3,969 px at x 65-127, y 1-63.
- Root cause of that 125 px remainder, read off the code rather than guessed:
  the hero marker fills pixels 0 to 62 of its cell (`tileShade`), while the
  target stamp fills pixels 1 to 63 (`stampTarget`), and the camera follows
  the hero so the two share a cell. The target therefore covers the hero cell
  except its top row and left column: 63 + 63 - 1 shared corner = 125. The
  click moved the hero marker one tile, and the visible remainder is the
  one-pixel seam between the two fills, not a leftover sprite.
- Capture trap, measured: `TestSmokeFramePNGArtifact` and `TestSmokeStoryFrames`
  both write `frame-1-before.png` into `SMOKE_ARTIFACT_DIR`, and the first of
  them clicks before it composes. Run them in one process, as `just smoke`
  does, and the "before" file is the overwritten one with the hero already
  moved. The captures above come from the story test alone. To reproduce them,
  run the single test rather than the recipe:

  ```sh
  SMOKE_ARTIFACT_DIR=/tmp/m1-smoke go test ./internal/game \
    -run TestSmokeStoryFrames -count=1 -v
  ```

- Root-cause note for the 100 percent non-black figure: the harness fills every
  tile, walkable or not, so there is no background to be black. The retail
  window capture above is the frame that shows the black diamond field.

## Windowed boot wiring

The reason steps 4 to 8 are headless-only is visible in the boot file, so
check it rather than trust this page:

```sh
grep -n 'Kernel\|IsKeyPressed\|CursorPosition' cmd/devilutionGo/*.go
```

Expected today: the kernel IS wired in the windowed boot
(`bootDeviceLane`, `device_lane.go`) — the device lane, the bridge session
and the movement verbs are live, and the selhero create submits
`KindCreatePack` through them (devilGo-5fr4). What is still missing is a
MOUSE read and the front flow's enter-town release: the arrow keys reach the
lane through `windowMovement` only once the flow leaves the front screens,
and the flow has no enter-town state (the honest gap above). The scripted
nine-stage walk is pinned in `internal/integration`, which is why steps 4 to
8 stay headless-only.

## Bean

This checklist is the documentation half of
[M1 gate: AC-4 manual smoke — checklist doc + screenshots](../../../.beans/devilGo-0qfg--m1-gate-ac-4-manual-smoke-checklist-doc-screenshot.md).
The screenshots are attached as evidence to the parent milestone,
[Milestone M1: Playable Spawn demo (town + cathedral slice)](../../../.beans/devilGo-ftbu--milestone-m1-playable-spawn-demo-town-cathedral-sl.md).
