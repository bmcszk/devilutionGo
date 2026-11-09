package ui

// The flask module ports the life/mana flasks of Source/control/control_flasks.cpp:
// the fixed cel regions, the discrete 0..81 fill and the four draw entry points
// the frame calls (scrollrt.cpp:1453-1454, :1876-1879).

// FlaskFillMax is the top of the discrete fill range: _pHPPer and _pManaPer are
// percentages on a 0..81 scale (player.h:691, :702).
const FlaskFillMax = 81

// The flask x offsets inside the main panel geometry
// (control_flasks.cpp:103-121).
const (
	LifeFlaskUpperOffset = 107
	ManaFlaskUpperOffset = 475
	LifeFlaskLowerOffset = 96
	ManaFlaskLowerOffset = 464
)

// The two flask halves.
const (
	FlaskUpper FlaskHalf = iota
	FlaskLower
)

// The two flasks.
const (
	FlaskLife Flask = iota
	FlaskMana
)

// The two fill-visibility states.
const (
	FillHidden Fill = iota
	FillVisible
)

// The flask cel regions (control_flasks.cpp:14-15).
var (
	// FlaskTopRect is the dome that protrudes above the panel's top line.
	FlaskTopRect = Rect{Position: Point{X: 11, Y: 3}, Size: Size{Width: 62, Height: 13}}
	// FlaskBottomRect is the bulb that sits on the panel.
	FlaskBottomRect = Rect{Position: Point{X: 0, Y: 16}, Size: Size{Width: 88, Height: 69}}
)

// FlaskHalf names the upper (dome, above the panel) or lower (bulb, on the
// panel) half of a flask.
type FlaskHalf uint8

// Flask names one of the two player flasks.
type Flask uint8

// Fill is the fill visibility scrollrt passes as drawFilledPortion: a
// composited control panel already carries the filled rows, so it draws only the
// empty part and skips the second blit (control_flasks.cpp:93,
// scrollrt.cpp:1876).
type Fill uint8

// FlaskFillRows is one half's row split: the rows that stay empty and the rows
// the filled part covers, both measured in the half's own cel region.
type FlaskFillRows struct {
	EmptyRows  int
	FilledRows int
}

// FlaskFill computes the row split of one flask half for one discrete fill
// percentage. Upstream keeps the two halves in integer rows with no float math
// (control_flasks.cpp:40-42, :81-83).
func FlaskFill(half FlaskHalf, fillPer int) FlaskFillRows {
	if half == FlaskUpper {
		return flaskUpperRows(fillPer)
	}

	return flaskLowerRows(fillPer)
}

// DrawFlaskUpper emits the dome: the whole empty cel keyed on index 0, then the
// filled rows taken from the composed panel buffer (DrawFlaskUpper,
// control_flasks.cpp:38-56).
func DrawFlaskUpper(sink PanelSink, areas Areas, flask Flask, fillPer int) {
	rect := FlaskTopRect
	offset := flaskUpperOffset(flask)
	rows := flaskUpperRows(fillPer)

	sink.DrawPanelBox(PanelBox{
		Buffer:  BufferFlask,
		Source:  rect,
		Target:  Point{X: areas.Main.Position.X + offset, Y: areas.Main.Position.Y - rect.Size.Height},
		Flavour: BlitSkipZero,
	})

	if rows.FilledRows <= 0 {
		return
	}

	sink.DrawPanelBox(PanelBox{
		Buffer: BufferPanel,
		Source: Rect{
			Position: Point{X: offset, Y: rect.Position.Y + rows.EmptyRows},
			Size:     Size{Width: rect.Size.Width, Height: rows.FilledRows},
		},
		Target: Point{
			X: areas.Main.Position.X + offset,
			Y: areas.Main.Position.Y - rect.Size.Height + rows.EmptyRows,
		},
		Flavour: BlitCopy,
	})
}

// DrawFlaskLower emits the bulb: the empty rows from the flask cel, then the
// filled rows from the composed panel buffer (DrawFlaskLower,
// control_flasks.cpp:79-98).
func DrawFlaskLower(sink PanelSink, areas Areas, flask Flask, fillPer int, fill Fill) {
	rect := FlaskBottomRect
	offset := flaskLowerOffset(flask)
	rows := flaskLowerRows(fillPer)

	if rows.EmptyRows > 0 {
		sink.DrawPanelBox(PanelBox{
			Buffer: BufferFlask,
			Source: Rect{
				Position: rect.Position,
				Size:     Size{Width: rect.Size.Width, Height: rows.EmptyRows},
			},
			Target:  Point{X: areas.Main.Position.X + offset, Y: areas.Main.Position.Y},
			Flavour: BlitCopy,
		})
	}

	if fill == FillHidden || rows.FilledRows <= 0 {
		return
	}

	sink.DrawPanelBox(PanelBox{
		Buffer: BufferPanel,
		Source: Rect{
			Position: Point{X: offset, Y: rect.Position.Y + rows.EmptyRows},
			Size:     Size{Width: rect.Size.Width, Height: rows.FilledRows},
		},
		Target:  Point{X: areas.Main.Position.X + offset, Y: areas.Main.Position.Y + rows.EmptyRows},
		Flavour: BlitCopy,
	})
}

// flaskUpperRows is the dome split: 81-fillPer empty rows clamped into the dome
// height, the rest filled (control_flasks.cpp:41-42).
func flaskUpperRows(fillPer int) FlaskFillRows {
	empty := clamp(FlaskFillMax-fillPer, 0, FlaskTopRect.Size.Height)

	return FlaskFillRows{EmptyRows: empty, FilledRows: FlaskTopRect.Size.Height - empty}
}

// flaskLowerRows is the bulb split: fillPer filled rows clamped into the bulb
// height, the rest empty (control_flasks.cpp:82-83).
func flaskLowerRows(fillPer int) FlaskFillRows {
	filled := clamp(fillPer, 0, FlaskBottomRect.Size.Height)

	return FlaskFillRows{EmptyRows: FlaskBottomRect.Size.Height - filled, FilledRows: filled}
}

// flaskUpperOffset is the dome's x offset inside the main panel
// (control_flasks.cpp:103-113).
func flaskUpperOffset(flask Flask) int {
	if flask == FlaskMana {
		return ManaFlaskUpperOffset
	}

	return LifeFlaskUpperOffset
}

// flaskLowerOffset is the bulb's x offset inside the main panel
// (control_flasks.cpp:115-125).
func flaskLowerOffset(flask Flask) int {
	if flask == FlaskMana {
		return ManaFlaskLowerOffset
	}

	return LifeFlaskLowerOffset
}

// clamp bounds v to [lo, hi], the std::clamp the flask rows use.
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}

	if v > hi {
		return hi
	}

	return v
}
