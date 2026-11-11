// Tiles producer: dumps the retail level tile records the tile pipeline
// reads (devilGo-gk5e tier-1), folded in from the one-off
// internal/testdata/render/producer/go-tiles main:
//
//	go run ./tools/goldens tiles -out internal/testdata/render/tiles
//
// The test suite only ever reads the committed hex text.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// Retail MPQ paths the tile records are dumped from; the refresh runs only
// where both exist.
const (
	diabdat  = "/home/blaze/Games/Heroic/Diablo/DIABDAT.MPQ"
	hellfire = "/home/blaze/Games/Heroic/Diablo/hellfire/hellfire.mpq"
)

// MPQ member path segments, cited from dun_tile_data.cpp:72-91.
const (
	dirLevels  = `levels\`
	dirNLevels = `nlevels\`
	dirTown    = `towndata\`
	dirL1      = `l1data\`
	dirL2      = `l2data\`
	dirL3      = `l3data\`
	dirL4      = `l4data\`
	dirL5      = `l5data\`
	dirL6      = `l6data\`
)

// tileMembers lists every retail record to dump; framesHead 0 means the
// whole file, otherwise the first framesHead CEL frames.
var tileMembers = []tileMember{
	{hellfire, dirNLevels + dirTown + `town.min`, "min/town.min.hex", 0},
	{hellfire, dirNLevels + dirTown + `town.sol`, "sol/town-nlevels.sol.hex", 0},
	{diabdat, dirLevels + dirTown + `town.sol`, "sol/town.sol.hex", 0},
	{diabdat, dirLevels + dirL1 + `l1.min`, "min/l1.min.hex", 0},
	{diabdat, dirLevels + dirL2 + `l2.min`, "min/l2.min.hex", 0},
	{diabdat, dirLevels + dirL3 + `l3.min`, "min/l3.min.hex", 0},
	{diabdat, dirLevels + dirL4 + `l4.min`, "min/l4.min.hex", 0},
	{hellfire, dirNLevels + dirL5 + `l5.min`, "min/l5.min.hex", 0},
	{hellfire, dirNLevels + dirL6 + `l6.min`, "min/l6.min.hex", 0},
	{diabdat, dirLevels + dirL1 + `l1.sol`, "sol/l1.sol.hex", 0},
	{diabdat, dirLevels + dirL2 + `l2.sol`, "sol/l2.sol.hex", 0},
	{diabdat, dirLevels + dirL3 + `l3.sol`, "sol/l3.sol.hex", 0},
	{diabdat, dirLevels + dirL4 + `l4.sol`, "sol/l4.sol.hex", 0},
	{hellfire, dirNLevels + dirL5 + `l5.sol`, "sol/l5.sol.hex", 0},
	{hellfire, dirNLevels + dirL6 + `l6.sol`, "sol/l6.sol.hex", 0},
	{diabdat, dirLevels + dirL1 + `l1.til`, "til/l1.til.hex", 0},
	{diabdat, dirLevels + dirL2 + `l2.til`, "til/l2.til.hex", 0},
	{diabdat, dirLevels + dirL3 + `l3.til`, "til/l3.til.hex", 0},
	{diabdat, dirLevels + dirL4 + `l4.til`, "til/l4.til.hex", 0},
	{diabdat, dirLevels + dirL1 + `l1.cel`, "cel/l1-head.hex", 96},
}

// tileMember is one retail record to dump.
type tileMember struct {
	mpq        string
	name       string
	out        string
	framesHead int
}

// archiveCache opens each MPQ once per run.
type archiveCache map[string]*assets.MPQArchive

// runTiles dumps every member as hex text under the -out directory.
func runTiles(args []string) error {
	outRoot, err := tilesOutRoot(args)
	if err != nil {
		return err
	}

	archives := newArchiveCache()
	defer archives.closeAll()

	for _, member := range tileMembers {
		data, err := archives.memberBytes(member)
		if err != nil {
			return err
		}

		if err := writeTileHex(outRoot, member, data); err != nil {
			return err
		}
	}

	slog.Info("goldens tiles done", "files", len(tileMembers), "out", outRoot)

	return nil
}

// tilesOutRoot parses the subcommand's sole flag: `tiles -out <dir>`; the
// corpus refresh passes internal/testdata/render/tiles.
func tilesOutRoot(args []string) (string, error) {
	if len(args) == 2 && args[0] == "-out" {
		return args[1], nil
	}

	return "", errors.New("usage: goldens tiles -out <corpus dir>")
}

// newArchiveCache starts an empty per-run cache.
func newArchiveCache() archiveCache {
	return archiveCache{}
}

// closeAll releases the cached archives.
func (c archiveCache) closeAll() {
	for _, archive := range c {
		_ = archive.Close()
	}
}

// memberBytes reads one member out of its cached archive and applies the
// CEL head cut.
func (c archiveCache) memberBytes(member tileMember) ([]byte, error) {
	archive, err := c.open(member.mpq)
	if err != nil {
		return nil, err
	}

	data, err := archive.ReadFile(member.name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", member.name, err)
	}

	return trimTileHead(member, data)
}

// open returns the cached archive, opening each MPQ once per run.
func (c archiveCache) open(path string) (*assets.MPQArchive, error) {
	if archive, ok := c[path]; ok {
		return archive, nil
	}

	archive, err := assets.OpenMPQArchive(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	c[path] = archive

	return archive, nil
}

// trimTileHead applies the CEL head cut; framesHead 0 keeps the whole member.
func trimTileHead(member tileMember, data []byte) ([]byte, error) {
	if member.framesHead == 0 {
		return data, nil
	}

	head, err := celHead(data, member.framesHead)
	if err != nil {
		return nil, fmt.Errorf("trim %s: %w", member.name, err)
	}

	return head, nil
}

// celHead trims a dungeon CEL to the header, the first frame and frames
// 1..n (reencode reads srcOffsets[frame+1] for passthrough squares, so
// frame n+1's start offset must be present).
func celHead(cel []byte, framesHead int) ([]byte, error) {
	if len(cel) < (2+framesHead+1)*4 {
		return nil, fmt.Errorf("cel shorter than %d frame offsets", framesHead+1)
	}

	last := le32(cel, (2+framesHead+1)*4)

	return cel[:last], nil
}

// le32 reads one little-endian u32 (Swap32LE on LE hosts, dun_render.hpp:123).
func le32(data []byte, at int) uint32 {
	return uint32(data[at]) | uint32(data[at+1])<<8 | uint32(data[at+2])<<16 | uint32(data[at+3])<<24
}

// writeTileHex writes one member's hex dump under outRoot.
func writeTileHex(outRoot string, member tileMember, data []byte) error {
	path := filepath.Join(outRoot, member.out)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, hexDump(data), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	slog.Info("goldens tiles dumped", "name", member.name, "out", member.out, "bytes", len(data))

	return nil
}

// hexDump writes one line of up to 32 bytes, the layout the render tests read.
func hexDump(data []byte) []byte {
	buf := make([]byte, 0, len(data)*3)
	for i := 0; i < len(data); i += 32 {
		for j := i; j < len(data) && j < i+32; j++ {
			const digits = "0123456789abcdef"
			buf = append(buf, digits[data[j]>>4], digits[data[j]&0xF])
		}

		buf = append(buf, '\n')
	}

	return buf
}
