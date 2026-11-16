package scene_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
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
