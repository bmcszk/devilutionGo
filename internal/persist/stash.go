package persist

import "fmt"

// Stash-save constants (Source/loadsave.cpp:2114-2115, :2415, :2714-2720).
const (
	// StashVersion is the current stash save version byte.
	StashVersion = 0
	// StashPageCells is 10*10 cells per stash page.
	StashPageCells = 100
	// StashEntrySingle is the stash entry name of single player games.
	StashEntrySingle = "spstashitems"
	// StashEntryMulti is the stash entry name of multiplayer games.
	StashEntryMulti = "mpstashitems"
	// ItemSaveSizeDiablo is DiabloItemSaveSize: the fixed SaveItem record
	// length without Hellfire.
	ItemSaveSizeDiablo = 368
	// ItemSaveSizeHellfire is HellfireItemSaveSize.
	ItemSaveSizeHellfire = 372
)

// StashPage is one saved stash page: its index plus the 10x10 cell grid of
// item indices (loadsave.cpp:2748-2755).
type StashPage struct {
	Index uint32
	Cells [StashPageCells]uint16
}

// StashSave is the stash stream SaveStash writes (loadsave.cpp:2712-2765):
// version, gold, the non-empty page grids, the item list and the current page.
type StashSave struct {
	// Gold is Stash.gold.
	Gold uint32
	// Pages holds only the pages containing at least one item.
	Pages []StashPage
	// Items holds the raw SaveItem records of Stash.stashList.
	Items [][]byte
	// CurrentPage is Stash.GetPage().
	CurrentPage uint32
}

// StashEntryName is the stash entry name for a game mode
// (loadsave.cpp:2714-2720).
func StashEntryName(mode GameMode) string {
	if mode.Multiplayer {
		return StashEntryMulti
	}

	return StashEntrySingle
}

// ItemSizeFor is the SaveItem record length for a game mode
// (loadsave.cpp:2113-2121).
func ItemSizeFor(mode GameMode) int {
	if mode.Hellfire {
		return ItemSaveSizeHellfire
	}

	return ItemSaveSizeDiablo
}

// Encode returns the stash stream; items must be itemSize bytes long.
func (s StashSave) Encode(itemSize int) ([]byte, error) {
	for i, item := range s.Items {
		if len(item) != itemSize {
			return nil, fmt.Errorf("persist: stash item %d len %d != %d: %w", i, len(item), itemSize,
				ErrShortRecord)
		}
	}

	var out recordWriter

	out.u8(StashVersion)
	out.le32(s.Gold)
	out.le32(uint32(len(s.Pages)))

	for _, page := range s.Pages {
		out.le32(page.Index)

		for _, cell := range page.Cells {
			out.le16(cell)
		}
	}

	out.le32(uint32(len(s.Items)))

	for _, item := range s.Items {
		out.bytes(item)
	}

	out.le32(s.CurrentPage)

	return out.buf, nil
}

// DecodeStashSave parses a stash stream; itemSize is the mode's SaveItem
// record length because the stream does not carry the Hellfire flag.
func DecodeStashSave(buf []byte, itemSize int) (StashSave, error) {
	in := recordReader{buf: buf}

	version := in.u8()
	if version != StashVersion {
		return StashSave{}, fmt.Errorf("persist: stash version %d: %w", version, ErrUnsupportedVersion)
	}

	save := StashSave{Gold: in.le32()}

	pageCount := in.le32()
	if pageCount > uint32(len(buf)) {
		return StashSave{}, fmt.Errorf("persist: stash page count %d: %w", pageCount, ErrShortRecord)
	}

	save.Pages = readStashPages(&in, pageCount)

	itemCount := in.le32()
	if itemCount > uint32(len(buf)) {
		return StashSave{}, fmt.Errorf("persist: stash item count %d: %w", itemCount, ErrShortRecord)
	}

	save.Items = readStashItems(&in, itemCount, itemSize)
	save.CurrentPage = in.le32()

	if in.err != nil {
		return StashSave{}, fmt.Errorf("persist: stash stream: %w", in.err)
	}

	return save, nil
}
