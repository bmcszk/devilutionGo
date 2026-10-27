package engine

import "math"

var (
	diabloLCGState uint32
	sglGameSeed    uint32
)

// Xoshiro128pp is a xoshiro128++ generator for seeds that must be less
// predictable than the vanilla LCG.
type Xoshiro128pp struct{ s [4]uint32 }

// SetRndSeed resets the LCG state and the game seed to the same value.
func SetRndSeed(seed uint32) {
	diabloLCGState = seed
	sglGameSeed = seed
}

// GetLCGEngineState returns the LCG state that callers capture as a level seed.
func GetLCGEngineState() uint32 { return sglGameSeed }

// DiscardRandomValues advances the LCG by count values.
func DiscardRandomValues(count uint) {
	for ; count != 0; count-- {
		GenerateRandomNumber()
	}
}

// GenerateRandomNumber advances the LCG and returns the new state.
func GenerateRandomNumber() uint32 {
	diabloLCGState = diabloLCGState*0x015A4E35 + 1
	sglGameSeed = diabloLCGState

	return sglGameSeed
}

// AdvanceRndSeed returns the absolute value of the next engine state, except
// that the minimum int32 is returned unchanged: it has no positive counterpart.
func AdvanceRndSeed() int32 {
	seed := int32(GenerateRandomNumber())
	if seed == math.MinInt32 {
		return math.MinInt32
	}

	if seed < 0 {
		return -seed
	}

	return seed
}

// GenerateRnd returns a random value in [0, v), or rarely a negative value in
// (-v, -1]; a non-positive v returns 0 without advancing the LCG.
func GenerateRnd(v int32) int32 {
	if v <= 0 {
		return 0
	}

	if v <= 0x7FFF {
		// The high bits carry less LCG bias than the low ones.
		return (AdvanceRndSeed() >> 16) % v
	}

	return AdvanceRndSeed() % v
}

// RandomIntLessThan returns a random value in [0, v): the clamped GenerateRnd
// (random.hpp:391-394), so the rare negative result becomes 0.
func RandomIntLessThan(v int32) int32 {
	return max(GenerateRnd(v), 0)
}

// RandomIntBetween returns a random value in [min, max], or in [min, max) when
// halfOpen is set (random.hpp:403-407).
//
//nolint:revive // the half-open flag is upstream's signature, not control coupling.
func RandomIntBetween(lo, hi int32, halfOpen bool) int32 {
	size := hi - lo + 1
	if halfOpen {
		size = hi - lo
	}

	return RandomIntLessThan(size) + lo
}

// FlipCoin returns true 1 time in v.
func FlipCoin(v uint) bool { return GenerateRnd(int32(v)) == 0 }

// NewXoshiro128pp returns a generator seeded from seed through SplitMix32.
func NewXoshiro128pp(seed uint32) *Xoshiro128pp {
	sm := seed

	var g Xoshiro128pp
	for i := range g.s {
		sm += 0x9e3779b9
		z := sm
		z = (z ^ (z >> 16)) * 0x85ebca6b
		z = (z ^ (z >> 13)) * 0xc2b2ae35
		g.s[i] = z ^ (z >> 16)
	}

	return &g
}

func rotl(x uint32, k uint) uint32 { return x<<k | x>>(32-k) }

// Next returns the next 32-bit value and advances the state.
func (g *Xoshiro128pp) Next() uint32 {
	res := rotl(g.s[0]+g.s[3], 7) + g.s[0]
	t := g.s[1] << 9
	g.s[2] ^= g.s[0]
	g.s[3] ^= g.s[1]
	g.s[1] ^= g.s[2]
	g.s[0] ^= g.s[3]
	g.s[2] ^= t
	g.s[3] = rotl(g.s[3], 11)

	return res
}

// Jump advances the generator by 2^64 steps, so the jumped generator yields a
// non-overlapping subsequence.
func (g *Xoshiro128pp) Jump() {
	j := [4]uint32{0x8764000b, 0xf542d2d3, 0x6fa035c3, 0x77f2db5b}

	var s0, s1, s2, s3 uint32

	for _, v := range j {
		for b := range uint(32) {
			if v&(1<<b) != 0 {
				s0 ^= g.s[0]
				s1 ^= g.s[1]
				s2 ^= g.s[2]
				s3 ^= g.s[3]
			}

			g.Next()
		}
	}

	g.s[0], g.s[1], g.s[2], g.s[3] = s0, s1, s2, s3
}
