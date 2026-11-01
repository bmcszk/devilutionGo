package stores

// The dialog panel model: the 104 fixed text lines upstream draws into, the
// helpers that write them, and the layout constants.

// NumStoreLines is NumStoreLines (stores.h:39): the fixed panel line count.
const NumStoreLines = 104

// BackButtonLine is BackButtonLine (stores.cpp:233-239) for the latin small
// font: line 22 regardless of the scrollbar. The CJK branch (20/21) keys on
// IsSmallFontTall(), which is a font-metrics question this data port does not
// own; see the deviation note in the commit report.
const BackButtonLine = 22

// The line kinds (stores.cpp:154-158).
const (
	// LineLabel is a static text line.
	LineLabel LineKind = iota
	// LineDivider is a horizontal rule (AddSLine).
	LineDivider
	// LineSelectable is a clickable option or list entry.
	LineSelectable
)

// ScrollRepeatFrames is the rearm value of both scrollbar counters
// (stores.cpp:2806, 2814).
const ScrollRepeatFrames int8 = 10

// NoneScrollCountdown is the idle counter value, -1 (stores.cpp:2851-2852).
const NoneScrollCountdown int8 = -1

// NoneTextLine is CurrentTextLine's empty value, -1 (stores.cpp:2560).
const NoneTextLine = -1

// PageRows is the page-jump step of StorePrior/StoreNext, four rows
// (stores.cpp:2636, 2648).
const PageRows = 4

// The alignments the store dialogs set.
const (
	AlignDefault Align = iota
	AlignCenter
	AlignRight
)

// The colors the store dialogs set.
const (
	ColorDefault Color = iota
	ColorWhite
	ColorBlue
	ColorWhitegold
	ColorRed
)

// LineKind is STextStruct::Type (stores.cpp:154-158).
// ScrollRepeat is CountdownScrollUp and CountdownScrollDown (stores.cpp:196-199):
// the push state of the two scrollbar buttons. A counter is -1 when idle and
// 0..10 while the button is held, so the panel draws it pushed for any value
// but -1 (DrawSSlider, stores.cpp:282-289).
type ScrollRepeat struct {
	// Up is CountdownScrollUp.
	Up int8
	// Down is CountdownScrollDown.
	Down int8
}

type LineKind uint8

// Align is the horizontal alignment subset of UiFlags the store dialogs use
// (DiabloUI/ui_flags.hpp).
type Align uint8

// Color is the text-color subset of UiFlags the store dialogs use
// (DiabloUI/ui_flags.hpp).
type Color uint8

// Line is STextStruct (stores.cpp:154-171): one panel line.
type Line struct {
	// Text is the rendered string; empty for dividers.
	Text string
	// Kind is Label, Divider or Selectable.
	Kind LineKind
	// Align is the horizontal alignment.
	Align Align
	// Color is the text color.
	Color Color
	// X is _sx, the left offset in pixels.
	X uint8
	// YOff is _syoff, the vertical nudge the back button uses.
	YOff uint8
	// SVal is _sval, the gold value AddSTextVal attaches to a line.
	SVal int
	// CursID is cursId, or -1.
	CursID int
	// CursIndent is cursIndent.
	CursIndent bool
}

// Dialog is the store panel: TextLine[NumStoreLines] plus the scroll and size
// flags that travel with it (stores.cpp:178-197).
type Dialog struct {
	// Lines is TextLine.
	Lines [NumStoreLines]Line
	// HasScrollbar is HasScrollbar.
	HasScrollbar bool
	// IsTextFullSize is IsTextFullSize.
	IsTextFullSize bool
	// RenderGold is RenderGold: whether the panel shows the player's gold.
	RenderGold bool
	// NumTextLines is NumTextLines, the scrollable entry count.
	NumTextLines int
	// ScrollPos is ScrollPos.
	ScrollPos int
	// CurrentTextLine is CurrentTextLine, or -1 when nothing is selectable.
	CurrentTextLine int
	// Scroll is the scrollbar's push-repeat state, the two countdown counters
	// stores.cpp keeps beside the cursor (CountdownScrollUp/Down, :196-199).
	Scroll ScrollRepeat
	// PreviousScrollPos is PreviousScrollPos, the page's first row line (:413).
	PreviousScrollPos int
	// NextScrollPos is NextScrollPos, the page's last row line (:421).
	NextScrollPos int
}

// MoveUp ports StoreUp (stores.cpp:2557-2591): with a scrollbar the cursor
// walks up to the page's first row and then pages the list, without one it
// wraps around the panel. Both branches skip the non-selectable lines.
func (d *Dialog) MoveUp() {
	if d.CurrentTextLine == NoneTextLine {
		return
	}

	if d.HasScrollbar && d.CurrentTextLine == d.PreviousScrollPos {
		if d.ScrollPos != 0 {
			d.ScrollPos--
		}

		return
	}

	d.CurrentTextLine = d.previousSelectable(d.CurrentTextLine)
}

// MoveDown ports StoreDown (stores.cpp:2594-2628), the mirror of MoveUp.
func (d *Dialog) MoveDown() {
	if d.CurrentTextLine == NoneTextLine {
		return
	}

	if d.HasScrollbar && d.CurrentTextLine == d.NextScrollPos {
		if d.ScrollPos < d.NumTextLines {
			d.ScrollPos++
		}

		return
	}

	d.CurrentTextLine = d.nextSelectable(d.CurrentTextLine)
}

// MovePrior ports StorePrior (stores.cpp:2630-2641): with a scrollbar, a key
// press on the page's first row pages the list up by four rows (stopping at
// zero), anywhere else it jumps the cursor back to that first row.
func (d *Dialog) MovePrior() {
	if d.CurrentTextLine == NoneTextLine || !d.HasScrollbar {
		return
	}

	if d.CurrentTextLine == d.PreviousScrollPos {
		d.ScrollPos = max(d.ScrollPos-PageRows, 0)

		return
	}

	d.CurrentTextLine = d.PreviousScrollPos
}

// MoveNext ports StoreNext (stores.cpp:2643-2655), the mirror of MovePrior:
// the list pages down by four rows and never scrolls past NumTextLines.
func (d *Dialog) MoveNext() {
	if d.CurrentTextLine == NoneTextLine || !d.HasScrollbar {
		return
	}

	if d.CurrentTextLine == d.NextScrollPos {
		d.ScrollPos += PageRows
		if d.ScrollPos > d.NumTextLines {
			d.ScrollPos = d.NumTextLines
		}

		return
	}

	d.CurrentTextLine = d.NextScrollPos
}

// NewScrollRepeat returns the counters at their idle value (stores.cpp:2851).
func NewScrollRepeat() ScrollRepeat {
	return ScrollRepeat{Up: NoneScrollCountdown, Down: NoneScrollCountdown}
}

// Release ports ReleaseStoreBtn (stores.cpp:2848-2852): letting go of the button
// idles both counters.
func (r *ScrollRepeat) Release() {
	r.Up = NoneScrollCountdown
	r.Down = NoneScrollCountdown
}

// PressUp is the scrollbar-up branch of the store click handler
// (stores.cpp:2803-2810): a spent counter steps the cursor now and rearms to
// ten, any other value just counts down. It reports whether the step fires.
func (r *ScrollRepeat) PressUp() bool {
	if r.Up <= 0 {
		r.Up = ScrollRepeatFrames

		return true
	}

	r.Up--

	return false
}

// PressDown is the scrollbar-down branch (stores.cpp:2811-2818).
func (r *ScrollRepeat) PressDown() bool {
	if r.Down <= 0 {
		r.Down = ScrollRepeatFrames

		return true
	}

	r.Down--

	return false
}

// IsPushedUp reports whether the up button draws pushed (stores.cpp:282).
func (r *ScrollRepeat) IsPushedUp() bool { return r.Up != NoneScrollCountdown }

// IsPushedDown reports whether the down button draws pushed (stores.cpp:286).
func (r *ScrollRepeat) IsPushedDown() bool { return r.Down != NoneScrollCountdown }

// IsDivider reports whether the line is a horizontal rule (stores.cpp:164-167).
func (l Line) IsDivider() bool { return l.Kind == LineDivider }

// IsSelectable reports whether the line can be chosen (stores.cpp:169-171).
func (l Line) IsSelectable() bool { return l.Kind == LineSelectable }

// HasText reports whether the line carries a rendered string (stores.cpp:173-176).
func (l Line) HasText() bool { return l.Text != "" }

// newDialog returns an empty dialog with every cursor id at upstream's -1
// default (AddSText's cursId default, stores.cpp:322).
func newDialog() Dialog {
	var d Dialog

	for i := range d.Lines {
		d.Lines[i].CursID = -1
	}

	return d
}

// ClearSText resets a line span to empty labels (stores.cpp:2244-2270).
func (d *Dialog) ClearSText(from, to int) {
	for i := from; i < to && i < NumStoreLines; i++ {
		d.Lines[i] = Line{CursID: -1}
	}
}

// AddSText writes one non-selectable text line (stores.cpp:322-333).
func (d *Dialog) AddSText(x uint8, y int, text string, color Color, align Align) {
	d.writeSText(x, y, text, color, align, LineLabel)
}

// AddSelectableSText writes one selectable text line (stores.cpp:322-333).
func (d *Dialog) AddSelectableSText(x uint8, y int, text string, color Color, align Align) {
	d.writeSText(x, y, text, color, align, LineSelectable)
}

// AddSTextVal attaches a gold value to a line (stores.cpp:317-320).
func (d *Dialog) AddSTextVal(y, val int) {
	d.Lines[y].SVal = val
}

// AddSLine writes a divider (stores.cpp:306-315).
func (d *Dialog) AddSLine(y int) {
	d.Lines[y] = Line{Kind: LineDivider, CursID: -1}
}

// AddOptionsBackButton writes the menu "Back" button (stores.cpp:334-340).
func (d *Dialog) AddOptionsBackButton() {
	d.AddSelectableSText(0, BackButtonLine, "Back", ColorWhite, AlignCenter)
	d.Lines[BackButtonLine].YOff = 6
}

// AddItemListBackButton is AddItemListBackButton(selectable=false)
// (stores.cpp:341-353) for the latin font branch.
func (d *Dialog) AddItemListBackButton() {
	d.writeItemListBackButton(LineLabel)
}

// AddSelectableItemListBackButton is AddItemListBackButton(selectable=true)
// (stores.cpp:341-353, stores.cpp:911 "nothing to recharge" uses it).
func (d *Dialog) AddSelectableItemListBackButton() {
	d.writeItemListBackButton(LineSelectable)
}

// FirstSelectable returns the first selectable line, or -1
// (stores.cpp:2422-2430).
func (d *Dialog) FirstSelectable() int {
	for i := range d.Lines {
		if d.Lines[i].IsSelectable() {
			return i
		}
	}

	return -1
}

// reset clears the panel for a new page: StartStore's ClearSText(0,
// NumStoreLines) plus the flag resets it performs (stores.cpp:2287-2297).
// previousSelectable steps one line up, wrapping at the top, and keeps stepping
// until it lands on a selectable line (stores.cpp:2571-2582). Upstream loops
// forever when a page carries no selectable line; every builder here adds the
// back button, so the loop always terminates.
func (d *Dialog) previousSelectable(line int) int {
	for {
		if line == 0 {
			line = NumStoreLines - 1
		} else {
			line--
		}

		if d.Lines[line].IsSelectable() {
			return line
		}
	}
}

// nextSelectable is previousSelectable's downward mirror
// (stores.cpp:2608-2619).
func (d *Dialog) nextSelectable(line int) int {
	for {
		if line == NumStoreLines-1 {
			line = 0
		} else {
			line++
		}

		if d.Lines[line].IsSelectable() {
			return line
		}
	}
}

func (d *Dialog) reset() {
	*d = newDialog()
}

// writeItemListBackButton writes the list "Back" button: a divider on the line
// above plus a centered label on BackButtonLine (stores.cpp:341-353).
func (d *Dialog) writeItemListBackButton(kind LineKind) {
	d.AddSLine(BackButtonLine - 1)
	d.writeSText(0, BackButtonLine, "Back", ColorWhite, AlignCenter, kind)
	d.Lines[BackButtonLine].YOff = 6
}

// writeSText writes one text line of the given kind (stores.cpp:322-333).
func (d *Dialog) writeSText(x uint8, y int, text string, color Color, align Align, kind LineKind) {
	d.Lines[y] = Line{
		Text:   text,
		Kind:   kind,
		Align:  align,
		Color:  color,
		X:      x,
		CursID: -1,
	}
}
