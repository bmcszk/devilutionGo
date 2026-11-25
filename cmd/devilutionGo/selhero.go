//go:build windowed

package main

import (
	"context"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// selheroCreateStep names the create flow's input steps, in selhero.cpp's
// own order: the name first, then the class pick.
type selheroCreateStep int

const (
	// selheroNameStep is the name-entry step (selhero.cpp's S text buffer).
	selheroNameStep selheroCreateStep = iota
	// selheroClassStep is the class-pick step (the six-class list).
	selheroClassStep
	// selheroDoneStep means the create submitted and the screen is closed.
	selheroDoneStep
)

// selhero controls indices inside the screen's menu list, upstream order
// (selhero.cpp:60-75): the hero rows first, then the fixed actions.
const (
	selheroActionCreate = iota // first action row after the hero rows
	selheroActionDelete
	selheroActionQuit
)

// selHeroScreen is the single-player entry screen (devilGo-5fr4): the hero
// list reads the persist store, the create flow submits KindCreatePack through
// the live kernel, delete hits DeleteHero and the class preview panel reads
// the committed class loaders. Presentation only: it invents no sim code, it
// submits the verbs the existing surface carries.
type selHeroScreen struct {
	store *persist.SQLiteStore
	lane  *deviceLane

	// list is the screen's menu: the hero names plus the three action rows.
	list    *ui.MenuList
	heroes  []persist.HeroRecord
	classes []txtdata.ClassData

	// classMenu is the create flow's class-pick list; nil outside the class
	// step of create mode.
	classMenu *ui.MenuList

	// preview is the focused class's stats panel; loaded from the committed
	// attributes.tsv through the typed loader (previewData marks validity).
	preview     txtdata.ClassAttributes
	previewData bool

	// focusedHero tracks the hero row navigation last highlighted: the
	// target the Delete action removes (selhero.cpp:139-146), nil before
	// any hero row was focused.
	focusedHero *persist.HeroRecord

	// creating carries the active create flow; nil outside create mode.
	creating *selheroCreate

	// lastCreateError carries the create flow's failure for the player:
	// a class the kernel refuses (the demo pack source serves warrior only)
	// surfaces here instead of silently leaving a ghost row (review F5).
	lastCreateError error

	// onBack asks the front flow to walk back to the main menu; wired at
	// construction so the screen never imports the flow.
	onBack func()
}

// selheroCreate is one create flow in progress: the name typed so far and the
// step the flow sits at.
type selheroCreate struct {
	name  string
	class int
	step  selheroCreateStep
}

// newSelHeroScreen wires the screen over the store and the lane's live kernel:
// exactly the two surfaces the screen consumes, nothing invented. onBack is
// the walk-back the flow owns (main menu), invoked by the screen's own Quit
// action and after a finished create.
func newSelHeroScreen(store *persist.SQLiteStore, lane *deviceLane, onBack func()) *selHeroScreen {
	s := &selHeroScreen{store: store, lane: lane, onBack: onBack}
	s.refresh(context.Background())

	return s
}

// selHeroRows reports the screen's label rows for the render pass: hero
// names first, then the three action rows (refresh's layout). The scene
// band draws one row per label (review F2).
func (s *selHeroScreen) SelHeroRows() []string {
	items := make([]string, 0, len(s.heroes)+3)
	items = append(items, s.heroNames()...)
	items = append(items, "Create Hero", "Delete Hero", "Quit")

	return items
}

// SelHeroFocused reports the menu list's focused index for the render pass,
// or -1 when nothing is focused.
func (s *selHeroScreen) SelHeroFocused() int {
	return s.list.FocusedIndex()
}

// SelHeroPreviewRows reports the stats preview panel's value rows: the class
// baselines the screen loaded (selhero.cpp:527-546: level, strength, magic,
// dexterity, vitality), or 0 while no preview is loaded.
func (s *selHeroScreen) SelHeroPreviewRows() int {
	if !s.previewData {
		return 0
	}

	return selheroPreviewStatRows
}

// SelHeroClassRows reports the class-pick rows while the create flow sits at
// the class step (xoap X3), or nil at every other step — the band then draws
// the regular hero/action list instead.
func (s *selHeroScreen) SelHeroClassRows() []string {
	if s.creating == nil || s.creating.step != selheroClassStep || s.classMenu == nil {
		return nil
	}

	return classNames(s.classes)
}

// SelHeroClassFocused reports the focused index inside the class-pick rows,
// or -1 when the class step is not active.
func (s *selHeroScreen) SelHeroClassFocused() int {
	if s.SelHeroClassRows() == nil || s.classMenu == nil {
		return -1
	}

	return s.classMenu.FocusedIndex()
}

// SelHeroHasErrorRow reports a surfaced create refusal (xoap X4): once the
// create flow has failed, the band draws one extra error row.
func (s *selHeroScreen) SelHeroHasErrorRow() bool {
	return s.lastCreateError != nil
}

// selheroPreviewStatRows is selhero.cpp:527-546's stat block height: the level
// line plus the four attributes the panel lists.
const selheroPreviewStatRows = 5

// heroNames reports the screen's hero rows in slot order: the list projection
// the tests read.
func (s *selHeroScreen) heroNames() []string {
	names := make([]string, 0, len(s.heroes))

	for _, h := range s.heroes {
		names = append(names, h.Name)
	}

	return names
}

// refresh reloads the hero list from the store and rebuilds the menu: the
// hero rows in slot order (pfile.cpp:700-710 order) plus the action rows.
func (s *selHeroScreen) refresh(ctx context.Context) error {
	heroes, err := s.store.ListHeroes(ctx)
	if err != nil {
		return fmt.Errorf("selhero list heroes: %w", err)
	}

	s.heroes = heroes

	s.list = ui.NewMenuList(s.SelHeroRows())
	s.list.OnSelect(s.selectItem)
	s.list.OnCancel(s.cancel)

	// The rebuilt list focuses its first row; upstream arms the Delete
	// button on the hero row the list opens with (selhero.cpp:564 UiInitList
	// selectedItem + 139-146), so the initial focus is a hero when one
	// exists.
	s.trackFocusedHero()

	return nil
}

// HandleKey routes one key press: create mode consumes keys in its own steps,
// otherwise the menu list navigates — and each navigation step updates the
// focused-hero tracking the Delete action reads (selhero.cpp:139-146).
func (s *selHeroScreen) HandleKey(key ui.Key) {
	if s.creating != nil {
		s.handleCreateKey(key)

		return
	}

	switch key {
	case ui.KeyUnknown:
		return
	case ui.KeyEscape:
		s.cancel()
	default:
		s.list.HandleKey(key)
		s.trackFocusedHero()
	}
}

// trackFocusedHero records the hero row navigation last highlighted: the
// row the Delete action removes. Navigation keys landed on the list just
// before this read, so the focused index is current.
func (s *selHeroScreen) trackFocusedHero() {
	idx := s.list.FocusedIndex()
	if idx >= 0 && idx < len(s.heroes) {
		s.focusedHero = &s.heroes[idx]

		return
	}

	// Leaving the hero rows (into the actions) keeps the last hero focused:
	// the Delete action itself sits below the rows, and upstream arms the
	// delete on the hero that WAS focused (selhero.cpp:139-146).
}

// Screen reports the screen kind the flow tracks: always the selhero screen.
// Kept for the ui.Screen contract's readability.

// selectItem runs the focused row's action: a hero row or create opens the
// create flow (the demo kernel raises one hero per session), delete removes
// the focused hero, quit walks back.
func (s *selHeroScreen) selectItem(idx int) {
	if s.creating != nil {
		return
	}

	switch {
	case idx < len(s.heroes), idx == len(s.heroes)+selheroActionCreate:
		// A hero row or the Create action opens the create flow: the
		// single-session lane raises its own hero through the kernel.
		_ = s.beginCreate()
	case idx == len(s.heroes)+selheroActionDelete:
		s.deleteFocused()
	case idx == len(s.heroes)+selheroActionQuit:
		s.cancel()
	default:
	}
}

// beginCreate opens the create flow at the name step: name first, then the
// class pick (selhero.cpp's own order).
func (s *selHeroScreen) beginCreate() error {
	s.creating = &selheroCreate{step: selheroNameStep}

	return nil
}

// typeName appends to the create flow's name buffer, byte by byte through
// the key set the widget kit carries.
func (s *selHeroScreen) typeName(name string) error {
	if s.creating == nil || s.creating.step != selheroNameStep {
		return fmt.Errorf("selhero: not at the name step")
	}

	s.creating.name += name

	return nil
}

// confirmName closes the name step and opens the class pick, loading the
// committed classdat rows for the pick list and preview.
func (s *selHeroScreen) confirmName() {
	if s.creating == nil {
		return
	}

	s.creating.step = selheroClassStep

	if data, err := scene.ReadCommittedFixture("classes", "classdat.tsv"); err == nil {
		if tbl, perr := enginetxtdata.Parse(data); perr == nil {
			if classes, cerr := txtdata.LoadClasses(tbl); cerr == nil {
				s.classes = classes
			}
		}
	}

	if len(s.classes) == 0 {
		s.classes = []txtdata.ClassData{{ClassName: "Warrior", FolderName: "warrior"}}
	}

	s.classMenu = ui.NewMenuList(classNames(s.classes))
	s.classMenu.OnSelect(s.pickClass)

	_ = s.loadPreview(0)
}

// handleCreateKey routes one key press inside create mode: the name step
// accepts only confirm keys (the window layer owns raw text input), the class
// step navigates the pick list.
func (s *selHeroScreen) handleCreateKey(key ui.Key) {
	if key == ui.KeyEscape {
		s.cancel()

		return
	}

	switch s.creating.step {
	case selheroNameStep:
		if key == ui.KeyEnter || key == ui.KeySpace {
			s.confirmName()
		}
	case selheroClassStep:
		if s.classMenu != nil {
			s.classMenu.HandleKey(key)
			// The preview panel follows the focused class (selhero.cpp:139-146
			// refills the stat block on focus), so the player sees the class
			// they are about to pick.
			_ = s.loadPreview(s.classMenu.FocusedIndex())
		}
	case selheroDoneStep:
		s.creating = nil
	}
}

// pickClass confirms the focused class: the preview panel already shows its
// stats, so activating submits KindCreatePack through the live kernel. The
// store row is written ONLY after the kernel raises the hero (the next
// settled tick's CreatedPack event): a class the pack source refuses — the
// demo serves warrior only — leaves no ghost row and surfaces the refusal
// through lastCreateError (review F5).
func (s *selHeroScreen) pickClass(idx int) {
	if s.creating == nil || s.lane == nil {
		return
	}

	class := idx
	name := s.creating.name

	cmd := kernel.Command{
		Kind: kernel.KindCreatePack,
		Tick: s.lane.game.TickCount,
		A:    int32(class),
		B:    heroSpawnX,
		C:    heroSpawnY,
	}

	if !s.lane.input.SubmitDecision(cmd) {
		return
	}

	s.lane.pendingCreate = true

	// The create closes at the class pick; the pending decision's outcome
	// lands with the next settled tick, so persistence hooks the lane's
	// post-tick read of the kernel trace instead of running before it.
	s.lane.onCreateEvent = func(ev kernel.Event) {
		s.lane.onCreateEvent = nil

		if ev.Kind != kernel.KindCreatedPack {
			s.lastCreateError = fmt.Errorf(
				"selhero create: kernel refused class %d (demo pack source serves warrior only)", class)

			// A refusal keeps the player on the selhero list so the error row
			// is visible; only a successful create closes the screen.
			s.creating = nil

			return
		}

		if err := s.store.SaveHero(context.Background(), persist.HeroRecord{
			Slot:    0,
			Name:    name,
			ClassID: class,
			Pack:    make([]byte, persist.PlayerPackSize),
		}); err != nil {
			s.lastCreateError = fmt.Errorf("selhero save hero: %w", err)

			return
		}

		// Only a successful create closes the screen; a refusal keeps the
		// player on the selhero list so the error row is visible.
		s.close()
	}

	s.creating.step = selheroDoneStep
	s.creating = nil
}

// deleteFocused removes the FOCUSED hero through the store's own
// transactional DeleteHero and refreshes the list (review F6): the screen
// tracks the last hero row navigation highlighted — the upstream contract
// (selhero.cpp:139-146), where the Delete button arms only while a hero row
// holds focus and deletes that hero — not a fixed offset from the action row.
func (s *selHeroScreen) deleteFocused() {
	if s.focusedHero == nil {
		// No hero row was ever focused: upstream disables the Delete button
		// until a hero holds focus (selhero.cpp:557 ElementDisabled), so the
		// action is a no-op, not a delete of some implicit row.
		return
	}

	if err := s.store.DeleteHero(context.Background(), s.focusedHero.Slot); err != nil {
		return
	}

	s.focusedHero = nil
	_ = s.refresh(context.Background())
}

// classPreview reports the class ordinal's stat baselines from the committed
// attributes.tsv through the typed loader: the data the preview panel shows.
func (s *selHeroScreen) classPreview(class int) (txtdata.ClassAttributes, error) {
	data, err := scene.ReadCommittedFixture(classDirName(class), "attributes.tsv")
	if err != nil {
		return txtdata.ClassAttributes{}, fmt.Errorf("selhero preview fixture: %w", err)
	}

	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return txtdata.ClassAttributes{}, fmt.Errorf("selhero preview parse: %w", err)
	}

	attrs, _, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		return txtdata.ClassAttributes{}, fmt.Errorf("selhero preview loader: %w", err)
	}

	s.preview = attrs
	s.previewData = true

	return attrs, nil
}

// loadPreview warms the preview panel for the focused class ordinal; failures
// leave the panel empty rather than showing stale numbers.
func (s *selHeroScreen) loadPreview(class int) error {
	_, err := s.classPreview(class)

	return err
}

// cancel walks back: create mode dissolves, otherwise the screen asks the flow
// to return to the main menu.
func (s *selHeroScreen) cancel() {
	if s.creating != nil {
		s.creating = nil

		return
	}

	if s.onBack != nil {
		s.onBack()
	}
}

// classNames projects the class rows to pick-list labels.
func classNames(classes []txtdata.ClassData) []string {
	names := make([]string, 0, len(classes))
	for _, c := range classes {
		names = append(names, c.ClassName)
	}

	return names
}

// classDirName maps a class ordinal to its fixture directory: classdat.tsv's
// own row order (Warrior, Rogue, Sorcerer, Monk, Bard, Barbarian).
func classDirName(class int) string {
	dirs := []string{"warrior", "rogue", "sorcerer", "monk", "bard", "barbarian"}
	if class < 0 || class >= len(dirs) {
		return ""
	}

	return dirs[class]
}

// close ends the screen: it asks the flow for the main menu.
func (s *selHeroScreen) close() {
	if s.onBack != nil {
		s.onBack()
	}
}

// keep the flow-scene import honest: the screen's back-walk is the flow's
// own walkBack, wired through SetSelHero.
var _ scene.FrontScreen = scene.FrontSelHero
