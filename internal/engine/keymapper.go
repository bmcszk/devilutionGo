package engine

import "slices"

// Keymapper binds keycodes to actions and persists them in diablo.ini under
// [Keymapping]. It is the Go port of KeymapperOptions (Source/options.cpp:
// 1106-1300) and the dispatch half KeymapperPress/Release
// (Source/controls/keymapper.cpp:53-77).
//
// The model holds no gameplay knowledge: an Action carries the two callbacks
// the caller wires, and a nil callback means "no handler on that edge", like
// upstream's empty std::function.

// KeyUnknown is SDLK_UNKNOWN: the unbound key (Source/options.cpp:1216).
const KeyUnknown Key = 0

// keymapperMouseButtonMask marks a synthetic key that names a MOUSE BUTTON
// rather than a keyboard key (KeymapperMouseButtonMask, Source/options.h:711).
const keymapperMouseButtonMask Key = 1 << 31

// sdlScancodeMask is SDL's SDLK_SCANCODE_MASK: SDL keycodes for scancode-named
// keys start here (SDL_keycode.h).
const sdlScancodeMask Key = 1 << 30

// Synthetic mouse-button and scroll-wheel keys (Source/options.h:711-715).
const (
	// KeymapperMouseButtonMask marks the synthetic mouse key range.
	KeymapperMouseButtonMask = keymapperMouseButtonMask
	// MouseScrollUpButton is the wheel-up key (Source/options.h:712).
	MouseScrollUpButton Key = 65536 | keymapperMouseButtonMask
	// MouseScrollDownButton is the wheel-down key (:713).
	MouseScrollDownButton Key = 65537 | keymapperMouseButtonMask
	// MouseScrollLeftButton is the wheel-left key (:714).
	MouseScrollLeftButton Key = 65538 | keymapperMouseButtonMask
	// MouseScrollRightButton is the wheel-right key (:715).
	MouseScrollRightButton Key = 65539 | keymapperMouseButtonMask
)

// KeymappingSection is the ini section the binds live in
// (Source/options.cpp:1110).
const KeymappingSection = "Keymapping"

// The two key tables of KeymapperOptions (Source/options.cpp:1108-1184). They
// are built once at init and never mutated: the model copies nothing into
// them, so a reader needs no lock.
var (
	// keyIDToKeyName is keyIDToKeyName, ported verbatim in injection order.
	keyIDToKeyName = buildKeyNameTable()
	// keyNameToKeyID is keyNameToKeyID, the inversion upstream builds at
	// Source/options.cpp:1180-1184.
	keyNameToKeyID = invertKeyNameTable(keyIDToKeyName)
)

// Key is a keycode in the keymapper's namespace. The values mirror SDL's
// SDL_Keycode so the ini key NAMES (the only thing diablo.ini stores) resolve
// to the same identity upstream uses.
type Key uint32

// InputFlags carries the entry-mode flags the dispatch guards read. Upstream
// reads two globals (ChatFlag, DropGoldFlag, Source/controls/keymapper.cpp:65,
// :75); injecting them keeps the dispatch free of mutable package state.
type InputFlags struct {
	// Chat is ChatFlag: while the chat line is open, presses are swallowed.
	Chat bool
	// DropGold is DropGoldFlag: while the gold-amount prompt is open, number
	// and backspace releases are swallowed.
	DropGold bool
}

// Action is one bindable action (KeymapperOptions::Action,
// Source/options.h:716-760): the ini key and its default, the current bind,
// the two callbacks and the enable predicate.
type Action struct {
	// Key is the ini key and the action's identity ("QuickSave").
	Key string
	// DefaultKey is the bind used when the ini has no usable value.
	DefaultKey Key
	// BoundKey is the current bind; KeyUnknown means unbound.
	BoundKey Key
	// Press runs on key-down (upstream actionPressed).
	Press func()
	// Release runs on key-up (upstream actionReleased).
	Release func()
	// Enabled gates both edges; nil means always enabled
	// (OptionEntryBase::isEnabled, Source/options.h:744).
	Enabled func() bool
}

// Keymapper is the bind model: actions in upstream insertion order plus the
// key-to-action index the dispatcher reads (KeymapperOptions,
// Source/options.h:719-790).
type Keymapper struct {
	// actions holds the actions in insertion order once CommitActions ran;
	// before the commit AddAction's front insertion leaves them reversed.
	actions []*Action
	// byAction indexes the actions by their ini key.
	byAction map[string]*Action
	// byKey is keyIDToAction: the current owner of each bound key.
	byKey map[Key]*Action
}

// NewKeymapper returns an empty model.
func NewKeymapper() *Keymapper {
	return &Keymapper{byAction: map[string]*Action{}, byKey: map[Key]*Action{}}
}

// AddAction registers an action, pushing it to the FRONT of the list
// (KeymapperOptions::AddAction, Source/options.cpp:1291). Pair it with
// CommitActions to restore insertion order.
func (m *Keymapper) AddAction(a Action) {
	action := &a
	m.actions = append([]*Action{action}, m.actions...)
	m.byAction[action.Key] = action
}

// CommitActions reverses the front-inserted list back into insertion order
// (KeymapperOptions::CommitActions, Source/options.cpp:1296). LoadFromIni and
// SaveToIni walk that order, so callers commit first — exactly upstream's
// InitKeymapActions-then-LoadOptions sequence (Source/diablo.cpp:1839, :2157).
func (m *Keymapper) CommitActions() {
	for i, j := 0, len(m.actions)-1; i < j; i, j = i+1, j-1 {
		m.actions[i], m.actions[j] = m.actions[j], m.actions[i]
	}
}

// Actions returns the actions in insertion order.
func (m *Keymapper) Actions() []*Action { return m.actions }

// Action returns the action registered under an ini key, or nil.
func (m *Keymapper) Action(key string) *Action { return m.byAction[key] }

// FindAction returns the action bound to a key, or nil
// (KeymapperOptions::findAction, Source/options.h:786-789).
func (m *Keymapper) FindAction(key Key) *Action { return m.byKey[key] }

// SetValue binds an action to a key, reporting whether the value was accepted
// (KeymapperOptions::Action::SetValue, Source/options.cpp:1262-1285). An
// unknown key value is refused; KeyUnknown unbinds; a key already held by
// another action unbinds that previous owner.
func (m *Keymapper) SetValue(actionKey string, key Key) bool {
	action := m.byAction[actionKey]
	if action == nil {
		return false
	}

	if key != KeyUnknown && !knownKey(key) {
		return false
	}

	return m.bind(action, key)
}

// LoadFromIni reads every action's bind from section "Keymapping"
// (KeymapperOptions::Action::LoadFromIni, Source/options.cpp:1210-1234): an
// absent key keeps the default, an empty value unbinds, and an unknown name
// falls back to the default.
func (m *Keymapper) LoadFromIni(in *Ini) {
	for _, action := range m.actions {
		m.loadAction(in, action)
	}
}

// SaveToIni writes every action's bind into section "Keymapping"
// (KeymapperOptions::Action::SaveToIni, Source/options.cpp:1236-1249): an
// unbound action writes an empty value, a bound one writes its key name.
func (m *Keymapper) SaveToIni(in *Ini) {
	for _, action := range m.actions {
		if action.BoundKey == KeyUnknown {
			in.SetString(KeymappingSection, action.Key, "")

			continue
		}

		if name, ok := KeyName(action.BoundKey); ok {
			in.SetString(KeymappingSection, action.Key, name)
		}
	}
}

// Press dispatches a key-down (KeymapperPress,
// Source/controls/keymapper.cpp:53-67): the guard order is unmapped action,
// then no press handler, then disabled, then the chat flag.
func (m *Keymapper) Press(key Key, flags InputFlags) {
	action := m.FindAction(ToAsciiUpper(key))
	if action == nil || action.Press == nil || !action.enabled() {
		return
	}

	if flags.Chat {
		return
	}

	action.Press()
}

// Release dispatches a key-up (KeymapperRelease,
// Source/controls/keymapper.cpp:69-77) with the release-side carve-outs: a chat
// session swallows text-entry keys, the gold prompt swallows number keys.
func (m *Keymapper) Release(key Key, flags InputFlags) {
	key = ToAsciiUpper(key)

	action := m.FindAction(key)
	if action == nil || action.Release == nil || !action.enabled() {
		return
	}

	if (flags.Chat && IsTextEntryKey(key)) || (flags.DropGold && IsNumberEntryKey(key)) {
		return
	}

	action.Release()
}

// KeyName returns the ini name of a key, or false when the key is not in the
// table (keyIDToKeyName, Source/options.cpp:1108-1178).
func KeyName(key Key) (string, bool) {
	name, ok := keyIDToKeyName[key]

	return name, ok
}

// KeyID returns the key a name resolves to, or false when unknown
// (keyNameToKeyID, Source/options.cpp:1180-1184).
func KeyID(name string) (Key, bool) {
	key, ok := keyNameToKeyID[name]

	return key, ok
}

// KeyNames returns every bindable key name, sorted: the list a rebind UI
// offers, and the exhaustive set the table round-trip test walks.
func KeyNames() []string {
	names := make([]string, 0, len(keyIDToKeyName))
	for _, name := range keyIDToKeyName {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// ToAsciiUpper folds a lower-case letter keycode onto its upper-case twin
// (ToAsciiUpper, Source/controls/keymapper.cpp:37-49).
func ToAsciiUpper(key Key) Key {
	if key >= 'a' && key <= 'z' {
		return key - ('a' - 'A')
	}

	return key
}

// IsTextEntryKey reports whether a key belongs to a text-entry session
// (IsTextEntryKey, Source/controls/keymapper.cpp:22-30).
func IsTextEntryKey(key Key) bool {
	switch key {
	case 27, 13, sdlScancodeMask | 88, 8, sdlScancodeMask | 81, sdlScancodeMask | 82:
		return true
	}

	return key >= ' ' && key <= 'z'
}

// IsNumberEntryKey reports whether a key belongs to a number-entry session
// (IsNumberEntryKey, Source/controls/keymapper.cpp:32-35).
func IsNumberEntryKey(key Key) bool {
	return (key >= '0' && key <= '9') || key == 8
}

// enabled mirrors OptionEntryBase::isEnabled (Source/options.h:744): a missing
// predicate means the action is always available.
func (a *Action) enabled() bool { return a.Enabled == nil || a.Enabled() }

// loadAction applies the absent/empty/unknown rules to one action
// (Action::LoadFromIni, Source/options.cpp:1210-1234).
func (m *Keymapper) loadAction(in *Ini, action *Action) {
	values := in.Get(KeymappingSection, action.Key)
	if len(values) == 0 {
		m.SetValue(action.Key, action.DefaultKey)

		return
	}

	name := values[len(values)-1].Value
	if name == "" {
		m.SetValue(action.Key, KeyUnknown)

		return
	}

	if key, ok := KeyID(name); ok {
		m.SetValue(action.Key, key)

		return
	}

	// An unknown key name keeps the default upstream (:1228-1231).
	m.SetValue(action.Key, action.DefaultKey)
}

// bind applies the drop-then-install half of SetValue
// (Action::SetValue, Source/options.cpp:1262-1285).
func (m *Keymapper) bind(action *Action, key Key) bool {
	if action.BoundKey != KeyUnknown {
		delete(m.byKey, action.BoundKey)
		action.BoundKey = KeyUnknown
	}

	if key == KeyUnknown {
		return true
	}

	if holder := m.byKey[key]; holder != nil {
		// Warn about overwriting keys (:1273-1277).
		holder.BoundKey = KeyUnknown
	}

	m.byKey[key] = action
	action.BoundKey = key

	return true
}

// knownKey reports whether the key is in the name table, the gate SetValue
// applies (Source/options.cpp:1264-1267).
func knownKey(key Key) bool {
	_, ok := keyIDToKeyName[key]

	return ok
}

// buildKeyNameTable builds keyIDToKeyName in upstream's insertion order
// (KeymapperOptions::KeymapperOptions, Source/options.cpp:1108-1178).
func buildKeyNameTable() map[Key]string {
	table := make(map[Key]string, 128)

	// A-Z and 0-9 (:1109-1115).
	for c := Key('A'); c <= 'Z'; c++ {
		table[c] = string(rune(c))
	}

	for c := Key('0'); c <= '9'; c++ {
		table[c] = string(rune(c))
	}

	addFunctionKeys(table)
	addKeypadKeys(table)
	addSpecialKeys(table)

	return table
}

// addFunctionKeys adds F1-F24: F1-F12 are scancodes 58..69, F13-F24 are
// 104..115 (Source/options.cpp:1116-1123).
func addFunctionKeys(table map[Key]string) {
	for i := range 12 {
		table[sdlScancodeMask|58+Key(i)] = "F" + itoa(i+1)
		table[sdlScancodeMask|104+Key(i)] = "F" + itoa(i+13)
	}
}

// addKeypadKeys adds the keypad digits and operators: KP_1..KP_9 are scancodes
// 89..97 and KP_0 is 98 (Source/options.cpp:1125-1128, :1151-1154).
func addKeypadKeys(table map[Key]string) {
	table[sdlScancodeMask|98] = "KEYPADNUM 0"
	for i := range 9 {
		table[sdlScancodeMask|89+Key(i)] = "KEYPADNUM " + itoa(i+1)
	}

	table[sdlScancodeMask|84] = "KEYPAD /"
	table[sdlScancodeMask|85] = "KEYPAD *"
	table[sdlScancodeMask|88] = "KEYPAD ENTER"
	table[sdlScancodeMask|99] = "KEYPAD DECIMAL"
}

// addSpecialKeys adds the modifiers, whitespace, mouse and punctuation names
// (Source/options.cpp:1130-1178).
func addSpecialKeys(table map[Key]string) {
	for key, name := range map[Key]string{
		sdlScancodeMask | 226:        "LALT",
		sdlScancodeMask | 230:        "RALT",
		' ':                          "SPACE",
		sdlScancodeMask | 228:        "RCONTROL",
		sdlScancodeMask | 224:        "LCONTROL",
		sdlScancodeMask | 70:         "PRINT",
		sdlScancodeMask | 72:         "PAUSE",
		9:                            "TAB",
		2 | keymapperMouseButtonMask: "MMOUSE",
		4 | keymapperMouseButtonMask: "X1MOUSE",
		5 | keymapperMouseButtonMask: "X2MOUSE",
		MouseScrollUpButton:          "SCROLLUPMOUSE",
		MouseScrollDownButton:        "SCROLLDOWNMOUSE",
		MouseScrollLeftButton:        "SCROLLLEFTMOUSE",
		MouseScrollRightButton:       "SCROLLRIGHTMOUSE",
		'`':                          "`",
		'[':                          "[",
		']':                          "]",
		'\\':                         "\\",
		';':                          ";",
		'\'':                         "'",
		',':                          ",",
		'.':                          ".",
		'/':                          "/",
		8:                            "BACKSPACE",
		sdlScancodeMask | 57:         "CAPSLOCK",
		sdlScancodeMask | 71:         "SCROLLLOCK",
		sdlScancodeMask | 73:         "INSERT",
		sdlScancodeMask | 76:         "DELETE",
		sdlScancodeMask | 74:         "HOME",
		sdlScancodeMask | 77:         "END",
	} {
		table[key] = name
	}
}

// invertKeyNameTable builds keyNameToKeyID from the forward table
// (Source/options.cpp:1180-1184).
func invertKeyNameTable(forward map[Key]string) map[string]Key {
	inverse := make(map[string]Key, len(forward))
	for key, name := range forward {
		inverse[name] = key
	}

	return inverse
}

// itoa renders a non-negative int without pulling strconv in for the two digits
// the table needs.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}

	var digits [3]byte

	pos := len(digits)

	for value > 0 {
		pos--
		digits[pos] = byte('0' + value%10)
		value /= 10
	}

	return string(digits[pos:])
}
