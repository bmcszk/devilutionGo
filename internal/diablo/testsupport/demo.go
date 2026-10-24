// Package testsupport: demo-tape playback foundations (devilGo-xcdz
// slice 2) — the recorded demo format of upstream demomode.cpp.
package testsupport

import (
	"encoding/binary"
	"fmt"
	"testing"
)

// Demo record types (upstream demomode.cpp DemoMsg::EventType).
const (
	DemoGameTick = 0
	DemoRender   = 1
	DemoQuit     = 8 // SDL_QUIT
)

// demoStep is one parsed demo record: type, tick-progress and the typed
// payload (empty for GameTick/Rendering-burst/Quit/custom events).
type demoStep struct {
	// Type is the EventType ordinal; Rendering events may arrive packed
	// (high bit set, progress in the low bits — upstream WriteDemoMsgHeader).
	Type byte
	// Progress is progressToNextGameTick.
	Progress byte
	// Motion/Button/Wheel/Key decode the payload per Type:
	//   MouseMotion (9): X, Y
	//   MouseButtonDown/Up (10, 11): Button, X, Y, Mod
	//   MouseWheel (12): X, Y (int16), Mod
	//   KeyDown/Up (13, 14): Sym (u32), Mod
	Button, Sym, Mod int
	X, Y             int
}

// IsInput reports whether the step carries an input payload.
func (s demoStep) IsInput() bool { return s.Type >= 8 }

// demoTape is a parsed .dmo recording: the fixed header plus the ordered
// step list.
type demoTape struct {
	// Version is the recording format version (3 current).
	Version int
	// SaveNumber is the recorded save slot (gSaveNumber).
	SaveNumber uint32
	// GraphicsWidth/Height is the recorded resolution.
	GraphicsWidth, GraphicsHeight int
	// steps is the ordered demo message list.
	steps []demoStep
}

// Steps reports the number of demo messages.
func (t *demoTape) Steps() int { return len(t.steps) }

// Step returns the i-th demo message (ordered as recorded).
func (t *demoTape) Step(i int) demoStep { return t.steps[i] }

// loadDemoTapeBytes parses an already decoded committed demo recording.
func loadDemoTapeBytes(t *testing.T, data []byte) *demoTape {
	t.Helper()
	parsed, err := parseDemoTape(data)
	if err != nil {
		t.Fatalf("parse committed demo fixture: %v", err)
	}
	return parsed
}

// parseDemoTape parses the version-3 layout (upstream demomode.cpp
// OpenDemoFile/ReadDemoMessage): u8 version, LE32 saveNumber, LE16 width,
// LE16 height, 16 per-feature bytes, 6 potion-counter bytes, then the
// ordered DemoMsg stream.
func parseDemoTape(data []byte) (*demoTape, error) {
	tape, err := parseDemoTapeHeader(data)
	if err != nil {
		return nil, err
	}

	for stream := data[31:]; len(stream) > 0; {
		step, consumed, err := decodeDemoStep(stream)
		if err != nil {
			return nil, err
		}
		tape.steps = append(tape.steps, step)
		stream = stream[consumed:]
	}
	return tape, nil
}

// parseDemoTapeHeader parses the fixed version-3 header (upstream
// demomode.cpp OpenDemoFile): u8 version, LE32 saveNumber, LE16 width,
// LE16 height, plus the feature/potion byte blocks up to offset 31; the
// remaining bytes are the ordered DemoMsg stream.
func parseDemoTapeHeader(data []byte) (*demoTape, error) {
	if len(data) < 31 {
		return nil, fmt.Errorf("size %d, want >= 31 header", len(data))
	}
	tape := &demoTape{}
	tape.Version = int(data[0])
	if tape.Version > 3 {
		return nil, fmt.Errorf("unsupported demo version %d", tape.Version)
	}
	tape.SaveNumber = binary.LittleEndian.Uint32(data[1:])
	tape.GraphicsWidth = int(binary.LittleEndian.Uint16(data[5:]))
	tape.GraphicsHeight = int(binary.LittleEndian.Uint16(data[7:]))
	return tape, nil
}

// decodeDemoStep decodes one demo record from the front of the stream and
// returns it with the number of bytes consumed (upstream demomode.cpp
// ReadDemoMessage): packed rendering events carry progress inline in the
// low type bits; other events are type, progress, then a typed payload.
func decodeDemoStep(stream []byte) (demoStep, int, error) {
	step := demoStep{Type: stream[0]}
	if step.Type&0x80 != 0 {
		// Packed rendering event: progress inline in the low bits.
		step.Type = DemoRender
		step.Progress = stream[0] & 0x7F
		return step, 1, nil
	}
	if len(stream) < 2 {
		return step, 0, fmt.Errorf("truncated progress byte after type %d", step.Type)
	}
	step.Progress = stream[1]
	consumed := 2
	payload := stream[2:]

	switch step.Type {
	case DemoGameTick, DemoQuit, DemoRender:
		// no payload
	case 9: // MouseMotionEvent
		if len(payload) < 4 {
			return step, 0, fmt.Errorf("truncated motion payload after type %d", step.Type)
		}
		step.X, step.Y = readLe16Pair(payload)
		consumed += 4
	case 10, 11: // MouseButtonDownEvent, MouseButtonUpEvent
		if len(payload) < 7 {
			return step, 0, fmt.Errorf("truncated button payload after type %d", step.Type)
		}
		step.Button = int(payload[0])
		step.X, step.Y = readLe16Pair(payload[1:])
		step.Mod = int(binary.LittleEndian.Uint16(payload[5:]))
		consumed += 7
	case 12: // MouseWheelEvent
		if len(payload) < 6 {
			return step, 0, fmt.Errorf("truncated wheel payload after type %d", step.Type)
		}
		step.X, step.Y = readLe16Pair(payload)
		step.Mod = int(binary.LittleEndian.Uint16(payload[4:]))
		consumed += 6
	case 13, 14: // KeyDownEvent, KeyUpEvent
		if len(payload) < 6 {
			return step, 0, fmt.Errorf("truncated key payload after type %d", step.Type)
		}
		step.Sym = int(binary.LittleEndian.Uint32(payload))
		step.Mod = int(binary.LittleEndian.Uint16(payload[4:]))
		consumed += 6
	default:
		if step.Type < 64 {
			return step, 0, fmt.Errorf("unknown event type %d", step.Type)
		}
		// custom event: no payload
	}
	return step, consumed, nil
}

// readLe16Pair pops two LE u16 values (x, y) off the stream.
func readLe16Pair(stream []byte) (x, y int) {
	return int(binary.LittleEndian.Uint16(stream)),
		int(binary.LittleEndian.Uint16(stream[2:]))
}
