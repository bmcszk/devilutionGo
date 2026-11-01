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
