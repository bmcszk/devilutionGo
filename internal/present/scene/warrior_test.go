package scene_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestWarriorStandFrameOverTown is the s7x0 acceptance gate: the warrior
// STAND frame 0 decodes from the real archives and draws over the town
// scene at the hero tile through the Presenter pipeline.
func TestWarriorStandFrameOverTown(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 || !scene.ArchivesPresent(dirs) {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	// when the warrior is drawn over the town scene
	composed, err := scene.ComposeTownFrameWithHero(dirs)
	if err != nil {
		t.Fatalf("ComposeTownFrameWithHero: %v", err)
	}

	// then the frame carries hero pixels at the expected screen rect
	if !scene.HasHeroAt(composed, scene.HeroTile()) {
		t.Fatal("composed frame shows no hero pixels at the hero tile")
	}

	// and the presented frame is bright enough to read: the non-black pixels'
	// mean channel value at the boot brightness (60) must stay above 55, so a
	// lost tone-map or a dark-palette regression can never silently return
	mean, nonBlack := scene.MeanBrightness(composed)
	if nonBlack < 0.5 {
		t.Fatalf("only %.1f%% of the frame is non-black, want >= 50%%", nonBlack*100)
	}
	if mean < 55 {
		t.Fatalf("non-black mean brightness = %.1f, want >= 55 at the boot brightness", mean)
	}

	// and the PNG artifact lands for the parent's visual check
	writeSceneArtifact(t, composed, "town_hero.png")
}

// TestComposeTownFrameAtBrightnessIsMonotone is the Gamma row's render oracle
// (Options slice S4b): the composed frame's non-black mean brightness rises
// strictly with the Graphics/"Brightness Correction" level at a FIXED hero tile,
// so a lost tone-map shows up as repeated stats. Reference:
// ApplyGlobalBrightness (Source/engine/palette.cpp:253-261) and the Gamma
// slider's 0..100 range (Source/gamemenu.cpp:152-156).
func TestComposeTownFrameAtBrightnessIsMonotone(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 || !scene.ArchivesPresent(dirs) {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	hero := scene.HeroTile()
	levels := []int{0, 50, 100}

	// when composing the same tile at each slider extreme and its middle
	means := composeAtEachLevel(t, dirs, hero, levels)

	// then every level is brighter than the one below it: three DISTINCT stats
	assertRisingMeans(t, levels, means)
}

// composeAtEachLevel composes the fixed hero tile at every gamma level and
// returns the measured means.
func composeAtEachLevel(t *testing.T, dirs []string, hero render.TilePos, levels []int) []float64 {
	t.Helper()

	means := make([]float64, 0, len(levels))

	for _, level := range levels {
		composed, err := scene.ComposeTownFrameAtBrightness(dirs, hero, level)
		if err != nil {
			t.Fatalf("ComposeTownFrameAtBrightness(%d): %v", level, err)
		}

		means = append(means, composeBrightnessStats(t, composed, level))
	}

	return means
}

// assertRisingMeans fails unless each level is strictly brighter than the one
// below it.
func assertRisingMeans(t *testing.T, levels []int, means []float64) {
	t.Helper()

	for i := 1; i < len(means); i++ {
		if means[i] <= means[i-1] {
			t.Fatalf("gamma %d mean %.2f did not rise above gamma %d's %.2f",
				levels[i], means[i], levels[i-1], means[i-1])
		}
	}
}

// composeBrightnessStats checks one level's frame and returns its mean,
// writing the visual PoW pair (PNG + .stats) on the way.
func composeBrightnessStats(t *testing.T, composed scene.TownScene, level int) float64 {
	t.Helper()

	mean, nonBlack := scene.MeanBrightness(composed)
	if nonBlack < 0.5 {
		t.Fatalf("gamma %d: only %.1f%% of the frame is non-black, want >= 50%%", level, nonBlack*100)
	}

	writeGammaArtifact(t, composed, level, mean, nonBlack)

	return mean
}

// writeSceneArtifact dumps the composed frame into the artifact dir.
func writeSceneArtifact(t *testing.T, composed scene.TownScene, name string) {
	t.Helper()

	artifact := os.Getenv("DG_SCENE_ARTIFACT_DIR")
	if artifact == "" {
		artifact = t.TempDir()
	}

	if err := scene.WritePNG(composed, filepath.Join(artifact, name)); err != nil {
		t.Fatalf("WritePNG: %v", err)
	}
}

// writeGammaArtifact dumps the gamma capture and its stats for the wave's
// visual PoW (lane-ops visual-verification: PNG + .stats, distinct means).
func writeGammaArtifact(t *testing.T, composed scene.TownScene, level int, mean, nonBlack float64) {
	t.Helper()

	name := fmt.Sprintf("gamma_%03d", level)
	writeSceneArtifact(t, composed, name+".png")

	artifact := os.Getenv("DG_SCENE_ARTIFACT_DIR")
	if artifact == "" {
		artifact = t.TempDir()
	}

	stats := fmt.Sprintf("gamma=%d\nmean=%.4f\nnonBlack=%.4f\ndistinct=%d\n",
		level, mean, nonBlack, scene.DistinctColors(composed))

	if err := os.WriteFile(filepath.Join(artifact, name+".stats"), []byte(stats), 0o600); err != nil {
		t.Fatalf("write %s.stats: %v", name, err)
	}
}
