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

// LineKind is STextStruct::Type (stores.cpp:154-158).
type LineKind uint8

// The line kinds (stores.cpp:154-158).
const (
	// LineLabel is a static text line.
	LineLabel LineKind = iota
	// LineDivider is a horizontal rule (AddSLine).
	LineDivider
	// LineSelectable is a clickable option or list entry.
	LineSelectable
)

// Align is the horizontal alignment subset of UiFlags the store dialogs use
// (DiabloUI/ui_flags.hpp).
type Align uint8

// The alignments the store dialogs set.
const (
	AlignDefault Align = iota
	AlignCenter
	AlignRight
)

// Color is the text-color subset of UiFlags the store dialogs use
// (DiabloUI/ui_flags.hpp).
type Color uint8

// The colors the store dialogs set.
const (
	ColorDefault Color = iota
	ColorWhite
	ColorBlue
	ColorWhitegold
	ColorRed
)

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

// IsDivider reports whether the line is a horizontal rule (stores.cpp:164-167).
func (l Line) IsDivider() bool { return l.Kind == LineDivider }

// IsSelectable reports whether the line can be chosen (stores.cpp:169-171).
func (l Line) IsSelectable() bool { return l.Kind == LineSelectable }

// HasText reports whether the line carries a rendered string (stores.cpp:173-176).
func (l Line) HasText() bool { return l.Text != "" }

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
}

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

// AddSText writes one text line (stores.cpp:322-333).
func (d *Dialog) AddSText(x uint8, y int, text string, color Color, align Align, sel bool) {
	kind := LineLabel
	if sel {
		kind = LineSelectable
	}

	d.Lines[y] = Line{
		Text:   text,
		Kind:   kind,
		Align:  align,
		Color:  color,
		X:      x,
		CursID: -1,
	}
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
	d.AddSText(0, BackButtonLine, "Back", ColorWhite, AlignCenter, true)
	d.Lines[BackButtonLine].YOff = 6
}

// AddItemListBackButton is AddItemListBackButton(selectable=false)
// (stores.cpp:341-353) for the latin font branch.
func (d *Dialog) AddItemListBackButton() {
	d.addItemListBackButton(false)
}

// AddSelectableItemListBackButton is AddItemListBackButton(selectable=true)
// (stores.cpp:341-353, stores.cpp:911 "nothing to recharge" uses it).
func (d *Dialog) AddSelectableItemListBackButton() {
	d.addItemListBackButton(true)
}

// addItemListBackButton writes the list "Back" button: a divider on the line
// above plus a centered label on BackButtonLine (stores.cpp:341-353).
func (d *Dialog) addItemListBackButton(selectable bool) {
	d.AddSLine(BackButtonLine - 1)
	d.AddSText(0, BackButtonLine, "Back", ColorWhite, AlignCenter, selectable)
	d.Lines[BackButtonLine].YOff = 6
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
func (d *Dialog) reset() {
	*d = newDialog()
}
