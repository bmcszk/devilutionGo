package levels

import "math/bits"

// Bitset2d is a fixed-size 2-D bitmap (utils/bitset2d.hpp). The bit index is
// y*width+x, matching the upstream template.
type Bitset2d struct {
	width  int
	height int
	words  []uint64
}

// NewBitset2d returns a cleared width×height bitmap.
func NewBitset2d(width, height int) Bitset2d {
	return Bitset2d{
		width:  width,
		height: height,
		words:  make([]uint64, (width*height+63)/64),
	}
}

// Set sets (or, with value false, clears) one bit
// (utils/bitset2d.hpp:17-19).
//
//nolint:revive // value is upstream Bitset2d::set's signature, not control coupling
func (b *Bitset2d) Set(x, y int, value bool) {
	i := b.index(x, y)

	mask := uint64(1) << (i % 64)
	if value {
		b.words[i/64] |= mask

		return
	}

	b.words[i/64] &^= mask
}

// Test reports one bit (utils/bitset2d.hpp:13-15).
func (b *Bitset2d) Test(x, y int) bool {
	i := b.index(x, y)

	return b.words[i/64]&(uint64(1)<<(i%64)) != 0
}

// ResetAt clears one bit (utils/bitset2d.hpp:21-23).
func (b *Bitset2d) ResetAt(x, y int) { b.Set(x, y, false) }

// Reset clears every bit (utils/bitset2d.hpp:24-26).
func (b *Bitset2d) Reset() {
	clear(b.words)
}

// Count returns the number of set bits (utils/bitset2d.hpp:28-30).
func (b *Bitset2d) Count() int {
	count := 0
	for _, word := range b.words {
		count += bits.OnesCount64(word)
	}

	return count
}

// index mirrors Bitset2d::index (utils/bitset2d.hpp:33-36).
func (b *Bitset2d) index(x, y int) int { return y*b.width + x }
