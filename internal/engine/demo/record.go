package demo

import (
	"errors"
	"fmt"
	"io"
)

// recorder serialises game events and tick boundaries into a demo stream in
// the order the game produced them.
//
// TODO remove nolint:unused // demo recording half; lands with the timedemo record bean (devilGo-w9or)
type recorder struct {
	w          *Writer
	saveNumber uint32
}

// newRecorder writes a header stamped with saveNumber and returns a recorder
// for the same stream.
//
// TODO remove nolint:unused // devilGo-w9or
func newRecorder(dst io.Writer, saveNumber uint32) (*recorder, error) {
	if err := writeHeader(dst, saveNumber); err != nil {
		return nil, err
	}
	return &recorder{w: NewWriter(dst), saveNumber: saveNumber}, nil
}

// SaveNumber returns the save-game id stamped into the stream header.
//
// TODO remove nolint:unused // devilGo-w9or
func (r *recorder) SaveNumber() uint32 { return r.saveNumber }

// RecordEvent appends one input event to the stream.
//
// TODO remove nolint:unused // devilGo-w9or
func (r *recorder) RecordEvent(rec Record) error {
	switch rec.(type) {
	case tickRecord, renderRecord:
		return errors.New("demo: RecordEvent rejects frame records, use RecordFrame")
	default:
		return r.w.WriteRecord(rec)
	}
}

// RecordFrame appends a game tick or render-frame boundary to the stream.
//
// TODO remove nolint:unused // devilGo-w9or
func (r *recorder) RecordFrame(rec Record) error {
	switch rec.(type) {
	case tickRecord, renderRecord:
		return r.w.WriteRecord(rec)
	default:
		return errors.New("demo: RecordFrame rejects event records, use RecordEvent")
	}
}

// Close flushes buffered bytes to the underlying writer.
//
// TODO remove nolint:unused // devilGo-w9or
func (r *recorder) Close() error {
	if err := r.w.Flush(); err != nil {
		return fmt.Errorf("demo: flush recording: %w", err)
	}
	return nil
}
