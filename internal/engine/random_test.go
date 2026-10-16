package engine_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: diasurgical/DevilutionX Source/engine/random.cpp (LCG a=0x015A4E35, c=1)
// and Source/engine/random.hpp (xoshiro128++, SplitMix32). Golden values were
// derived by simulating the cited C++ formulas, not by calling the Go code.

// lcgNext is a step of the cited Borland LCG, used only to cross-check the
// implementation's state advance against the C++ formula.
func lcgNext(state uint32) uint32 {
	const multiplier = 0x015A4E35 // random.cpp:17 (std::linear_congruential_engine<uint32_t, 0x015A4E35, 1, 0>)
	const increment = 1           // random.cpp:17
	return state*multiplier + increment
}

func TestSetRndSeedGoldenSequence(t *testing.T) {
	engine.SetRndSeed(2588)
	// Hand-derived from random.cpp:83-85 (state = state*0x015A4E35 + 1 mod 2^32),
	// starting from SetRndSeed(2588):
	//  2588*22695477+1        = 2901319629
	//  2901319629*22695477+1  = 2099874674
	//  2099874674*22695477+1  =   170564251
	//    170564251*22695477+1 =  191577112
	//   191577112*22695477+1  = 1106428153
	//  1106428153*22695477+1  = 1855443342
	//  1855443342*22695477+1  = 1630774887
	//  1630774887*22695477+1  = 3281704276
	want := []uint32{
		2901319629, 2099874674, 170564251, 191577112,
		1106428153, 1855443342, 1630774887, 3281704276,
	}
	for i, w := range want {
		if got := engine.GenerateRandomNumber(); got != w {
			t.Fatalf("GenerateRandomNumber() #%d = %d, want %d", i+1, got, w)
		}
	}
}

func TestSetRndSeedStateAdvance(t *testing.T) {
	engine.SetRndSeed(2588)
	state := uint32(2588)
	for i := 0; i < 1000; i++ {
		state = lcgNext(state)
		if got := engine.GenerateRandomNumber(); got != state {
			t.Fatalf("after %d rounds: GenerateRandomNumber() = %d, want %d", i+1, got, state)
		}
		if got := engine.GetLCGEngineState(); got != state {
			t.Fatalf("after %d rounds: GetLCGEngineState() = %d, want %d", i+1, got, state)
		}
	}
}

func TestDiscardRandomValues(t *testing.T) {
	engine.SetRndSeed(2588)
	engine.DiscardRandomValues(5)
	// 2588 advanced 5 times per TestSetRndSeedGoldenSequence.
	if got, want := engine.GetLCGEngineState(), uint32(1106428153); got != want {
		t.Fatalf("GetLCGEngineState() = %d, want %d", got, want)
	}
}

func TestAdvanceRndSeed(t *testing.T) {
	tests := []struct {
		seed uint32
		want int32
	}{
		// All seed/expectation pairs from upstream test/random_test.cpp (AbsDistribution).
		{seed: 1457187811, want: math.MinInt32}, // yields 0x80000000, random.cpp:89-93 special case
		{seed: 3604671459, want: 0},             // yields 0
		{seed: 0, want: 1},                      // yields +1
		{seed: 2914375622, want: 1},             // yields -1
		{seed: 3604671460, want: 22695477},      // yields +22695477
		{seed: 3604671458, want: 22695477},      // yields -22695477
		{seed: 1902003768, want: 429496729},     // yields +429496729
		{seed: 1012371854, want: 429496729},     // yields -429496729
		{seed: 189776845, want: 1212022642},     // yields +1212022642
		{seed: 2724598777, want: 1212022642},    // yields -1212022642
		{seed: 76596137, want: 2147483646},      // yields +2147483646
		{seed: 2837779485, want: 2147483646},    // yields -2147483646
		{seed: 766891974, want: 2147483647},     // yields +2147483647
		{seed: 2147483648, want: 2147483647},    // yields -2147483647
	}
	for _, tt := range tests {
		engine.SetRndSeed(tt.seed)
		if got := engine.AdvanceRndSeed(); got != tt.want {
			t.Errorf("AdvanceRndSeed(seed=%d) = %d, want %d", tt.seed, got, tt.want)
		}
	}
}

func TestGenerateRndInvalidRange(t *testing.T) {
	// random.cpp:96-98: v <= 0 returns 0 without advancing the engine.
	const initialSeed = uint32(44444444)
	engine.SetRndSeed(initialSeed)
	for _, v := range []int32{0, -1, math.MinInt32} {
		if got := engine.GenerateRnd(v); got != 0 {
			t.Errorf("GenerateRnd(%d) = %d, want 0", v, got)
		}
		if got, want := engine.GetLCGEngineState(), initialSeed; got != want {
			t.Fatalf("GenerateRnd(%d) advanced the state: %d, want %d", v, got, want)
		}
	}
}

func TestGenerateRndSingleRange(t *testing.T) {
	// Any seed must give 0 for v=1 while still advancing the engine
	// (random.cpp:96-102). Seeds come from a local fixed-seed generator,
	// standing in for gtest's arbitrary default seed.
	arbitrary := rand.New(rand.NewSource(1))
	const initialSeed = uint32(2588)
	engine.SetRndSeed(initialSeed)
	for i := 0; i < 100; i++ {
		engine.SetRndSeed(arbitrary.Uint32())
		if got := engine.GenerateRnd(1); got != 0 {
			t.Fatalf("GenerateRnd(1) = %d, want 0", got)
		}
	}
	if engine.GetLCGEngineState() == initialSeed {
		t.Fatal("GenerateRnd(1) never advanced the engine state")
	}
}

func TestGenerateRndShiftMod(t *testing.T) {
	const maxBound = int32(32767)
	tests := []struct {
		seed uint32
		want int32
	}{
		// random.cpp:100-101: v <= 0x7FFF uses (AdvanceRndSeed() >> 16) % v.
		// Seed/expectation pairs from upstream test/random_test.cpp.
		{seed: 1625910243, want: 0},     // raw 32768,  sign-carry shift -> -1, mod -> 0
		{seed: 3267226595, want: 1},     // raw 65536,  shift -> 1
		{seed: 3942116323, want: 1},     // raw -65536, sign-carry shift -> -1, mod -> -1+32767
		{seed: 4279561187, want: 2},     // raw 131072, shift -> 2
		{seed: 659483619, want: 4},      // raw 262144, shift -> 4
		{seed: 3604671458, want: 346},   // raw 22695477
		{seed: 3604671460, want: 346},   // raw -22695477
		{seed: 1012371854, want: 6553},  // raw 429496729
		{seed: 1902003768, want: 6553},  // raw -429496729
		{seed: 189776845, want: 18493},  // raw 1212022642
		{seed: 2724598777, want: 18493}, // raw -1212022642
		{seed: 76596137, want: 0},       // raw 2147483646, shift -> 32767, mod 32767 -> 0
		{seed: 766891974, want: 0},      // raw 2147483647, shift -> 32767, mod 32767 -> 0
		{seed: 2147483648, want: 0},     // raw -2147483647, shift -> -32767, mod -> 0
	}
	for _, tt := range tests {
		engine.SetRndSeed(tt.seed)
		if got := engine.GenerateRnd(maxBound); got != tt.want {
			t.Errorf("GenerateRnd(%d, seed=%d) = %d, want %d", maxBound, tt.seed, got, tt.want)
		}
	}

	// Sign-preserving right shift of INT_MIN: -32768 % 32767 = -1.
	// Upstream: ShiftModDistributionSignCarry.
	engine.SetRndSeed(1457187811)
	if got, want := engine.GenerateRnd(32767), int32(-1); got != want {
		t.Errorf("GenerateRnd(32767, seed=1457187811) = %d, want %d", got, want)
	}
}

func TestGenerateRndMod(t *testing.T) {
	tests := []struct {
		v    int32
		want int32
	}{
		// random.cpp:102: plain AdvanceRndSeed() % v for v > 0x7FFF, sign-preserving.
		// Upstream: ModDistributionSignPreserving (seed yields INT_MIN).
		{v: 32768, want: 0},
		{v: 32769, want: -2},
		{v: 65535, want: -32768},
		{v: math.MaxInt32, want: -1},
	}
	for _, tt := range tests {
		engine.SetRndSeed(1457187811)
		if got := engine.GenerateRnd(tt.v); got != tt.want {
			t.Errorf("GenerateRnd(%d, seed=1457187811) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestGenerateRndNegativeReturnValues(t *testing.T) {
	// The vanilla bug: sign-preserving modulo of a negative intermediate can
	// return values in (-v, -1]. All pairs from upstream test/random_test.cpp
	// (NegativeReturnValues); every seed below yields INT_MIN.
	tests := []struct {
		v    int32
		want int32
	}{
		{v: 3, want: -2},
		{v: 5, want: -3},
		{v: 6, want: -2},
		{v: 7, want: -1},
		{v: 9, want: -8},
		{v: 10, want: -8},
		{v: 11, want: -10},
		{v: 12, want: -8},
		{v: 13, want: -8},
		{v: 14, want: -8},
		{v: 15, want: -8},
		{v: 17, want: -9},
		{v: 18, want: -8},
		{v: 19, want: -12},
		{v: 20, want: -8},
		{v: 21, want: -8},
		{v: 22, want: -10},
		{v: 23, want: -16},
		{v: 24, want: -8},
		{v: 25, want: -18},
		{v: 26, want: -8},
		{v: 27, want: -17},
		{v: 28, want: -8},
		{v: 29, want: -27},
		{v: 30, want: -8},
		{v: 31, want: -1},
		{v: 37, want: -23},
		{v: 38, want: -12},
		{v: 40, want: -8}, // commonly used for dungeon megatile coordinates
		{v: 41, want: -9},
		{v: 52, want: -8},
		{v: 80, want: -48}, // commonly used for dungeon tile coordinates
		{v: 100, want: -68},
	}
	for _, tt := range tests {
		engine.SetRndSeed(1457187811)
		if got := engine.GenerateRnd(tt.v); got != tt.want {
			t.Errorf("GenerateRnd(%d, seed=1457187811) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestGenerateRndPowerOfTwoBounds(t *testing.T) {
	// Powers of two cleanly divide INT_MIN: result is always 0.
	for v := int32(1); v < 32768; v *= 2 {
		engine.SetRndSeed(1457187811)
		if got := engine.GenerateRnd(v); got != 0 {
			t.Errorf("GenerateRnd(%d, seed=1457187811) = %d, want 0", v, got)
		}
	}
}

func TestGenerateRndRange(t *testing.T) {
	for _, v := range []int32{2, 7, 1000, 32767, 32768, 65534, 100000} {
		rng := rand.New(rand.NewSource(int64(v)))
		for i := 0; i < 200; i++ {
			engine.SetRndSeed(rng.Uint32())
			got := engine.GenerateRnd(v)
			if got < 0 || got >= v {
				t.Fatalf("GenerateRnd(%d) = %d, outside [0,%d)", v, got, v)
			}
		}
	}
}

func TestFlipCoin(t *testing.T) {
	tests := []struct {
		seed      uint32
		frequency uint
		want      bool
	}{
		// random.cpp:105-109: FlipCoin is GenerateRnd(frequency) == 0.
		// seed 1457187811 -> AdvanceRndSeed = INT_MIN -> (>>16) = -32768 -> -32768%2 = 0.
		{seed: 1457187811, frequency: 2, want: true},
		// seed 3604671459 -> AdvanceRndSeed = 0 -> 0%2 = 0.
		{seed: 3604671459, frequency: 2, want: true},
		// seed 0 -> AdvanceRndSeed = 1 -> (1>>16) = 0 -> 0%3 = 0.
		{seed: 0, frequency: 3, want: true},
		// seed 2588 -> AdvanceRndSeed = 2901319629 -> (>>16) = 44267 -> 44267%2 = 1.
		{seed: 2588, frequency: 2, want: false},
	}
	for _, tt := range tests {
		engine.SetRndSeed(tt.seed)
		if got := engine.FlipCoin(tt.frequency); got != tt.want {
			t.Errorf("FlipCoin(%d, seed=%d) = %v, want %v", tt.frequency, tt.seed, got, tt.want)
		}
	}
}

func TestXoshiro128plusplusNextGoldenStream(t *testing.T) {
	// Seeding path: uint32 seed -> SplitMix32 (random.hpp:150-163), then the
	// output/scramble steps of random.cpp:23-36.
	g := engine.NewXoshiro128pp(2588)
	want := []uint32{0x6afe3f2e, 0xb4f92059, 0xb337e0a7, 0x4fe119cd}
	for i, w := range want {
		if got := g.Next(); got != w {
			t.Fatalf("Next() #%d = %#08x, want %#08x", i+1, got, w)
		}
	}
}

func TestXoshiro128plusplusReproducible(t *testing.T) {
	g := engine.NewXoshiro128pp(2588)
	again := engine.NewXoshiro128pp(2588)
	for i := 0; i < 100; i++ {
		if g.Next() != again.Next() {
			t.Fatalf("streams diverged at output #%d for identical seed", i+1)
		}
	}
	other := engine.NewXoshiro128pp(2589)
	if g.Next() == other.Next() {
		t.Fatal("seeds 2588 and 2589 produced identical first outputs")
	}
}

func TestXoshiro128plusplusJumpChangesStream(t *testing.T) {
	// jump() (random.hpp:219-241) is equivalent to 2^64 Next() calls, so a
	// jumped generator must produce a different stream than the unjumped one.
	base := engine.NewXoshiro128pp(2588)
	jumped := engine.NewXoshiro128pp(2588)
	jumped.Jump()

	same := true
	for i := 0; i < 64; i++ {
		if base.Next() != jumped.Next() {
			same = false
			break
		}
	}
	if same {
		t.Fatal("jump() left the stream unchanged")
	}
}

func TestXoshiro128plusplusDoubleJump(t *testing.T) {
	// Two jumps advance 2*2^64 steps: distinct from a single jump's stream.
	jumped := engine.NewXoshiro128pp(2588)
	jumped.Jump()
	double := engine.NewXoshiro128pp(2588)
	double.Jump()
	double.Jump()
	for i := 0; i < 64; i++ {
		a, b := jumped.Next(), double.Next()
		if a == b {
			t.Fatalf("single-jumped and double-jumped streams collided at output #%d (%#08x)", i+1, a)
		}
	}
}

func TestXoshiro128plusplusJumpDeterministic(t *testing.T) {
	// Identical seed + identical jump must replay the identical stream.
	jumped := engine.NewXoshiro128pp(2588)
	jumped.Jump()
	replay := engine.NewXoshiro128pp(2588)
	replay.Jump()
	for i := 0; i < 64; i++ {
		if replay.Next() != jumped.Next() {
			t.Fatalf("jumped streams diverged at output #%d", i+1)
		}
	}
}
