package scene_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// TestTownSceneComposesFromRetailArchives is the 6asa acceptance gate: the
// loader reads town.pal, town.til, town.min, town.sol and the sector .dun
// layers out of the retail archives (levels\towndata\*, diablo.cpp:1371-1387,
// dun_tile_data.cpp:103-106, town.cpp:202-207), CreateTown builds the grids
// and the adapter renders a full frame through DrawWorld/RenderTileFrame.
//
// The archives are environment inputs, never committed fixtures; an absence
// skips (mpq_smoke_test.go findAssetArchives politeness).
func TestTownSceneComposesFromRetailArchives(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 || !scene.ArchivesPresent(dirs) {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	// when the town scene loads and composes one frame
	composed, err := scene.ComposeTownFrame(dirs)
	if err != nil {
		t.Fatalf("ComposeTownFrame: %v", err)
	}

	// then the frame carries real pixel variety: many distinct palette
	// colours (not flat) and the diamond tile structure (a floor tile's
	// left-triangle left edge grows one pixel per row).
	if got := scene.DistinctColors(composed); got < 64 {
		t.Fatalf("composed frame has %d distinct palette colours, want >= 64 (a real town frame is richly coloured)", got)
	}

	if !scene.HasDiamondEdge(composed) {
		t.Fatal("composed frame shows no tile-diamond edge structure")
	}

	// and the PNG artifact lands for the parent's visual check
	artifact := os.Getenv("DG_SCENE_ARTIFACT_DIR")
	if artifact == "" {
		artifact = t.TempDir()
	}

	if err := scene.WritePNG(composed, filepath.Join(artifact, "town.png")); err != nil {
		t.Fatalf("WritePNG: %v", err)
	}
}
