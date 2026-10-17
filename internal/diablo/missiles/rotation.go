// Package missiles: RotateBlockedMissile (Source/missiles.cpp:449-465)
// — the hellfire blocked-shot 45° re-aim CheckMissileCol applies instead
// of deleting the missile (missiles.cpp:529-530); devilGo-mwlm slice 2.
//
// Upstream draws the rotation through PickRandomlyAmong({-1, 1})
// (random.hpp:377-384: GenerateRnd(size) indexes the list). The draw is
// injected as rnd func(int) int so tests script it (the repo's tape
// pattern, cf. internal/gameplay tap()); no math/rand global.
package missiles

// GraphicArrow mirrors MissileGraphicID::Arrow (Source/tables/misdat.h:36):
// the first missile_sprites.tsv row — the only graphic with the
// hardcoded 1-indexed 16-frame rotation branch.
const GraphicArrow = 0

// RotateBlockedMissile ports RotateBlockedMissile
// (Source/missiles.cpp:449-465) 1:1. rnd mirrors GenerateRnd: it takes
// the draw range (always 2 here — PickRandomlyAmong({-1, 1})) and
// returns the picked index; index 0 -> -1 (rotate left), 1 -> +1
// (rotate right).
//
// Arrow branch (missiles.cpp:451-456): _miAnimFrame is 1-indexed, so the
// wrap is (frame + rotation + 15) % 16 + 1.
//
// Every other graphic (missiles.cpp:458-464): the frame group _mimfnum
// is 0-indexed over the sheet's animFAmt directions, so the wrap is
// modular — MFNum carries the group (missiles.h:95, getFrameGroupRaw
// missiles.h:215-217).
func RotateBlockedMissile(m *Missile, rnd func(int) int) {
	rotation := rnd(2)
	if rotation == 0 {
		rotation = -1 // PickRandomlyAmong index 0 -> -1 (rotate left)
	}

	if m.AnimType == GraphicArrow {
		dir := m.AnimFrame + rotation
		m.AnimFrame = (dir+15)%16 + 1
		return
	}

	dir := m.MFNum + rotation
	animFAmt := GraphicFrameCount(m.AnimType)
	if dir < 0 {
		dir = animFAmt - 1
	} else if dir >= animFAmt {
		dir = 0
	}

	m.MFNum = dir
}

// GraphicFrameCount ports GetMissileSpriteData(animtype).animFAmt
// (Source/tables/misdat.h:214, missiles.cpp:459): the number of
// direction sprites on the graphic sheet (missile_sprites.tsv
// numFrames). Only rows the rotation path can reach are declared;
// slice-1's Go graphic table has no loader yet, so unknown ids keep
// their sheet frames (never 0).
func GraphicFrameCount(animType int) int {
	if animType == GraphicArrow {
		return 1 // Arrow: 1 frame (NotAnimated, missiles.cpp:451 branch)
	}
	return 16 // Fireball and friends: 16 direction sprites (row 2)
}
