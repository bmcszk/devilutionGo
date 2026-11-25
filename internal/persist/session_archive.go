package persist

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// SessionArchive is the production fill of the world.SessionArchive seam
// (world/session.go:87-97): the persist store's three blocks — the heroes row,
// the games row (kind "single") and the levels row with the seed tail — freeze
// and restore one session record per slot. This is the shape the M1 scripted
// session's test-local slotArchive promoted to production
// (internal/integration/session_stage_test.go:242-770).
//
// Deviations from the upstream layout (recorded in the bean's PoW):
//   - the level seed lives in a 4-byte LE tail of the level blob, not in the
//     save game's own level-seed array (loadsave.cpp:1907-1918) — no schema
//     column is invented; the column rides the next schema version;
//   - the hero pack is re-derived from the classdat pack source, because the
//     create flow persists the row's Pack blob as zeros today
//     (cmd/devilutionGo/selhero.go pickClass) — the encoded-pack follow-up
//     makes this row authoritative.
type SessionArchive struct {
	store *SQLiteStore
	packs world.PackSource
}

// NewSessionArchive builds the production SessionArchive over the store: packs
// serves the starting pack a load restores (the row's own Pack blob is zeros
// until the create flow encodes one, so the class ordinal decides). The
// concrete type returns so callers reach HasGame — the value still satisfies
// world.SessionArchive wherever the seam takes it.
func NewSessionArchive(store *SQLiteStore, packs world.PackSource) *SessionArchive {
	return &SessionArchive{store: store, packs: packs}
}

// SaveSlot freezes rec into its slot: the hero row keeps its own name and
// creation time (upstream pfile_write_hero writes the whole hero, pfile.cpp),
// the game heading lands under kind "single" and the level blob carries the
// seed tail (pfile.cpp SaveSlot, loadsave.cpp SaveLevel).
func (a *SessionArchive) SaveSlot(rec world.SessionRecord) error {
	if err := a.saveHero(rec); err != nil {
		return err
	}

	if err := a.saveGame(rec); err != nil {
		return err
	}

	return a.saveLevel(rec)
}

// LoadSlot reads one slot's session back: world.ErrNoSlot when the slot holds
// no hero row (errors.Is ErrHeroNotFound), the wrapped store failure otherwise
// — a broken store is never disguised as an empty slot.
func (a *SessionArchive) LoadSlot(slot int) (world.SessionRecord, error) {
	record, err := a.readSlot(slot)
	if err != nil {
		return world.SessionRecord{}, err
	}

	record.Slot = slot

	return record, nil
}

// HasGame reports whether the slot holds a frozen session: the "game" entry
// exists (upstream ArchiveContainsGame reads the archive's "game" entry,
// pfile.cpp:247-253). Mapped on LoadGame + ErrHeroNotFound — no extra query,
// no sqlc regen.
func (a *SessionArchive) HasGame(ctx context.Context, slot int, kind string) (bool, error) {
	_, err := a.store.LoadGame(ctx, slot, kind)
	if errors.Is(err, ErrHeroNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("persist session archive: has game: %w", err)
	}

	return true, nil
}

// readSlot decodes the three blocks of one slot into the session record
// (without the slot number itself, which LoadSlot stamps).
func (a *SessionArchive) readSlot(slot int) (world.SessionRecord, error) {
	ctx := context.Background()

	header, hero, err := a.readHeading(slot)
	if err != nil {
		return world.SessionRecord{}, err
	}

	level, err := a.store.LoadLevel(ctx, slot, int(header.CurrLevel))
	if err != nil {
		return world.SessionRecord{}, fmt.Errorf("persist session archive: load level: %w", err)
	}

	seed, err := decodeLevelSeed(level)
	if err != nil {
		return world.SessionRecord{}, err
	}

	pack, ok := a.packs.StartingPack(hero.ClassID)
	if !ok {
		return world.SessionRecord{}, fmt.Errorf(
			"persist session archive: hero class %d has no starting pack", hero.ClassID)
	}

	return world.SessionRecord{
		LevelSeed: seed,
		Level:     int(header.CurrLevel),
		LevelType: int(header.LevelType),
		Loaded:    header.CurrLevel != 0,
		Hero:      world.Tile{X: int(header.ViewX), Y: int(header.ViewY)},
		Pack:      pack,
	}, nil
}

// readHeading reads one slot's hero row and game heading: the two blocks a
// level read needs (the level's own number lives in the heading).
func (a *SessionArchive) readHeading(slot int) (GameHeader, HeroRecord, error) {
	ctx := context.Background()

	hero, err := a.store.LoadHero(ctx, slot)
	if err != nil {
		if errors.Is(err, ErrHeroNotFound) {
			return GameHeader{}, HeroRecord{}, fmt.Errorf("persist session archive: %w", world.ErrNoSlot)
		}

		return GameHeader{}, HeroRecord{}, fmt.Errorf("persist session archive: load hero: %w", err)
	}

	game, err := a.store.LoadGame(ctx, slot, string(KindSingle))
	if err != nil {
		return GameHeader{}, HeroRecord{}, fmt.Errorf("persist session archive: load game: %w", err)
	}

	header, _, err := DecodeGameHeader(game)
	if err != nil {
		return GameHeader{}, HeroRecord{}, fmt.Errorf("persist session archive: decode game header: %w", err)
	}

	return header, hero, nil
}

// saveHero writes the record's hero row keeping the existing row's name and
// creation time (the create flow wrote them; a save must not clobber them).
func (a *SessionArchive) saveHero(rec world.SessionRecord) error {
	ctx := context.Background()

	hero, err := a.store.LoadHero(ctx, rec.Slot)
	if err != nil {
		return fmt.Errorf("persist session archive: load hero: %w", err)
	}

	hero.ClassID = rec.Pack.Class
	hero.Spawn = true

	if err := a.store.SaveHero(ctx, hero); err != nil {
		return fmt.Errorf("persist session archive: save hero: %w", err)
	}

	return nil
}

// saveGame writes the record's game heading: the record's level identity and
// the hero tile a load drops the hero back onto.
func (a *SessionArchive) saveGame(rec world.SessionRecord) error {
	levelType, currLevel := uint32(0), uint32(0)
	if rec.Loaded {
		levelType, currLevel = uint32(rec.LevelType), uint32(rec.Level)
	}

	header, err := GameHeader{
		Spawn:     true,
		CurrLevel: currLevel, LevelType: levelType,
		ViewX: int32(rec.Hero.X), ViewY: int32(rec.Hero.Y),
	}.Encode()
	if err != nil {
		return fmt.Errorf("persist session archive: encode game header: %w", err)
	}

	if err := a.store.SaveGame(context.Background(), rec.Slot, string(KindSingle), header); err != nil {
		return fmt.Errorf("persist session archive: save game: %w", err)
	}

	return nil
}

// saveLevel writes the record's level blob: the counter header, the grids the
// hero lit, then the 4-byte LE seed tail. Upstream keeps the seed in the save
// game's own level-seed array (loadsave.cpp:1907-1918); the block keeps it
// here instead of inventing a schema column (deviation 2 of the bean's PoW).
func (a *SessionArchive) saveLevel(rec world.SessionRecord) error {
	header := LevelHeader{ActiveItems: 1}
	grids := TileGrids{}

	index := rec.Hero.Y*DungeonWidth + rec.Hero.X
	grids.Lights[index] = 15
	grids.Automap[rec.Hero.Y%AutomapHeight*AutomapWidth+rec.Hero.X%AutomapWidth] = 1

	block := binary.LittleEndian.AppendUint32(append(header.Encode(), grids.Encode()...), rec.LevelSeed)

	if err := a.store.SaveLevel(context.Background(), rec.Slot, rec.Level, block); err != nil {
		return fmt.Errorf("persist session archive: save level: %w", err)
	}

	return nil
}

// decodeLevelSeed reads one level blob's seed tail, the 4-byte LE word the
// save wrote after the header and the grids.
func decodeLevelSeed(data []byte) (uint32, error) {
	if len(data) < LevelHeaderSize+4 {
		return 0, fmt.Errorf("persist session archive: level block = %d bytes, want at least %d",
			len(data), LevelHeaderSize+4)
	}

	if _, err := DecodeLevelHeader(data); err != nil {
		return 0, fmt.Errorf("persist session archive: decode level header: %w", err)
	}

	return binary.LittleEndian.Uint32(data[len(data)-4:]), nil
}
