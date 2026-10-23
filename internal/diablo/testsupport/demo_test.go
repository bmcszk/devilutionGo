// Package testsupport_test: devilGo-xcdz slice-2 acceptance tests — StateHash
// stability (in-process and across OS processes) and the demo-tape loader.
//
// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package testsupport

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

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

// TestStateHashStableAcrossProcesses re-execs the compiled test binary in a
// child OS process and compares the child's hash against the parent's —
// genuine cross-process stability (no Go test package caches this).
func TestStateHashStableAcrossProcesses(t *testing.T) {
	if os.Getenv("DG_STATEHASH_CHILD") != "" {
		return // child runs only the helper below
	}
	// when
	//TODO remove nolint:gosec // test binary self-exec (goleak idiom)
	cmd := exec.Command(os.Args[0], "-test.run=TestStateHashChild$")
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

// TestStateHashChild is the isolated helper the parent executes: it prints
// the canonical sample hash to stdout (child mode) and skips otherwise.
func TestStateHashChild(t *testing.T) {
	if os.Getenv("DG_STATEHASH_CHILD") == "" {
		return // parent mode: the child helper below is the tested path
	}
	fmt.Println(gametestsupport.StateHash(sampleState()))
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

func TestLoadDemoTapeParsesHeaderAndIterates(t *testing.T) {
	// given
	encoded, err := os.ReadFile(demoTapePath)
	if err != nil {
		t.Fatalf("read committed demo fixture: %v", err)
	}
	raw, err := decodeHexDump(encoded)
	if err != nil {
		t.Fatalf("decode committed demo fixture: %v", err)
	}
	// when
	tape := loadDemoTapeBytes(t, raw)
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
	tick := 0
	for i := 0; i < tape.Steps(); i++ {
		step := tape.Step(i)
		switch {
		case step.Type == DemoGameTick:
			tick++
		case step.IsInput(), step.Type == DemoQuit, step.Type == DemoRender:
		default:
			t.Fatalf("step %d: unexpected type %d", i, step.Type)
		}
	}
	if tick < 30 {
		t.Fatalf("game ticks = %d, want >= 30", tick)
	}
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

func decodeHexDump(data []byte) ([]byte, error) {
	var raw []byte
	for line := range strings.SplitSeq(string(data), "\n") {
		_, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
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
	}
	return raw, nil
}

func TestStateHashStableInProcess(t *testing.T) {
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

func (s testSpelled) CanonicalFields() []gametestsupport.HashField { return s.fields }
