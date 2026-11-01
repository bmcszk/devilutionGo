package persist_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestLevelSeedsEncodeGolden(t *testing.T) {
	// given two generated levels and the rest absent
	seeds := persist.LevelSeeds{{Present: true, Seed: 0x01020304}, {}, {Present: true, Seed: 7}}
	// when the Diablo stream is encoded (loadsave.cpp:1911-1917)
	image := seeds.Encode(persist.GameMode{})
	// then each level is a presence byte plus a LE seed only when present
	want := []byte{0x01, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x07, 0x00, 0x00, 0x00}
	if !bytes.HasPrefix(image, want) {
		t.Fatalf("head = %x, want prefix %x", image[:len(want)], want)
	}

	if len(image) != 25 {
		t.Fatalf("len = %d, want 25 (17 levels: 2 seeds of 5 bytes + 15 flags)", len(image))
	}
}

func TestLevelSeedsHellfireLength(t *testing.T) {
	// when the Hellfire stream is encoded (LevelsHellfire = 25)
	image := persist.LevelSeeds{}.Encode(persist.GameMode{Hellfire: true})
	// then every level carries a single absent flag
	if len(image) != 25 {
		t.Fatalf("len = %d, want 25", len(image))
	}

	for i, flag := range image {
		if flag != 0 {
			t.Fatalf("level %d flag = %d, want 0", i, flag)
		}
	}
}

func TestLevelSeedsRoundTrip(t *testing.T) {
	// given a seeds list with gaps
	seeds := persist.LevelSeeds{{}, {Present: true, Seed: 42}, {}, {Present: true, Seed: 0xdeadbeef}}
	// when it is encoded and decoded back
	image := seeds.Encode(persist.GameMode{})

	decoded, err := persist.DecodeLevelSeeds(image, persist.GameMode{})
	// then the present entries keep their seeds and the rest stay absent
	if err != nil {
		t.Fatalf("DecodeLevelSeeds: %v", err)
	}

	if len(decoded) != 17 {
		t.Fatalf("len = %d, want 17", len(decoded))
	}

	if decoded[1] != seeds[1] || decoded[3] != seeds[3] {
		t.Fatalf("decoded = %+v", decoded[:4])
	}

	if decoded[0].Present || decoded[2].Present {
		t.Fatalf("absent levels decoded as present: %+v", decoded[:4])
	}
}

func TestDecodeLevelSeedsTruncated(t *testing.T) {
	// given a stream cut right after a present flag
	image := []byte{0x01}
	// when it is decoded
	_, err := persist.DecodeLevelSeeds(image, persist.GameMode{})
	// then the missing seed is reported
	if err == nil {
		t.Fatal("truncated stream decoded without error")
	}
}
