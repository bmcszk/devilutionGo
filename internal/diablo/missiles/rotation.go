// Package missiles: the re-aim of a blocked missile.
package missiles

// GraphicArrow is the arrow graphic id: the only graphic whose rotation
// uses the hardcoded 1-indexed 16-frame branch.
const GraphicArrow = 0

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

// graphicFrameCount stands in for the graphic table's direction sprite
// count: the arrow sheet holds one frame, the sheets the rotation path
// reaches hold 16 directions. Unknown ids keep the 16-direction count until
// the table loader lands.
func graphicFrameCount(animType int) int {
	if animType == GraphicArrow {
		return 1
	}

	return 16
}
