package main

import (
	"strings"
	"testing"
)

func TestLookupRefresherResolvesSubcommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"dun takes two paths", []string{"dun", "artifacts", "corpus"}, "dun"},
		{"dpiece takes two paths", []string{"dpiece", "artifacts", "corpus"}, "dpiece"},
		{"town takes two paths", []string{"town", "artifacts", "corpus"}, "town"},
		{"til takes two paths", []string{"til", "raw", "assets"}, "til"},
		{"assets takes two paths", []string{"assets", "mpq", "corpus"}, "assets"},
		{"town-inputs takes two paths", []string{"town-inputs", "raw", "corpus"}, "town-inputs"},
		{"gamedata takes no arguments", []string{"gamedata"}, "gamedata"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			ref, ok := lookupRefresher(tc.args)

			// then
			if !ok || ref.name != tc.want {
				t.Fatalf("lookupRefresher(%q) = (%q, %v), want (%q, true)", tc.args, ref.name, ok, tc.want)
			}
		})
	}
}

func TestLookupRefresherRejectsOtherArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"bare dun is short an argument", []string{"dun"}},
		{"dun with three paths", []string{"dun", "a", "b", "c"}},
		{"til is short an argument", []string{"til"}},
		{"assets with three paths", []string{"assets", "a", "b", "c"}},
		{"gamedata with a path", []string{"gamedata", "corpus"}},
		{"unknown name", []string{"affix", "a", "b"}},
		{"no arguments", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			ref, ok := lookupRefresher(tc.args)

			// then
			if ok {
				t.Fatalf("lookupRefresher(%q) resolved to %q, want rejection", tc.args, ref.name)
			}
		})
	}
}

func TestErrUsageNamesEverySubcommand(t *testing.T) {
	// when
	err := errUsage()

	// then
	if err == nil {
		t.Fatal("errUsage() = nil, want an error")
	}
	message := err.Error()
	if strings.Contains(message, "\n") {
		t.Fatalf("usage spans %d lines, want one: %q", strings.Count(message, "\n")+1, message)
	}
	for _, name := range []string{"dun", "dpiece", "town", "til", "assets", "town-inputs", "gamedata"} {
		if !strings.Contains(message, name) {
			t.Fatalf("usage %q does not name subcommand %q", message, name)
		}
	}
}
