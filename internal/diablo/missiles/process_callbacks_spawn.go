// Package missiles: the spawn sinks of ProcessCallbacks, split out to keep
// process.go within the file-length limit.
package missiles

// spawnHooks groups the missile-spawn sinks (missiles.cpp:747-750, 1236,
// 1385, 1698, 2823). ProcessCallbacks embeds it, so the fields stay reachable
// as cb.SpawnMissile and friends.
type spawnHooks struct {
	// SpawnMissile forwards to AddMissile (missiles.cpp:2823; params: src,
	// dst, dir, spawnID, caster, source, dam, spellLevel, parent — the
	// AddMissile argument pack).
	SpawnMissile func(src, dst TilePosition, dir Direction,
		spawnID MissileID, caster, source, dam, spellLevel int, parent *Missile)
	// SpawnMissileRow is the argument-pack form of SpawnMissile, used by the
	// add bodies that need the child row back to detect a fizzle and to
	// propagate the cast (missiles.cpp:1236, 1385).
	SpawnMissileRow func(args AddMissileArgs) *Missile
	// PlaceWallMissile ports PlaceWall (missiles.cpp:747-750), the wall spawn
	// that hands back the missile so the grower can nudge it: a segment at
	// position facing direction, cast by both sides.
	PlaceWallMissile func(src, dst TilePosition, dir Direction, spawnID MissileID,
		caster, source, damage, spellLevel int) *Missile
	// OtherMissiles returns the live rows AddSearch merges with
	// (missiles.cpp:1704).
	OtherMissiles func() []*Missile
	// IsNestLevel reports leveltype == DTYPE_NEST (missiles.cpp:1340).
	IsNestLevel func() bool
	// ShowAutoMapItems ports `AutoMapShowItems = true` (missiles.cpp:1698).
	ShowAutoMapItems func()
}
