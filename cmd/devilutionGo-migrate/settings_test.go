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

	// then the copy is cmp-equal to the source and the exit is clean
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
