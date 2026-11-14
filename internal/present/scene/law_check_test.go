package scene_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// Town palette index 83, 102, 110 = the top terrain colors from the surface histogram.
// What RGB does shade 0 expansion give them?
func TestShadeZeroBright(t *testing.T) {
	// given the retail archives and town palette
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	stack := assets.NewMPQStack()
	defer func() { _ = stack.Close() }()
	stack.Load(dirs, "hellfire", assets.HellfireMpqPriority)
	stack.Load(dirs, "DIABDAT", assets.MainMpqPriority, ".MPQ")

	palData, ok := stack.ReadFile("levels/towndata/town.pal")
	if !ok || len(palData) == 0 {
		t.Fatalf("read %s: not found in archives", "levels/towndata/town.pal")
	}
	pal, err := assets.LoadPaletteFromBytes(palData)
	if err != nil {
		t.Fatal(err)
	}
	tables := light.MakeLightTables(light.DTYPE_TOWN)
	t.Logf("tables[0][83]=%d tables[0][102]=%d tables[0][110]=%d",
		tables.Tables[0][83], tables.Tables[0][102], tables.Tables[0][110])
	t.Logf("pal[83]=%v pal[102]=%v pal[110]=%v", pal[83], pal[102], pal[110])
	t.Logf("pal[64]=%v (hero synthetic index)", pal[64])

	// when the palette is expanded through shade 0
	state := render.NewPaletteState(*pal, tables, assets.BlendingFor(0))
	t.Logf("Variants[0][83]=%v Variants[0][102]=%v", state.Variants[0][83], state.Variants[0][102])

	// then the frame composes with readable brightness
	composed, err := scene.ComposeTownFrameWithHero(dirs)
	if err != nil {
		t.Fatal(err)
	}
	frame := composed.Surface.ExpandTile(&composed.State.Variants, 0)
	nonBlack := countNonBlack(frame)
	t.Logf("nonBlack=%.1f%%", nonBlack*100/float64(len(frame.Pixels)))
}

func countNonBlack(frame render.Frame) float64 {
	n := 0
	for _, px := range frame.Pixels {
		if px.R != 0 || px.G != 0 || px.B != 0 {
			n++
		}
	}
	return float64(n)
}
