package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// keymappingFixture is the committed defaults table: one row per
// InitKeymapActions entry (Source/diablo.cpp:1839-2157), transcribed by the
// corpus-measurement law. The row count is measured from the fixture at run
// time, never written as a literal.
const keymappingFixture = "keymapping_defaults.tsv"

// keymappingRow is one fixture row: the action's ini key and its default key
// name (empty for SDLK_UNKNOWN).
type keymappingRow struct {
	Action string
	Key    string
}

func TestKeymapperDefaultsMatchTheUpstreamTable(t *testing.T) {
	// given the committed InitKeymapActions defaults
	rows := loadKeymappingFixtures(t)
	km := buildDefaultsKeymapper(t, rows)

	// when loading from a document with NO [Keymapping] section at all
	km.LoadFromIni(mustParseIni(t, nil))

	// then every action is bound to exactly the upstream default
	// (Action::LoadFromIni's absent branch, Source/options.cpp:1213-1216)
	for _, row := range rows {
		t.Run(row.Action, func(t *testing.T) { assertFixtureBind(t, km, row) })
	}
}

func TestKeymapperDefaultsAreUnique(t *testing.T) {
	// Each row's claim is judged on its own: the phases live in the subtest.
	rows := loadKeymappingFixtures(t)

	for i, row := range rows {
		t.Run(row.Action, func(t *testing.T) {
			// given this row and every row after it
			later := rows[i+1:]

			// when checking the later rows for a repeat
			repeatedAction, repeatedKey := firstRepeat(later, row)

			// then neither the action name nor its default key repeats
			if repeatedAction != "" || repeatedKey != "" {
				t.Fatalf("fixture repeats a claim on row %q: action %q, key %q",
					row.Action, repeatedAction, repeatedKey)
			}
		})
	}
}

func TestKeymapperIniRebindsExactlyOneAction(t *testing.T) {
	// given the defaults registered
	rows := loadKeymappingFixtures(t)
	km := buildDefaultsKeymapper(t, rows)

	// when loading [Keymapping] QuickSave=F4 — F4 is the one function key no
	// default claims (F1 Help, F2 QuickSave, F3 QuickLoad, F5-F8 QuickSpell,
	// F9-F12 QuickMessage)
	km.LoadFromIni(mustParseIni(t, []byte("[Keymapping]\nQuickSave=F4\n")))

	// then QuickSave moved to F4 and nothing else did
	assertBound(t, km, "QuickSave", mustKeyID(t, "F4"))

	for _, row := range rows {
		if row.Action == "QuickSave" {
			continue
		}

		t.Run(row.Action, func(t *testing.T) { assertFixtureBind(t, km, row) })
	}
}

func TestKeymapperDefaultsRoundTripThroughTheIni(t *testing.T) {
	// given the defaults registered and loaded
	rows := loadKeymappingFixtures(t)
	km := buildDefaultsKeymapper(t, rows)
	km.LoadFromIni(mustParseIni(t, nil))

	// when saving them and loading the result into a fresh model
	out := mustParseIni(t, nil)
	km.SaveToIni(out)

	reloaded := buildDefaultsKeymapper(t, rows)
	reloaded.LoadFromIni(out)

	// then every bind survived the round trip
	for _, row := range rows {
		assertBound(t, reloaded, row.Action, km.Action(row.Action).BoundKey)
	}
}

// loadKeymappingFixtures reads the committed defaults table.
func loadKeymappingFixtures(t *testing.T) []keymappingRow {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(configFixtureDir, keymappingFixture))
	if err != nil {
		t.Fatalf("read %s: %v", keymappingFixture, err)
	}

	var rows []keymappingRow

	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if row, ok := parseKeymappingRow(line); ok {
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		t.Fatalf("%s carries no rows", keymappingFixture)
	}

	return rows
}

// parseKeymappingRow parses one fixture line; comments and blanks are skipped.
func parseKeymappingRow(line string) (keymappingRow, bool) {
	if line == "" || strings.HasPrefix(line, "#") {
		return keymappingRow{}, false
	}

	action, key, found := strings.Cut(line, "\t")
	if !found {
		return keymappingRow{}, false
	}

	return keymappingRow{Action: action, Key: key}, true
}

// buildDefaultsKeymapper registers one action per fixture row with the
// fixture's default key and no callbacks (the data half of
// InitKeymapActions).
func buildDefaultsKeymapper(t *testing.T, rows []keymappingRow) *engine.Keymapper {
	t.Helper()

	km := engine.NewKeymapper()

	for _, row := range rows {
		km.AddAction(engine.Action{Key: row.Action, DefaultKey: fixtureDefaultKey(t, row)})
	}

	km.CommitActions()

	return km
}

// fixtureDefaultKey resolves a row's default key (KeyUnknown when empty).
func fixtureDefaultKey(t *testing.T, row keymappingRow) engine.Key {
	t.Helper()

	if row.Key == "" {
		return engine.KeyUnknown
	}

	key, ok := engine.KeyID(row.Key)
	if !ok {
		t.Fatalf("fixture names %q for %q, which is not in the key table", row.Key, row.Action)
	}

	return key
}

// assertFixtureBind checks one row's bind against the fixture default.
func assertFixtureBind(t *testing.T, km *engine.Keymapper, row keymappingRow) {
	t.Helper()

	want := fixtureDefaultKey(t, row)

	action := km.Action(row.Action)
	if action == nil {
		t.Fatalf("action %q was not registered", row.Action)
	}

	if action.BoundKey != want {
		gotName, _ := engine.KeyName(action.BoundKey)

		t.Fatalf("bound keys differ: %q bound to %q (%d), want %q (%d)",
			row.Action, gotName, action.BoundKey, row.Key, want)
	}
}

// firstRepeat returns the first later action name and the first later action
// that repeats this row's default key ("" for neither).
func firstRepeat(later []keymappingRow, row keymappingRow) (action, key string) {
	for _, other := range later {
		if other.Action == row.Action {
			return other.Action, ""
		}

		if row.Key != "" && other.Key == row.Key {
			return "", other.Action
		}
	}

	return "", ""
}
