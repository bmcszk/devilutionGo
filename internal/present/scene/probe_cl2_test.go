package scene_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestProbeWarriorCL2(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		t.Skip("no archives")
	}
	stack := assets.NewMPQStack()
	defer func() { _ = stack.Close() }()
	stack.Load(dirs, "hellfire", assets.HellfireMpqPriority)
	stack.Load(dirs, "DIABDAT", assets.MainMpqPriority, ".MPQ")

	for _, name := range []string{
		`plrgfx\warrior\wln\wlnst.cl2`,
		`plrgfx\warrior\whnst.cl2`,
		`plrgfx\warrior\wlnst.cl2`,
	} {
		// when reading and decoding the CL2 sheet
		data, ok := stack.ReadFile(name)
		if !ok || len(data) == 0 {
			t.Logf("%s: NOT FOUND", name)
			continue
		}
		t.Logf("%s: %d bytes", name, len(data))
		list, err := assets.DecodeCL2(data, 96, 96)
		if err != nil {
			t.Logf("%s: decode err: %v", name, err)
			continue
		}
		// then the sheet decodes with frame counts and group structure
		t.Logf("%s: %d frames, groups %v", name, len(list.Frames), list.GroupSizes)
	}
}
