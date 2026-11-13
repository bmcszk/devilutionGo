package persist_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// goldenStreamPath is the committed composed-stream fixture.
const goldenStreamPath = "testdata/compose_game_stream.hex"

// monsterIDTableOff is the monster id table offset in the fixed compose
// stream: header 43 + seeds 136 + hero 1266 + quests 44 + portals 48 + kill
// counts 800.
const monsterIDTableOff = persist.GameHeaderSize + 17*8 + 1266 + 44 + 48 + 800

// missileIndexBytes is the two 125-entry missile index tables plus the
// saved-count skip (0 saved missiles in the fixed input).
const missileIndexBytes = 2 * persist.MaxMissilesForSaveGame

// composeInput is the fixed game-side state the composed golden is built
// from: a retail (non-Hellfire, non-spawn) session on Cathedral level 1 with
// one quest, two portals, two kill counts, one light, two objects, one
// dropped item, and the saved grid state.
type composeInput struct {
	Mode    persist.GameMode
	Header  persist.GameHeader
	Hero    persist.PlayerPack
	Quests  []persist.QuestSave
	Portals []persist.PortalRecord
	Kills   []int32
	Objects []persist.ObjectRecord
	Lights  []persist.LightRecord
	Items   persist.DroppedItems
	Grids   composeGrids
}

// composeGrids is the grid state of composeInput.
type composeGrids struct {
	DLight         [persist.DungeonWidth * persist.DungeonHeight]byte
	DFlags         [persist.DungeonWidth * persist.DungeonHeight]byte
	DPlayer        [persist.DungeonWidth * persist.DungeonHeight]int8
	DItem          [persist.DungeonWidth * persist.DungeonHeight]byte
	Monsters       [persist.DungeonWidth * persist.DungeonHeight]int32
	Corpses        [persist.DungeonWidth * persist.DungeonHeight]int8
	DObjects       [persist.DungeonWidth * persist.DungeonHeight]int8
	PreLights      [persist.DungeonWidth * persist.DungeonHeight]byte
	Automap        [persist.AutomapWidth * persist.AutomapHeight]byte
	MonsterRecords [][]byte
	MissileRecords [][]byte
}

// TestComposeSaveGameDataOrder pins the section order of the composed stream
// against the pin walk (loadsave.cpp:2763-2935): header, hero, quests,
// portals, kill counts, dungeon records, dropped items, unique item flags,
// the common grids, the dungeon grid tail, premium tail and automap flags.
func TestComposeSaveGameDataOrder(t *testing.T) {
	// given the fixed compose input
	in := goldenCompose()

	// when the stream is composed
	stream, err := persist.ComposeSaveGameData(composeRequest(in))
	if err != nil {
		t.Fatalf("ComposeSaveGameData: %v", err)
	}

	// then each section sits at its walk offset
	assertComposeOffsets(t, in, stream)
}

// TestComposeSaveGameDataIdempotent pins compose determinism: the same input
// composes to byte-identical streams.
func TestComposeSaveGameDataIdempotent(t *testing.T) {
	// given the fixed compose input
	input := composeRequest(goldenCompose())

	// when it is composed twice
	first, err := persist.ComposeSaveGameData(input)
	if err != nil {
		t.Fatalf("first compose: %v", err)
	}

	second, err := persist.ComposeSaveGameData(input)
	if err != nil {
		t.Fatalf("second compose: %v", err)
	}

	// then the streams are identical
	if !bytes.Equal(first, second) {
		t.Fatal("compose is not deterministic")
	}
}

// TestComposeSaveGameDataGolden pins the composed stream against the
// committed golden; its byte layout maps 1:1 to the pin walk with offset
// comments in the fixture header.
func TestComposeSaveGameDataGolden(t *testing.T) {
	// given the composed stream
	stream := composeStream(t)

	// when the committed golden is decoded
	want, err := hex.DecodeString(readGoldenLines(t))
	if err != nil {
		t.Fatalf("golden hex: %v", err)
	}

	// then the composed stream matches byte-for-byte
	if !bytes.Equal(stream, want) {
		t.Fatalf("composed stream differs from golden at offset %d (len got %d want %d)",
			firstDiffOffset(stream, want), len(stream), len(want))
	}
}

// TestComposeSaveGameDataStoreRoundTrip composes the stream, stores it via
// SQLiteStore.SaveGame, loads it back and checks the bytes survived
// (roundtrip through the store, size floor at GameHeaderSize included).
func TestComposeSaveGameDataStoreRoundTrip(t *testing.T) {
	// given a composed stream saved into a fresh hero store
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hero.db")
	stream := composeStream(t)

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	defer func() { _ = store.Close() }()

	hero := persist.HeroRecord{Slot: 0, Name: "Golden", ClassID: 0, Pack: composeHero().Encode()}
	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.SaveGame(ctx, 0, "single", stream); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	// when the stream is loaded back
	loaded, err := store.LoadGame(ctx, 0, "single")
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}

	// then it survived byte-for-byte and respects the header size floor
	if len(loaded) < persist.GameHeaderSize {
		t.Fatalf("loaded stream = %d bytes, below GameHeaderSize %d", len(loaded), persist.GameHeaderSize)
	}

	if !bytes.Equal(loaded, stream) {
		t.Fatal("store roundtrip changed the composed stream")
	}
}

// TestComposeSaveGameDataSeeds pins the DungeonSeeds loop between the header
// counts and the hero: giNumberOfLevels x (BE uint32 seed + BE int32
// hellfire level type), 17 entries retail / 25 Hellfire (loadsave.cpp:2824-2828
// at ca91d778f, memory_map game.txt rows 'C_DA 17 levelSeed levelSeeds' /
// 'C_HF 25').
func TestComposeSaveGameDataSeeds(t *testing.T) {
	t.Parallel()

	// given a request with per-level seeds and types
	req := composeRequest(goldenCompose())
	req.LevelSeeds = make([]uint32, 17)
	req.LevelTypes = make([]int32, 17)
	for i := range req.LevelSeeds {
		req.LevelSeeds[i] = uint32(1000 + i)
	}
	req.LevelTypes[0] = 1

	// when the stream is composed
	stream, err := persist.ComposeSaveGameData(req)
	if err != nil {
		t.Fatalf("ComposeSaveGameData: %v", err)
	}

	// then the 17x8-byte seed block sits right after the 43-byte header and
	// the hero follows it
	assertSeedsBlock(t, stream)
}

// TestComposeSaveGameDataFixedIDTables pins the fixed-width id tables of the
// walk: ActiveMonsters is ALWAYS 200 BE uint32 over the fixed array
// (monster.h:496, memory_map 'M_DL 200 32'), ActiveObjects and
// AvailableObjects are ALWAYS 127+127 int8 (objects.h MAXOBJECTS, two
// 'M_DL 127 8' rows), ActiveLights is ALWAYS 32 uint8 (lighting.h MAXLIGHTS,
// 'M_DL 32 8'); short inputs pad with zero exactly as the upstream fixed
// arrays do (loadsave.cpp:2845-2847, :2870-2872, :2874-2875).
func TestComposeSaveGameDataFixedIDTables(t *testing.T) {
	t.Parallel()

	// given a request whose id lists are shorter than the fixed widths
	req := composeRequest(goldenCompose())
	req.Dungeon.ActiveObjectIDs = []int8{0}
	req.Dungeon.AvailableObjectIDs = []int8{1}
	req.Dungeon.ActiveLightIDs = []uint8{0}

	// when the stream is composed
	stream, err := persist.ComposeSaveGameData(req)
	if err != nil {
		t.Fatalf("ComposeSaveGameData: %v", err)
	}

	// then the monster id table is 800 bytes with the single id 0 first and
	// zero padding after, the object id tables are 127+127 bytes with the
	// supplied ids first, and the light id list is 32 bytes
	if !bytes.Equal(stream[monsterIDTableOff:monsterIDTableOff+4], []byte{0, 0, 0, 0}) {
		t.Fatal("first monster id not 0")
	}

	if !allZero(stream[monsterIDTableOff+4 : monsterIDTableOff+800]) {
		t.Fatal("monster id padding not zero")
	}

	objectIDOff := monsterIDTableOff + 800 + monsterRecordBytes(req) + missileIndexBytes
	if !bytes.Equal(stream[objectIDOff:objectIDOff+127], objectIDTable(0)) {
		t.Fatal("ActiveObjects table wrong")
	}

	if !bytes.Equal(stream[objectIDOff+127:objectIDOff+254], objectIDTable(1)) {
		t.Fatal("AvailableObjects table wrong")
	}

	lightIDOff := objectIDOff + 254 + 2*persist.ObjectSaveSize + 4
	if !bytes.Equal(stream[lightIDOff:lightIDOff+32], lightIDTable(0)) {
		t.Fatal("ActiveLights table wrong")
	}
}

// TestComposeSaveGameDataFixedIDTablesFullWidth pins the full-width case: a
// table of exactly the fixed width passes through unchanged.
func TestComposeSaveGameDataFixedIDTablesFullWidth(t *testing.T) {
	t.Parallel()

	// given a request with the full 200-monster id table
	req := composeRequest(goldenCompose())
	ids := make([]uint32, 200)
	ids[199] = 199
	req.ActiveMonsters = ids

	// when the stream is composed
	stream, err := persist.ComposeSaveGameData(req)
	if err != nil {
		t.Fatalf("ComposeSaveGameData: %v", err)
	}

	// then the table is exactly 800 bytes and the last id sits in slot 199
	last := monsterIDTableOff + 4*199
	if binary.BigEndian.Uint32(stream[last:]) != 199 {
		t.Fatalf("monster id 199 = %x", stream[last:last+4])
	}
}

// assertSeedsBlock asserts the fixed seeds of TestComposeSaveGameDataSeeds at
// the header offset and the hero directly after the block.
func assertSeedsBlock(t *testing.T, stream []byte) {
	t.Helper()

	off := persist.GameHeaderSize
	for i := range 17 {
		seed := binary.BigEndian.Uint32(stream[off:])
		if seed != uint32(1000+i) {
			t.Fatalf("seed %d = %d, want %d", i, seed, 1000+i)
		}

		if i == 0 && !bytes.Equal(stream[off+4:off+8], []byte{0, 0, 0, 1}) {
			t.Fatalf("level type 0 = %x, want 1", stream[off+4:off+8])
		}
		off += 8
	}

	hero := goldenCompose().Hero.Encode()
	if !bytes.Equal(stream[off:off+len(hero)], hero) {
		t.Fatalf("hero not at offset %d after the seed block", off)
	}
}

// monsterRecordBytes sums the monster record images in the request.
func monsterRecordBytes(req persist.SaveGameDataRequest) int {
	total := 0
	for _, record := range req.MonsterRecords {
		total += len(record)
	}

	return total
}

// objectIDTable builds the expected 127-byte id table: id first, zero padding.
func objectIDTable(id int8) []byte {
	table := make([]byte, 127)
	table[0] = uint8(id)

	return table
}

// lightIDTable builds the expected 32-byte id table: id first, zero padding.
func lightIDTable(id uint8) []byte {
	table := make([]byte, 32)
	table[0] = id

	return table
}

// allZero reports whether every byte is 0.
func allZero(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}

	return true
}

// composeStream composes the fixed golden input, failing the test on error.
func composeStream(t *testing.T) []byte {
	t.Helper()

	stream, err := persist.ComposeSaveGameData(composeRequest(goldenCompose()))
	if err != nil {
		t.Fatalf("ComposeSaveGameData: %v", err)
	}

	return stream
}

// goldenCompose builds the fixed compose input: seed the grids with
// recognizable row-outer values so offsets map 1:1 to the pin walk.
func goldenCompose() composeInput {
	in := composeInput{
		Mode: persist.GameMode{},
		Header: persist.GameHeader{
			Hellfire: false, Spawn: false,
			SetLevel: false, SetLevelNum: 0, CurrLevel: 1, LevelType: 1,
			ViewX: 75, ViewY: 68,
			InventoryFlag: true, CharFlag: false,
			ActiveMonsters: 1, ActiveItems: 1, ActiveMissiles: 0, ActiveObjects: 2,
		},
	}
	in.Hero = composeHero()
	in.Quests = []persist.QuestSave{{Level: 1, Idx: 13, Active: 1}}
	in.Portals = []persist.PortalRecord{
		{Open: true, X: 75, Y: 68, Level: 4, LType: 1},
		{Open: false, X: 0, Y: 0, Level: 0, LType: 0},
	}
	in.Kills = []int32{3, 7}
	in.Objects = []persist.ObjectRecord{
		{Type: 1157, X: 63, Y: 62, ApplyLighting: true, SolidFlag: true, SelectionRegion: 4, OLid: 1, ORndSeed: 11},
		{Type: 3, X: 64, Y: 62, SelectionRegion: 4, OLid: 2, ORndSeed: 22, Var1: -7},
	}
	in.Lights = []persist.LightRecord{{X: 63, Y: 62, Radius: 11, HasChanged: true, OldRadius: 11}}
	in.Items = composeDroppedItems()
	in.Grids = composeGridsState()

	return in
}

// composeHero is the fixed warrior pack of the composed golden.
func composeHero() persist.PlayerPack {
	player := persist.PlayerPack{Level: 2}
	copy(player.Name[:], "Golden")
	player.Px, player.Py = 75, 68
	player.Exp, player.Gold = 350, 1200

	return player
}

// composeDroppedItems is the fixed one-item dropped set.
func composeDroppedItems() persist.DroppedItems {
	record := make([]byte, persist.ItemSaveSizeDiablo)
	record[0] = 0xAB

	return persist.DroppedItems{ActiveCount: 1, Records: [][]byte{record}}
}

// composeGridsState seeds the grids with recognizable values.
func composeGridsState() composeGrids {
	var grids composeGrids

	for i := range grids.DLight {
		grids.DLight[i] = 15
	}

	grids.DFlags[0] = 0xC8
	grids.DPlayer[1] = 1
	grids.DItem[2] = 1
	grids.Monsters[3] = 1
	grids.Corpses[4] = 1
	grids.DObjects[5] = 6
	grids.PreLights[6] = 1
	grids.Automap[7] = 1
	grids.MonsterRecords = [][]byte{make([]byte, persist.MonsterSaveSize)}

	return grids
}

// composeRequest adapts composeInput to the persist request struct.
func composeRequest(in composeInput) persist.SaveGameDataRequest {
	dungeon := dungeonHalf(in)

	seeds := make([]uint32, 17)
	types := make([]int32, 17)
	for i := range types {
		types[i] = 1 // cathedral across the 17 levels of the fixed input
	}

	return persist.SaveGameDataRequest{
		Mode:           in.Mode,
		Header:         in.Header,
		LevelSeeds:     seeds,
		LevelTypes:     types,
		Hero:           in.Hero.Encode(),
		Quests:         in.Quests,
		Portals:        in.Portals,
		KillCounts:     in.Kills,
		ActiveMonsters: dungeon.ActiveMonsterIDs,
		MonsterRecords: in.Grids.MonsterRecords,
		MissileRecords: in.Grids.MissileRecords,
		Dungeon:        dungeon,
		VisionLights:   visionLights(),
		ActiveObjects:  in.Objects,
		ActiveLights:   in.Lights,
		DroppedItems:   in.Items,
		UniqueItemFlag: make([]bool, 128),
		DLight:         in.Grids.DLight,
		DFlags:         in.Grids.DFlags,
		DPlayer:        in.Grids.DPlayer,
		DItem:          in.Grids.DItem,
		PremiumCount:   1,
		PremiumLevel:   2,
		PremiumItems:   [][]byte{make([]byte, persist.ItemSaveSizeDiablo)},
		AutomapActive:  true,
		AutoMapScaleX:  50,
	}
}

// dungeonHalf maps the dungeon-only fields of composeInput.
func dungeonHalf(in composeInput) persist.SaveGameDataDungeon {
	return persist.SaveGameDataDungeon{
		MonsterGrid:      in.Grids.Monsters,
		CorpseGrid:       in.Grids.Corpses,
		ObjectGrid:       in.Grids.DObjects,
		PreLightGrid:     in.Grids.PreLights,
		AutomapView:      in.Grids.Automap,
		ActiveMonsterIDs: []uint32{0},
		ActiveObjectIDs:  []int8{0, 1},
		ActiveLightIDs:   []uint8{0},
		VisionLights:     visionLights(),
	}
}

// visionLights is the per-player vision list of the compose input: two vision
// light records.
func visionLights() []persist.LightRecord {
	return []persist.LightRecord{
		{X: 75, Y: 68, Radius: 10},
		{X: 76, Y: 69, Radius: 10},
	}
}

// assertComposeOffsets walks the stream and checks each section boundary.
func assertComposeOffsets(t *testing.T, in composeInput, stream []byte) {
	t.Helper()

	off := assertHeadSections(t, in, stream)
	off = assertDungeonSections(t, in, stream, off)
	off = assertCommonSections(t, in, stream, off)
	off = assertTailSections(t, in, stream, off)

	if off != len(stream) {
		t.Fatalf("stream has %d trailing bytes after the walk", len(stream)-off)
	}
}

// expectSection asserts stream[off:off+len(image)] equals image and returns
// the next offset.
func expectSection(t *testing.T, stream []byte, off int, image []byte, name string) int {
	t.Helper()

	if !bytes.Equal(stream[off:off+len(image)], image) {
		t.Fatalf("%s section at %d mismatch", name, off)
	}

	return off + len(image)
}

// assertHeadSections checks header, hero, quests, portals and kill counts.
func assertHeadSections(t *testing.T, in composeInput, stream []byte) int {
	t.Helper()

	header, err := in.Header.Encode()
	if err != nil {
		t.Fatalf("header encode: %v", err)
	}

	off := expectSection(t, stream, 0, header, "header")
	off = assertSeeds(t, stream, off)
	off = expectSection(t, stream, off, in.Hero.Encode(), "hero")

	for _, quest := range in.Quests {
		off = expectSection(t, stream, off, quest.Encode(in.Mode), "quest")
	}

	for _, portal := range in.Portals {
		off = expectSection(t, stream, off, portal.Encode(), "portal")
	}

	return expectSection(t, stream, off, persist.MonsterKillCountsBlock(in.Kills), "kill counts")
}

// assertSeeds checks the 17x8 DungeonSeeds block between the counts and the
// hero (loadsave.cpp:2824-2828 at ca91d778f).
func assertSeeds(t *testing.T, stream []byte, off int) int {
	t.Helper()

	for i := range 17 {
		if !bytes.Equal(stream[off+8*i+4:off+8*i+8], []byte{0, 0, 0, 1}) {
			t.Fatalf("level type %d not 1 at %d", i, off+8*i+4)
		}
	}

	return off + 17*8
}

// assertDungeonSections checks the dungeon-only record blocks.
func assertDungeonSections(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	off = assertMonsterBlock(t, in, stream, off)
	off = assertObjectBlock(t, in, stream, off)

	return assertLightBlock(t, in, stream, off)
}

// assertMonsterBlock checks monster ids and records plus the missile tables.
func assertMonsterBlock(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	// monster ids: one BE uint32 in the fixed 200-slot table, the rest zero
	// padding (loadsave.cpp:2845-2847, monster.h:496)
	off += 4 * persist.MaxMonsters
	off = expectSection(t, stream, off, in.Grids.MonsterRecords[0], "monster record")

	// missile index tables: 125 active ids then the available list, then the
	// saved-count skip (loadsave.cpp:2849-2857)
	return off + 2*persist.MaxMissilesForSaveGame
}

// assertObjectBlock checks object ids and records.
func assertObjectBlock(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	off += 2 * persist.MaxObjects // ActiveObjects + AvailableObjects fixed tables

	for i := range in.Objects {
		off = expectSection(t, stream, off, in.Objects[i].Encode(), "object record")
	}

	return off
}

// assertLightBlock checks the light count, ids and records, then the vision
// block (VisionId/VisionCount BE + per-player vision records,
// loadsave.cpp:2869-2882).
func assertLightBlock(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	if stream[off] != 0 || stream[off+3] != 1 {
		t.Fatalf("active light count at %d mismatch", off)
	}

	off += 4 + persist.MaxLights

	for i := range in.Lights {
		off = expectSection(t, stream, off, in.Lights[i].Encode(), "light record")
	}

	off += 8

	for i := range 2 {
		off = expectSection(t, stream, off, visionLights()[i].EncodeVision(), "vision record")
	}

	return off
}

// assertCommonSections checks dropped items, unique flags and the common
// grids, returning the offset after dItem.
func assertCommonSections(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	items, err := in.Items.Encode(persist.ItemSaveSizeDiablo)
	if err != nil {
		t.Fatalf("dropped items encode: %v", err)
	}

	off = expectSection(t, stream, off, items, "dropped items")

	for i := range 128 {
		if stream[off+i] != 0 {
			t.Fatalf("unique item flag %d not zero", i)
		}
	}

	off += 128
	off = expectSection(t, stream, off, persist.EncodeDLightGrid(in.Grids.DLight), "dLight")
	off = expectSection(t, stream, off, persist.EncodeDFlagsGrid(in.Grids.DFlags), "dFlags")
	off = expectSection(t, stream, off, persist.EncodeDPlayerGrid(in.Grids.DPlayer), "dPlayer")
	off = expectSection(t, stream, off, persist.EncodeDroppedItemLocations(in.Grids.DItem), "dItem")

	return off
}

// assertTailSections checks the dungeon grid tail, premium tail and automap
// flags, returning the end offset.
func assertTailSections(t *testing.T, in composeInput, stream []byte, off int) int {
	t.Helper()

	dungeon := persist.EncodeDungeonGrids(persist.DungeonGrids{
		Monsters:  in.Grids.Monsters,
		Corpses:   in.Grids.Corpses,
		Objects:   in.Grids.DObjects,
		Lights:    in.Grids.DLight,
		PreLights: in.Grids.PreLights,
		Automap:   in.Grids.Automap,
	})
	off = expectSection(t, stream, off, dungeon, "dungeon grids")

	tail, err := persist.PremiumTailBlock(1, 2, [][]byte{make([]byte, persist.ItemSaveSizeDiablo)}, in.Mode)
	if err != nil {
		t.Fatalf("premium tail encode: %v", err)
	}

	off = expectSection(t, stream, off, tail, "premium tail")

	if stream[off] != 1 {
		t.Fatalf("automap active flag = %d, want 1", stream[off])
	}

	off++
	scale := uint32(stream[off])<<24 | uint32(stream[off+1])<<16 | uint32(stream[off+2])<<8 | uint32(stream[off+3])
	if got := int32(scale); got != 50 {
		t.Fatalf("automap scale = %d, want 50", got)
	}

	return off + 4
}

// readGoldenLines reads the concatenated hex payload of the golden file.
func readGoldenLines(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile(goldenStreamPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	var hexText []byte

	for _, line := range bytes.Split(raw, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}

		hexText = append(hexText, trimmed...)
	}

	return string(hexText)
}

// firstDiffOffset returns the first differing offset, or -1.
func firstDiffOffset(got, want []byte) int {
	n := min(len(got), len(want))
	for i := range n {
		if got[i] != want[i] {
			return i
		}
	}

	if len(got) != len(want) {
		return n
	}

	return -1
}
