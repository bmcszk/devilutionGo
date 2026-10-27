package save_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/save"
)

// Reference: diasurgical/DevilutionX Source/sha.cpp (Diablo "SHA1" — arithmetic
// right shift, sign-extending). Golden values were produced by compiling the
// verbatim upstream sha.cpp on x86-64 and dumping state (see /tmp/dg-w67z-ref),
// not by calling the Go code.

func TestSHA1CircularShiftSignExtensionQuirk(t *testing.T) {
	// given
	// Hand-derived from sha.cpp:16-24. With bit31 set the arithmetic
	// right shift sign-extends, so the high (32-bits) bits are all ones:
	//   (0x80000000, 5): 0xFFFFFFFF<<5 = 0xFFFFFFE0, 0x80000000>>27 = 0x10
	//                    0xFFFFFFE0 | 0x10 = 0xFFFFFFF0
	//   (0xF0F0F0F0, 3): 0xFFFFFFFF<<3 = 0xFFFFFFF8, 0xF0F0F0F0>>29 = 0x7
	//                    0xFFFFFFF8 | 0x7 = 0xFFFFFFFF
	// Clear bit31 falls through to the plain rotate:
	//   (0x12345678, 7): 0x12345678<<7 = 0x91A2B3C00 truncated to 0x1A2B3C00,
	//                    0x12345678>>25 = 0x9, or = 0x1A2B3C09
	//   (0x00000001, 1): 0x1<<1 | 0x1>>31 = 0x2
	tests := []struct {
		name string
		word uint32
		bits uint
		want uint32
	}{
		{"sign-extended high bit", 0x80000000, 5, 0xFFFFFFF0},
		{"sign-extended pattern", 0xF0F0F0F0, 3, 0xFFFFFFFF},
		{"plain rotate", 0x12345678, 7, 0x1A2B3C09},
		{"single bit", 0x00000001, 1, 0x00000002},
	}
	// when
	shifts := make([]uint32, len(tests))
	for i, tc := range tests {
		shifts[i] = save.SHA1CircularShift(tc.word, tc.bits)
	}

	// then
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if shifts[i] != tc.want {
				t.Errorf("save.SHA1CircularShift(%08X, %d) = %08X, want %08X", tc.word, tc.bits, shifts[i], tc.want)
			}
		})
	}
}

func TestSHA1CalculateGoldenFromFreshContext(t *testing.T) {
	// given
	// Initial IV per sha.h:10 (0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0).
	// "abc" followed by zero bytes, loaded as one little-endian word per 4 bytes.
	var block [save.SHA1BlockSize]uint32

	block[0] = 0x00636261
	context := save.NewSHA1Context()
	// when
	save.SHA1Calculate(&context, &block)
	// then
	want := [5]uint32{0x45A07558, 0xD5B60157, 0xA31F2F7B, 0x0BCB00D4, 0xB7E9798B}
	if context.State != want {
		t.Errorf("diablo-SHA1(abc block) = %08X, want %08X", context.State, want)
	}
	// Standard SHA1("abc") = A9993E36 4706816A BA3E2571 7850C26C 9CD0D89D;
	// the mismatch above is exactly the sha.cpp:14 sign-extension deviation.
}

func TestSHA1CalculatePasswordBlockGolden(t *testing.T) {
	// given
	// save.CodecInitKey first step (codec.cpp:25-38): "password" repeated to 64
	// bytes as LE words, one message block from the fresh IV.
	var block [save.SHA1BlockSize]uint32
	for i := range block {
		block[i] = [...]uint32{0x73736170, 0x64726F77}[i%2]
	}

	context := save.NewSHA1Context()
	// when
	save.SHA1Calculate(&context, &block)
	// then
	want := [5]uint32{0x32022197, 0xA96C5CFB, 0x7DFBE1EF, 0xF4685499, 0x8ED39B25}
	if context.State != want {
		t.Errorf("diablo-SHA1(password block) = %08X, want %08X", context.State, want)
	}
}
