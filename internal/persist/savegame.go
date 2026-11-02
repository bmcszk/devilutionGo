package persist

import "fmt"

// Game-save constants (Source/loadsave.cpp:55, :2763-2800).
const (
	// MaxMissilesForSaveGame caps the saved missile count so vanilla can read
	// the file (loadsave.cpp:55).
	MaxMissilesForSaveGame = 125
	// LevelsHellfire is giNumberOfLevels with Hellfire (loadsave.cpp:2779).
	LevelsHellfire = 25
	// LevelsDiablo is giNumberOfLevels without Hellfire (loadsave.cpp:2784).
	LevelsDiablo = 17
	// QuestsHellfire is giNumberQuests with Hellfire.
	QuestsHellfire = 24
	// QuestsDiablo is giNumberQuests without Hellfire.
	QuestsDiablo = 16
	// SmithPremiumHellfire is giNumberOfSmithPremiumItems with Hellfire.
	SmithPremiumHellfire = 15
	// SmithPremiumDiablo is giNumberOfSmithPremiumItems without Hellfire.
	SmithPremiumDiablo = 6
	// GameHeaderSize is the byte size of the SaveGameData header written by
	// loadsave.cpp:2765-2800: 4+1+4+4+4+4+4+1+1+4+4+4+4.
	GameHeaderSize = 43
)

// Game magic values (Source/loadsave.cpp:2765-2776), stored as the ASCII
// bytes of the four-character tokens.
const (
	// MagicRetail is "RETL": retail Diablo.
	MagicRetail = "RETL"
	// MagicHellfire is "HELF": Hellfire.
	MagicHellfire = "HELF"
	// MagicShareware is "SHAR": shareware Diablo.
	MagicShareware = "SHAR"
	// MagicSharewareHellfire is "SHLF": shareware Hellfire.
	MagicSharewareHellfire = "SHLF"
)

// LevelCounts mirrors the giNumberOfLevels/giNumberQuests/
// giNumberOfSmithPremiumItems triple SaveGameData sets (loadsave.cpp:2778-2787).
type LevelCounts struct {
	Levels       int
	Quests       int
	SmithPremium int
}

// GameMode is the game-state mode SaveGameData depends on: gbIsHellfire and
// gbIsSpawn (loadsave.cpp:2765-2776).
type GameMode struct {
	Hellfire    bool
	Spawn       bool
	Multiplayer bool
}

// GameHeader is the fixed part of SaveGameData before the per-level blocks
// (Source/loadsave.cpp:2765-2800). The mixed endianness is upstream's: the
// magic, setlevel, invflag and CharFlag are little-endian, everything else is
// big-endian.
type GameHeader struct {
	Hellfire       bool
	Spawn          bool
	SetLevel       bool
	SetLevelNum    uint32
	CurrLevel      uint32
	LevelType      uint32
	ViewX          int32
	ViewY          int32
	InventoryFlag  bool
	CharFlag       bool
	ActiveMonsters int32
	ActiveItems    int32
	ActiveMissiles int32
	ActiveObjects  int32
}

// GameMagic returns the four-character magic for the game mode
// (loadsave.cpp:2765-2776), or an error for an impossible combination.
func GameMagic(mode GameMode) (string, error) {
	switch {
	case mode.Spawn && !mode.Hellfire:
		return MagicShareware, nil
	case mode.Spawn && mode.Hellfire:
		return MagicSharewareHellfire, nil
	case !mode.Spawn && mode.Hellfire:
		return MagicHellfire, nil
	case !mode.Spawn && !mode.Hellfire:
		return MagicRetail, nil
	default:
		return "", fmt.Errorf("persist: game magic: %w", ErrInvalidGameState)
	}
}

// LevelCountsFor is the giNumberOfLevels block of SaveGameData
// (loadsave.cpp:2778-2787).
func LevelCountsFor(mode GameMode) LevelCounts {
	if mode.Hellfire {
		return LevelCounts{Levels: LevelsHellfire, Quests: QuestsHellfire, SmithPremium: SmithPremiumHellfire}
	}

	return LevelCounts{Levels: LevelsDiablo, Quests: QuestsDiablo, SmithPremium: SmithPremiumDiablo}
}

// ClampMissiles is the min(Missiles.size(), MaxMissilesForSaveGame) cap
// (loadsave.cpp:2795).
func ClampMissiles(count int) int {
	return min(count, MaxMissilesForSaveGame)
}

// Encode returns the header image.
func (h GameHeader) Encode() ([]byte, error) {
	magic, err := GameMagic(GameMode{Hellfire: h.Hellfire, Spawn: h.Spawn})
	if err != nil {
		return nil, err
	}

	var out recordWriter

	out.le32(uint32(magic[0]) | uint32(magic[1])<<8 | uint32(magic[2])<<16 | uint32(magic[3])<<24)
	out.u8(h.flagBytes()[0])
	out.be32(h.SetLevelNum)
	out.be32(h.CurrLevel)
	out.be32(h.LevelType)
	out.be32(uint32(h.ViewX))
	out.be32(uint32(h.ViewY))
	out.u8(h.flagBytes()[1])
	out.u8(h.flagBytes()[2])
	out.be32(uint32(h.ActiveMonsters))
	out.be32(uint32(h.ActiveItems))
	out.be32(uint32(ClampMissiles(int(h.ActiveMissiles))))
	out.be32(uint32(h.ActiveObjects))

	return out.buf, nil
}

// DecodeGameHeader reads a header image and returns it with the magic token.
func DecodeGameHeader(buf []byte) (GameHeader, string, error) {
	in := recordReader{buf: buf}

	magicBytes := in.bytes(4)
	magic := string(magicBytes)
	header := GameHeader{
		Hellfire:       magic == MagicHellfire || magic == MagicSharewareHellfire,
		Spawn:          magic == MagicShareware || magic == MagicSharewareHellfire,
		SetLevel:       in.u8() != 0,
		SetLevelNum:    in.be32(),
		CurrLevel:      in.be32(),
		LevelType:      in.be32(),
		ViewX:          int32(in.be32()),
		ViewY:          int32(in.be32()),
		InventoryFlag:  in.u8() != 0,
		CharFlag:       in.u8() != 0,
		ActiveMonsters: int32(in.be32()),
		ActiveItems:    int32(in.be32()),
		ActiveMissiles: int32(in.be32()),
		ActiveObjects:  int32(in.be32()),
	}

	if in.err != nil {
		return GameHeader{}, "", in.err
	}

	if !knownMagic(magic) {
		return GameHeader{}, magic, fmt.Errorf("persist: decode header %q: %w", magic, ErrInvalidGameState)
	}

	return header, magic, nil
}

// knownMagic reports whether the token is one of the four SaveGameData
// magics (loadsave.cpp:2765-2776).
func knownMagic(magic string) bool {
	switch magic {
	case MagicRetail, MagicHellfire, MagicShareware, MagicSharewareHellfire:
		return true
	default:
		return false
	}
}

// flagBytes packs setlevel, invflag and CharFlag in stream order
// (loadsave.cpp:2779, :2792-2793).
func (h GameHeader) flagBytes() [3]uint8 {
	var out [3]uint8

	if h.SetLevel {
		out[0] = 1
	}

	if h.InventoryFlag {
		out[1] = 1
	}

	if h.CharFlag {
		out[2] = 1
	}

	return out
}
