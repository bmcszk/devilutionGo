package engine

import "math"

// Reference: diasurgical/DevilutionX Source/engine/random.cpp and
// Source/engine/random.hpp. The Borland LCG and xoshiro128++ streams must be
// byte-compatible with the C++ engine for save and network parity.

const (
	lcgMultiplier = 0x015A4E35 // random.cpp:17 — std::linear_congruential_engine<uint32_t, 0x015A4E35, 1, 0>
	lcgIncrement  = 1          // random.cpp:17
)

// diabloLCGState is the Borland C/C++ pseudo-random number generator state
// needed for vanilla compatibility (random.cpp:19 diabloGenerator).
var diabloLCGState uint32

// sglGameSeed mirrors C++ sglGameSeed (random.cpp:16): the last raw LCG output
// exposed for debugging and parity tests (random.cpp:70 GetLCGEngineState).
var sglGameSeed uint32

// SetRndSeed sets the vanilla LCG engine state to seed (random.cpp:64-67).
func SetRndSeed(seed uint32) {
	diabloLCGState = seed
	sglGameSeed = seed
}

// GetLCGEngineState returns the current vanilla LCG engine state
// (random.cpp:70-72 GetLCGEngineState, backed by sglGameSeed random.cpp:16).
func GetLCGEngineState() uint32 {
	return sglGameSeed
}

// DiscardRandomValues advances the vanilla LCG state by count rounds without
// using the outputs (random.cpp:75-80 DiscardRandomValues).
func DiscardRandomValues(count uint) {
	for ; count != 0; count-- {
		GenerateRandomNumber()
	}
}

// GenerateRandomNumber advances the vanilla LCG and returns the new 32-bit
// state (random.cpp:83-85 GenerateRandomNumber; random.cpp:17 constants).
func GenerateRandomNumber() uint32 {
	diabloLCGState = diabloLCGState*lcgMultiplier + lcgIncrement
	sglGameSeed = diabloLCGState
	return sglGameSeed
}

// AdvanceRndSeed advances the LCG and returns the new state interpreted as
// int32. INT_MIN is returned as-is because abs(INT_MIN) is undefined behavior
// (random.cpp:89-94, "since abs(INT_MIN) is undefined behavior, handle this
// value specially"); every other value is std::abs.
func AdvanceRndSeed() int32 {
	seed := int32(GenerateRandomNumber()) // random.cpp:91 static_cast<int32_t>(GenerateRandomNumber())
	if seed == math.MinInt32 {
		return math.MinInt32
	}
	if seed < 0 {
		return -seed // random.cpp:93 std::abs(seed)
	}
	return seed
}

// GenerateRnd returns a random integer less than v using the vanilla RNG.
//
// v <= 0 returns 0 without advancing the engine. For v <= 0x7FFF the high bits
// are used to correct for LCG bias, (AdvanceRndSeed() >> 16) % v (random.cpp:100-101);
// otherwise the plain AdvanceRndSeed() % v (random.cpp:102). Go's % on negative
// int32 matches C++11+ sign-preserving modulo, so the vanilla bug of results in
// (-v, -1] is preserved.
func GenerateRnd(v int32) int32 {
	if v <= 0 {
		return 0 // random.cpp:97-98
	}
	if v <= 0x7FFF {
		return (AdvanceRndSeed() >> 16) % v // random.cpp:100-101
	}
	return AdvanceRndSeed() % v // random.cpp:102
}

// FlipCoin returns GenerateRnd(frequency) == 0 (random.cpp:105-109): true 1 in
// frequency of the time, e.g. the default frequency of 2 is a 50/50 chance.
func FlipCoin(frequency uint) bool {
	return GenerateRnd(int32(frequency)) == 0
}

// Xoshiro128pp ports xoshiro128plusplus (random.hpp:142-283), the less
// predictable generator used for new game/level seeds. The zero value is not
// usable; construct with NewXoshiro128pp.
//
// Its Jump is equivalent to 2^64 calls to Next and generates non-overlapping
// subsequences for parallel computations (random.hpp:216-219).
type Xoshiro128pp struct {
	s [4]uint32 // random.hpp:282 state s
}

// splitMix32 ports SplitMix32 (random.hpp:147-168), used to expand a 32-bit
// seed into the four state words.
type splitMix32 struct {
	state uint32
}

func newSplitMix32(state uint32) splitMix32 {
	return splitMix32{state: state}
}

// next ports SplitMix32::next (random.hpp:157-163, fmix32 from MurmurHash3).
func (m *splitMix32) next() uint32 {
	m.state += 0x9e3779b9 // random.hpp:158
	z := m.state
	z = (z ^ (z >> 16)) * 0x85ebca6b // random.hpp:159
	z = (z ^ (z >> 13)) * 0xc2b2ae35 // random.hpp:160
	return z ^ (z >> 16)
}

// NewXoshiro128pp expands the 32-bit seed through SplitMix32 and returns a
// ready generator (random.hpp:267-270 seed(uint32_t) path).
func NewXoshiro128pp(seed uint32) *Xoshiro128pp {
	sm := newSplitMix32(seed)
	var g Xoshiro128pp
	for i := range g.s {
		g.s[i] = sm.next()
	}
	return &g
}

// Next ports xoshiro128plusplus::next (random.cpp:23-36).
func (g *Xoshiro128pp) Next() uint32 {
	s := &g.s
	result := rotl32(s[0]+s[3], 7) + s[0] // random.cpp:24 std::rotl(s[0] + s[3], 7) + s[0]

	t := s[1] << 9 // random.cpp:26

	s[2] ^= s[0] // random.cpp:28
	s[3] ^= s[1]
	s[1] ^= s[2]
	s[0] ^= s[3]

	s[2] ^= t // random.cpp:31

	s[3] = rotl32(s[3], 11) // random.cpp:35 std::rotl(s[3], 11)

	return result
}

// Jump advances the state as if 2^64 Next calls had been made, producing
// non-overlapping subsequences (random.hpp:219-241 xoshiro128plusplus::jump).
func (g *Xoshiro128pp) Jump() {
	jump := [4]uint32{0x8764000b, 0xf542d2d3, 0x6fa035c3, 0x77f2db5b} // random.hpp:222 JUMP

	var s0, s1, s2, s3 uint32
	for _, entry := range jump {
		for b := uint(0); b < 32; b++ {
			if entry&(1<<b) != 0 {
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

// rotl32 is Go's spelling of C++ std::rotl(x, k) on uint32.
func rotl32(x uint32, k uint) uint32 {
	return x<<k | x>>(32-k)
}
