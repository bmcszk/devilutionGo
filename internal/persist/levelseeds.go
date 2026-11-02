package persist

// LevelSeedsEntryName is the archive entry SaveLevelSeeds writes
// (Source/loadsave.cpp:1909).
const LevelSeedsEntryName = "levelseeds"

// LevelSeed is one entry of LevelSeeds: absent means the level was never
// generated (Source/loadsave.cpp:1913-1917).
type LevelSeed struct {
	Present bool
	Seed    uint32
}

// LevelSeeds is the whole LevelSeeds array in level order
// (Source/loadsave.cpp:1907-1918).
type LevelSeeds []LevelSeed

// Encode returns the levelseeds stream for the mode's level count
// (LevelsDiablo 17 / LevelsHellfire 25, loadsave.cpp:2778-2787).
func (s LevelSeeds) Encode(mode GameMode) []byte {
	var out recordWriter

	for level := range LevelCountsFor(mode).Levels {
		seed := LevelSeed{}
		if level < len(s) {
			seed = s[level]
		}

		out.u8(seed.presentByte())

		if seed.Present {
			out.le32(seed.Seed)
		}
	}

	return out.buf
}

// DecodeLevelSeeds parses a levelseeds stream; a short stream leaves the
// remaining levels absent.
func DecodeLevelSeeds(buf []byte, mode GameMode) (LevelSeeds, error) {
	in := recordReader{buf: buf}

	levels := LevelCountsFor(mode).Levels
	seeds := make(LevelSeeds, 0, levels)

	for range levels {
		seed := LevelSeed{Present: in.u8() != 0}
		if seed.Present {
			seed.Seed = in.le32()
		}

		if in.err != nil {
			return nil, in.err
		}

		seeds = append(seeds, seed)
	}

	return seeds, nil
}

// presentByte is the 0/1 presence flag (loadsave.cpp:1913).
func (s LevelSeed) presentByte() uint8 {
	if s.Present {
		return 1
	}

	return 0
}
