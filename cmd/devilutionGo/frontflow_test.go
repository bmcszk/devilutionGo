//go:build windowed

package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// TestFrontBootLandsOnTitle pins the boot screen (devilGo-uuyh): the windowed
// boot starts on the title screen — today it lands on the town, which is the
// RED this slice fixes.
func TestFrontBootLandsOnTitle(t *testing.T) {
	// given the windowed boot's front-screen wiring
	boot := newFrontBoot()

	// then the boot screen is the title, not the town
	if got := boot.flow.Screen(); got != scene.FrontTitle {
		t.Fatalf("boot screen = %v, want FrontTitle", got)
	}
}

// TestFrontFlowTransitions drives the whole front-screen flow through key
// presses and asserts the active screen after each one — the black-box path
// the window's key mapping feeds (devilGo-uuyh ACs).
func TestFrontFlowTransitions(t *testing.T) {
	tests := []struct {
		name string
		keys []ui.Key
		want scene.FrontScreen
	}{
		{
			name: "title any key advances to the main menu",
			keys: []ui.Key{ui.KeySpace},
			want: scene.FrontMainMenu,
		},
		{
			name: "enter on Single Player reaches the selhero placeholder",
			keys: []ui.Key{ui.KeySpace, ui.KeyEnter},
			want: scene.FrontSelHero,
		},
		{
			name: "escape twice walks back to the title",
			keys: []ui.Key{ui.KeySpace, ui.KeyEnter, ui.KeyEscape, ui.KeyEscape},
			want: scene.FrontTitle,
		},
		{
			name: "escape on the selhero placeholder walks back to the main menu",
			keys: []ui.Key{ui.KeySpace, ui.KeyEnter, ui.KeyEscape},
			want: scene.FrontMainMenu,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given the flow the boot installs
			boot := newFrontBoot()

			// when the keys drive the flow one press per frame
			for _, key := range test.keys {
				boot.handleKey(key)
			}

			// then the active screen matches the walk
			if got := boot.flow.Screen(); got != test.want {
				t.Fatalf("after %v keys screen = %v, want %v", test.keys, got, test.want)
			}

			// and no walk requested a quit
			if boot.flow.Quit() {
				t.Fatal("flow quit on a navigation walk, want it active")
			}
		})
	}
}

// TestFrontFlowQuitPaths pins the rc-0 exits (devilGo-uuyh): Exit on the main
// menu and Escape on the title both end the boot; anything on the selhero
// placeholder never does (its Escape walks back one screen).
func TestFrontFlowQuitPaths(t *testing.T) {
	tests := []struct {
		name string
		keys []ui.Key
	}{
		{
			name: "escape on the title quits",
			keys: []ui.Key{ui.KeyEscape},
		},
		{
			name: "enter on Exit quits",
			keys: []ui.Key{ui.KeySpace, ui.KeyArrowDown, ui.KeyEnter},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given the flow the boot installs
			boot := newFrontBoot()

			// when the quit path's keys drive it
			for _, key := range test.keys {
				boot.handleKey(key)
			}

			// then the flow asked to end the boot
			if !boot.flow.Quit() {
				t.Fatalf("after %v keys flow still active, want quit", test.keys)
			}
		})
	}
}

// TestFrontFlowSelHeroPlaceholderNeverQuits pins the placeholder's contract
// (devilGo-uuyh): it renders and accepts back-navigation, nothing more — every
// key on it either stays put or walks back one screen.
func TestFrontFlowSelHeroPlaceholderNeverQuits(t *testing.T) {
	// given the flow parked on the placeholder
	boot := newFrontBoot()
	for _, key := range []ui.Key{ui.KeySpace, ui.KeyEnter} {
		boot.handleKey(key)
	}

	// when every navigation key the placeholder accepts is pressed
	for _, key := range []ui.Key{
		ui.KeyEnter, ui.KeySpace, ui.KeyArrowUp, ui.KeyArrowDown,
		ui.KeyArrowLeft, ui.KeyArrowRight, ui.KeyTab, ui.KeyShiftTab,
		ui.KeyEscape,
	} {
		boot.handleKey(key)

		// then the flow never quit: the placeholder carries no exit
		if boot.flow.Quit() {
			t.Fatalf("flow quit on placeholder key %v, want back-navigation only", key)
		}

		if got := boot.flow.Screen(); got != scene.FrontMainMenu && got != scene.FrontSelHero {
			t.Fatalf("placeholder key %v landed on %v, want SelHero or MainMenu", key, got)
		}
	}
}

// TestFrontTitleIgnoresIdleTicks pins the real-key gate (review F1): the
// window seam feeds KeyUnknown on every idle frame (window_ebiten.go
// windowFlowKey), so the title's press-any-key rule must accept only keys
// carrying a concrete code — idle ticks never advance the title.
func TestFrontTitleIgnoresIdleTicks(t *testing.T) {
	// given the flow parked on the title
	boot := newFrontBoot()

	// when a full second of idle frames feeds the flow nothing but KeyUnknown
	for i := 0; i < 20; i++ {
		boot.handleKey(ui.KeyUnknown)
	}

	// then the title holds: no key was pressed
	if got := boot.flow.Screen(); got != scene.FrontTitle {
		t.Fatalf("after 20 idle ticks screen = %v, want FrontTitle", got)
	}

	// when a concrete key arrives
	boot.handleKey(ui.KeySpace)

	// then the title advances: a real press opens the main menu
	if got := boot.flow.Screen(); got != scene.FrontMainMenu {
		t.Fatalf("after a concrete key screen = %v, want FrontMainMenu", got)
	}
}

// TestFrontFlowEscapeWalksBackNotQuit pins the back-walk seam (review F2):
// Escape on the menu and on the selhero placeholder walks BACK one screen and
// must not trip the legacy quit overlay the window checks alongside the flow —
// only Escape-on-title and Exit quit rc 0.
func TestFrontFlowEscapeWalksBackNotQuit(t *testing.T) {
	tests := []struct {
		name string
		keys []ui.Key
		want scene.FrontScreen
	}{
		{
			name: "escape on the menu walks back to the title",
			keys: []ui.Key{ui.KeySpace, ui.KeyEscape},
			want: scene.FrontTitle,
		},
		{
			name: "escape on the selhero placeholder walks back to the menu",
			keys: []ui.Key{ui.KeySpace, ui.KeyEnter, ui.KeyEscape},
			want: scene.FrontMainMenu,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given the flow the boot installs
			boot := newFrontBoot()

			// when the walk's keys drive it
			for _, key := range test.keys {
				boot.handleKey(key)
			}

			// then the walk happened without a quit request: the legacy
			// level-triggered quit overlay is gone, so only the flow's own
			// quit paths (Escape-on-title, Exit) can end the boot
			if got := boot.flow.Screen(); got != test.want {
				t.Fatalf("after %v keys screen = %v, want %v", test.keys, got, test.want)
			}
			if boot.flow.Quit() {
				t.Fatalf("after %v keys flow quit, want walk-back", test.keys)
			}
		})
	}
}
