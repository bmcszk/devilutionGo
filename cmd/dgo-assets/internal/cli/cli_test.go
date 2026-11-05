package cli_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/cmd/dgo-assets/internal/cli"
	"github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

const gfxFixtureDir = "../../../../internal/testdata/gfx"

func TestCheckReportsMissingEntriesWithTheirLoader(t *testing.T) {
	// given
	// One entry present on disk, one absent.
	root := t.TempDir()
	present := filepath.Join(root, "data/health.clx")
	if err := os.MkdirAll(filepath.Dir(present), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(present, []byte("clx"), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	manifestPath := writeManifest(t, root, fixtureManifest())

	var stdout, stderr bytes.Buffer
	// when
	code := cli.Run([]string{"check", "--manifest", manifestPath, "--root", root}, &stdout, &stderr)
	// then
	if code != cli.ExitFailure {
		t.Fatalf("exit code = %d, want %d", code, cli.ExitFailure)
	}
	want := "missing fonts/smallfont.trn (requested by Source/engine/assets.cpp:410)"
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), want)
	}
	if !strings.Contains(stderr.String(), "1 of 2 entries") {
		t.Fatalf("stderr = %q, want the missing count", stderr.String())
	}
}

func TestCheckPassesWhenEveryEntryResolves(t *testing.T) {
	// given
	root := t.TempDir()
	for _, path := range []string{"data/health.clx", "fonts/smallfont.trn"} {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	manifestPath := writeManifest(t, root, fixtureManifest())

	var stdout, stderr bytes.Buffer
	// when
	code := cli.Run([]string{"check", "--manifest", manifestPath, "--root", root}, &stdout, &stderr)
	// then
	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, cli.ExitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "2 entries in 2 families") {
		t.Fatalf("stdout = %q, want the coverage summary", stdout.String())
	}
}

func TestCheckRejectsInvalidManifest(t *testing.T) {
	// given
	broken := fixtureManifest()
	broken.Entries[0].Format = "gif"
	root := t.TempDir()
	manifestPath := writeManifest(t, root, broken)

	var stdout, stderr bytes.Buffer
	// when
	code := cli.Run([]string{"check", "--manifest", manifestPath, "--root", root}, &stdout, &stderr)
	// then
	if code != cli.ExitFailure {
		t.Fatalf("exit code = %d, want %d", code, cli.ExitFailure)
	}
	if !strings.Contains(stderr.String(), `unknown format "gif"`) {
		t.Fatalf("stderr = %q, want the format violation", stderr.String())
	}
}

func TestListPrintsEveryEntry(t *testing.T) {
	// The same manifest renders as tab-separated lines or a markdown table.
	root := t.TempDir()
	manifestPath := writeManifest(t, root, fixtureManifest())

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "tab separated",
			args: []string{"list", "--manifest", manifestPath},
			want: "ui.data\tdata/health.clx\tclx\t1\tSource/diablo.cpp:1359\n",
		},
		{
			name: "markdown",
			args: []string{"list", "--manifest", manifestPath, "--markdown"},
			want: "| ui.data | data/health.clx | clx | 1 | Source/diablo.cpp:1359 |\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			code, stdout := runCLI(t, test.args)
			// then
			assertListOutput(t, code, stdout, test.want)
		})
	}
}

func TestCheckDefaultsToTheGraphicsInventory(t *testing.T) {
	// given a working directory whose assets/inventory.yaml is the graphics list
	dir := t.TempDir()
	t.Chdir(dir)
	for _, path := range []string{"data/health.clx", "fonts/smallfont.trn"} {
		full := filepath.Join(dir, "assets/loose", path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	writeManifest(t, filepath.Join(dir, "assets"), fixtureManifest())

	// when check runs without --manifest or --root
	code, output := runCLI(t, []string{"check"})

	// then the assets/inventory.yaml default resolves and the check passes
	assertRunOutput(t, code, output, cli.ExitOK, "2 entries in 2 families")
}

func TestConvertWritesTheGoldensTheConvertersProduce(t *testing.T) {
	// given
	// A committed CEL input and the C++ oracle's CLX output for it.
	input := readHexFixture(t, filepath.Join(gfxFixtureDir, "cel/data_talkbutton.cel.hex"))
	want := readHexFixture(t, filepath.Join(gfxFixtureDir, "cel/data_talkbutton.cel.clx.hex"))
	inputPath := filepath.Join(t.TempDir(), "talkbutton.cel")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "talkbutton.clx")

	var stdout, stderr bytes.Buffer
	// when
	code := cli.Run([]string{
		"convert", "--format", "cel", "--in", inputPath, "--out", outPath, "--widths", "53,53",
	}, &stdout, &stderr)
	// then
	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, cli.ExitOK, stderr.String())
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("converted %d bytes, want %d", len(got), len(want))
	}
}

func TestRunRejectsBadInvocations(t *testing.T) {
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{name: "no arguments", args: nil, code: cli.ExitUsage, want: "usage:"},
		{name: "unknown command", args: []string{"frobnicate"}, code: cli.ExitUsage, want: `unknown command "frobnicate"`},
		{
			name: "convert without flags", args: []string{"convert"},
			code: cli.ExitUsage, want: "--format, --in and --out are required",
		},
		{
			name: "convert with a bad width list",
			args: []string{"convert", "--format", "cel", "--in", "x", "--out", "y", "--widths", "a,b"},
			code: cli.ExitUsage, want: `width "a" is not an integer`,
		},
		{
			name: "convert with an unknown format",
			args: []string{"convert", "--format", "gif", "--in", "x", "--out", "y"},
			code: cli.ExitUsage, want: `unknown format "gif"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			code, output := runCLIOutput(t, test.args)
			// then
			assertRunOutput(t, code, output, test.code, test.want)
		})
	}
}

// writeManifest writes m as a manifest YAML under dir and returns its path.
func writeManifest(t *testing.T, dir string, m manifest.Manifest) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	path := filepath.Join(dir, "inventory.yaml")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	if err := manifest.Write(file, m); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close manifest: %v", err)
	}

	return path
}

// readHexFixture reads a committed hex dump.
func readHexFixture(t *testing.T, path string) []byte {
	t.Helper()

	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	compact := strings.NewReplacer("\n", "", " ", "", "\t", "").Replace(string(text))
	decoded, err := hex.DecodeString(compact)
	if err != nil {
		t.Fatalf("fixture %s: %v", path, err)
	}

	return decoded
}

// fixtureManifest is the baseline manifest the CLI tests check and list.
func fixtureManifest() manifest.Manifest {
	m := manifest.Manifest{
		Tool:    manifest.Tool,
		Version: manifest.Version,
		Entries: []manifest.Entry{
			{
				Family: "ui.data", Path: "data/health.clx", Format: "clx", Frames: 1, Width: 264,
				WidthSource: "Source/items.cpp:2424", Loader: "Source/diablo.cpp:1359",
			},
			{
				Family: "fonts", Path: "fonts/smallfont.trn", Format: "trn",
				Loader: "Source/engine/assets.cpp:410",
			},
		},
	}
	m.Finalize()

	return m
}

// runCLI runs one invocation and returns its exit code and stdout.
func runCLI(t *testing.T, args []string) (int, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	if code != cli.ExitOK && stderr.Len() > 0 {
		t.Logf("stderr: %s", stderr.String())
	}

	return code, stdout.String()
}

// runCLIOutput runs one invocation and returns its exit code with the combined
// stdout and stderr, for cases where the message stream does not matter.
func runCLIOutput(t *testing.T, args []string) (int, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)

	return code, stdout.String() + stderr.String()
}

// assertListOutput checks a successful list invocation and its output.
func assertListOutput(t *testing.T, code int, stdout, want string) {
	t.Helper()

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d", code, cli.ExitOK)
	}
	if !strings.Contains(stdout, want) {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
	}
}

// assertRunOutput checks an expected failing invocation against its message.
func assertRunOutput(t *testing.T, code int, output string, wantCode int, want string) {
	t.Helper()

	if code != wantCode {
		t.Fatalf("exit code = %d, want %d", code, wantCode)
	}
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want it to contain %q", output, want)
	}
}
