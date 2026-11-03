package assets

import "encoding/binary"

// MPQ hash types (StormLib MPQ_HASH_TABLE_OFFSET..MPQ_HASH_FILE_KEY, SBaseCommon.cpp:42-45).
const (
	mpqHashTableOffset = 0
	mpqHashNameA       = 1
	mpqHashNameB       = 2
	mpqHashFileKey     = 3
)

// mpqCryptTable is the 0x500-entry MPQ crypt table. Entry 0x400..0x4FF mixes
// encryption keys, the rest index the name hash (SBaseCommon.cpp:52-95).
var mpqCryptTable = newMPQCryptTable()

// newMPQCryptTable builds the crypt table from its linear-congruential seed.
func newMPQCryptTable() [0x500]uint32 {
	var table [0x500]uint32

	seed := uint32(0x00100001)

	for index1 := range 0x100 {
		index2 := index1

		for range 5 {
			seed = (seed*125 + 3) % 0x2AAAAB
			high := (seed & 0xFFFF) << 0x10

			seed = (seed*125 + 3) % 0x2AAAAB
			low := seed & 0xFFFF

			table[index2] = high | low
			index2 += 0x100
		}
	}

	return table
}

// mpqHash hashes an MPQ inner name for hashType, uppercasing ASCII and folding
// forward slashes to backslashes first (SBaseCommon.cpp:262-305).
func mpqHash(name string, hashType uint32) uint32 {
	seed1 := uint32(0x7FED7FED)
	seed2 := uint32(0xEEEEEEEE)

	for i := range len(name) {
		ch := name[i]
		if ch >= 'a' && ch <= 'z' {
			ch -= 'a' - 'A'
		}

		if ch == '/' {
			ch = '\\'
		}

		seed1 = mpqCryptTable[hashType*0x100+uint32(ch)] ^ (seed1 + seed2)
		seed2 = uint32(ch) + seed1 + seed2 + (seed2 << 5) + 3
	}

	return seed1
}

// mpqDecryptBytes decrypts the complete 32-bit words of data in place with key.
// A trailing partial word is left untouched, which is what StormLib's
// DecryptMpqBlock callers do by passing a word count (SBaseCommon.cpp:490-530):
// silently skipping the whole buffer when its length is not a multiple of four
// is the defect that made every imploded sector unreadable.
func mpqDecryptBytes(data []byte, key uint32) {
	words := len(data) / 4
	decoded := make([]uint32, words)

	for i := range words {
		decoded[i] = binary.LittleEndian.Uint32(data[i*4:])
	}

	mpqDecryptWords(decoded, key)

	for i, word := range decoded {
		binary.LittleEndian.PutUint32(data[i*4:], word)
	}
}

// mpqDecryptWords decrypts words in place with key, mixing the key and seed the
// way StormLib's DecryptMpqBlock does (SBaseCommon.cpp:490-530).
func mpqDecryptWords(words []uint32, key uint32) {
	seed := uint32(0xEEEEEEEE)

	for i, value := range words {
		seed += mpqCryptTable[0x400+(key&0xFF)]

		plain := value ^ (key + seed)

		key = ((^key << 0x15) + 0x11111111) | (key >> 0x0B)
		seed = plain + seed + (seed << 5) + 3

		words[i] = plain
	}
}

// mpqFileKey derives the decryption key of one record. The key hashes the bare
// file name and, when the record carries MPQ_FILE_FIX_KEY, mixes in its
// position and size (SBaseCommon.cpp:681-700).
func mpqFileKey(name string, record MPQRecord) uint32 {
	key := mpqHash(mpqBaseName(name), mpqHashFileKey)

	if record.Flags&mpqFileFixKey != 0 {
		key = (key + record.FilePos) ^ record.FileSize
	}

	return key
}

// mpqBaseName returns the name after the last path separator, which is the part
// the file key hashes.
func mpqBaseName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '\\' || name[i] == '/' {
			return name[i+1:]
		}
	}

	return name
}
