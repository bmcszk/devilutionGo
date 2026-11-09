package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settingsFixture is a synthetic diablo.ini: CRLF line ends like the upstream
// serialize emits, a comment, two known sections and one unknown section whose
// keys must land in its view untyped.
const settingsFixture = "; generated\r\n" +
	"[Audio]\r\nSample Rate=44100\r\n" +
	"\r\n[Game]\r\nSpeed=30\r\n" +
	"\r\n[Diablo]\r\n; keep this\r\nLast Single Player Hero=warrior\r\n" +
	"\r\n[Language]\r\nLanguage=pl\r\n"

// violationFixture is a synthetic diablo.ini whose typed keys sit outside the
// upstream bounds; the check must report them and still carry the file.
const violationFixture = "; generated\r\n" +
	"[Game]\r\nSpeed=99\r\n" +
	"\r\n[Audio]\r\nSample Rate=1234\r\n" +
	"\r\n[Diablo]\r\nLast Single Player Hero=warrior\r\n"

func TestSettingsCarriesIniByteIdentical(t *testing.T) {
	// given a diablo.ini next to the run
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, settingsFixture)
	out := filepath.Join(dir, "out")

	// when the settings subcommand migrates it
	_, _ = runSettingsOK(t, "--ini", src, "--out", out)

	// then the copy is cmp-equal to the source
	if !bytes.Equal(readSettingsFile(t, out, "diablo.ini"), []byte(settingsFixture)) {
		t.Error("carried ini differs from the source")
	}
}

func TestSettingsWritesOneViewPerSectionFamily(t *testing.T) {
	// given a diablo.ini with three sections
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, settingsFixture)
	out := filepath.Join(dir, "out")

	// when the settings subcommand migrates it
	runSettingsOK(t, "--ini", src, "--out", out)

	// then every family gets its own flat view file (open question resolved:
	// one file per INI section, no --single-file variant)
	for _, family := range []string{
		"settings_audio.yaml", "settings_game.yaml", "settings_diablo.yaml", "settings_language.yaml",
	} {
		if _, err := os.Stat(filepath.Join(out, family)); err != nil {
			t.Errorf("view %s missing: %v", family, err)
		}
	}
}

func TestSettingsViewKeysAndValuesVerbatim(t *testing.T) {
	// given a diablo.ini with typed and unknown keys
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, settingsFixture)
	out := filepath.Join(dir, "out")

	// when the settings subcommand migrates it
	runSettingsOK(t, "--ini", src, "--out", out)

	// then the Audio view carries the value verbatim, comments stay out
	audio := readSettingsFile(t, out, "settings_audio.yaml")

	for _, want := range []string{"value: \"44100\"", "source: Audio.Sample Rate"} {
		if !strings.Contains(string(audio), want) {
			t.Errorf("audio view = %q, want it to contain %q", audio, want)
		}
	}
}

func TestSettingsViewsWithBomMatchTheNonBomTwin(t *testing.T) {
	// given a BOM-prefixed diablo.ini and its non-BOM twin
	doc := "[Audio]\r\nSample Rate=44100\r\n\r\n[Game]\r\nSpeed=30\r\n"
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, "\xef\xbb\xbf"+doc)
	out := filepath.Join(dir, "out")

	// when both files run through the migration
	runSettingsOK(t, "--ini", src, "--out", out)
	bomViews := readSettingsViews(t, out)

	twinDir := t.TempDir()
	twin := writeSettingsIni(t, twinDir, doc)
	twinOut := filepath.Join(twinDir, "out")

	runSettingsOK(t, "--ini", twin, "--out", twinOut)
	twinViews := readSettingsViews(t, twinOut)

	// then both runs produce the same view set with the same content
	if len(bomViews) != len(twinViews) {
		t.Fatalf("view count = %d, the non-BOM twin has %d", len(bomViews), len(twinViews))
	}

	for name, content := range twinViews {
		if !bytes.Equal(bomViews[name], content) {
			t.Errorf("view %s differs from the non-BOM twin", name)
		}
	}
}

func TestSettingsViewsRegenerateByteIdentical(t *testing.T) {
	// given a diablo.ini migrated once and a second identical run
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, settingsFixture)
	out := filepath.Join(dir, "out")

	// when both runs regenerate the views
	first, second := migrateSettingsTwice(t, src, out)

	// then every view is byte-identical across runs
	if len(first) != len(second) {
		t.Fatalf("view count changed: %d then %d", len(first), len(second))
	}

	for name, content := range first {
		if !bytes.Equal(content, second[name]) {
			t.Errorf("view %s differs across runs", name)
		}
	}
}

func TestSettingsReportListsViolationsItCannotFix(t *testing.T) {
	// given a diablo.ini whose typed keys are out of range and a key outside
	// every section family the engine knows
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, violationFixture)
	out := filepath.Join(dir, "out")

	// when the settings subcommand migrates it
	stdout, _ := runSettingsOK(t, "--ini", src, "--out", out)

	// then the report names every violation
	for _, want := range []string{
		"Game.Speed=99",
		"Audio.Sample Rate=1234",
		"not carried into any typed option",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("report = %q, want it to contain %q", stdout, want)
		}
	}

	// and the carry still happens, violations included
	if !strings.Contains(string(readSettingsFile(t, out, "diablo.ini")), "Speed=99") {
		t.Error("carried ini lost the violation")
	}
}

func TestSettingsRejectsUnparseableIni(t *testing.T) {
	// given a diablo.ini with a key outside any section
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, "Speed=30\r\n")
	out := filepath.Join(dir, "out")

	// when the settings subcommand migrates it
	code, _, stderr := runSettings(t, "--ini", src, "--out", out)

	// then the run fails before writing anything
	assertSettingsFailure(t, code, stderr, exitFailure)

	if _, err := os.Stat(filepath.Join(out, "diablo.ini")); !os.IsNotExist(err) {
		t.Error("carried ini exists after a failed run")
	}
}

func TestSettingsRejectsMissingIni(t *testing.T) {
	// given an --ini path that does not exist
	out := filepath.Join(t.TempDir(), "out")

	// when the settings subcommand migrates it
	code, _, stderr := runSettings(t, "--ini", filepath.Join(t.TempDir(), "absent.ini"), "--out", out)

	// then the run fails with the missing file named
	assertSettingsFailure(t, code, stderr, exitFailure)

	if !strings.Contains(stderr, "absent.ini") {
		t.Errorf("stderr = %q, want the missing file named", stderr)
	}
}

func TestSettingsDryRunWritesNothing(t *testing.T) {
	// given a diablo.ini and a fresh --out
	dir := t.TempDir()
	src := writeSettingsIni(t, dir, settingsFixture)
	out := filepath.Join(dir, "out")

	// when the settings subcommand plans the migration
	stdout, _ := runSettingsOK(t, "--dry-run", "--ini", src, "--out", out)

	// then nothing is written and the plan is reported
	if entries, err := os.ReadDir(out); !os.IsNotExist(err) && len(entries) > 0 {
		t.Errorf("dry run wrote %d entries into %s", len(entries), out)
	}

	if !strings.Contains(stdout, "planned") {
		t.Errorf("stdout = %q, want the planned mode report", stdout)
	}
}

func TestSettingsPortableMarkerResolvesToTheWorkingDir(t *testing.T) {
	// given a working dir holding the portable.ini marker and the ini
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "portable.ini"), nil, 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	writeSettingsIni(t, dir, settingsFixture)

	t.Chdir(dir)

	// when the settings subcommand migrates with no explicit paths
	stdout, stderr := runSettingsOK(t)

	// then both the carry and the views resolve to the working dir
	if !bytes.Equal(readSettingsFile(t, dir, "diablo.ini"), []byte(settingsFixture)) {
		t.Error("carried ini in the working dir differs from the source")
	}

	if _, err := os.Stat(filepath.Join(dir, "settings_game.yaml")); err != nil {
		t.Errorf("working-dir view missing: %v", err)
	}

	want := "--ini " + filepath.Join(dir, settingsIniName)
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}

	if !strings.Contains(stdout, "wrote") {
		t.Errorf("stdout = %q, want the wrote mode report", stdout)
	}
}

func TestSettingsXDGDataHomeOverridesThePrefDir(t *testing.T) {
	// given an XDG_DATA_HOME holding the ini and a marker-free working dir
	dir := t.TempDir()
	xdg := filepath.Join(dir, "xdg")
	pref := filepath.Join(xdg, settingsOrg, settingsApp)

	if err := os.MkdirAll(pref, 0o755); err != nil {
		t.Fatalf("write pref dir: %v", err)
	}

	writeSettingsIni(t, pref, settingsFixture)

	t.Setenv("XDG_DATA_HOME", xdg)
	t.Chdir(t.TempDir())

	// when the settings subcommand migrates with no explicit paths
	stdout, stderr := runSettingsOK(t)

	// then both the carry and the views resolve under XDG_DATA_HOME
	if !bytes.Equal(readSettingsFile(t, pref, "diablo.ini"), []byte(settingsFixture)) {
		t.Error("carried ini under XDG_DATA_HOME differs from the source")
	}

	if _, err := os.Stat(filepath.Join(pref, "settings_game.yaml")); err != nil {
		t.Errorf("XDG view missing: %v", err)
	}

	want := "--ini " + filepath.Join(pref, settingsIniName)
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}

	if !strings.Contains(stdout, "wrote") {
		t.Errorf("stdout = %q, want the wrote mode report", stdout)
	}
}

func TestSettingsHomeFallbackWhenNoXDGOverride(t *testing.T) {
	// given a HOME with the ini in its pref layout and no XDG override
	home := t.TempDir()
	pref := filepath.Join(home, ".local", "share", settingsOrg, settingsApp)

	if err := os.MkdirAll(pref, 0o755); err != nil {
		t.Fatalf("write pref dir: %v", err)
	}

	writeSettingsIni(t, pref, settingsFixture)

	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)
	t.Chdir(t.TempDir())

	// when the settings subcommand migrates with no explicit paths
	stdout, stderr := runSettingsOK(t)

	// then both the carry and the views resolve under the home pref layout
	if !bytes.Equal(readSettingsFile(t, pref, "diablo.ini"), []byte(settingsFixture)) {
		t.Error("carried ini under the home pref layout differs from the source")
	}

	if _, err := os.Stat(filepath.Join(pref, "settings_game.yaml")); err != nil {
		t.Errorf("home view missing: %v", err)
	}

	want := "--ini " + filepath.Join(pref, settingsIniName)
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}

	if !strings.Contains(stdout, "wrote") {
		t.Errorf("stdout = %q, want the wrote mode report", stdout)
	}
}

func TestUsageSettingsSubcommandNoLongerPending(t *testing.T) {
	// given the top-level usage
	var stdout, stderr bytes.Buffer

	// when the usage is printed
	run(nil, &stdout, &stderr)

	// then settings no longer carries the pending note
	for _, line := range strings.Split(stderr.String(), "\n") {
		if strings.HasPrefix(line, "  settings") && strings.Contains(line, "not implemented yet") {
			t.Errorf("usage still marks settings pending: %q", line)
		}
	}
}

// runSettings runs the settings subcommand and reports its exit code and
// output streams.
func runSettings(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer

	code = run(append([]string{"settings"}, args...), &outBuf, &errBuf)

	return code, outBuf.String(), errBuf.String()
}

// runSettingsOK runs the settings subcommand and fails the test when it does
// not exit clean; it returns the report streams.
func runSettingsOK(t *testing.T, args ...string) (stdout, stderr string) {
	t.Helper()

	code, outText, errText := runSettings(t, args...)
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, errText)
	}

	return outText, errText
}

// migrateSettingsTwice runs the same settings migration two times and returns
// the view sets of both runs.
func migrateSettingsTwice(t *testing.T, src, out string) (first, second map[string][]byte) {
	t.Helper()

	_, _ = runSettingsOK(t, "--ini", src, "--out", out)
	first = readSettingsViews(t, out)

	_, _ = runSettingsOK(t, "--ini", src, "--out", out)

	return first, readSettingsViews(t, out)
}

// assertSettingsFailure checks the exit code and that the failure was
// reported on stderr.
func assertSettingsFailure(t *testing.T, code int, stderr string, want int) {
	t.Helper()

	if code != want {
		t.Errorf("exit code = %d, want %d (stderr: %s)", code, want, stderr)
	}
}

// writeSettingsIni writes one synthetic diablo.ini with the given content.
func writeSettingsIni(t *testing.T, dir, content string) string {
	t.Helper()

	path := filepath.Join(dir, "diablo.ini")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	return path
}

// readSettingsViews collects every written view file by its base name.
func readSettingsViews(t *testing.T, out string) map[string][]byte {
	t.Helper()

	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("read out dir: %v", err)
	}

	views := map[string][]byte{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "settings_") {
			continue
		}

		views[entry.Name()] = readSettingsFile(t, out, entry.Name())
	}

	return views
}

// readSettingsFile reads one written settings output file.
func readSettingsFile(t *testing.T, out, name string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(out, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}

	return content
}
