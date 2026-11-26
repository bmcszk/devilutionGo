package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// The pref-path resolver MOVED here from cmd/devilutionGo-migrate (ADR-004:
// portable.ini marker -> XDG_DATA_HOME -> ~/.local/share/<org>/<app>; upstream
// equivalent Source/utils/paths.cpp PrefPath, read via
// Source/options.cpp:136 GetIniPath). Every case is driven through the injected
// SettingsEnv, so no process state is mutated.

func TestSettingsDirPortableMarkerWins(t *testing.T) {
	// given a working dir holding the portable.ini marker
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, engine.PortableMarkerName), nil, 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	env := engine.SettingsEnv{
		Cwd:         cwd,
		Home:        t.TempDir(),
		XDGDataHome: filepath.Join(t.TempDir(), "xdg"),
	}

	// when resolving the settings directory and path
	dir, err := engine.SettingsDir(env)
	if err != nil {
		t.Fatalf("SettingsDir: %v", err)
	}

	path, err := engine.SettingsPath(env)
	if err != nil {
		t.Fatalf("SettingsPath: %v", err)
	}

	// then both land in the working dir
	if dir != cwd {
		t.Fatalf("SettingsDir = %q, want the working dir %q", dir, cwd)
	}

	if want := filepath.Join(cwd, engine.SettingsIniName); path != want {
		t.Fatalf("SettingsPath = %q, want %q", path, want)
	}
}

func TestSettingsDirXDGDataHomeOverridesTheHomeFallback(t *testing.T) {
	// given XDG_DATA_HOME set, no portable marker
	xdg := filepath.Join(t.TempDir(), "xdg")
	env := engine.SettingsEnv{Cwd: t.TempDir(), Home: t.TempDir(), XDGDataHome: xdg}

	// when resolving
	dir, err := engine.SettingsDir(env)
	if err != nil {
		t.Fatalf("SettingsDir: %v", err)
	}

	// then the pref dir hangs off XDG_DATA_HOME
	want := filepath.Join(xdg, "diasurgical", "devilution")
	if dir != want {
		t.Fatalf("SettingsDir = %q, want %q", dir, want)
	}
}

func TestSettingsDirHomeFallbackWithoutXDG(t *testing.T) {
	// given no XDG_DATA_HOME and no portable marker
	home := t.TempDir()
	env := engine.SettingsEnv{Cwd: t.TempDir(), Home: home}

	// when resolving
	dir, err := engine.SettingsDir(env)
	if err != nil {
		t.Fatalf("SettingsDir: %v", err)
	}

	// then the POSIX pref layout under home is used
	want := filepath.Join(home, ".local", "share", "diasurgical", "devilution")
	if dir != want {
		t.Fatalf("SettingsDir = %q, want %q", dir, want)
	}
}

func TestSettingsDirRejectsAnEmptyEnvironment(t *testing.T) {
	// when resolving the settings dir with no working directory and no home
	_, err := engine.SettingsDir(engine.SettingsEnv{})

	// then the missing input is reported instead of a relative path
	if err == nil {
		t.Fatal("SettingsDir accepted an empty environment")
	}
}

func TestSettingsDirRejectsAMissingHomeWithoutAMarker(t *testing.T) {
	// when resolving the settings dir with a marker-free cwd and no home
	_, err := engine.SettingsDir(engine.SettingsEnv{Cwd: t.TempDir()})

	// then the missing home is reported
	if err == nil {
		t.Fatal("SettingsDir accepted an environment with no home")
	}
}

func TestCurrentSettingsEnvReadsTheAmbientValues(t *testing.T) {
	// given the ambient working dir, home and XDG override
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "xdg"))
	t.Chdir(t.TempDir())

	// when reading the ambient inputs
	env, err := engine.CurrentSettingsEnv()
	if err != nil {
		t.Fatalf("CurrentSettingsEnv: %v", err)
	}

	// then they are the live process values
	if env.Cwd == "" || env.Home == "" || env.XDGDataHome == "" {
		t.Fatalf("CurrentSettingsEnv = %+v, want all three populated", env)
	}
}
