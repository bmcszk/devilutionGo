// The AC-3 scene-row runner (devilGo-ht4l): one measurement row per process.
//
// The retention that bounds every AC-3 measurement is a property of the PROCESS,
// not of the scene. Presenter.Draw uploads its index and shade textures with
// Image.WritePixels, and ebiten's graphics command queue keeps those buffers live
// until a running loop flushes them — there is no flush a caller outside RunGame
// can reach. Measured on the town scene in one process: the 32-frame p95 holds
// ~5 ms through ~192 cumulative presented frames, then climbs (frames 224-255
// measured 72.8 ms, frames 576-607 measured 162.0 ms) as the collector runs
// against the retained queue.
//
// One row presents frameSamples + frameWarmup = 136 frames, which fits that
// window. AC-3 has two rows, so measuring both in one process puts 272 presented
// frames into one series' window and the later row fails on the earlier row's
// retained queue rather than on its own frame cost (measured: 22-84 ms p95 on
// otherwise idle hardware, both rows, either order). Each row is therefore
// measured in its own process, which is the single-row window the merged gate was
// pinned inside.
//
// A row is measured in the runner class' passes (frameBudgetClass): the CI class
// runs 2 fresh processes and holds the row on any passing pass (measureRowPasses).
package integration_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The environment contract between a row's own process and the runner.
const (
	// rowEnv names the scene row the child process measures.
	rowEnv = "DEVILUTIONGO_FRAME_ROW"
	// rowChildTest is the child test name; the parent runs exactly this one.
	rowChildTest = "TestM1FrameRowChild"
	// rowReportPrefix starts the child's one report line.
	rowReportPrefix = "frame-row p95 "
)

// rowPass is one scripted child pass: its combined output and its exit error.
type rowPass struct {
	output string
	err    error
}

// rowPassesCase is one scripted pass-rule case: how many passes the class asks for,
// the child's passes in order, and the verdict, text and pass count it must earn.
type rowPassesCase struct {
	name     string
	passes   int
	script   []rowPass
	wantErr  bool
	wantText []string
	notWant  []string
	wantRuns int
}

// TestM1FrameRowChild measures the row $DEVILUTIONGO_FRAME_ROW names and asserts
// that row's budget. It is the child half of measureRowInOwnProcess and skips
// without the variable, so a plain package run cannot measure a row here and
// quietly admit the retained queue.
func TestM1FrameRowChild(t *testing.T) {
	// given the row label this process was started to measure
	label := os.Getenv(rowEnv)
	if label == "" {
		t.Skip("row child: run through measureRowInOwnProcess")
	}

	name, build, ok := sceneRowFor(label)
	if !ok {
		t.Fatalf("row child: %q names no AC-3 scene row", label)
	}

	samples := budgetSamples()

	// when the row's whole pipeline is sampled frame by frame
	times := measureFrames(build(t), samples)

	// then the row holds this runner class' budget and reports its table
	assertFrameBudgetPin(t, name, samples, frameBudgetClass(), times)

	t.Logf("%s%q: %s", rowReportPrefix, name, stageTable(times))
}

// TestMeasureRowPassesFirstPassThatHolds pins the pass rule the CI class relies on:
// a row holds as soon as one pass holds it, and fails only when every pass missed,
// quoting each. The pre-fix parent (fail on the first failing child) reds the
// fail-then-pass case, so this test is the pin that fix earns.
func TestMeasureRowPassesFirstPassThatHolds(t *testing.T) {
	// when
	subtests := rowPassesCases()
	// then
	for _, test := range subtests {
		t.Run(test.name, func(t *testing.T) {
			checkRowPassesCase(t, test)
		})
	}
}

// measureRowInOwnProcess measures one AC-3 row in the child processes the runner
// class asks for and returns the p95 table of a pass that held the row's budget. A
// row fails the parent only when every pass missed it, with each missed pass's own
// failure text and stage table quoted, so the process boundary cannot hide a
// regression while one stalled process window cannot red the row.
func measureRowInOwnProcess(t *testing.T, label string) string {
	t.Helper()

	table, err := measureRowPasses(label, frameBudgetClass().passes, rowChildProcess(t, label))
	if err != nil {
		t.Fatalf("%v", err)
	}

	return table
}

// rowChildProcess starts one child process told to measure the row label names and
// returns its combined output, logging the p95 table it reported so a passing run's
// log carries every pass' figure, not only the pass that held the row.
func rowChildProcess(t *testing.T, label string) func() (string, error) {
	t.Helper()

	return func() (string, error) {
		cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^"+rowChildTest+"$", "-test.v")
		cmd.Env = append(os.Environ(), rowEnv+"="+label)

		out, err := cmd.CombinedOutput()

		if table, ok := rowReport(string(out), label); ok {
			t.Logf("%s row pass p95 over %d frames: %s", label, budgetSamples(), table)
		}

		return string(out), err
	}
}

// measureRowPasses measures one row by running runChild once per pass and holds the
// row when any pass exits clean and reports its p95 table. It fails only when every
// pass missed, quoting each: a stall lives in one process window, a real regression
// slows all of them.
func measureRowPasses(label string, passes int, runChild func() (string, error)) (string, error) {
	held := ""
	failures := make([]string, 0, passes)

	for pass := 1; pass <= passes; pass++ {
		output, err := runChild()
		if err == nil {
			if table, ok := rowReport(output, label); ok {
				held = table

				continue
			}
		}

		failures = append(failures, rowPassFailure(pass, passes, label, output, err))
	}

	if held != "" {
		return held, nil
	}

	return "", fmt.Errorf("%s row: %d passes, every pass missed:\n%s", label, passes, strings.Join(failures, "\n\n"))
}

// rowPassFailure is one missed pass' text: the child's exit error and output, or the
// missing table when it exited clean. The pass prefix is left off a single pass so
// the local class reports exactly what it reported before.
func rowPassFailure(pass, passes int, label, output string, err error) string {
	prefix := ""
	if passes > 1 {
		prefix = fmt.Sprintf("pass %d/%d: ", pass, passes)
	}

	if err != nil {
		return fmt.Sprintf("%s%s row process failed: %v\n%s", prefix, label, err, output)
	}

	return fmt.Sprintf("%s%s row process reported no p95 table:\n%s", prefix, label, output)
}

// rowPassesCases is the scripted pass-rule matrix: the fail-then-pass walk the rule
// exists for, the all-missed walk that must still fail, the clean-exit-without-table
// walk, and the local single pass whose text must not change.
func rowPassesCases() []rowPassesCase {
	stalled := missedRowPass()
	held := heldRowPass()

	return []rowPassesCase{
		{
			name:     "a second pass holds the row the first pass missed",
			passes:   2,
			script:   []rowPass{stalled, held},
			wantText: []string{"world 15.412"},
			notWant:  []string{"pass 1/", "27.195ms"},
			wantRuns: 2,
		},
		{
			name:     "a row every pass missed fails with every pass quoted",
			passes:   2,
			script:   []rowPass{stalled, stalled},
			wantErr:  true,
			wantText: []string{"pass 1/2:", "pass 2/2:", "27.195ms", "world 22.777", "present 5.169"},
			wantRuns: 2,
		},
		{
			name:     "a clean pass without a table does not hold the row",
			passes:   2,
			script:   []rowPass{{output: "=== RUN   TestM1FrameRowChild\n"}, held},
			wantText: []string{"world 15.412"},
			notWant:  []string{"pass 1/"},
			wantRuns: 2,
		},
		{
			name:     "the local single pass keeps the unfixed failure text",
			passes:   1,
			script:   []rowPass{stalled},
			wantErr:  true,
			wantText: []string{"town row process failed: exit status 1", "27.195ms"},
			notWant:  []string{"pass 1/"},
			wantRuns: 1,
		},
	}
}

// heldRowPass is a child pass that held the town row: its own report line, verbatim.
func heldRowPass() rowPass {
	return rowPass{output: "=== RUN   TestM1FrameRowChild\n" +
		"    " + rowReportPrefix + `"` + townRowLabel + `": world 15.412 panel 0.008 text 0.051 present 1.060` + "\n"}
}

// missedRowPass is a child pass that missed the town row: the child's own budget
// failure, verbatim, with the non-zero exit that carries it.
func missedRowPass() rowPass {
	return rowPass{
		output: "    town_test.go:1: town row: frame pipeline p95 = 27.195ms over 128 frames, " +
			"budget 26.672ms (CI runner): world 22.777 panel 0.008 text 0.051 present 5.169\n",
		err: errors.New("exit status 1"),
	}
}

// checkRowPassesCase runs one scripted case through the pass rule and asserts its
// verdict, the text a reviewer reads and how many passes it was allowed to run.
func checkRowPassesCase(t *testing.T, test rowPassesCase) {
	t.Helper()

	// given a row measured over the class' passes, scripted pass by pass
	runChild, runs := scriptedRowChild(t, test.script)

	// when the parent applies the pass rule
	table, err := measureRowPasses(townRowLabel, test.passes, runChild)

	// then the rule's verdict, text and pass count are the ones it owes
	if got := err != nil; got != test.wantErr {
		t.Fatalf("measureRowPasses error = %v, wantErr %v", err, test.wantErr)
	}

	text := table
	if err != nil {
		text = err.Error()
	}

	checkRowPassesText(t, text, test.wantText, test.notWant)

	if *runs != test.wantRuns {
		t.Fatalf("child ran %d passes, want %d", *runs, test.wantRuns)
	}
}

// checkRowPassesText asserts the pieces the rule's text must carry and must not.
func checkRowPassesText(t *testing.T, text string, want, notWant []string) {
	t.Helper()

	for _, piece := range want {
		if !strings.Contains(text, piece) {
			t.Fatalf("measureRowPasses text %q does not carry %q", text, piece)
		}
	}

	for _, piece := range notWant {
		if strings.Contains(text, piece) {
			t.Fatalf("measureRowPasses text %q carries %q", text, piece)
		}
	}
}

// scriptedRowChild returns a child runner that answers with the scripted passes in
// order and the count of passes it was asked for.
func scriptedRowChild(t *testing.T, script []rowPass) (func() (string, error), *int) {
	t.Helper()

	runs := 0

	return func() (string, error) {
		if runs >= len(script) {
			t.Fatalf("child ran past the scripted %d passes", len(script))
		}

		pass := script[runs]
		runs++

		return pass.output, pass.err
	}, &runs
}

// sceneRowFor resolves a row label to its name and scene builder.
func sceneRowFor(label string) (string, func(testing.TB) *frameScene, bool) {
	for _, row := range frameRows() {
		if row.label == label {
			return row.label, row.build, true
		}
	}

	return "", nil, false
}

// rowReport picks the p95 table one child reported for label.
func rowReport(output, label string) (string, bool) {
	prefix := rowReportPrefix + fmt.Sprintf("%q: ", label)

	for _, line := range strings.Split(output, "\n") {
		// the child's t.Logf line carries the -test.v log prefix, so the marker is
		// located rather than matched at the start
		start := strings.Index(line, prefix)
		if start < 0 {
			continue
		}

		return strings.TrimSpace(line[start+len(prefix):]), true
	}

	return "", false
}
