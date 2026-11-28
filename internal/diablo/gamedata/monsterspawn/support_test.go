package monsterspawn_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monsterspawn"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

const (
	// tapesPath is the committed scripted-tape fixture of PlaceGroup.
	tapesPath = "testdata/placegroup_tapes.tsv"
	// tapeColumns is the column count of one tape row.
	tapeColumns = 17
)

// tapeCase is one scripted PlaceGroup case, parsed from the committed tape.
type tapeCase struct {
	name  string
	world string

	num, active, total int
	leader             *monsterspawn.Leader
	leashed            bool
	base               monsterspawn.Stats

	script      []int32
	wantTiles   []wantTile
	wantRolls   int
	wantRewinds int
	wantPack    int
}

// wantTile is one hand-derived placement expectation.
type wantTile struct {
	x, y         int
	direction    int
	startFrame   int
	maxHitPoints int
	hitPoints    int
	intelligence int
	leashed      bool
}

// flatWorld is a test dungeon: a placeable set plus fixed transparency regions.
type flatWorld struct {
	// placeable holds the accepted tiles when it is non-nil; a nil set accepts
	// the whole 64x64 test dungeon.
	placeable map[[2]int]bool
	// transAt maps a tile to its dTransVal region, defaulting to 0.
	transAt map[[2]int]int8
	// regionBoundary makes every tile with x >= 20 a second region.
	regionBoundary bool
}

// request is the PlaceGroup input the row describes.
func (c tapeCase) request() monsterspawn.GroupRequest {
	return monsterspawn.GroupRequest{
		Num:     c.num,
		Active:  c.active,
		Total:   c.total,
		Base:    c.base,
		Leader:  c.leader,
		Leashed: c.leashed,
	}
}

// CanPlaceMonster is CanPlaceMonster (monster.cpp:282-291) for the test world.
func (w flatWorld) CanPlaceMonster(x, y int) bool {
	if x < 0 || y < 0 || x >= 64 || y >= 64 {
		return false
	}

	if w.placeable == nil {
		return true
	}

	return w.placeable[[2]int{x, y}]
}

// TransAt is dTransVal[x][y].
func (w flatWorld) TransAt(x, y int) int8 {
	if region, ok := w.transAt[[2]int{x, y}]; ok {
		return region
	}

	if w.regionBoundary && x >= 20 {
		return 1
	}

	return 0
}

// openWorld is the open-floor world: every 64x64 tile placeable, one region.
func openWorld() flatWorld { return flatWorld{} }

// scriptedRnd is a GenerateRnd source that returns the committed tape in order
// and fails the test on one draw too many — the roll count is part of the oracle.
func scriptedRnd(t *testing.T, script []int32) monsterspawn.Rnd {
	t.Helper()

	next := 0

	return monsterspawn.Rnd{GenerateRnd: func(bound int32) int32 {
		if next >= len(script) {
			t.Fatalf("roll %d past the committed script (%d rolls), bound %d", next+1, len(script), bound)

			return 0
		}

		value := script[next]
		next++

		if value < 0 || value >= bound {
			t.Fatalf("script roll %d = %d outside [0,%d)", next, value, bound)
		}

		return value
	}}
}

// constRnd is a GenerateRnd source that always returns value.
func constRnd(value int32) monsterspawn.Rnd {
	return monsterspawn.Rnd{GenerateRnd: func(int32) int32 { return value }}
}

// tileWorld accepts exactly the given tiles, all in one transparency region.
func tileWorld(tiles [][2]int) flatWorld {
	placeable := make(map[[2]int]bool, len(tiles))
	for _, tile := range tiles {
		placeable[tile] = true
	}

	return flatWorld{placeable: placeable}
}

// testWorld builds the world a tape row names.
func testWorld(t *testing.T, name string) flatWorld {
	t.Helper()

	switch name {
	case "open":
		return openWorld()
	case "blocked":
		return tileWorld(placeableExcept(17, 18))
	case "twotrans":
		return flatWorld{regionBoundary: true}
	default:
		t.Fatalf("unknown tape world %q", name)

		return flatWorld{}
	}
}

// placeableExcept is the open floor with one blocked tile.
func placeableExcept(blockX, blockY int) [][2]int {
	tiles := make([][2]int, 0, 64*64)

	for x := range 64 {
		for y := range 64 {
			if x == blockX && y == blockY {
				continue
			}

			tiles = append(tiles, [2]int{x, y})
		}
	}

	return tiles
}

// assertTiles compares the placement with the hand-derived expectation.
func assertTiles(t *testing.T, got []monsterspawn.Placed, want []wantTile) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("placed %d monsters, want %d: %+v", len(got), len(want), got)
	}

	for i := range want {
		assertTile(t, i, got[i], want[i])
	}
}

// assertTile compares one placed monster with its expectation.
func assertTile(t *testing.T, index int, got monsterspawn.Placed, want wantTile) {
	t.Helper()

	assertTilePlacement(t, index, got, want)
	assertTileStats(t, index, got, want)
}

// assertTilePlacement checks the tile and the rolled direction of one monster.
func assertTilePlacement(t *testing.T, index int, got monsterspawn.Placed, want wantTile) {
	t.Helper()

	if got.X != want.x || got.Y != want.y {
		t.Fatalf("monster %d at (%d,%d), want (%d,%d)", index, got.X, got.Y, want.x, want.y)
	}

	if got.Direction != want.direction {
		t.Fatalf("monster %d direction %d, want %d", index, got.Direction, want.direction)
	}

	if got.StartFrame != want.startFrame {
		t.Fatalf("monster %d start frame %d, want %d", index, got.StartFrame, want.startFrame)
	}
}

// assertTileStats checks the minion transform of one monster.
func assertTileStats(t *testing.T, index int, got monsterspawn.Placed, want wantTile) {
	t.Helper()

	if got.MaxHitPoints != want.maxHitPoints || got.HitPoints != want.hitPoints {
		t.Fatalf("monster %d hit points %d/%d, want %d/%d",
			index, got.HitPoints, got.MaxHitPoints, want.hitPoints, want.maxHitPoints)
	}

	if got.Intelligence != want.intelligence {
		t.Fatalf("monster %d intelligence %d, want %d", index, got.Intelligence, want.intelligence)
	}

	if got.Leashed != want.leashed {
		t.Fatalf("monster %d leashed = %v, want %v", index, got.Leashed, want.leashed)
	}
}

// runTapes runs every scripted case.
func runTapes(t *testing.T, cases []tapeCase) []monsterspawn.Placement {
	t.Helper()

	placements := make([]monsterspawn.Placement, len(cases))

	for i, tc := range cases {
		placements[i] = monsterspawn.PlaceGroup(testWorld(t, tc.world), scriptedRnd(t, tc.script), tc.request())
	}

	return placements
}

// assertTapeCase checks one placement against its row.
func assertTapeCase(t *testing.T, tc tapeCase, placement monsterspawn.Placement) {
	t.Helper()

	assertTiles(t, placement.Tiles, tc.wantTiles)

	if placement.Rolls != tc.wantRolls {
		t.Fatalf("rolls = %d, want %d", placement.Rolls, tc.wantRolls)
	}

	if placement.Rewinds != tc.wantRewinds {
		t.Fatalf("rewinds = %d, want %d", placement.Rewinds, tc.wantRewinds)
	}

	if placement.PackSize != tc.wantPack {
		t.Fatalf("pack size = %d, want %d", placement.PackSize, tc.wantPack)
	}
}

// gatePasses returns the IDs of the rows of one availability class that pass
// the gate at their own minimum level.
func gatePasses(t *testing.T, rows []monstdat.MonsterData,
	class monsterspawn.Availability, mode monsterspawn.GameMode,
) []string {
	t.Helper()

	var passed []string

	for _, monster := range rows {
		rule, err := monsterspawn.ParseSpawnRule(monster)
		if err != nil {
			t.Fatalf("ParseSpawnRule(%s): %v", monster.ID, err)
		}

		if rule.Availability == class && monsterspawn.IsMonsterAvailable(rule, mode, rule.MinDunLvl) {
			passed = append(passed, monster.ID)
		}
	}

	return passed
}

// gateMismatches returns one line per row whose gate disagrees with its class
// and level band, at the given levels.
func gateMismatches(t *testing.T, rows []monstdat.MonsterData, levels []int) []string {
	t.Helper()

	mismatches := make([]string, 0, len(rows)*len(levels))

	for _, monster := range rows {
		rule, err := monsterspawn.ParseSpawnRule(monster)
		if err != nil {
			t.Fatalf("ParseSpawnRule(%s): %v", monster.ID, err)
		}

		for _, level := range levels {
			mismatches = append(mismatches, gateMismatch(t, monster, rule, level)...)
		}
	}

	return mismatches
}

// gateMismatch lists the mode disagreements of one row at one level.
func gateMismatch(t *testing.T, monster monstdat.MonsterData, rule monsterspawn.SpawnRule, level int,
) []string {
	t.Helper()

	inBand := level >= rule.MinDunLvl && level <= rule.MaxDunLvl
	wanted := map[monsterspawn.GameMode]bool{
		monsterspawn.Retail:    inBand && rule.Availability != monsterspawn.AvailabilityNever,
		monsterspawn.Shareware: inBand && rule.Availability == monsterspawn.AvailabilityAlways,
	}

	var mismatches []string

	for mode, want := range wanted {
		if got := monsterspawn.IsMonsterAvailable(rule, mode, level); got != want {
			mismatches = append(mismatches, fmt.Sprintf("%s level %d mode %d: gate = %v, want %v",
				monster.ID, level, mode, got, want))
		}
	}

	return mismatches
}

// readTapes parses the committed scripted tapes; the row count comes from the
// file, never from a literal.
func readTapes(t *testing.T) []tapeCase {
	t.Helper()

	raw, err := os.ReadFile(tapesPath)
	if err != nil {
		t.Fatalf("read committed tapes: %v", err)
	}

	var cases []tapeCase

	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		cases = append(cases, parseTape(t, i+1, strings.Split(line, "\t")))
	}

	return cases
}

// parseTape parses one tape row.
func parseTape(t *testing.T, lineNo int, fields []string) tapeCase {
	t.Helper()

	if len(fields) != tapeColumns {
		t.Fatalf("tape line %d has %d columns, want %d", lineNo, len(fields), tapeColumns)
	}

	return tapeCase{
		name:        fields[0],
		world:       fields[1],
		num:         atoiColumn(t, lineNo, fields[2]),
		active:      atoiColumn(t, lineNo, fields[3]),
		total:       atoiColumn(t, lineNo, fields[4]),
		leader:      parseLeader(t, lineNo, fields[5], fields[11]),
		leashed:     fields[6] == "true",
		base:        parseBase(t, lineNo, fields),
		script:      parseInt32s(t, lineNo, fields[12]),
		wantTiles:   parseTiles(t, lineNo, fields[13]),
		wantRolls:   atoiColumn(t, lineNo, fields[14]),
		wantRewinds: atoiColumn(t, lineNo, fields[15]),
		wantPack:    atoiColumn(t, lineNo, fields[16]),
	}
}

// parseBase reads the base stat block columns.
func parseBase(t *testing.T, lineNo int, fields []string) monsterspawn.Stats {
	t.Helper()

	return monsterspawn.Stats{
		MaxHitPoints: atoiColumn(t, lineNo, fields[7]),
		Intelligence: atoiColumn(t, lineNo, fields[8]),
		HitPoints:    atoiColumn(t, lineNo, fields[7]),
		Frames:       atoiColumn(t, lineNo, fields[9]),
		Gargoyle:     fields[10] == "true",
	}
}

// parseLeader reads the leader column ("-" for an unled pack).
func parseLeader(t *testing.T, lineNo int, position, intelligence string) *monsterspawn.Leader {
	t.Helper()

	if position == "-" {
		return nil
	}

	parts := strings.Split(position, ",")
	if len(parts) != 2 {
		t.Fatalf("tape line %d leader %q, want x,y", lineNo, position)
	}

	return &monsterspawn.Leader{
		X:            atoiColumn(t, lineNo, parts[0]),
		Y:            atoiColumn(t, lineNo, parts[1]),
		Intelligence: atoiColumn(t, lineNo, intelligence),
	}
}

// parseTiles reads the hand-derived placement column.
func parseTiles(t *testing.T, lineNo int, column string) []wantTile {
	t.Helper()

	entries := strings.Fields(column)
	tiles := make([]wantTile, 0, len(entries))

	for _, entry := range entries {
		parts := strings.Split(entry, ",")
		if len(parts) != 8 {
			t.Fatalf("tape line %d tile %q, want 8 fields", lineNo, entry)
		}

		tiles = append(tiles, wantTile{
			x:            atoiColumn(t, lineNo, parts[0]),
			y:            atoiColumn(t, lineNo, parts[1]),
			direction:    atoiColumn(t, lineNo, parts[2]),
			startFrame:   atoiColumn(t, lineNo, parts[3]),
			maxHitPoints: atoiColumn(t, lineNo, parts[4]),
			hitPoints:    atoiColumn(t, lineNo, parts[5]),
			intelligence: atoiColumn(t, lineNo, parts[6]),
			leashed:      parts[7] == "1",
		})
	}

	return tiles
}

// parseInt32s parses the space-separated roll script.
func parseInt32s(t *testing.T, lineNo int, column string) []int32 {
	t.Helper()

	fields := strings.Fields(column)
	values := make([]int32, 0, len(fields))

	for _, field := range fields {
		values = append(values, int32(atoiColumn(t, lineNo, field)))
	}

	return values
}

// atoiColumn parses one integer column.
func atoiColumn(t *testing.T, lineNo int, column string) int {
	t.Helper()

	value, err := strconv.Atoi(column)
	if err != nil {
		t.Fatalf("tape line %d column %q: %v", lineNo, column, err)
	}

	return value
}

// committedMonsters loads the base and Hellfire monstdat tables.
func committedMonsters(t *testing.T) []monstdat.MonsterData {
	t.Helper()

	base, err := monstdat.ParseMonstdat(corpus(t, "assets/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}

	hf, err := monstdat.ParseMonstdat(corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}

	return append(base, hf...)
}

// corpus reads one committed gamedata fixture.
func corpus(t *testing.T, rel string) []byte {
	t.Helper()

	return testsupport.Corpus(t, "monstdat", rel)
}
