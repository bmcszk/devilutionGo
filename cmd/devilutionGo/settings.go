package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// runtimeSettings is the boot-owned settings document: the parsed diablo.ini
// plus the typed values read out of it. Saving writes the values back into the
// SAME document, so comments and unknown keys survive a round trip
// (Source/utils/ini.cpp:415 serialize).
//
// Untagged on purpose: the device-lane tests drive the boot and must stay
// asset- and display-free.
type runtimeSettings struct {
	ini  *engine.Ini
	sim  engine.SimOptions
	user engine.UserOptions
}

// loadUserSettings reads dir/diablo.ini and returns the typed tree. A missing
// file boots on the upstream defaults and starts an empty document: a recorded
// deviation — upstream app_fatal()s only on a PARSE error, never on a missing
// file (Source/options.cpp:144-162).
func loadUserSettings(dir string) (runtimeSettings, error) {
	path := filepath.Join(dir, engine.SettingsIniName)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return newRuntimeSettings()
	}

	if err != nil {
		return runtimeSettings{}, fmt.Errorf("read settings %s: %w", path, err)
	}

	return parseUserSettings(data)
}

// newRuntimeSettings starts an empty document on the upstream defaults
// (Source/options.cpp:514-516, :796, :843).
func newRuntimeSettings() (runtimeSettings, error) {
	in, err := engine.ParseIni(nil)
	if err != nil {
		return runtimeSettings{}, fmt.Errorf("parse empty settings: %w", err)
	}

	return runtimeSettings{ini: in, sim: engine.DefaultSimOptions(), user: engine.DefaultUserOptions()}, nil
}

// parseUserSettings parses a diablo.ini document and reads the typed tree out
// of it (Source/options.cpp:221-252 LoadOptions).
func parseUserSettings(data []byte) (runtimeSettings, error) {
	in, err := engine.ParseIni(data)
	if err != nil {
		return runtimeSettings{}, fmt.Errorf("parse settings: %w", err)
	}

	return runtimeSettings{ini: in, sim: engine.LoadSimOptions(in), user: engine.LoadUserOptions(in)}, nil
}

// saveUserSettings writes the typed tree back into dir/diablo.ini, but only
// when the document changed, then marks it unchanged (SaveIni,
// Source/options.cpp:167-183).
func saveUserSettings(dir string, s runtimeSettings) error {
	if s.ini == nil {
		return nil
	}

	// GetCategories order puts Graphics before Audio before Game, so writing in
	// this order gives a fresh file the upstream section order
	// (Source/options.h:905-920).
	s.user.SaveUserOptions(s.ini)
	s.sim.SaveSimOptions(s.ini)

	if !s.ini.Changed() {
		return nil
	}

	path := filepath.Join(dir, engine.SettingsIniName)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create settings dir %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(s.ini.Serialize()), 0o644); err != nil {
		return fmt.Errorf("write settings %s: %w", path, err)
	}

	s.ini.MarkAsUnchanged()

	return nil
}
