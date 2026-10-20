// Package demo: playback driver.
package demo

import (
	"fmt"
	"io"
)

// Consumer receives decoded demo records in stream order. The game wiring
// (forwarding these calls into the running game loop, mirroring upstream's
// handlers at demomode.cpp:626 and :674) happens in a later bean; this
// interface is the seam it will implement.
type Consumer interface {
	Tick(progress uint8)
	MouseMotion(x, y uint16)
	MouseButton(button uint8, x, y, mod uint16, down bool)
	MouseWheel(x, y int16, mod uint16)
	Key(sym uint32, mod uint16, down bool)
	Quit()
}

// Playback replays a recorded demo stream into a Consumer.
type Playback struct {
	stream   io.Reader
	reader   *Reader
	consumer Consumer
	started  bool
}

// NewPlayback binds a demo stream to a consumer. The stream must start with
// a valid header; Run validates it before decoding the first record.
func NewPlayback(stream io.Reader, c Consumer) *Playback {
	return &Playback{stream: stream, consumer: c}
}

// Run drains the stream into the consumer.
//
// This is the --timedemo wiring point: upstream's GetRunGameLoop branches on
// Timedemo to skip wall-clock pacing so the replay runs uncapped
// (demomode.cpp:644-645). The uncapped semantics here are structural — Run
// never sleeps and consults no clock — so a future paced mode would add its
// own throttle rather than this loop growing one.
func (p *Playback) Run() error {
	if !p.started {
		p.started = true
		if _, err := ReadHeader(p.stream); err != nil {
			return fmt.Errorf("demo: playback: %w", err)
		}
		p.reader = NewReader(p.stream)
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

func (p *Playback) dispatch(rec Record) {
	switch r := rec.(type) {
	case TickRecord:
		p.consumer.Tick(r.Progress)
	case RenderRecord:
		p.consumer.Tick(r.Progress)
	case MouseMotionRecord:
		p.consumer.MouseMotion(r.X, r.Y)
	case MouseButtonRecord:
		p.consumer.MouseButton(r.Button, r.X, r.Y, r.Mod, r.Down)
	case MouseWheelRecord:
		p.consumer.MouseWheel(r.X, r.Y, r.Mod)
	case KeyRecord:
		p.consumer.Key(r.Sym, r.Mod, r.Down)
	case QuitRecord:
		// The quit message still owns one loop pass — upstream injects the
		// SDL_QUIT via FetchMessage during that pass (demomode.cpp:674-687)
		// and the loop exits only on the NEXT GetRunGameLoop call.
		p.consumer.Tick(0)
		p.consumer.Quit()
	default:
		// unreachable: Record implementations are sealed to this package
	}
}
