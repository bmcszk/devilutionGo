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
package integration_test

import (
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

// measureRowInOwnProcess measures one AC-3 row in a child process and returns the
// p95 table it asserted. A row that misses its budget fails the parent with the
// child's own failure text and stage table, so the process boundary cannot hide a
// regression.
func measureRowInOwnProcess(t *testing.T, label string) string {
	t.Helper()

	// given this test binary as the child, told which one row to measure
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^"+rowChildTest+"$", "-test.v")
	cmd.Env = append(os.Environ(), rowEnv+"="+label)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s row process failed: %v\n%s", label, err, out)
	}

	// then the child reported a complete row table
	table, ok := rowReport(string(out), label)
	if !ok {
		t.Fatalf("%s row process reported no p95 table:\n%s", label, out)
	}

	return table
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
