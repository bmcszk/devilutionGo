// Package demo implements the devilutionX demo-mode stream format and the
// recorder and playback drivers built on it.
//
// Stream layout:
//
//	version byte (always 3; readers reject any other version)
//	uint32le save number
//	records...
//
// Every record starts with a one-byte tag. Frame records carry a progress
// byte after the tag, input records carry a little-endian payload, and a
// render-only frame is packed into the tag byte itself (0x80 | progress).
//
// Records carry no length prefix and are byte-identical to the records a
// devilutionX .dmo file holds. Only the framing differs: ours is the version
// byte and save number, while devilutionX inserts a settings block between
// the save number and the first record.
package demo
