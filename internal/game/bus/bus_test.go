package bus_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/game/bus"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// passScript is the deterministic three-tick command script every case runs:
// two commands per tick over three kinds, two of them drawing from a stream.
var passScript = [][]kernel.Command{
	{
		{Tick: 0, Kind: kernel.KindPlayerMove, A: 11, B: 22},
		{Tick: 0, Kind: kernel.KindPlayerAttack, A: 33, B: 44},
	},
	{
		{Tick: 1, Kind: kernel.KindMonsterDied, A: 55, B: 66},
	},
	{
		{Tick: 2, Kind: kernel.KindPlayerMove, A: 77, B: 88},
		{Tick: 2, Kind: kernel.KindMonsterDied, A: 99, B: 100},
	},
}

// spec is one case: which bus wiring to run and what to do inside a pass.
type spec struct {
	withBus     bool
	subscribers int
}

type runResult struct {
	traceHash uint64
	trace     []kernel.Event
	perTick   [][]kernel.Event
	received  [][][]kernel.Event
}

func TestPass_SubscribersDoNotChangeThePassResult(t *testing.T) {
	// given the pre-bus kernel as the reference every case must match
	baseline := runPasses(t, spec{})

	cases := []struct {
		name string
		spec spec
	}{
		{"noBus", spec{}},
		{"busZeroSubscribers", spec{withBus: true}},
		{"busThreeSubscribers", spec{withBus: true, subscribers: 3}},
	}

	// when every case wiring runs the same script
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given the case wiring
			s := tc.spec

			// when it runs the script
			got := runPasses(t, s)

			// then state and event trace are identical to the pre-bus reference
			assertSamePassResult(t, baseline, got)
		})
	}

	// then the reference itself emitted events, so the comparison is not vacuous
	if len(baseline.trace) == 0 {
		t.Fatal("reference pass emitted no events: the comparison would prove nothing")
	}
}

func TestPass_EventsArePublishedOnlyAfterThePassCompletes(t *testing.T) {
	// given a subscriber counting delivered events while the pass runs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pub := bus.New()
	t.Cleanup(func() { _ = pub.Close() })

	sub, err := pub.Subscribe(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	delivered := pumpEvents(sub)

	// onDispatch runs inside the pass, so an event in flight there means the
	// kernel published before the pass settled.
	onDispatch := func() {
		if len(delivered) > 0 {
			t.Error("a handler saw a published event while the pass was still dispatching")
		}
	}

	k := kernel.New(kernel.DropUnnamed(), nil)
	k.SetPublisher(pub)
	k.Register(kernel.KindPlayerMove, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		onDispatch()

		return []kernel.Event{{Tick: c.Tick, Kind: kernel.KindPlayerMoved, A: c.A}}
	})

	// when two passes run
	for tick := range 2 {
		k.Submit(kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerMove, A: int32(tick)})
		k.Pass(uint32(tick))

		// then this pass's event arrives only after Pass returned
		waitForEvent(t, delivered, uint32(tick))
	}
}

func TestSubscribe_MutatingReceivedEventsCannotLeak(t *testing.T) {
	// given the same pass published to two subscribers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pub := bus.New()
	t.Cleanup(func() { _ = pub.Close() })

	first, err := pub.Subscribe(ctx)
	if err != nil {
		t.Fatalf("subscribe first: %v", err)
	}

	second, err := pub.Subscribe(ctx)
	if err != nil {
		t.Fatalf("subscribe second: %v", err)
	}

	k := kernel.New(kernel.DropUnnamed(), nil)
	k.SetPublisher(pub)
	k.Register(kernel.KindPlayerMove, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		return []kernel.Event{{Tick: c.Tick, Kind: kernel.KindPlayerMoved, A: c.A}}
	})

	k.Submit(kernel.Command{Tick: 7, Kind: kernel.KindPlayerMove, A: 1})
	k.Pass(7)

	published := k.Trace()
	before := k.TraceHash()

	// when the first subscriber mutates everything it received
	batch := collect(t, first, 1)[0]
	for i := range batch {
		batch[i].Kind = kernel.KindPlayerStruck
		batch[i].A = -1
		batch[i].Tick = 99
	}

	// then the kernel trace and the other subscriber are untouched: this guard
	// fails the test if either saw the mutation.
	if after := k.TraceHash(); after != before {
		t.Fatalf("trace hash after subscriber mutation = %#x, want %#x", after, before)
	}

	if !slices.Equal(k.Trace(), published) {
		t.Fatalf("trace = %v, want %v", k.Trace(), published)
	}

	if other := collect(t, second, 1)[0]; !slices.Equal(other, published) {
		t.Fatalf("second subscriber received %v, want %v", other, published)
	}
}

func TestSubscribe_ClosingTheContextEndsTheSubscription(t *testing.T) {
	// given a subscriber whose presentation consumer is going away
	ctx, cancel := context.WithCancel(context.Background())

	pub := bus.New()
	t.Cleanup(func() { _ = pub.Close() })

	batches, err := pub.Subscribe(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	// when the context is cancelled
	cancel()

	// then the batch channel closes, so no consumer goroutine is left blocked
	select {
	case _, ok := <-batches:
		if ok {
			t.Fatal("received a batch after cancellation, want a closed channel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("batch channel stayed open after cancellation")
	}
}

func assertSamePassResult(t *testing.T, want, got runResult) {
	t.Helper()

	if got.traceHash != want.traceHash {
		t.Fatalf("trace hash = %#x, want %#x (pre-bus reference)", got.traceHash, want.traceHash)
	}

	if !slices.Equal(got.trace, want.trace) {
		t.Fatalf("trace = %v, want %v (pre-bus reference)", got.trace, want.trace)
	}

	assertSamePasses(t, want.perTick, got.perTick)
	assertSameDeliveries(t, want.trace, got.received)
}

func assertSamePasses(t *testing.T, want, got [][]kernel.Event) {
	t.Helper()

	for i, tickEvents := range got {
		if !slices.Equal(tickEvents, want[i]) {
			t.Fatalf("pass %d returned %v, want %v", i, tickEvents, want[i])
		}
	}
}

func assertSameDeliveries(t *testing.T, want []kernel.Event, got [][][]kernel.Event) {
	t.Helper()

	for i, batches := range got {
		if flat := slices.Concat(batches...); !slices.Equal(flat, want) {
			t.Fatalf("subscriber %d received %v, want %v", i, flat, want)
		}
	}
}

func runPasses(t *testing.T, s spec) runResult {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k := newScriptedKernel()
	subscriptions := subscribe(ctx, t, k, s)

	res := runResult{perTick: runScript(k)}

	res.traceHash = k.TraceHash()
	res.trace = k.Trace()

	res.received = make([][][]kernel.Event, len(subscriptions))
	for i, ch := range subscriptions {
		res.received[i] = collect(t, ch, len(res.perTick))
	}

	return res
}

// newScriptedKernel registers one handler per scripted kind; the move and
// attack handlers draw from a named stream so events carry rolls.
func newScriptedKernel() *kernel.Kernel {
	k := kernel.New(kernel.DropUnnamed(), kernel.NewStreams(0xA5A5, "ai", "items"))

	k.Register(kernel.KindPlayerMove, func(c kernel.Command, st *kernel.Streams) []kernel.Event {
		roll, _ := st.Draw("ai")

		return []kernel.Event{{
			Tick: c.Tick, Kind: kernel.KindPlayerMoved, A: c.A, B: c.B,
			Stream: "ai", Roll: roll,
		}}
	})
	k.Register(kernel.KindPlayerAttack, func(c kernel.Command, st *kernel.Streams) []kernel.Event {
		roll, _ := st.Draw("items")

		return []kernel.Event{{
			Tick: c.Tick, Kind: kernel.KindPlayerStruck, A: c.A, B: c.B,
			Stream: "items", Roll: roll,
		}}
	})
	k.Register(kernel.KindMonsterDied, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		return []kernel.Event{{Tick: c.Tick, Kind: kernel.KindMonsterDied, A: c.A, B: c.B, C: c.C}}
	})

	return k
}

func runScript(k *kernel.Kernel) [][]kernel.Event {
	perTick := make([][]kernel.Event, 0, len(passScript))

	for _, tickCommands := range passScript {
		for _, c := range tickCommands {
			k.Submit(c)
		}

		perTick = append(perTick, k.Pass(tickCommands[0].Tick))
	}

	return perTick
}

func subscribe(
	ctx context.Context, t *testing.T, k *kernel.Kernel, s spec,
) []<-chan []kernel.Event {
	t.Helper()

	if !s.withBus {
		if s.subscribers != 0 {
			t.Fatalf("case asks for %d subscribers without a bus", s.subscribers)
		}

		return nil
	}

	pub := bus.New()
	t.Cleanup(func() { _ = pub.Close() })
	k.SetPublisher(pub)

	chans := make([]<-chan []kernel.Event, 0, s.subscribers)

	for range s.subscribers {
		ch, err := pub.Subscribe(ctx)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		chans = append(chans, ch)
	}

	return chans
}

func pumpEvents(sub <-chan []kernel.Event) chan kernel.Event {
	events := make(chan kernel.Event, 4)

	go func() {
		for batch := range sub {
			for _, ev := range batch {
				events <- ev
			}
		}
	}()

	return events
}

// collect reads want batches, ordering them by tick: cross-pass delivery order
// is not guaranteed (see bus.Publisher), so only batch content is asserted.
func collect(t *testing.T, ch <-chan []kernel.Event, want int) [][]kernel.Event {
	t.Helper()

	got := make([][]kernel.Event, 0, want)

	for len(got) < want {
		select {
		case batch, ok := <-ch:
			if !ok {
				t.Fatalf("subscriber channel closed after %d of %d batches", len(got), want)
			}

			got = append(got, batch)
		case <-time.After(2 * time.Second):
			t.Fatalf("subscriber received %d of %d batches before the timeout", len(got), want)
		}
	}

	slices.SortFunc(got, func(a, b []kernel.Event) int { return int(a[0].Tick) - int(b[0].Tick) })

	return got
}

func waitForEvent(t *testing.T, delivered <-chan kernel.Event, want uint32) {
	t.Helper()

	select {
	case ev := <-delivered:
		if ev.Tick != want {
			t.Fatalf("delivered event tick = %d, want %d", ev.Tick, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("pass %d published nothing after it completed", want)
	}
}
