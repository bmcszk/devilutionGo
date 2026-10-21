// Package demo implements the devilutionX demo-mode stream format (version 3)
// and the record/replay drivers around it.
//
// Stream layout:
//
//	version byte (always 3; readers reject any other version)
//	uint32le gSaveNumber
//	records...
//
// Every record starts with a one-byte tag. Frame records carry a progress
// byte after the tag, input records carry a little-endian payload, and a
// render-only frame is packed into the tag byte itself (0x80 | progress).
//
// Records carry no length prefix: a recording is byte-identical to a
// devilutionX .dmo file from the first record on, and only the version byte
// and save number framing is ours.
package demo
