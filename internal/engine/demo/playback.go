package demo

import (
	"fmt"
	"io"
)

// Consumer receives decoded demo records in stream order.
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

// NewPlayback binds a demo stream to a consumer. The stream must begin with
// a valid header, which Run validates before decoding the first record.
func NewPlayback(stream io.Reader, c Consumer) *Playback {
	return &Playback{stream: stream, consumer: c}
}

// Run drains the stream into the consumer.
//
// Run never sleeps or reads a clock: a paced mode would add its own
// throttle rather than this loop growing one.
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
		// The quit record still owns one tick pass before Quit is delivered.
		p.consumer.Tick(0)
		p.consumer.Quit()
	default:
		// unreachable: Record is sealed to this package
	}
}
