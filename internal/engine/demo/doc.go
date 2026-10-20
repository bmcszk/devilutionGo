// Package demo implements the devilutionX demo-mode stream format (version 3)
// and the record/replay drivers around it.
//
// The stream layout mirrors Source/engine/demomode.cpp:
//
//	version byte (always 3; readers reject anything larger)
//	uint32le   gSaveNumber
//	records...
//
// Each record starts with a one-byte tag. Tags 0–7 are frame records; the
// writer packs pure rendering frames into the tag byte itself
// (0x80 | progress) as upstream does, so tags 2–127 never appear on disk.
// Tags 8–14 carry input payloads little-endian, exactly the byte order
// upstream's ReadLE16/ReadLE32 produce. The format is deliberately
// unlength-prefixed so the byte stream matches a devilutionX .dmo file
// from tag 1 onward; only the version byte and save number framing is
// ours.
package demo
