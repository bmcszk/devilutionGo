package engine

import "math"

var diabloLCGState uint32
var sglGameSeed uint32

// SetRndSeed (random.cpp:64-67)
func SetRndSeed(seed uint32) {
	diabloLCGState = seed
	sglGameSeed = seed
}

// GetLCGEngineState (random.cpp:70)
func GetLCGEngineState() uint32 { return sglGameSeed }

// DiscardRandomValues (random.cpp:75-80)
func DiscardRandomValues(count uint) {
	for ; count != 0; count-- {
		GenerateRandomNumber()
	}
}

// GenerateRandomNumber (random.cpp:83-85)
func GenerateRandomNumber() uint32 {
	diabloLCGState = diabloLCGState*0x015A4E35 + 1
	sglGameSeed = diabloLCGState
	return sglGameSeed
}

// AdvanceRndSeed (random.cpp:89-94)
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

// GenerateRnd (random.cpp:97-102)
func GenerateRnd(v int32) int32 {
	if v <= 0 {
		return 0
	}
	if v <= 0x7FFF {
		return (AdvanceRndSeed() >> 16) % v
	}
	return AdvanceRndSeed() % v
}

// FlipCoin (random.cpp:105-109)
func FlipCoin(v uint) bool { return GenerateRnd(int32(v)) == 0 }

// Xoshiro128pp (random.hpp:142-283)
type Xoshiro128pp struct{ s [4]uint32 }

// NewXoshiro128pp (random.hpp:267-270)
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

// Next (random.cpp:23-36)
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

// Jump (random.hpp:222-241)
func (g *Xoshiro128pp) Jump() {
	j := [4]uint32{0x8764000b, 0xf542d2d3, 0x6fa035c3, 0x77f2db5b}
	var s0, s1, s2, s3 uint32
	for _, v := range j {
		for b := uint(0); b < 32; b++ {
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

// rotl (std::rotl)
func rotl(x uint32, k uint) uint32 { return x<<k | x>>(32-k) }
