package entities

import "github.com/bmcszk/devilutionGo/internal/assets"

// Corpse side constants.
const (
	// MaxCorpses is MaxCorpses (dead.h:17).
	MaxCorpses = 31
	// CorpseIndexMask selects the corpse id bits of a dCorpse byte
	// (dun_tile_data.hpp:123-125).
	CorpseIndexMask = 0x1F
	// CorpseDirectionShift is the ``>> 0x5`` of the dCorpse encoding
	// (dun_tile_data.hpp:124).
	CorpseDirectionShift = 5
	// Dungeon grid bounds (gendung_defs.hpp:8-9 MAXDUNX/MAXDUNY); entities has
	// no dependency on internal/game, so the map sizes live here too.
	MaxDunX = 112
	MaxDunY = 112
)

// AnimData ports struct AnimStruct (monster.h:157-173): one animation's sprite
// handle plus its frame count, rate and sprite width.
type AnimData struct {
	// Sprites is OptionalClxSpriteListOrSheet: nil is std::nullopt.
	Sprites *assets.ClxSpriteList
	Width   uint16
	Frames  int8
	Rate    int8
}

// Corpse ports struct Corpse (dead.h:19-35).
type Corpse struct {
	// Sprites is OptionalClxSpriteListOrSheet, shared not copied:
	// InitDeadAnimationFromMonster copies the handle only (dead.cpp:26-30).
	Sprites *assets.ClxSpriteList
	Frame   int
	// Width drives the render-side tile-centre offset (scrollrt.cpp:803).
	Width uint16
	// TranslationPaletteIndex: 0 means no unique-monster palette remap
	// (dead.cpp:61,74,83).
	TranslationPaletteIndex uint8
}

// CorpseRef is a decoded dCorpse entry: the 1-based corpse id and the
// direction bits (scrollrt.cpp:802-804).
type CorpseRef struct {
	Index     int8
	Direction int
}

// CorpseDraw is one resolved corpse draw: the direction's frame list, the
// death frame, the sprite width and the palette remap switch, all the data
// scrollrt.cpp:801-806 needs before pixels.
type CorpseDraw struct {
	Sprites                 []assets.ClxFrame
	Frame                   int
	Width                   uint16
	TranslationPaletteIndex uint8
}

// CorpseDrawer is the render-layer seam for one corpse draw call
// (scrollrt.cpp:801-806); the renderer owns position, palette and pixels.
type CorpseDrawer interface {
	DrawCorpse(tile Point, draw CorpseDraw, lightIndex int)
}

// Corpses ports the corpse state: the Corpse records (dead.cpp:19
// Corpses[MaxCorpses]) and the per-tile dCorpse occupancy map
// (dun_tile_data.hpp:125). Both are upstream globals; one value lets a level
// load reset them together.
type Corpses struct {
	// Records is indexed by corpse id - 1 (dead.cpp:60,82, scrollrt.cpp:802).
	Records [MaxCorpses]Corpse
	// Tiles is dCorpse stored [y][x] so the save loop's j-outer/i-inner order
	// (loadsave.cpp:1948-1950) walks it row-major. The value is
	// (corpse id & 0x1F) + (direction << 5) narrowed to int8.
	Tiles [MaxDunY][MaxDunX]int8
}

// AddCorpse ports AddCorpse (dead.cpp:93-96): an unconditional per-tile store,
// overwriting whatever corpse occupied the tile. Upstream has no corpse pool
// eviction here; the MaxCorpses bound is enforced by InitCorpses' assert
// (dead.cpp:90).
func (c *Corpses) AddCorpse(tile Point, corpseID int8, direction int) {
	c.Tiles[tile.Y][tile.X] = int8(int(corpseID&CorpseIndexMask) + (direction << CorpseDirectionShift))
}

// ClearCorpse zeroes a tile once its corpse is gone (monster.cpp:2227).
func (c *Corpses) ClearCorpse(tile Point) {
	c.Tiles[tile.Y][tile.X] = 0
}

// CorpseAt decodes the tile's corpse reference; ok is false for an empty tile
// (scrollrt.cpp:801: bDead != 0). The int8 shift keeps upstream's sign
// extension before the 3-bit mask (scrollrt.cpp:804).
func (c *Corpses) CorpseAt(tile Point) (ref CorpseRef, ok bool) {
	raw := c.Tiles[tile.Y][tile.X]
	if raw == 0 {
		return CorpseRef{}, false
	}

	return CorpseRef{
		Index:     raw & CorpseIndexMask,
		Direction: int(raw>>CorpseDirectionShift) & 7,
	}, true
}

// InitFromMonsterAnim ports InitDeadAnimationFromMonster (dead.cpp:23-33) in
// upstream line order: sprites (share or clear), then frame, then width.
func (c *Corpse) InitFromMonsterAnim(anim AnimData) {
	c.Sprites = anim.Sprites
	c.Frame = int(anim.Frames) - 1
	c.Width = anim.Width
}

// SpritesForDirection ports Corpse::spritesForDirection (dead.h:31-34): a sheet
// indexes one list per direction, a plain list serves every direction.
func (c *Corpse) SpritesForDirection(direction int) []assets.ClxFrame {
	if c.Sprites == nil {
		return nil
	}

	if c.Sprites.GroupSizes == nil {
		return c.Sprites.Frames
	}

	return c.Sprites.Group(direction)
}

// AddDeadMonsterCorpse ports the corpse store of MonsterDeath
// (monster.cpp:1518-1522): a unique monster drops its own CorpseID, every other
// monster the type's. typeCorpseID is CMonster.corpseId (monster.h:191) and is
// supplied by the caller, which owns the monster type table.
func (c *Corpses) AddDeadMonsterCorpse(mon *Monster, typeCorpseID int8) {
	corpseID := typeCorpseID
	if mon.UniqueType != UniqueMonsterNone {
		corpseID = mon.CorpseID
	}

	c.AddCorpse(mon.Position.Tile, corpseID, int(mon.Direction))
}

// DrawTileCorpse hands the tile's corpse to the renderer and reports whether a
// draw call was made (scrollrt.cpp:801-806). The LightsMax check belongs to the
// renderer, which owns the light table.
func (c *Corpses) DrawTileCorpse(drawer CorpseDrawer, tile Point, lightIndex int) bool {
	ref, ok := c.CorpseAt(tile)
	if !ok || ref.Index < 1 || int(ref.Index) > len(c.Records) {
		return false
	}

	corpse := c.Records[ref.Index-1]

	drawer.DrawCorpse(tile, CorpseDraw{
		Sprites:                 corpse.SpritesForDirection(ref.Direction),
		Frame:                   corpse.Frame,
		Width:                   corpse.Width,
		TranslationPaletteIndex: corpse.TranslationPaletteIndex,
	}, lightIndex)

	return true
}
