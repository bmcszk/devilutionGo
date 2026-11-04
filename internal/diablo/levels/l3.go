package levels

import (
	"errors"
	"fmt"
)

// errCavesNotPorted marks the caves stages that are not ported yet.
var errCavesNotPorted = errors.New("caves generator: stage not ported")

// GenerateL3 runs the caves generator for the level d was created with
// (levels/drlg_l3.cpp:2185-2193, CreateL3Dungeon).
func GenerateL3(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	return fmt.Errorf("%w: entry %d", errCavesNotPorted, opts.Entry)
}
