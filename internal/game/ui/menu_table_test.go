package ui_test

import (
	"reflect"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
)

// speedRow is the Speed row's index in the single-player table
// (X:gamemenu.cpp:49-57, row 3).
const speedRow = 3

func TestMenuTableRowsInUpstreamOrder(t *testing.T) {
	// given: the single-player rows in upstream order (X:gamemenu.cpp:49-57
	// sgSingleMenu); every row starts GMENU_ENABLED.
	table := ui.NewMenuTable([]ui.MenuItem{
		{Label: "Options", Enabled: true},
		{Label: "Save Game", Enabled: true},
		{Label: "Load Game", Enabled: true},
		{Label: "Exit to Main Menu", Enabled: true},
		{Label: "Quit Game", Enabled: true},
	})

	// when: the rows are read back
	rows := table.Rows()

	// then: five enabled rows in the upstream order
	if len(rows) != 5 {
		t.Fatalf("Rows() len = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if !r.Enabled {
			t.Fatalf(
				"row %d (%q) Enabled = false, want true (every row starts GMENU_ENABLED)",
				i,
				r.Label,
			)
		}
	}
	if labels := rowLabels(table); !reflect.DeepEqual(
		labels,
		[]string{"Options", "Save Game", "Load Game", "Exit to Main Menu", "Quit Game"},
	) {
		t.Fatalf("rows = %v, want the upstream order (X:gamemenu.cpp:49-57)", labels)
	}
}

func TestMenuTableNextPrevSkipDisabledRowsAndWrap(t *testing.T) {
	rows := gameMenuRows()
	rows[2].Enabled = false // Load Game disabled: no save exists
	table := ui.NewMenuTable(rows)

	// when: the walks run around the disabled Load row
	nextFromSave := table.NextEnabled(1)
	prevFromExit := table.PrevEnabled(3)
	nextFromQuit := table.NextEnabled(4)
	prevFromOptions := table.PrevEnabled(0)
	nextFromLoad := table.NextEnabled(2)

	// then: disabled rows are skipped, with wrap in both directions
	if nextFromSave != 3 {
		t.Fatalf("NextEnabled(1) = %d, want 3 (skip the disabled Load row)", nextFromSave)
	}
	if prevFromExit != 1 {
		t.Fatalf("PrevEnabled(3) = %d, want 1 (skip back over the disabled row)", prevFromExit)
	}
	if nextFromQuit != 0 {
		t.Fatalf("NextEnabled(4) = %d, want 0 (wrap)", nextFromQuit)
	}
	if prevFromOptions != 4 {
		t.Fatalf("PrevEnabled(0) = %d, want 4 (wrap)", prevFromOptions)
	}
	if nextFromLoad != 3 {
		t.Fatalf("NextEnabled(2) = %d, want 3 (leave the disabled row)", nextFromLoad)
	}
}

func TestMenuTableFirstEnabled(t *testing.T) {
	// given: the first row disabled
	rows := gameMenuRows()
	rows[0].Enabled = false
	table := ui.NewMenuTable(rows)

	// when: the first enabled row is asked for
	first := table.FirstEnabled()

	// then: it is the second row
	if first != 1 {
		t.Fatalf(
			"FirstEnabled() = %d, want 1 (gmenu_set_items focuses the first enabled row)",
			first,
		)
	}
}

func TestMenuTableAllDisabledStaysPut(t *testing.T) {
	// given: a table with every row disabled
	rows := gameMenuRows()
	for i := range rows {
		rows[i].Enabled = false
	}
	table := ui.NewMenuTable(rows)

	// when: navigation is asked from a focused row
	next := table.NextEnabled(2)
	prev := table.PrevEnabled(2)
	first := table.FirstEnabled()

	// then: every walk stays put and no row reports enabled
	if next != 2 {
		t.Fatalf("NextEnabled(2) = %d, want 2 (no enabled row to move to)", next)
	}
	if prev != 2 {
		t.Fatalf("PrevEnabled(2) = %d, want 2", prev)
	}
	if first != -1 {
		t.Fatalf("FirstEnabled() = %d, want -1 (nothing enabled)", first)
	}
}

func TestMenuSliderRoundTripsThroughTheUpstreamArithmetic(t *testing.T) {
	// X:gmenu.cpp:375-392 with (min,max,steps) = (20,50,46): the round-trip
	// is exact at the 20/35/50 anchor points.
	for _, rate := range []int{20, 35, 50} {
		t.Run("", func(t *testing.T) {
			// given: a fresh 46-step slider
			table := speedTable(rate)

			// when: the rate is set onto the slider
			table.SliderSet(speedRow, 20, 50, rate)

			// then: the read round-trips exactly
			if got := table.SliderGet(speedRow, 20, 50); got != rate {
				t.Fatalf("SliderGet after SliderSet(%d) = %d, want exact round-trip", rate, got)
			}
		})
	}
}

func TestMenuSliderRightFromStepZeroReads21(t *testing.T) {
	// given: the slider at its bottom step
	table := speedTable(20)

	// when: one step right
	if got := table.SliderStepRight(speedRow); got != 1 {
		t.Fatalf("SliderStepRight = step %d, want 1", got)
	}

	// then: the value reads one quantized step up
	if got := table.SliderGet(speedRow, 20, 50); got != 21 {
		t.Fatalf("SliderGet after one Right = %d, want 21 (20 + 52/46)", got)
	}
}

func TestMenuSliderStepsSaturateAtBothEnds(t *testing.T) {
	// given: the slider at its top step
	table := speedTable(50)

	if got := table.SliderStepRight(speedRow); got != 46 {
		t.Fatalf("SliderStepRight at max = %d, want 46 (saturate)", got)
	}

	// when: a fresh slider at its bottom step, stepping left
	table = speedTable(20)

	// then: the step saturates at zero
	if got := table.SliderStepLeft(speedRow); got != 0 {
		t.Fatalf("SliderStepLeft at zero = %d, want 0 (saturate)", got)
	}
}

func TestMenuSliderStepsBelowTwoBehaveAsTwo(t *testing.T) {
	// X:gmenu.cpp:377,:386 clamp sliderSteps to min 2 at use.
	// given: a slider declaring only one step
	table := ui.NewMenuTable(gameMenuRows())
	table.SetSliderSteps(speedRow, 1)
	table.SliderSet(speedRow, 20, 50, 20)

	if got := table.SliderGet(speedRow, 20, 50); got != 20 {
		t.Fatalf("SliderGet with steps=1 (clamped to 2) at min = %d, want 20", got)
	}

	// when: the slider is set to the max rate
	table.SliderSet(speedRow, 20, 50, 50)

	// then: the two-step clamp still resolves it exactly
	if got := table.SliderGet(speedRow, 20, 50); got != 50 {
		t.Fatalf("SliderGet with steps=1 (clamped to 2) at max = %d, want 50", got)
	}
}

func TestMenuSliderMidRangeRoundTrips(t *testing.T) {
	// upstream twin: Slider_MidRange (X:test/game_menu_test.cpp:207-219):
	// a 100-step slider over 0..100 holds 75 exactly
	table := ui.NewMenuTable(gameMenuRows())
	table.SetSliderSteps(speedRow, 100)

	// when: 75 is set over [0,100] with 100 steps
	table.SliderSet(speedRow, 0, 100, 75)

	// then: it reads back exactly
	if got := table.SliderGet(speedRow, 0, 100); got != 75 {
		t.Fatalf("SliderGet after SliderSet(75) over [0,100]/100 steps = %d, want 75", got)
	}
}

func TestMenuSliderCustomRangeRoundTrips(t *testing.T) {
	// upstream twin: Slider_CustomRange (X:test/game_menu_test.cpp:221-233):
	// the bounds are parameters, not the Speed constants — [10,60] over 50
	// steps holds 35 exactly (a port that hard-coded (20,50) fails here)
	table := ui.NewMenuTable(gameMenuRows())
	table.SetSliderSteps(speedRow, 50)

	// when: 35 is set over the custom bounds [10,60] with 50 steps
	table.SliderSet(speedRow, 10, 60, 35)

	// then: it reads back exactly
	if got := table.SliderGet(speedRow, 10, 60); got != 35 {
		t.Fatalf("SliderGet after SliderSet(35) over [10,60]/50 steps = %d, want 35", got)
	}
}

// gameMenuRows are the single-player menu rows in upstream order
// (X:gamemenu.cpp:49-57 sgSingleMenu); every row starts GMENU_ENABLED.
func gameMenuRows() []ui.MenuItem {
	return []ui.MenuItem{
		{Label: "Options", Enabled: true},
		{Label: "Save Game", Enabled: true},
		{Label: "Load Game", Enabled: true},
		{Label: "Exit to Main Menu", Enabled: true},
		{Label: "Quit Game", Enabled: true},
	}
}

// rowLabels returns the table's row labels.
func rowLabels(table *ui.MenuTable) []string {
	rows := table.Rows()
	labels := make([]string, len(rows))
	for i, r := range rows {
		labels[i] = r.Label
	}

	return labels
}

// speedTable builds the game menu with the Speed row's 46-step slider set
// to the given tick rate (X:gamemenu.cpp:158-178).
func speedTable(rate int) *ui.MenuTable {
	table := ui.NewMenuTable(gameMenuRows())
	table.SetSliderSteps(speedRow, 46)
	table.SliderSet(speedRow, 20, 50, rate)

	return table
}
