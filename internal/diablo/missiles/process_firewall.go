// Package missiles: the FireWall row's per-tick body.
package missiles

// fireWallFrameStart/fireWallFrameIdle are the FireWallFrame ordinals
// (missiles.h:74-77, enum class FireWallFrame).
const (
	fireWallFrameStart = 0
	fireWallFrameIdle  = 1
)

// fireWallExpLight is the expansion light table of ProcessFireWall
// (missiles.cpp:3062-3063).
var fireWallExpLight = [12]uint8{2, 3, 4, 5, 5, 6, 7, 8, 9, 10, 11, 12}

// ProcessFireWall ports ProcessFireWall (missiles.cpp:3060-3096): the two
// frame-group flips, the collision probe, the delete tail and the expansion
// light sweep that walks ExpLight out and back.
func ProcessFireWall(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3066

	if m.Duration == m.Var1 { // cpp:3067-3069
		m.setFrameGroup(fireWallFrameIdle)
		m.AnimFrame = randomIntLessThan(w, 11) + 1
	}

	if m.Duration == m.AnimLen-1 { // cpp:3070-3073
		m.setFrameGroup(fireWallFrameStart)
		m.AnimFrame = 13
		m.AnimAdd = -1
	}

	var onlyHitWalking *Direction // cpp:3074-3076
	if m.Var3 != 0 {
		direction := Direction(m.Var3 - 1)
		onlyHitWalking = &direction
	}

	CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{ // cpp:3077
		DamageType:     DamageTypeFor(m.Type),
		MinDamage:      m.Dam,
		MaxDamage:      m.Dam,
		Shifted:        true,
		Keep:           true,
		OnlyHitWalking: onlyHitWalking,
	})

	if m.Duration == 0 { // cpp:3078-3081
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	m.sweepFireWallLight(w) // cpp:3083-3093
	PutMissile(w, m)
}

// sweepFireWallLight ports the expansion light sweep of ProcessFireWall
// (missiles.cpp:3082-3093): while the start group runs it adds the light on
// the first tick and walks ExpLight out to its last entry and back.
func (m *Missile) sweepFireWallLight(w *HitWorld) {
	maxIndex := len(fireWallExpLight) - 1
	if m.MFNum != fireWallFrameStart || m.Duration == 0 || m.Var2 > maxIndex*2 {
		return
	}

	if m.Var2 == 0 {
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, fireWallExpLight[0])
	}

	delta := m.Var2 - maxIndex
	if delta < 0 {
		delta = -delta
	}

	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
		fireWallExpLight[maxIndex-delta])
	m.Var2++
}

// ProcessFlameWave ports ProcessFlameWave (missiles.cpp:3475-3507): the tile
// shifts that undo the add-time render offset, the var1 animation counter, the
// collision probe (which alone can zero the duration — the wave has no tick
// decrement, cpp:3489-3492) and the ExpLight sweep that clamps at its last
// entry instead of wrapping.
func ProcessFlameWave(w *HitWorld, m *Missile) {
	m.Position.Tile = shiftTile(m.Position.Tile, North) // cpp:3481
	m.Position.Offset.DeltaY += 32                      // cpp:3482

	m.Var1++                 // cpp:3484
	if m.Var1 == m.AnimLen { // cpp:3485-3488
		m.setFrameGroup(fireWallFrameIdle)
		m.AnimFrame = randomIntLessThan(w, 11) + 1
	}

	kept := m.Duration // cpp:3489
	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), m.Dam, m.Dam, false, false)

	if m.HitFlag { // cpp:3491-3492
		m.Duration = kept
	}

	if m.Duration == 0 { // cpp:3493-3495
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	} else {
		m.sweepFlameWaveLight(w) // cpp:3497-3501
	}

	m.Position.Tile = shiftTile(m.Position.Tile, South) // cpp:3504
	m.Position.Offset.DeltaY -= 32                      // cpp:3505
	PutMissile(w, m)
}

// sweepFlameWaveLight ports the light half of ProcessFlameWave
// (missiles.cpp:3497-3501): the light starts on the first tick and the radius
// walks ExpLight, stopping at its last entry.
func (m *Missile) sweepFlameWaveLight(w *HitWorld) {
	if m.Var2 == 0 {
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, fireWallExpLight[0])
	}

	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
		fireWallExpLight[m.Var2])

	if m.Var2 < len(fireWallExpLight)-1 {
		m.Var2++
	}
}
