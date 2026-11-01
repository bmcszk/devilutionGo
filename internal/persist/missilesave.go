package persist

import "fmt"

// Additional-missile constants (Source/loadsave.cpp:1868-1890).
const (
	// AdditionalMissilesEntryName is the archive entry SaveAdditionalMissiles
	// writes.
	AdditionalMissilesEntryName = "additionalMissiles"
	// VersionAdditionalMissiles is the stream version this build writes.
	VersionAdditionalMissiles uint32 = 0
	// MissileSaveSize is BytesWrittenBySaveMissile, the fixed SaveMissile
	// record length.
	MissileSaveSize = 180
	// AdditionalMissilesHeaderSize is the version plus count prefix.
	AdditionalMissilesHeaderSize = 8
)

// AdditionalMissiles is the overflow list of missiles past
// MaxMissilesForSaveGame (Source/loadsave.cpp:1868-1890). The records are the
// opaque SaveMissile images of the missiles package, kept at their fixed
// 180-byte stride.
type AdditionalMissiles struct {
	// Records holds one SaveMissile image per overflow missile.
	Records [][]byte
}

// OverflowCount is the recorded missile count for a live missile list: the
// part above MaxMissilesForSaveGame (loadsave.cpp:1870-1871).
func OverflowCount(total int) int {
	if total <= MaxMissilesForSaveGame {
		return 0
	}

	return total - MaxMissilesForSaveGame
}

// Encode returns the additionalMissiles stream: version, count, then the
// fixed-stride records (loadsave.cpp:1872-1888).
func (a AdditionalMissiles) Encode() ([]byte, error) {
	out := make([]byte, 0, AdditionalMissilesHeaderSize+len(a.Records)*MissileSaveSize)

	var header recordWriter

	header.u32(VersionAdditionalMissiles)
	header.u32(uint32(len(a.Records)))
	out = append(out, header.buf...)

	for i, record := range a.Records {
		if len(record) != MissileSaveSize {
			return nil, fmt.Errorf("persist: missile record %d len %d != %d: %w", i, len(record),
				MissileSaveSize, ErrShortRecord)
		}

		out = append(out, record...)
	}

	return out, nil
}

// DecodeAdditionalMissiles parses the additionalMissiles stream; a newer
// version is ignored like upstream does (loadsave.cpp:1896-1899).
func DecodeAdditionalMissiles(buf []byte) (AdditionalMissiles, error) {
	in := recordReader{buf: buf}

	version := in.le32()
	if version > VersionAdditionalMissiles {
		return AdditionalMissiles{}, nil
	}

	count := in.le32()
	if in.err != nil {
		return AdditionalMissiles{}, fmt.Errorf("persist: additional missiles header: %w", in.err)
	}

	if uint64(count)*MissileSaveSize > uint64(len(buf)) {
		return AdditionalMissiles{}, fmt.Errorf("persist: additional missiles count %d: %w", count, ErrShortRecord)
	}

	save := AdditionalMissiles{Records: make([][]byte, 0, count)}

	for range count {
		record := in.bytes(MissileSaveSize)
		if in.err != nil {
			return AdditionalMissiles{}, fmt.Errorf("persist: additional missiles record: %w", in.err)
		}

		save.Records = append(save.Records, record)
	}

	return save, nil
}
