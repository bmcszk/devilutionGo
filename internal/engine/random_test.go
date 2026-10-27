package engine_test

import (
	"math"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: diasurgical/DevilutionX Source/engine/random.cpp (LCG
// a=0x015A4E35, c=1) and random.hpp (xoshiro128++, SplitMix32). Golden
// values derived by simulating the cited C++ formulas, not the Go code.

func TestSetRndSeedGoldenSequence(t *testing.T) {
	// given
	// Hand-derived from random.cpp:83-85 (state = state*0x015A4E35 + 1
	// mod 2^32) from engine.SetRndSeed(2588): 2588→2901319629→2099874674→170564251
	// →191577112→1106428153→1855443342→1630774887→3281704276.
	engine.SetRndSeed(2588)

	want := []uint32{2901319629, 2099874674, 170564251, 191577112, 1106428153, 1855443342, 1630774887, 3281704276}
	// when
	for i, w := range want {
		if got := engine.GenerateRandomNumber(); got != w {
			t.Fatalf("engine.GenerateRandomNumber() #%d = %d, want %d", i+1, got, w)
		}
	}
	// The state mirror rides the same arithmetic (getLCGEngineState,
	// random.cpp:70 — sglGameSeed is save-format bound).
	// then
	if got, want := engine.GetLCGEngineState(), uint32(3281704276); got != want {
		t.Fatalf("engine.GetLCGEngineState() = %d, want %d", got, want)
	}
}

func TestDiscardRandomValues(t *testing.T) {
	// given
	engine.SetRndSeed(2588)
	// when
	engine.DiscardRandomValues(5)
	// then
	// 2588 advanced 5 times per TestSetRndSeedGoldenSequence.
	if got, want := engine.GetLCGEngineState(), uint32(1106428153); got != want {
		t.Fatalf("engine.GetLCGEngineState() = %d, want %d", got, want)
	}
}

func TestAdvanceRndSeed(t *testing.T) {
	// given
	// Seed/expectation pairs from upstream test/random_test.cpp
	// (AbsDistribution); row 1 is the INT_MIN special case (random.cpp:89-93).
	tests := []struct {
		seed uint32
		want int32
	}{
		{seed: 1457187811, want: math.MinInt32}, // yields 0x80000000, random.cpp:89-93 special case
		{seed: 3604671459, want: 0},             // yields 0
		{seed: 0, want: 1},
		{seed: 2914375622, want: 1},
		{seed: 3604671460, want: 22695477},
		{seed: 3604671458, want: 22695477},
		{seed: 1902003768, want: 429496729},
		{seed: 1012371854, want: 429496729},
		{seed: 189776845, want: 1212022642},
		{seed: 2724598777, want: 1212022642},
		{seed: 76596137, want: 2147483646},
		{seed: 2837779485, want: 2147483646},
		{seed: 766891974, want: 2147483647},
		{seed: 2147483648, want: 2147483647},
	}
	// when every seed is advanced once
	got := make([]int32, len(tests))
	for i, tt := range tests {
		got[i] = seedAndCall(tt.seed, engine.AdvanceRndSeed)
	}

	// then engine.AdvanceRndSeed returns tt.want for every row
	for i, tt := range tests {
		assertRnd(t, got[i], tt.want, "engine.AdvanceRndSeed")
	}
}

func TestGenerateRndInvalidRange(t *testing.T) {
	// given
	// random.cpp:96-98: v <= 0 returns 0 without advancing the
	// v = 1 (folded from the removed loop): x mod 1 = 0 for every seed
	// while engine.AdvanceRndSeed advances (random.cpp:100-101).
	const initialSeed = uint32(44444444)
	engine.SetRndSeed(initialSeed)
	// when
	for _, v := range []int32{0, -1, math.MinInt32} {
		if got := engine.GenerateRnd(v); got != 0 {
			t.Errorf("engine.GenerateRnd(%d) = %d, want 0", v, got)
		}

		if got, want := engine.GetLCGEngineState(), initialSeed; got != want {
			t.Fatalf("engine.GenerateRnd(%d) advanced the state: %d, want %d", v, got, want)
		}
	}
	// then
	if got := engine.GenerateRnd(1); got != 0 {
		t.Errorf("engine.GenerateRnd(1) = %d, want 0", got)
	}

	if engine.GetLCGEngineState() == initialSeed {
		t.Fatal("engine.GenerateRnd(1) never advanced the engine state")
	}
}

func TestGenerateRndShiftMod(t *testing.T) {
	// given
	// random.cpp:100-101: v <= 0x7FFF uses (engine.AdvanceRndSeed() >> 16) % v;
	// pairs verbatim from upstream test/random_test.cpp, the last row the
	// INT_MIN sign-carry shift (-32768 % 32767 = -1).
	const maxBound = int32(32767)

	tests := []struct {
		seed uint32
		want int32
	}{
		{seed: 1625910243, want: 0},
		{seed: 3267226595, want: 1},
		{seed: 3942116323, want: 1},
		{seed: 4279561187, want: 2},
		{seed: 659483619, want: 4},
		{seed: 3604671458, want: 346},
		{seed: 3604671460, want: 346},
		{seed: 1012371854, want: 6553},
		{seed: 1902003768, want: 6553},
		{seed: 189776845, want: 18493},
		{seed: 2724598777, want: 18493},
		{seed: 76596137, want: 0},
		{seed: 766891974, want: 0},
		{seed: 2147483648, want: 0},
		{seed: 1457187811, want: -1}, // INT_MIN shift, sign-carry
	}
	// when every seed is advanced once
	got := make([]int32, len(tests))
	for i, tt := range tests {
		got[i] = seedAndCall(tt.seed, func() int32 { return engine.GenerateRnd(maxBound) })
	}

	// then engine.GenerateRnd(maxBound) returns tt.want for every row
	for i, tt := range tests {
		assertRnd(t, got[i], tt.want, "engine.GenerateRnd(shift)")
	}
}

func TestGenerateRndMod(t *testing.T) {
	// given
	// random.cpp:102: plain engine.AdvanceRndSeed() % v for v > 0x7FFF,
	// sign-preserving (upstream ModDistributionSignPreserving).
	tests := []struct {
		v    int32
		want int32
	}{
		{v: 32768, want: 0},
		{v: 32769, want: -2},
		{v: 65535, want: -32768},
		{v: math.MaxInt32, want: -1},
	}
	// when every bound is evaluated on the INT_MIN seed
	got := make([]int32, len(tests))
	for i, tt := range tests {
		got[i] = seedAndCall(1457187811, func() int32 { return engine.GenerateRnd(tt.v) })
	}

	// then engine.GenerateRnd(tt.v) returns tt.want for every row
	for i, tt := range tests {
		assertRnd(t, got[i], tt.want, "engine.GenerateRnd(mod)")
	}
}

// The powers-of-two premise (folded from the removed loop) and the vanilla
// modulo bug: INT_MIN % 2^k == 0 (pinned by the rows above), and
// sign-preserving modulo of a negative intermediate returns (-v, -1]. All
// pairs from upstream test/random_test.cpp; every seed below yields INT_MIN.
func TestGenerateRndNegativeReturnValues(t *testing.T) {
	// given
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
	// when every bound is evaluated on the INT_MIN seed
	got := make([]int32, len(tests))
	for i, tt := range tests {
		got[i] = seedAndCall(1457187811, func() int32 { return engine.GenerateRnd(tt.v) })
	}

	// then engine.GenerateRnd(tt.v) returns tt.want for every row
	for i, tt := range tests {
		assertRnd(t, got[i], tt.want, "engine.GenerateRnd(neg)")
	}
}

func TestFlipCoin(t *testing.T) {
	// given
	// random.cpp:105-109: flipCoin is engine.GenerateRnd(frequency) == 0.
	// Hand-derived: 1457187811 → INT_MIN → (>>16) = -32768 → %2 = 0;
	// 3604671459 → 0; 0 → 1 → (1>>16) = 0 → %3 = 0; 2588 → …44267 → %2 = 1.
	tests := []struct {
		seed      uint32
		frequency uint
		want      bool
	}{
		{seed: 1457187811, frequency: 2, want: true},
		{seed: 3604671459, frequency: 2, want: true},
		{seed: 0, frequency: 3, want: true},
		{seed: 2588, frequency: 2, want: false},
	}
	// when every seed is flipped at its frequency
	got := make([]bool, len(tests))
	for i, tt := range tests {
		got[i] = seedAndCall(tt.seed, func() bool { return engine.FlipCoin(tt.frequency) })
	}

	// then engine.FlipCoin(tt.frequency) returns tt.want for every row
	for i, tt := range tests {
		assertRnd(t, got[i], tt.want, "flipCoin")
	}
}

func TestXoshiro128plusplusNextGoldenStream(t *testing.T) {
	// given
	// Seeding path: SplitMix32 (random.hpp:150-163) followed by the
	// output/scramble steps of random.cpp:23-36.
	g := engine.NewXoshiro128pp(2588)
	want := []uint32{0x6afe3f2e, 0xb4f92059, 0xb337e0a7, 0x4fe119cd}
	// when
	got := []uint32{g.Next(), g.Next(), g.Next(), g.Next()}
	// then
	if !slices.Equal(got, want) {
		t.Fatalf("Next() = %v, want %v", got, want)
	}
}

func TestXoshiro128plusplusJumpChangesStream(t *testing.T) {
	// given
	// jump() (random.hpp:219-241) = 2^64 Next() calls: a jumped
	// generator must produce a different stream than the unjumped one.
	base, jumped := engine.NewXoshiro128pp(2588), engine.NewXoshiro128pp(2588)
	// when
	jumped.Jump()

	baseStream := make([]uint32, 0, 64)
	jumpedStream := make([]uint32, 0, 64)
	for range 64 {
		baseStream = append(baseStream, base.Next())
		jumpedStream = append(jumpedStream, jumped.Next())
	}
	// then
	if slices.Equal(baseStream, jumpedStream) {
		t.Fatal("jump() left the stream unchanged")
	}
}

// seedAndCall seeds the LCG then returns one call's exact result.
func seedAndCall[V comparable](seed uint32, call func() V) V {
	engine.SetRndSeed(seed)
	return call()
}

// assertRnd compares one call's result with its expectation.
func assertRnd[V comparable](t *testing.T, got, want V, label string) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}
