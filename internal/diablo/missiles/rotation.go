// Package missiles: the re-aim of a blocked missile.
package missiles

// GraphicArrow is the arrow graphic id: the only graphic whose rotation
// uses the hardcoded 1-indexed 16-frame branch.
const GraphicArrow = 0

// missileAnimFrameAmounts is the animFAmt column of missile_sprites.tsv —
// the per-sheet frame group count (misdat.cpp:152 reads numFrames into
// MissileFileData::animFAmt). Values taken from the committed corpus
// (internal/testdata/gamedata/misdat/missile_sprites.tsv); a sheet missing
// here is treated as 1, the MissileGraphicID::None shape.
var missileAnimFrameAmounts = map[int]int{
	0:  1,  // Arrow (NotAnimated)
	1:  16, // Fireball
	20: 1,  // MagmaBallExplosion
	22: 1,  // BloodStar
	24: 8,  // MagmaBall
	31: 16, // Acid
	32: 1,  // AcidSplat
	40: 1,  // BloodStarBlue
	42: 1,  // BloodStarYellow
	44: 1,  // BloodStarRed
}

// RotateBlockedMissile re-aims a blocked missile one 45° step left or right.
//
// rnd takes the draw range and returns the picked index: 0 turns left,
// 1 turns right. Arrow frames are 1-indexed over 16 directions; every other
// graphic keeps a 0-indexed direction group that wraps over the sheet's
// direction count.
func RotateBlockedMissile(m *Missile, rnd func(int) int) {
	rotation := rnd(2)
	if rotation == 0 {
		rotation = -1
	}

	if m.AnimType == GraphicArrow {
		dir := m.AnimFrame + rotation
		m.AnimFrame = (dir+15)%16 + 1

		return
	}

	dir := m.MFNum + rotation

	animFAmt := graphicFrameCount(m.AnimType)
	if dir < 0 {
		dir = animFAmt - 1
	} else if dir >= animFAmt {
		dir = 0
	}

	m.MFNum = dir
}

// graphicFrameCount reports the sheet's animFAmt — the number of frame
// groups (directions) the sheet holds.
func graphicFrameCount(animType int) int {
	if frames, ok := missileAnimFrameAmounts[animType]; ok {
		return frames
	}

	return 1
}
