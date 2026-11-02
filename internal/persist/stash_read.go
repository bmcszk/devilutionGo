package persist

// readStashPages reads the non-empty page blocks of a stash stream
// (loadsave.cpp:2745-2757).
func readStashPages(in *recordReader, count uint32) []StashPage {
	pages := make([]StashPage, 0, count)

	for range count {
		page := StashPage{Index: in.le32()}

		for cell := range page.Cells {
			page.Cells[cell] = in.le16()
		}

		if in.err != nil {
			return pages
		}

		pages = append(pages, page)
	}

	return pages
}

// readStashItems reads the fixed-stride SaveItem records of a stash stream
// (loadsave.cpp:2760-2765).
func readStashItems(in *recordReader, count uint32, itemSize int) [][]byte {
	items := make([][]byte, 0, count)

	for range count {
		item := in.bytes(itemSize)
		if in.err != nil {
			return items
		}

		items = append(items, item)
	}

	return items
}
