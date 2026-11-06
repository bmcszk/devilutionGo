package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// Asset element widths, i.e. the byte width of one archive member value: .til
// and .min ids are uint16 (levels/dun_tile_data.hpp:56-61), .sol is
// TileProperties[MAXTILES], one property byte per world tile
// (levels/dun_tile.hpp:105, levels/dun_tile_data.hpp:81).
const (
	assetUint16Bytes = 2
	assetUint8Bytes  = 1
)

// hiveAssets is the nlevels\l6data set CreateL3Dungeon's nest branch loads
// (levels/diablo.cpp:1409-1413, levels/dun_tile_data.cpp:88-89,154).
var hiveAssets = []assetMember{
	{"l6.til", `nlevels\l6data\l6.til`, 4, assetUint16Bytes},
	{"l6.min", `nlevels\l6data\l6.min`, 1, assetUint16Bytes},
	{"l6.sol", `nlevels\l6data\l6.sol`, 1, assetUint8Bytes},
}

// assetMember is one archive member the hive generator installs and the CSV
// shape it is committed in: `count,cols` then count*cols decimal values.
type assetMember struct {
	name   string
	member string
	cols   int
	width  int
}

// runAssets converts the hive level's archive members into the committed asset
// CSVs, so the fixtures come from the retail archive instead of a hand edit.
func runAssets(archivePath, corpusDir string) error {
	archive, err := assets.OpenMPQArchive(archivePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", archivePath, err)
	}
	defer func() { _ = archive.Close() }()

	for _, member := range hiveAssets {
		if err := convertAsset(archive, member, corpusDir); err != nil {
			return err
		}
	}

	slog.Info("dunconv assets done", "files", len(hiveAssets), "corpus", corpusDir)

	return nil
}

// convertAsset reads one member and writes its CSV.
func convertAsset(archive *assets.MPQArchive, member assetMember, corpusDir string) error {
	data, err := archive.ReadFile(member.member)
	if err != nil {
		return fmt.Errorf("read %s: %w", member.member, err)
	}

	values, err := assetValues(data, member)
	if err != nil {
		return err
	}

	path := filepath.Join(corpusDir, member.name+".csv")
	if err := os.WriteFile(path, encodeAssetCSV(member, values), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// assetValues decodes the member into its element values, one per CSV value.
func assetValues(data []byte, member assetMember) ([]uint16, error) {
	width := member.width
	if len(data)%width != 0 {
		return nil, fmt.Errorf("%s holds %d bytes, not a multiple of %d", member.member, len(data), width)
	}

	values := make([]uint16, len(data)/width)
	for i := range values {
		if width == assetUint16Bytes {
			values[i] = binary.LittleEndian.Uint16(data[2*i:])
		} else {
			values[i] = uint16(data[i])
		}
	}

	return values, nil
}

// encodeAssetCSV writes the two-record asset shape: `count,cols`, then the
// values. cols is the member's element count per row (4 for .til, 1 otherwise).
func encodeAssetCSV(member assetMember, values []uint16) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(len(values) / member.cols), strconv.Itoa(member.cols)})

	fields := make([]string, len(values))
	for i, value := range values {
		fields[i] = strconv.Itoa(int(value))
	}
	_ = writer.Write(fields)
	writer.Flush()

	return out.Bytes()
}
