// Package missiles_test ports test/missiles_test.cpp:1-64
// (TestArrowRotatesUniformly / TestAnimatedMissileRotatesUniformly and
// the TEST(Missiles, RotateBlockedMissileArrow) block) onto the rnd-tape
// parity harness: the upstream statistical 30%-70% histogram bounds are
// replayed against a FIXED seed through the repo's ported Borland LCG
// (internal/engine port of random.cpp:83-102), so every rotation draw is
// deterministic — no flaky statistical asserts (devilGo-mwlm slice 2).
package missiles_test

import "github.com/bmcszk/devilutionGo/internal/diablo/missiles"

import (
	"testing"
)

// rotationTape is GenerateRnd(2) x 600 with SetRndSeed(0x3A7B9E21),
// generated through the repo's engine RNG port (random.cpp:97-102:
// (AdvanceRndSeed() >> 16) % v for v <= 0x7FFF). Value 0 draws
// PickRandomlyAmong's index 0 (-1, left), value 1 draws index 1 (+1,
// right). Six 100-draw groups drive the six TestArrowRotatesUniformly /
// TestAnimatedMissileRotatesUniformly calls of the upstream TEST block
// in order (test/missiles_test.cpp:59-64). Regenerated from the ported
// LCG at slice-2 integration (an earlier hand-written tape diverged from
// the real sequence and carried literal ",," typos).
var rotationTape = [600]int{
	1, 0, 1, 1, 0, 1, 1, 0, 0, 0,
	1, 0, 1, 1, 0, 0, 0, 1, 1, 1,
	1, 0, 0, 0, 0, 1, 1, 0, 0, 0,
	0, 1, 0, 1, 0, 1, 0, 1, 1, 0,
	1, 0, 1, 0, 0, 0, 1, 0, 1, 1,
	0, 1, 0, 0, 1, 1, 0, 0, 1, 0,
	1, 1, 0, 1, 1, 1, 0, 0, 1, 1,
	0, 0, 1, 1, 0, 0, 1, 1, 0, 0,
	1, 0, 0, 0, 1, 1, 1, 0, 1, 0,
	1, 0, 0, 0, 1, 1, 1, 1, 1, 0,
	0, 1, 0, 1, 0, 0, 1, 1, 1, 0,
	1, 1, 0, 0, 0, 1, 0, 1, 0, 0,
	0, 0, 1, 0, 1, 1, 1, 0, 0, 1,
	1, 0, 0, 1, 1, 0, 1, 1, 0, 0,
	0, 1, 0, 1, 0, 1, 1, 0, 0, 1,
	0, 1, 0, 1, 1, 1, 1, 0, 1, 1,
	1, 0, 0, 1, 0, 1, 1, 1, 1, 0,
	0, 0, 0, 0, 1, 0, 1, 0, 0, 1,
	0, 0, 1, 0, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 1, 1, 1, 1, 0,
	1, 1, 1, 1, 1, 1, 0, 1, 0, 1,
	1, 0, 0, 0, 0, 0, 1, 0, 0, 0,
	1, 0, 1, 0, 1, 1, 1, 1, 1, 0,
	1, 0, 1, 1, 0, 1, 1, 1, 1, 0,
	1, 1, 1, 1, 1, 1, 0, 0, 1, 0,
	1, 1, 1, 1, 0, 1, 0, 1, 1, 0,
	0, 0, 1, 0, 0, 0, 1, 0, 1, 0,
	0, 1, 0, 0, 0, 0, 0, 1, 1, 0,
	0, 1, 0, 1, 0, 1, 0, 0, 1, 1,
	1, 1, 0, 0, 1, 0, 0, 0, 0, 0,
	0, 0, 1, 0, 0, 0, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 1, 0, 0, 1,
	1, 1, 1, 1, 1, 0, 0, 1, 1, 0,
	0, 1, 0, 1, 1, 0, 1, 1, 0, 0,
	0, 0, 0, 1, 1, 1, 0, 1, 1, 0,
	0, 0, 0, 1, 0, 1, 0, 1, 0, 0,
	1, 1, 0, 1, 0, 1, 1, 0, 1, 1,
	1, 1, 0, 0, 1, 0, 1, 0, 1, 0,
	1, 1, 0, 0, 1, 0, 1, 0, 1, 0,
	1, 1, 0, 1, 1, 0, 1, 0, 1, 1,
	1, 0, 1, 1, 1, 0, 1, 1, 0, 1,
	1, 1, 1, 0, 1, 0, 1, 0, 1, 1,
	1, 0, 0, 1, 0, 0, 0, 0, 0, 0,
	1, 1, 1, 1, 1, 0, 0, 0, 1, 1,
	0, 0, 1, 0, 0, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 0, 0, 1, 1, 0,
	1, 1, 0, 1, 0, 1, 1, 0, 1, 0,
	0, 0, 0, 0, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 0, 0, 1, 0, 0, 1, 1,
	1, 1, 0, 0, 0, 0, 1, 1, 0, 0,
	1, 1, 1, 0, 0, 1, 1, 1, 0, 1,
	0, 1, 0, 1, 1, 1, 0, 0, 0, 1,
	0, 1, 1, 1, 0, 1, 0, 0, 1, 0,
	1, 1, 0, 1, 0, 0, 0, 0, 1, 0,
	0, 1, 0, 0, 1, 0, 0, 1, 1, 0,
	1, 0, 1, 0, 0, 1, 0, 1, 0, 1,
	0, 1, 0, 1, 0, 1, 0, 1, 1, 0,
	1, 1, 1, 0, 1, 0, 0, 1, 0, 0,
	0, 0, 0, 0, 1, 1, 1, 0, 0, 1,
	0, 0, 0, 1, 1, 1, 1, 1, 0, 0,
}

// TestArrowRotatesUniformly replays test/missiles_test.cpp:12-23 with
// the tape: 100 rotations from startingFrame, then the observed
// histogram must be exactly {leftFrame: leftCount, rightFrame:
// rightCount} — both counts inside the upstream 30-70 bounds
// (EXPECT_THAT ... Gt(30U), Lt(70U), test/missiles_test.cpp:22).
type arrowRotateRow = struct {
	name          string
	startingFrame int
	leftFrame     int
	rightFrame    int
	tapeFrom      int
}

func TestArrowRotatesUniformly(t *testing.T) {
	// missiles.Arrow sprite frames are hardcoded 1-indexed over 16 directions
	// (cpp:451-456); the graphic id drives the branch.
	const frameCount = 16

	tests := []arrowRotateRow{
		{name: "mid-range 5 -> 4/6", startingFrame: 5, leftFrame: 4, rightFrame: 6, tapeFrom: 0},
		{name: "low wrap 1 -> 16/2", startingFrame: 1, leftFrame: 16, rightFrame: 2, tapeFrom: 100},
		{name: "high wrap 16 -> 15/1", startingFrame: 16, leftFrame: 15, rightFrame: 1, tapeFrom: 200},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			arrowRotateCase(t, tc, frameCount)
		})
	}
}

func arrowRotateCase(t *testing.T, tc arrowRotateRow, frameCount int) {
	t.Helper()
	// given
	missile := arrowFixture(missiles.Arrow, missiles.GraphicArrow)
	observed := make(map[int]int, 2)
	rnd := rotationTapeAt(t, tc.tapeFrom)

	// when
	for range 100 {
		missile.AnimFrame = tc.startingFrame
		missiles.RotateBlockedMissile(&missile, rnd)
		observed[missile.AnimFrame]++
	}

	// then
	for frame := range observed {
		if frame < 1 || frame > frameCount {
			t.Fatalf("AnimFrame = %d, out of 1..%d", frame, frameCount)
		}
	}
	assertHistogram(t, observed, tc.leftFrame, tc.rightFrame)
}

// TestAnimatedMissileRotatesUniformly replays
// test/missiles_test.cpp:25-35 for the missiles.Firebolt graphic (16 sprite
// directions — missile_sprites.tsv row 2, numFrames=16); the frame
// group MFNum is 0-indexed (cpp:458-464).
type animRotateRow = struct {
	name       string
	startingDI int
	leftDir    int
	rightDir   int
	tapeFrom   int
}

func TestAnimatedMissileRotatesUniformly(t *testing.T) {
	// missiles.Firebolt sheet: 16 directions, 0-indexed (missile_sprites.tsv
	// row 2 numFrames=16; GetMissileSpriteData(_miAnimType).animFAmt,
	// cpp:459).
	const frameCount = 16

	tests := []animRotateRow{
		{name: "mid-range 5 -> 4/6", startingDI: 5, leftDir: 4, rightDir: 6, tapeFrom: 300},
		{name: "low wrap 0 -> 15/1", startingDI: 0, leftDir: 15, rightDir: 1, tapeFrom: 400},
		{name: "high wrap 15 -> 14/0", startingDI: 15, leftDir: 14, rightDir: 0, tapeFrom: 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			animRotateCase(t, tc, frameCount)
		})
	}
}

func animRotateCase(t *testing.T, tc animRotateRow, frameCount int) {
	t.Helper()
	// given
	missile := arrowFixture(missiles.Firebolt, 1 /* missiles.Fireball graphic, misdat.h:37 */)
	observed := make(map[int]int, 2)
	rnd := rotationTapeAt(t, tc.tapeFrom)

	// when
	for range 100 {
		missile.MFNum = tc.startingDI
		missiles.RotateBlockedMissile(&missile, rnd)
		observed[missile.MFNum]++
	}

	// then
	for frame := range observed {
		if frame < 0 || frame >= frameCount {
			t.Fatalf("MFNum = %d, out of 0..%d", frame, frameCount-1)
		}
	}
	assertHistogram(t, observed, tc.leftDir, tc.rightDir)
}

func rotationTapeAt(t *testing.T, from int) func(int) int {
	t.Helper()
	if from > len(rotationTape) {
		t.Fatalf("tape exhausted at %d (len %d)", from, len(rotationTape))
	}
	i := from
	return func(int) int {
		if i >= len(rotationTape) {
			t.Fatalf("tape exhausted at draw %d", i)
		}
		v := rotationTape[i]
		i++
		return v
	}
}

// arrowFixture rebuilds the TEST setup upstream needs
// (test/missiles_test.cpp:41-45 Players.resize/LoadMissileData/
// AddMissile): AddMissile's emplace_back + field seeding
// (Source/cpp:2801-2839) for missiles.MissileID::missiles.Arrow,
// TARGET_MONSTERS, missiles.Direction::missiles.South, at tile (0,0). Upstream seeds the
// graphic via AddMissile -> setAnimation (cpp:2838 + 76-107);
// setAnimation leaves _miAnimFrame = 1 (cpp:96), and
// AddArrow's dst += midir makes GetDirection16(start, dst) = missiles.South, so
// _miAnimFrame = 0 + 1 = 1 (cpp:1796) — both hold in the
// fixture, asserted by the TEST (missiles_test.cpp:57). The player
// handle is an index — no global Players table yet.
func arrowFixture(id missiles.MissileID, animType int) missiles.Missile {
	// setAnimation seeds _miAnimDelay/_miAnimLen from the graphic's
	// per-direction frameDelay/frameLength columns (cpp:90-91,
	// missiles_sprites.tsv rows 1-2: missiles.Arrow 0/16, missiles.Fireball 0/14).
	delay, animLen := 0, 16
	if animType != missiles.GraphicArrow {
		animLen = 14 // missiles.Fireball frameLength (missile_sprites.tsv row 2)
	}
	m := missiles.Missile{
		Type:      id,
		AnimType:  animType,
		AnimLen:   animLen,
		AnimDelay: delay,
		Source:    0,
		Caster:    int(missiles.TargetMonsters),
	}
	m.SetTraveledFromTile(0, 0)
	return m
}

// assertHistogram checks the upstream EXPECT_THAT
// UnorderedElementsAre(Pair(leftFrame, Gt(30) && Lt(70)), Pair(rightFrame, ...))
// (test/missiles_test.cpp:22-23 / 33-34): after 100 rotations the
// observed frame histogram must be exactly {leftFrame, rightFrame} with
// both counts inside the 30-70 bounds.
func assertHistogram(t *testing.T, observed map[int]int, leftFrame, rightFrame int) {
	t.Helper()
	if len(observed) != 2 || observed[leftFrame] == 0 || observed[rightFrame] == 0 {
		t.Fatalf("observed = %v, want exactly {%d, %d}", observed, leftFrame, rightFrame)
	}
	for _, frame := range []int{leftFrame, rightFrame} {
		if observed[frame] <= 30 || observed[frame] >= 70 {
			t.Fatalf("frame %d count = %d, want 30 < count < 70 (missiles_test.cpp:22)", frame, observed[frame])
		}
	}
}
