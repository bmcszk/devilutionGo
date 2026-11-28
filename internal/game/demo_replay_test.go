package game_test

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
)

const (
	// demoTapeHexPath is the committed tape of timerun demo 0.
	demoTapeHexPath = "../../internal/testdata/demo/demo_0.dmo.hex"
	// demoTapeStatsPath is its committed record and pass accounting.
	demoTapeStatsPath = "../../internal/testdata/demo/demo_0.replay_stats.tsv"
	// cmdPerTickPrefix marks the per-tick command histogram rows.
	cmdPerTickPrefix = "cmd_per_tick:"
)

// replayOutcome is one replay's hash and accounting.
type replayOutcome struct {
	Hash  uint64
	Stats game.DemoReplayStats
}

func TestDemoReplayRealTapeAccounting(t *testing.T) {
	// given the committed tape of timerun demo 0 and its committed accounting
	raw := readDemoTape(t)
	table := gametestsupport.MetricTable(t, demoTapeStatsPath)

	// when the tape replays through a fresh kernel
	replay := game.NewDemoReplay(kernel.New(nil, nil))
	if _, err := replay.ReplayUpstream(bytes.NewReader(raw)); err != nil {
		t.Fatalf("ReplayUpstream: %v", err)
	}

	// then every committed count matches the replay
	stats := replay.Stats()

	assertReplayCounts(t, stats, table)
	assertCommandHistogram(t, stats, table)
}

func TestDemoReplayRealTapeIsDeterministic(t *testing.T) {
	// given the committed tape
	raw := readDemoTape(t)

	// when it replays twice through fresh kernels
	first := replayTape(t, raw)
	second := replayTape(t, raw)

	// then the trace hash AND the accounting agree. The hash alone proves
	// little here: with no handlers registered the trace stays empty and the
	// hash is the FNV offset basis whatever the stream carried.
	if first.Hash != second.Hash {
		t.Fatalf("hash = %d then %d, want the same", first.Hash, second.Hash)
	}

	if !equalStats(first.Stats, second.Stats) {
		t.Fatalf("accounting differs between replays:\nfirst  %+v\nsecond %+v", first.Stats, second.Stats)
	}
}

// replayTape replays the tape once through a fresh kernel.
func replayTape(t *testing.T, raw []byte) replayOutcome {
	t.Helper()

	replay := game.NewDemoReplay(kernel.New(nil, nil))

	hash, err := replay.ReplayUpstream(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("ReplayUpstream: %v", err)
	}

	return replayOutcome{Hash: hash, Stats: replay.Stats()}
}

// readDemoTape decodes the committed xxd tape fixture.
func readDemoTape(t *testing.T) []byte {
	t.Helper()

	dump, err := os.ReadFile(demoTapeHexPath)
	if err != nil {
		t.Fatalf("read committed tape: %v", err)
	}

	raw, err := gametestsupport.DecodeXXD(dump)
	if err != nil {
		t.Fatalf("decode committed tape: %v", err)
	}

	return raw
}

// assertReplayCounts compares the pass, command and record counts with the
// committed rows.
func assertReplayCounts(t *testing.T, stats game.DemoReplayStats, table map[string]string) {
	t.Helper()

	counts := []struct {
		metric string
		got    int
	}{
		{"replay_passes", stats.Ticks},
		{"replay_commands", stats.Commands},
		{"records_game_tick", stats.Records.GameTicks},
		{"records_render", stats.Records.Renders},
		{"records_mouse_motion", stats.Records.MouseMotions},
		{"records_mouse_button_down", stats.Records.MouseButtonDowns},
		{"records_mouse_button_up", stats.Records.MouseButtonUps},
		{"records_mouse_wheel", stats.Records.MouseWheels},
		{"records_key_down", stats.Records.KeyDowns},
		{"records_key_up", stats.Records.KeyUps},
		{"records_quit", stats.Records.Quits},
		{"records_custom", stats.Records.Customs},
	}

	for _, count := range counts {
		assertMetric(t, count.metric, count.got, gametestsupport.MetricInt(t, table, count.metric))
	}

	assertMetricText(t, "commands_per_tick_hash",
		strconv.FormatUint(commandsPerTickHash(stats.CommandsPerTick), 10), table["commands_per_tick_hash"])
}

// assertCommandHistogram compares the per-tick command histogram with the
// committed cmd_per_tick:<n> rows, in both directions.
func assertCommandHistogram(t *testing.T, stats game.DemoReplayStats, table map[string]string) {
	t.Helper()

	want := committedHistogram(t, table)
	got := commandHistogram(stats.CommandsPerTick)

	if len(got) != len(want) {
		t.Fatalf("histogram buckets = %d, want %d (got %v, want %v)", len(got), len(want), got, want)
	}

	for count, ticks := range want {
		assertMetric(t, cmdPerTickPrefix+strconv.Itoa(count), got[count], ticks)
	}
}

// committedHistogram reads the committed cmd_per_tick:<n> rows.
func committedHistogram(t *testing.T, table map[string]string) map[int]int {
	t.Helper()

	histogram := make(map[int]int)

	for metric, value := range table {
		key, found := strings.CutPrefix(metric, cmdPerTickPrefix)
		if !found {
			continue
		}

		count, err := strconv.Atoi(key)
		if err != nil {
			t.Fatalf("metric %q: %v", metric, err)
		}

		histogram[count] = parseMetric(t, metric, value)
	}

	return histogram
}

// commandHistogram counts how many passes dispatched each command count.
func commandHistogram(perTick []int) map[int]int {
	histogram := make(map[int]int, len(perTick))

	for _, count := range perTick {
		histogram[count]++
	}

	return histogram
}

// commandsPerTickHash pins the ORDER of the per-tick command counts, which the
// histogram alone loses: FNV-1a 64 over each count as a little-endian word.
func commandsPerTickHash(perTick []int) uint64 {
	hash := fnv.New64a()

	var buf [4]byte

	for _, count := range perTick {
		binary.LittleEndian.PutUint32(buf[:], uint32(count))
		_, _ = hash.Write(buf[:])
	}

	return hash.Sum64()
}

// equalStats reports whether two accountings agree.
func equalStats(first, second game.DemoReplayStats) bool {
	if first.Ticks != second.Ticks || first.Commands != second.Commands || first.Records != second.Records {
		return false
	}

	if len(first.CommandsPerTick) != len(second.CommandsPerTick) {
		return false
	}

	for i := range first.CommandsPerTick {
		if first.CommandsPerTick[i] != second.CommandsPerTick[i] {
			return false
		}
	}

	return true
}

// assertMetric compares one measured count with its committed value.
func assertMetric(t *testing.T, metric string, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %d, want %d", metric, got, want)
	}
}

// assertMetricText compares one measured value with its committed text.
func assertMetricText(t *testing.T, metric, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %s, want %s", metric, got, want)
	}
}

// parseMetric parses one committed count.
func parseMetric(t *testing.T, metric, value string) int {
	t.Helper()

	count, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("metric %q = %q: %v", metric, value, err)
	}

	return count
}
