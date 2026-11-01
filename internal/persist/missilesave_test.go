package persist_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestOverflowCount(t *testing.T) {
	// when the overflow of a live missile list is counted
	// (loadsave.cpp:1870-1871)
	cases := []struct{ total, want int }{
		{0, 0}, {125, 0}, {126, 1}, {200, 75},
	}
	// then only the part above MaxMissilesForSaveGame counts
	for _, tc := range cases {
		if got := persist.OverflowCount(tc.total); got != tc.want {
			t.Fatalf("OverflowCount(%d) = %d, want %d", tc.total, got, tc.want)
		}
	}
}

func TestAdditionalMissilesEncodeGolden(t *testing.T) {
	// given no overflow missiles
	save := persist.AdditionalMissiles{}
	// when the stream is encoded (loadsave.cpp:1872-1875)
	image, err := save.Encode()
	// then it is version 0 plus a zero count, little-endian
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	want := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	if !bytes.Equal(image, want) {
		t.Fatalf("image = %x, want %x", image, want)
	}
}

func TestAdditionalMissilesRecordStride(t *testing.T) {
	// given two overflow missile records
	first := make([]byte, persist.MissileSaveSize)
	first[0] = 1
	second := make([]byte, persist.MissileSaveSize)
	second[0] = 2

	save := persist.AdditionalMissiles{Records: [][]byte{first, second}}
	image, err := save.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when the stream is decoded back
	decoded, err := persist.DecodeAdditionalMissiles(image)
	// then the count prefix and both 180-byte records round-trip
	if err != nil {
		t.Fatalf("DecodeAdditionalMissiles: %v", err)
	}

	if got := binary.LittleEndian.Uint32(image[4:8]); got != 2 {
		t.Fatalf("count = %d, want 2", got)
	}

	if len(image) != persist.AdditionalMissilesHeaderSize+2*persist.MissileSaveSize {
		t.Fatalf("len = %d", len(image))
	}

	if !bytes.Equal(decoded.Records[0], first) || !bytes.Equal(decoded.Records[1], second) {
		t.Fatal("records changed")
	}
}

func TestAdditionalMissilesRejectsWrongStride(t *testing.T) {
	// given a record shorter than the SaveMissile stride
	save := persist.AdditionalMissiles{Records: [][]byte{make([]byte, 10)}}
	// when it is encoded
	_, err := save.Encode()
	// then the stride mismatch is reported
	if err == nil {
		t.Fatal("short missile record encoded")
	}
}

func TestDecodeAdditionalMissilesNewerVersionIgnored(t *testing.T) {
	// given a stream from a newer build (loadsave.cpp:1896-1899)
	buf := make([]byte, persist.AdditionalMissilesHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:4], persist.VersionAdditionalMissiles+1)
	// when it is decoded
	save, err := persist.DecodeAdditionalMissiles(buf)
	// then it is ignored without an error
	if err != nil {
		t.Fatalf("DecodeAdditionalMissiles: %v", err)
	}

	if len(save.Records) != 0 {
		t.Fatalf("records = %d, want 0", len(save.Records))
	}
}

func TestDecodeAdditionalMissilesTruncated(t *testing.T) {
	// given a stream claiming one record but carrying none
	buf := make([]byte, persist.AdditionalMissilesHeaderSize)
	binary.LittleEndian.PutUint32(buf[4:8], 1)
	// when it is decoded
	_, err := persist.DecodeAdditionalMissiles(buf)
	// then the truncation is reported
	if err == nil {
		t.Fatal("truncated stream decoded without error")
	}
}
