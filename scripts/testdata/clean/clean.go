// Package clean is the purity control: the gate must report nothing here, even
// though the clock arrives under an alias and is used for Duration arithmetic.
package clean

import tt "time"

// TickDelay is the fixed sim step; Duration arithmetic is legal purity-wise.
const TickDelay = 50 * tt.Millisecond

// Step advances the sim by one fixed step.
func Step(tick int) int {
	return tick + int(TickDelay/tt.Millisecond)
}
