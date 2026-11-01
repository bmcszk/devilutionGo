package persist_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestTileGridsSize(t *testing.T) {
	// when the grid block is encoded (loadsave.cpp:1986-2008)
	image := persist.TileGrids{}.Encode()
	// then it is 112*112*7 + 40*40 bytes: dMonster is BE int32, the rest bytes
	if len(image) != 89_408 {
		t.Fatalf("grid block len = %d, want 89408", len(image))
	}

	if len(image) != persist.TileGridsSize {
		t.Fatalf("TileGridsSize = %d", persist.TileGridsSize)
	}
}

func TestTileGridsFieldOffsets(t *testing.T) {
	// given grids with a value in the first and last cell of each block
	var grids persist.TileGrids
	grids.Monsters[0] = 7
	grids.Monsters[persist.DungeonWidth*persist.DungeonHeight-1] = -3
	grids.Objects[0] = -1
	grids.Lights[0] = 9
	grids.PreLights[0] = 8
	grids.Automap[0] = 5
	// when the block is encoded
	image := grids.Encode()
	// then each block starts where the previous one ends, monsters big-endian
	if got := int32(binary.BigEndian.Uint32(image[0:4])); got != 7 {
		t.Fatalf("monster[0] = %d, want 7", got)
	}

	lastMonster := 112*112*4 - 4
	if got := int32(binary.BigEndian.Uint32(image[lastMonster : lastMonster+4])); got != -3 {
		t.Fatalf("last monster = %d, want -3", got)
	}

	cases := []struct {
		name   string
		offset int
		want   uint8
	}{
		{"dObject[0]", 112 * 112 * 4, 0xff},
		{"dLight[0]", 112 * 112 * 5, 9},
		{"dPreLight[0]", 112 * 112 * 6, 8},
		{"AutomapView[0]", 112 * 112 * 7, 5},
	}
	for _, tc := range cases {
		if image[tc.offset] != tc.want {
			t.Fatalf("%s at %d = %d, want %d", tc.name, tc.offset, image[tc.offset], tc.want)
		}
	}
}

func TestTileGridsRoundTrip(t *testing.T) {
	// given grids with negative monster and object indices
	var grids persist.TileGrids
	grids.Monsters[1] = -2
	grids.Monsters[2] = 26000
	grids.Objects[3] = -1
	grids.Lights[4] = 255
	grids.PreLights[5] = 1
	grids.Automap[6] = 3
	// when the block is decoded back
	decoded, err := persist.DecodeTileGrids(grids.Encode())
	// then every value round-trips
	if err != nil {
		t.Fatalf("DecodeTileGrids: %v", err)
	}

	if decoded != grids {
		t.Fatal("grids changed")
	}
}

func TestDecodeTileGridsShort(t *testing.T) {
	// when a truncated block is decoded
	_, err := persist.DecodeTileGrids(make([]byte, persist.TileGridsSize-1))
	// then the truncation is reported
	if err == nil {
		t.Fatal("truncated block decoded without error")
	}
}

func TestSavedFlags(t *testing.T) {
	// when a cell carries missile/visible/dead-player bits plus saved flags
	cell := byte(0xCF)
	// then only Populated|Lit|Explored survive (dun_tile_data.hpp:45)
	if got := persist.SavedFlags(cell); got != persist.DungeonFlagSaved {
		t.Fatalf("SavedFlags(0xCF) = %#x, want %#x", got, persist.DungeonFlagSaved)
	}

	if got := persist.SavedFlags(0x30); got != 0 {
		t.Fatalf("SavedFlags(0x30) = %#x, want 0", got)
	}
}

func TestEncodeFlagAndCorpseGrids(t *testing.T) {
	// given a dFlags cell with a missile bit and a corpse cell
	var flags [persist.DungeonWidth * persist.DungeonHeight]uint8
	flags[0] = 0xCF

	var corpses [persist.DungeonWidth * persist.DungeonHeight]int8
	corpses[0] = -5
	// when the blocks are encoded (loadsave.cpp:1948-1953, :1978-1984)
	flagImage := persist.EncodeFlagGrid(flags)
	corpseImage := persist.EncodeCorpseGrid(corpses)
	// then the flag grid is masked and both are one byte per cell
	if len(flagImage) != persist.DungeonWidth*persist.DungeonHeight ||
		len(corpseImage) != persist.CorpseGridSize {
		t.Fatalf("lens = %d / %d", len(flagImage), len(corpseImage))
	}

	if flagImage[0] != persist.DungeonFlagSaved || corpseImage[0] != 0xfb {
		t.Fatalf("first cells = %#x / %#x", flagImage[0], corpseImage[0])
	}
}
