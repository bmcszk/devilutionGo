// Package misdat loads the missile data tables (misdat.tsv,
// missile_sprites.tsv) and holds the Go registry replacing the C++
// addFn/processFn function-pointer tables.
package misdat

import (
	"fmt"
	"strconv"
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// MissileData is one misdat.tsv row.
type MissileData struct {
	ID                   string
	AddFn                string
	ProcessFn            string
	CastSound            string
	HitSound             string
	Graphic              string
	Flags                []string
	MovementDistribution string
}

// MissileSprite is one missile_sprites.tsv row.
type MissileSprite struct {
	ID          string
	Width       int
	Width2      int
	Name        string
	NumFrames   int
	Flags       string
	FrameDelay  []int
	FrameLength []int
}

// ParseMisdat returns one MissileData per tsv row, in file order.
func ParseMisdat(data []byte) ([]MissileData, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	rows := tbl.Rows
	out := make([]MissileData, 0, len(rows))
	for _, r := range rows {
		out = append(out, MissileData{
			ID:                   r[0],
			AddFn:                r[1],
			ProcessFn:            r[2],
			CastSound:            r[3],
			HitSound:             r[4],
			Graphic:              r[5],
			Flags:                strings.Split(r[6], ","),
			MovementDistribution: r[7],
		})
	}
	return out, nil
}

// ParseMissileSprites returns one MissileSprite per missile_sprites.tsv row;
// frameDelay/frameLength are 16-wide comma lists (misdat.cpp sprite table).
func ParseMissileSprites(data []byte) ([]MissileSprite, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	rows := tbl.Rows
	out := make([]MissileSprite, 0, len(rows))
	for _, r := range rows {
		s, err := parseMissileSpriteRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// parseMissileSpriteRow parses a single tsv row into a MissileSprite.
// Mirrors the per-row parsing in parseMissileSprites (misdat.cpp).
func parseMissileSpriteRow(r []string) (MissileSprite, error) {
	s := MissileSprite{ID: r[0], Name: r[3], Flags: r[5]}
	var err error
	if s.Width, err = strconv.Atoi(r[1]); err != nil {
		return s, fmt.Errorf("misdat: %s width: %w", s.ID, err)
	}
	if s.Width2, err = strconv.Atoi(r[2]); err != nil {
		return s, fmt.Errorf("misdat: %s width2: %w", s.ID, err)
	}
	if s.NumFrames, err = strconv.Atoi(r[4]); err != nil {
		return s, fmt.Errorf("misdat: %s numFrames: %w", s.ID, err)
	}
	if s.FrameDelay, err = parse16(r[6]); err != nil {
		return s, fmt.Errorf("misdat: %s frameDelay: %w", s.ID, err)
	}
	if s.FrameLength, err = parse16(r[7]); err != nil {
		return s, fmt.Errorf("misdat: %s frameLength: %w", s.ID, err)
	}
	return s, nil
}

func parse16(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 16 {
		return nil, fmt.Errorf("want 16 entries, got %d", len(parts))
	}
	out := make([]int, 16)
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// AddFn is the addFn behaviour of a missile row (misdat.cpp add function).
type AddFn func(w World, self *MissileRef)

// ProcessFn is the processFn behaviour of a missile row (misdat.cpp process).
type ProcessFn func(w World, self *MissileRef)

// World is the surface missile behaviours act on (missiles.cpp global state,
// narrowed to the spawn/delete surface the misdat behaviours need).
type World interface {
	SpawnMissile(kind string, x, y int) *MissileRef
	Delete(m *MissileRef)
}

// MissileRef identifies one live missile in the world (MissileStruct identity).
type MissileRef struct {
	Kind    string
	X, Y    int
	Deleted bool
}

// AddRegistry maps misdat addFn column values to implementations; hf-only
// names live in HFAddRegistry.
// ProcessRegistry maps misdat processFn column values to implementations.
var (
	AddRegistry     = map[string]AddFn{}
	ProcessRegistry = map[string]ProcessFn{}
)

// AddOpenNest spawns four BigExplosion missiles, then self-deletes
// (Hive niche open, hellfire missions.cpp AddOpenNest).
func AddOpenNest(w World, self *MissileRef) {
	for _, p := range [][2]int{{80, 62}, {81, 62}, {80, 63}, {81, 63}} {
		w.SpawnMissile("BigExplosion", p[0], p[1])
	}
	w.Delete(self)
	self.Deleted = true
}

// DedupTable is the shared-array view ToIndex builds; consecutive equal
// entries collapse to one slot (sgFrames shared-frame-table layout).
type DedupTable struct {
	values  []int
	indexes []int
}

// Values is the run-collapsed shared frame table.
func (d DedupTable) Values() []int { return d.values }

// Indexes maps each frame to its slot in Values.
func (d DedupTable) Indexes() []int { return d.indexes }

// ToIndex collapses consecutive equal values into shared slots.
func ToIndex(v []int) DedupTable {
	var d DedupTable
	d.indexes = make([]int, len(v))
	for i, x := range v {
		if len(d.values) == 0 || d.values[len(d.values)-1] != x {
			d.values = append(d.values, x)
		}
		d.indexes[i] = len(d.values) - 1
	}
	return d
}
