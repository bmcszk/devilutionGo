package assets

import (
	"fmt"
)

var (
	infravisionTRNTable = make([]byte, 256)
	stoneTRNTable       = make([]byte, 256)
	pauseTRNTable       = make([]byte, 256)
)

// TrnFrame represents a decoded TRN sprite frame (hero transformation).
// TRN files are palettes or transformation matrices used for player sprites.
// In devilutionX, TRN files contain 256-byte lookup tables for palette swaps.
type TrnFrame struct {
	// Data contains 256 entries (each entry is palette index for a pixel position)
	Data []byte
}

// TrnSpriteList represents a set of TRN sprite frames.
type TrnSpriteList struct {
	Frames []TrnFrame
}

// DecodeTRN loads and decodes a TRN file. TRN files in devilutionX
// are used as palette transformations or sprite data transformations.
// The format contains a 256-byte lookup table for palette indices.
func DecodeTRN(data []byte) (*TrnSpriteList, error) {
	if len(data) < 256 {
		return nil, fmt.Errorf("TRN data too short for 256-byte lookup table (have %d)", len(data))
	}

	// TRN contains 256 bytes of lookup table data
	trn := &TrnSpriteList{Frames: make([]TrnFrame, 0)}
	trn.Frames = append(trn.Frames, TrnFrame{
		Data: data[:256],
	})

	return trn, nil
}

// GetInfravisionTRN returns infravision TRN table
func GetInfravisionTRN() *TrnSpriteList {
	return &TrnSpriteList{Frames: []TrnFrame{{Data: infravisionTRNTable}}}
}

// GetStoneTRN returns stone TRN table
func GetStoneTRN() *TrnSpriteList {
	return &TrnSpriteList{Frames: []TrnFrame{{Data: stoneTRNTable}}}
}

// GetPauseTRN returns pause TRN table
func GetPauseTRN() *TrnSpriteList {
	return &TrnSpriteList{Frames: []TrnFrame{{Data: pauseTRNTable}}}
}
