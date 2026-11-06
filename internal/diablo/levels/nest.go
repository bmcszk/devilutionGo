package levels

import "errors"

// errNotImplemented marks a seam the hive port has not filled yet, so a RED
// parity gate reports the missing port instead of a layout mismatch.
var errNotImplemented = errors.New("hive generator not implemented")

// GenerateNest is the hive generator entry: CreateL3Dungeon's nest branch
// (levels/drlg_l3.cpp:2192-2200) with the `leveltype == DTYPE_NEST` variants of
// the pools (:1499-1501), stairs (:1982-1989) and isles (:2034-2042), then
// Pass3 (:2198) and PlaceHiveLights (:2182-2187). The port is the next unit
// against the 64 committed nest dPiece goldens.
func GenerateNest(_ *Dungeon, _ Options) error {
	return errNotImplemented
}
