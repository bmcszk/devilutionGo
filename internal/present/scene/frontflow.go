package scene

import "github.com/bmcszk/devilutionGo/internal/game/ui"

const (
	// FrontTitle is the title screen: any key advances to the main menu,
	// Escape quits (mainmenu.cpp's pre-menu title flow).
	FrontTitle FrontScreen = iota
	// FrontMainMenu is the main menu: Single Player / Exit
	// (mainmenu.cpp's submenu).
	FrontMainMenu
	// FrontSelHero is the single-player entry screen: the wired boot installs
	// the real hero list/create/delete UI over it (devilGo-5fr4); without an
	// installed screen it keeps the placeholder's back-walk-only contract.
	FrontSelHero
)

// menuItemSinglePlayer and menuItemExit are the menu indices in upstream
// order (mainmenu.cpp:SinglePlayerDir -> mainmenu.cpp:72-79: SINGLEPLAYER
// then EXIT).
const (
	menuItemSinglePlayer = iota
	menuItemExit
)

// FrontScreen identifies the active front screen of the windowed boot
// (devilGo-uuyh).
type FrontScreen int

// SelHeroContent is the selhero screen content the render pass draws: the
// band rows the screen's list carries. The flow delegates through this
// interface so the scene package never imports the cmd wiring (the screen
// itself is behind the windowed build in cmd/devilutionGo).
type SelHeroContent interface {
	// SelHeroRows reports the screen's label rows: hero names first, then
	// the action labels (refresh's layout).
	SelHeroRows() []string
	// SelHeroFocused reports the focused list index, or -1.
	SelHeroFocused() int
	// SelHeroPreviewRows reports the stats preview panel's value rows: the
	// class baselines the screen loaded (selhero.cpp:527-546 stat block),
	// or 0 while no preview is loaded.
	SelHeroPreviewRows() int
	// SelHeroClassRows reports the class-pick rows while the create flow
	// sits at the class step (selhero.cpp's class list), or nil/empty at
	// every other step — the band then draws the regular list instead.
	SelHeroClassRows() []string
	// SelHeroClassFocused reports the focused index inside the class-pick
	// rows, or -1 when the class step is not active.
	SelHeroClassFocused() int
	// SelHeroHasErrorRow reports a surfaced create refusal: once true the
	// band draws one extra error row under the list rows.
	SelHeroHasErrorRow() bool
}

// FrontFlow is the front-screen flow: title → main menu → single-player entry
// (devilGo-uuyh). Presentation only: it drives no kernel state and reads none —
// the town scene stays the terminal render behind it, and the flow sits IN
// FRONT of it in the draw order. The single-player entry delegates its keys,
// rows and entry-time refresh to the installed screen (devilGo-5fr4).
type FrontFlow struct {
	screen FrontScreen
	menu   *ui.MenuList
	quit   bool
	// selhero is the single-player screen the flow delegates keys to when
	// set; without it FrontSelHero keeps the placeholder's back-walk-only
	// contract.
	selhero ui.Screen
	// selheroContent is the same screen's render half: the rows the band
	// draws on FrontSelHero (review F2). Set together with selhero by
	// SetSelHero; the placeholder draws the flow's own fallback rows.
	selheroContent SelHeroContent
	// onSelHeroEnter runs once per entry INTO FrontSelHero: the screen's
	// entry-time refresh (review F2 — the list populates when the screen is
	// entered, not only after a delete). Nil on the placeholder path.
	onSelHeroEnter func()
}

// NewFrontFlow builds the flow the windowed boot installs: it lands on the
// title (mainmenu.cpp boots to the title screen, not straight into the world).
func NewFrontFlow() *FrontFlow {
	flow := &FrontFlow{screen: FrontTitle}
	flow.menu = flow.buildMenu()

	return flow
}

// SetSelHero installs the real single-player screen: HandleKey routes
// FrontSelHero keys to it instead of the placeholder walk-back, its rows
// render on the band and its entry hook refreshes the list each time the
// flow walks INTO FrontSelHero (review F2).
func (f *FrontFlow) SetSelHero(s ui.Screen, content SelHeroContent, onEnter func()) {
	f.selhero = s
	f.selheroContent = content
	f.onSelHeroEnter = onEnter
}

// RequestBackWalk moves one screen toward the title: the selhero screen's own
// back-walk hook (Quit, or a finished create), the same walk the flow's
// Escape route applies.
func (f *FrontFlow) RequestBackWalk() { f.walkBack() }

// Screen reports the active front screen: the black-box probe the boot test
// reads.
func (f *FrontFlow) Screen() FrontScreen {
	return f.screen
}

// Quit reports that the flow asked to end the boot: Exit, or Escape on the
// title. The window maps it to a clean rc-0 exit (ebiten.Termination).
func (f *FrontFlow) Quit() bool {
	return f.quit
}

// HandleKey routes one key press to the active screen: the title advances on
// any key except Escape; the other screens delegate to their widgets.
func (f *FrontFlow) HandleKey(key ui.Key) {
	switch f.screen {
	case FrontTitle:
		f.handleTitleKey(key)
	case FrontMainMenu:
		f.menu.HandleKey(key)
	case FrontSelHero:
		f.handleSelHeroKey(key)
	default:
	}
}

// buildMenu constructs the main menu's MenuList from the widget kit AS-IS:
// Single Player / Exit, navigation and selection delegated to the kit.
func (f *FrontFlow) buildMenu() *ui.MenuList {
	menu := ui.NewMenuList([]string{"Single Player", "Exit"})
	menu.OnSelect(func(idx int) {
		switch idx {
		case menuItemSinglePlayer:
			f.enterSelHero()
		case menuItemExit:
			f.quit = true
		default:
		}
	})
	menu.OnCancel(func() {
		f.walkBack()
	})

	return menu
}

// handleTitleKey applies the title screen's rule: Escape quits (diablo.cpp
// quit surface), any other key opens the main menu (title.cpp's press-any-key
// flow).
func (f *FrontFlow) handleTitleKey(key ui.Key) {
	// A zero-code key is an idle tick, not a press (review F1): the window
	// seam (window_ebiten.go:windowFlowKey) feeds KeyUnknown on every frame
	// the user touches nothing, and the title's press-any-key rule must not
	// read that as a key down. The flow is the honest layer: the gate holds
	// for every seam that feeds it, windowed or test.
	if key == ui.KeyUnknown {
		return
	}

	if key == ui.KeyEscape {
		f.quit = true

		return
	}

	f.screen = FrontMainMenu
}

// handleSelHeroKey routes FrontSelHero keys to the real screen when one is
// installed (the screen also consumes Escape), else the placeholder's
// back-walk-only contract applies.
func (f *FrontFlow) handleSelHeroKey(key ui.Key) {
	if f.selhero != nil {
		f.selhero.HandleKey(key)

		return
	}

	if key == ui.KeyEscape {
		f.walkBack()
	}
}

// enterSelHero walks INTO the selhero screen and runs the screen's entry
// hook once per entry (review F2): the list refreshes from the store at
// entry time, not only after a delete.
func (f *FrontFlow) enterSelHero() {
	f.screen = FrontSelHero

	if f.onSelHeroEnter != nil {
		f.onSelHeroEnter()
	}
}

// walkBack moves one screen toward the title: placeholder → menu → title,
// where the title screen's own Escape quit takes over.
func (f *FrontFlow) walkBack() {
	switch f.screen {
	case FrontSelHero:
		f.screen = FrontMainMenu
	case FrontMainMenu:
		f.screen = FrontTitle
	default:
	}
}
