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
}

// NewDemoReplay binds the adapter to the kernel that will run the passes.
func NewDemoReplay(k *kernel.Kernel) *DemoReplay {
	return &DemoReplay{kernel: k}
}

// Replay runs the whole stream through the adapter and returns the event-trace
// hash: the same stream and the same kernel seed must yield the same hash.
func (d *DemoReplay) Replay(stream io.Reader) (uint64, error) {
	if err := demo.NewPlayback(stream, d).Run(); err != nil {
		return 0, fmt.Errorf("game: demo replay: %w", err)
	}

	return d.kernel.TraceHash(), nil
}

// Tick closes the tick boundary the stream recorded: the buffered inputs are
// submitted at the current tick and the kernel runs its single ordered pass.
func (d *DemoReplay) Tick(uint8) {
	for _, cmd := range d.pending {
		d.kernel.Submit(cmd)
	}

	d.pending = nil

	d.kernel.Pass(d.tick)
	d.tick++
}

// MouseMotion records the recorded pointer position for the current tick.
func (d *DemoReplay) MouseMotion(x, y uint16) {
	d.buffer(kernel.Command{Kind: kernel.KindMouseMotion, A: int32(x), B: int32(y)})
}

// MouseButton records a button press or release for the current tick.
func (d *DemoReplay) MouseButton(button uint8, x, y, mod uint16, down bool) {
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
	d.buffer(kernel.Command{Kind: kernel.KindMouseWheel, A: int32(x), B: int32(y), C: int32(mod)})
}

// Key records a key press or release for the current tick.
func (d *DemoReplay) Key(sym uint32, mod uint16, down bool) {
	d.buffer(kernel.Command{Kind: kernel.KindKey, A: int32(sym), B: int32(mod), Down: down})
}

// Quit submits the quit command at once: a stream may end right after it, with
// no further tick record to flush a buffered command.
func (d *DemoReplay) Quit() {
	d.kernel.Submit(kernel.Command{Tick: d.tick, Kind: kernel.KindQuit})
}

// buffer holds one input command until the tick boundary that owns it.
func (d *DemoReplay) buffer(cmd kernel.Command) {
	cmd.Tick = d.tick
	d.pending = append(d.pending, cmd)
}
