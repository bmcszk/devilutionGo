package persist_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestExportSVHeroEntryLayoutGolden pins the .sv hero-entry byte layout against
// a committed hex image: ExportSV must write the entry upstream reads, i.e.
// "hero" holding codec_encode(PlayerPack, pfile_get_password())
// (Source/pfile.cpp:171-179 EncodeHero).
//
// The image is the committed upstream PlayerPack record (testdata/
// player_pack_golden.hex) encoded with PASSWORD_SINGLE, so it locks the
// pack + password + codec composition as one artifact. The container itself is
// deliberately not compared byte-for-byte: upstream records that the archive
// layout is not byte-stable (test/writehero_test.cpp:415-419), which is why the
// pin is at entry level.
func TestExportSVHeroEntryLayoutGolden(t *testing.T) {
	// given a save whose hero is the committed PlayerPack image
	game, err := persist.GameHeader{CurrLevel: 6, LevelType: 5}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	archive := persist.LegacySVStore{
		Path: filepath.Join(t.TempDir(), "single_0.sv"),
		Mode: persist.GameMode{},
	}

	if err := archive.Write(persist.HeroRecord{Slot: 0, Pack: goldenPlayerPack().Encode()}, game); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// when the archive entries are read back
	heroEntry, err := archive.ReadEntry(persist.HeroEntryName)
	if err != nil {
		t.Fatalf("ReadEntry(hero): %v", err)
	}

	gameEntry, err := archive.ReadEntry(persist.GameEntryName)
	if err != nil {
		t.Fatalf("ReadEntry(game): %v", err)
	}

	// then the hero entry is byte-identical to the committed layout image
	if want := loadSVHeroEntryHex(t); !bytes.Equal(heroEntry, want) {
		t.Fatalf("hero entry = %d bytes %x, want %d bytes %x",
			len(heroEntry), heroEntry[:8], len(want), want[:8])
	}
	// ...and the game stream is carried verbatim
	if !bytes.Equal(gameEntry, game) {
		t.Fatalf("game entry = %x, want %x", gameEntry, game)
	}
}

// loadSVHeroEntryHex reads the committed hero-entry image, ignoring the
// provenance comments.
func loadSVHeroEntryHex(t *testing.T) []byte {
	t.Helper()

	raw, err := os.ReadFile("testdata/sv_hero_entry_single.hex")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	var image strings.Builder

	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}

		image.WriteString(strings.TrimSpace(line))
	}

	decoded, err := hex.DecodeString(image.String())
	if err != nil {
		t.Fatalf("decode golden: %v", err)
	}

	return decoded
}
