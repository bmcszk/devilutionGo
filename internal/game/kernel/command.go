// Package kernel holds the deterministic simulation kernel: typed commands in,
// one ordered pass per tick, typed events out (ADR-015, ADR-018).
//
// The laws this package owns:
//
//	Law 1: everything that changes the game sends a Command carrying its tick.
//	Law 2: every real state change emits an Event; no change, no event.
//	Law 3: one ordered pass per tick, handlers run in registration order.
//	Law 6: the Gate drops commands that cannot change anything, before dispatch.
//
// The kernel never reads a clock, never starts a goroutine, and never lets map
// iteration decide an order: handler chains are slices in registration order
// and RNG streams are kept in sorted name order.
package kernel

// Kinds. Commands are domain + verb; the events they raise are past tense.
const (
	KindPlayerMove    Kind = "PlayerMove"
	KindPlayerAttack  Kind = "PlayerAttack"
	KindMouseMotion   Kind = "MouseMotion"
	KindMouseButton   Kind = "MouseButton"
	KindMouseWheel    Kind = "MouseWheel"
	KindKey           Kind = "Key"
	KindQuit          Kind = "Quit"
	KindMonsterDied   Kind = "MonsterDied"
	KindPlayerMoved   Kind = "PlayerMoved"
	KindPlayerStruck  Kind = "PlayerStruck"
	KindPointerMoved  Kind = "PointerMoved"
	KindGameQuit      Kind = "GameQuit"
	KindTickProcessed Kind = "TickProcessed"
	// KindDebugJumpLevel is the debug warp cheat (warp.cpp:18-30); its A word
	// carries the dungeon level.
	KindDebugJumpLevel Kind = "DebugJumpLevel"
)

// Kind identifies a command or an event type by domain name. The literal
// string is part of the trace hash contract, so renaming a kind changes hashes
// (ADR-018 law 5: names live in one package).
type Kind string

// Command is one intent applied to the simulation. The payload is integer-only
// per ADR-015: no float64 enters a per-tick path.
type Command struct {
	// Tick is the tick the command belongs to (ADR-018 law 1); dispatch stamps
	// events from the pass tick, so this value documents intent only.
	Tick uint32
	// Kind names the command.
	Kind Kind
	// A, B and C are integer payload words, interpreted per Kind.
	A, B, C int32
	// Down is the pressed state for input commands.
	Down bool
}

// Event is one applied state change: the unit presentation reads and the tick
// log records (ADR-018 laws 2 and 4).
type Event struct {
	// Tick is the tick whose pass emitted the event.
	Tick uint32
	// Kind names the event.
	Kind Kind
	// A, B and C mirror the originating command payload words.
	A, B, C int32
	// Stream is the RNG stream the emitting handler drew from, empty when it
	// drew nothing.
	Stream string
	// Roll is the value drawn from Stream, when one was drawn.
	Roll uint32
}

// Gate decides whether a command may be dispatched. A dropped command emits no
// event (ADR-018 law 6).
type Gate interface {
	Allow(Command) bool
}

// Handler applies one command and returns the events it caused. An empty result
// is the honest answer for a command that changed nothing.
type Handler func(Command, *Streams) []Event
