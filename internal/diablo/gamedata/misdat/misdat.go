// Package misdat loads the missile data tables (misdat.tsv,
// missile_sprites.tsv) and holds the Go registry replacing the C++
// addFn/processFn function-pointer tables.
package misdat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// ParseMisdat returns one MissileData per tsv row, in file order.
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

// parseMissileSprites returns one MissileSprite per tsv row;
// frameDelay/frameLength are 16-wide comma lists.
func parseMissileSprites(data []byte) ([]MissileSprite, error) {
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

// addFn is the addFn behaviour of a missile row.
type addFn func(w World, self *missileRef)

// processFn is the processFn behaviour of a missile row.
type processFn func(w World, self *missileRef)

// World is the surface missile behaviours act on.
type World interface {
	SpawnMissile(kind string, x, y int) *missileRef
	Delete(m *missileRef)
}

// missileRef identifies one live missile in the world.
type missileRef struct {
	Kind    string
	X, Y    int
	Deleted bool
}

// AddRegistry maps misdat addFn column values to implementations; hf-only
// names live in HFAddRegistry.
var AddRegistry = map[string]addFn{}

// ProcessRegistry maps misdat processFn column values to implementations.
var ProcessRegistry = map[string]processFn{}

// addOpenNest spawns four BigExplosion missiles, then self-deletes.
func addOpenNest(w World, self *missileRef) {
	for _, p := range [][2]int{{80, 62}, {81, 62}, {80, 63}, {81, 63}} {
		w.SpawnMissile("BigExplosion", p[0], p[1])
	}
	w.Delete(self)
	self.Deleted = true
}

// dedupTable is the shared-array view toIndex builds; consecutive equal
// entries collapse to one slot.
type dedupTable struct {
	values  []int
	indexes []int
}

// Values is the run-collapsed shared frame table.
func (d dedupTable) Values() []int { return d.values }

// Indexes maps each frame to its slot in Values.
func (d dedupTable) Indexes() []int { return d.indexes }

// toIndex collapses consecutive equal values into shared slots.
func toIndex(v []int) dedupTable {
	var d dedupTable
	d.indexes = make([]int, len(v))
	for i, x := range v {
		if len(d.values) == 0 || d.values[len(d.values)-1] != x {
			d.values = append(d.values, x)
		}
		d.indexes[i] = len(d.values) - 1
	}
	return d
}
