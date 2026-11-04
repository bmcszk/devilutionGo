package kernel_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// mutatingPublisher overwrites the events it receives in place, the way a
// badly behaved presentation consumer would.
type mutatingPublisher struct{}

func (mutatingPublisher) Publish(events []kernel.Event) {
	for i := range events {
		events[i] = kernel.Event{Tick: 99, Kind: kernel.KindPlayerStruck, A: -1}
	}
}

func TestPass_TwoHandlersOnOneCommandEmitInRegistrationOrder(t *testing.T) {
	// given
	var fired []string

	k := kernel.New(nil, nil)
	k.Register(kernel.KindPlayerMove, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		fired = append(fired, "first")

		return []kernel.Event{{Kind: kernel.KindPlayerMoved, A: c.A, B: c.B}}
	})
	k.Register(kernel.KindPlayerMove, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		fired = append(fired, "second")

		return []kernel.Event{{Kind: kernel.KindPlayerStruck, A: c.A, B: c.B}}
	})

	// when
	for tick := range 2 {
		k.Submit(kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerMove, A: 1, B: 2})
		k.Pass(uint32(tick))
	}

	// then
	want := []string{"first", "second", "first", "second"}
	if !slices.Equal(fired, want) {
		t.Fatalf("handler order = %v, want %v (registration order, every pass)", fired, want)
	}

	got := make([]kernel.Kind, 0, len(k.Trace()))
	for _, ev := range k.Trace() {
		got = append(got, ev.Kind)
	}

	wantEvents := []kernel.Kind{
		kernel.KindPlayerMoved, kernel.KindPlayerStruck,
		kernel.KindPlayerMoved, kernel.KindPlayerStruck,
	}
	if !slices.Equal(got, wantEvents) {
		t.Fatalf("event order = %v, want %v", got, wantEvents)
	}
}

func TestPass_RunsOncePerTick(t *testing.T) {
	// given
	runs := 0

	k := kernel.New(nil, nil)
	k.Register(kernel.KindKey, func(kernel.Command, *kernel.Streams) []kernel.Event {
		runs++

		return []kernel.Event{{Kind: kernel.KindPointerMoved}}
	})
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindKey})
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindKey})

	// when
	first := k.Pass(0)
	runsAfterFirst := runs
	second := k.Pass(0)
	k.Submit(kernel.Command{Tick: 1, Kind: kernel.KindKey})
	third := k.Pass(1)

	// then
	if len(first) != 2 || runsAfterFirst != 2 {
		t.Fatalf("first pass ran %d handlers for %d events, want 2", runsAfterFirst, len(first))
	}

	if second != nil {
		t.Fatalf("second pass for tick 0 emitted %d events, want 0: one pass per tick", len(second))
	}

	if len(third) != 1 {
		t.Fatalf("tick 1 pass emitted %d events, want 1", len(third))
	}
}

func TestSubmit_GateDropsCommandAndEmitsNothing(t *testing.T) {
	// given
	runs := 0

	k := kernel.New(kernel.DropUnnamed(), kernel.NewStreams(1, "ai"))
	k.Register(kernel.KindQuit, func(kernel.Command, *kernel.Streams) []kernel.Event {
		runs++

		return []kernel.Event{{Kind: kernel.KindGameQuit}}
	})

	// when
	accepted := k.Submit(kernel.Command{Tick: 0})
	events := k.Pass(0)

	// then
	if accepted {
		t.Fatal("gate accepted an unnamed command, want it dropped")
	}

	if runs != 0 {
		t.Fatalf("dropped command reached a handler %d times, want 0", runs)
	}

	if len(events) != 0 || len(k.Trace()) != 0 {
		t.Fatalf("dropped command emitted %d events, want 0", len(events))
	}
}

func TestPass_HandlerDrawingFromStreamIsDeterministic(t *testing.T) {
	// given
	draw := func() kernel.Event { return drawOnce() }

	// when
	first := draw()
	second := draw()
	unknown := kernel.NewStreams(42, "ai").Names()

	// then
	if first != second {
		t.Fatalf("same seed drew %v then %v, want the same roll", first, second)
	}

	if first.Stream != "ai" {
		t.Fatalf("event stream = %q, want ai", first.Stream)
	}

	if !slices.Equal(unknown, []string{"ai"}) {
		t.Fatalf("stream names = %v, want [ai]", unknown)
	}
}

func TestTraceHash_DiffersOnDifferentTrace(t *testing.T) {
	// given
	build := func(roll uint32) *kernel.Kernel {
		k := kernel.New(nil, nil)
		k.Register(kernel.KindKey, func(kernel.Command, *kernel.Streams) []kernel.Event {
			return []kernel.Event{{Kind: kernel.KindPointerMoved, Roll: roll}}
		})
		k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindKey})
		k.Pass(0)

		return k
	}

	// when
	a := build(1).TraceHash()
	b := build(1).TraceHash()
	c := build(2).TraceHash()

	// then
	if a != b {
		t.Fatalf("identical traces hashed %d and %d, want equal", a, b)
	}

	if a == c {
		t.Fatalf("different traces share hash %d, want a differing roll to change the hash", a)
	}
}

func TestTraceHash_PerTickHookMatchesTheWholeTrace(t *testing.T) {
	// given
	run := func() ([]uint64, uint64) {
		k := kernel.New(nil, kernel.NewStreams(7, "ai"))
		k.Register(kernel.KindPlayerAttack, func(c kernel.Command, s *kernel.Streams) []kernel.Event {
			roll, _ := s.Draw("ai")

			return []kernel.Event{{Kind: kernel.KindPlayerStruck, A: c.A, Stream: "ai", Roll: roll}}
		})

		perTick := make([]uint64, 0, 3)

		for tick := range 3 {
			k.Submit(kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerAttack, A: int32(tick)})
			perTick = append(perTick, kernel.TraceHash(k.Pass(uint32(tick))))
		}

		return perTick, k.TraceHash()
	}

	// when
	firstTicks, firstWhole := run()
	secondTicks, secondWhole := run()

	// then
	if !slices.Equal(firstTicks, secondTicks) {
		t.Fatalf("per-tick hashes = %v then %v, want the same sequence", firstTicks, secondTicks)
	}

	if firstWhole != secondWhole {
		t.Fatalf("whole-trace hash = %d then %d, want equal", firstWhole, secondWhole)
	}
}

func TestPass_APublisherCannotReachTheSimulation(t *testing.T) {
	// given a kernel whose publisher rewrites everything it is handed, the way
	// a badly behaved presentation consumer would
	k := kernel.New(nil, nil)
	k.Register(kernel.KindMonsterDied, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		return []kernel.Event{{Tick: c.Tick, Kind: kernel.KindMonsterDied, A: c.A}}
	})
	k.SetPublisher(mutatingPublisher{})

	// when
	k.Submit(kernel.Command{Tick: 4, Kind: kernel.KindMonsterDied, A: 7})
	returned := k.Pass(4)

	// then the pass result and the trace still carry the emitted event
	want := []kernel.Event{{Tick: 4, Kind: kernel.KindMonsterDied, A: 7}}
	if !slices.Equal(returned, want) {
		t.Fatalf("pass returned %v, want %v", returned, want)
	}

	if !slices.Equal(k.Trace(), want) {
		t.Fatalf("trace = %v, want %v", k.Trace(), want)
	}
}

// drawOnce seeds "ai" and "items" from 42, dies once and returns the event the
// AI handler emitted, so two calls prove the same seed replays the same roll.
func drawOnce() kernel.Event {
	k := kernel.New(nil, kernel.NewStreams(42, "ai", "items"))
	k.Register(kernel.KindMonsterDied, func(_ kernel.Command, s *kernel.Streams) []kernel.Event {
		roll, ok := s.Draw("ai")
		if !ok {
			return nil
		}

		return []kernel.Event{{Kind: kernel.KindMonsterDied, Stream: "ai", Roll: roll}}
	})
	k.Submit(kernel.Command{Tick: 3, Kind: kernel.KindMonsterDied})
	k.Pass(3)

	return k.Trace()[0]
}
