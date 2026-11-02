package engine_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: Source/utils/ini.cpp:151-160 (comment span includes interior
// blank lines), ini.cpp:206 (value skips ' ' and '\t'), ini.cpp:162-168
// (section errors), options.cpp:166/180 (changed()/markAsUnchanged).

func TestIniCommentSpanKeepsInteriorBlankLines(t *testing.T) {
	// given a comment block with a blank line before the key
	in, err := engine.ParseIni([]byte("; head\n\n[s]\nk=1\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when serializing
	got := in.Serialize()

	// then the blank line survives inside the comment span
	want := "; head\r\n\r\n[s]\r\nk=1\r\n"
	if got != want {
		t.Errorf("serialize = %q; want %q", got, want)
	}
}

func TestIniValueTrimSkipsTabs(t *testing.T) {
	// given a value indented by a tab (ini.cpp:206)
	in, err := engine.ParseIni([]byte("[s]\nk=\tvalue\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when reading the value
	got := in.GetString("s", "k", "")

	// then the tab was trimmed like upstream
	if got != "value" {
		t.Errorf("s.k = %q; want value", got)
	}
}

func TestIniParseErrorBranches(t *testing.T) {
	// given inputs that trip each upstream error branch
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"unclosed section", "[s\nk=1\n", "unclosed section name"},
		{"content after section", "[s] junk\n", "content after section"},
		{"key without value", "[s]\nk\n", "has no value"},
		{"key outside section", "k=1\n", "key not in any section"},
	} {
		// when parsing
		_, err := engine.ParseIni([]byte(tc.in))

		// then the upstream error text is returned
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v; want containing %q", tc.name, err, tc.want)
		}
	}
}

func TestIniFreshParseStartsUnchanged(t *testing.T) {
	// given a freshly parsed document
	in, err := engine.ParseIni([]byte("[s]\nk=1\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when reading the dirty flag
	got := in.Changed()

	// then it starts unchanged
	if got {
		t.Error("fresh parse reported changed")
	}
}

func TestIniSetMarksChangedUntilCleared(t *testing.T) {
	// given a parsed document
	in, err := engine.ParseIni([]byte("[s]\nk=1\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when a set mutates it and the save is acknowledged
	in.SetString("s", "k", "2")
	in.MarkAsUnchanged()

	// then the flag is cleared like the upstream save flow
	if in.Changed() {
		t.Error("markAsUnchanged did not clear")
	}
}

func TestIniCommentNormalizationKeepsCRLF(t *testing.T) {
	// given a comment already containing CRLF and a lone LF
	in, err := engine.ParseIni([]byte("; a\r\n; b\n[s]\nk=1\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when serializing
	got := in.Serialize()

	// then lone \n gains \r but CRLF pairs stay intact (ini.cpp:389)
	want := "; a\r\n; b\r\n[s]\r\nk=1\r\n"
	if got != want {
		t.Errorf("serialize = %q; want %q", got, want)
	}
}
