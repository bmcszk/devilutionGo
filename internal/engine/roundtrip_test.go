package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: test/ini_test.cpp round-trip semantics - serialize after a
// no-edit parse must reproduce diablo.ini byte for byte.

func TestIniNoEditRoundTripIsByteIdentical(t *testing.T) {
	// given the committed serialized golden fixture converted to CRLF
	// (the canonical diablo.ini shape upstream serialize emits)
	raw, err := os.ReadFile(filepath.Join("..", "testdata", "config", "ini_basic_serialized.ini"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	crlf := strings.TrimPrefix(string(raw), "\n")
	crlf = strings.ReplaceAll(crlf, "\r\n", "\n")
	crlf = strings.ReplaceAll(crlf, "\n", "\r\n")

	// when parsing and serializing with no edits
	in, err := engine.ParseIni([]byte(crlf))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}
	got := in.Serialize()

	// then the bytes are identical
	if got != crlf {
		t.Errorf("no-edit round trip differs:\n got %q\nwant %q", got, crlf)
	}
}
