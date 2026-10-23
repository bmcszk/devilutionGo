// Package demo: frame format, version 3, per devilutionX demomode.cpp.
package demo

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// Version is the demo stream format version; upstream writes it as the
// first byte of every .dmo file (demomode.cpp:41).
const Version uint8 = 3

// Record tag values, named after DemoMsg::EventType (demomode.cpp:46-63).
const (
	tagGameTick     uint8 = 0
	tagMouseMotion  uint8 = 9
	tagMouseDown    uint8 = 10
	tagMouseUp      uint8 = 11
	tagMouseWheel   uint8 = 12
	tagKeyDown      uint8 = 13
	tagKeyUp        uint8 = 14
	tagQuit         uint8 = 8
	tagRenderPrefix uint8 = 0x80
)

// Record is one entry in a demo stream: a frame boundary or an input event.
type Record interface {
	record()
}

// tickRecord marks a game logic tick; Progress carries
// progressToNextGameTick (0–127) as recorded upstream (demomode.cpp:795-797).
type tickRecord struct{ Progress uint8 }

// renderRecord marks a render-only frame between ticks, decoded from the
// packed 0x80|progress tag byte (demomode.cpp:539-542).
type renderRecord struct{ Progress uint8 }

// mouseMotionRecord carries the pointer position (demomode.cpp:813-817).
type mouseMotionRecord struct{ X, Y uint16 }

// mouseButtonRecord carries a button press or release (demomode.cpp:833-840).
type mouseButtonRecord struct {
	Button uint8
	X, Y   uint16
	Mod    uint16
	Down   bool
}

// mouseWheelRecord carries wheel motion; X and Y are signed like
// SDL_MouseWheelEvent (demomode.cpp:846-851).
type mouseWheelRecord struct {
	X, Y int16
	Mod  uint16
}

// keyRecord carries a key press or release (demomode.cpp:891-901).
type keyRecord struct {
	Sym  uint32
	Mod  uint16
	Down bool
}

// quitRecord marks the SDL quit event, which upstream records with an empty
// payload (demomode.cpp:907-919).
type quitRecord struct{}

func (tickRecord) record()        {}
func (renderRecord) record()      {}
func (mouseMotionRecord) record() {}
func (mouseButtonRecord) record() {}
func (mouseWheelRecord) record()  {}
func (keyRecord) record()         {}
func (quitRecord) record()        {}

// Writer serialises records in upstream byte order. Stream framing (version
// byte + save number) is separate; recorder owns it via writeHeader.
type Writer struct {
	dst *bufio.Writer
}

// writeHeader appends the version byte and LE32 gSaveNumber that open every
// .dmo file (demomode.cpp:519-523).
func writeHeader(dst io.Writer, saveNumber uint32) error {
	header := []byte{Version, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(header[1:], saveNumber)
	if _, err := dst.Write(header); err != nil {
		return fmt.Errorf("demo: write header: %w", err)
	}
	return nil
}

// readHeader consumes the version byte and LE32 gSaveNumber, rejecting any
// version this reader cannot decode: upstream refuses versions above its own
// (demomode.cpp:519-523) and this reader implements version 3 only.
func readHeader(src io.Reader) (uint32, error) {
	var b [5]byte
	if _, err := io.ReadFull(src, b[:]); err != nil {
		return 0, fmt.Errorf("demo: read header: %w", err)
	}
	if v := b[0]; v != Version {
		return 0, fmt.Errorf("demo: unsupported version %d (want %d)", v, Version)
	}
	return binary.LittleEndian.Uint32(b[1:]), nil
}

// NewWriter buffers writes to dst; call Flush when done recording.
func NewWriter(dst io.Writer) *Writer {
	return &Writer{dst: bufio.NewWriter(dst)}
}

// WriteRecord appends one record. It rejects renderRecord values with
// Progress > 127 because the packed tag byte can only hold seven bits
// (demomode.cpp:592-599).
func (w *Writer) WriteRecord(rec Record) error {
	switch r := rec.(type) {
	case renderRecord:
		return w.writeRender(r.Progress)
	case tickRecord:
		w.writeTag(tagGameTick, r.Progress)
		return nil
	case mouseMotionRecord:
		w.writeTag(tagMouseMotion, 0)
		writeLE16(w.dst, r.X)
		writeLE16(w.dst, r.Y)
		return nil
	case mouseButtonRecord:
		tag := tagMouseUp
		if r.Down {
			tag = tagMouseDown
		}
		w.writeTag(tag, 0)
		writeByte(w.dst, r.Button)
		writeLE16(w.dst, r.X)
		writeLE16(w.dst, r.Y)
		writeLE16(w.dst, r.Mod)
		return nil
	case mouseWheelRecord:
		w.writeTag(tagMouseWheel, 0)
		writeLE16(w.dst, uint16(r.X))
		writeLE16(w.dst, uint16(r.Y))
		writeLE16(w.dst, r.Mod)
		return nil
	case keyRecord:
		tag := tagKeyUp
		if r.Down {
			tag = tagKeyDown
		}
		w.writeTag(tag, 0)
		writeLE32(w.dst, r.Sym)
		writeLE16(w.dst, r.Mod)
		return nil
	case quitRecord:
		w.writeTag(tagQuit, 0)
		return nil
	default:
		return fmt.Errorf("demo: unsupported record type %T", rec)
	}
}

// Flush pushes buffered bytes to the underlying writer.
func (w *Writer) Flush() error { return w.dst.Flush() }

func (w *Writer) writeRender(progress uint8) error {
	if progress > 127 {
		return fmt.Errorf("demo: render progress %d exceeds 127", progress)
	}
	return w.dst.WriteByte(tagRenderPrefix | progress)
}

func (w *Writer) writeTag(tag, progress uint8) {
	writeByte(w.dst, tag)
	writeByte(w.dst, progress)
}

func writeByte(w *bufio.Writer, v uint8) {
	if err := w.WriteByte(v); err != nil {
		panic(fmt.Sprintf("demo: bufio write: %v", err))
	}
}

func writeLE16(w *bufio.Writer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	writeAll(w, b[:])
}

func writeLE32(w *bufio.Writer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	writeAll(w, b[:])
}

// writeAll swallows the sticky bufio error; Flush is the checkpoint.
func writeAll(w *bufio.Writer, b []byte) {
	if _, err := w.Write(b); err != nil {
		panic(fmt.Sprintf("demo: bufio write: %v", err))
	}
}

// Reader decodes record payloads from a stream whose header was already
// consumed via readHeader.
type Reader struct {
	src *bufio.Reader
}

// NewReader starts decoding records at the current stream position.
func NewReader(src io.Reader) *Reader {
	return &Reader{src: bufio.NewReader(src)}
}

// NextRecord decodes the next record and returns nil at end of stream.
func (r *Reader) NextRecord() (Record, error) {
	tag, err := r.src.ReadByte()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("demo: read tag: %w", err)
	}
	if tag&tagRenderPrefix != 0 {
		return renderRecord{Progress: tag & 0x7F}, nil
	}
	return r.decodeTagged(tag)
}

func (r *Reader) decodeTagged(tag uint8) (Record, error) {
	progress, err := r.src.ReadByte()
	if err == io.EOF && tag == tagGameTick {
		// Upstream treats feof after the tag byte as end-of-stream
		// (demomode.cpp:535-538): a bare tick tag is the natural terminator.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("demo: read progress: %w", err)
	}
	switch tag {
	case tagGameTick:
		return tickRecord{Progress: progress}, nil
	case tagMouseMotion:
		return r.decodeMouseMotion()
	case tagMouseDown, tagMouseUp:
		return r.decodeMouseButton(tag == tagMouseDown)
	case tagMouseWheel:
		return r.decodeMouseWheel()
	case tagKeyDown, tagKeyUp:
		return r.decodeKey(tag == tagKeyDown)
	case tagQuit:
		return quitRecord{}, nil
	default:
		return nil, fmt.Errorf("demo: unknown record tag %d", tag)
	}
}

func (r *Reader) decodeMouseMotion() (Record, error) {
	x, err := r.le16()
	if err != nil {
		return nil, err
	}
	y, err := r.le16()
	if err != nil {
		return nil, err
	}
	return mouseMotionRecord{X: x, Y: y}, nil
}

func (r *Reader) decodeMouseButton(down bool) (Record, error) {
	button, err := r.src.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("demo: read button: %w", err)
	}
	x, err := r.le16()
	if err != nil {
		return nil, err
	}
	y, err := r.le16()
	if err != nil {
		return nil, err
	}
	mod, err := r.le16()
	if err != nil {
		return nil, err
	}
	return mouseButtonRecord{Button: button, X: x, Y: y, Mod: mod, Down: down}, nil
}

func (r *Reader) decodeMouseWheel() (Record, error) {
	x, err := r.le16()
	if err != nil {
		return nil, err
	}
	y, err := r.le16()
	if err != nil {
		return nil, err
	}
	mod, err := r.le16()
	if err != nil {
		return nil, err
	}
	return mouseWheelRecord{X: int16(x), Y: int16(y), Mod: mod}, nil
}

func (r *Reader) decodeKey(down bool) (Record, error) {
	sym, err := r.le32()
	if err != nil {
		return nil, err
	}
	mod, err := r.le16()
	if err != nil {
		return nil, err
	}
	return keyRecord{Sym: sym, Mod: mod, Down: down}, nil
}

func (r *Reader) le16() (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r.src, b[:]); err != nil {
		return 0, fmt.Errorf("demo: read LE16: %w", err)
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

func (r *Reader) le32() (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r.src, b[:]); err != nil {
		return 0, fmt.Errorf("demo: read LE32: %w", err)
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}
