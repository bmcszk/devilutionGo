//go:build windowed

package main

import (
	"context"
	"fmt"
	"time"

	snd "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// clearOutcomes disarms every flow: the Exit-to-Main-Menu row resets the
// lane's pending hooks before the flow walks back.
func (l *deviceLane) clearOutcomes() {
	for i := range l.outcomes {
		l.outcomes[i].pending, l.outcomes[i].hook = false, nil
	}
}

// gamemenu row indices inside the root table, upstream order
// (X:gamemenu.cpp:49-57 sgSingleMenu).
const (
	gamemenuRowOptions = iota
	gamemenuRowSave
	gamemenuRowLoad
	gamemenuRowExit
	gamemenuRowQuit
)

// gamemenu submenu row indices (X:gamemenu.cpp:68-75 sgOptionsMenu).
const (
	gamemenuRowMusic = iota
	gamemenuRowSound
	gamemenuRowGamma
	gamemenuRowSpeed
	gamemenuRowPrevious
)

// speedSliderSteps is the Speed row's slider granularity
// (X:gamemenu.cpp:175 gmenu_slider_steps(row, 46)).
const speedSliderSteps = 46

// speedSliderMin/Max are nTickRate's bounds (X:gamemenu.cpp:175
// gmenu_slider_set(row, 20, 50, nTickRate)).
const (
	speedSliderMin = 20
	speedSliderMax = 50
)

// gammaSliderSteps is the Gamma row's slider granularity
// (gmenu_slider_steps(&sgOptionsMenu[2], 21), X:gamemenu.cpp:154).
const gammaSliderSteps = 21

// The volume rows' toggle labels: upstream swaps the label with gbSndInited
// (MusicToggleNames / SoundToggleNames, X:gamemenu.cpp:80-90).
const (
	musicLabelOn  = "Music"
	musicLabelOff = "Music Disabled"
	soundLabelOn  = "Sound"
	soundLabelOff = "Sound Disabled"
)

// brightnessKnob is the Gamma row's read/write seam: upstream's
// UpdateBrightness(v) applies v >= 0 and always returns the current level, with
// -1 as the READ sentinel (Source/engine/palette.cpp:253-261).
type brightnessKnob struct {
	read  func() int
	write func(int)
}

// gameMenuScreen is the in-game Esc menu (X:gamemenu.cpp, devilGo-1a4r S4/S5):
// cmd composition over the lane (submit site + tick), the persist archive
// (HasGame for the Load row) and the front flow (LeaveGame / RequestQuit).
// The Diablo-neutral table model lives in internal/game/ui; this screen keys
// its actions by row index.
type gameMenuScreen struct {
	lane    *deviceLane
	flow    *scene.FrontFlow
	stores  *persist.HeroStoreSet
	slot    int
	seed    uint32
	enabled func() bool // the dead-player gate stand-in (today: constant alive)

	// active marks the menu open (gmenu_is_active): the flow's MenuOpen
	// closure reads it and ProcessInput pauses the sim on it.
	active bool

	// table is the active table model; the root table until the Options
	// submenu swaps it in (X:gamemenu.cpp:185-191 GamemenuOptions).
	table *ui.MenuTable

	// inSubmenu marks the options table active; Previous Menu is the only
	// way back (X:gamemenu.cpp:101-104 GamemenuPrevious).
	inSubmenu bool

	// onRateSubmit submits one SetTickRate decision: wired by the boot to the
	// lane's bridge session so the screen holds no kernel knowledge.
	onRateSubmit func(rate int32) bool

	// sound is the audio engine the Music/Sound rows drive; nil (or an engine
	// that never initialised) keeps those rows disabled, gbSndInited false
	// (X:gamemenu.cpp:126-134).
	sound *audioEngine

	// levelTrack names the track a rising volume restarts (music_start(
	// GetLevelMusic(leveltype)), X:gamemenu.cpp:220-226).
	levelTrack func() string

	// brightness is the Gamma row's read/write seam; a zero knob means no
	// display is wired (the row still renders, it just cannot apply).
	brightness brightnessKnob

	// onClose notifies the boot that the menu fully closed, where upstream
	// calls SaveOptions (gmenu_set_items(nullptr), X:gmenu.cpp:246-248).
	onClose func()

	// lastError surfaces a refused action as the screen's error row.
	lastError error
}

// newGameMenuScreen builds the screen over the boot-owned surfaces. slot is
// the hero slot threaded from the selhero entry (the save target); seed is
// the session's level seed threaded the same way (0 for a fresh town — the
// load handler does not validate B for an unloaded town record, disclosed in
// the lane). enabled stays the dead-player gate stand-in field (nil =
// constant alive); the rate callback lives on the lane only — the screen
// reads the applied rate through the game's live TickDelay (review F9).
func newGameMenuScreen(
	lane *deviceLane, flow *scene.FrontFlow, stores *persist.HeroStoreSet,
	slot int, seed uint32,
) *gameMenuScreen {
	m := &gameMenuScreen{
		lane: lane, flow: flow, stores: stores,
		slot: slot, seed: seed,
		table: newGameMenuRootTable(),
	}
	m.onRateSubmit = func(rate int32) bool {
		return lane.input.SubmitDecision(kernel.Command{
			Kind: kernel.KindSetTickRate, Tick: lane.game.TickCount, A: rate,
		})
	}

	return m
}

// newGameMenuRootTable builds the single-player root table in upstream order
// (X:gamemenu.cpp:49-57); enabled flags refresh at open().
func newGameMenuRootTable() *ui.MenuTable {
	return ui.NewMenuTable([]ui.MenuItem{
		{Label: "Options", Enabled: true},
		{Label: "Save Game", Enabled: true},
		{Label: "Load Game", Enabled: true},
		{Label: "Exit to Main Menu", Enabled: true},
		{Label: "Quit Game", Enabled: true},
	})
}

// isOpen reports the menu's active state (gmenu_is_active): the flow's
// MenuOpen closure and the window's movement gate read it.
func (m *gameMenuScreen) isOpen() bool { return m.active }

// LastError reports the surfaced refusal of the last action, or nil.
func (m *gameMenuScreen) LastError() error { return m.lastError }

// Rows exposes the active table's rows to the render pass.
func (m *gameMenuScreen) Rows() []ui.MenuItem { return m.table.Rows() }

// Focused reports the focused row index.
func (m *gameMenuScreen) Focused() int { return m.table.Focused() }

// GameMenuOpen implements scene.GameMenuContent.
func (m *gameMenuScreen) GameMenuOpen() bool { return m.active }

// GameMenuRows implements scene.GameMenuContent: the active table's labels.
func (m *gameMenuScreen) GameMenuRows() []string {
	rows := m.table.Rows()
	labels := make([]string, len(rows))
	for i, r := range rows {
		labels[i] = r.Label
	}

	return labels
}

// GameMenuFocused implements scene.GameMenuContent.
func (m *gameMenuScreen) GameMenuFocused() int { return m.table.Focused() }

// GameMenuSlider implements scene.GameMenuContent: the focused row's slider
// fill in 0..1, negative when it carries no slider.
func (m *gameMenuScreen) GameMenuSlider() float64 {
	row := m.table.Focused()
	if !m.table.Rows()[row].Slider {
		return -1
	}

	// Each row divides by its OWN step count: 21 for Gamma, 46 for Speed, 64
	// for the volumes (X:gamemenu.cpp:128, :154, :175). The earlier version
	// divided every slider by the Speed row's 46.
	steps := max(m.table.SliderSteps(row), 2)

	return float64(m.table.Step(row)) / float64(steps)
}

// newGameMenuOptionsTable builds the options submenu in upstream order
// (X:gamemenu.cpp:68-75 sgOptionsMenu). Gamma is ALWAYS an enabled slider
// (upstream arms it unconditionally, X:gamemenu.cpp:153-155 — it needs no audio
// device); Music and Sound are disabled until the audio engine reports
// gbSndInited (X:gamemenu.cpp:126-134); Speed is wired; Previous Menu is the
// only way back.
func (m *gameMenuScreen) optionsTable() *ui.MenuTable {
	ready := m.soundReady()
	multiplayer := m.lane != nil && m.lane.game.Multiplayer

	table := ui.NewMenuTable([]ui.MenuItem{
		{Label: volumeLabel(musicLabelOn, musicLabelOff, ready), Enabled: ready, Slider: ready},
		{Label: volumeLabel(soundLabelOn, soundLabelOff, ready), Enabled: ready, Slider: ready},
		{Label: "Gamma", Enabled: true, Slider: true},
		{Label: "Speed", Enabled: !multiplayer, Slider: !multiplayer},
		{Label: "Previous Menu", Enabled: true},
	})

	table.SetSliderSteps(gamemenuRowSpeed, speedSliderSteps)
	table.SetSliderSteps(gamemenuRowGamma, gammaSliderSteps)
	table.SetSliderSteps(gamemenuRowMusic, engine.VolumeSteps)
	table.SetSliderSteps(gamemenuRowSound, engine.VolumeSteps)

	// GamemenuGetBrightness / GamemenuGetMusic / GamemenuGetSound seed the
	// sliders from the live values (X:gamemenu.cpp:126-156).
	table.SliderSet(gamemenuRowGamma, engine.BrightnessMin, engine.BrightnessMax, m.brightnessLevel())
	table.SliderSet(gamemenuRowSpeed, speedSliderMin, speedSliderMax, m.pendingTickRate())

	if ready {
		table.SliderSet(gamemenuRowMusic, engine.VolumeMin, engine.VolumeMax, m.sound.MusicVolume())
		table.SliderSet(gamemenuRowSound, engine.VolumeMin, engine.VolumeMax, m.sound.SoundVolume())
	}

	return table
}

// volumeLabel picks the on/off label pair by the engine state.
func volumeLabel(on, off string, ready bool) string {
	if ready {
		return on
	}

	return off
}

// soundReady reports gbSndInited for the menu rows (X:gamemenu.cpp:126-134).
func (m *gameMenuScreen) soundReady() bool {
	return m.sound != nil && m.sound.Initialised()
}

// open activates the menu: the update-fn role (X:gamemenu.cpp:91-99
// GamemenuUpdateSingle) recomputes the enabled flags — Load iff the slot
// holds a game — then focus lands on the FIRST ENABLED row
// (X:gmenu.cpp:238-244).
func (m *gameMenuScreen) open(slot int) {
	m.slot = slot
	m.table = newGameMenuRootTable()

	// Load Game enabled iff the slot holds a frozen session (gbValidSaveFile's
	// Go equivalent: persist.SessionArchive.HasGame, selhero.go:370-372 shape).
	archive := persist.NewSessionArchive(m.stores, lanePackSource)
	has, err := archive.HasGame(context.Background(), m.slot, string(persist.KindSingle))
	if err != nil {
		m.lastError = fmt.Errorf("game menu load check: %w", err)
		has = false
	}
	rows := m.table.Rows()
	rows[gamemenuRowLoad].Enabled = has

	// Options enabled iff the player is alive (X:gamemenu.cpp:97-99); the
	// death gate has not landed, so the stand-in is constant alive.
	if m.enabled != nil && !m.enabled() {
		rows[gamemenuRowOptions].Enabled = false
	}

	// Save Game enabled stands in for the hand-cursor gate
	// (X:gamemenu.cpp:333-335 pcurs != CURSOR_HAND refusal) — constant true
	// until the cursor lands. DISCLOSED stand-in.

	m.table.Focus()
	m.active = true
	m.inSubmenu = false
}

// speedLabel renders the Speed row's live label: in the submenu it reads the
// SLIDER's pending rate — upstream's GamemenuGetSpeed reads nTickRate, which
// the handler set immediately at the last keypress, while our event-sourced
// rate dispatches on the next settled tick; the slider position IS the
// pending rate (X:gamemenu.cpp:158-178).
func (m *gameMenuScreen) speedLabel() string {
	ticks := m.pendingTickRate()

	switch {
	case ticks >= 50:
		return "Speed - Fastest"
	case ticks >= 40:
		return "Speed - Faster"
	case ticks >= 30:
		return "Speed - Fast"
	default:
		return "Speed - Normal"
	}
}

// pendingTickRate reports the rate the Speed knob currently names: the
// slider position in the submenu, else the live delay's rate.
func (m *gameMenuScreen) pendingTickRate() int {
	if m.inSubmenu {
		return m.table.SliderGet(gamemenuRowSpeed, speedSliderMin, speedSliderMax)
	}

	return 1000 / int(m.lane.game.TickDelay()/time.Millisecond)
}

// refreshSpeedLabel re-reads the Speed row's label into the active table
// (the update-fn's per-frame label role, X:gmenu.cpp:256-257).
func (m *gameMenuScreen) refreshSpeedLabel() {
	if !m.inSubmenu {
		return
	}

	rows := m.table.Rows()
	rows[gamemenuRowSpeed].Label = m.speedLabel()
}

// close deactivates the menu (X:gmenu.cpp:290-292: Esc closes the menu) and
// notifies the boot, which persists the settings document on the same event
// (gmenu_set_items(nullptr) -> SaveOptions(), X:gmenu.cpp:246-248). Walking
// back to the root table is NOT a close: the menu is still up, so no save.
func (m *gameMenuScreen) close() {
	m.active = false
	m.inSubmenu = false

	if m.onClose != nil {
		m.onClose()
	}
}

// HandleKey routes the menu keys: navigation, activation, Esc close
// (X:gmenu.cpp:280-311 gmenu_presskeys). KeyUnknown is an idle frame and is
// never a press.
func (m *gameMenuScreen) HandleKey(key ui.Key) {
	if key == ui.KeyUnknown {
		return
	}

	if !m.active {
		// Esc on FrontGame with the menu closed opens it (X:diablo.cpp:546-551).
		if key == ui.KeyEscape {
			m.open(m.slot)
		}

		return
	}

	m.refreshSpeedLabel()

	switch key {
	case ui.KeyEscape:
		// Esc closes the WHOLE menu from the submenu too
		// (X:gmenu.cpp:290-292 gmenu_set_items(nullptr)).
		m.close()
	case ui.KeyEnter:
		m.selectRow(m.table.Focused())
	case ui.KeyArrowDown:
		m.table.MoveDown()
	case ui.KeyArrowUp:
		m.table.MoveUp()
	case ui.KeyArrowRight:
		m.stepRight()
	case ui.KeyArrowLeft:
		m.stepLeft()
	default:
	}
}

// stepRight steps the FOCUSED row's slider one step up and runs its handler
// (GmenuLeftRight(true), X:gmenu.cpp:107-124); a no-op on non-slider rows and
// outside the submenu.
func (m *gameMenuScreen) stepRight() { m.stepFocused(m.table.SliderStepRight) }

// stepLeft is the left-arrow twin of stepRight (GmenuLeftRight(false)).
func (m *gameMenuScreen) stepLeft() { m.stepFocused(m.table.SliderStepLeft) }

// stepFocused moves the focused row one step and applies it.
func (m *gameMenuScreen) stepFocused(step func(row int) int) {
	row := m.table.Focused()
	if !m.inSubmenu || !m.table.Rows()[row].Slider {
		return
	}

	step(row)
	m.applySlider(row)
}

// applySlider runs the row's own handler after a step: the same functions
// upstream's per-item fnMenu carries (X:gmenu.cpp:120-123).
func (m *gameMenuScreen) applySlider(row int) {
	switch row {
	case gamemenuRowMusic:
		m.applyMusic()
	case gamemenuRowSound:
		m.applySound()
	case gamemenuRowGamma:
		m.applyGamma()
	case gamemenuRowSpeed:
		m.submitSpeed()
	case gamemenuRowPrevious:
	}

	m.refreshSpeedLabel()
}

// brightnessLevel reads the live gamma level, defaulting to the upstream 0 when
// no display is wired (Source/options.cpp:796).
func (m *gameMenuScreen) brightnessLevel() int {
	if m.brightness.read == nil {
		return engine.BrightnessMin
	}

	return m.brightness.read()
}

// setBrightness applies a gamma level through the boot's display seam
// (UpdateBrightness, Source/engine/palette.cpp:253-261).
func (m *gameMenuScreen) setBrightness(level int) {
	if m.brightness.write != nil {
		m.brightness.write(engine.CapBrightness(level))
	}
}

// applyGamma writes the Gamma slider's position into the display
// (GamemenuBrightness's non-activate arm, X:gamemenu.cpp:246-262).
func (m *gameMenuScreen) applyGamma() {
	m.setBrightness(m.table.SliderGet(gamemenuRowGamma, engine.BrightnessMin, engine.BrightnessMax))
}

// gammaToggle is the Gamma row's Enter rule: 0 becomes 100 and anything else
// becomes 0 (GamemenuBrightness(true), X:gamemenu.cpp:250-256).
func (m *gameMenuScreen) gammaToggle() {
	level := engine.BrightnessMin
	if m.brightnessLevel() == engine.BrightnessMin {
		level = engine.BrightnessMax
	}

	m.setBrightness(level)
	m.table.SliderSet(gamemenuRowGamma, engine.BrightnessMin, engine.BrightnessMax, level)
}

// submitSpeed submits the quantized rate at the Speed row's slider position.
func (m *gameMenuScreen) submitSpeed() {
	rate := m.table.SliderGet(gamemenuRowSpeed, speedSliderMin, speedSliderMax)
	m.onRateSubmit(int32(rate))
}

// speedToggle applies the Enter rule: reset to 20 unless already 20, then 50
// (X:gamemenu.cpp:265-271), and submits the new rate.
func (m *gameMenuScreen) speedToggle() {
	if !m.inSubmenu || !m.table.Rows()[gamemenuRowSpeed].Enabled {
		return
	}

	ms := int(m.lane.game.TickDelay() / time.Millisecond)
	rate := 1000 / ms
	if rate == speedSliderMin {
		rate = speedSliderMax
	} else {
		rate = speedSliderMin
	}

	m.table.SliderSet(gamemenuRowSpeed, speedSliderMin, speedSliderMax, rate)
	m.onRateSubmit(int32(rate))
}

// applyMusic is GamemenuMusicVolume's non-activate arm (X:gamemenu.cpp:209-227):
// set the volume, then stop at VOLUME_MIN and restart on the way back up.
func (m *gameMenuScreen) applyMusic() {
	if !m.soundReady() {
		return
	}

	m.sound.SetMusicVolume(m.table.SliderGet(gamemenuRowMusic, engine.VolumeMin, engine.VolumeMax))

	if !m.sound.MusicOn() {
		m.sound.StopMusic()

		return
	}

	if m.sound.Track() == musicTrackNone {
		m.startLevelMusic()
	}
}

// applySound is GamemenuSoundVolume's non-activate arm (X:gamemenu.cpp:229-251);
// it also replays the menu move cue.
func (m *gameMenuScreen) applySound() {
	if !m.soundReady() {
		return
	}

	m.sound.SetSoundVolume(m.table.SliderGet(gamemenuRowSound, engine.VolumeMin, engine.VolumeMax))
	m.playMenuMove()
}

// musicToggle is the Music row's Enter rule: on becomes MIN (and stops), off
// becomes MAX (and restarts the level track) — X:gamemenu.cpp:194-208.
func (m *gameMenuScreen) musicToggle() {
	if !m.soundReady() {
		return
	}

	if m.sound.MusicOn() {
		m.sound.SetMusicVolume(engine.VolumeMin)

		return
	}

	m.sound.SetMusicVolume(engine.VolumeMax)
	m.startLevelMusic()
}

// soundToggle is the Sound row's Enter rule (X:gamemenu.cpp:229-243).
func (m *gameMenuScreen) soundToggle() {
	if !m.soundReady() {
		return
	}

	if m.sound.SoundOn() {
		m.sound.SetSoundVolume(engine.VolumeMin)

		return
	}

	m.sound.SetSoundVolume(engine.VolumeMax)
}

// startLevelMusic restarts the level track through the boot's track seam
// (music_start(GetLevelMusic(leveltype)), X:gamemenu.cpp:220-226).
func (m *gameMenuScreen) startLevelMusic() {
	if m.levelTrack == nil {
		return
	}

	m.sound.startTrack(m.levelTrack())
}

// playMenuMove replays the menu's own cue (PlaySFX(SfxID::MenuMove),
// X:gamemenu.cpp:248).
func (m *gameMenuScreen) playMenuMove() {
	if err := m.sound.PlaySFX(snd.SfxMenuMove); err != nil {
		m.lastError = err
	}
}

// selectRow runs the focused row's action (X:gmenu.cpp:285-289: Enter on an
// enabled row only).
func (m *gameMenuScreen) selectRow(idx int) {
	if !m.table.Rows()[idx].Enabled {
		return
	}

	if m.inSubmenu {
		switch idx {
		case gamemenuRowMusic:
			m.musicToggle()
		case gamemenuRowSound:
			m.soundToggle()
		case gamemenuRowGamma:
			m.gammaToggle()
		case gamemenuRowSpeed:
			m.speedToggle()
		case gamemenuRowPrevious:
			// GamemenuPrevious (X:gamemenu.cpp:101-104): swap the root
			// table back in.
			m.table = newGameMenuRootTable()
			m.table.Focus()
			m.inSubmenu = false
		}

		return
	}

	switch idx {
	case gamemenuRowOptions:
		// GamemenuOptions (X:gamemenu.cpp:185-191): swap the options table
		// in place, refreshed from the live rate, focused on the first
		// enabled row.
		m.table = m.optionsTable()
		m.table.Focus()
		m.inSubmenu = true
	case gamemenuRowSave:
		m.saveGame()
	case gamemenuRowLoad:
		m.loadGame()
	case gamemenuRowExit:
		m.exitToMainMenu()
	case gamemenuRowQuit:
		m.quitGame()
	}
}

// saveGame submits KindSaveSlot{A: slot, B: seed} through the lane's bridge
// session (X:gamemenu.cpp:331-364 gamemenu_save_game's SaveGame verb) and
// arms the lane's save outcome hook. Upstream-faithful settle path (review
// F1 fix, shape a): gamemenu_off() runs FIRST (X:gamemenu.cpp:331-336), so
// the menu closes here and the command dispatches on the now-running tick;
// the outcome hook reports the result onto the game screen.
func (m *gameMenuScreen) saveGame() {
	if m.lane == nil {
		return
	}

	m.close()

	cmd := kernel.Command{
		Kind: kernel.KindSaveSlot,
		Tick: m.lane.game.TickCount,
		A:    int32(m.slot),
		B:    int32(m.seed),
	}

	if !m.lane.input.SubmitDecision(cmd) {
		return
	}

	m.lane.armOutcome(kernel.KindSaveSlot, kernel.KindSavedSlot, func(ev kernel.Event) {
		if ev.Kind != kernel.KindSavedSlot {
			m.lastError = fmt.Errorf("game menu save: kernel refused the slot")
		}
	})
}

// loadGame submits KindLoadSlot{A: slot, B: seed} with the selhero load's
// peek (selhero.go:421-428) and arms the load outcome hook. Upstream-faithful
// settle path (review F1 fix, shape a): gamemenu_off() runs FIRST
// (X:gamemenu.cpp:297-301), so the menu closes here and the command
// dispatches on the now-running tick; the hook reports onto the game screen.
func (m *gameMenuScreen) loadGame() {
	if m.lane == nil {
		return
	}

	m.close()

	archive := persist.NewSessionArchive(m.stores, lanePackSource)

	peek, err := archive.LoadSlot(m.slot)
	if err != nil {
		m.lastError = fmt.Errorf("game menu load peek: %w", err)

		return
	}

	cmd := kernel.Command{
		Kind: kernel.KindLoadSlot,
		Tick: m.lane.game.TickCount,
		A:    int32(m.slot),
		B:    int32(peek.LevelSeed),
	}

	if !m.lane.input.SubmitDecision(cmd) {
		return
	}

	m.lane.armOutcome(kernel.KindLoadSlot, kernel.KindLoadedSlot, func(ev kernel.Event) {
		if ev.Kind != kernel.KindLoadedSlot {
			m.lastError = fmt.Errorf("game menu load: kernel refused the slot")
		}
	})
}

// exitToMainMenu clears the lane's hooks, closes the menu and walks the flow
// back to the main menu (X:gamemenu.cpp:106-121 GamemenuNewGame's
// gbRunGame = false arm).
func (m *gameMenuScreen) exitToMainMenu() {
	m.lane.clearOutcomes()
	m.close()
	m.flow.LeaveGame()
}

// quitGame asks the flow to end the boot rc-0 (X:gamemenu.cpp:287-295
// gamemenu_quit_game).
func (m *gameMenuScreen) quitGame() {
	m.close()
	m.flow.RequestQuit()
}
