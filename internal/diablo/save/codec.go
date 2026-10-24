package save

import "encoding/binary"

// codecBlockSizeBytes is BlockSizeBytes in codec.cpp:16.
// codecSignatureSize is SignatureSize in codec.cpp:17.
const (
	codecBlockSizeBytes = SHA1BlockSize * 4
	codecSignatureSize  = 8
)

// codecPassword is the fixed multiplayer save-game password (pfile.cpp:60-63);
// the upstream codec_encode/codec_decode pair always receives it (codec.cpp:140/92).
const codecPassword = "xrgyrkj1"

// codecKeyMask holds the obfuscation constants xor-ed into the password
// digest by codecInitKey (codec.cpp:40-42), in source order.
var codecKeyMask = [SHA1BlockSize]uint32{
	2908958655, 4146550480, 658981742, 1113311088, 3927878744, 679301322, 1760465731, 3305370375,
	2269115995, 3928541685, 580724401, 2607446661, 2233092279, 2416822349, 4106933702, 3046442503,
}

// codecInitKey stretches a password into the codec's SHA1 state
// (codec.cpp:19-53): the password repeats to fill one 64-byte block of
// little-endian words, is hashed, the digest xor-masks an obfuscation
// constant block, and that block is hashed into the returned context.
func codecInitKey(password string) sHA1Context {
	var pw [SHA1BlockSize]uint32
	j := 0
	for i := range pw {
		if j >= len(password) {
			j = 0
		}
		pw[i] = codecLoadLE32(password, j)
		j += 4
	}

	context := newSHA1Context()
	sHA1Calculate(&context, &pw)
	digest := sHA1Result(&context)

	var key [SHA1BlockSize]uint32
	for i := range key {
		key[i] = codecKeyMask[i] ^ digest[(i+3)%SHA1HashSize]
	}

	context = newSHA1Context()
	sHA1Calculate(&context, &key)
	return context
}

// codecGetEncodedLen pads to a full block and adds the signature
// (codec.cpp:132-138).
func codecGetEncodedLen(srcBytes int) int {
	if srcBytes%codecBlockSizeBytes != 0 {
		srcBytes += codecBlockSizeBytes - srcBytes%codecBlockSizeBytes
	}
	return srcBytes + codecSignatureSize
}

// codecEncode encrypts the first size bytes of data in place, zero-pads the
// trailing chunk and appends the signature (codec.cpp:140-173). data must
// hold codecGetEncodedLen(size) bytes; size64 must equal it, matching the
// upstream app_fatal guard.
func codecEncode(data []byte, size int, size64 int) {
	if size64 != codecGetEncodedLen(size) {
		panic("Invalid encode parameters")
	}
	context := codecInitKey(codecPassword)

	lastChunk := 0
	for offset := 0; size != 0; offset += codecBlockSizeBytes {
		chunk := min(size, codecBlockSizeBytes)
		block := codecLoadBlock(data[offset : offset+chunk])
		digest := sHA1Result(&context)
		sHA1Calculate(&context, &block)
		xorBlock(digest, &block)
		codecStoreBlock(data[offset:offset+codecBlockSizeBytes], block)
		lastChunk = chunk
		size -= chunk
	}

	digest := sHA1Result(&context)
	sig := data[len(data)-codecSignatureSize:]
	binary.LittleEndian.PutUint32(sig[0:4], digest[0])
	sig[4] = 0
	sig[5] = byte(lastChunk)
	sig[6] = 0
	sig[7] = 0
}

// codecDecode decrypts data in place and returns the original payload
// length, or 0 on any failure (codec.cpp:92-130). data must hold exactly
// codecGetEncodedLen(payload) bytes.
func codecDecode(data []byte, password string) int {
	context := codecInitKey(password)
	size := len(data)
	if size <= codecSignatureSize {
		return 0
	}
	size -= codecSignatureSize
	if size%SHA1BlockSize != 0 {
		return 0
	}
	for offset := 0; offset < size; offset += codecBlockSizeBytes {
		block := codecLoadBlock(data[offset : offset+codecBlockSizeBytes])
		digest := sHA1Result(&context)
		xorBlock(digest, &block)
		sHA1Calculate(&context, &block)
		codecStoreBlock(data[offset:offset+codecBlockSizeBytes], block)
	}

	sig := data[size:]
	if sig[4] > 0 {
		return 0
	}
	checksum := binary.LittleEndian.Uint32(sig[0:4])
	digest := sHA1Result(&context)
	if checksum != digest[0] {
		return 0
	}
	return size + int(sig[5]) - codecBlockSizeBytes
}

func xorBlock(shaResult [SHA1HashSize]uint32, out *[SHA1BlockSize]uint32) {
	for i := range out {
		out[i] ^= shaResult[i%SHA1HashSize]
	}
}

// codecLoadLE32 reads four password bytes zero-padded past the end as one
// little-endian word (endian_read.hpp:24 LoadLE32). Upstream indexes past
// the terminator; production passwords are multiples of 4 bytes, so the
// zero pad only covers short final words.
func codecLoadLE32(password string, offset int) uint32 {
	var word [4]byte
	for i := range word {
		if offset+i < len(password) {
			word[i] = password[offset+i]
		}
	}
	return binary.LittleEndian.Uint32(word[:])
}

// codecLoadBlock little-endian loads a 64-byte chunk, zero-padded to a full
// block — byte-exact with upstream's LE-host ByteSwapBlock no-op
// (codec.cpp:66-70, endian_swap.hpp:31 SDL_Swap32LE).
func codecLoadBlock(chunk []byte) [SHA1BlockSize]uint32 {
	var block [SHA1BlockSize]uint32
	for i := range block {
		if i*4 >= len(chunk) {
			break
		}
		var word [4]byte
		copy(word[:], chunk[i*4:])
		block[i] = binary.LittleEndian.Uint32(word[:])
	}
	return block
}

// codecStoreBlock little-endian stores a full block back into the buffer,
// the LE-host mirror of codecLoadBlock.
func codecStoreBlock(dst []byte, block [SHA1BlockSize]uint32) {
	for i := range block {
		binary.LittleEndian.PutUint32(dst[i*4:(i+1)*4], block[i])
	}
}
