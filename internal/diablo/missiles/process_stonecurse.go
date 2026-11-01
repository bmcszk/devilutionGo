// Package missiles: the Stone Curse bodies.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Monster modes this row switches on (monster.h:37-58).
const (
	monsterModeFadeIn    = 8
	monsterModeFadeOut   = 9
	monsterModeCharge    = 14
	monsterModePetrified = 15
)

// graphicStoneCurseShatter is MissileGraphicID::StoneCurseShatter — the 19th
// entry of an enum that starts at Arrow (tables/misdat.h:36,54).
const graphicStoneCurseShatter GraphicID = 18

// stoneCurseSteps is the 15-step cap of the petrified life (missiles.cpp:2414).
const stoneCurseSteps = 15

// stoneCurseShatterLife is the 11-tick shatter animation (missiles.cpp:3702).
const stoneCurseShatterLife = 11

// Monster type ordinals Stone Curse refuses (tables/monstdat.h:270-271,298).
// enum _monster_id starts at MT_NZOMBIE = 0, so the ordinal is the monstdat.tsv
// data row index; MT_NAKRUL is a hellfire row (hf-monstdat.tsv).
const (
	MonsterTypeGolem  = 109
	MonsterTypeDiablo = 110
	MonsterTypeNakrul = 137
)

// AddStoneCurse ports AddStoneCurse (missiles.cpp:2363-2416): the search takes
// the closest monster within 5 of the aim that is in bounds, not a golem or a
// boss, and not fading in, fading out or charging (cpp:2365-2387). A failed
// search fizzles the spell (cpp:2389-2393) and an already petrified monster is
// left alone (cpp:2399-2403); otherwise the caster's spell saves the monster's
// mode in var1 and its id in var2, petrifies it (cpp:2405-2407) and lives
// min(spellLevel + 6, 15) 16-tick steps (cpp:2410-2415).
func AddStoneCurse(w *HitWorld, m *Missile, p *AddParameter) {
	spot, monster, ok := stoneCurseTarget(w, p.Dst)
	if !ok { // cpp:2389-2393
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	if w.ProcessCallbacks.MonsterMode(monster) == monsterModePetrified { // cpp:2399-2403
		m.DelFlag = true

		return
	}

	m.Var1 = w.ProcessCallbacks.MonsterMode(monster)          // cpp:2405
	m.Var2 = monster                                          // cpp:2406
	petrifyMonster(w, monster)                                // cpp:2407
	m.Position.Tile = spot                                    // cpp:2410
	m.Position.Start = spot                                   // cpp:2411
	m.Duration = minInt(m.SpellLevel+6, stoneCurseSteps) << 4 // cpp:2412-2415
}

// ProcessStoneCurse ports ProcessStoneCurse (missiles.cpp:3694-3720): a monster
// that died while petrified swaps to the shatter frames for 11 ticks
// (cpp:3698-3703), a monster that is no longer petrified ends the missile
// (cpp:3704-3707), and the last tick either restores the saved mode and clears
// the petrified flag (cpp:3711-3713) or drops the petrified remains
// (cpp:3715).
func ProcessStoneCurse(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3696

	monster := m.Var2 // cpp:3697
	if switchToStoneCurseShatter(w, m, monster) {
		m.Duration = stoneCurseShatterLife // cpp:3702
	}

	if w.ProcessCallbacks.MonsterMode(monster) != monsterModePetrified { // cpp:3704-3707
		m.DelFlag = true

		return
	}

	if m.Duration == 0 { // cpp:3709-3717
		m.DelFlag = true
		stoneCurseExpire(w, m, monster)
	}

	if GraphicID(m.AnimType) == graphicStoneCurseShatter { // cpp:3718-3719
		PutMissile(w, m)
	}
}

// switchToStoneCurseShatter ports the shatter swap (missiles.cpp:3698-3703): a
// monster that died while petrified and has not shattered yet switches to the
// shatter sheet, drawn over the walls.
func switchToStoneCurseShatter(w *HitWorld, m *Missile, monster int) bool {
	if w.ProcessCallbacks.MonsterHitPoints(monster) != 0 { // cpp:3698
		return false
	}

	if GraphicID(m.AnimType) == graphicStoneCurseShatter { // cpp:3698
		return false
	}

	m.setDefaultFrameGroup()                 // cpp:3699
	m.DrawFlag = true                        // cpp:3700
	m.setAnimation(graphicStoneCurseShatter) // cpp:3701

	return true
}

// stoneCurseExpire ports the last-tick fork (missiles.cpp:3711-3716): a living
// monster keeps its saved mode with the petrified flag cleared; a dead one
// leaves the corpse AddCorpse(monster.position.tile, stonendx,
// monster.direction) drops (cpp:3715, dead.cpp:93-96).
func stoneCurseExpire(w *HitWorld, m *Missile, monster int) {
	if w.ProcessCallbacks.MonsterHitPoints(monster) > 0 { // cpp:3711
		w.ProcessCallbacks.SetMonsterMode(monster, m.Var1)     // cpp:3712
		w.ProcessCallbacks.SetMonsterPetrified(monster, false) // cpp:3713

		return
	}

	w.ProcessCallbacks.AddCorpse(monster) // cpp:3715
}

// stoneCurseTarget ports the FindClosestValidPosition search of the add body
// (missiles.cpp:2365-2387) and hands back the monster id it found. Upstream
// re-reads dMonster at the found tile (cpp:2396); the predicate's read is the
// same cell, so the id is carried out of the search.
func stoneCurseTarget(w *HitWorld, dst TilePosition) (TilePosition, int, bool) {
	found := -1
	posOk := func(candidate engine.Point) bool {
		monster, ok := stoneCurseTileMonster(w, candidate)
		if ok {
			found = monster
		}

		return ok
	}

	spot, ok := engine.FindClosestValidPosition(posOk, engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !ok || found < 0 { // cpp:2387-2393
		return TilePosition{}, 0, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, found, true
}

// stoneCurseTileMonster ports the per-tile half of the search
// (missiles.cpp:2366-2385).
func stoneCurseTileMonster(w *HitWorld, candidate engine.Point) (int, bool) {
	target := TilePosition{X: candidate.X, Y: candidate.Y}
	if !world.InDungeonBounds(target.X, target.Y) { // cpp:2367-2369
		return 0, false
	}

	monster := absInt(int(w.DMonster[target.Y][target.X])) - 1 // cpp:2371
	if monster < 0 {                                           // cpp:2372-2374
		return 0, false
	}

	if stoneCurseTypeBlocks(w, monster) { // cpp:2378-2380
		return 0, false
	}

	if stoneCurseModeBlocks(w.ProcessCallbacks.MonsterMode(monster)) { // cpp:2381-2383
		return 0, false
	}

	return monster, true
}

// stoneCurseTypeBlocks ports the IsAnyOf(monster.type().type, MT_GOLEM,
// MT_DIABLO, MT_NAKRUL) refusal (missiles.cpp:2378).
func stoneCurseTypeBlocks(w *HitWorld, monster int) bool {
	switch w.AddSource.monsterStats(monster).Type {
	case MonsterTypeGolem, MonsterTypeDiablo, MonsterTypeNakrul:
		return true
	}

	return false
}

// stoneCurseModeBlocks reports the modes the add body refuses
// (missiles.cpp:2381).
func stoneCurseModeBlocks(mode int) bool {
	return mode == monsterModeFadeIn || mode == monsterModeFadeOut || mode == monsterModeCharge
}

// petrifyMonster ports Monster::petrify (monster.cpp:4925-4929): the mode and
// the animation flag, nothing else.
func petrifyMonster(w *HitWorld, monster int) {
	w.ProcessCallbacks.SetMonsterMode(monster, monsterModePetrified)
	w.ProcessCallbacks.SetMonsterPetrified(monster, true)
}
