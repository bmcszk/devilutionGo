package game_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// tickDelay is gnTickDelay, diablo.cpp:146 (50ms, 20 ticks per second); the
// unexported constant is not visible from the external test package.
const tickDelay = 50 * time.Millisecond

func TestDemoReplay_SameSeedAndStreamHashIdentically(t *testing.T) {
	// given
	stream := demoStream(t)

	// when
	firstHash, firstEvents := replayKernel(t, 1234, stream)
	secondHash, secondEvents := replayKernel(t, 1234, stream)

	// then
	if firstEvents == 0 {
		t.Fatal("replay emitted no events, want the recorded inputs replayed")
	}

	if firstHash != secondHash {
		t.Fatalf("replay twice hashed %d then %d, want the same trace hash", firstHash, secondHash)
	}

	if firstEvents != secondEvents {
		t.Fatalf("replay twice emitted %d then %d events, want the same count", firstEvents, secondEvents)
	}
}

func TestDemoReplay_DifferentSeedChangesTheHash(t *testing.T) {
	// given
	stream := demoStream(t)

	// when
	sameHash, _ := replayKernel(t, 1, stream)
	otherHash, _ := replayKernel(t, 2, stream)

	// then
	if sameHash == otherHash {
		t.Fatalf("different seeds shared hash %d, want the seed to change the trace", sameHash)
	}
}

func TestGame_Update_HandsEachSettledTickToTheKernel(t *testing.T) {
	// given
	tick := &engine.MockTickSource{}
	g := game.NewGame(tick)
	g.Kernel = kernel.New(nil, nil)
	g.Kernel.Register(kernel.KindTickProcessed, func(kernel.Command, *kernel.Streams) []kernel.Event {
		return []kernel.Event{{Kind: kernel.KindTickProcessed}}
	})
	g.StartupBurst = false

	// when
	ran := g.Update(3 * tickDelay)

	// then
	if ran != 3 {
		t.Fatalf("Update ran %d ticks, want 3: the tick budget must not change", ran)
	}

	if g.TickCount != 3 {
		t.Fatalf("kernel tick count = %d, want 3 (one pass per settled tick)", g.TickCount)
	}
}

func TestGame_Update_KernelPassDoesNotChangeTickBudget(t *testing.T) {
	// given
	unwired := game.NewGame(&engine.MockTickSource{})
	wired := game.NewGame(&engine.MockTickSource{})
	wired.Kernel = kernel.New(kernel.DropUnnamed(), kernel.NewStreams(9, "ai"))

	// when
	unwiredTicks := unwired.Update(5 * tickDelay)
	wiredTicks := wired.Update(5 * tickDelay)

	// then
	if unwiredTicks != wiredTicks {
		t.Fatalf("wired kernel ran %d ticks, unwired ran %d: the budget must not change", wiredTicks, unwiredTicks)
	}

	if unwiredTicks == 0 {
		t.Fatal("Update ran no ticks, want the startup burst to run the due ticks")
	}
}

// demoStream writes a small but mixed recording: pointer motion, a click, wheel
// motion, a key press and release, a quit, all split across two tick records.
func demoStream(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer

	if err := demo.WriteHeader(&buf, 7); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}

	w := demo.NewWriter(&buf)

	records := []demo.Record{
		demo.MouseMotionRecord{X: 10, Y: 20},
		demo.MouseButtonRecord{Button: 1, X: 10, Y: 20, Mod: 0, Down: true},
		demo.TickRecord{Progress: 40},
		demo.MouseWheelRecord{X: -1, Y: 2, Mod: 4},
		demo.KeyRecord{Sym: 27, Mod: 0, Down: true},
		demo.KeyRecord{Sym: 27, Mod: 0},
		demo.TickRecord{Progress: 80},
		demo.QuitRecord{},
	}

	for _, rec := range records {
		if err := w.WriteRecord(rec); err != nil {
			t.Fatalf("WriteRecord %T: %v", rec, err)
		}
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return buf.Bytes()
}

// replayKernel builds a kernel whose handlers turn every input command into one
// event, seeded from seed, and replays stream through the demo adapter.
func replayKernel(t *testing.T, seed uint32, stream []byte) (uint64, int) {
	t.Helper()

	k := kernel.New(kernel.DropUnnamed(), kernel.NewStreams(seed, "ai", "items"))
	for _, kind := range []kernel.Kind{
		kernel.KindMouseMotion,
		kernel.KindMouseButton,
		kernel.KindMouseWheel,
		kernel.KindKey,
		kernel.KindQuit,
	} {
		k.Register(kind, func(c kernel.Command, s *kernel.Streams) []kernel.Event {
			roll, _ := s.Draw("ai")

			return []kernel.Event{{Kind: kernel.KindPointerMoved, A: c.A, B: c.B, C: c.C, Stream: "ai", Roll: roll}}
		})
	}

	hash, err := game.NewDemoReplay(k).Replay(bytes.NewReader(stream))
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	return hash, len(k.Trace())
}
