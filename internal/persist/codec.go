package persist

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/save"
)

// Hero-file passwords (Source/pfile.cpp:60-63).
const (
	// PasswordSpawnSingle is PASSWORD_SPAWN_SINGLE.
	PasswordSpawnSingle = "adslhfb1"
	// PasswordSpawnMulti is PASSWORD_SPAWN_MULTI.
	PasswordSpawnMulti = "lshbkfg1"
	// PasswordSingle is PASSWORD_SINGLE.
	PasswordSingle = "xrgyrkj1"
	// PasswordMulti is PASSWORD_MULTI.
	PasswordMulti = "szqnlsk1"
)

// MaxCharacters is MAX_CHARACTERS (Source/pfile.h:24): the hero-slot count.
const MaxCharacters = 99

// HeroPassword is pfile_get_password (Source/pfile.cpp:638-643).
func HeroPassword(mode SaveNamespace) string {
	if mode.Spawn {
		if mode.Multiplayer {
			return PasswordSpawnMulti
		}

		return PasswordSpawnSingle
	}

	if mode.Multiplayer {
		return PasswordMulti
	}

	return PasswordSingle
}

// HeroEntryLen is codec_get_encoded_len (codec.cpp:132-138) for a hero file:
// the record is padded to a 32-byte block and gets the 8-byte signature.
func HeroEntryLen(payload int) int {
	return save.CodecGetEncodedLen(payload)
}

// EncodeHeroEntry runs the codec over a hero-file record
// (pfile.cpp:177 pfile_write_hero -> codec_encode).
func EncodeHeroEntry(payload []byte, password string) ([]byte, error) {
	if len(password) == 0 {
		return nil, fmt.Errorf("persist: encode hero: %w", ErrEmptyPassword)
	}

	encoded := make([]byte, save.CodecGetEncodedLen(len(payload)))
	copy(encoded, payload)

	if err := save.CodecEncodeWithPassword(encoded, len(payload), len(encoded), password); err != nil {
		return nil, fmt.Errorf("persist: encode hero: %w", err)
	}

	return encoded, nil
}

// DecodeHeroEntry runs the codec backwards (pfile.cpp:628
// pfile_read_hero -> codec_decode) and returns the payload.
func DecodeHeroEntry(encoded []byte, password string) ([]byte, error) {
	decodedLen := save.CodecDecode(encoded, password)
	if decodedLen <= 0 || decodedLen > len(encoded) {
		return nil, fmt.Errorf("persist: decode hero: %w", ErrCodecRejected)
	}

	payload := make([]byte, decodedLen)
	copy(payload, encoded[:decodedLen])

	return payload, nil
}
