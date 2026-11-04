package levels

import "fmt"

// Quests is the quest availability the cathedral generator consults
// (drlg_l1.cpp:386-394, 1116-1145). Each field is a Quest::IsAvailable()
// result for the level being generated; the caller decides it, the generator
// only reads the snapshot.
type Quests struct {
	// PWATER is Quests[Q_PWATER].IsAvailable().
	PWATER bool
	// LTBanner is Quests[Q_LTBANNER].IsAvailable().
	LTBanner bool
	// Butcher is Quests[Q_BUTCHER].IsAvailable().
	Butcher bool
	// Skelking is Quests[Q_SKELKING].IsAvailable() (and not multiplayer).
	Skelking bool
	// CornerStone is CornerStone.isAvailable() (items.cpp:4865-4868).
	CornerStone bool
	// Blind is Quests[Q_BLIND].IsAvailable(): the catacombs level 7 set piece
	// plus the outer-wall patch (levels/drlg_l2.cpp:1638-1656, 2708-2716).
	Blind bool
	// Blood is Quests[Q_BLOOD].IsAvailable(): the catacombs level 5 set piece
	// plus the blood room (levels/drlg_l2.cpp:1638-1656, 2708-2716).
	Blood bool
	// Schamb is Quests[Q_SCHAMB].IsAvailable(): the catacombs level 6 bone
	// chamber (levels/drlg_l2.cpp:1638-1656, 2708-2716).
	Schamb bool
}

// Options carries the inputs CreateL5Dungeon reads besides the dungeon state.
type Options struct {
	// Entry is the level entry (lvl_entry).
	Entry Entry
	// OriginalCathedral is MyPlayer->pOriginalCathedral.
	OriginalCathedral bool
	// Quests is the quest availability snapshot.
	Quests Quests
	// SetPieces maps a .dun resource path to its bytes, replacing the
	// upstream LoadFileInMem calls.
	SetPieces map[string][]byte
}

// GenerateL1 runs the L1 generator: CreateL5Dungeon's GenerateLevel
// (levels/drlg_l1.cpp:1212-1294) for a cathedral or crypt level. The level type
// selects the crypt branches, exactly as upstream.
//
// The caller seeds d.Rnd; GenerateLevel only re-seeds when it finds a
// level seed recorded by an earlier run of the same level
// (levels/drlg_l1.cpp:1214-1215, `LevelSeeds[currlevel]`), which d.LevelSeed
// models. The fixture probe resets that slot per variant, so a fresh Dungeon
// enters the generator unseeded and keeps the caller's RNG state.
//
//nolint:revive // 1:1 port of GenerateLevel's pass sequence (levels/drlg_l1.cpp:1212-1294)
func GenerateL1(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	minArea := levelMinArea(d.CurrLevel)

	for attempt := 1; ; attempt++ {
		d.InitTrans()

		d.retryLayout(minArea)
		d.dumpStage(fmt.Sprintf("01-firstroom-a%d", attempt))

		d.initDungeonFlags()
		d.makeDmt()
		if err := d.fillChambers(opts); err != nil {
			return err
		}

		d.fixTilesPatterns()
		d.addWall()
		d.FloodTransparencyValues(catFloor)
		d.dumpStage(fmt.Sprintf("02-predstairs-a%d", attempt))

		if d.placeStairs(opts) {
			break
		}
	}

	d.carryEntranceStairsTrans()
	d.dumpStage("03-stairs")

	d.fixTransparency()
	d.dumpStage("04-fixtrans")

	if d.LevelType == Crypt {
		d.fixCryptDirtTiles()
	} else {
		d.fixDirtTiles()
	}
	d.dumpStage("05-fixdirt")

	d.fixCornerTiles()
	d.dumpStage("06-fixcorner")

	if d.LevelType == Crypt {
		d.cryptSubstitution()
	} else {
		d.decorateCathedral()
	}

	d.copyToPrevTiles()
	d.checkQuests(opts.Quests)
	d.dumpStage("11-quests")

	return nil
}

// decorateCathedral runs the four passes only the cathedral branch of
// GenerateLevel executes (levels/drlg_l1.cpp:1274-1289).
func (d *Dungeon) decorateCathedral() {
	d.substitution()
	d.dumpStage("07-subst")

	d.applyShadowsPatterns()
	d.dumpStage("08-shadows")

	d.placeLamps()
	d.dumpStage("09-lamps")

	d.fillFloor()
	d.dumpStage("10-fillfloor")
}

// checkQuests is DRLG_CheckQuests' call site on the cathedral path
// (levels/drlg_l1.cpp:1294, levels/drlg_quests.cpp:89-118). Of the quest
// branches only DrawButcher is ported here: it is the one branch whose effect
// falls inside the state this generator owns (it widens the Butcher's room
// transparency sector, levels/drlg_quests.cpp:19-23). The remaining branches
// (DrawLTBanner, DrawSChamber, DrawBlood, DrawWarLord, DrawBlind, DrawSkelKing)
// write pdungeon or a quest position and stay with the quest unit; the hook
// below is their documented call site.
func (d *Dungeon) checkQuests(quests Quests) {
	if quests.Butcher {
		d.drawButcher()
	}

	if d.CheckQuests != nil {
		d.CheckQuests(d.SetPiece.Position)
	}
}

// drawButcher opens the transparency sector the Butcher's room occupies
// (levels/drlg_quests.cpp:19-23).
func (d *Dungeon) drawButcher() {
	d.RectTrans(Rectangle{
		Position: d.SetPiece.Position.MegaToWorld().Add(Point{X: 3, Y: 3}),
		Size:     Size{Width: 7, Height: 7},
	})
}

// levelMinArea is GenerateLevel's layout area threshold per level
// (levels/drlg_l1.cpp:1217-1227).
func levelMinArea(level int) int {
	switch level {
	case 1:
		return 533
	case 2:
		return 693
	default:
		return 761
	}
}

// retryLayout rerolls the room layout until it covers minArea mega tiles,
// recording the RNG state the next attempt would resume from
// (levels/drlg_l1.cpp:1235-1238).
func (d *Dungeon) retryLayout(minArea int) {
	for {
		d.LevelSeed = d.Rnd.State()

		d.firstRoom()
		if d.findArea() >= minArea {
			return
		}
	}
}

// carryEntranceStairsTrans copies the entrance-stairs transparency down one
// world row so the staircase graphic covers both halves of the tile
// (levels/drlg_l1.cpp:1239-1249).
func (d *Dungeon) carryEntranceStairsTrans() {
	for j := range DMAXY {
		for i := range DMAXX {
			if d.Tiles[i][j] != catEntranceStairs {
				continue
			}

			xx := 2*i + 16
			yy := 2*j + 16

			d.CopyTrans(xx, yy+1, xx, yy)
			d.CopyTrans(xx+1, yy+1, xx+1, yy)
		}
	}
}

// dumpStage reports the generator state after one named pass, mirroring the
// DumpStage calls the offline oracle build inserts (levels/drlg_l1.cpp:1240,
// 1247, 1256, ...). A nil hook is a no-op, so production runs pay one branch.
func (d *Dungeon) dumpStage(name string) {
	if d.DumpStage != nil {
		d.DumpStage(name)
	}
}
