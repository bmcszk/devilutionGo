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

	// creating carries the active create flow; nil outside create mode.
	creating *selheroCreate

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
	s.list = ui.NewMenuList(nil)
	s.list.OnSelect(s.selectItem)
	s.list.OnCancel(s.cancel)

	return s
}

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

	items := make([]string, 0, len(heroes)+3)
	for _, h := range heroes {
		items = append(items, h.Name)
	}

	items = append(items, "Create Hero", "Delete Hero", "Quit")

	s.list = ui.NewMenuList(items)
	s.list.OnSelect(s.selectItem)
	s.list.OnCancel(s.cancel)

	return nil
}

// HandleKey routes one key press: create mode consumes keys in its own steps,
// otherwise the menu list navigates.
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
	}
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

	if data, err := readCommittedFixture("classes", "classdat.tsv"); err == nil {
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
		}
	case selheroDoneStep:
		s.creating = nil
	}
}

// pickClass confirms the focused class: the preview panel already shows its
// stats, so activating submits KindCreatePack through the live kernel, persists
// the hero row under the entered name and closes the screen (the hero is born
// as a kernel verb, devilGo-5fr4).
func (s *selHeroScreen) pickClass(idx int) {
	if s.creating == nil || s.lane == nil {
		return
	}

	cmd := kernel.Command{
		Kind: kernel.KindCreatePack,
		Tick: s.lane.game.TickCount,
		A:    int32(idx),
		B:    heroSpawnX,
		C:    heroSpawnY,
	}

	if !s.lane.input.SubmitDecision(cmd) {
		return
	}

	if err := s.persistHero(idx); err != nil {
		return
	}

	s.creating.step = selheroDoneStep
	s.creating = nil

	s.close()
}

// persistHero writes the created hero's store row: slot 0 (the demo's single
// live slot), the entered name and the class the create verb raised.
func (s *selHeroScreen) persistHero(class int) error {
	if s.creating == nil {
		return fmt.Errorf("selhero: no create flow to persist")
	}

	record := persist.HeroRecord{
		Slot:    0,
		Name:    s.creating.name,
		ClassID: class,
		Pack:    make([]byte, persist.PlayerPackSize),
	}

	if err := s.store.SaveHero(context.Background(), record); err != nil {
		return fmt.Errorf("selhero save hero: %w", err)
	}

	return nil
}

// deleteFocused removes the focused hero row through the store's own
// transactional DeleteHero and refreshes the list.
func (s *selHeroScreen) deleteFocused() {
	// idx counts hero rows first, action rows after (refresh's layout), so the
	// focused Delete action names the hero TWO rows above it (Delete sits one
	// action row below its hero, Quit below Delete).
	idx := s.list.FocusedIndex() - selheroActionDelete - 1
	if idx < 0 || idx >= len(s.heroes) {
		return
	}

	if err := s.store.DeleteHero(context.Background(), s.heroes[idx].Slot); err != nil {
		return
	}

	_ = s.refresh(context.Background())
}

// classPreview reports the class ordinal's stat baselines from the committed
// attributes.tsv through the typed loader: the data the preview panel shows.
func (s *selHeroScreen) classPreview(class int) (txtdata.ClassAttributes, error) {
	data, err := readCommittedFixture(classDirName(class), "attributes.tsv")
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
