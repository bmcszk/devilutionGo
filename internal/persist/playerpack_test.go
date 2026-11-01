package persist_test

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestPlayerPackSize(t *testing.T) {
	// when
	size := persist.PlayerPackSize
	// then the pack.h:33-88 layout sums to 1266 bytes
	if size != 1266 {
		t.Fatalf("PlayerPackSize = %d, want 1266", size)
	}
}

func TestPlayerPackGoldenImage(t *testing.T) {
	// given the fixture state and its independently generated image
	pack := goldenPlayerPack()
	want := loadPlayerPackHex(t)
	// when
	encoded := pack.Encode()
	// then the bytes match field for field
	if got := hex.EncodeToString(encoded); got != want {
		t.Fatalf("Encode = %s, want %s", got, want)
	}
}

func TestPlayerPackFieldOffsets(t *testing.T) {
	// given the golden image
	image, err := hex.DecodeString(loadPlayerPackHex(t))
	if err != nil {
		t.Fatalf("decode hex: %v", err)
	}
	// when the fixed offsets of pack.h are read back
	cases := []struct {
		name   string
		offset int
		bytes  string
	}{
		{"pSplLvl", 79, "00010203040506"},
		{"pMemSpells", 116, "0807060504030201"},
		{"InvGrid", 1017, strings.Repeat("00", 8)},
		{"NumInv", 1057, "01"},
		{"TownWarps", 1210, "01"},
		{"SplLvl2", 1222, "0102030405060708090a"},
		{"wReserved8", 1232, "feff"},
	}
	// then every field sits where pack.h puts it
	for _, tc := range cases {
		got := hex.EncodeToString(image[tc.offset : tc.offset+len(tc.bytes)/2])
		if got != tc.bytes {
			t.Fatalf("%s at %d = %s, want %s", tc.name, tc.offset, got, tc.bytes)
		}
	}
}

func TestPlayerPackRoundTrip(t *testing.T) {
	// given the golden record
	pack := goldenPlayerPack()
	// when
	decoded, err := persist.DecodePlayerPack(pack.Encode())
	// then every field survives
	if err != nil {
		t.Fatalf("DecodePlayerPack: %v", err)
	}

	if decoded != pack {
		t.Fatalf("round trip = %+v, want %+v", decoded, pack)
	}
}

func TestDecodePlayerPackShortBuffer(t *testing.T) {
	// given a truncated hero record
	buf := make([]byte, persist.PlayerPackSize-1)
	// when
	_, err := persist.DecodePlayerPack(buf)
	// then the decode fails instead of over-reading
	if err == nil {
		t.Fatal("DecodePlayerPack(short) = nil error, want error")
	}
}

// goldenPlayerPack is the state pinned by testdata/player_pack_golden.hex.
func goldenPlayerPack() persist.PlayerPack {
	pack := persist.PlayerPack{
		LowDateTime:  1,
		HighDateTime: 2,
		DestAction:   -1,
		DestParam1:   3,
		DestParam2:   4,
		PlrLevel:     5,
		Px:           10,
		Py:           20,
		Class:        0,
		BaseStr:      30,
		BaseMag:      10,
		BaseDex:      25,
		BaseVit:      20,
		Level:        12,
		StatPts:      5,
		Exp:          1000,
		Gold:         500,
		HPBase:       96,
		MaxHP:        128,
		ManaBase:     32,
		MaxMana:      48,
		MemSpells:    0x0102030405060708,
		NumInv:       1,
		TownWarps:    1,
		DungMsgs:     2,
		ManaShield:   1,
		IsHellfire:   1,
		Reflections:  3,
		Reserved8:    -2,
	}

	copy(pack.Name[:], "Conan")
	for i := range pack.SplLvl {
		pack.SplLvl[i] = uint8(i % 10)
	}

	for i := range pack.SplLvl2 {
		pack.SplLvl2[i] = uint8(i + 1)
	}

	goldenPlayerItems(&pack)
	pack.DiabloKillLevel = 7
	pack.Difficulty = 8
	pack.DamAcFlags = 9

	return pack
}

// goldenPlayerItems fills the three item blocks of the golden record.
func goldenPlayerItems(pack *persist.PlayerPack) {
	for i := range pack.InvBody {
		pack.InvBody[i].Idx = persist.PackedItemNone
	}

	for i := range pack.InvList {
		pack.InvList[i].Idx = persist.PackedItemNone
	}

	for i := range pack.SpdList {
		pack.SpdList[i].Idx = persist.PackedItemNone
	}

	pack.InvBody[0] = persist.ItemPack{
		Seed: 1, CreateInfo: 2, Idx: 3, BId: 4, BDur: 5, BMDur: 6, BCh: 7, BMCh: 8, Value: 9, Buff: 10,
	}
	pack.InvList[39] = persist.ItemPack{
		Seed: 11, CreateInfo: 12, Idx: 13, BId: 14, BDur: 15, BMDur: 16, BCh: 17, BMCh: 18, Value: 19, Buff: 20,
	}
}

// loadPlayerPackHex reads the committed record image.
func loadPlayerPackHex(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("testdata/player_pack_golden.hex")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		return strings.TrimSpace(line)
	}

	t.Fatal("golden has no image line")

	return ""
}
