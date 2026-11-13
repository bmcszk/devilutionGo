package persist

// DungeonGrids carries the six grid arrays the non-town SaveGameData tail
// writes (Source/loadsave.cpp:1885-1908): the BE-int32 dMonster grid, the
// signed byte grids dCorpse and dObject, dLight, dPreLight, AutomapView. The
// duplicated dLight copy of loadsave.cpp:1901-1906 is reproduced by
// EncodeDungeonGrids.
type DungeonGrids struct {
	// Monsters is dMonster (BE int32 per cell).
	Monsters [dungeonGridCells]int32
	// Corpses is dCorpse.
	Corpses [dungeonGridCells]int8
	// Objects is dObject.
	Objects [dungeonGridCells]int8
	// Lights is dLight.
	Lights [dungeonGridCells]uint8
	// PreLights is dPreLight.
	PreLights [dungeonGridCells]uint8
	// Automap is AutomapView (DMAXX x DMAXY).
	Automap [automapGridCells]uint8
}

// EncodeDLightGrid writes the dLight byte grid in row-outer order
// (loadsave.cpp:2869-2873).
func EncodeDLightGrid(cells [dungeonGridCells]uint8) []byte {
	grid := make([]byte, dungeonGridCells)
	copy(grid, cells[:])

	return grid
}

// EncodeDFlagsGrid writes the dFlags byte grid with every cell masked to
// DungeonFlag::SavedFlags (loadsave.cpp:1874-1878).
func EncodeDFlagsGrid(cells [dungeonGridCells]uint8) []byte {
	grid := make([]byte, dungeonGridCells)

	for i, cell := range cells {
		grid[i] = SavedFlags(cell)
	}

	return grid
}

// EncodeDPlayerGrid writes the dPlayer signed byte grid
// (loadsave.cpp:1880-1884).
func EncodeDPlayerGrid(cells [dungeonGridCells]int8) []byte {
	grid := make([]byte, dungeonGridCells)

	for i, cell := range cells {
		grid[i] = uint8(cell)
	}

	return grid
}

// EncodeDroppedItemLocations writes the dItem location block: the
// save-relative item index per cell (loadsave.cpp:1858-1866).
func EncodeDroppedItemLocations(cells [dungeonGridCells]uint8) []byte {
	grid := make([]byte, dungeonGridCells)
	copy(grid, cells[:])

	return grid
}

// EncodeDMonsterGrid writes the dMonster grid as big-endian int32 cells
// (loadsave.cpp:1886-1890).
func EncodeDMonsterGrid(cells [dungeonGridCells]int32) []byte {
	var out recordWriter

	for _, cell := range cells {
		out.be32(uint32(cell))
	}

	return out.buf
}

// EncodeDCorpseGrid writes the dCorpse signed byte grid
// (loadsave.cpp:1892-1896).
func EncodeDCorpseGrid(cells [dungeonGridCells]int8) []byte {
	grid := make([]byte, dungeonGridCells)

	for i, cell := range cells {
		grid[i] = uint8(cell)
	}

	return grid
}

// EncodeDObjectGrid writes the dObject signed byte grid
// (loadsave.cpp:1897-1900).
func EncodeDObjectGrid(cells [dungeonGridCells]int8) []byte {
	grid := make([]byte, dungeonGridCells)

	for i, cell := range cells {
		grid[i] = uint8(cell)
	}

	return grid
}

// EncodeDungeonGrids composes the non-town grid tail in the upstream write
// order: dMonster, dCorpse, dObject, the BUGFIX dLight copy, dPreLight and
// AutomapView (loadsave.cpp:1885-1916). The dLight copy here is the second
// write of the array in one stream — the common section already wrote it at
// loadsave.cpp:2869-2873 — and is load-bearing parity, not a copy-paste bug
// on this side.
func EncodeDungeonGrids(grids DungeonGrids) []byte {
	out := make([]byte, 0,
		4*dungeonGridCells+3*dungeonGridCells+automapGridCells)

	out = append(out, EncodeDMonsterGrid(grids.Monsters)...)
	out = append(out, EncodeDCorpseGrid(grids.Corpses)...)
	out = append(out, EncodeDObjectGrid(grids.Objects)...)
	out = append(out, EncodeDLightGrid(grids.Lights)...) // BUGFIX duplication
	out = append(out, grids.PreLights[:]...)
	out = append(out, grids.Automap[:]...)

	return out
}
