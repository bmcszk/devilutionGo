package game

import (
	"strings"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// jumpLevelVerb is the one cheat line whose payload is already a number, so it
// encodes into a kernel command without an id table.
const jumpLevelVerb = "jump-level"

// DebugCommand maps one cheat command line to the kernel command it raises.
// Only the level warp is mapped: spawn-monster and spawn-item carry names, and
// naming needs the monster and item tables this layer cannot import yet, so
// those stay on the text path (DebugRouter).
func DebugCommand(line string) (kernel.Command, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[0] != jumpLevelVerb {
		return kernel.Command{}, false
	}

	level, err := parseLevelArg(fields[1:])
	if err != nil {
		return kernel.Command{}, false
	}

	return kernel.Command{Kind: kernel.KindDebugJumpLevel, A: int32(level)}, true
}
