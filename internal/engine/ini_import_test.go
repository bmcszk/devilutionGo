package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// configFixtureDir holds committed diablo.ini goldens; no test reads the
// upstream C++ checkout (zero-dependency parity, AGENTS.md).
const configFixtureDir = "../testdata/config"

// Reference: diasurgical/DevilutionX test/ini_test.cpp (TEST(IniTest, BasicTest)).
// The parse fixture is byte-identical to that test's input raw string.
func TestIniBasicParse(t *testing.T) {
	// given the upstream ini_test.cpp fixture
	in := mustParseIni(t, readConfigFixture(t, "ini_basic.ini"))

	// when reading the parsed scalars
	num, numErr := in.GetInt("sectionA", "int", 0)
	flt, fltErr := in.GetFloat("sectionA", "float", 0)

	// then every upstream expectation holds
	assertInt(t, "sectionA.int", num, numErr, -3)
	assertFloat(t, "sectionA.float", flt, fltErr, 2.5)
	assertString(t, in, "sectionA", "key1", "value1")
	assertString(t, in, "sectionA", "key2", "value2")
	assertString(t, in, "sectionB", "key", "value")
}

func TestIniBasicParseKeepsDuplicateKeyValues(t *testing.T) {
	// given the upstream fixture with "multi" set twice
	in := mustParseIni(t, readConfigFixture(t, "ini_basic.ini"))

	// when reading the multi-valued key
	values := in.Get("sectionA", "multi")

	// then both values survive in document order
	if len(values) != 2 || values[0].Value != "a" || values[1].Value != "b" {
		t.Fatalf("multi = %+v; want [a b]", values)
	}
}

func TestIniSetValuesSerialize(t *testing.T) {
	// given the fixture with upstream's edits applied
	in := mustParseIni(t, readConfigFixture(t, "ini_basic.ini"))
	applyUpstreamEdits(in)

	// when
	got := in.Serialize()

	// then the bytes equal upstream's expected serialization
	want := string(readConfigFixture(t, "ini_basic_serialized.ini"))
	if got != want {
		t.Fatalf("Serialize mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestIniSerializeIsIdempotent(t *testing.T) {
	// given the expected serialized bytes
	golden := string(readConfigFixture(t, "ini_basic_serialized.ini"))

	// when parsing and re-serializing them
	got := mustParseIni(t, []byte(golden)).Serialize()

	// then nothing changes (properties: parse->serialize is a fixed point)
	if got != golden {
		t.Fatalf("round-trip changed bytes\n got: %q\nwant: %q", got, golden)
	}
}

func TestIniSerializeRoundTripsRepeatedly(t *testing.T) {
	// given the expected serialized bytes
	data := readConfigFixture(t, "ini_basic_serialized.ini")

	// when re-parsing and re-serializing three times
	var last string
	for range 3 {
		last = mustParseIni(t, data).Serialize()
		data = []byte(last)
	}

	// then the third pass still matches the golden bytes
	if want := string(readConfigFixture(t, "ini_basic_serialized.ini")); last != want {
		t.Fatalf("third pass drifted\n got: %q\nwant: %q", last, want)
	}
}

func TestIniParseRejectsKeyOutsideSection(t *testing.T) {
	// given a key before any section header
	data := []byte("key=value\n")

	// when
	_, err := engine.ParseIni(data)

	// then parsing fails like Ini::parse (Source/utils/ini.cpp:210)
	if err == nil || !strings.Contains(err.Error(), "key not in any section") {
		t.Fatalf("err = %v; want key-not-in-section error", err)
	}
}

func TestIniParseSkipsUTF8BOM(t *testing.T) {
	// given a document with a UTF-8 byte order mark
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("[s]\nk=v\n")...)

	// when
	in := mustParseIni(t, data)

	// then the BOM is skipped (Source/utils/ini.cpp:63) and the key is readable
	if got := in.GetString("s", "k", ""); got != "v" {
		t.Fatalf("s.k = %q; want v", got)
	}
}

func readConfigFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(configFixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func mustParseIni(t *testing.T, data []byte) *engine.Ini {
	t.Helper()
	ini, err := engine.ParseIni(data)
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}
	return ini
}

// applyUpstreamEdits replays the set() block of test/ini_test.cpp BasicTest.
func applyUpstreamEdits(in *engine.Ini) {
	in.SetString("newSection", "newKey", "hello")
	in.SetString("sectionA", "key1", "newValue")
	in.SetInt("sectionA", "int", 1337)
	in.SetBool("sectionA", "bool yes", false)
	in.SetStrings("sectionA", "multi", []string{"x", "y", "z"})
	in.SetFloat("sectionA", "float", 10.5)
}

// assertInt checks one GetInt result against its upstream expectation.
func assertInt(t *testing.T, name string, got int, err error, want int) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if got != want {
		t.Errorf("%s = %d; want %d", name, got, want)
	}
}

// assertFloat checks one GetFloat result against its upstream expectation.
func assertFloat(t *testing.T, name string, got float64, err error, want float64) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if got < want-0.001 || got > want+0.001 {
		t.Errorf("%s = %v; want %v", name, got, want)
	}
}

// assertString checks one GetString result against its upstream expectation.
func assertString(t *testing.T, in *engine.Ini, section, key, want string) {
	t.Helper()
	if got := in.GetString(section, key, ""); got != want {
		t.Errorf("%s.%s = %q; want %q", section, key, got, want)
	}
}
