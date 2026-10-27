// Package testsupport_test: devilGo-xcdz slice-2 acceptance tests — StateHash
// stability (in-process and across OS processes) and the demo-tape loader.

package testsupport_test

import (
	"bufio"
	"context"
	"encoding/hex"
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
	// given
	encoded, err := os.ReadFile(demoTapePath)
	if err != nil {
		t.Fatalf("read committed demo fixture: %v", err)
	}

	raw := decodeHexDump(encoded)
	// when
	tape := testsupport.LoadDemoTapeBytes(t, raw)
	// then
	if got := tape.Version; got != 3 {
		t.Fatalf("version = %d, want 3", got)
	}

	if tape.GraphicsWidth == 0 || tape.GraphicsHeight == 0 {
		t.Fatalf("resolution %dx%d not parsed", tape.GraphicsWidth, tape.GraphicsHeight)
	}

	if got := tape.Steps(); got < 30 {
		t.Fatalf("steps = %d, want >= 30 (pinned DemoFrameCount)", got)
	}

	tick := countGameTicks(t, tape)
	if tick < 30 {
		t.Fatalf("game ticks = %d, want >= 30", tick)
	}
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

func decodeHexDump(data []byte) []byte {
	var raw []byte

	for line := range strings.SplitSeq(string(data), "\n") {
		_, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		raw = appendHexFields(raw, after)
	}

	return raw
}

// appendHexFields decodes the 4-hex-digit fields of one dump line, stopping
// at the first malformed field.
func appendHexFields(raw []byte, after string) []byte {
	for field := range strings.FieldsSeq(after) {
		if len(field) != 4 {
			break
		}

		decoded, err := hex.DecodeString(field)
		if err != nil {
			break
		}

		raw = append(raw, decoded...)
	}

	return raw
}

func (s testSpelled) CanonicalFields() []gametestsupport.HashField { return s.fields }
