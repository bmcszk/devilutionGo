// One-off offline fixture producer (devilGo-gk5e tier-1): dumps the retail
// level tile records the tile pipeline reads. Run ONCE, offline:
//
//	go run ./internal/testdata/render/producer/go-tiles
//
// The test suite only ever reads the committed hex text.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const (
	diabdat  = "/home/blaze/Games/Heroic/Diablo/DIABDAT.MPQ"
	hellfire = "/home/blaze/Games/Heroic/Diablo/hellfire/hellfire.mpq"
	outRoot  = "internal/testdata/render/tiles"
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

// archives caches one open MPQ per path.
var archives = map[string]*assets.MPQArchive{}

// member is one retail record to dump; framesHead 0 means the whole file,
// otherwise the first framesHead CEL frames.

var members = []member{
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

// member is one retail record to dump; framesHead 0 means the whole file,
// otherwise the first framesHead CEL frames.
type member struct {
	mpq        string
	name       string
	out        string
	framesHead int
}

// main dumps every member; failure panics, this is a one-off tool.
func main() {
	defer closeAll()

	for _, m := range members {
		data := trim(m, read(m))
		path := filepath.Join(outRoot, m.out)

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			panic(err)
		}

		if err := os.WriteFile(path, hexDump(data), 0o644); err != nil {
			panic(err)
		}

		slog.Info("dumped", "name", m.name, "out", m.out, "bytes", len(data))
	}
}

// trim applies the CEL head cut.
func trim(m member, data []byte) []byte {
	if m.framesHead == 0 {
		return data
	}

	head, err := celHead(data, m.framesHead)
	if err != nil {
		panic(err)
	}

	return head
}

// read fetches one member out of its archive.
func read(m member) []byte {
	data, err := archive(m.mpq).ReadFile(m.name)
	if err != nil {
		panic(err)
	}

	return data
}

// archive opens each MPQ once.
func archive(path string) *assets.MPQArchive {
	if a, ok := archives[path]; ok {
		return a
	}

	a, err := assets.OpenMPQArchive(path)
	if err != nil {
		panic(fmt.Errorf("open %s: %w", path, err))
	}

	archives[path] = a

	return a
}

// closeAll releases the archives.
func closeAll() {
	for _, a := range archives {
		_ = a.Close()
	}
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
