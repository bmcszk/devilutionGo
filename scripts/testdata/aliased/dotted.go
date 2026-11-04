//go:build ignore

// Package aliased also imports the wall clock with a dot import, so the banned
// readings appear as bare calls the gate must still name.
package aliased

import . "time"

// DotElapsed reads the wall clock through the dot import.
func DotElapsed() int64 {
	ticks := NewTicker(Second)
	defer ticks.Stop()

	<-ticks.C
	<-Tick(Second)

	return Since(Now()).Microseconds()
}
