package levels

import "errors"

// errL2NotPorted keeps the catacombs gate red until the generator lands.
var errL2NotPorted = errors.New("catacombs generator not ported")

// GenerateL2 runs the catacombs generator: CreateL2Dungeon's GenerateLevel
// (levels/drlg_l2.cpp:2679-2829) for levels 5-8.
func GenerateL2(_ *Dungeon, _ Options) error {
	return errL2NotPorted
}
