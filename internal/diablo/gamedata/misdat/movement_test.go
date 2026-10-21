package misdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
)

func TestMovementDistributionDomain(t *testing.T) {
	// given
	base, err := misdat.ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat: %v", err)
	}

	// when
	counts := map[string]int{}
	for _, row := range base {
		counts[row.MovementDistribution]++
	}

	// then
	for _, tc := range []struct {
		name  string
		value string
		want  int
	}{
		{name: "blockable", value: "Blockable", want: 14},
		{name: "unblockable", value: "Unblockable", want: 3},
		// The empty movementDistribution is the upstream default, not a parse gap.
		{name: "empty", value: "", want: 51},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := counts[tc.value]; got != tc.want {
				t.Errorf("movementDistribution %q = %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}
