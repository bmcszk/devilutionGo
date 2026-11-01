// Package missiles_test: the shared Add* fixture. Every add body reads the
// same AddMissile seeding (missiles.cpp:2810-2838) plus a slice of the
// Players/Monsters globals, so one script struct covers them all.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// addScript is the scripted input of one Add* run: the AddMissile argument
// pack, the missile fields AddMissile seeds, and the injected stat reads.
type addScript struct {
	start  [2]int
	dst    [2]int
	midir  missiles.Direction
	caster int
	source int
	// spllvl is _mispllvl, dam is the pre-rolled _midam.
	spllvl int
	dam    int
	// magic is the source player's _pMagic.
	magic int
	// monsterMin/monsterMax are the source monster's damage range.
	monsterMin int
	monsterMax int
	// currlevel feeds ProjectileTrapDamage (missiles.cpp:266-269).
	currlevel int
	// light is the AddLight id the sink hands back.
	light    int
	hellfire bool
	rnd      []int
}

// newAddFixture wires the scripted globals into a HitWorld and seeds the
// missile exactly as AddMissile does before calling the addFn
// (missiles.cpp:2810-2838).
func newAddFixture(t *testing.T, s addScript) (*missiles.HitWorld, *missiles.Missile, *missiles.AddParameter) {
	t.Helper()

	w := &missiles.HitWorld{
		IsHellfire:       s.hellfire,
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Magic: s.magic}
			},
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{
					Intelligence: 5,
					MinDamage:    s.monsterMin,
					MaxDamage:    s.monsterMax,
				}
			},
		},
	}
	w.ProcessCallbacks.Currlevel = s.currlevel
	w.ProcessCallbacks.AddLight = func(int, int, uint8) int { return s.light }
	w.ProcessCallbacks.GenerateRnd = drawTape(t, s.rnd)

	m := &missiles.Missile{
		Type:       missiles.Firebolt,
		Caster:     s.caster,
		Source:     s.source,
		SpellLevel: s.spllvl,
		Dam:        s.dam,
		Lid:        missiles.NoLight,
	}
	m.SetTraveledFromTile(s.start[0], s.start[1])

	p := &missiles.AddParameter{
		Dst:   missiles.TilePosition{X: s.dst[0], Y: s.dst[1]},
		Midir: s.midir,
	}

	return w, m, p
}
