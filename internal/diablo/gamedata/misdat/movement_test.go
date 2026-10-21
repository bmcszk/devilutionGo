package misdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
)

// The empty movementDistribution is the upstream default, not a parse gap.
func TestMovementDistributionDomain(t *testing.T) {
	base, err := misdat.ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat: %v", err)
	}
	seen := map[string]int{}
	for _, m := range base {
		seen[m.MovementDistribution]++
	}
	if seen["Blockable"] != 14 || seen["Unblockable"] != 3 {
		t.Fatalf("Blockable=%d Unblockable=%d, want 14/3", seen["Blockable"], seen["Unblockable"])
	}
	if seen[""] != 51 {
		t.Fatalf("empty movementDistribution = %d, want 51", seen[""])
	}
}
