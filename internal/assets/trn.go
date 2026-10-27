package assets

import "fmt"

// trnSize is the fixed TRN table length (lighting.cpp:37).
const trnSize = 256

const (
	// MTCounslr is the Councilor monster type.
	MTCounslr MonsterType = iota
	// MTMagistr is the Magistr monster type.
	MTMagistr
	// MTCabalist is the Cabalist monster type.
	MTCabalist
	// MTAdvocate is the Advocate monster type.
	MTAdvocate
)

// TRN is a 256-entry palette translation table: dst = trn[src]. Every colour
// effect in the engine is an index->index lookup (utils/trn.hpp, clx_render.cpp:393).
type TRN [256]byte

// MonsterType is the subset of the monster enum whose animation 1 has no TRN
// (monster.cpp:181-184). Ordinals stay local so this package never imports
// internal/diablo.
type MonsterType int

// TRNRegistry holds the three engine-wide translation tables loaded at start.
type TRNRegistry struct {
	Infravision TRN
	Stone       TRN
	Pause       TRN
}

// DecodeTRN parses a TRN file; it must be exactly 256 bytes.
func DecodeTRN(b []byte) (TRN, error) {
	var trn TRN

	if len(b) != trnSize {
		return trn, fmt.Errorf("TRN: expected %d bytes, got %d", trnSize, len(b))
	}

	copy(trn[:], b)

	return trn, nil
}

// ApplyTRN rewrites palette indices in place through the table, the Go
// equivalent of ClxApplyTrans (clx_render.cpp:393).
func ApplyTRN(pixels []byte, t TRN) {
	for i, idx := range pixels {
		pixels[i] = t[idx]
	}
}

// LoadTrns loads the engine tables in the order lighting.cpp:214-219 requests
// them, stopping at the first failure with the path in the error.
func LoadTrns(src AssetSource) (*TRNRegistry, error) {
	paths := []string{`plrgfx\infra.trn`, `plrgfx\stone.trn`, `gendata\pause.trn`}

	tables := make([]TRN, len(paths))
	for i, path := range paths {
		trn, err := loadRequiredTRN(src, path)
		if err != nil {
			return nil, err
		}

		tables[i] = trn
	}

	return &TRNRegistry{Infravision: tables[0], Stone: tables[1], Pause: tables[2]}, nil
}

// ClassTRN loads the class tint table for a player class (trn.cpp:29-45).
// A missing file is not an error: ok is false and the zero TRN is returned.
func (*TRNRegistry) ClassTRN(src AssetSource, trnName string) (TRN, bool, error) {
	return loadOptionalTRN(src, fmt.Sprintf(`plrgfx\%s.trn`, trnName))
}

// PlayerGraphicTRN loads the per-graphic tint that sits next to the CL2
// (trn.cpp:48-60); a missing file is not an error.
func (*TRNRegistry) PlayerGraphicTRN(src AssetSource, baseName string) (TRN, bool, error) {
	return loadOptionalTRN(src, baseName+".trn")
}

// MonsterTRN loads a monster type's tint table (monster.cpp:173-180); the file
// is required, so a missing one is an error.
func (*TRNRegistry) MonsterTRN(src AssetSource, trnFile string) (TRN, error) {
	return loadRequiredTRN(src, fmt.Sprintf(`monsters\%s.trn`, trnFile))
}

// MonsterTRNSanitized applies the InitMonsterTRN quirk: 255 means "map to
// black", so every 255 entry becomes 0 before the table is used
// (monster.cpp:177).
func MonsterTRNSanitized(t TRN) TRN {
	for i, v := range t {
		if v == 255 {
			t[i] = 0
		}
	}

	return t
}

// MonsterTRNSkipAnim reports whether a monster type skips TRN application for
// an animation: animation 1 of the Councilor family has none
// (monster.cpp:181-184).
func MonsterTRNSkipAnim(monsterType MonsterType, animIndex int) bool {
	if animIndex != 1 {
		return false
	}

	switch monsterType {
	case MTCounslr, MTMagistr, MTCabalist, MTAdvocate:
		return true
	}

	return false
}

func loadRequiredTRN(src AssetSource, path string) (TRN, error) {
	data, ok := src.ReadFile(path)
	if !ok {
		return TRN{}, fmt.Errorf("TRN: asset %s not found", path)
	}

	trn, err := DecodeTRN(data)
	if err != nil {
		return TRN{}, fmt.Errorf("TRN %s: %w", path, err)
	}

	return trn, nil
}

func loadOptionalTRN(src AssetSource, path string) (TRN, bool, error) {
	data, ok := src.ReadFile(path)
	if !ok {
		return TRN{}, false, nil
	}

	trn, err := DecodeTRN(data)
	if err != nil {
		return TRN{}, false, fmt.Errorf("TRN %s: %w", path, err)
	}

	return trn, true, nil
}
