package controls

// CURSOR_* values (cursor_defs.hpp:20-32), ordinals preserved. All 13
// ported — the dispatch branches on the range pcurs > CURSOR_HAND &&
// pcurs < CURSOR_FIRSTITEM, so both the action cursors AND the FIRSTITEM
// sentinel must exist; HOURGLASS is a named branch.
const (
	// CursorNone (CURSOR_NONE): no cursor. Render-only at the dispatch.
	CursorNone CursorID = iota
	// CursorHand (CURSOR_HAND): the default hand cursor; the "no special
	// cursor" baseline of every range check.
	CursorHand
	// CursorIdentify (CURSOR_IDENTIFY): scroll of identify targeting.
	CursorIdentify
	// CursorRepair (CURSOR_REPAIR): repair targeting.
	CursorRepair
	// CursorRecharge (CURSOR_RECHARGE): staff recharge targeting.
	CursorRecharge
	// CursorDisarm (CURSOR_DISARM): trap disarm targeting.
	CursorDisarm
	// CursorOil (CURSOR_OIL): oil use targeting.
	CursorOil
	// CursorTelekinesis (CURSOR_TELEKINESIS): telekinesis targeting.
	CursorTelekinesis
	// CursorResurrect (CURSOR_RESURRECT): resurrect targeting.
	CursorResurrect
	// CursorTeleport (CURSOR_TELEPORT): teleport targeting.
	CursorTeleport
	// CursorHealOther (CURSOR_HEALOTHER): heal-other targeting.
	CursorHealOther
	// CursorHourglass (CURSOR_HOURGLASS): a pending time-cost action is in
	// flight; the dispatch refuses (hourglass no-op, plrctrls.cpp:2319-2320).
	CursorHourglass
	// CursorFirstItem (CURSOR_FIRSTITEM): the first item-drag cursor — the
	// exclusive END of the icon-cursor range; not itself an icon cursor.
	CursorFirstItem
)

// The panel regions MousePosition can be in (plrctrls.cpp:2323-2336). A
// position may satisfy more than one upstream rectangle at once; the guard
// order (right/main before left) IS the resolution, so the port takes the
// pre-resolved winner.
const (
	// PanelNone: the mouse is over no dispatch-relevant panel.
	PanelNone PanelRegion = iota
	// PanelRightOrMain: GetRightPanel().contains || GetMainPanel().contains.
	PanelRightOrMain
	// PanelLeft: GetLeftPanel().contains only.
	PanelLeft
)

// CursorID mirrors enum cursor_id (Source/cursor_defs.hpp:19-33 @ fork
// ca91d778f): the pcurs value space. SelectionRegion (cursor_defs.hpp:11-18)
// is not pcurs state and is not ported (screen geometry seam, later slice).
type CursorID uint8

// PanelRegion collapses the panel-containment queries the dispatch makes —
// GetRightPanel().contains(MousePosition), GetMainPanel().contains(MousePosition),
// GetLeftPanel().contains(MousePosition) (plrctrls.cpp:2323,2325,2326,2328,
// 2331,2335) — into one input value. The rectangles themselves stay at the
// wiring layer (screen-geometry seam, later slice).
type PanelRegion uint8

// ActionView is the narrow read surface the primary-action dispatch needs
// over upstream globals — house PlayerView pattern (explicit inputs so the
// decision order is testable before entity wiring lands; the adapter onto
// the game state lives at the call site; the game kernel never imports
// controls). Every field cites its upstream source.
type ActionView interface {
	// SpellSelectFlag (plrctrls.cpp:2311): the speed-book selection mode is
	// armed — the next primary action assigns the spell to the slot.
	SpellSelectFlag() bool
	// InventoryOpen is invflag (plrctrls.cpp:2316).
	InventoryOpen() bool
	// Cursor is pcurs (cursor_defs.hpp; read at plrctrls.cpp:2318-2319).
	Cursor() CursorID
	// CharFlag (plrctrls.cpp:2340): the character panel is open.
	CharFlag() bool
	// CharPanelButtonActive (plrctrls.cpp:2340,2343): a stat button is
	// currently held down.
	CharPanelButtonActive() bool
	// StatPoints is MyPlayer->_pStatPts (plrctrls.cpp:2340).
	StatPoints() int
	// StashOpen is IsStashOpen (plrctrls.cpp:2325).
	StashOpen() bool
	// VisualStoreOpen is IsVisualStoreOpen (plrctrls.cpp:2327).
	VisualStoreOpen() bool
	// HoldItemEmpty is MyPlayer->HoldItem.isEmpty() (plrctrls.cpp:2328).
	HoldItemEmpty() bool
	// StoreButton is pcursstorebtn, -1 when none (plrctrls.cpp:2331).
	StoreButton() int
	// MouseInPanel is the pre-resolved panel containment of MousePosition
	// (plrctrls.cpp:2323-2336): right/main wins over left by guard order.
	MouseInPanel() PanelRegion
	// OverlaysActive models the upstream CALL-SITE refusal: GetController
	// ActionForButton returns false while PadMenuNavigatorActive ||
	// PadHotspellMenuActive (game_controls.cpp:197-198), and the InGameMenu
	// family (plrctrls.cpp:73-82) gates menu-consumed states — the machine
	// is simply never invoked upstream. When true, PerformPrimaryAction
	// performs nothing.
	OverlaysActive() bool
}

// ActionHooks are the environment functions the dispatch emits through: the
// upstream action functions (plrctrls.h:54-64 signatures) whose internals
// live in later slices. A nil hook is skipped — the dispatch's DECISION is
// the contract, not the side effects; the wiring layer installs real
// implementations as the dependency slices land.
type ActionHooks struct {
	// SetSpell ports SetSpell (plrctrls.h; called plrctrls.cpp:2312).
	SetSpell func()
	// TryIconCurs ports TryIconCurs (called plrctrls.cpp:2321).
	TryIconCurs func()
	// NewCursor ports NewCursor (engine; called plrctrls.cpp:2322).
	NewCursor func(CursorID)
	// LiftInventoryItem ports LiftInventoryItem (called plrctrls.cpp:2324).
	LiftInventoryItem func()
	// LiftStashItem ports LiftStashItem (called plrctrls.cpp:2326).
	LiftStashItem func()
	// CheckVisualStorePaste ports CheckVisualStorePaste(MousePosition)
	// (called plrctrls.cpp:2329) — MousePosition is fixed at the call, so
	// the hook is niladic.
	CheckVisualStorePaste func()
	// CheckVisualStoreButtonPress ports CheckVisualStoreButtonPress
	// (MousePosition) (called plrctrls.cpp:2333) — press only; the matching
	// release happens on button-up via PerformPrimaryActionRelease.
	CheckVisualStoreButtonPress func()
	// CheckVisualStoreButtonRelease ports CheckVisualStoreButtonRelease
	// (MousePosition) (called plrctrls.cpp:2351 from
	// PerformPrimaryActionRelease).
	CheckVisualStoreButtonRelease func()
	// CheckVisualStoreItem ports CheckVisualStoreItem(MousePosition, false,
	// false) (called plrctrls.cpp:2335) — the (false, false) flags are fixed
	// at the upstream call site, so the hook is niladic.
	CheckVisualStoreItem func()
	// CheckChrBtns ports CheckChrBtns (called plrctrls.cpp:2341) — returns
	// whether a stat button took the press (upstream's
	// CharPanelButtonActive flip).
	CheckChrBtns func() bool
	// ReleaseChrBtns ports ReleaseChrBtns(false) (called plrctrls.cpp:2342)
	// — the const flag is fixed at the upstream call site.
	ReleaseChrBtns func()
	// Interact ports Interact (called plrctrls.cpp:2345) — the world-
	// interaction fallthrough (attack/talk/open targets resolved by
	// CheckCursMove).
	Interact func()
}

// String keeps test and log output readable ("NewCursor:CURSOR_HAND").
func (c CursorID) String() string {
	names := [...]string{
		"CURSOR_NONE", "CURSOR_HAND", "CURSOR_IDENTIFY", "CURSOR_REPAIR",
		"CURSOR_RECHARGE", "CURSOR_DISARM", "CURSOR_OIL", "CURSOR_TELEKINESIS",
		"CURSOR_RESURRECT", "CURSOR_TELEPORT", "CURSOR_HEALOTHER",
		"CURSOR_HOURGLASS", "CURSOR_FIRSTITEM",
	}
	if int(c) < len(names) {
		return names[c]
	}
	return "CURSOR_INVALID"
}

// PerformPrimaryAction ports PerformPrimaryAction (Source/controls/
// plrctrls.cpp:2309-2346 @ fork ca91d778f): given the virtual-cursor state,
// decide which upstream action fires — or nothing. Guard order IS the
// behavior and is mirrored exactly, early returns included.
//
// Deviations, both input-modeling only (no branch reordered, no branch cut):
//   - OverlaysActive: upstream refuses BEFORE the machine at the call site
//     (game_controls.cpp:197-198); the port takes it as an input flag.
//   - Panel containment: upstream queries three rectangles at MousePosition
//     (plrctrls.cpp:2323-2336); the port takes the pre-resolved winner
//     (right/main beats left upstream via guard order).
func PerformPrimaryAction(view ActionView, hooks ActionHooks) {
	if view.OverlaysActive() {
		return
	}
	if view.SpellSelectFlag() {
		hooks.SetSpell() // plrctrls.cpp:2312; return :2313
		return
	}
	if view.InventoryOpen() {
		dispatchInventory(view, hooks)
		return // plrctrls.cpp:2337: the invflag block always returns
	}
	// plrctrls.cpp:2340-2344: char panel stat allocation.
	if view.CharFlag() && !view.CharPanelButtonActive() && view.StatPoints() > 0 {
		if hooks.CheckChrBtns() {
			hooks.ReleaseChrBtns()
		}
		return
	}
	hooks.Interact() // plrctrls.cpp:2345: the world-interaction fallthrough
}

// PerformPrimaryActionRelease ports PerformPrimaryActionRelease
// (plrctrls.cpp:2349-2353): the button-up half of the visual-store press —
// release the pressed button while the store is open with the mouse over
// its panel. Upstream guards it with the same ControllerActionHeld ==
// GameActionType_PRIMARY_ACTION release path (game_controls.cpp:186-190),
// so the overlay refusal carries over as the same OverlaysActive gate.
func PerformPrimaryActionRelease(view ActionView, hooks ActionHooks) {
	if view.OverlaysActive() {
		return
	}
	// plrctrls.cpp:2350: IsVisualStoreOpen && GetLeftPanel().contains(MousePosition).
	if view.VisualStoreOpen() && view.MouseInPanel() == PanelLeft {
		hooks.CheckVisualStoreButtonRelease() // plrctrls.cpp:2351
	}
}

// dispatchInventory ports the invflag block (plrctrls.cpp:2316-2338): the
// icon-cursor family, then the panel lift/store family, then the mandatory
// return. The block ALWAYS returns — nothing past it runs while the
// inventory is open.
func dispatchInventory(view ActionView, hooks ActionHooks) {
	if tryIconCursor(view, hooks) {
		return
	}
	// plrctrls.cpp:2323-2324: right panel or main panel — lift inventory.
	if view.MouseInPanel() == PanelRightOrMain {
		hooks.LiftInventoryItem()
		return
	}
	overLeft := view.MouseInPanel() == PanelLeft
	// plrctrls.cpp:2325-2326: stash open + left panel — lift stash item.
	if view.StashOpen() && overLeft {
		hooks.LiftStashItem()
		return
	}
	// plrctrls.cpp:2327-2336: visual store open + left panel — the
	// paste/button/item family.
	if view.VisualStoreOpen() && overLeft {
		dispatchVisualStore(view, hooks)
	}
}

// tryIconCursor ports the icon-cursor branch (plrctrls.cpp:2318-2322): a
// cursor in (CURSOR_HAND, CURSOR_FIRSTITEM) is an item-targeting cursor
// loaded over the inventory. Reports whether it consumed the action —
// including the HOURGLASS refusal, which deliberately does NOT reset the
// cursor (plrctrls.cpp:2319-2320).
func tryIconCursor(view ActionView, hooks ActionHooks) bool {
	c := view.Cursor()
	if c <= CursorHand || c >= CursorFirstItem {
		return false
	}
	if c == CursorHourglass {
		return true
	}
	hooks.TryIconCurs()         // plrctrls.cpp:2321
	hooks.NewCursor(CursorHand) // plrctrls.cpp:2322
	return true
}

// dispatchVisualStore ports the visual-store inner branches
// (plrctrls.cpp:2328-2336): paste a held item, press a hovered button
// (release comes on button-up), or inspect the hovered item.
func dispatchVisualStore(view ActionView, hooks ActionHooks) {
	if !view.HoldItemEmpty() {
		hooks.CheckVisualStorePaste() // plrctrls.cpp:2329
		return
	}
	if view.StoreButton() != -1 {
		// plrctrls.cpp:2332-2333: only press; release is called when the
		// button is released (PerformPrimaryActionRelease).
		hooks.CheckVisualStoreButtonPress()
		return
	}
	hooks.CheckVisualStoreItem() // plrctrls.cpp:2335 (false,false fixed at call)
}
