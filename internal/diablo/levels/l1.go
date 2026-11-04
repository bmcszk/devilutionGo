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

// GenerateL1 runs the cathedral generator: CreateL5Dungeon's GenerateLevel
// (levels/drlg_l1.cpp:1212-1312) for a cathedral level.
//
// The caller seeds d.Rnd; GenerateLevel only re-seeds when it finds a
// level seed recorded by an earlier run of the same level
// (levels/drlg_l1.cpp:1214-1215, `LevelSeeds[currlevel]`), which d.LevelSeed
// models. The fixture probe resets that slot per variant, so a fresh Dungeon
// enters the generator unseeded and keeps the caller's RNG state.
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

	d.fixDirtTiles()
	d.dumpStage("05-fixdirt")

	d.fixCornerTiles()
	d.dumpStage("06-fixcorner")

	d.substitution()
	d.dumpStage("07-subst")

	d.applyShadowsPatterns()
	d.dumpStage("08-shadows")

	d.placeLamps()
	d.dumpStage("09-lamps")

	d.fillFloor()
	d.dumpStage("10-fillfloor")

	d.copyToPrevTiles()
	d.checkQuests()
	d.dumpStage("11-quests")

	return nil
}

// checkQuests is DRLG_CheckQuests' call site on the cathedral path
// (levels/drlg_l1.cpp:1294). On this level type it only draws the banner set
// piece into pdungeon, which the quest hook owns (unit C); the call site is
// here so the pass order stays 1:1.
func (d *Dungeon) checkQuests() {
	if d.CheckQuests != nil {
		d.CheckQuests(d.SetPiece.Position)
	}
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
