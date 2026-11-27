package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Settings file names and the SDL pref path components
// (Source/options.cpp:136 GetIniPath; Source/utils/paths.cpp PrefPath).
const (
	// SettingsIniName is the settings file upstream reads.
	SettingsIniName = "diablo.ini"
	// PortableMarkerName is the portable-install marker per ADR-004
	// ("portable.ini supported"); upstream instead treats a writeable
	// diablo.ini in the working dir as the marker (paths.cpp PrefPath).
	PortableMarkerName = "portable.ini"
	// settingsOrg and settingsApp are the SDL_GetPrefPath components.
	settingsOrg = "diasurgical"
	settingsApp = "devilution"
)

// SettingsEnv is the ambient input the pref-path resolver reads. Injecting it
// keeps the resolver testable without mutating process state.
type SettingsEnv struct {
	// Cwd is the working directory the portable.ini marker is probed in.
	Cwd string
	// Home is the user's home directory the POSIX pref layout hangs off.
	Home string
	// XDGDataHome is the XDG_DATA_HOME override; empty means "use Home".
	XDGDataHome string
}

// CurrentSettingsEnv reads the live process values SettingsEnv carries.
func CurrentSettingsEnv() (SettingsEnv, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return SettingsEnv{}, fmt.Errorf("working dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return SettingsEnv{}, fmt.Errorf("home dir: %w", err)
	}

	return SettingsEnv{Cwd: cwd, Home: home, XDGDataHome: os.Getenv("XDG_DATA_HOME")}, nil
}

// SettingsDir resolves the directory the engine reads diablo.ini from
// (ADR-004): the working dir when it holds the portable.ini marker, else the
// SDL pref dir SDL_GetPrefPath("diasurgical", "devilution")
// (Source/utils/paths.cpp PrefPath).
//
// POSIX-only pref layout: macOS and Windows resolve differently upstream — a
// recorded deviation carried over from the migrate CLI.
func SettingsDir(env SettingsEnv) (string, error) {
	if env.Cwd == "" {
		return "", errors.New("settings dir: no working directory")
	}

	if _, err := os.Stat(filepath.Join(env.Cwd, PortableMarkerName)); err == nil {
		return env.Cwd, nil
	}

	if env.Home == "" {
		return "", errors.New("settings dir: no home directory")
	}

	base := filepath.Join(env.Home, ".local", "share")
	if env.XDGDataHome != "" {
		base = env.XDGDataHome
	}

	return filepath.Join(base, settingsOrg, settingsApp), nil
}

// SettingsPath is the diablo.ini path inside SettingsDir
// (Source/options.cpp:136).
func SettingsPath(env SettingsEnv) (string, error) {
	dir, err := SettingsDir(env)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, SettingsIniName), nil
}
