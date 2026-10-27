package assets_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// BenchmarkCLXDecode decodes the committed small-CLX corpus (internal/testdata/gfx/clx).
func BenchmarkCLXDecode(b *testing.B) {
	// given the committed corpus read once
	names := clxFixtureNames(b)
	corpus := make([][]byte, len(names))
	for i, name := range names {
		corpus[i] = readHexFixture(b, clxHexPath(name))
	}

	// when every file is decoded b.N times
	for b.Loop() {
		for _, raw := range corpus {
			if _, err := assets.DecodeCLX(raw); err != nil {
				b.Fatalf("DecodeCLX: %v", err)
			}
		}
	}
}
