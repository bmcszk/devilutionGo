// Package monsterspawn holds the monster pack placement algorithm as pure,
// injectable code: upstream's PlaceGroup and its spawn-availability gate,
// ported from devilutionX monster.cpp at commit e1ad0c2b6.
//
// The port owns PlaceGroup's OWN random draws only. InitMonster's draws (the
// animation tick/frame rolls, the hit-point range roll and the two seed
// advances per placed monster, monster.cpp:961-984) belong to the deferred
// Monster entity port, so a roll tape can only be compared against a C++ call
// stream once that lands. Until then the oracles are scripted tapes with
// hand-derived expectations.
//
// Two upstream behaviours are reproduced VERBATIM rather than fixed:
//
//   - The scatter step adds the rolled direction's X delta to BOTH axes
//     (monster.cpp:341, `/// BUGFIX: yp += Point.y`). A "fixed" port would
//     scatter diagonally instead of on the vanilla cross pattern, which is a
//     parity defect, not an improvement.
//   - dMonster occupancy is tracked internally instead of in the global grid,
//     because the caller applies the returned placement in one step. The set of
//     occupied tiles, and therefore every CanPlaceMonster answer, is identical.
//
// Upstream's `num` is a size_t: `totalmonsters - ActiveMonsterCount` wraps when
// the level already holds more monsters than the budget. Real callers clamp
// before calling (totalmonsters >= ActiveMonsterCount by construction), so the
// wrap is documented here and not reproduced: this port returns zero placements
// for the non-wrapping boundary total == active.
package monsterspawn
