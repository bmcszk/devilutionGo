// Package misdat ports the missile data tables: misdat.tsv
// (Source/tables/misdat.cpp:373-396 LoadMissileData) and missile_sprites.tsv
// (misdat.cpp:134 MissileFileData), plus the Go registry replacing the C++
// addFn/processFn function-pointer tables.
package misdat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MissileData is one misdat.tsv row (misdat.h:166-196).
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

// MissileSprite is one missile_sprites.tsv row (MissileFileData).
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

func splitTSV(data []byte) ([][]string, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var rows [][]string
	var header []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		cols := strings.Split(ln, "\t")
		if header == nil {
			header = cols
			continue
		}
		if len(cols) != len(header) {
			return nil, fmt.Errorf("misdat: row has %d cols, want %d: %q", len(cols), len(header), ln)
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("misdat: empty table")
	}
	return rows, nil
}

// ParseMisdat parses misdat.tsv (id|addFn|processFn|castSound|hitSound|
// graphic|flags|movementDistribution) in row order.
func ParseMisdat(data []byte) ([]MissileData, error) {
	rows, err := splitTSV(data)
	if err != nil {
		return nil, err
	}
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

// ParseMissileSprites parses missile_sprites.tsv. frameDelay/frameLength are
// 16-wide comma lists (misdat.cpp:47-55 de-dups them at load).
func ParseMissileSprites(data []byte) ([]MissileSprite, error) {
	rows, err := splitTSV(data)
	if err != nil {
		return nil, err
	}
	out := make([]MissileSprite, 0, len(rows))
	for _, r := range rows {
		s := MissileSprite{ID: r[0], Name: r[3], Flags: r[5]}
		var err error
		if s.Width, err = strconv.Atoi(r[1]); err != nil {
			return nil, fmt.Errorf("misdat: %s width: %w", s.ID, err)
		}
		if s.Width2, err = strconv.Atoi(r[2]); err != nil {
			return nil, fmt.Errorf("misdat: %s width2: %w", s.ID, err)
		}
		if s.NumFrames, err = strconv.Atoi(r[4]); err != nil {
			return nil, fmt.Errorf("misdat: %s numFrames: %w", s.ID, err)
		}
		if s.FrameDelay, err = parse16(r[6]); err != nil {
			return nil, fmt.Errorf("misdat: %s frameDelay: %w", s.ID, err)
		}
		if s.FrameLength, err = parse16(r[7]); err != nil {
			return nil, fmt.Errorf("misdat: %s frameLength: %w", s.ID, err)
		}
		out = append(out, s)
	}
	return out, nil
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

// AddFn ports one Add* missile behaviour (missiles.cpp addFn dispatch region).
type AddFn func(w World, self *MissileRef)

// ProcessFn ports one Process* behaviour.
type ProcessFn func(w World, self *MissileRef)

// World is the behaviour port surface (stub in this bean; the real world
// lands with devilGo-mwlm).
type World interface {
	SpawnMissile(kind string, x, y int) *MissileRef
	Delete(m *MissileRef)
}

// MissileRef identifies one live missile in the world.
type MissileRef struct {
	Kind    string
	X, Y    int
	Deleted bool
}

// AddRegistry maps misdat addFn column values to implementations (base
// name set; hf-only names live in HFAddRegistry).
var AddRegistry = map[string]AddFn{}

// ProcessRegistry maps misdat processFn column values to implementations.
// Populated as behaviours port; the base 37 names resolve by the
// completeness test.
var ProcessRegistry = map[string]ProcessFn{}

// AddOpenNest ports missiles.cpp:1238-1246: spawns 4 BigExplosion missiles
// at (80,62) (81,62) (80,63) (81,63), then self-deletes.
func AddOpenNest(w World, self *MissileRef) {
	for _, p := range [][2]int{{80, 62}, {81, 62}, {80, 63}, {81, 63}} {
		w.SpawnMissile("BigExplosion", p[0], p[1])
	}
	w.Delete(self)
	self.Deleted = true
}

// DedupTable is the shared-array view ToIndex builds (misdat.cpp:47-55):
// consecutive equal entries collapse to one slot; the index list preserves
// per-frame lookup.
type DedupTable struct {
	values  []int
	indexes []int
}

// Values returns the de-duplicated shared table.
func (d DedupTable) Values() []int { return d.values }

// Indexes returns per-frame indices into Values.
func (d DedupTable) Indexes() []int { return d.indexes }

// ToIndex de-dups a 16-wide frame array into shared values + indices,
// matching ToIndex's assignment order (misdat.cpp:47-55).
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
