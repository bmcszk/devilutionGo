package engine_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// The keymapper bind model (Options slice S2): the Go port of
// KeymapperOptions (Source/options.cpp:1106-1300) and the dispatch guards of
// KeymapperPress/Release (Source/controls/keymapper.cpp:22-77).

// spyAction is one registered action plus the edge counters the guard-order
// cases read.
type spyAction struct {
	pressed  int
	released int
}

// keyTableCase is one documented name/keycode identity.
type keyTableCase struct {
	name string
	want engine.Key
}

// toAsciiUpperCase is one folding case.
type toAsciiUpperCase struct {
	name string
	in   engine.Key
	want engine.Key
}

// loadRuleCase is one absent/empty/known/unknown load case.
type loadRuleCase struct {
	name string
	doc  string
	want string
}

// pressGuardCase is one Press guard-order case: the registered action, the ini
// bind it loads and the flags the press sees.
type pressGuardCase struct {
	name      string
	register  func(*engine.Keymapper) *spyAction
	bind      string
	key       engine.Key
	flags     engine.InputFlags
	wantCalls int
}

// releaseCarveOutCase is one Release carve-out case: the bound key, its ini
// name and the flags the release sees.
type releaseCarveOutCase struct {
	name      string
	key       engine.Key
	keyName   string
	flags     engine.InputFlags
	wantCalls int
}

// entryRuleCase is one IsTextEntryKey/IsNumberEntryKey case.
type entryRuleCase struct {
	name   string
	key    engine.Key
	text   bool
	number bool
}

func TestKeymapperKeyNamesRoundTrip(t *testing.T) {
	for _, name := range engine.KeyNames() {
		t.Run(name, func(t *testing.T) {
			// given one name from the ported keyIDToKeyName table
			tableName := name

			// when resolving it to a key and back
			back, found := keyNameRoundTrip(tableName)

			// then the name survives the round trip
			if !found || back != tableName {
				t.Fatalf("KeyName(KeyID(%q)) = %q (found %v), want %q", tableName, back, found, tableName)
			}
		})
	}
}

func TestKeymapperKeyTableEntries(t *testing.T) {
	for _, tc := range keyTableCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			// given the case's documented keycode identity
			want := tc.want

			// when looking the name up
			got, ok := engine.KeyID(tc.name)

			// then it resolves to that keycode
			if !ok || got != want {
				t.Fatalf("KeyID(%q) = %d (found %v), want %d", tc.name, got, ok, want)
			}
		})
	}
}

func TestKeymapperToAsciiUpperFoldsOnlyLowercaseLetters(t *testing.T) {
	for _, tc := range toAsciiUpperCases() {
		t.Run(tc.name, func(t *testing.T) {
			// given the case's keycode
			want := tc.want

			// when folding it
			got := engine.ToAsciiUpper(tc.in)

			// then the -32 letter rule applies and nothing else
			// (Source/controls/keymapper.cpp:37-49)
			if got != want {
				t.Fatalf("ToAsciiUpper(%d) = %d, want %d", tc.in, got, want)
			}
		})
	}
}

func TestKeymapperEntryKeyPredicates(t *testing.T) {
	for _, tc := range entryRuleCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			// given the case's key and its expected classification
			wantText, wantNumber := tc.text, tc.number

			// when classifying the key for both entry modes
			gotText := engine.IsTextEntryKey(tc.key)
			gotNumber := engine.IsNumberEntryKey(tc.key)

			// then each predicate matches
			// (Source/controls/keymapper.cpp:22-35)
			if gotText != wantText || gotNumber != wantNumber {
				t.Fatalf("IsTextEntryKey/IsNumberEntryKey(%d) = %v/%v, want %v/%v",
					tc.key, gotText, gotNumber, wantText, wantNumber)
			}
		})
	}
}

func TestKeymapperLoadAppliesTheUpstreamRules(t *testing.T) {
	for _, tc := range loadRuleCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			// given a keymapper whose action defaults to F2
			km := newFixtureKeymapper(t, mustKeyID(t, "F2"))

			// when loading the case's document
			km.LoadFromIni(mustParseIni(t, []byte(tc.doc)))

			// then the bind follows the absent/empty/known/unknown rules
			// (Action::LoadFromIni, Source/options.cpp:1210-1234)
			assertBoundName(t, km, "QuickSave", tc.want)
		})
	}
}

func TestKeymapperLoadUnbindsThePreviousOwnerOfADuplicate(t *testing.T) {
	// given two actions, both defaulting onto F2
	km := engine.NewKeymapper()
	registerSpy(km, "ActionA", mustKeyID(t, "F2"), nil)
	registerSpy(km, "ActionB", mustKeyID(t, "F2"), nil)
	km.CommitActions()

	// when both load onto F2
	km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nActionA=F2\nActionB=F2\n")))

	// then the LAST loader owns the key and the first was unbound
	// (Action::SetValue, Source/options.cpp:1262-1285)
	assertBound(t, km, "ActionB", mustKeyID(t, "F2"))
	assertBound(t, km, "ActionA", engine.KeyUnknown)
}

func TestKeymapperSetValueRejectsAnUnknownKey(t *testing.T) {
	// given an action bound to F2 through the ini
	km := newFixtureKeymapper(t, mustKeyID(t, "F2"))
	km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nQuickSave=F2\n")))

	// when binding a keycode that is not in the name table
	accepted := km.SetValue("QuickSave", 0xDEADBEEF)

	// then the value is refused and the old bind survives
	if accepted {
		t.Fatal("SetValue accepted an unknown keycode")
	}

	assertBound(t, km, "QuickSave", mustKeyID(t, "F2"))
}

func TestKeymapperSaveWritesNamesAndEmptyForUnbound(t *testing.T) {
	// given one bound and one unbound action
	km := engine.NewKeymapper()
	registerSpy(km, "QuickSave", engine.KeyUnknown, nil)
	registerSpy(km, "StopHero", engine.KeyUnknown, nil)
	km.CommitActions()

	km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nQuickSave=F6\nStopHero=\n")))

	out := mustParseIni(t, nil)

	// when saving the section
	km.SaveToIni(out)

	// then the bound action writes its key name and the unbound one an empty
	// value (Action::SaveToIni, Source/options.cpp:1236-1249)
	if got := out.GetString(engine.KeymappingSection, "QuickSave", "<absent>"); got != "F6" {
		t.Fatalf("saved QuickSave = %q, want F6", got)
	}

	values := out.Get(engine.KeymappingSection, "StopHero")
	if len(values) != 1 || values[0].Value != "" {
		t.Fatalf("saved StopHero = %+v, want one empty (unbound) value", values)
	}
}

func TestKeymapperPressGuardOrder(t *testing.T) {
	for _, tc := range pressGuardCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			// given the case's action registered and bound
			km := engine.NewKeymapper()
			spy := tc.register(km)
			km.CommitActions()

			km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nA="+tc.bind+"\n")))

			// when pressing the case's key with its flags
			km.Press(tc.key, tc.flags)

			// then the press handler ran exactly as often as the guard allows
			// (KeymapperPress, Source/controls/keymapper.cpp:53-67)
			if spy.pressed != tc.wantCalls {
				t.Fatalf("press calls = %d, want %d", spy.pressed, tc.wantCalls)
			}
		})
	}
}

func TestKeymapperReleaseCarveOuts(t *testing.T) {
	for _, tc := range releaseCarveOutCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			// given an action bound to the case's key
			km := engine.NewKeymapper()
			spy := registerSpy(km, "A", engine.KeyUnknown, nil)
			km.CommitActions()

			km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nA="+tc.keyName+"\n")))

			// when releasing with the case's flags
			km.Release(tc.key, tc.flags)

			// then the release handler ran exactly as often as the carve-outs
			// allow (KeymapperRelease, Source/controls/keymapper.cpp:69-77)
			if spy.released != tc.wantCalls {
				t.Fatalf("release calls = %d, want %d", spy.released, tc.wantCalls)
			}
		})
	}
}

func TestKeymapperAddActionPushesFrontAndCommitReverses(t *testing.T) {
	// given two actions added in order
	km := engine.NewKeymapper()
	km.AddAction(engine.Action{Key: "First"})
	km.AddAction(engine.Action{Key: "Second"})

	// when reading them before the commit
	before := actionKeys(km.Actions())

	// then the list is reversed (AddAction pushes front, Source/options.cpp:1291)
	if len(before) != 2 || before[0] != "Second" {
		t.Fatalf("actions before CommitActions = %v, want Second first", before)
	}

	// when committing
	km.CommitActions()

	// then insertion order is restored (:1296)
	after := actionKeys(km.Actions())
	if len(after) != 2 || after[0] != "First" || after[1] != "Second" {
		t.Fatalf("actions after CommitActions = %v, want [First Second]", after)
	}
}

func TestKeymapperFindActionFollowsTheBind(t *testing.T) {
	// given an action rebound through the ini
	km := newFixtureKeymapper(t, mustKeyID(t, "F2"))
	km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nQuickSave=F6\n")))

	// when looking the OLD key up
	oldOwner := km.FindAction(mustKeyID(t, "F2"))

	// then it resolves to nothing (keyIDToAction, Source/options.cpp:1274)
	if oldOwner != nil {
		t.Fatalf("FindAction(F2) = %q, want nil (the action rebound away)", oldOwner.Key)
	}

	// and the NEW key resolves to the action
	newOwner := km.FindAction(mustKeyID(t, "F6"))
	if newOwner == nil || newOwner.Key != "QuickSave" {
		t.Fatalf("FindAction(F6) = %v, want the QuickSave action", newOwner)
	}
}

// registerSpy adds an action whose callbacks count their invocations.
func registerSpy(km *engine.Keymapper, key string, defaultKey engine.Key, enabled func() bool) *spyAction {
	spy := &spyAction{}
	km.AddAction(engine.Action{
		Key:        key,
		DefaultKey: defaultKey,
		Press:      func() { spy.pressed++ },
		Release:    func() { spy.released++ },
		Enabled:    enabled,
	})

	return spy
}

// newFixtureKeymapper registers one QuickSave action with the given default.
func newFixtureKeymapper(t *testing.T, defaultKey engine.Key) *engine.Keymapper {
	t.Helper()

	km := engine.NewKeymapper()
	registerSpy(km, "QuickSave", defaultKey, nil)
	km.CommitActions()

	return km
}

// keyNameRoundTrip resolves a name to a key and back, reporting whether both
// directions succeeded.
func keyNameRoundTrip(name string) (string, bool) {
	key, ok := engine.KeyID(name)
	if !ok {
		return "", false
	}

	return engine.KeyName(key)
}

// mustKeyID resolves a key name in the ported table or fails the test.
func mustKeyID(t *testing.T, name string) engine.Key {
	t.Helper()

	key, ok := engine.KeyID(name)
	if !ok {
		t.Fatalf("KeyID(%q) not found", name)
	}

	return key
}

// actionKeys projects the action identities in list order.
func actionKeys(actions []*engine.Action) []string {
	keys := make([]string, len(actions))
	for i, action := range actions {
		keys[i] = action.Key
	}

	return keys
}

// assertBound checks an action's current bind.
func assertBound(t *testing.T, km *engine.Keymapper, action string, want engine.Key) {
	t.Helper()

	registered := km.Action(action)
	if registered == nil {
		t.Fatalf("action %q is not registered", action)
	}

	if registered.BoundKey != want {
		gotName, _ := engine.KeyName(registered.BoundKey)

		t.Fatalf("%q bound key = %q (%d), want %d", action, gotName, registered.BoundKey, want)
	}
}

// assertBoundName checks an action's bind against a key NAME ("" = unbound).
func assertBoundName(t *testing.T, km *engine.Keymapper, action, want string) {
	t.Helper()

	if want == "" {
		assertBound(t, km, action, engine.KeyUnknown)

		return
	}

	assertBound(t, km, action, mustKeyID(t, want))
}

// keyTableCases lists the keycode identities the ini format depends on.
func keyTableCases(t *testing.T) []keyTableCase {
	t.Helper()

	return []keyTableCase{
		{name: "TAB", want: 9},
		{name: "SPACE", want: ' '},
		{name: "LALT", want: mustKeyID(t, "LALT")},
		{name: "PRINT", want: mustKeyID(t, "PRINT")},
		{name: "MMOUSE", want: 2 | engine.KeymapperMouseButtonMask},
		{name: "SCROLLUPMOUSE", want: engine.MouseScrollUpButton},
		{name: "SCROLLDOWNMOUSE", want: engine.MouseScrollDownButton},
		{name: "SCROLLLEFTMOUSE", want: engine.MouseScrollLeftButton},
		{name: "SCROLLRIGHTMOUSE", want: engine.MouseScrollRightButton},
	}
}

// toAsciiUpperCases lists the letter-folding cases.
func toAsciiUpperCases() []toAsciiUpperCase {
	return []toAsciiUpperCase{
		{name: "a folds to A", in: 'a', want: 'A'},
		{name: "z folds to Z", in: 'z', want: 'Z'},
		{name: "an upper letter is unchanged", in: 'A', want: 'A'},
		{name: "a digit is unchanged", in: '7', want: '7'},
		{name: "a mouse key is unchanged", in: engine.MouseScrollUpButton, want: engine.MouseScrollUpButton},
	}
}

// entryRuleCases lists the entry-predicate boundaries.
func entryRuleCases(t *testing.T) []entryRuleCase {
	t.Helper()

	return []entryRuleCase{
		{name: "escape", key: 27, text: true},
		{name: "return", key: 13, text: true},
		{name: "backspace", key: 8, text: true, number: true},
		{name: "space", key: ' ', text: true},
		{name: "a lower letter", key: 'z', text: true},
		{name: "a digit", key: '5', text: true, number: true},
		{name: "a function key", key: mustKeyID(t, "F2")},
		{name: "print", key: mustKeyID(t, "PRINT")},
	}
}

// loadRuleCases lists the four LoadFromIni branches.
func loadRuleCases(t *testing.T) []loadRuleCase {
	t.Helper()

	return []loadRuleCase{
		{name: "absent key keeps the default", doc: "[Other]\nX=1\n", want: "F2"},
		{name: "empty value unbinds", doc: "[Keymapping]\nQuickSave=\n", want: ""},
		{name: "a known name rebinds", doc: "[Keymapping]\nQuickSave=F6\n", want: "F6"},
		{name: "an unknown name falls back", doc: "[Keymapping]\nQuickSave=NOSUCHKEY\n", want: "F2"},
	}
}

// pressGuardCases lists the Press guard order: unmapped, no handler, disabled,
// chat-open, then the firing cases.
func pressGuardCases(t *testing.T) []pressGuardCase {
	t.Helper()

	return []pressGuardCase{
		{
			name:      "unmapped key does nothing",
			register:  func(km *engine.Keymapper) *spyAction { return registerSpy(km, "A", mustKeyID(t, "F2"), nil) },
			bind:      "F2",
			key:       mustKeyID(t, "F3"),
			wantCalls: 0,
		},
		{
			name: "no press handler does nothing",
			register: func(km *engine.Keymapper) *spyAction {
				spy := registerSpy(km, "A", mustKeyID(t, "F2"), nil)
				km.Action("A").Press = nil

				return spy
			},
			bind:      "F2",
			key:       mustKeyID(t, "F2"),
			wantCalls: 0,
		},
		{
			name: "a disabled action does nothing",
			register: func(km *engine.Keymapper) *spyAction {
				return registerSpy(km, "A", mustKeyID(t, "F2"), func() bool { return false })
			},
			bind:      "F2",
			key:       mustKeyID(t, "F2"),
			wantCalls: 0,
		},
		{
			name:      "the chat flag swallows the press",
			register:  func(km *engine.Keymapper) *spyAction { return registerSpy(km, "A", mustKeyID(t, "F2"), nil) },
			bind:      "F2",
			key:       mustKeyID(t, "F2"),
			flags:     engine.InputFlags{Chat: true},
			wantCalls: 0,
		},
		{
			name:      "an enabled unmuted press fires",
			register:  func(km *engine.Keymapper) *spyAction { return registerSpy(km, "A", mustKeyID(t, "F2"), nil) },
			bind:      "F2",
			key:       mustKeyID(t, "F2"),
			wantCalls: 1,
		},
		{
			name:      "a lowercase letter finds the upper-case bind",
			register:  func(km *engine.Keymapper) *spyAction { return registerSpy(km, "A", mustKeyID(t, "S"), nil) },
			bind:      "S",
			key:       's',
			wantCalls: 1,
		},
	}
}

// releaseCarveOutCases lists the chat and gold-entry release rules.
func releaseCarveOutCases(t *testing.T) []releaseCarveOutCase {
	t.Helper()

	return []releaseCarveOutCase{
		{name: "chat swallows a letter", key: mustKeyID(t, "A"), keyName: "A", flags: engine.InputFlags{Chat: true}},
		{name: "chat swallows space", key: mustKeyID(t, "SPACE"), keyName: "SPACE", flags: engine.InputFlags{Chat: true}},
		{
			name: "chat does NOT swallow a function key", key: mustKeyID(t, "F2"), keyName: "F2",
			flags: engine.InputFlags{Chat: true}, wantCalls: 1,
		},
		{name: "gold entry swallows a digit", key: mustKeyID(t, "5"), keyName: "5", flags: engine.InputFlags{DropGold: true}},
		{
			name: "gold entry swallows backspace", key: mustKeyID(t, "BACKSPACE"), keyName: "BACKSPACE",
			flags: engine.InputFlags{DropGold: true},
		},
		{
			name: "gold entry does NOT swallow a letter", key: mustKeyID(t, "S"), keyName: "S",
			flags: engine.InputFlags{DropGold: true}, wantCalls: 1,
		},
		{name: "no flags releases", key: mustKeyID(t, "S"), keyName: "S", wantCalls: 1},
	}
}
