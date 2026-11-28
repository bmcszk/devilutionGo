package game

import (
	"fmt"
	"io"

	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// DemoReplay turns a recorded demo stream into kernel commands, so a .dmo file
// replays through the deterministic kernel (ADR-015 consequences). It is the
// only place demo records become commands; the input direction stays
// game -> engine/demo.
//
// Record mapping, all integer payloads per ADR-015:
//
//	MouseMotion(x, y)          -> MouseMotion  A=x, B=y
//	MouseButton(button, x, y)  -> MouseButton  A=button, B=x<<16|y, C=mod, Down
//	MouseWheel(x, y, mod)      -> MouseWheel   A=x, B=y, C=mod
//	Key(sym, mod)              -> Key          A=sym, B=mod, Down
//	Quit                       -> Quit         no payload
//	Tick(progress)             -> one ordered kernel pass for the buffered
//	                              inputs, then the tick advances
type DemoReplay struct {
	kernel  *kernel.Kernel
	pending []kernel.Command
	tick    uint32
	stats   DemoReplayStats
}

// DemoReplayStats is the accounting of one replay: the record histogram, the
// kernel passes the stream drove and how many commands each pass dispatched.
// A replay fixture pins these numbers because TraceHash alone proves little
// for a kernel with no handlers (the trace stays empty).
type DemoReplayStats struct {
	// Ticks is the number of kernel passes the stream drove.
	Ticks int
	// Commands is the number of kernel commands the stream submitted.
	Commands int
	// Records is the record-type histogram of the stream.
	Records DemoRecordCounts
	// CommandsPerTick is the command count of every pass, in pass order.
	CommandsPerTick []int
}

// DemoRecordCounts is the record-type histogram of a demo stream.
type DemoRecordCounts struct {
	// GameTicks counts frame-boundary records.
	GameTicks int
	// Renders counts render-only frames.
	Renders int
	// MouseMotions, MouseButtonDowns, MouseButtonUps, MouseWheels count the
	// pointer records.
	MouseMotions, MouseButtonDowns, MouseButtonUps, MouseWheels int
	// KeyDowns and KeyUps count the keyboard records.
	KeyDowns, KeyUps int
	// Quits counts the recorded quit events.
	Quits int
	// Customs counts recorded SDL custom events, which carry no command.
	Customs int
}

// NewDemoReplay binds the adapter to the kernel that will run the passes.
func NewDemoReplay(k *kernel.Kernel) *DemoReplay {
	return &DemoReplay{kernel: k}
}

// Replay runs the whole stream through the adapter and returns the event-trace
// hash: the same stream and the same kernel seed must yield the same hash.
// The stream uses this engine's 5-byte header framing.
func (d *DemoReplay) Replay(stream io.Reader) (uint64, error) {
	return d.drain(stream, demo.EngineFraming)
}

// ReplayUpstream runs a devilutionX .dmo through the adapter. Its header
// carries the settings block between the save number and the first record
// (demomode.cpp:157-187), so the first record sits at offset 32.
func (d *DemoReplay) ReplayUpstream(stream io.Reader) (uint64, error) {
	return d.drain(stream, demo.UpstreamFraming)
}

// Stats returns the accounting of the last replay.
func (d *DemoReplay) Stats() DemoReplayStats { return d.stats }

// Tick closes the tick boundary the stream recorded: the buffered inputs are
// submitted at the current tick and the kernel runs its single ordered pass.
func (d *DemoReplay) Tick(uint8) {
	d.stats.Records.GameTicks++
	d.pass()
}

// Render closes a render-only frame, which drives the same single pass as a
// tick boundary (demomode.cpp:539-542).
func (d *DemoReplay) Render(uint8) {
	d.stats.Records.Renders++
	d.pass()
}

// QuitPass runs the pass a quit record owns before Quit is delivered: it is a
// kernel pass, but not a recorded game tick.
func (d *DemoReplay) QuitPass() { d.pass() }

// Custom accounts one recorded SDL custom event, which carries no command.
func (d *DemoReplay) Custom(uint8) {
	d.stats.Records.Customs++
}

// MouseMotion records the recorded pointer position for the current tick.
func (d *DemoReplay) MouseMotion(x, y uint16) {
	d.stats.Records.MouseMotions++
	d.buffer(kernel.Command{Kind: kernel.KindMouseMotion, A: int32(x), B: int32(y)})
}

// MouseButton records a button press or release for the current tick.
//
//nolint:revive // "down" is the recorded event direction (demomode.cpp:833-840), not a control flag.
func (d *DemoReplay) MouseButton(button uint8, x, y, mod uint16, down bool) {
	if down {
		d.stats.Records.MouseButtonDowns++
	} else {
		d.stats.Records.MouseButtonUps++
	}

	d.buffer(kernel.Command{
		Kind: kernel.KindMouseButton,
		A:    int32(button),
		B:    int32(x)<<16 | int32(y),
		C:    int32(mod),
		Down: down,
	})
}

// MouseWheel records wheel motion for the current tick.
func (d *DemoReplay) MouseWheel(x, y int16, mod uint16) {
	d.stats.Records.MouseWheels++
	d.buffer(kernel.Command{Kind: kernel.KindMouseWheel, A: int32(x), B: int32(y), C: int32(mod)})
}

// Key records a key press or release for the current tick.
//
//nolint:revive // "down" is the recorded event direction (demomode.cpp:891-901), not a control flag.
func (d *DemoReplay) Key(sym uint32, mod uint16, down bool) {
	if down {
		d.stats.Records.KeyDowns++
	} else {
		d.stats.Records.KeyUps++
	}

	d.buffer(kernel.Command{Kind: kernel.KindKey, A: int32(sym), B: int32(mod), Down: down})
}

// Quit submits the quit command at once: a stream may end right after it, with
// no further tick record to flush a buffered command.
func (d *DemoReplay) Quit() {
	d.stats.Records.Quits++
	d.stats.Commands++

	d.kernel.Submit(kernel.Command{Tick: d.tick, Kind: kernel.KindQuit})
}

// drain consumes one stream with the given framing and reports the trace hash.
func (d *DemoReplay) drain(stream io.Reader, framing demo.Framing) (uint64, error) {
	d.stats = DemoReplayStats{}

	if err := demo.NewPlaybackFramed(stream, d, framing).Run(); err != nil {
		return 0, fmt.Errorf("game: demo replay: %w", err)
	}

	return d.kernel.TraceHash(), nil
}

// pass submits the buffered inputs and runs one kernel pass.
func (d *DemoReplay) pass() {
	d.stats.Ticks++
	d.stats.CommandsPerTick = append(d.stats.CommandsPerTick, len(d.pending))
	d.stats.Commands += len(d.pending)

	for _, cmd := range d.pending {
		d.kernel.Submit(cmd)
	}

	d.pending = nil

	d.kernel.Pass(d.tick)
	d.tick++
}

// buffer holds one input command until the tick boundary that owns it.
func (d *DemoReplay) buffer(cmd kernel.Command) {
	cmd.Tick = d.tick
	d.pending = append(d.pending, cmd)
}
