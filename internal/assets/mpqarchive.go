package assets

import (
	"bytes"
	"compress/bzip2"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// MPQ on-disk constants. The magic is "MPQ\x1A" read as a little-endian word;
// the header is 32 bytes in format version 0 (StormLib SBaseFileTable.cpp).
const (
	mpqMagic      = 0x1A51504D
	mpqHeaderSize = 32
	mpqEntrySize  = 16
	mpqEntryWords = mpqEntrySize / 4

	// mpqListfileName is the record name MPQ archives use for their own name list.
	mpqListfileName = "(listfile)"
)

// MPQ record flags (StormLib MPQ_FILE_*, SBaseCommon.cpp:16-32).
const (
	mpqFileImplode    = 0x00000100
	mpqFileCompress   = 0x00000200
	mpqFileEncrypted  = 0x00010000
	mpqFileFixKey     = 0x00020000
	mpqFileSingleUnit = 0x01000000
	mpqFileExists     = 0x80000000
)

// MPQ multi-method compression bytes carried in the first sector byte when
// MPQ_FILE_COMPRESS is set (mpq_stream.c:162-231).
const (
	mpqCompressPKWare = 0x08
	mpqCompressZlib   = 0x02
	mpqCompressBzip2  = 0x10
)

// mpqHashEntry is one hash table slot.
type mpqHashEntry struct {
	hashA      uint32
	hashB      uint32
	locale     uint16
	platform   uint16
	blockIndex uint32
}

// MPQArchive is a read-only MPQ archive that reads records by name.
//
// It replaces go-mpq's read path because that path returns IMPLODE payloads
// verbatim (silent corruption) and skips decryption for sectors whose length is
// not a multiple of four; see docs/pow/asset-verification-wave.md F7. go-mpq
// still writes the synthetic archives the tests build.
type MPQArchive struct {
	file       *os.File
	path       string
	sectorSize uint32
	hashTable  []mpqHashEntry
	blockTable []MPQRecord
}

// mpqLayout is the fixed part of an MPQ header: where the tables sit and how big
// their entries are.
type mpqLayout struct {
	sectorSize uint32
	hashPos    uint32
	blockPos   uint32
	hashCount  uint32
	blockCount uint32
}

// MPQRecord is one block table entry: where a record's payload sits, how big it
// is compressed and decompressed, and the flags that say how to read it.
type MPQRecord struct {
	FilePos        uint32
	CompressedSize uint32
	FileSize       uint32
	Flags          uint32
}

// Used reports whether the record is referenced by the hash table; unused
// blocks carry no flags (DIABDAT.MPQ has one such record).
func (r MPQRecord) Used() bool {
	return r.Flags&mpqFileExists != 0
}

// Imploded reports whether the record's sectors are PKWARE DCL compressed.
func (r MPQRecord) Imploded() bool {
	return r.Flags&mpqFileImplode != 0
}

// Compressed reports whether the record's sectors carry an MPQ compression
// method byte.
func (r MPQRecord) Compressed() bool {
	return r.Flags&mpqFileCompress != 0
}

// Encrypted reports whether the record's sectors are ARC-encrypted.
func (r MPQRecord) Encrypted() bool {
	return r.Flags&mpqFileEncrypted != 0
}

// OpenMPQArchive opens a format-version-0 MPQ archive and decrypts its tables.
func OpenMPQArchive(path string) (*MPQArchive, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open mpq %s: %w", path, err)
	}

	archive := &MPQArchive{file: file, path: path}

	if err := archive.readTables(); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("mpq %s: %w (closing after: %w)", path, err, closeErr)
		}

		return nil, fmt.Errorf("mpq %s: %w", path, err)
	}

	return archive, nil
}

// Close releases the archive file.
func (a *MPQArchive) Close() error {
	if err := a.file.Close(); err != nil {
		return fmt.Errorf("close mpq %s: %w", a.path, err)
	}

	return nil
}

// RecordCount reports how many block records the archive holds, used or not.
func (a *MPQArchive) RecordCount() int {
	return len(a.blockTable)
}

// Record returns the block record at index.
func (a *MPQArchive) Record(index int) MPQRecord {
	return a.blockTable[index]
}

// RecordIndex resolves an inner name to its block index.
func (a *MPQArchive) RecordIndex(name string) (int, bool) {
	if len(a.hashTable) == 0 {
		return 0, false
	}

	index := mpqHash(name, mpqHashTableOffset) & uint32(len(a.hashTable)-1)
	hashA := mpqHash(name, mpqHashNameA)
	hashB := mpqHash(name, mpqHashNameB)

	for probe := range len(a.hashTable) {
		entry := a.hashTable[(index+uint32(probe))&uint32(len(a.hashTable)-1)]
		if entry.hashA != hashA || entry.hashB != hashB {
			continue
		}

		if entry.blockIndex >= uint32(len(a.blockTable)) {
			return 0, false
		}

		return int(entry.blockIndex), true
	}

	return 0, false
}

// HasFile reports whether the archive holds name.
func (a *MPQArchive) HasFile(name string) bool {
	_, found := a.RecordIndex(name)

	return found
}

// Names lists the archive's inner names as its (listfile) record spells them.
// An archive without that record (retail DIABDAT.MPQ, hellfire.mpq) returns nil.
func (a *MPQArchive) Names() ([]string, error) {
	if !a.HasFile(mpqListfileName) {
		return nil, nil
	}

	content, err := a.ReadFile(mpqListfileName)
	if err != nil {
		return nil, fmt.Errorf("read listfile: %w", err)
	}

	names := make([]string, 0, 256)

	for line := range bytes.SplitSeq(content, []byte("\n")) {
		name := string(bytes.TrimRight(line, "\r"))
		if name == "" {
			continue
		}

		names = append(names, name)
	}

	return names, nil
}

// ReadFile extracts and decompresses one record.
func (a *MPQArchive) ReadFile(name string) ([]byte, error) {
	index, found := a.RecordIndex(name)
	if !found {
		return nil, fmt.Errorf("mpq %s: file %q not found: %w", a.path, name, os.ErrNotExist)
	}

	record := a.blockTable[index]

	// A record is only ARC-encrypted when it says so; the derived key is a hash
	// of the name, so passing it unconditionally would decrypt stored payloads.
	if !record.Encrypted() {
		return a.readRecord(record, 0)
	}

	return a.readRecord(record, mpqFileKey(name, record))
}

// readTables parses the header and decrypts the hash and block tables.
func (a *MPQArchive) readTables() error {
	layout, err := a.readLayout()
	if err != nil {
		return err
	}

	a.sectorSize = layout.sectorSize

	hashTable, err := a.readTable(layout.hashPos, layout.hashCount, "(hash table)", "hash")
	if err != nil {
		return err
	}

	blockTable, err := a.readTable(layout.blockPos, layout.blockCount, "(block table)", "block")
	if err != nil {
		return err
	}

	a.hashTable = buildHashTable(hashTable)
	a.blockTable = buildBlockTable(blockTable)

	return nil
}

// readLayout reads and validates the archive header.
func (a *MPQArchive) readLayout() (mpqLayout, error) {
	magic := make([]byte, 4)
	if _, err := a.file.ReadAt(magic, 0); err != nil {
		return mpqLayout{}, fmt.Errorf("bad magic: header shorter than 4 bytes: %w", err)
	}

	if value := binary.LittleEndian.Uint32(magic); value != mpqMagic {
		return mpqLayout{}, fmt.Errorf("bad magic %08X, want %08X", value, uint32(mpqMagic))
	}

	header := make([]byte, mpqHeaderSize)
	if _, err := a.file.ReadAt(header, 0); err != nil {
		return mpqLayout{}, fmt.Errorf("read header: %w", err)
	}

	if version := binary.LittleEndian.Uint16(header[12:]); version != 0 {
		return mpqLayout{}, fmt.Errorf("unsupported format version %d", version)
	}

	return mpqLayout{
		sectorSize: 512 << binary.LittleEndian.Uint16(header[14:]),
		hashPos:    binary.LittleEndian.Uint32(header[16:]),
		blockPos:   binary.LittleEndian.Uint32(header[20:]),
		hashCount:  binary.LittleEndian.Uint32(header[24:]),
		blockCount: binary.LittleEndian.Uint32(header[28:]),
	}, nil
}

// buildHashTable groups the decrypted hash words into entries.
func buildHashTable(words []uint32) []mpqHashEntry {
	entries := make([]mpqHashEntry, len(words)/mpqEntryWords)

	for i := range entries {
		base := i * mpqEntryWords
		entries[i] = mpqHashEntry{
			hashA:      words[base],
			hashB:      words[base+1],
			locale:     uint16(words[base+2]),
			platform:   uint16(words[base+2] >> 16),
			blockIndex: words[base+3],
		}
	}

	return entries
}

// buildBlockTable groups the decrypted block words into records.
func buildBlockTable(words []uint32) []MPQRecord {
	records := make([]MPQRecord, len(words)/mpqEntryWords)

	for i := range records {
		base := i * mpqEntryWords
		records[i] = MPQRecord{
			FilePos:        words[base],
			CompressedSize: words[base+1],
			FileSize:       words[base+2],
			Flags:          words[base+3],
		}
	}

	return records
}

// readTable reads and decrypts one table of count 16-byte entries, keyed by the
// special name MPQ hashes for that table — "(hash table)" or "(block table)".
func (a *MPQArchive) readTable(pos, count uint32, source, what string) ([]uint32, error) {
	raw := make([]byte, count*mpqEntrySize)
	if _, err := a.file.ReadAt(raw, int64(pos)); err != nil {
		return nil, fmt.Errorf("read %s table: %w", what, err)
	}

	words := make([]uint32, count*4)
	for i := range words {
		words[i] = binary.LittleEndian.Uint32(raw[i*4:])
	}

	mpqDecryptWords(words, mpqHash(source, mpqHashFileKey))

	return words, nil
}

// readRecord decompresses one record, decrypting when it carries a file key.
func (a *MPQArchive) readRecord(record MPQRecord, key uint32) ([]byte, error) {
	payload := make([]byte, record.CompressedSize)
	if _, err := a.file.ReadAt(payload, int64(record.FilePos)); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	if !sectored(record) {
		if key != 0 {
			mpqDecryptBytes(payload, key)
		}

		return decompressSector(payload, record, record.FileSize)
	}

	return a.readSectors(payload, record, key)
}

// sectored reports whether a record's payload starts with a sector offset table.
// MPQ only sectors payloads that carry a compression flag and are not flagged
// single-unit (mpq_stream.c:341); a stored record is one raw run.
func sectored(record MPQRecord) bool {
	if record.Flags&mpqFileSingleUnit != 0 {
		return false
	}

	return record.Imploded() || record.Compressed()
}

// readSectors splits a sectored record, decrypts each sector, and joins the
// decompressed output.
func (a *MPQArchive) readSectors(payload []byte, record MPQRecord, key uint32) ([]byte, error) {
	count := int((record.FileSize + a.sectorSize - 1) / a.sectorSize)

	offsets, err := sectorOffsets(payload, uint32(count), key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, record.FileSize)

	for index := range count {
		sector, err := decryptSector(payload, offsets, index, key)
		if err != nil {
			return nil, err
		}

		decoded, err := decompressSector(sector, record, a.sectorPlainSize(record, index, count))
		if err != nil {
			return nil, fmt.Errorf("sector %d: %w", index, err)
		}

		out = append(out, decoded...)
	}

	return out, nil
}

// decryptSector slices one sector out of the payload and decrypts it.
func decryptSector(payload []byte, offsets []uint32, index int, key uint32) ([]byte, error) {
	start, end := offsets[index], offsets[index+1]
	if end < start || end > uint32(len(payload)) {
		return nil, fmt.Errorf("sector %d spans %d..%d of %d bytes",
			index, start, end, len(payload))
	}

	sector := payload[start:end]
	if key != 0 {
		mpqDecryptBytes(sector, key+uint32(index))
	}

	return sector, nil
}

// sectorPlainSize reports how many bytes one sector decompresses to; the last
// sector holds the remainder.
func (a *MPQArchive) sectorPlainSize(record MPQRecord, index, count int) uint32 {
	if index == count-1 {
		return record.FileSize - uint32(index)*a.sectorSize
	}

	return a.sectorSize
}

// sectorOffsets reads the sector offset table that starts a sectored record.
// The table is key-1 encrypted like the sectors are key+i encrypted.
func sectorOffsets(payload []byte, count, key uint32) ([]uint32, error) {
	tableSize := (count + 1) * 4
	if uint32(len(payload)) < tableSize {
		return nil, fmt.Errorf("payload of %d bytes cannot carry a %d-byte sector table",
			len(payload), tableSize)
	}

	offsets := make([]uint32, count+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(payload[i*4:])
	}

	if key != 0 {
		mpqDecryptWords(offsets, key-1)
	}

	return offsets, nil
}

// decompressSector returns one sector's plaintext. A payload at least as long as
// the expected output is stored, not compressed: the encoder keeps the raw bytes
// when compression does not pay (mpq_stream.c:96-99), and a stored payload may
// carry trailing bytes past the recorded size (hellfire.mpq's two .smk videos
// carry 292 extra), so the recorded size governs the read.
func decompressSector(sector []byte, record MPQRecord, expected uint32) ([]byte, error) {
	if uint32(len(sector)) >= expected {
		return storedSector(sector, record, expected)
	}

	switch {
	case record.Imploded():
		decoded, err := ExplodeDCL(sector, int(expected))
		if err != nil {
			return nil, fmt.Errorf("explode: %w", err)
		}

		return decoded, nil
	case record.Compressed():
		return decompressMPQMethod(sector, expected)
	default:
		return nil, fmt.Errorf("payload of %d bytes, want %d, with flags %08X",
			len(sector), expected, record.Flags)
	}
}

// storedSector trims a stored payload to its recorded size.
func storedSector(sector []byte, record MPQRecord, expected uint32) ([]byte, error) {
	if uint32(len(sector)) > expected && (record.Imploded() || record.Compressed()) {
		return nil, fmt.Errorf("sector of %d bytes exceeds its %d-byte output",
			len(sector), expected)
	}

	return sector[:expected], nil
}

// decompressMPQMethod decodes one MPQ-compressed sector by its method byte
// (mpq_stream.c:162-231).
func decompressMPQMethod(sector []byte, expected uint32) ([]byte, error) {
	if len(sector) == 0 {
		return nil, errors.New("empty compressed sector")
	}

	method := sector[0]
	body := sector[1:]

	switch method {
	case mpqCompressZlib:
		return inflateZlib(body, expected)
	case mpqCompressBzip2:
		return inflateBzip2(body, expected)
	case mpqCompressPKWare:
		decoded, err := ExplodeDCL(body, int(expected))
		if err != nil {
			return nil, fmt.Errorf("pkware: %w", err)
		}

		return decoded, nil
	default:
		return nil, fmt.Errorf("unsupported compression method %02X", method)
	}
}

// inflateZlib decodes one zlib sector.
func inflateZlib(body []byte, expected uint32) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("zlib reader: %w", err)
	}

	decoded, err := inflateExact(reader, expected)

	if closeErr := reader.Close(); closeErr != nil && err == nil {
		return nil, fmt.Errorf("zlib close: %w", closeErr)
	}

	return decoded, err
}

// inflateBzip2 decodes one bzip2 sector.
func inflateBzip2(body []byte, expected uint32) ([]byte, error) {
	return inflateExact(bzip2.NewReader(bytes.NewReader(body)), expected)
}

// inflateExact reads exactly expected bytes from a decompressor.
func inflateExact(reader io.Reader, expected uint32) ([]byte, error) {
	decoded, err := io.ReadAll(io.LimitReader(reader, int64(expected)))
	if err != nil {
		return nil, fmt.Errorf("inflate: %w", err)
	}

	if uint32(len(decoded)) != expected {
		return nil, fmt.Errorf("inflated %d bytes, want %d", len(decoded), expected)
	}

	return decoded, nil
}
