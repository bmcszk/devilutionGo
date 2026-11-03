package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRoot holds the synthetic packages the gate proofs feed to the checker.
const fixtureRoot = "../testdata"

// verdict is one gate run's exit code and report, kept for the then phase.
type verdict struct {
	name   string
	code   int
	report string
}

// TestRun_ImpureFixture_FailsNamingPackageFileAndOffender is the negative proof:
// one synthetic package holding every banned line must fail the gate, and the
// report must name the package, the file and the offending import or call.
func TestRun_ImpureFixture_FailsNamingPackageFileAndOffender(t *testing.T) {
	// given
	var stdout, stderr bytes.Buffer

	// when
	code := run([]string{filepath.Join(fixtureRoot, "impure")}, &stdout, &stderr)

	// then
	if code != 1 {
		t.Fatalf("run(impure) = %d, want 1; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	wantPrefix := filepath.Join(fixtureRoot, "impure") + ": " +
		filepath.Join(fixtureRoot, "impure", "impure.go") + ":"
	if !strings.Contains(stdout.String(), wantPrefix) {
		t.Fatalf("report = %q, want it to start findings with %q", stdout.String(), wantPrefix)
	}

	for _, want := range []string{
		`rendering imports "github.com/hajimehoshi/ebiten/v2"`,
		`rendering imports "github.com/bmcszk/devilutionGo/internal/engine/render"`,
		`file io imports "os"`,
		"wall clock time.Now",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("report = %q, want it to contain %q", stdout.String(), want)
		}
	}
}

// TestBannedImport_ClassifiesPaths pins the ban list itself, path by path.
func TestBannedImport_ClassifiesPaths(t *testing.T) {
	// given
	cases := []struct {
		path string
		want string
	}{
		{path: "github.com/hajimehoshi/ebiten/v2", want: "rendering"},
		{path: "github.com/hajimehoshi/ebiten/v2/audio", want: "rendering"},
		{path: "github.com/bmcszk/devilutionGo/internal/engine/render", want: "rendering"},
		{path: "os", want: "file io"},
		{path: "io/ioutil", want: "file io"},
		{path: "", want: ""},
		{path: "time", want: ""},
		{path: "io", want: ""},
		{path: "github.com/bmcszk/devilutionGo/internal/game/kernel", want: ""},
	}

	// when
	verdicts := make([]string, 0, len(cases))
	for _, tc := range cases {
		verdicts = append(verdicts, bannedImport(tc.path))
	}

	// then
	for i, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if verdicts[i] != tc.want {
				t.Fatalf("bannedImport(%q) = %q, want %q", tc.path, verdicts[i], tc.want)
			}
		})
	}
}

// TestRun_TestFileExempt_NotReported pins the exemption: the fixture's
// impure_test.go reads the disk and must not appear in the report.
func TestRun_TestFileExempt_NotReported(t *testing.T) {
	// given
	var stdout, stderr bytes.Buffer

	// when
	code := run([]string{filepath.Join(fixtureRoot, "impure")}, &stdout, &stderr)

	// then
	if code != 1 {
		t.Fatalf("run(impure) = %d, want 1", code)
	}

	if strings.Contains(stdout.String(), "impure_test.go") {
		t.Fatalf("report = %q, want _test.go files exempt", stdout.String())
	}
}

// TestRun_CleanFixtures_Pass covers the controls: pure sim code with a
// time.Duration constant passes, and test scaffolding under testsupport may
// open files.
func TestRun_CleanFixtures_Pass(t *testing.T) {
	// given
	names := []string{"clean", "exempt"}

	// when
	verdicts := make([]verdict, 0, len(names))
	for _, name := range names {
		code, report := gateVerdict(t, filepath.Join(fixtureRoot, name))
		verdicts = append(verdicts, verdict{name: name, code: code, report: report})
	}

	// then
	for _, got := range verdicts {
		if got.code != 0 {
			t.Fatalf("run(%s) = %d, want 0; report=%q", got.name, got.code, got.report)
		}

		if !strings.Contains(got.report, "clean") {
			t.Fatalf("run(%s) report = %q, want a clean verdict", got.name, got.report)
		}
	}
}

// TestRun_MissingTree_FailsClosed proves an unreadable tree is never a pass.
func TestRun_MissingTree_FailsClosed(t *testing.T) {
	// given
	var stdout, stderr bytes.Buffer

	// when
	code := run([]string{filepath.Join(fixtureRoot, "absent")}, &stdout, &stderr)

	// then
	if code != 2 {
		t.Fatalf("run(absent) = %d, want 2 (fail-closed)", code)
	}

	if !strings.Contains(stderr.String(), "testdata/absent") {
		t.Fatalf("stderr = %q, want it to name the missing tree", stderr.String())
	}
}

// TestRun_EmptyTree_FailsClosed proves a pass must have scanned a source: an
// empty tree is a misconfigured gate, not a clean one.
func TestRun_EmptyTree_FailsClosed(t *testing.T) {
	// given
	root := filepath.Join(fixtureRoot, "empty")

	// when
	code, report := gateVerdict(t, root)

	// then
	if code != 2 {
		t.Fatalf("run(%s) = %d, want 2 (fail-closed on an empty tree)", root, code)
	}

	if !strings.Contains(report, "no sim sources scanned") {
		t.Fatalf("report = %q, want it to explain the empty scan", report)
	}
}

// TestRun_AliasedTimeImport_FailsNamingOffendingCall is the negative proof for
// the alias evasion: `import t "time"` with a wall-clock read must fail the
// gate and name the file:line of each banned call.
func TestRun_AliasedTimeImport_FailsNamingOffendingCall(t *testing.T) {
	// given
	want := []string{
		aliasedFinding("aliased.go", 13, "wall clock t.NewTicker"),
		aliasedFinding("aliased.go", 17, "wall clock t.Tick"),
		aliasedFinding("aliased.go", 19, "wall clock t.Since"),
		aliasedFinding("aliased.go", 19, "wall clock t.Now"),
	}

	// when
	code, report := gateVerdict(t, filepath.Join(fixtureRoot, "aliased"))

	// then
	if code != 1 {
		t.Fatalf("run(aliased) = %d, want 1; report=%q", code, report)
	}

	for _, finding := range want {
		if !strings.Contains(report, finding) {
			t.Fatalf("report = %q, want it to contain %q", report, finding)
		}
	}
}

// TestRun_DotTimeImport_FailsNamingBareCall proves the dot-import evasion is
// closed too: bare Now, Tick, NewTicker and Since calls are wall-clock reads.
func TestRun_DotTimeImport_FailsNamingBareCall(t *testing.T) {
	// given
	want := []string{
		aliasedFinding("dotted.go", 11, "wall clock NewTicker"),
		aliasedFinding("dotted.go", 15, "wall clock Tick"),
		aliasedFinding("dotted.go", 17, "wall clock Since"),
		aliasedFinding("dotted.go", 17, "wall clock Now"),
	}

	// when
	code, report := gateVerdict(t, filepath.Join(fixtureRoot, "aliased"))

	// then
	if code != 1 {
		t.Fatalf("run(aliased) = %d, want 1; report=%q", code, report)
	}

	for _, finding := range want {
		if !strings.Contains(report, finding) {
			t.Fatalf("report = %q, want it to contain %q", report, finding)
		}
	}
}

// gateVerdict runs the gate over root and returns its exit code and its whole
// report (stdout plus stderr), so both the pass line and scan errors are visible.
func gateVerdict(t *testing.T, root string) (int, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	code := run([]string{root}, &stdout, &stderr)

	return code, stdout.String() + stderr.String()
}

// aliasedFinding renders one expected finding for the aliased fixture.
func aliasedFinding(file string, line int, detail string) string {
	return fmt.Sprintf("%s:%d: %s", filepath.Join(fixtureRoot, "aliased", file), line, detail)
}
