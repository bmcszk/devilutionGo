# M1 manual smoke checklist

This page is the human pass over the M1 slice: boot to a window, create a
warrior, walk the town, descend into the cathedral, kill, loot, save, load,
quit. Run every step from the repo root. Each step names the screen state you
should see and the exact command that produces it.

Read the honest status first, because two of the nine steps are not reachable
from a window yet.

## What the window shows today

The windowed binary boots, presents the retail Tristram frame with the warrior
standing in it, ticks at 20 TPS and quits cleanly on ESC or `q`. That is the
whole interactive surface. The windowed boot wires no command kernel and reads
no mouse or movement keys, so nothing you press moves the hero. See
[Windowed boot wiring](#windowed-boot-wiring) for the code paths.

The nine-step session itself is real and covered end to end: it runs headless
through the landed kernel verbs in `internal/integration`, and the AC-1, AC-2
and AC-3 gates pin it. So this checklist splits each step into what a human can
do in the window and what the same step proves headless.

| Step | Window (human) | Headless proof |
|------|----------------|----------------|
| 1. Boot to window | yes | `TestM1ScriptedSessionNineStages` covers the boot stage |
| 2. Create warrior | no input path | `TestM1ScriptedSessionCreateIsAKernelVerb` |
| 3. Walk town | no input path | `TestM1ScriptedSessionNineStages` town-enter stage |
| 4. Descend to cathedral | no input path | `TestM1ScriptedSessionDungeonRunsTheCommittedLayout` |
| 5. Kill | no input path | `TestM1ScriptedSessionKillsAreKernelDriven` |
| 6. Loot | no input path | `TestM1ScriptedSessionDescendAndLootAreKernelVerbs` |
| 7. Save | no input path | `TestM1ScriptedSessionSaveAndLoadAreKernelVerbs` |
| 8. Load | no input path | `TestM1ScriptedSessionRoundtripRestoresTheWalk` |
| 9. Quit | yes | `TestWindowBootQuitKeys` |

The input path and the hero-creation screens are tracked as
[Input: keyboard/mouse/gamepad incl. fork stick turn/walk](../../../.beans/devilGo-k736--input-keyboardmousegamepad-incl-fork-stick-turnwal.md)
and
[Hero select/create/delete + class stats](../../../.beans/devilGo-dcob--hero-selectcreatedelete-class-stats.md).
Until those land, steps 2 to 8 are headless-only.

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

Window: not reachable. There is no character-creation screen and no input path
into the kernel, so the warrior you see is the composed first scene, not a
created hero.

Headless proof: run the create stage.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionCreateIsAKernelVerb -v
```

Expected: `--- PASS`. The pin asserts exactly one `KindCreatePack` command and
one `KindCreatedPack` event whose payload carries the warrior class and the
spawn tile.

## Step 3: walk the town

Window: not reachable. The boot reads only ESC and `q`
(`cmd/devilutionGo/window_scene.go`, `windowKey`); mouse clicks and movement
keys are not read at all.

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

Window: not reachable, same missing input path.

Headless proof: the walk loads the committed cathedral L1 layout at its own
seed and freezes it.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionDungeonRunsTheCommittedLayout -v
```

Expected: `--- PASS`. The pin asserts the dungeon stage runs the committed
layout fixture, not a stand-in grid.

## Step 5: kill

Window: not reachable.

Headless proof: ten kills, all kernel-driven.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionKillsAreKernelDriven -v
```

Expected: `--- PASS`. The pin asserts the kill stages advance the frozen state
and carry kernel commands, so a kill cannot be a test-side no-op.

## Step 6: loot

Window: not reachable.

Headless proof: descend and loot both run as kernel verbs.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionDescendAndLootAreKernelVerbs -v
```

Expected: `--- PASS`, with both transitions appearing exactly once.

## Step 7: save

Window: not reachable.

Headless proof: the save stage freezes the session through the kernel.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionSaveAndLoadAreKernelVerbs -v
```

Expected: `--- PASS`. The pin asserts one `KindSaveSlot` command and one
`KindSavedSlot` event, with payloads checked against the walk's slot and level.

## Step 8: load

Window: not reachable.

Headless proof: the roundtrip restores exactly what the save froze.

```sh
go test ./internal/integration -count=1 -run TestM1ScriptedSessionRoundtripRestoresTheWalk -v
```

Expected: `--- PASS`. The pin asserts the loaded level identity, hero tile,
warrior pack and live-monster count equal the snapshot the save took.

## Step 9: quit

Window: reachable, and this is the one interactive step to run by hand.

With the window focused, press `ESC` or `q`.

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

The reason steps 2 to 8 are headless-only is visible in the boot file, so check
it rather than trust this page:

```sh
grep -n 'Kernel\|IsKeyPressed\|CursorPosition' cmd/devilutionGo/*.go
```

Expected today: only the two `IsKeyPressed` quit-key reads in
`window_scene.go`. There is no `Kernel` reference and no mouse read in the
windowed boot, so no command reaches the world slice and nothing moves the
hero. The kernel and its verbs are wired in the harness
(`internal/game/smoke.go`, `internal/integration/session_stage_test.go`), which
is why the headless walk proves all nine stages while the window proves two.

## Bean

This checklist is the documentation half of
[M1 gate: AC-4 manual smoke — checklist doc + screenshots](../../../.beans/devilGo-0qfg--m1-gate-ac-4-manual-smoke-checklist-doc-screenshot.md).
The screenshots are attached as evidence to the parent milestone,
[Milestone M1: Playable Spawn demo (town + cathedral slice)](../../../.beans/devilGo-ftbu--milestone-m1-playable-spawn-demo-town-cathedral-sl.md).
