package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// errUnaddressable marks an archive none of whose records can be named, so the
// run cannot extract anything from it. MPQ decrypts by name, so without a
// (listfile) or a committed inventory the records stay out of reach.
var errUnaddressable = errors.New("ships no (listfile) and no committed name inventory")

// extraction is what one archive's pass produced: the checkpoint state to record
// and the per-record failures, which are reported but never hide the rest of the
// archive.
type extraction struct {
	state    archiveState
	failures []string
	resumed  bool
	// unaddressable names an archive whose records cannot even be named; it is
	// reported and left alone instead of failing the run.
	unaddressable string
}

// rendered is one record decoded for output: the bytes to write, the manifest
// metadata, and the output file name inside its archive's tree.
type rendered struct {
	payload    []byte
	outputName string
	family     string
	format     string
	frames     *frameLayout
	rows       int
}

// extractOrResume extracts archive unless the journal already records this exact
// archive, in which case the recorded files are reused and nothing is decoded.
func extractOrResume(state *journal, archive string) (extraction, error) {
	if recorded, resumed := state.resume(archive); resumed {
		return extraction{state: recorded, resumed: true}, nil
	}

	extracted, err := extractArchive(archive)
	if err != nil {
		return extraction{}, err
	}

	state.Archives[archive] = extracted.state

	return extracted, nil
}

// extractArchive decodes every record of archive, mirroring the archive's inner
// paths below a directory named after the archive, and returns the state to
// checkpoint. Record failures are collected here and reported by the caller;
// only an unreadable archive is fatal.
func extractArchive(path string) (extraction, error) {
	archive, err := assets.OpenMPQArchive(path)
	if err != nil {
		return extraction{}, fmt.Errorf("open archive: %w", err)
	}

	defer func() { _ = archive.Close() }()

	names, err := archiveNames(archive, path)
	if errors.Is(err, errUnaddressable) {
		return extraction{unaddressable: filepath.Base(path)}, nil
	}

	if err != nil {
		return extraction{}, err
	}

	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	result := extractRecords(archive, path, stem, names)

	info, err := os.Stat(path)
	if err != nil {
		return extraction{}, fmt.Errorf("stat archive: %w", err)
	}

	result.state.Bytes = info.Size()
	result.state.Modified = info.ModTime().UnixNano()

	return result, nil
}

// extractRecords decodes the named records, collecting the ones that fail.
func extractRecords(archive *assets.MPQArchive, path, stem string, names []string) extraction {
	result := extraction{state: archiveState{Files: make([]fileEntry, 0, len(names))}}

	for _, name := range names {
		if isBookkeepingRecord(name) {
			continue
		}

		entry, failure := extractRecord(archive, path, stem, name)
		if failure != "" {
			result.failures = append(result.failures, failure)

			continue
		}

		result.state.Files = append(result.state.Files, entry)
	}

	slices.SortFunc(result.state.Files, compareEntries)

	return result
}

// compareEntries orders produced files by their output path, so a walk order
// change cannot reorder the manifest.
func compareEntries(a, b fileEntry) int { return strings.Compare(a.Path, b.Path) }

// extractRecord reads and decodes one archive record into a manifest entry.
func extractRecord(archive *assets.MPQArchive, archivePath, stem, name string) (fileEntry, string) {
	data, err := archive.ReadFile(name)
	if err != nil {
		return fileEntry{}, failure(archivePath, name, err)
	}

	decoded, err := render(name, data)
	if err != nil {
		return fileEntry{}, failure(archivePath, name, err)
	}

	return fileEntry{
		Path:    filepath.ToSlash(filepath.Join(stem, decoded.outputName)),
		Archive: archivePath,
		Source:  name,
		Family:  decoded.family,
		Format:  decoded.format,
		Frames:  decoded.frames,
		Rows:    decoded.rows,
		Bytes:   len(decoded.payload),
		SHA256:  digest(decoded.payload),
		payload: decoded.payload,
	}, ""
}

// writeEntires writes every extracted file below out.
func writeEntries(out string, entries []fileEntry) error {
	for _, entry := range entries {
		if err := writeOutput(filepath.Join(out, filepath.FromSlash(entry.Path)), entry.payload); err != nil {
			return err
		}
	}

	return nil
}

// failure formats one record problem in the operator's terms: archive, record,
// cause.
func failure(archivePath, name string, err error) string {
	return fmt.Sprintf("%s %s: %v", filepath.Base(archivePath), name, err)
}

// isBookkeepingRecord reports whether name is an MPQ metadata record rather than
// an asset. The (listfile) is read as the name source, never migrated as a file.
func isBookkeepingRecord(name string) bool {
	base := filepath.Base(filepath.FromSlash(strings.ReplaceAll(name, `\`, "/")))

	return strings.HasPrefix(base, "(") && strings.HasSuffix(base, ")")
}

// render decodes one record by its format. A format the package cannot decode is
// copied byte-identically and listed as raw, so the assets dir stays complete
// while the manifest says exactly which files a loader must not assume pixels
// for.
func render(name string, data []byte) (rendered, error) {
	flat := strings.ReplaceAll(name, `\`, "/")
	extension := strings.ToLower(filepath.Ext(flat))

	switch extension {
	case ".tsv":
		return renderTable(flat, data)
	case ".clx":
		if isFontSheet(flat) {
			return renderFont(flat, data)
		}

		return renderCLX(flat, data)
	case ".pcx":
		return renderPCX(flat, data)
	case ".pal":
		return renderPalette(flat, data)
	case ".trn":
		return renderTRN(flat, data)
	case ".cel", ".cl2":
		// CEL and CL2 frame dimensions live in the caller's tables upstream
		// (cel_to_clx.cpp:84-86, cl2_to_clx.cpp:66), not in the container, so
		// these two formats are extracted as-is and carry no frame layout.
		return rendered{payload: data, outputName: flat, family: familySprite, format: extension[1:]}, nil
	default:
		return rendered{payload: data, outputName: flat, family: familyRaw, format: extension[1:]}, nil
	}
}

// renderTable parses a .tsv game table and renders it as the CSV file ADR-013
// fixes for game data.
func renderTable(name string, data []byte) (rendered, error) {
	table, err := assets.Load(bytes.NewReader(data))
	if err != nil {
		return rendered{}, fmt.Errorf("parse tsv: %w", err)
	}

	csvBytes, err := tableCSV(table)
	if err != nil {
		return rendered{}, fmt.Errorf("render csv: %w", err)
	}

	return rendered{
		payload:    csvBytes,
		outputName: changesExtension(name, ".csv"),
		family:     familyTable,
		format:     "csv",
		rows:       table.Len(),
	}, nil
}

// renderCLX decodes a CLX sprite list; its frame geometry goes into the manifest.
func renderCLX(name string, data []byte) (rendered, error) {
	list, err := assets.DecodeCLX(data)
	if err != nil {
		return rendered{}, fmt.Errorf("decode clx: %w", err)
	}

	return rendered{
		payload:    data,
		outputName: name,
		family:     familySprite,
		format:     "clx",
		frames:     clxLayout(list),
	}, nil
}

// renderFont decodes a font sheet: the same CLX container, decoded as glyphs.
func renderFont(name string, data []byte) (rendered, error) {
	if _, err := assets.DecodeFontSheet(data); err != nil {
		return rendered{}, fmt.Errorf("decode font sheet: %w", err)
	}

	return rendered{
		payload:    data,
		outputName: name,
		family:     familyFont,
		format:     "clx",
		frames:     &frameLayout{Count: assets.FontGlyphCount},
	}, nil
}

// renderPCX decodes a PCX image to its single frame.
func renderPCX(name string, data []byte) (rendered, error) {
	frame, err := assets.DecodePCX(data)
	if err != nil {
		return rendered{}, fmt.Errorf("decode pcx: %w", err)
	}

	return rendered{
		payload:    data,
		outputName: name,
		family:     familySprite,
		format:     "pcx",
		frames:     &frameLayout{Count: 1, Width: frame.Width, Height: frame.Height},
	}, nil
}

// renderPalette validates a 256-colour palette record.
func renderPalette(name string, data []byte) (rendered, error) {
	if _, err := assets.LoadPaletteFromBytes(data); err != nil {
		return rendered{}, fmt.Errorf("decode palette: %w", err)
	}

	return rendered{payload: data, outputName: name, family: familyPalette, format: "pal"}, nil
}

// renderTRN validates a colour-translation record.
func renderTRN(name string, data []byte) (rendered, error) {
	if _, err := assets.DecodeTRN(data); err != nil {
		return rendered{}, fmt.Errorf("decode trn: %w", err)
	}

	return rendered{payload: data, outputName: name, family: familyTRN, format: "trn"}, nil
}

// clxLayout reports a CLX list's frame geometry: the frame count always, and the
// frame size when every frame shares one (sheets and animations often mix
// sizes, so a mixed list lists only the count).
func clxLayout(list *assets.ClxSpriteList) *frameLayout {
	layout := &frameLayout{Count: len(list.Frames), Groups: list.GroupSizes}

	if len(list.Frames) == 0 {
		return layout
	}

	width, height := list.Frames[0].Width, list.Frames[0].Height
	uniform := true

	for _, frame := range list.Frames[1:] {
		if frame.Width != width || frame.Height != height {
			uniform = false

			break
		}
	}

	if uniform {
		layout.Width, layout.Height = width, height
	}

	return layout
}

// isFontSheet reports whether a CLX record is a font sheet: upstream loads fonts
// from the "fonts" directory of its own archives (text_render.cpp, fonts\<name>.clx).
func isFontSheet(flatPath string) bool {
	dir, _, _ := strings.Cut(flatPath, "/")

	return strings.EqualFold(dir, "fonts")
}

// changesExtension replaces a path's extension.
func changesExtension(name, extension string) string {
	return strings.TrimSuffix(name, filepath.Ext(name)) + extension
}
