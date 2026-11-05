package levels

// This file drives the hell generator: CreateL4Dungeon's GenerateLevel
// (levels/drlg_l4.cpp:1148-1238) for DTYPE_HELL levels 13-16, layout through
// decoration. Pass3 (levels/drlg_l4.cpp:1238-1284) lives in piece.go: ExpandPieces
// writes dPiece/dSpecial through the level's mega-tile table, which the caller
// supplies as data.

const (
	// hellRock is the tile InitDungeonFlags fills the hell layout with
	// (levels/drlg_l4.cpp:191, memset(dungeon, 30, sizeof(dungeon))). Hell has no
	// Tile enum upstream, so the passes spell their ids inline.
	hellRock Tile = 30
	// hellFloor is the open floor tile every hell pass tests against
	// (levels/drlg_l4.cpp:1166, FloodTransparencyValues(6)).
	hellFloor Tile = 6
	// hellHWall is the horizontal wall tile a corridor is filled with
	// (levels/drlg_l4.cpp:387).
	hellHWall Tile = 2
	// hellVWall is the vertical wall tile (levels/drlg_l4.cpp:428).
	hellVWall Tile = 1
	// hellEdgeWall is the lowest wall id FixCornerTiles redraws
	// (levels/drlg_l4.cpp:1068).
	hellEdgeWall Tile = 18
	// hellCornerOffset is the +98 a corner wall tile is redrawn with
	// (levels/drlg_l4.cpp:1070).
	hellCornerOffset Tile = 98
	// hellDWall and hellDWallB are the two double-wide wall tiles FixTransparency
	// halves copy their region onto (levels/drlg_l4.cpp:1029, 1035).
	hellDWall  Tile = 18
	hellDWallB Tile = 19
	// hellDoorH and hellDoorV are the horizontal and vertical doorway tiles whose
	// transparency spills onto the neighbouring tile
	// (levels/drlg_l4.cpp:1050, 1054).
	hellDoorH Tile = 57
	hellDoorV Tile = 53
	// hellTransOrigin is where the doubled transparency window starts
	// (levels/drlg_l4.cpp:1023).
	hellTransOrigin = 16
	// hellFirstRoomSide is the level 16 first room size (levels/drlg_l4.cpp:277).
	hellFirstRoomSide = 14
	// hellQuadSide is the 14x14 corner ProtectQuads protects
	// (levels/drlg_l4.cpp:957).
	hellQuadSide = 14
	// hellMinArea is the carved quadrant area an attempt must reach
	// (levels/drlg_l4.cpp:1155).
	hellMinArea = 692
	// hellPlacementTries is PlaceMiniSet's default scan length
	// (levels/gendung.h:87). All four hell calls take the default, which passes
	// no tries argument (levels/drlg_l4.cpp:1106, 1119, 1129, 1137).
	hellPlacementTries = 199
	// hellThemeMin, hellThemeMax and hellThemeFreq are the hell
	// DRLG_PlaceThemeRooms arguments (levels/drlg_l4.cpp:1194).
	hellThemeMin  = 7
	hellThemeMax  = 10
	hellThemeFreq = 8
	// warlordLevel, hellPentagramLevel and hellQuadLevel are the levels the
	// generator branches on (levels/drlg_l4.cpp:1156, 1176, 1179, 1194).
	warlordLevel       = 13
	hellPentagramLevel = 15
	hellQuadLevel      = 16
	// hellWarlordSetPiece and hellWarlordQuestSetPiece are the two warlord rooms:
	// the one the layout reserves and the one the quest draws over it
	// (levels/drlg_l4.cpp:175, levels/drlg_quests.cpp:37).
	hellWarlordSetPiece      = "levels/l4data/warlord.dun"
	hellWarlordQuestSetPiece = "levels/l4data/warlord2.dun"
	// hellDiabloQuad1 is the shared north-west Diablo quadrant
	// (levels/drlg_l4.cpp:969).
	hellDiabloQuad1 = "levels/l4data/diab1.dun"
)

// GenerateL4 runs the hell generator for the level d was created with.
// The caller seeds d.Rnd; GenerateLevel only re-seeds when it finds a level seed
// recorded by an earlier run of the same level (levels/drlg_l4.cpp:1150-1151),
// which d.LevelSeed models.
func GenerateL4(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	for {
		ok, err := d.rollHellLevel(opts)
		if err != nil {
			return err
		}

		if ok {
			break
		}
	}

	return d.finishHellLevel(opts)
}

// rollHellLevel is one GenerateLevel attempt: layout, tile patterns, walls,
// transparency, set pieces and stairs. It reports false when the hell gate or the
// stairs do not fit, which rerolls the whole attempt
// (levels/drlg_l4.cpp:1153-1185).
func (d *Dungeon) rollHellLevel(opts Options) (bool, error) {
	d.InitTrans()
	d.rollHellLayout(opts)

	d.prepareHellInnerBorders()
	d.mirrorHellLayout()

	d.makeHellDmt()
	d.fixHellTilesPatterns()

	if d.CurrLevel == hellQuadLevel {
		d.protectQuads()
	}

	if opts.Quests.Warlord {
		d.protectHellSetPieceRoom()
	}

	d.dumpStage("fixpat")
	d.addHellWalls()
	d.FloodTransparencyValues(hellFloor)
	d.fixHellTransparency()
	d.dumpStage("trans")

	if err := d.initHellSetPiece(opts); err != nil {
		return false, err
	}

	if d.CurrLevel == hellQuadLevel {
		if err := d.loadPreGateDiabloQuads(opts); err != nil {
			return false, err
		}
	}

	return d.placeHellStairs(opts), nil
}

// rollHellLayout rerolls the quadrant layout until it is large enough, recording
// the RNG state the next attempt would resume from
// (levels/drlg_l4.cpp:1155-1162).
func (d *Dungeon) rollHellLayout(opts Options) {
	for {
		d.LevelSeed = d.Rnd.State()
		d.initHellDungeonFlags()
		d.firstHellRoom(opts)
		d.closeOuterBorders()

		if d.hellArea() >= hellMinArea {
			return
		}
	}
}

// finishHellLevel runs the passes GenerateLevel executes once an attempt
// succeeded, in upstream order (levels/drlg_l4.cpp:1187-1238).
func (d *Dungeon) finishHellLevel(opts Options) error {
	d.dumpStage("stairsok")
	d.generalFix()

	if d.CurrLevel != hellQuadLevel {
		d.placeHellThemeRooms()
	}

	d.dumpStage("theme")
	d.applyHellShadowPatterns()
	d.fixHellCornerTiles()
	d.hellSubstitution()
	d.dumpStage("shadows")

	d.copyToPrevTiles()

	if err := d.drawHellQuests(opts); err != nil {
		return err
	}

	if d.CurrLevel == hellPentagramLevel {
		d.closeHellGate(opts)
	}

	if d.CurrLevel == hellQuadLevel {
		return d.loadPostGateDiabloQuads(opts)
	}

	return nil
}

// closeHellGate redraws the pentagram over the portal while the gate is closed
// (levels/drlg_l4.cpp:1209-1212).
//
// The loop that follows upstream records the portal position for the quest and
// marks the 5x5 corner populated in dFlags; both are quest and dPiece state, not
// generator state (levels/drlg_l4.cpp:1214-1222).
func (d *Dungeon) closeHellGate(opts Options) {
	if opts.Quests.DiabloGateOpen {
		return
	}

	l4Penta.Place(d, d.DiabloPosition, false)
}

// protectHellSetPieceRoom protects the interior of the warlord set piece room so
// no later pass overwrites it (levels/drlg_l4.cpp:1160-1167).
func (d *Dungeon) protectHellSetPieceRoom() {
	room := d.SetPieceRoom

	for i := room.Position.X; i < room.Position.X+room.Size.Width-1; i++ {
		for j := room.Position.Y; j < room.Position.Y+room.Size.Height-1; j++ {
			d.Protected.Set(i, j, true)
		}
	}
}
