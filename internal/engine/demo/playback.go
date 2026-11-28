package demo

import (
	"fmt"
	"io"
)

// consumer receives decoded demo records in stream order.
type Consumer interface {
	Tick(progress uint8)
	MouseMotion(x, y uint16)
	MouseButton(button uint8, x, y, mod uint16, down bool)
	MouseWheel(x, y int16, mod uint16)
	Key(sym uint32, mod uint16, down bool)
	Quit()
}

// renderConsumer is an optional Consumer extension: a consumer that implements
// it is told render-only frames apart from tick boundaries. Without it both
// arrive as Tick, which is what a plain consumer wants.
type renderConsumer interface {
	Render(progress uint8)
}

// customConsumer is an optional Consumer extension for recorded SDL custom
// events, which carry no payload and no command (demomode.cpp:589-596).
type customConsumer interface {
	Custom(eventType uint8)
}

// quitPasser is an optional Consumer extension: the pass a quit record owns
// before Quit is delivered, kept apart from a recorded game tick so a consumer
// can count the two record kinds separately.
type quitPasser interface {
	QuitPass()
}

// Playback replays a recorded demo stream into a consumer.
type Playback struct {
	stream   io.Reader
	reader   *Reader
	consumer Consumer
	framing  Framing
	started  bool
}

// makePlayback binds a demo stream to a consumer. The stream must begin with
// a valid header, which Run validates before decoding the first record.
// NewPlayback builds a playback over this engine's 5-byte-header framing;
// upstream's public entry is StartPlay (demomode.cpp).
func NewPlayback(stream io.Reader, c Consumer) *Playback {
	return makePlayback(stream, c, EngineFraming)
}

// NewPlaybackFramed builds a playback over a stream with the given framing: an
// upstream .dmo carries the settings block after the save number
// (demomode.cpp:157-187), so its first record sits at offset 32, not 5.
func NewPlaybackFramed(stream io.Reader, c Consumer, framing Framing) *Playback {
	return makePlayback(stream, c, framing)
}

func makePlayback(stream io.Reader, c Consumer, framing Framing) *Playback {
	return &Playback{stream: stream, consumer: c, framing: framing}
}

// Run drains the stream into the consumer.
//
// Run never sleeps or reads a clock: a paced mode would add its own
// throttle rather than this loop growing one.
func (p *Playback) Run() error {
	if err := p.start(); err != nil {
		return err
	}

	for {
		rec, err := p.reader.NextRecord()
		if err != nil {
			return fmt.Errorf("demo: playback: %w", err)
		}

		if rec == nil {
			return nil
		}

		p.dispatch(rec)
	}
}

// start lazily validates the header and opens the record reader once.
// Mirrors the first-run branch of Run (demomode.cpp).
func (p *Playback) start() error {
	if p.started {
		return nil
	}

	p.started = true
	if _, err := ReadHeader(p.stream); err != nil {
		return fmt.Errorf("demo: playback: %w", err)
	}

	if p.framing == UpstreamFraming {
		if _, err := ReadSettings(p.stream); err != nil {
			return fmt.Errorf("demo: playback: %w", err)
		}
	}

	p.reader = NewReader(p.stream)

	return nil
}

func (p *Playback) dispatch(rec Record) {
	switch r := rec.(type) {
	case TickRecord:
		p.consumer.Tick(r.Progress)
	case RenderRecord:
		p.dispatchRender(r.Progress)
	case QuitRecord:
		// The quit record still owns one tick pass before Quit is delivered.
		p.dispatchQuitPass()
		p.consumer.Quit()
	case CustomRecord:
		p.dispatchCustom(r.Type)
	default:
		p.dispatchInput(rec)
	}
}

// dispatchInput delivers one pointer or keyboard record to the consumer.
func (p *Playback) dispatchInput(rec Record) {
	switch r := rec.(type) {
	case MouseMotionRecord:
		p.consumer.MouseMotion(r.X, r.Y)
	case MouseButtonRecord:
		p.consumer.MouseButton(r.Button, r.X, r.Y, r.Mod, r.Down)
	case MouseWheelRecord:
		p.consumer.MouseWheel(r.X, r.Y, r.Mod)
	case KeyRecord:
		p.consumer.Key(r.Sym, r.Mod, r.Down)
	default:
		// unreachable: Record is sealed to this package and every kind is handled above
	}
}

// dispatchCustom hands a recorded custom event to a consumer that wants it.
func (p *Playback) dispatchCustom(eventType uint8) {
	if custom, ok := p.consumer.(customConsumer); ok {
		custom.Custom(eventType)
	}
}

// dispatchQuitPass runs the pass a quit record owns.
func (p *Playback) dispatchQuitPass() {
	if passer, ok := p.consumer.(quitPasser); ok {
		passer.QuitPass()

		return
	}

	p.consumer.Tick(0)
}

// dispatchRender hands a render-only frame to a consumer that separates
// render frames from tick boundaries, else treats it as the tick pass it owns.
func (p *Playback) dispatchRender(progress uint8) {
	if render, ok := p.consumer.(renderConsumer); ok {
		render.Render(progress)

		return
	}

	p.consumer.Tick(progress)
}
