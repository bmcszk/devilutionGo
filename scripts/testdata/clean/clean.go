// Package clean is the purity control: the gate must report nothing here.
package clean

import "time"

// TickDelay is the fixed sim step; Duration arithmetic is legal purity-wise.
const TickDelay = 50 * time.Millisecond

// Step advances the sim by one fixed step.
func Step(tick int) int {
	return tick + int(TickDelay/time.Millisecond)
}
