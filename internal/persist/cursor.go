package persist

import "encoding/binary"

// recordWriter builds the byte image of a fixed-layout record. Every
// multi-byte field is little-endian: upstream writes records with
// Swap*LE/SDL_SwapLE memory order (Source/utils/endian_swap.hpp:27-49), which
// is byte-identical to an explicit LE write on a little-endian host.
type recordWriter struct {
	buf []byte
}

// recordReader reads a little-endian record image. The first error sticks, so
// callers can read a whole record and check once.
type recordReader struct {
	buf []byte
	off int
	err error
}

// u8 appends one byte.
func (w *recordWriter) u8(value uint8) {
	w.buf = append(w.buf, value)
}

// le16 appends a little-endian uint16.
func (w *recordWriter) le16(value uint16) {
	w.buf = binary.LittleEndian.AppendUint16(w.buf, value)
}

// le32 appends a little-endian uint32.
func (w *recordWriter) le32(value uint32) {
	w.buf = binary.LittleEndian.AppendUint32(w.buf, value)
}

// le64 appends a little-endian uint64.
func (w *recordWriter) le64(value uint64) {
	w.buf = binary.LittleEndian.AppendUint64(w.buf, value)
}

// be32 appends a big-endian uint32 (SaveHelper's WriteBE, used for the
// game-header fields at loadsave.cpp:2783-2797).
func (w *recordWriter) be32(value uint32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, value)
}

// bytes appends a fixed-size block, as file.WriteBytes does upstream.
func (w *recordWriter) bytes(block []byte) {
	w.buf = append(w.buf, block...)
}

// u8 reads one byte, or zero past the end.
func (r *recordReader) u8() uint8 {
	if r.err != nil {
		return 0
	}

	if r.off+1 > len(r.buf) {
		r.err = ErrShortRecord
		return 0
	}

	value := r.buf[r.off]
	r.off++

	return value
}

// le16 reads a little-endian uint16, or zero past the end.
func (r *recordReader) le16() uint16 {
	block := r.take(2)
	if len(block) < 2 {
		return 0
	}

	return binary.LittleEndian.Uint16(block)
}

// le32 reads a little-endian uint32, or zero past the end.
func (r *recordReader) le32() uint32 {
	block := r.take(4)
	if len(block) < 4 {
		return 0
	}

	return binary.LittleEndian.Uint32(block)
}

// le64 reads a little-endian uint64, or zero past the end.
func (r *recordReader) le64() uint64 {
	block := r.take(8)
	if len(block) < 8 {
		return 0
	}

	return binary.LittleEndian.Uint64(block)
}

// be32 reads a big-endian uint32.
func (r *recordReader) be32() uint32 {
	block := r.take(4)
	if len(block) < 4 {
		return 0
	}

	return binary.BigEndian.Uint32(block)
}

// bytes reads n raw bytes, or nil past the end.
func (r *recordReader) bytes(n int) []byte {
	block := r.take(n)
	out := make([]byte, len(block))
	copy(out, block)

	return out
}

// take returns the next n bytes, marking the reader failed when short.
func (r *recordReader) take(n int) []byte {
	if r.err != nil {
		return nil
	}

	if r.off+n > len(r.buf) {
		r.err = ErrShortRecord
		return nil
	}

	block := r.buf[r.off : r.off+n]
	r.off += n

	return block
}
