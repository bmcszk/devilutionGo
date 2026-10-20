// Package demo: recording driver.
package demo

import (
	"errors"
	"fmt"
	"io"
)

// Recorder accepts game events and tick boundaries in the order the game
// produced them and serialises them into a demo stream.
type Recorder struct {
	w          *Writer
	saveNumber uint32
}

// NewRecorder opens a recording stream stamped with the current
// gSaveNumber, mirroring the header upstream writes once per file
// (demomode.cpp:523).
func NewRecorder(dst io.Writer, saveNumber uint32) (*Recorder, error) {
	if err := WriteHeader(dst, saveNumber); err != nil {
		return nil, err
	}
	return &Recorder{w: NewWriter(dst), saveNumber: saveNumber}, nil
}

// SaveNumber returns the save-game id stamped into the stream header.
func (r *Recorder) SaveNumber() uint32 { return r.saveNumber }

// RecordEvent appends one input event (demomode.cpp:799-924).
func (r *Recorder) RecordEvent(rec Record) error {
	switch rec.(type) {
	case TickRecord, RenderRecord:
		return errors.New("demo: RecordEvent rejects frame records, use RecordFrame")
	default:
		return r.w.WriteRecord(rec)
	}
}

// RecordFrame appends a game tick or render-frame boundary
// (RecordGameLoopResult, demomode.cpp:795-797).
func (r *Recorder) RecordFrame(rec Record) error {
	switch rec.(type) {
	case TickRecord, RenderRecord:
		return r.w.WriteRecord(rec)
	default:
		return errors.New("demo: RecordFrame rejects event records, use RecordEvent")
	}
}

// Close flushes buffered bytes to the underlying writer.
func (r *Recorder) Close() error {
	if err := r.w.Flush(); err != nil {
		return fmt.Errorf("demo: flush recording: %w", err)
	}
	return nil
}
