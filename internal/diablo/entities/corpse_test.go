package entities_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
)

type drawCall struct {
	tile       entities.Point
	draw       entities.CorpseDraw
	lightIndex int
}

type recordingDrawer struct {
	calls []drawCall
}

type corpseFixtureLine struct {
	number int
	text   string
}

type corpseFixtureRow struct {
	tile      entities.Point
	corpseID  int8
	direction int
	want      int8
}

// TestCorpses_TilePacking pins the dCorpse encoding byte for byte:
// value = (corpseId & 0x1F) + (direction << 5) stored in an int8_t, so index
// and direction bits never overlap (dead.cpp:95, dun_tile_data.hpp:123-125).
func TestCorpses_TilePacking(t *testing.T) {
	tests := []struct {
		name      string
		tile      entities.Point
		corpseID  int8
		direction int
		want      int8
	}{
		{name: "index 1 south", tile: entities.Point{X: 11, Y: 7}, corpseID: 1, direction: 0, want: 1},
		{name: "index 7 north west", tile: entities.Point{X: 40, Y: 3}, corpseID: 7, direction: 3, want: wantByte(7, 3)},
		{name: "index 31 south wraps signed", tile: entities.Point{}, corpseID: 31, direction: 4, want: wantByte(31, 4)},
		{name: "index above mask", tile: entities.Point{X: 1, Y: 2}, corpseID: 33, direction: 7, want: wantByte(1, 7)},
		{name: "negative direction", tile: entities.Point{X: 5, Y: 5}, corpseID: 2, direction: -1, want: wantByte(2, -1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given an empty occupancy map
			var corpses entities.Corpses

			// when the corpse lands on the tile
			corpses.AddCorpse(tt.tile, tt.corpseID, tt.direction)

			// then the raw tile byte matches the upstream packing
			if got := corpses.Tiles[tt.tile.Y][tt.tile.X]; got != tt.want {
				t.Fatalf("dCorpse[%d][%d] = %d, want %d", tt.tile.X, tt.tile.Y, got, tt.want)
			}
		})
	}
}

// TestCorpses_CorpseAtDecodesReference pins the read side: the 1-based corpse
// id sits in bits 0-4, the direction in bits 5-7 (scrollrt.cpp:802-804).
func TestCorpses_CorpseAtDecodesReference(t *testing.T) {
	// given a corpse packed at direction 5 with corpse id 9
	var corpses entities.Corpses
	corpses.AddCorpse(entities.Point{X: 20, Y: 21}, 9, 5)

	// when the tile is read back
	got, ok := corpses.CorpseAt(entities.Point{X: 20, Y: 21})

	// then the decoded reference matches the stored values
	if !ok {
		t.Fatal("CorpseAt reported an empty tile")
	}

	if got.Index != 9 || got.Direction != 5 {
		t.Fatalf("CorpseAt = {Index:%d Direction:%d}, want {9 5}", got.Index, got.Direction)
	}
}

// TestCorpses_CorpseAtEmptyTile pins the sentinel: a zero byte is no corpse,
// it never decodes to corpse index 0 (scrollrt.cpp:801).
func TestCorpses_CorpseAtEmptyTile(t *testing.T) {
	// given an occupancy map with nothing added
	var corpses entities.Corpses

	// when an untouched tile is read
	_, ok := corpses.CorpseAt(entities.Point{X: 3, Y: 4})

	// then the tile reports empty
	if ok {
		t.Fatal("CorpseAt reported a corpse on an empty tile")
	}
}

// TestCorpses_AddCorpseReusesOnlyItsOwnTile pins the reuse policy: AddCorpse
// overwrites the tile it lands on and nothing else; there is no corpse pool
// eviction (dead.cpp:93-96 stores one byte per tile).
func TestCorpses_AddCorpseReusesOnlyItsOwnTile(t *testing.T) {
	// given two corpses on neighbouring tiles
	var corpses entities.Corpses
	corpses.AddCorpse(entities.Point{X: 30, Y: 30}, 3, 1)
	corpses.AddCorpse(entities.Point{X: 31, Y: 30}, 4, 2)

	// when a new corpse lands on the first tile
	corpses.AddCorpse(entities.Point{X: 30, Y: 30}, 8, 6)

	// then that tile is replaced and the neighbour is untouched
	replaced, ok := corpses.CorpseAt(entities.Point{X: 30, Y: 30})
	if !ok || replaced.Index != 8 || replaced.Direction != 6 {
		t.Fatalf("replaced tile = {Index:%d Direction:%d ok:%v}, want {8 6 true}", replaced.Index, replaced.Direction, ok)
	}

	neighbour, ok := corpses.CorpseAt(entities.Point{X: 31, Y: 30})
	if !ok || neighbour.Index != 4 || neighbour.Direction != 2 {
		t.Fatalf("neighbour tile = {Index:%d Direction:%d ok:%v}, want {4 2 true}", neighbour.Index, neighbour.Direction, ok)
	}
}

// TestCorpses_ClearCorpseEmptiesTile covers the scavenger path that zeroes a
// tile once the corpse is eaten (monster.cpp:2227).
func TestCorpses_ClearCorpseEmptiesTile(t *testing.T) {
	// given a tile holding a corpse
	var corpses entities.Corpses
	corpses.AddCorpse(entities.Point{X: 12, Y: 13}, 5, 3)

	// when the corpse is cleared
	corpses.ClearCorpse(entities.Point{X: 12, Y: 13})

	// then the tile is empty again
	if _, ok := corpses.CorpseAt(entities.Point{X: 12, Y: 13}); ok {
		t.Fatal("ClearCorpse left a corpse on the tile")
	}
}

// TestCorpse_InitFromMonsterAnimSharesSprites pins
// InitDeadAnimationFromMonster (dead.cpp:23-33): the record holds the monster's
// sprite handle instead of a copy, and mirrors the last death frame + width.
func TestCorpse_InitFromMonsterAnimSharesSprites(t *testing.T) {
	// given monster death animation data
	sprites := listSprites()
	anim := entities.AnimData{Sprites: sprites, Width: 128, Frames: 6, Rate: 3}
	var corpse entities.Corpse

	// when the corpse record is initialised from it
	corpse.InitFromMonsterAnim(anim)

	// then the record shares the sprite handle and mirrors frame and width
	if corpse.Sprites != sprites {
		t.Fatal("InitFromMonsterAnim copied the sprite list instead of sharing the handle")
	}

	if corpse.Frame != 5 {
		t.Fatalf("Frame = %d, want 5 (frames-1)", corpse.Frame)
	}

	if corpse.Width != 128 {
		t.Fatalf("Width = %d, want 128", corpse.Width)
	}
}

// TestCorpse_InitFromMonsterAnimWithoutSprites pins the nullopt branch
// (dead.cpp:28-29): sprite-less animation data clears the record's handle.
func TestCorpse_InitFromMonsterAnimWithoutSprites(t *testing.T) {
	// given a corpse record that already holds a sprite handle
	corpse := entities.Corpse{Sprites: listSprites()}

	// when it is re-initialised from sprite-less animation data
	corpse.InitFromMonsterAnim(entities.AnimData{Width: 64, Frames: 4})

	// then the handle is cleared and the remaining fields still apply
	if corpse.Sprites != nil {
		t.Fatal("InitFromMonsterAnim kept a sprite handle for sprite-less animation data")
	}

	if corpse.Frame != 3 || corpse.Width != 64 {
		t.Fatalf("Frame/Width = %d/%d, want 3/64", corpse.Frame, corpse.Width)
	}
}

// TestCorpse_SpritesForDirectionPicksSheetOrList pins
// Corpse::spritesForDirection (dead.h:31-34): a sheet indexes per direction, a
// plain list is shared by every direction.
func TestCorpse_SpritesForDirectionPicksSheetOrList(t *testing.T) {
	// given one sheet corpse and one plain-list corpse
	sheetCorpse := entities.Corpse{Sprites: sheetSprites()}
	listCorpse := entities.Corpse{Sprites: listSprites()}

	// when each corpse resolves direction 3
	sheetFrames := sheetCorpse.SpritesForDirection(3)
	listFrames := listCorpse.SpritesForDirection(3)

	// then the sheet yields its third list and the list yields all frames
	if len(sheetFrames) != 2 {
		t.Fatalf("sheet frames = %d, want 2", len(sheetFrames))
	}

	if len(listFrames) != 9 {
		t.Fatalf("list frames = %d, want 9", len(listFrames))
	}
}

// TestCorpse_SpritesForDirectionOutOfRange pins the documented deviation: a
// direction past the sheet's list count is an assert-only index upstream
// (clx_sprite.hpp:238-247); here it yields nothing instead of reading OOB.
func TestCorpse_SpritesForDirectionOutOfRange(t *testing.T) {
	// given a sheet corpse with 8 direction lists
	corpse := entities.Corpse{Sprites: sheetSprites()}

	// when direction 9 is resolved
	frames := corpse.SpritesForDirection(9)

	// then nothing is returned
	if frames != nil {
		t.Fatalf("frames = %v, want nil", frames)
	}
}

// TestCorpse_SpritesForDirectionWithoutSprites asserts a nil handle resolves to
// nothing instead of panicking.
func TestCorpse_SpritesForDirectionWithoutSprites(t *testing.T) {
	// given a corpse record without sprite data
	var corpse entities.Corpse

	// when a direction is resolved
	frames := corpse.SpritesForDirection(0)

	// then there is nothing to draw
	if frames != nil {
		t.Fatalf("frames = %v, want nil", frames)
	}
}

// TestCorpses_DrawTileCorpseDispatchesDraw pins the render seam: the tile's
// corpse resolves to one draw call carrying width, frame and palette index
// (scrollrt.cpp:801-806).
func TestCorpses_DrawTileCorpseDispatchesDraw(t *testing.T) {
	// given a corpse on a tile with a monster palette remap
	var corpses entities.Corpses
	corpses.Records[0].InitFromMonsterAnim(entities.AnimData{Sprites: sheetSprites(), Width: 128, Frames: 6})
	corpses.Records[0].TranslationPaletteIndex = 3
	corpses.AddCorpse(entities.Point{X: 8, Y: 9}, 1, 2)
	drawer := &recordingDrawer{}

	// when the tile is rendered
	drawn := corpses.DrawTileCorpse(drawer, entities.Point{X: 8, Y: 9}, 12)

	// then exactly one draw call carries the resolved corpse
	if !drawn || len(drawer.calls) != 1 {
		t.Fatalf("draw calls = %d, drawn = %v, want 1 and true", len(drawer.calls), drawn)
	}

	call := drawer.calls[0]
	if call.draw.Width != 128 || call.draw.Frame != 5 || call.draw.TranslationPaletteIndex != 3 {
		t.Fatalf("draw = %+v, want width 128 frame 5 palette 3", call.draw)
	}

	if call.lightIndex != 12 {
		t.Fatalf("light index = %d, want 12", call.lightIndex)
	}
}

// TestCorpses_DrawTileCorpseSkipsEmptyTile asserts an empty tile produces no
// draw call (scrollrt.cpp:801: bDead != 0).
func TestCorpses_DrawTileCorpseSkipsEmptyTile(t *testing.T) {
	// given an empty occupancy map
	var corpses entities.Corpses
	drawer := &recordingDrawer{}

	// when the tile is rendered
	drawn := corpses.DrawTileCorpse(drawer, entities.Point{X: 1, Y: 1}, 12)

	// then nothing is drawn
	if drawn || len(drawer.calls) != 0 {
		t.Fatalf("draw calls = %d, drawn = %v, want 0 and false", len(drawer.calls), drawn)
	}
}

// TestCorpses_GobRoundTrip pins the save-facing layout: the occupancy map and
// the corpse records survive a gob round-trip unchanged, so the oqho serializer
// can write this shape as-is.
func TestCorpses_GobRoundTrip(t *testing.T) {
	// given an occupancy map with one corpse and one initialised record
	var corpses entities.Corpses
	corpses.AddCorpse(entities.Point{X: 50, Y: 60}, 2, 4)
	corpses.Records[1].InitFromMonsterAnim(entities.AnimData{Sprites: listSprites(), Width: 128, Frames: 6})

	// when it round-trips through gob
	var got entities.Corpses
	roundTripGob(t, &corpses, &got)

	// then the tile byte and the record fields survive
	assertRestored(t, &got)
}

// TestCorpses_JSONRoundTrip is the JSON half of the layout pin: field names
// stay stable for a future save/wire format.
func TestCorpses_JSONRoundTrip(t *testing.T) {
	// given an occupancy map with one corpse and one initialised record
	var corpses entities.Corpses
	corpses.AddCorpse(entities.Point{X: 50, Y: 60}, 2, 4)
	corpses.Records[1].InitFromMonsterAnim(entities.AnimData{Sprites: listSprites(), Width: 128, Frames: 6})

	// when it round-trips through JSON
	var got entities.Corpses
	roundTripJSON(t, &corpses, &got)

	// then the tile byte and the record fields survive
	assertRestored(t, &got)
}

// TestCorpses_AddCorpseFixtureParity replays the committed dCorpse byte fixture
// (hand-derived from dead.cpp:95) and asserts every raw tile byte.
func TestCorpses_AddCorpseFixtureParity(t *testing.T) {
	// given the committed occupancy fixture rows
	rows := loadCorpseFixture(t, "add-corpse.tsv")

	// when every corpse is placed on its tile
	var corpses entities.Corpses
	for _, row := range rows {
		corpses.AddCorpse(row.tile, row.corpseID, row.direction)
	}

	// then each raw tile byte and its decoded reference match upstream
	for _, row := range rows {
		got := corpses.Tiles[row.tile.Y][row.tile.X]
		if got != row.want {
			t.Fatalf("dCorpse[%d][%d] = %d, want %d", row.tile.X, row.tile.Y, got, row.want)
		}

		ref, ok := corpses.CorpseAt(row.tile)
		if !ok || ref.Index != row.corpseID&entities.CorpseIndexMask || ref.Direction != row.direction {
			t.Fatalf("decode at [%d][%d] = {Index:%d Direction:%d ok:%v}, want {%d %d true}",
				row.tile.X, row.tile.Y, ref.Index, ref.Direction, ok,
				row.corpseID&entities.CorpseIndexMask, row.direction)
		}
	}
}

// TestCorpses_AddDeadMonsterCorpse pins the corpse id choice of MonsterDeath
// (monster.cpp:1518-1522): a unique monster drops its own CorpseID, a normal one
// drops the monster type's.
func TestCorpses_AddDeadMonsterCorpse(t *testing.T) {
	tests := []struct {
		name       string
		uniqueType uint8
		monsterID  int8
		typeID     int8
		wantIndex  int8
	}{
		{name: "unique monster uses its own id", uniqueType: 3, monsterID: 12, typeID: 4, wantIndex: 12},
		{name: "normal monster uses type id", uniqueType: entities.UniqueMonsterNone, monsterID: 12, typeID: 4, wantIndex: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given a dying monster at tile 17,18 facing direction 6
			mon := entities.Monster{
				UniqueType: tt.uniqueType,
				CorpseID:   tt.monsterID,
				Direction:  6,
			}
			mon.Position.Tile = entities.Point{X: 17, Y: 18}
			var corpses entities.Corpses

			// when its death is recorded
			corpses.AddDeadMonsterCorpse(&mon, tt.typeID)

			// then the tile holds the chosen corpse id and the monster direction
			ref, ok := corpses.CorpseAt(entities.Point{X: 17, Y: 18})
			if !ok || ref.Index != tt.wantIndex || ref.Direction != 6 {
				t.Fatalf("corpse = {Index:%d Direction:%d ok:%v}, want {%d 6 true}",
					ref.Index, ref.Direction, ok, tt.wantIndex)
			}
		})
	}
}

// TestCorpses_FindCorpseTile pins the light-move scan (dead.cpp:37-45): a corpse
// id is searched x-major then y-major, and an absent id reports not-found.
func TestCorpses_FindCorpseTile(t *testing.T) {
	tests := []struct {
		name      string
		corpseID  int8
		want      entities.Point
		wantFound bool
	}{
		{name: "x-major order wins", corpseID: 5, want: entities.Point{X: 10, Y: 40}, wantFound: true},
		{name: "absent id", corpseID: 9, wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given corpse 5 on (10,40) and on the smaller-y tile (11,2)
			var corpses entities.Corpses
			corpses.AddCorpse(entities.Point{X: 10, Y: 40}, 5, 0)
			corpses.AddCorpse(entities.Point{X: 11, Y: 2}, 5, 0)

			// when the corpse tile is searched
			tile, ok := corpses.FindCorpseTile(tt.corpseID)

			// then the x-smaller tile is found, or nothing for an absent id
			if ok != tt.wantFound || tile != tt.want {
				t.Fatalf("FindCorpseTile(%d) = {%+v ok:%v}, want {%+v ok:%v}",
					tt.corpseID, tile, ok, tt.want, tt.wantFound)
			}
		})
	}
}

// TestCorpses_ResetClearsMap pins InitGlobals' dCorpse memset
// (gendung.cpp:298): every tile is emptied and the records survive, since
// InitCorpses rebuilds them separately (dead.cpp:49-91).
func TestCorpses_ResetClearsMap(t *testing.T) {
	// given two corpses and one initialised record
	var corpses entities.Corpses
	corpses.Records[0].InitFromMonsterAnim(entities.AnimData{Frames: 6, Width: 128})
	corpses.AddCorpse(entities.Point{X: 4, Y: 5}, 1, 2)
	corpses.AddCorpse(entities.Point{X: 6, Y: 7}, 2, 3)

	// when the map is reset for a new level
	corpses.Reset()

	// then both tiles are empty and the record is untouched
	for _, tile := range []entities.Point{{X: 4, Y: 5}, {X: 6, Y: 7}} {
		if _, ok := corpses.CorpseAt(tile); ok {
			t.Fatalf("Reset left a corpse on tile %+v", tile)
		}
	}

	if corpses.Records[0].Frame != 5 || corpses.Records[0].Width != 128 {
		t.Fatalf("Reset cleared the record: frame %d width %d, want 5/128",
			corpses.Records[0].Frame, corpses.Records[0].Width)
	}
}

// TestCorpses_RecordBound pins the corpse pool size (dead.h:17 MaxCorpses = 31),
// the bound InitCorpses asserts against (dead.cpp:90).
func TestCorpses_RecordBound(t *testing.T) {
	// given the empty corpse state
	var corpses entities.Corpses

	// when the pool size is read
	size := len(corpses.Records)

	// then it matches the upstream MaxCorpses bound
	if entities.MaxCorpses != 31 || size != 31 {
		t.Fatalf("MaxCorpses = %d, records = %d, want 31 and 31", entities.MaxCorpses, size)
	}
}

// sheetSprites is one CLX sheet: 8 direction lists of 2 frames each
// (OptionalClxSpriteListOrSheet sheet branch, dead.h spritesForDirection).
func sheetSprites() *assets.ClxSpriteList {
	sizes := make([]int, 8)
	for i := range sizes {
		sizes[i] = 2
	}

	return &assets.ClxSpriteList{Frames: make([]assets.ClxFrame, 16), GroupSizes: sizes}
}

// listSprites is one plain CLX list of 9 frames shared by every direction.
func listSprites() *assets.ClxSpriteList {
	return &assets.ClxSpriteList{Frames: make([]assets.ClxFrame, 9)}
}

// wantByte is the int8_t narrowing dead.cpp:95 performs on the stored byte.
func wantByte(corpseID int8, direction int) int8 {
	return int8(int(corpseID&0x1F) + (direction << 5))
}

func assertRestored(t *testing.T, got *entities.Corpses) {
	t.Helper()

	ref, ok := got.CorpseAt(entities.Point{X: 50, Y: 60})
	if !ok || ref.Index != 2 || ref.Direction != 4 {
		t.Fatalf("restored tile = {Index:%d Direction:%d ok:%v}, want {2 4 true}", ref.Index, ref.Direction, ok)
	}

	if got.Records[1].Frame != 5 || got.Records[1].Width != 128 {
		t.Fatalf("restored record = frame %d width %d, want 5/128", got.Records[1].Frame, got.Records[1].Width)
	}
}

func roundTripGob(t *testing.T, in, out *entities.Corpses) {
	t.Helper()

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(in); err != nil {
		t.Fatalf("gob encode: %v", err)
	}

	if err := gob.NewDecoder(&buf).Decode(out); err != nil {
		t.Fatalf("gob decode: %v", err)
	}
}

func roundTripJSON(t *testing.T, in, out *entities.Corpses) {
	t.Helper()

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
}

func (d *recordingDrawer) DrawCorpse(tile entities.Point, draw entities.CorpseDraw, lightIndex int) {
	d.calls = append(d.calls, drawCall{tile: tile, draw: draw, lightIndex: lightIndex})
}

// loadCorpseFixture reads a committed tab-separated corpse fixture.
func loadCorpseFixture(t *testing.T, name string) []corpseFixtureRow {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "diablo", "corpses", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed corpse fixture %s: %v", path, err)
	}

	lines := corpseFixtureLines(raw)
	rows := make([]corpseFixtureRow, 0, len(lines))

	for _, line := range lines {
		row, err := parseCorpseFixtureRow(line.text)
		if err != nil {
			t.Fatalf("parse corpse fixture %s line %d: %v", name, line.number, err)
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		t.Fatalf("corpse fixture %s has no data rows", name)
	}

	return rows
}

// corpseFixtureLines strips comments, blanks and the header row from a
// committed fixture, keeping each data line with its file line number.
func corpseFixtureLines(raw []byte) []corpseFixtureLine {
	var lines []corpseFixtureLine

	headerSeen := false

	for i, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !headerSeen {
			headerSeen = true
			continue
		}

		lines = append(lines, corpseFixtureLine{number: i + 1, text: line})
	}

	return lines
}

func parseCorpseFixtureRow(line string) (corpseFixtureRow, error) {
	fields := strings.Split(line, "\t")
	if len(fields) != 5 {
		return corpseFixtureRow{}, fmt.Errorf("want 5 fields, got %d", len(fields))
	}

	values := make([]int, len(fields))
	for i, field := range fields {
		v, err := strconv.Atoi(field)
		if err != nil {
			return corpseFixtureRow{}, fmt.Errorf("field %d: %w", i, err)
		}

		values[i] = v
	}

	return corpseFixtureRow{
		tile:      entities.Point{X: values[0], Y: values[1]},
		corpseID:  int8(values[2]),
		direction: values[3],
		want:      int8(values[4]),
	}, nil
}
