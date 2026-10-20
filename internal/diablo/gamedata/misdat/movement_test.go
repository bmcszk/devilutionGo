package misdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
)

// Data-side enumeration of the movementDistribution domain: the base corpus
// uses exactly Blockable / Unblockable (51 control rows carry the empty
// value = upstream default). The 16.16 traveled arithmetic itself is the
// missile-engine bean's scope (devilGo-mwlm, MoveMissileAndCheckMissileCol
// missiles.cpp:636).
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
