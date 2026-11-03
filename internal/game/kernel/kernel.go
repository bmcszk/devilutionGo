package kernel

import (
	"encoding/binary"
	"hash/fnv"
	"slices"
)

// Kernel applies commands to state in one ordered pass per tick and collects
// the events those changes raise. It is the only code that drives handlers
// (ADR-018 law 3): no goroutines, no clock reads, no locks.
type Kernel struct {
	gate    Gate
	streams *Streams
	chains  map[Kind][]Handler
	queue   []Command
	trace   []Event
	last    uint32
	passed  bool
}

// New returns a kernel with no handlers. A nil gate accepts every command; a
// nil streams value draws nothing, which keeps RNG use explicit.
func New(gate Gate, streams *Streams) *Kernel {
	return &Kernel{gate: gate, streams: streams, chains: make(map[Kind][]Handler)}
}

// Register appends a handler for kind. Registration order is dispatch order,
// permanently: every later pass runs the chain in the same order.
func (k *Kernel) Register(kind Kind, h Handler) {
	k.chains[kind] = append(k.chains[kind], h)
}

// Submit queues one command, applying the gate first. It reports whether the
// gate accepted the command; a rejected command never reaches a handler and
// emits nothing.
func (k *Kernel) Submit(c Command) bool {
	if k.gate != nil && !k.gate.Allow(c) {
		return false
	}

	k.queue = append(k.queue, c)

	return true
}

// Pass runs the one ordered pass for tick: it drains the queue in submission
// order, runs each command's handlers in registration order, records the
// resulting events in the trace, and returns them. A second call for a tick
// already passed runs nothing and returns nil, so a tick has exactly one pass.
func (k *Kernel) Pass(tick uint32) []Event {
	if k.passed && tick <= k.last {
		return nil
	}

	k.passed = true
	k.last = tick

	queue := k.queue
	k.queue = nil

	events := make([]Event, 0, len(queue))

	for _, cmd := range queue {
		events = append(events, k.dispatch(tick, cmd)...)
	}

	k.trace = append(k.trace, events...)

	return events
}

// Trace returns the ordered event trace since the kernel was created.
func (k *Kernel) Trace() []Event { return slices.Clone(k.trace) }

// TraceHash is the checksum of the whole trace so far.
func (k *Kernel) TraceHash() uint64 { return TraceHash(k.trace) }

// TraceHash is the event-trace checksum the tick log and desync checks compare
// (ADR-018 consequences): FNV-1a 64 over tick, kind, payload, stream and roll,
// in trace order. Same events, same hash, on every run. The tick log feeds it
// one tick's events at a time; a replay compares the whole trace with
// TraceHash(k.Trace()).
func TraceHash(events []Event) uint64 {
	h := fnv.New64a()

	var buf [4]byte

	for _, ev := range events {
		binary.LittleEndian.PutUint32(buf[:], ev.Tick)
		_, _ = h.Write(buf[:])

		_, _ = h.Write([]byte(ev.Kind))

		for _, word := range [3]int32{ev.A, ev.B, ev.C} {
			binary.LittleEndian.PutUint32(buf[:], uint32(word))
			_, _ = h.Write(buf[:])
		}

		_, _ = h.Write([]byte(ev.Stream))

		binary.LittleEndian.PutUint32(buf[:], ev.Roll)
		_, _ = h.Write(buf[:])
	}

	return h.Sum64()
}

// dispatch runs one command's chain in registration order and stamps every
// event with the pass tick.
func (k *Kernel) dispatch(tick uint32, cmd Command) []Event {
	chain := k.chains[cmd.Kind]

	var out []Event

	for _, h := range chain {
		for _, ev := range h(cmd, k.streams) {
			ev.Tick = tick
			out = append(out, ev)
		}
	}

	return out
}
