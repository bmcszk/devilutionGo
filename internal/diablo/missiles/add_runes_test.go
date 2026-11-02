// Package missiles_test pins the Hellfire rune add bodies: AddRune and the
// RuneOfFire / RuneOfLight / RuneOfNova rows (missiles.cpp:668-693, 1250-1266).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

const (
	runeFree runeBlock = iota
	runeObject
	runeMissile
	runeSolid
	runeMissileBlocked
)

// runeBlock selects which tile predicate keeps the rune off its aim tile.
type runeBlock int

// runeCase pins one rune add row.
type runeCase struct {
	name   string
	script addScript
	// block makes the destination tile reject the rune search.
	block     runeBlock
	wantTile  [2]int
	wantVar1  int
	wantDel   bool
	wantDam   int
	wantLight bool
}

func TestAddRuneOfFire(t *testing.T) {
	// given
	cases := []runeCase{
		{
			name:      "clear line lands the rune on the aim tile",
			script:    addScript{start: [2]int{30, 30}, dst: [2]int{34, 30}, source: 0, light: 7},
			wantTile:  [2]int{34, 30},
			wantVar1:  int(missiles.BigExplosion),
			wantLight: true,
		},
		{
			name:      "an object on the aim tile pushes the rune one step south",
			script:    addScript{start: [2]int{30, 30}, dst: [2]int{34, 30}, source: 0, light: 7},
			block:     runeObject,
			wantTile:  [2]int{34, 31}, // crawl radius 1 starts at Displacement{0, 1}
			wantVar1:  int(missiles.BigExplosion),
			wantLight: true,
		},
		{
			name:      "a missile on the aim tile pushes the rune to the next free tile",
			script:    addScript{start: [2]int{30, 30}, dst: [2]int{34, 30}, source: 0, light: 7},
			block:     runeMissile,
			wantTile:  [2]int{34, 31},
			wantVar1:  int(missiles.BigExplosion),
			wantLight: true,
		},
		{
			name:      "a solid tile rejects the rune",
			script:    addScript{start: [2]int{30, 30}, dst: [2]int{34, 30}, source: 0, light: 7},
			block:     runeSolid,
			wantTile:  [2]int{34, 31}, // the search walks on to the neighbouring ring
			wantVar1:  int(missiles.BigExplosion),
			wantLight: true,
		},
		{
			name:      "a blocked missile line deletes the rune",
			script:    addScript{start: [2]int{30, 30}, dst: [2]int{34, 30}, source: 0, light: 7},
			block:     runeMissileBlocked,
			wantDel:   true,
			wantLight: false,
		},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newRuneFixture(t, tc)

			// when
			missiles.AddRuneOfFire(fx.world, fx.missile, fx.param)

			// then
			assertRune(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddRuneOfNova(t *testing.T) {
	// given
	fx := newRuneFixture(t, runeCase{
		script: addScript{start: [2]int{30, 30}, dst: [2]int{32, 30}, source: 0, light: 7},
	})

	// when
	missiles.AddRuneOfNova(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Var1 != int(missiles.Nova) { // cpp:1263-1266
		t.Fatalf("var1 = %d, want %d (Nova)", fx.missile.Var1, int(missiles.Nova))
	}

	if fx.missile.Position.Tile != (missiles.TilePosition{X: 32, Y: 30}) {
		t.Fatalf("tile = %v, want the aim tile", fx.missile.Position.Tile)
	}
}

func TestAddRuneOfLight(t *testing.T) {
	// given
	cases := []runeCase{
		{
			name: "player caster adds its level to the 2d10 roll",
			script: addScript{
				start: [2]int{30, 30}, dst: [2]int{33, 30}, source: 0, level: 5, light: 7,
				rnd: []int{3, 4},
			},
			wantTile:  [2]int{33, 30},
			wantVar1:  int(missiles.LightningWall),
			wantDam:   16 * (7 + 5 + 2),
			wantLight: true,
		},
		{
			name: "a trap caster contributes no level",
			script: addScript{
				start: [2]int{30, 30}, dst: [2]int{33, 30}, source: -1, caster: 0, light: 7,
				rnd: []int{3, 4},
			},
			wantTile:  [2]int{33, 30},
			wantVar1:  int(missiles.LightningWall),
			wantDam:   16 * (7 + 0 + 2),
			wantLight: true,
		},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runRuneOfLight(t, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// runRuneOfLight runs one AddRuneOfLight row.
func runRuneOfLight(t *testing.T, tc runeCase) {
	t.Helper()

	// given
	fx := newRuneFixture(t, tc)

	// when
	missiles.AddRuneOfLight(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	assertRune(t, fx, tc)
}

// newRuneFixture builds an add fixture whose tile predicates follow tc.block.
func newRuneFixture(t *testing.T, tc runeCase) *addFixture {
	t.Helper()

	fx := newAddFixture(t, tc.script)
	dst := missiles.TilePosition{X: tc.script.dst[0], Y: tc.script.dst[1]}

	switch tc.block {
	case runeObject:
		fx.world.ProcessCallbacks.IsObjectAtPosition = func(pos missiles.TilePosition) bool {
			return pos == dst
		}
	case runeMissile:
		fx.world.ProcessCallbacks.TileContainsMissile = func(pos missiles.TilePosition) bool {
			return pos == dst
		}
	case runeSolid:
		fx.world.Grid.DPiece[dst.Y][dst.X] = 1
		fx.world.Grid.SOLData[1] = world.TileSolid
	case runeMissileBlocked:
		// The walk tests the tiles after the launch tile, so the block sits
		// midway to the aim (monster.cpp:1633, 1666 skip startPoint).
		fx.world.Grid.DPiece[dst.Y][dst.X-2] = 2
		fx.world.Grid.SOLData[2] = world.TileBlockMissile
	default: // runeFree: every predicate accepts the aim tile
	}

	return fx
}

// assertRune checks the placement, var1, damage, light and delete flag of a
// rune row.
func assertRune(t *testing.T, fx *addFixture, tc runeCase) {
	t.Helper()

	if fx.missile.DelFlag != tc.wantDel {
		t.Fatalf("DelFlag = %v, want %v", fx.missile.DelFlag, tc.wantDel)
	}

	if tc.wantDam != 0 && fx.missile.Dam != tc.wantDam {
		t.Fatalf("dam = %d, want %d", fx.missile.Dam, tc.wantDam)
	}

	assertRuneLight(t, fx, tc)
	assertRunePlacement(t, fx, tc)
}

// assertRuneLight checks the 8-radius light a placed rune seeds, and that a
// rejected rune seeds none.
func assertRuneLight(t *testing.T, fx *addFixture, tc runeCase) {
	t.Helper()

	if !tc.wantLight {
		if fx.missile.Lid != missiles.NoLight || fx.light.radius != 0 {
			t.Fatalf("lid/radius = %d/%d, want no light seeded", fx.missile.Lid, fx.light.radius)
		}

		return
	}

	if fx.light.radius != 8 { // cpp:688
		t.Fatalf("light radius = %d, want 8", fx.light.radius)
	}

	if fx.missile.Lid != tc.script.light {
		t.Fatalf("lid = %d, want the AddLight id %d", fx.missile.Lid, tc.script.light)
	}
}

// assertRunePlacement checks where the rune landed and the missile id it
// remembers in var1.
func assertRunePlacement(t *testing.T, fx *addFixture, tc runeCase) {
	t.Helper()

	if tc.wantDel || tc.wantTile == [2]int{} {
		return
	}

	want := missiles.TilePosition{X: tc.wantTile[0], Y: tc.wantTile[1]}
	if fx.missile.Position.Tile != want {
		t.Fatalf("tile = %v, want %v", fx.missile.Position.Tile, want)
	}

	if fx.missile.Var1 != tc.wantVar1 {
		t.Fatalf("var1 = %d, want %d", fx.missile.Var1, tc.wantVar1)
	}
}
