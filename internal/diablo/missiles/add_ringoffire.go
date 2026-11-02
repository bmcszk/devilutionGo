// Package missiles: the Hellfire Ring of Fire add body.
package missiles

// AddRingOfFire ports AddRingOfFire (missiles.cpp:1686-1691): the ring only
// remembers its launch tile in var1/var2 and burns for 7 ticks.
func AddRingOfFire(_ *HitWorld, m *Missile, _ *AddParameter) {
	m.Var1 = m.Position.Start.X // cpp:1688
	m.Var2 = m.Position.Start.Y // cpp:1689
	m.Duration = 7              // cpp:1690
}
