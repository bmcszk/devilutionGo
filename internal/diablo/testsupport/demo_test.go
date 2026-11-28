// Package testsupport_test: devilGo-xcdz slice-2 acceptance tests — StateHash
// stability (in-process and across OS processes) and the demo-tape loader.

package testsupport_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
)

// demoTapePath is the committed demo tape, stored as an xxd hex dump of the
// recorded .dmo bytes.
const demoTapePath = "../../../internal/testdata/demo/demo_0.dmo.hex"

// demoTapeStatsPath is the committed replay accounting of the same tape.
const demoTapeStatsPath = "../../../internal/testdata/demo/demo_0.replay_stats.tsv"

// testSpelled adapts a field list to the HashableState seam.
type testSpelled struct {
	name   string
	fields []gametestsupport.HashField
}

func TestStateHashStableAcrossProcesses(t *testing.T) {
	// given the parent process, which is the one that re-execs a child
	if os.Getenv("DG_STATEHASH_CHILD") != "" {
		return // child runs only the helper below
	}
	// when
	// TODO remove nolint:gosec // test binary self-exec (goleak idiom)
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestStateHashChild$")

	cmd.Env = append(os.Environ(), "DG_STATEHASH_CHILD=1")

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("child run: %v", err)
	}
	// then
	childHash := parseChildHash(t, string(out))

	parentHash := gametestsupport.StateHash(sampleState())
	if childHash != parentHash {
		t.Fatalf("hash differs across processes: child %d, parent %d", childHash, parentHash)
	}
}

func TestStateHashChild(t *testing.T) {
	// given the child-mode flag decides whether this process hashes at all
	if os.Getenv("DG_STATEHASH_CHILD") == "" {
		return // parent mode: the child helper below is the tested path
	}
	// when the child writes the sample state hash to stdout
	_, err := fmt.Fprintf(os.Stdout, "%d\n", gametestsupport.StateHash(sampleState()))
	// then a failed write fails the child
	if err != nil {
		t.Fatalf("child hash write: %v", err)
	}
}

func TestLoadDemoTapeParsesHeaderAndIterates(t *testing.T) {
	// given the committed demo tape and its committed accounting
	tape := loadDemoTape(t)
	table := gametestsupport.MetricTable(t, demoTapeStatsPath)

	// when the header and the record stream are read
	steps := tape.Steps()
	ticks := countGameTicks(t, tape)

	// then the header carries the recorded version, save and resolution
	// (demomode.cpp:519-525, ReadSettings :157-187)
	if got := tape.Version; got != 3 {
		t.Fatalf("version = %d, want 3", got)
	}

	if got := tape.SaveNumber; got != 0 {
		t.Fatalf("save number = %d, want 0", got)
	}

	if got := [2]int{tape.GraphicsWidth, tape.GraphicsHeight}; got != [2]int{768, 480} {
		t.Fatalf("resolution = %dx%d, want 768x480", got[0], got[1])
	}

	// ...and the exact record counts match the committed rows
	if got, want := steps, gametestsupport.MetricInt(t, table, "tape_steps"); got != want {
		t.Fatalf("steps = %d, want %d", got, want)
	}

	if got, want := ticks, gametestsupport.MetricInt(t, table, "tape_game_ticks"); got != want {
		t.Fatalf("game ticks = %d, want %d", got, want)
	}
}

func TestLoadDemoTapeFirstRecords(t *testing.T) {
	// given the committed demo tape
	tape := loadDemoTape(t)

	// when the first records after the 32-byte header are read
	first, second, third := tape.Step(0), tape.Step(1), tape.Step(2)

	// then the stream starts with the recorded KeyUp of SDLK_RETURN
	// (hex offset 0x20: 0e 80 0d 00 00 00 00 00)
	assertStep(t, first, testsupport.DemoStep{Type: testsupport.DemoKeyUp, Progress: 0x80, Sym: 13})
	// ...followed by the first game tick and the first render-only frame
	assertStep(t, second, testsupport.DemoStep{Type: testsupport.DemoGameTick, Progress: 0x80})
	assertStep(t, third, testsupport.DemoStep{Type: testsupport.DemoRender, Progress: 2})

	// ...and the first pointer motion is the recorded (331, 314)
	assertStep(t, firstStepOfType(t, tape, testsupport.DemoMouseMotion),
		testsupport.DemoStep{Type: testsupport.DemoMouseMotion, X: 331, Y: 314})
}

func TestStateHashStableInProcess(t *testing.T) {
	// given the canonical state fixture
	state := sampleState()
	// when
	first := gametestsupport.StateHash(state)
	second := gametestsupport.StateHash(state)
	// then
	if first != second {
		t.Fatalf("hash unstable in-process: %d vs %d", first, second)
	}

	if first == 0 {
		t.Fatal("degenerate zero hash")
	}
}

func parseChildHash(t *testing.T, out string) uint64 {
	t.Helper()

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--- FAIL") {
			t.Fatalf("child failed: %s", out)
		}

		var hash uint64
		if _, err := fmt.Sscanf(line, "%d", &hash); err == nil && line != "" {
			return hash
		}
	}

	t.Fatal("child hash line not found (child must print it)")

	return 0
}

// countGameTicks walks every step, failing on unexpected types, and counts
// DemoGameTick steps.
func countGameTicks(t *testing.T, tape *testsupport.DemoTape) int {
	t.Helper()

	tick := 0

	for i := range tape.Steps() {
		step := tape.Step(i)
		switch {
		case step.Type == testsupport.DemoGameTick:
			tick++
		case step.IsInput(), step.Type == testsupport.DemoQuit, step.Type == testsupport.DemoRender:
		default:
			t.Fatalf("step %d: unexpected type %d", i, step.Type)
		}
	}

	return tick
}

// sampleState is the canonical state fixture: identity, position, vitals
// and mode fields in a fixed field order (caches/derived geometry are
// deliberately absent — see hash.go's canonical-field contract).
func sampleState() testSpelled {
	return testSpelled{
		name: "hero-0",
		fields: []gametestsupport.HashField{
			{Name: "player.id", Bytes: []byte{1}},
			{Name: "player.tileX", Bytes: []byte{42}},
			{Name: "player.tileY", Bytes: []byte{7}},
			{Name: "player.hp", Bytes: []byte{100}},
			{Name: "player.mode", Bytes: []byte{3}},
		},
	}
}

func (s testSpelled) CanonicalFields() []gametestsupport.HashField { return s.fields }

// loadDemoTape reads and parses the committed tape.
func loadDemoTape(t *testing.T) *testsupport.DemoTape {
	t.Helper()

	dump, err := os.ReadFile(demoTapePath)
	if err != nil {
		t.Fatalf("read committed demo fixture: %v", err)
	}

	raw, err := gametestsupport.DecodeXXD(dump)
	if err != nil {
		t.Fatalf("decode committed demo fixture: %v", err)
	}

	return testsupport.LoadDemoTapeBytes(t, raw)
}

// assertStep checks one parsed record against its hand-decoded expectation.
func assertStep(t *testing.T, got, want testsupport.DemoStep) {
	t.Helper()

	if got != want {
		t.Fatalf("step = %+v, want %+v", got, want)
	}
}

// firstStepOfType returns the first parsed record of the given type.
func firstStepOfType(t *testing.T, tape *testsupport.DemoTape, stepType byte) testsupport.DemoStep {
	t.Helper()

	for i := range tape.Steps() {
		if step := tape.Step(i); step.Type == stepType {
			return step
		}
	}

	t.Fatalf("no record of type %d in the tape", stepType)

	return testsupport.DemoStep{}
}
