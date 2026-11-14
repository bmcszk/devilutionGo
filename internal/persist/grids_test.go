package persist_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestDLightGridPins the dLight byte grid block (loadsave.cpp:2869-2873).
func TestDLightGrid(t *testing.T) {
	// given a dLight grid with distinct cell values
	var lights [persist.DungeonWidth * persist.DungeonHeight]byte
	fillGrid(&lights)

	// when the grid is written
	image := persist.EncodeDLightGrid(lights)

	// then it is 12544 bytes in row-outer order
	if len(image) != persist.DungeonWidth*persist.DungeonHeight {
		t.Fatalf("dLight grid = %d bytes", len(image))
	}

	assertGridOffset(t, image, (2+3)%251)
}

// TestDFlagsGridPins the dFlags grid block: cells are masked to SavedFlags
// (loadsave.cpp:1874-1878, DungeonFlag::SavedFlags).
func TestDFlagsGrid(t *testing.T) {
	// given a dFlags grid mixing saved and unsaved bits
	var flags [persist.DungeonWidth * persist.DungeonHeight]byte
	for i := range flags {
		flags[i] = 0xFF // every bit set
	}

	flags[0] = 0xC8 // exactly SavedFlags
	flags[1] = 0xCF // LoadedFlags: masked to 0xC8
	flags[2] = 0x37 // no saved bits

	// when the grid is written
	image := persist.EncodeDFlagsGrid(flags)

	// then every cell is masked to SavedFlags
	if image[0] != 0xC8 || image[1] != 0xC8 || image[2] != 0 {
		t.Fatalf("dFlags mask = %02x %02x %02x, want c8 c8 00", image[0], image[1], image[2])
	}

	if len(image) != persist.DungeonWidth*persist.DungeonHeight {
		t.Fatalf("dFlags grid = %d bytes", len(image))
	}
}

// TestDPlayerGridPins the dPlayer signed byte grid (loadsave.cpp:1880-1884).
func TestDPlayerGrid(t *testing.T) {
	// given a dPlayer grid holding a negative cell
	var players [persist.DungeonWidth * persist.DungeonHeight]int8
	fillGridSigned(&players)
	players[4] = -1

	// when the grid is written
	image := persist.EncodeDPlayerGrid(players)

	// then the signed value round-trips through the byte
	if image[4] != 0xFF {
		t.Fatalf("dPlayer[4] = %d, want 0xff (-1)", image[4])
	}

	assertGridOffset(t, image, (2+3)%251)
}

// TestDroppedItemLocationsGridPins the dItem location block: the values are
// the save-relative item indexes (loadsave.cpp:1858-1866, written at
// :2884 via SaveDroppedItemLocations).
func TestDroppedItemLocationsGrid(t *testing.T) {
	// given a dItem grid
	var items [persist.DungeonWidth * persist.DungeonHeight]byte
	items[0] = 1
	items[5*persist.DungeonWidth+6] = 2

	// when the locations block is written
	image := persist.EncodeDroppedItemLocations(items)

	// then it is the raw cell sequence
	if len(image) != persist.DungeonWidth*persist.DungeonHeight || image[0] != 1 ||
		image[5*persist.DungeonWidth+6] != 2 {
		t.Fatal("dropped item locations block mismatch")
	}
}

// monsterGridWantTail is the tail of the BE-int32 dMonster grid: the last
// cell holds 7 (loadsave.cpp:1886-1890).
func TestDMonsterGrid(t *testing.T) {
	// given a dMonster grid
	var monsters [persist.DungeonWidth * persist.DungeonHeight]int32
	for i := range monsters {
		monsters[i] = int32(i % 9)
	}

	// when the grid is written
	image := persist.EncodeDMonsterGrid(monsters)

	// then it is 4 bytes per cell big-endian
	if len(image) != 4*persist.DungeonWidth*persist.DungeonHeight {
		t.Fatalf("dMonster grid = %d bytes", len(image))
	}

	last := 4 * (persist.DungeonWidth*persist.DungeonHeight - 1)
	if image[last] != 0 || image[last+3] != 6 { // 12543 %% 9 = 6
		t.Fatalf("last dMonster cell = %x, want 00000007", image[last:last+4])
	}
}

// TestDCorpseAndDObjectGrids pin the signed byte dungeon grids
// (loadsave.cpp:1892-1900).
func TestDCorpseAndDObjectGrids(t *testing.T) {
	// given corpse and object grids
	var corpses [persist.DungeonWidth * persist.DungeonHeight]int8
	var objects [persist.DungeonWidth * persist.DungeonHeight]int8
	corpses[1] = -2
	objects[2] = 3

	// when the grids are written
	corpseImage := persist.EncodeDCorpseGrid(corpses)
	objectImage := persist.EncodeDObjectGrid(objects)

	// then the signed values survive
	if corpseImage[1] != 0xFE || objectImage[2] != 3 {
		t.Fatalf("corpse/object = %d/%d, want fe/03", corpseImage[1], objectImage[2])
	}
}

// TestDungeonGridQuirkDoubleDLight pins the upstream BUGFIX line: the dungeon
// tail writes dLight a SECOND time after dObject (loadsave.cpp:1901-1906,
// "BUGFIX: dLight got saved already").
func TestDungeonGridQuirkDoubleDLight(t *testing.T) {
	// given distinct dLight and dPreLight grids
	var lights [persist.DungeonWidth * persist.DungeonHeight]byte
	var preLights [persist.DungeonWidth * persist.DungeonHeight]byte
	for i := range lights {
		lights[i] = 7
		preLights[i] = 9
	}

	grids := persist.DungeonGrids{
		Monsters:  [persist.DungeonWidth * persist.DungeonHeight]int32{},
		Corpses:   [persist.DungeonWidth * persist.DungeonHeight]int8{},
		Objects:   [persist.DungeonWidth * persist.DungeonHeight]int8{},
		Lights:    lights,
		PreLights: preLights,
	}

	// when the dungeon tail is composed
	image := persist.EncodeDungeonGrids(grids)

	// then the layout is dMonster + dCorpse + dObject + the BUGFIX dLight
	// copy + dPreLight + AutomapView
	monsterBytes := 4 * persist.DungeonWidth * persist.DungeonHeight
	cellBytes := persist.DungeonWidth * persist.DungeonHeight
	quirkCopy := image[monsterBytes+2*cellBytes : monsterBytes+3*cellBytes]
	preLightBlock := image[monsterBytes+3*cellBytes : monsterBytes+4*cellBytes]
	automapBlock := image[monsterBytes+4*cellBytes:]

	if !bytes.Equal(quirkCopy, lights[:]) {
		t.Fatal("double-dLight quirk: the dungeon tail must repeat dLight")
	}

	if !bytes.Equal(preLightBlock, preLights[:]) {
		t.Fatal("dPreLight block after the BUGFIX dLight copy")
	}

	if len(automapBlock) != persist.AutomapWidth*persist.AutomapHeight {
		t.Fatalf("automap block = %d bytes", len(automapBlock))
	}

	wantLen := monsterBytes + 4*cellBytes + persist.AutomapWidth*persist.AutomapHeight
	if len(image) != wantLen {
		t.Fatalf("dungeon grids = %d bytes, want %d", len(image), wantLen)
	}
}

// gridCells mirrors the MAXDUNX×MAXDUNY loop order of the grid writers: row
// j outer, column i inner (loadsave.cpp:1869-1881, :1942-1985).
func fillGrid(cells *[persist.DungeonWidth * persist.DungeonHeight]byte) {
	for j := range persist.DungeonHeight {
		for i := range persist.DungeonWidth {
			cells[j*persist.DungeonWidth+i] = byte((i + j) % 251)
		}
	}
}

// assertGridPins the column-major-visual (row-outer) write order: cell (2,3)
// lands at offset 3*112+2.
func assertGridOffset(t *testing.T, image []byte, want byte) {
	t.Helper()

	const offset = 3*persist.DungeonWidth + 2
	if image[offset] != want {
		t.Fatalf("grid[%d] = %d, want %d (row-outer order)", offset, image[offset], want)
	}
}

// fillGridSigned fills a signed grid with distinct values.
func fillGridSigned(cells *[persist.DungeonWidth * persist.DungeonHeight]int8) {
	for j := range persist.DungeonHeight {
		for i := range persist.DungeonWidth {
			cells[j*persist.DungeonWidth+i] = int8((i + j) % 100)
		}
	}
}
