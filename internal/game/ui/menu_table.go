package ui

// MenuItem is one row of a menu table: the Go shape of upstream TMenuItem's
// dwFlags + pszStr pair (X:Source/gmenu.h:20-23). GMENU_ENABLED 0x80000000 ->
// Enabled, GMENU_SLIDER 0x40000000 -> Slider.
type MenuItem struct {
	// Label is the row's text (TMenuItem.pszStr).
	Label string
	// Enabled carries GMENU_ENABLED: a disabled row is skipped by navigation
	// and never activates (X:gmenu.cpp:82-105, :285-289).
	Enabled bool
	// Slider carries GMENU_SLIDER: the row draws a fill bar and steps on
	// Left/Right (X:gmenu.cpp:107-124).
	Slider bool
}

// MenuTable is the menu-as-data model: rows with enabled flags plus the
// gmenu slider arithmetic, ported from X:Source/gmenu.cpp verbatim. It holds
// no callbacks and imports nothing — the screen that presents it keys its
// actions by row index.
type MenuTable struct {
	rows []MenuItem
	// steps holds the per-row slider step count (gmenu_slider_steps stores it
	// in the dwFlags bits 12..23, X:gmenu.cpp:388-391).
	steps []int
	// step holds the per-row current slider step (TMenuItem sliderStep).
	step []int
	// focused is the index of the focused row (sgpCurrItem's offset).
	focused int
}

// NewMenuTable builds the table over rows in display order, focused on the
// last row before the first navigation — gmenu_set_items focuses the last
// row then wraps forward onto the first enabled one (X:gmenu.cpp:238-244).
func NewMenuTable(rows []MenuItem) *MenuTable {
	t := &MenuTable{
		rows:    append([]MenuItem(nil), rows...),
		steps:   make([]int, len(rows)),
		step:    make([]int, len(rows)),
		focused: len(rows) - 1,
	}
	t.focused = t.NextEnabled(t.focused)

	return t
}

// Rows returns the table's rows in display order.
func (t *MenuTable) Rows() []MenuItem { return t.rows }

// Focused returns the index of the focused row (sgpCurrItem).
func (t *MenuTable) Focused() int { return t.focused }

// Focus moves the focus to the first enabled row (GmenuUpDown(true) from the
// last row wraps, X:gmenu.cpp:82-105); -1 when nothing is enabled.
func (t *MenuTable) Focus() bool {
	first := t.FirstEnabled()
	if first < 0 {
		return false
	}
	t.focused = first

	return true
}

// FirstEnabled returns the first enabled row index, or -1 when none is.
func (t *MenuTable) FirstEnabled() int {
	for i, r := range t.rows {
		if r.Enabled {
			return i
		}
	}

	return -1
}

// NextEnabled returns the index of the first enabled row strictly after from,
// wrapping; from itself when every row is disabled. Disabled rows are skipped
// (X:Source/gmenu.cpp:82-105 GmenuUpDown).
func (t *MenuTable) NextEnabled(from int) int {
	return t.walk(from, 1)
}

// PrevEnabled returns the index of the first enabled row strictly before from,
// wrapping; from itself when every row is disabled (X:gmenu.cpp:82-105).
func (t *MenuTable) PrevEnabled(from int) int {
	return t.walk(from, -1)
}

// MoveDown moves the focus one enabled row down with wrap (GmenuUpDown(true))
// and reports the new focus.
func (t *MenuTable) MoveDown() int {
	t.focused = t.NextEnabled(t.focused)

	return t.focused
}

// MoveUp moves the focus one enabled row up with wrap (GmenuUpDown(false))
// and reports the new focus.
func (t *MenuTable) MoveUp() int {
	t.focused = t.PrevEnabled(t.focused)

	return t.focused
}

// SetSliderSteps stores a row's slider step count (gmenu_slider_steps,
// X:gmenu.cpp:388-391; the min-2 clamp applies at use).
func (t *MenuTable) SetSliderSteps(row, steps int) {
	t.steps[row] = steps
}

// SliderSet maps a value in [min,max] onto the row's step with the upstream
// integer arithmetic (gmenu_slider_set, X:Source/gmenu.cpp:375-379):
//
//	step = ((max-min-1)/2 + (value-min)*nSteps) / (max-min)
//
// with nSteps clamped to at least 2 at use (:377).
func (t *MenuTable) SliderSet(row, lo, hi, value int) {
	steps := max2(t.steps[row], 2)
	t.step[row] = ((hi-lo-1)/2 + (value-lo)*steps) / (hi - lo)
}

// SliderGet maps the row's step back onto a value in [min,max] with the
// upstream integer arithmetic (gmenu_slider_get, X:Source/gmenu.cpp:381-386):
//
//	value = min + (step*(max-min) + (nSteps-1)/2) / nSteps
func (t *MenuTable) SliderGet(row, lo, hi int) int {
	step := t.step[row]
	steps := max2(t.steps[row], 2)

	return lo + (step*(hi-lo)+(steps-1)/2)/steps
}

// SliderStepRight steps the row one step up, saturating at nSteps, and
// reports the new step (GmenuLeftRight(true), X:gmenu.cpp:107-124).
func (t *MenuTable) SliderStepRight(row int) int { return t.stepSlider(row, 1) }

// SliderStepLeft steps the row one step down, saturating at 0, and reports
// the new step (GmenuLeftRight(false), X:gmenu.cpp:107-124).
func (t *MenuTable) SliderStepLeft(row int) int { return t.stepSlider(row, -1) }

// Step reports the row's current slider step (TMenuItem sliderStep).
func (t *MenuTable) Step(row int) int { return t.step[row] }

// walk moves from in direction dir (±1), one row per hop with wrap, and stops
// at the first enabled row. With no enabled row it returns from unchanged.
func (t *MenuTable) walk(from, dir int) int {
	n := len(t.rows)
	if n == 0 {
		return from
	}

	i := from
	for range n {
		i = (i + dir + n) % n
		if t.rows[i].Enabled {
			return i
		}
	}

	return from
}

// stepSlider applies the saturated one-step move of GmenuLeftRight.
func (t *MenuTable) stepSlider(row, dir int) int {
	steps := max2(t.steps[row], 2)
	next := t.step[row] + dir
	if next < 0 || next > steps {
		return t.step[row]
	}
	t.step[row] = next

	return t.step[row]
}

// max2 returns the larger of a and b (std::max<uint16_t> at
// X:gmenu.cpp:377,:386).
func max2(a, b int) int {
	if a > b {
		return a
	}

	return b
}
