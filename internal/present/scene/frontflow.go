package scene

import "github.com/bmcszk/devilutionGo/internal/game/ui"

const (
	// FrontTitle is the title screen: any key advances to the main menu,
	// Escape quits (mainmenu.cpp's pre-menu title flow).
	FrontTitle FrontScreen = iota
	// FrontMainMenu is the main menu: Single Player / Exit
	// (mainmenu.cpp's submenu).
	FrontMainMenu
	// FrontSelHero is the single-player entry PLACEHOLDER screen: it
	// renders and accepts back-navigation, nothing more — the real
	// list/create/delete UI is devilGo-5fr4.
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

// FrontFlow is the minimal front-screen flow: title → main menu →
// single-player entry placeholder (devilGo-uuyh). Presentation only: it
// drives no kernel state and reads none — the town scene stays the terminal
// render behind it, and the flow sits IN FRONT of it in the draw order.
type FrontFlow struct {
	screen FrontScreen
	menu   *ui.MenuList
	quit   bool
	// selhero is the single-player screen the flow delegates to when set;
	// without it FrontSelHero keeps the placeholder's back-walk-only contract
	// (devilGo-5fr4 wires the real screen in).
	selhero ui.Screen
}

// NewFrontFlow builds the flow the windowed boot installs: it lands on the
// title (mainmenu.cpp boots to the title screen, not straight into the world).
func NewFrontFlow() *FrontFlow {
	flow := &FrontFlow{screen: FrontTitle}
	flow.menu = flow.buildMenu()

	return flow
}

// SetSelHero installs the real single-player screen: HandleKey routes
// FrontSelHero keys to it instead of the placeholder walk-back.
func (f *FrontFlow) SetSelHero(s ui.Screen) { f.selhero = s }

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
		if f.selhero != nil {
			f.selhero.HandleKey(key)

			return
		}

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
			f.screen = FrontSelHero
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

// handleSelHeroKey applies the placeholder's contract: Escape walks back to
// the main menu, every other key leaves the placeholder unchanged (the real
// selhero screen is devilGo-5fr4).
func (f *FrontFlow) handleSelHeroKey(key ui.Key) {
	if key == ui.KeyEscape {
		f.walkBack()
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
