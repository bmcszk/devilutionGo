// Package missiles: the no-op callback defaults the world constructors start from.
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine/appfat"

// noopFindGolem is the default FindGolemForPlayer sink.
func noopFindGolem(int) bool { return false }

// noopKillGolem is the default KillGolem sink.
func noopKillGolem(int) {}

// noopSpawnGolem is the default SpawnGolem sink.
func noopSpawnGolem(int, TilePosition, int) {}

// noopRequestSpawnGolem is the default RequestSpawnGolem sink.
func noopRequestSpawnGolem(TilePosition) {}

// noopPlaceWallMissile is the default PlaceWall sink: no wall is created.
func noopPlaceWallMissile(TilePosition, TilePosition, Direction, MissileID, int, int, int, int) *Missile {
	return nil
}

// noopSpawnMissile is the default SpawnMissile sink.
func noopSpawnMissile(TilePosition, TilePosition, Direction, MissileID, int, int, int, int, *Missile) {
}

// noopPlayerAtPosition is the default PlayerAtPosition sink: nobody there.
func noopPlayerAtPosition(TilePosition) int { return -1 }

// noopMyPlayer is the default MyPlayer sink.
func noopMyPlayer() int { return -1 }

// noopRelocatePlayer is the default RelocatePlayer sink.
func noopRelocatePlayer(int, TilePosition) {}

// noopChangeLightXY is the default ChangeLightXY sink.
func noopChangeLightXY(int, TilePosition) {}

// noopChangeVisionXY is the default ChangeVisionXY sink.
func noopChangeVisionXY(int, TilePosition) {}

// noopSetViewPosition is the default SetViewPosition sink.
func noopSetViewPosition(TilePosition) {}

// noopPlayerFalse is the default "no" player predicate.
func noopPlayerFalse(int) bool { return false }

// noopIntSink is the default int sink.
func noopIntSink(int) {}

// applySharedNoops fills the callbacks every world shares but which are not
// part of the missile container itself: party, sfx, tile trans and the fatal
// reporter.
func applySharedNoops(cb *ProcessCallbacks) {
	cb.PlayerFutureTile = func(int) TilePosition { return TilePosition{} }
	cb.SetInfravision = func(int) {}
	cb.RecalcPlayerItems = func(int) {}
	cb.PlayerDirection = func(int) Direction { return South }
	cb.PlaySfxLoc = func(string, int, int) {}
	cb.TransAt = func(int, int) bool { return false }
	cb.CheckBlock = func(int, int, int, int) bool { return true }
	cb.Fatal = func(*appfat.FatalError) {}
}

// noopPlayerLife is the default PlayerLife sink.
func noopPlayerLife(int) PlayerLife { return PlayerLife{} }

// noopTrigNear is the default TrigNear sink: no trigger nearby.
func noopTrigNear(TilePosition) bool { return false }

// noopPlaceSink is the default position sink.
func noopPlaceSink(TilePosition) {}

// noopPortalReadyPlayer is the default PortalReadyPlayer sink.
func noopPortalReadyPlayer(int, TilePosition) bool { return false }

// noopVoid is the default zero-argument sink.
func noopVoid() {}

// applyPlayerNoops fills the player-state sinks the life and rage rows read.
func applyPlayerNoops(cb *ProcessCallbacks) {
	cb.PlayerLife = noopPlayerLife
	cb.ClassLifeRestore = func(int) int { return fixed26_6One }
	cb.SetPlayerLife = func(int, int, int) {}
	cb.RedrawHealth = noopVoid
	cb.PlayerSpellFlags = func(int) int { return 0 }
	cb.SetPlayerSpellFlags = func(int, int) {}
	cb.SetPlayerHitPoints = func(int, int) {}
	cb.DamagePlayer = func(int, int, int, int) {}
	cb.Say = func(int, int) {}
	cb.RedrawEverything = noopVoid
}
