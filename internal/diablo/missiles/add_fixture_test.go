// Package missiles_test: the shared Add* fixture. Every add body reads the
// same AddMissile seeding (missiles.cpp:2810-2838) plus a slice of the
// Players/Monsters globals, so one script struct covers them all.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// rndTape scripts the GenerateRnd draws of one add body: the reads must match
// the scripted order, and every scripted value must be consumed — the seed
// parity gate of the ported bodies (missiles.cpp draw order).
type rndTape struct {
	t    *testing.T
	tape []int
	used int
}

// addFixture is one scripted add-body run: the world, the seeded missile, the
// AddParameter and the Rnd tape that must be fully consumed.
type addFixture struct {
	world   *missiles.HitWorld
	missile *missiles.Missile
	param   *missiles.AddParameter
	tape    *rndTape
	// light captures the AddLight call the body made.
	light *lightCall
}

// lightCall records one AddLight request so tests can pin its radius.
type lightCall struct {
	x      int
	y      int
	radius uint8
}

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
	// level is the source player's character level.
	level int
	// mfnum seeds the missile's frame group before the add body runs.
	mfnum int
	// tile is the source player's position.tile (missiles.cpp:1950).
	tile [2]int
	// animLen seeds _miAnimLen (the row's per-group frame length, misdat.cpp:400-408).
	animLen int
	// monsterMin/monsterMax are the source monster's damage range.
	monsterMin int
	monsterMax int
	// monsterType is the source monster's type id (tables/monstdat.h).
	monsterType int
	// currlevel feeds ProjectileTrapDamage (missiles.cpp:266-269).
	currlevel int
	// light is the AddLight id the sink hands back.
	light    int
	hellfire bool
	rnd      []int
}

// draw hands back the next scripted value; an extra draw fails.
func (r *rndTape) draw(int) int {
	r.t.Helper()

	if r.used >= len(r.tape) {
		r.t.Fatalf("unexpected GenerateRnd draw %d", r.used)
	}

	v := r.tape[r.used]
	r.used++

	return v
}

// assertConsumed fails when the body drew fewer times than the C++ order says.
func (r *rndTape) assertConsumed() {
	r.t.Helper()

	if r.used != len(r.tape) {
		r.t.Fatalf("consumed %d of %d scripted GenerateRnd draws", r.used, len(r.tape))
	}
}

// newAddFixture wires the scripted globals into a HitWorld and seeds the
// missile exactly as AddMissile does before calling the addFn
// (missiles.cpp:2810-2838).
func newAddFixture(t *testing.T, s addScript) *addFixture {
	t.Helper()

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		IsHellfire:       s.hellfire,
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{
					Magic: s.magic,
					Level: s.level,
					Tile:  missiles.TilePosition{X: s.tile[0], Y: s.tile[1]},
				}
			},
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{
					Type:         s.monsterType,
					Intelligence: 5,
					MinDamage:    s.monsterMin,
					MaxDamage:    s.monsterMax,
				}
			},
		},
	}
	w.ProcessCallbacks.Currlevel = s.currlevel
	light := &lightCall{}
	w.ProcessCallbacks.AddLight = func(x, y int, radius uint8) int {
		*light = lightCall{x: x, y: y, radius: radius}

		return s.light
	}
	tape := &rndTape{t: t, tape: s.rnd}
	w.ProcessCallbacks.GenerateRnd = tape.draw

	m := &missiles.Missile{
		Type:       missiles.Firebolt,
		Caster:     s.caster,
		Source:     s.source,
		SpellLevel: s.spllvl,
		Dam:        s.dam,
		MFNum:      s.mfnum,
		AnimLen:    s.animLen,
		Lid:        missiles.NoLight,
	}
	m.SetTraveledFromTile(s.start[0], s.start[1])

	p := &missiles.AddParameter{
		Dst:   missiles.TilePosition{X: s.dst[0], Y: s.dst[1]},
		Midir: s.midir,
	}

	return &addFixture{world: w, missile: m, param: p, tape: tape, light: light}
}
