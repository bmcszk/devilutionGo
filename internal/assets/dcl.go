package assets

import (
	"errors"
	"fmt"
)

// PKWARE DCL ("implode") decoding constants. The two stream header bytes carry
// the literal model and the log2 dictionary size (mpq_explode.h:334-340).
const (
	dclHeaderSize    = 2
	dclModelBinary   = 0
	dclModelASCII    = 1
	dclMinWindowBits = 4
	dclMaxWindowBits = 6

	// dclMinimumMatch is the shortest back-reference the encoder emits; a
	// length-2 match carries only two distance bits (mpq_explode.h:422-426).
	dclMinimumMatch = 2
)

// Shannon-Fano tables, verbatim from the reference explode implementation
// (mpq_explode.h:56-108, itself taken from StormLib's explode.c).
var (
	dclDistanceBits = []byte{
		0x02, 0x04, 0x04, 0x05, 0x05, 0x05, 0x05, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06,
		0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07,
		0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07, 0x07,
		0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08, 0x08,
	}

	dclDistanceCodes = []byte{
		0x03, 0x0D, 0x05, 0x19, 0x09, 0x11, 0x01, 0x3E, 0x1E, 0x2E, 0x0E, 0x36, 0x16, 0x26, 0x06, 0x3A,
		0x1A, 0x2A, 0x0A, 0x32, 0x12, 0x22, 0x42, 0x02, 0x7C, 0x3C, 0x5C, 0x1C, 0x6C, 0x2C, 0x4C, 0x0C,
		0x74, 0x34, 0x54, 0x14, 0x64, 0x24, 0x44, 0x04, 0x78, 0x38, 0x58, 0x18, 0x68, 0x28, 0x48, 0x08,
		0xF0, 0x70, 0xB0, 0x30, 0xD0, 0x50, 0x90, 0x10, 0xE0, 0x60, 0xA0, 0x20, 0xC0, 0x40, 0x80, 0x00,
	}

	dclLengthBits = []byte{
		0x03, 0x02, 0x03, 0x03, 0x04, 0x04, 0x04, 0x05, 0x05, 0x05, 0x05, 0x06, 0x06, 0x06, 0x07, 0x07,
	}

	dclLengthCodes = []byte{
		0x05, 0x03, 0x01, 0x06, 0x0A, 0x02, 0x0C, 0x14, 0x04, 0x18, 0x08, 0x30, 0x10, 0x20, 0x40, 0x00,
	}

	dclLengthExtraBits = []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	}

	dclLengthBase = []uint16{
		0x0000, 0x0001, 0x0002, 0x0003, 0x0004, 0x0005, 0x0006, 0x0007,
		0x0008, 0x000A, 0x000E, 0x0016, 0x0026, 0x0046, 0x0086, 0x0106,
	}

	// dclLengthCode and dclDistanceCode are the two match symbol tables.
	dclLengthCode   = dclCode{codes: dclLengthCodes, bits: dclLengthBits}
	dclDistanceCode = dclCode{codes: dclDistanceCodes, bits: dclDistanceBits}

	// errDCLTruncated reports a stream that ends before it produced the
	// expected number of bytes.
	errDCLTruncated = errors.New("dcl: stream ended before the expected output size")
)

// dclModel is the decoded two-byte stream header.
type dclModel struct {
	windowBits int
}

// dclCode is one Shannon-Fano symbol table: the code values and their widths,
// read least-significant-bit first.
type dclCode struct {
	codes []byte
	bits  []byte
}

// dclMatch is one decoded back-reference.
type dclMatch struct {
	length   int
	distance int
}

// dclDecoder walks one imploded stream.
type dclDecoder struct {
	bits       dclBits
	windowBits int
}

// dclBits reads a bit stream least-significant-bit first, refilling the
// accumulator from the stream bytes.
type dclBits struct {
	data  []byte
	pos   int
	acc   uint32
	count int
}

// ExplodeDCL decodes one PKWARE DCL "implode" stream to exactly outSize bytes.
//
// The stream starts with the two header bytes (literal model, window bits).
// Only the binary model is supported: every Diablo record that reaches this
// decoder declares model 0, including the whole retail corpus the extraction
// smoke test walks, so the ASCII model is rejected by name rather than
// half-implemented. Output is truncated to outSize because MPQ sectors carry an
// exact expected size and the encoder may end a stream with a longer match.
func ExplodeDCL(src []byte, outSize int) ([]byte, error) {
	if outSize < 0 {
		return nil, fmt.Errorf("dcl: negative output size %d", outSize)
	}

	model, err := parseDCLHeader(src)
	if err != nil {
		return nil, err
	}

	stream := &dclDecoder{
		bits:       dclBits{data: src[dclHeaderSize:]},
		windowBits: model.windowBits,
	}

	return stream.explode(outSize)
}

// parseDCLHeader validates the two header bytes and reports the window size.
func parseDCLHeader(src []byte) (dclModel, error) {
	if len(src) < dclHeaderSize {
		return dclModel{}, fmt.Errorf("dcl: %d byte(s) cannot carry the %d-byte header",
			len(src), dclHeaderSize)
	}

	if src[0] != dclModelBinary {
		return dclModel{}, fmt.Errorf("dcl: unsupported ascii literal model %d", src[0])
	}

	windowBits := int(src[1])
	if windowBits < dclMinWindowBits || windowBits > dclMaxWindowBits {
		return dclModel{}, fmt.Errorf("dcl: window bits %d outside %d..%d",
			windowBits, dclMinWindowBits, dclMaxWindowBits)
	}

	return dclModel{windowBits: windowBits}, nil
}

// explode drives the stream until outSize bytes are produced
// (mpq_explode.h:344-450).
func (d *dclDecoder) explode(outSize int) ([]byte, error) {
	out := make([]byte, 0, outSize)

	for len(out) < outSize {
		var err error

		out, err = d.appendUnit(out)
		if err != nil {
			return nil, err
		}
	}

	return out[:outSize], nil
}

// appendUnit appends the next literal byte or back-reference to out.
func (d *dclDecoder) appendUnit(out []byte) ([]byte, error) {
	isMatch, err := d.matchFlag()
	if err != nil {
		return nil, err
	}

	if !isMatch {
		value, err := d.literal()
		if err != nil {
			return nil, err
		}

		return append(out, value), nil
	}

	match, err := d.match()
	if err != nil {
		return nil, err
	}

	return appendDCLMatch(out, match)
}

// matchFlag reports whether the next unit is a back-reference or a literal.
func (d *dclDecoder) matchFlag() (bool, error) {
	flag, err := d.bits.read(1)
	if err != nil {
		return false, err
	}

	return flag == 1, nil
}

// literal reads one binary literal byte.
func (d *dclDecoder) literal() (byte, error) {
	value, err := d.bits.read(8)
	if err != nil {
		return 0, err
	}

	return byte(value), nil
}

// match reads one back-reference (mpq_explode.h:381-433).
func (d *dclDecoder) match() (dclMatch, error) {
	length, err := d.matchLength()
	if err != nil {
		return dclMatch{}, err
	}

	distance, err := d.matchDistance(length)
	if err != nil {
		return dclMatch{}, err
	}

	return dclMatch{length: length, distance: distance}, nil
}

// matchLength decodes the length symbol plus its extra bits.
func (d *dclDecoder) matchLength() (int, error) {
	index, err := d.index(dclLengthCode)
	if err != nil {
		return 0, err
	}

	length := int(dclLengthBase[index]) + dclMinimumMatch

	extraBits := int(dclLengthExtraBits[index])
	if extraBits == 0 {
		return length, nil
	}

	extra, err := d.bits.read(extraBits)
	if err != nil {
		return 0, err
	}

	return length + int(extra), nil
}

// matchDistance decodes the distance symbol and its low bits. A length-2 match
// carries two distance bits instead of the full window (mpq_explode.h:422-431).
func (d *dclDecoder) matchDistance(length int) (int, error) {
	index, err := d.index(dclDistanceCode)
	if err != nil {
		return 0, err
	}

	lowBits := d.windowBits
	if length == dclMinimumMatch {
		lowBits = 2
	}

	low, err := d.bits.read(lowBits)
	if err != nil {
		return 0, err
	}

	return index<<lowBits | int(low), nil
}

// index decodes one symbol of code, matching the stream's low bits against each
// table entry in order (mpq_explode.h:252-292).
func (d *dclDecoder) index(code dclCode) (int, error) {
	for index := range code.codes {
		width := int(code.bits[index])
		if d.bits.available() < width {
			return 0, errDCLTruncated
		}

		if d.bits.peek(width) != uint32(code.codes[index]) {
			continue
		}

		if _, err := d.bits.read(width); err != nil {
			return 0, err
		}

		return index, nil
	}

	return 0, errors.New("dcl: no table code matches the stream bits")
}

// appendDCLMatch copies match.length bytes from match.distance+1 bytes back.
// Copying byte by byte is what lets an overlapping match repeat a run
// (mpq_explode.h:444-448).
func appendDCLMatch(out []byte, match dclMatch) ([]byte, error) {
	from := len(out) - match.distance - 1
	if from < 0 {
		return nil, fmt.Errorf("dcl: match distance %d reaches before the output start", match.distance+1)
	}

	for i := range match.length {
		out = append(out, out[from+i])
	}

	return out, nil
}

// read consumes n bits.
func (b *dclBits) read(n int) (uint32, error) {
	if b.available() < n {
		return 0, errDCLTruncated
	}

	value := b.peek(n)
	b.acc >>= n
	b.count -= n

	return value, nil
}

// peek returns the next n bits without consuming them, zero-padding past the
// end of the stream.
func (b *dclBits) peek(n int) uint32 {
	b.refill(n)

	return b.acc & (1<<n - 1)
}

// available reports how many bits are buffered, refilling up to 25.
func (b *dclBits) available() int {
	b.refill(25)

	return b.count
}

// refill buffers at least n bits, or everything the stream still holds.
func (b *dclBits) refill(n int) {
	for b.count < n && b.pos < len(b.data) {
		b.acc |= uint32(b.data[b.pos]) << b.count
		b.pos++
		b.count += 8
	}
}
