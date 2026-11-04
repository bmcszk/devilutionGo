//go:build ignore

// Package aliased is a deliberately impure fixture: the wall clock arrives under
// an alias, and the gate must still name every banned reading. The ignore tag
// keeps the Go formatters and the linter out; the gate parses sources regardless
// of build tags.
package aliased

import t "time"

// WaitElapsed reads the wall clock through the aliased import.
func WaitElapsed() int64 {
	ticks := t.NewTicker(t.Second)
	defer ticks.Stop()

	<-ticks.C
	<-t.Tick(t.Second)

	return t.Since(t.Now()).Microseconds()
}
