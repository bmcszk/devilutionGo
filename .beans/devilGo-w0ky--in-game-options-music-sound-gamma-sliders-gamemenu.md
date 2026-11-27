---
# devilGo-w0ky
title: 'In-game Options: Music / Sound / Gamma sliders (gamemenu.cpp:193-262)'
status: completed
type: task
priority: normal
created_at: 2026-11-25T22:53:27Z
updated_at: 2026-11-26T22:15:32Z
parent: devilGo-xih1
---

# In-game Options: Music / Sound / Gamma sliders (gamemenu.cpp:193-262)

## Goal
The three DISABLED rows the game-menu bean ships become real: Left/Right/Enter dispatch on the FOCUSED row. Music/Sound (64 steps over [-1600,0]) through the audio engine's volume seam with off-at-MIN / restart-on-rise rules; Gamma (21 steps over [0,100]) through brightness read/write-apply, ALWAYS enabled (no audio dependency). Rows enabled iff the audio engine is initialised (upstream !gbSndInited). Full evidence: .hermes/lane/options-slice/triage.md (E5, E-G).

## Research
- The rows: newGameMenuOptionsTable Music/Sound/Gamma Enabled:false (gamemenu.go:170-180). Key handling hard-keyed to Speed (HandleKey :268 -> speedStepRight/Left :295-323; Enter only Speed/Previous :361-363) - enabling is NOT a flag flip.
- Latent defect fixed this wave: GameMenuSlider divides by speedSliderSteps=46 for EVERY row (gamemenu.go:156-164); upstream uses 64 for volumes (X:gamemenu.cpp:124) and 21 for gamma (:154). Fix: read the focused row's own step count (MenuTable is row-generic, game/ui/menu_table.go:107-152).
- Render oracle: composeGameMenuLabels draws the fill only on the focused row (frontflow_render.go:85-123, windowed).
- Gamma seam: ApplyBrightness/WithBrightness ported+tested (render/lightvariants.go:51-78); compose takes brightness as a parameter (warrior.go:138,164); the public ComposeTownFrameAtHero hardcodes 60 (boot switches to the ini value - disclosed visible change, before/after captures); presenter needs a brightness invalidation hook (window_scene.go:89-96 recomposes only on hero change).
- Upstream: UpdateBrightness(v) v>=0 set+apply, always returns current, v==-1 read sentinel (palette.cpp:253-261); rows disabled when !gbSndInited (gamemenu.cpp:131-134); Enter on gamma toggles 0<->100 (gamemenu.cpp:250-262). Brightness is display state, NOT a kernel command (hl4 triage ruling).
- Windowless front-screen law: every step machine pinned with explicit ui.KeyUnknown between arm and answer through the REAL Update (a branch above the guard answers itself in one frame).

## Acceptance criteria
- [x] AC4 (options rows): with the engine initialised Music (64 steps), Sound (64) and Gamma (21) rows are ENABLED and Left/Right/Enter act on the FOCUSED row: Gamma changes the composed frame's MeanBrightness monotonically (0 < 50 < 100 at a fixed hero tile, stats DIFFER) and persists across reboot; Music/Sound change engine volumes and follow off-at-MIN / restart-on-rise; engine NOT initialised -> Music/Sound stay DISABLED, Gamma still works; fill bar uses each row's own step count. Every step machine pinned with ui.KeyUnknown through the real Update. RED: windowed TestGameMenu(Gamma|Music|Sound) fails on today's DISABLED rows; visual PoW = three gamma captures with distinct .stats. RED: go test -tags windowed ./cmd/devilutionGo -run 'TestGameMenu(Gamma|Music|Sound)' -count=1.

## Libraries
stdlib + ebiten (windowed seams only).

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/port-parity-testing/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Combined triage 2026-11-26 (task-opts-triage-brainer): evidence E5 + E-G, the gamma independence AC, and the replace-not-weaken rule for the DISABLED pin (gamemenu_test.go:447-468) in .hermes/lane/options-slice/triage.md.



## Proof of Work (Options-slice wave, parent-verified 2026-11-26)

- AC4 landed (S4b + S4a): rows dispatch on the FOCUSED row; Gamma 21 steps [0,100] via brightness read/write-apply, Music/Sound 64 steps [-1600,0] with off-at-MIN / restart-on-rise; rows enabled iff engine initialised; GameMenuSlider per-row steps defect fixed; KeyUnknown pins in every new step machine.
- Parent re-ran: windowed menu battery (Gamma|Music|Sound|Options|Speed_SliderFill|ClosePersists) ok asset-less; presenter recompose test ok; TestComposeTownFrameAtBrightnessIsMonotone PASS with real archives (the render oracle: mean brightness monotone in level).
- Disclosed visible change: boot brightness now comes from Graphics/Brightness Correction (default 0) instead of the first-scene hardcoded 60 - the town renders darker by default; recorded for the PR body.
- The DISABLED pin was REPLACED keeping the not-initialised disabled leg (upstream !gbSndInited); no assertion weakened.
- Review APPROVE (static); status ticked from parent re-run evidence.
