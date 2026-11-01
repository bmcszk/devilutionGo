// Package persist ports devilutionX persistence: the packed player and item
// records (Source/pack.h), the savegame stream layout (Source/loadsave.cpp),
// the savegame codec (Source/codec.cpp), the pfile lifecycle
// (Source/pfile.cpp), and the SQLite live store pinned by ADR-003.
//
// Placement: the contract pins this package as internal/persist, a layer-3
// sibling of internal/diablo. It imports internal/diablo packages (entities,
// save) and never renders or reads input, so the layer rule in AGENTS.md
// ("internal/diablo may import engine and game") holds: persist depends on
// diablo, nothing in diablo depends on persist.
package persist

import "errors"

var (
	// ErrShortRecord reports a record decode against a truncated buffer.
	ErrShortRecord = errors.New("persist: short record")
	// ErrEmptyPassword reports a codec call without a pfile password.
	ErrEmptyPassword = errors.New("persist: empty password")
	// ErrCodecRejected reports a hero entry the codec refuses to decrypt.
	ErrCodecRejected = errors.New("persist: codec rejected the hero entry")
	// ErrSlotOutOfRange reports a hero slot outside MAX_CHARACTERS.
	ErrSlotOutOfRange = errors.New("persist: hero slot out of range")
	// ErrInvalidGameState reports a game mode the save format cannot express
	// (loadsave.cpp:2774 app_fatal("Invalid game state")).
	ErrInvalidGameState = errors.New("persist: invalid game state")
)
