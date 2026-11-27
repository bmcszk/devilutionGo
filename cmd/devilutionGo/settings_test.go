package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// The boot's settings document (Options slice S1): diablo.ini read from
// engine.SettingsDir, the four keys the game menu drives written back on
// menu close. Reference: Source/options.cpp:144-183 (LoadIni/SaveIni),
// :221-280 (LoadOptions/SaveOptions), Source/options.h:905-920
// (GetCategories order: Graphics, Audio, Game).

func TestSettingsMissingFileBootsOnDefaults(t *testing.T) {
	// given a settings dir with no diablo.ini
	dir := t.TempDir()

	// when the boot reads its settings
	got, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	// then it runs on the upstream defaults (a missing file is not fatal —
	// recorded deviation from upstream, which app_fatal()s only on a PARSE
	// error, Source/options.cpp:151-162)
	if want := engine.DefaultUserOptions(); got.user != want {
		t.Fatalf("user = %+v, want defaults %+v", got.user, want)
	}

	if want := engine.DefaultSimOptions(); got.sim != want {
		t.Fatalf("sim = %+v, want defaults %+v", got.sim, want)
	}
}

func TestSettingsRoundTripThroughTheIni(t *testing.T) {
	// given a first boot on a missing file
	dir := t.TempDir()

	boot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	// when the player moves the three knobs and the rate, and the menu close
	// persists them (X:gmenu.cpp:246-248 SaveOptions)
	boot.user = engine.UserOptions{SoundVolume: -600, MusicVolume: -300, Brightness: 50}
	boot.sim.TickRate = 30
	boot.sim.SampleRate = 44100

	if err := saveUserSettings(dir, boot); err != nil {
		t.Fatalf("saveUserSettings: %v", err)
	}

	// then a re-boot reads the same values back
	reboot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("re-loadUserSettings: %v", err)
	}

	if reboot.user != boot.user {
		t.Fatalf("user after reboot = %+v, want %+v (no setting survived reboot)", reboot.user, boot.user)
	}

	if reboot.sim != boot.sim {
		t.Fatalf("sim after reboot = %+v, want %+v (no setting survived reboot)", reboot.sim, boot.sim)
	}
}

func TestSettingsWrittenFileMatchesTheUpstreamShape(t *testing.T) {
	// given a first boot on a missing file
	dir := t.TempDir()

	boot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	// when the menu close persists non-default values
	boot.user = engine.UserOptions{SoundVolume: -600, MusicVolume: -300, Brightness: 50}
	boot.sim.TickRate = 30

	if err := saveUserSettings(dir, boot); err != nil {
		t.Fatalf("saveUserSettings: %v", err)
	}

	// then the four keys land under the upstream section and key names in
	// GetCategories order with a CRLF separator (Source/options.h:905-920,
	// Source/utils/ini.cpp:415)
	data, err := os.ReadFile(filepath.Join(dir, engine.SettingsIniName))
	if err != nil {
		t.Fatalf("read written ini: %v", err)
	}

	want := strings.Join([]string{
		"[Graphics]",
		"Brightness Correction=50",
		"",
		"[Audio]",
		"Sound Volume=-600",
		"Music Volume=-300",
		"Sample Rate=22050",
		"",
		"[Game]",
		"Speed=30",
		"",
	}, "\r\n")

	if got := string(data); got != want {
		t.Fatalf("written ini =\n%q\nwant\n%q", got, want)
	}
}

func TestSettingsCleanFileIsNotRewritten(t *testing.T) {
	// given a written file that already carries the boot's values
	dir := writeSettingsFixture(t)

	path := filepath.Join(dir, engine.SettingsIniName)

	before := statSettings(t, path)

	// when a re-boot saves again with no value moved
	reboot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("re-loadUserSettings: %v", err)
	}

	if err := saveUserSettings(dir, reboot); err != nil {
		t.Fatalf("saveUserSettings (clean): %v", err)
	}

	// then nothing was written: SaveIni returns early on an unchanged document
	// (Source/options.cpp:169)
	after := statSettings(t, path)

	if !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() {
		t.Fatalf("clean save rewrote the file: mtime %v -> %v, size %d -> %d",
			before.ModTime(), after.ModTime(), before.Size(), after.Size())
	}
}

func TestSettingsUpstreamWrittenFileLoadsBack(t *testing.T) {
	// given a diablo.ini as a devilutionX build writes the four keys
	dir := t.TempDir()
	upstream := "[Audio]\r\nSound Volume=-800\r\nMusic Volume=-200\r\nSample Rate=22050\r\n\r\n" +
		"[Game]\r\nSpeed=40\r\n\r\n[Graphics]\r\nBrightness Correction=25\r\n"

	if err := os.WriteFile(filepath.Join(dir, engine.SettingsIniName), []byte(upstream), 0o644); err != nil {
		t.Fatalf("write upstream ini: %v", err)
	}

	// when the boot reads it
	got, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	// then every key is picked up
	wantUser := engine.UserOptions{SoundVolume: -800, MusicVolume: -200, Brightness: 25}
	if got.user != wantUser {
		t.Fatalf("user = %+v, want %+v", got.user, wantUser)
	}

	if got.sim.TickRate != 40 || got.sim.SampleRate != 22050 {
		t.Fatalf("sim = %+v, want TickRate 40 / SampleRate 22050", got.sim)
	}
}

func TestSettingsUnchangedSaveLeavesAnUpstreamFileByteIdentical(t *testing.T) {
	// given a devilutionX-shaped file
	dir := t.TempDir()
	upstream := "[Audio]\r\nSound Volume=-800\r\nMusic Volume=-200\r\nSample Rate=22050\r\n\r\n" +
		"[Game]\r\nSpeed=40\r\n\r\n[Graphics]\r\nBrightness Correction=25\r\n"

	path := filepath.Join(dir, engine.SettingsIniName)
	if err := os.WriteFile(path, []byte(upstream), 0o644); err != nil {
		t.Fatalf("write upstream ini: %v", err)
	}

	boot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	// when saving the unchanged values
	if err := saveUserSettings(dir, boot); err != nil {
		t.Fatalf("saveUserSettings: %v", err)
	}

	// then the file is untouched
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-read ini: %v", err)
	}

	if string(after) != upstream {
		t.Fatalf("unchanged save rewrote the upstream file:\n%q\nwant\n%q", after, upstream)
	}
}

func TestSettingsParseErrorIsReported(t *testing.T) {
	// given a diablo.ini carrying an unparsable line
	dir := t.TempDir()
	path := filepath.Join(dir, engine.SettingsIniName)
	if err := os.WriteFile(path, []byte("not a key value line\n"), 0o644); err != nil {
		t.Fatalf("write broken ini: %v", err)
	}

	// when the boot reads it
	_, err := loadUserSettings(dir)

	// then the parse failure reaches the caller (upstream app_fatal()s here,
	// Source/options.cpp:151-162; this port reports it instead)
	if err == nil {
		t.Fatal("loadUserSettings accepted an unparsable ini")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("err = %v, want a parse failure", err)
	}
}

func TestSettingsSaveWithoutADocumentIsANoOp(t *testing.T) {
	// given a boot that never loaded a document
	dir := t.TempDir()

	// when saving
	if err := saveUserSettings(dir, runtimeSettings{}); err != nil {
		t.Fatalf("saveUserSettings: %v", err)
	}

	// then no file appeared
	if _, err := os.Stat(filepath.Join(dir, engine.SettingsIniName)); !os.IsNotExist(err) {
		t.Fatalf("stat = %v, want no written ini", err)
	}
}

// writeSettingsFixture persists one non-default settings document into a fresh
// temp dir and returns that dir.
func writeSettingsFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	boot, err := loadUserSettings(dir)
	if err != nil {
		t.Fatalf("loadUserSettings: %v", err)
	}

	boot.user = engine.UserOptions{SoundVolume: -600, MusicVolume: -300, Brightness: 50}

	if err := saveUserSettings(dir, boot); err != nil {
		t.Fatalf("saveUserSettings: %v", err)
	}

	return dir
}

// statSettings stats the settings file or fails the test.
func statSettings(t *testing.T, path string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}

	return info
}
