package persist

import "fmt"

// SaveGameDataDungeon carries the dungeon-only walk state: the active
// monster/object/light id lists, the per-record images, the vision list, and
// the grids the tail writes (Source/loadsave.cpp:2845-2864, :1885-1916).
type SaveGameDataDungeon struct {
	// ActiveMonsterIDs is the ActiveMonsters id list (BE uint32 per id,
	// :2845-2847).
	ActiveMonsterIDs []uint32
	// MonsterRecords holds one opaque SaveMonster image per active monster.
	MonsterRecords [][]byte
	// MissileRecords holds one opaque SaveMissile image per saved missile
	// (:2858-2864); the index tables are derived from the count.
	MissileRecords [][]byte
	// ActiveObjectIDs is ActiveObjects as int8 ids over the fixed MAXOBJECTS
	// array, zero-padded to 127 (loadsave.cpp:2870-2872, objects.h MAXOBJECTS,
	// memory_map 'M_DL 127 8 ActiveObjects').
	ActiveObjectIDs []int8
	// AvailableObjectIDs is AvailableObjects over the same fixed array, zero-
	// padded to 127 (loadsave.cpp:2873, memory_map 'M_DL 127 8
	// AvailableObjects').
	AvailableObjectIDs []int8
	// ActiveLightIDs is the ActiveLights id list over the fixed MAXLIGHTS
	// array, zero-padded to 32 (LE uint8, :2874-2875, lighting.h MAXLIGHTS,
	// memory_map 'M_DL 32 8 ActiveLights').
	ActiveLightIDs []uint8
	// VisionLights is the per-player vision list, saved with the vision flag
	// (:2878-2882).
	VisionLights []LightRecord
	// MonsterGrid is dMonster (BE int32).
	MonsterGrid [dungeonGridCells]int32
	// CorpseGrid is dCorpse.
	CorpseGrid [dungeonGridCells]int8
	// ObjectGrid is the dungeon dObject grid.
	ObjectGrid [dungeonGridCells]int8
	// PreLightGrid is dPreLight.
	PreLightGrid [dungeonGridCells]uint8
	// AutomapView is AutomapView (DMAXX x DMAXY).
	AutomapView [automapGridCells]uint8
	// MissileTiles holds the linear indexes of missile-occupied tiles; the
	// dMissile tail grid writes -1 at these and 0 elsewhere
	// (loadsave.cpp:2907-2911, TileContainsMissile).
	MissileTiles []int
}

// SaveGameDataRequest is the game-side state ComposeSaveGameData turns into
// the SaveGameData stream (Source/loadsave.cpp:2763-2935).
type SaveGameDataRequest struct {
	// Mode selects magic, level counts and record widths.
	Mode GameMode
	// Header is the fixed stream head (GameHeader.Encode).
	Header GameHeader
	// Hero is the opaque PlayerPack image written after the header (:2803).
	Hero []byte
	// LevelSeeds is DungeonSeeds, one BE uint32 per level, written after the
	// header counts and before the hero (loadsave.cpp:2824-2826 at
	// ca91d778f; memory_map game.txt 'C_DA 17 / C_HF 25 levelSeed').
	LevelSeeds []uint32
	// LevelTypes is the per-level GetLevelType image, one BE int32 per level
	// after each seed (loadsave.cpp:2827). Callers pass the type already
	// mapped through getHellfireLevelType.
	LevelTypes []int32
	// Quests holds one record per giNumberQuests (:2805-2807).
	Quests []QuestSave
	// Portals holds MAXPORTAL records (:2808-2810).
	Portals []PortalRecord
	// KillCounts is MonsterKillCounts; the encoder pads to MaxMonsters
	// (:2811-2816).
	KillCounts []int32
	// ActiveMonsters is ActiveMonsters over the fixed MaxMonsters array; the
	// writer zero-pads to 200 BE uint32 (loadsave.cpp:2845-2847,
	// monster.h:496, memory_map 'M_DL 200 32 ActiveMonsters').
	ActiveMonsters []uint32
	// MonsterRecords holds one SaveMonster image per active monster.
	MonsterRecords [][]byte
	// MissileRecords holds one SaveMissile image per saved missile.
	MissileRecords [][]byte
	// ActiveObjects holds the active objects; ids come from their records'
	// positions (:2865-2867, :2870-2872).
	ActiveObjects []ObjectRecord
	// ActiveLights holds the active lights (:2873-2877).
	ActiveLights []LightRecord
	// VisionLights is the per-player vision list (:2878-2882).
	VisionLights []LightRecord
	// DroppedItems is the floor item set (:2866 SaveDroppedItems).
	DroppedItems DroppedItems
	// UniqueItemFlag is the 128-slot flag array (:2868-2870).
	UniqueItemFlag []bool
	// DLight/DFlags/DPlayer are the common-section grids (:2871-2883).
	DLight  [dungeonGridCells]uint8
	DFlags  [dungeonGridCells]uint8
	DPlayer [dungeonGridCells]int8
	// DItem is the dItem grid remapped to save positions (:2884).
	DItem [dungeonGridCells]uint8
	// Dungeon carries the non-town tail state; empty on town.
	Dungeon SaveGameDataDungeon
	// PremiumCount and PremiumLevel are the BE premium counters (:2921-2922).
	PremiumCount int32
	PremiumLevel int32
	// PremiumItems holds the premium SaveItem images (:2923-2924).
	PremiumItems [][]byte
	// AutomapActive and AutoMapScaleX end the stream (:2925-2926).
	AutomapActive bool
	AutoMapScaleX int32
}

// ComposeSaveGameData assembles the whole SaveGameData stream in the upstream
// write order (Source/loadsave.cpp:2763-2935): header, hero, quests, portals,
// kill counts, the dungeon record blocks, dropped items, unique item flags,
// the common grids, the dungeon grid tail with the double-dLight quirk, the
// premium tail and the automap flags.
func ComposeSaveGameData(req SaveGameDataRequest) ([]byte, error) {
	var out recordWriter

	if err := composeHead(&out, req); err != nil {
		return nil, err
	}

	if err := composeHeroAndLists(&out, req); err != nil {
		return nil, err
	}
	composeDungeonRecords(&out, req)

	if err := composeCommonGrids(&out, req); err != nil {
		return nil, err
	}

	if req.Header.LevelType != 0 { // DTYPE_TOWN
		dungeon := DungeonGrids{
			Monsters:  req.Dungeon.MonsterGrid,
			Corpses:   req.Dungeon.CorpseGrid,
			Objects:   req.Dungeon.ObjectGrid,
			Lights:    req.DLight,
			PreLights: req.Dungeon.PreLightGrid,
			Automap:   req.Dungeon.AutomapView,
		}
		out.bytes(EncodeDungeonGrids(dungeon))
		out.bytes(EncodeMissileGrid(req.Dungeon.MissileTiles))
	}

	if err := composeTail(&out, req); err != nil {
		return nil, err
	}

	return out.buf, nil
}

// composeHead writes the fixed header (loadsave.cpp:2765-2800).
func composeHead(out *recordWriter, req SaveGameDataRequest) error {
	header, err := req.Header.Encode()
	if err != nil {
		return fmt.Errorf("persist: compose header: %w", err)
	}

	out.bytes(header)

	return nil
}

// composeHeroAndLists writes the seeds loop, the hero pack, quests, portals
// and kill counts (loadsave.cpp:2824-2828 at ca91d778f, :2803-2816).
func composeHeroAndLists(out *recordWriter, req SaveGameDataRequest) error {
	counts := LevelCountsFor(req.Mode)
	if err := validateFixedCounts(req, counts); err != nil {
		return err
	}

	for i, seed := range req.LevelSeeds {
		out.be32(seed)
		out.be32(uint32(req.LevelTypes[i]))
	}

	out.bytes(req.Hero)

	if len(req.Quests) != counts.Quests {
		return fmt.Errorf("persist: %d quests != giNumberQuests %d: %w", len(req.Quests), counts.Quests,
			ErrInvalidGameState)
	}

	if len(req.Portals) != MaxPortals {
		return fmt.Errorf("persist: %d portals != MAXPORTAL %d: %w", len(req.Portals), MaxPortals,
			ErrInvalidGameState)
	}

	for _, quest := range req.Quests {
		out.bytes(quest.Encode(req.Mode))
	}

	for _, portal := range req.Portals {
		out.bytes(portal.Encode())
	}

	out.bytes(MonsterKillCountsBlock(req.KillCounts))

	return nil
}

// validateFixedCounts rejects requests whose fixed-count sections do not
// match the walk: one seed and type per level, giNumberQuests quests and
// MAXPORTAL portals (loadsave.cpp:2824-2810).
func validateFixedCounts(req SaveGameDataRequest, counts LevelCounts) error {
	if len(req.LevelSeeds) != counts.Levels || len(req.LevelTypes) != counts.Levels {
		return fmt.Errorf("persist: level seeds %d / types %d != %d levels: %w",
			len(req.LevelSeeds), len(req.LevelTypes), counts.Levels, ErrInvalidGameState)
	}

	if len(req.Quests) != counts.Quests {
		return fmt.Errorf("persist: %d quests != giNumberQuests %d: %w", len(req.Quests), counts.Quests,
			ErrInvalidGameState)
	}

	if len(req.Portals) != MaxPortals {
		return fmt.Errorf("persist: %d portals != MAXPORTAL %d: %w", len(req.Portals), MaxPortals,
			ErrInvalidGameState)
	}

	return nil
}

// composeDungeonRecords writes the dungeon-only record blocks
// (loadsave.cpp:2845-2882): monster ids and records, the missile index tables
// and records, object ids and records, the light block and the vision block.
func composeDungeonRecords(out *recordWriter, req SaveGameDataRequest) {
	if req.Header.LevelType == 0 { // DTYPE_TOWN
		return
	}

	for i, id := range req.ActiveMonsters {
		if i >= MaxMonsters {
			break
		}

		out.be32(id)
	}

	out.zeroes(4 * (MaxMonsters - min(len(req.ActiveMonsters), MaxMonsters)))

	for _, record := range req.MonsterRecords {
		out.bytes(record)
	}

	writeMissileIndexTables(out, len(req.MissileRecords))

	for _, record := range req.MissileRecords {
		out.bytes(record)
	}

	writeFixedIDTable(out, req.Dungeon.ActiveObjectIDs, MaxObjects)
	writeFixedIDTable(out, req.Dungeon.AvailableObjectIDs, MaxObjects)

	for i := range req.ActiveObjects {
		out.bytes(req.ActiveObjects[i].Encode())
	}

	writeLightBlock(out, req)
	writeVisionBlock(out, req)
}

// int8ToByte keeps the two's-complement image of a signed id byte.
func int8ToByte(value int8) byte {
	return uint8(value)
}

// writeFixedIDTable writes an int8 id table over an upstream fixed array:
// the ids in order, then zero padding to width (loadsave.cpp:2870-2873).
func writeFixedIDTable(out *recordWriter, ids []int8, width int) {
	for _, id := range ids {
		out.u8(uint8(int8ToByte(id)))
	}

	out.zeroes(width - min(len(ids), width))
}

// writeFixedLightIDTable writes the uint8 light id table over the fixed
// MAXLIGHTS array: the ids in order, then zero padding to width
// (loadsave.cpp:2874-2875, lighting.h MAXLIGHTS).
func writeFixedLightIDTable(out *recordWriter, ids []uint8, width int) {
	for _, id := range ids {
		out.u8(id)
	}

	out.zeroes(width - min(len(ids), width))
}

// writeMissileIndexTables writes the active/available missile id tables and
// the saved-count skip (loadsave.cpp:2849-2857).
func writeMissileIndexTables(out *recordWriter, savedCount int) {
	for activeMissile := range MaxMissilesForSaveGame {
		out.u8(uint8(activeMissile))
	}

	for available := savedCount; available < MaxMissilesForSaveGame; available++ {
		out.u8(uint8(available))
	}

	out.zeroes(savedCount)
}

// writeLightBlock writes the active-light count, ids and records
// (loadsave.cpp:2869-2877).
func writeLightBlock(out *recordWriter, req SaveGameDataRequest) {
	out.be32(uint32(len(req.ActiveLights)))

	writeFixedLightIDTable(out, req.Dungeon.ActiveLightIDs, MaxLights)

	for i := range req.ActiveLights {
		out.bytes(req.ActiveLights[i].Encode())
	}
}

// writeVisionBlock writes the VisionId/VisionCount pair and the per-player
// vision records with the vision flag set (loadsave.cpp:2878-2882).
func writeVisionBlock(out *recordWriter, req SaveGameDataRequest) {
	players := len(req.VisionLights)
	out.be32(uint32(players + 1))
	out.be32(uint32(players))

	for i := range req.VisionLights {
		out.bytes(req.VisionLights[i].EncodeVision())
	}
}

// composeCommonGrids writes the dropped items, unique item flags and the
// common grids, all sections town and dungeon share
// (loadsave.cpp:2866-2884).
func composeCommonGrids(out *recordWriter, req SaveGameDataRequest) error {
	items, err := req.DroppedItems.Encode(SaveItemSize(req.Mode))
	if err != nil {
		return fmt.Errorf("persist: compose dropped items: %w", err)
	}

	out.bytes(items)

	for _, flag := range req.UniqueItemFlag {
		out.u8(uint8(boolField{set: flag}.image()))
	}

	out.bytes(EncodeDLightGrid(req.DLight))
	out.bytes(EncodeDFlagsGrid(req.DFlags))
	out.bytes(EncodeDPlayerGrid(req.DPlayer))
	out.bytes(EncodeDroppedItemLocations(req.DItem))

	return nil
}

// composeTail writes the premium tail and the automap flags
// (loadsave.cpp:2921-2926).
func composeTail(out *recordWriter, req SaveGameDataRequest) error {
	tail, err := PremiumTailBlock(req.PremiumCount, req.PremiumLevel, req.PremiumItems, req.Mode)
	if err != nil {
		return fmt.Errorf("persist: compose premium tail: %w", err)
	}

	out.bytes(tail)
	out.u8(uint8(boolField{set: req.AutomapActive}.image()))
	out.be32(uint32(req.AutoMapScaleX))

	return nil
}
