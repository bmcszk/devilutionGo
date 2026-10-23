// Package save implements the Diablo savegame codec and its "SHA1" hash.
package save

// SHA1HashSize is the number of 32-bit words in a digest.
// SHA1BlockSize is the number of 32-bit words in a message block.
const (
	SHA1HashSize  = 5
	SHA1BlockSize = 16
)

// sHA1Context is the Diablo "SHA1" state: digest plus the last message block.
type sHA1Context struct {
	State  [SHA1HashSize]uint32
	Buffer [SHA1BlockSize]uint32
}

// newSHA1Context returns the standard SHA-1 initial vector.
func newSHA1Context() sHA1Context {
	return sHA1Context{State: [SHA1HashSize]uint32{
		0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0,
	}}
}

// sHA1CircularShift is the Diablo-"SHA1" left rotate. The original treated the
// word as signed and used arithmetic right shifts, so words with bit31 set get
// their high bits sign-extended to one instead of wrapping around — the
// deviation from real SHA-1.
func sHA1CircularShift(word uint32, bits uint) uint32 {
	if word&(1<<31) != 0 {
		return (0xFFFFFFFF << bits) | (word >> (32 - bits))
	}
	return (word << bits) | (word >> (32 - bits))
}

func sha1ProcessMessageBlock(context *sHA1Context) {
	var w [80]uint32
	copy(w[:], context.Buffer[:])
	for i := 16; i < 80; i++ {
		w[i] = w[i-16] ^ w[i-14] ^ w[i-8] ^ w[i-3]
	}

	a := context.State[0]
	b := context.State[1]
	c := context.State[2]
	d := context.State[3]
	e := context.State[4]

	for i := 0; i < 20; i++ {
		temp := sHA1CircularShift(a, 5) + ((b & c) | (^b & d)) + e + w[i] + 0x5A827999
		e = d
		d = c
		c = sHA1CircularShift(b, 30)
		b = a
		a = temp
	}

	for i := 20; i < 40; i++ {
		temp := sHA1CircularShift(a, 5) + (b ^ c ^ d) + e + w[i] + 0x6ED9EBA1
		e = d
		d = c
		c = sHA1CircularShift(b, 30)
		b = a
		a = temp
	}

	for i := 40; i < 60; i++ {
		temp := sHA1CircularShift(a, 5) + ((b & c) | (b & d) | (c & d)) + e + w[i] + 0x8F1BBCDC
		e = d
		d = c
		c = sHA1CircularShift(b, 30)
		b = a
		a = temp
	}

	for i := 60; i < 80; i++ {
		temp := sHA1CircularShift(a, 5) + (b ^ c ^ d) + e + w[i] + 0xCA62C1D6
		e = d
		d = c
		c = sHA1CircularShift(b, 30)
		b = a
		a = temp
	}

	context.State[0] += a
	context.State[1] += b
	context.State[2] += c
	context.State[3] += d
	context.State[4] += e
}

// sHA1Result copies the digest without finalizing; the context stays usable
// for further sHA1Calculate calls.
func sHA1Result(context *sHA1Context) [SHA1HashSize]uint32 {
	return context.State
}

// sHA1Calculate absorbs one message block and transforms the state.
func sHA1Calculate(context *sHA1Context, data *[SHA1BlockSize]uint32) {
	copy(context.Buffer[:], data[:])
	sha1ProcessMessageBlock(context)
}
