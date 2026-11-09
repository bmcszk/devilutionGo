// Package ui_test proves the life and mana flask fill (Source/control/control_flasks.cpp).
package ui_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// The two palette indexes the fill-growth test paints with: the dome art and the
// filled rows the composed panel buffer carries.
const (
	flaskDomeIndex = 5
	flaskFillIndex = 7
)

func TestFlaskFillRowsFollowTheIntegerFormula(t *testing.T) {
	// given: the two flask regions (control_flasks.cpp:14-15) and the discrete
	// 0..81 fill the player struct maintains (player.h:691, :702)
	top := ui.FlaskTopRect
	bottom := ui.FlaskBottomRect

	for _, row := range []struct {
		fillPer  int
		topFill  int
		lowFill  int
		lowEmpty int
	}{
		{fillPer: 0, topFill: 0, lowFill: 0, lowEmpty: bottom.Size.Height},
		{fillPer: 1, topFill: 0, lowFill: 1, lowEmpty: bottom.Size.Height - 1},
		{fillPer: 2, topFill: 0, lowFill: 2, lowEmpty: bottom.Size.Height - 2},
		{fillPer: 67, topFill: 0, lowFill: 67, lowEmpty: bottom.Size.Height - 67},
		{fillPer: 68, topFill: 0, lowFill: 68, lowEmpty: bottom.Size.Height - 68},
		{fillPer: 69, topFill: 1, lowFill: bottom.Size.Height, lowEmpty: 0},
		{fillPer: 80, topFill: top.Size.Height - 1, lowFill: bottom.Size.Height, lowEmpty: 0},
		{fillPer: 81, topFill: top.Size.Height, lowFill: bottom.Size.Height, lowEmpty: 0},
	} {
		// when: the upper flask clamps 81-fillPer into its 13 rows and the lower
		// flask clamps fillPer into its 69 rows (control_flasks.cpp:41-42, :82-83)
		upper := ui.FlaskFill(ui.FlaskUpper, row.fillPer)
		lower := ui.FlaskFill(ui.FlaskLower, row.fillPer)

		// then
		if upper.FilledRows != row.topFill || lower.FilledRows != row.lowFill {
			t.Fatalf("fillPer %d: upper filled=%d lower filled=%d, want %d and %d",
				row.fillPer, upper.FilledRows, lower.FilledRows, row.topFill, row.lowFill)
		}

		if lower.EmptyRows != row.lowEmpty {
			t.Fatalf("fillPer %d: lower empty=%d, want %d", row.fillPer, lower.EmptyRows, row.lowEmpty)
		}
	}
}

func TestFlaskFillClampsOutOfRangeFillPer(t *testing.T) {
	// when: upstream clamps both ends, so a negative or over-full percentage
	// lands on the empty or the full flask (control_flasks.cpp:41, :82)
	negUpper := ui.FlaskFill(ui.FlaskUpper, -5)
	overUpper := ui.FlaskFill(ui.FlaskUpper, 200)
	negLower := ui.FlaskFill(ui.FlaskLower, -5)
	overLower := ui.FlaskFill(ui.FlaskLower, 200)

	// then
	if negUpper.FilledRows != 0 || negLower.FilledRows != 0 {
		t.Fatalf("negative fill: upper=%d lower=%d, want 0 and 0", negUpper.FilledRows, negLower.FilledRows)
	}

	if overUpper.FilledRows != ui.FlaskTopRect.Size.Height {
		t.Fatalf("over-full upper = %d, want %d", overUpper.FilledRows, ui.FlaskTopRect.Size.Height)
	}

	if overLower.FilledRows != ui.FlaskBottomRect.Size.Height {
		t.Fatalf("over-full lower = %d, want %d", overLower.FilledRows, ui.FlaskBottomRect.Size.Height)
	}
}

func TestDrawFlaskUpperAtRestCopiesTheWholeDome(t *testing.T) {
	// given: an empty flask draws only the dome cel, keyed on index 0 so the
	// panel behind it shows through (control_flasks.cpp:31-34, :43-46)
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when
	ui.DrawFlaskUpper(sink, areas, ui.FlaskLife, 0)

	// then
	if len(sink.boxes) != 1 {
		t.Fatalf("empty upper flask emitted %d boxes, want 1", len(sink.boxes))
	}

	expectRect(t, "dome source", sink.boxes[0].Source, ui.FlaskTopRect)
	expectPoint(t, "dome target", sink.boxes[0].Target, ui.Point{
		X: areas.Main.Position.X + ui.LifeFlaskUpperOffset,
		Y: areas.Main.Position.Y - ui.FlaskTopRect.Size.Height,
	})

	if sink.boxes[0].Flavour != ui.BlitSkipZero {
		t.Fatalf("dome flavour = %v, want BlitSkipZero", sink.boxes[0].Flavour)
	}
}

func TestDrawFlaskUpperAddsFilledRows(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when: a full flask paints the filled part of the dome from the panel below
	// (control_flasks.cpp:50-54)
	ui.DrawFlaskUpper(sink, areas, ui.FlaskLife, 81)

	// then
	if len(sink.boxes) != 2 {
		t.Fatalf("full upper flask emitted %d boxes, want 2", len(sink.boxes))
	}

	expectRect(t, "filled source", sink.boxes[1].Source, ui.Rect{
		Position: ui.Point{X: ui.LifeFlaskUpperOffset, Y: ui.FlaskTopRect.Position.Y},
		Size:     ui.Size{Width: ui.FlaskTopRect.Size.Width, Height: ui.FlaskTopRect.Size.Height},
	})
	expectPoint(t, "filled target", sink.boxes[1].Target, ui.Point{
		X: areas.Main.Position.X + ui.LifeFlaskUpperOffset,
		Y: areas.Main.Position.Y - ui.FlaskTopRect.Size.Height,
	})
}

func TestDrawFlaskLowerAtRestCopiesTheEmptyBulb(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when: an empty lower flask copies the whole bulb, unkeyed, and draws no
	// filled portion (control_flasks.cpp:85-91)
	ui.DrawFlaskLower(sink, areas, ui.FlaskMana, 0, ui.FillVisible)

	// then
	if len(sink.boxes) != 1 {
		t.Fatalf("empty lower flask emitted %d boxes, want 1", len(sink.boxes))
	}

	expectRect(t, "bulb source", sink.boxes[0].Source, ui.Rect{
		Position: ui.FlaskBottomRect.Position,
		Size:     ui.Size{Width: ui.FlaskBottomRect.Size.Width, Height: ui.FlaskBottomRect.Size.Height},
	})
	expectPoint(t, "bulb target", sink.boxes[0].Target, ui.Point{
		X: areas.Main.Position.X + ui.ManaFlaskLowerOffset,
		Y: areas.Main.Position.Y,
	})

	if sink.boxes[0].Flavour != ui.BlitCopy {
		t.Fatalf("bulb flavour = %v, want BlitCopy", sink.boxes[0].Flavour)
	}
}

func TestDrawFlaskLowerDrawsTheFilledPortionFromThePanel(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when: half a flask leaves the empty rows above the fill
	// (control_flasks.cpp:84-97)
	ui.DrawFlaskLower(sink, areas, ui.FlaskLife, 34, ui.FillVisible)

	// then
	if len(sink.boxes) != 2 {
		t.Fatalf("half lower flask emitted %d boxes, want 2", len(sink.boxes))
	}

	emptyRows := ui.FlaskBottomRect.Size.Height - 34

	expectRect(t, "empty source", sink.boxes[0].Source, ui.Rect{
		Position: ui.FlaskBottomRect.Position,
		Size:     ui.Size{Width: ui.FlaskBottomRect.Size.Width, Height: emptyRows},
	})
	expectRect(t, "filled source", sink.boxes[1].Source, ui.Rect{
		Position: ui.Point{X: ui.LifeFlaskLowerOffset, Y: ui.FlaskBottomRect.Position.Y + emptyRows},
		Size:     ui.Size{Width: ui.FlaskBottomRect.Size.Width, Height: 34},
	})
	expectPoint(t, "filled target", sink.boxes[1].Target, ui.Point{
		X: areas.Main.Position.X + ui.LifeFlaskLowerOffset,
		Y: areas.Main.Position.Y + emptyRows,
	})
}

func TestDrawFlaskLowerSkipsFillWhenThePanelIsComposited(t *testing.T) {
	// given: scrollrt passes !drawCtrlPan, so a composited panel draws only the
	// empty part and the fill comes from the panel buffer itself
	// (scrollrt.cpp:1876, control_flasks.cpp:93)
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when: a half-full flask still draws its empty rows on the panel
	ui.DrawFlaskLower(sink, areas, ui.FlaskLife, 34, ui.FillHidden)

	// then
	if len(sink.boxes) != 1 {
		t.Fatalf("composited lower flask emitted %d boxes, want 1", len(sink.boxes))
	}
}

func TestFlaskPaintingShowsFillGrowthInTheIndexBuffer(t *testing.T) {
	// given: a flask dome cel of index 5 over a panel of index 9, plus the panel
	// buffer the filled rows come from at index 7
	sources := map[ui.PanelBuffer]render.Surface{
		ui.BufferPanel: solidSurface(ui.MainPanelWidth, ui.MainPanelHeight+ui.PanelPaddingHeight, flaskFillIndex),
		ui.BufferFlask: solidSurface(ui.FlaskBufferSize, ui.FlaskBufferSize, flaskDomeIndex),
	}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	dome := ui.Rect{
		Position: ui.Point{
			X: areas.Main.Position.X + ui.LifeFlaskUpperOffset,
			Y: areas.Main.Position.Y - ui.FlaskTopRect.Size.Height,
		},
		Size: ui.FlaskTopRect.Size,
	}

	// when: an empty dome and a full dome, each painted into its own surface
	emptySink := &recordSink{}
	ui.DrawFlaskUpper(emptySink, areas, ui.FlaskLife, 0)
	emptyOut := paintBoxes(t, emptySink.boxes, sources)

	fullSink := &recordSink{}
	ui.DrawFlaskUpper(fullSink, areas, ui.FlaskLife, 81)
	fullOut := paintBoxes(t, fullSink.boxes, sources)

	// then: the empty dome is pure dome art, the full one is pure fill art
	expectIndexes(t, emptyOut, dome, flaskDomeIndex)
	expectIndexes(t, fullOut, dome, flaskFillIndex)
}
