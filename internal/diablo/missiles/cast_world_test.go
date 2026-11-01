// Package missiles_test: the casting world the corpus-row checks drive.
package missiles_test

import "github.com/bmcszk/devilutionGo/internal/diablo/missiles"

// castWorld is newSilentWorld with a live caster behind it. Rows whose add body
// reads the player — AddRage's life guard (missiles.cpp:2882-2884), for one —
// fizzle on the zero-value player newSilentWorld reports, so the corpus-row
// checks would see a nil missile for a row that is in fact wired.
func castWorld() *missiles.HitWorld {
	w := newSilentWorld()
	w.AddSource.PlayerStats = func(int) missiles.AddPlayerStats {
		return missiles.AddPlayerStats{Level: 1, HitPoints: 200 << 6}
	}

	return w
}
