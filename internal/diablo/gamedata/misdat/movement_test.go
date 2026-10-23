// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package misdat

import (
	"testing"
)

func TestMovementDistributionDomain(t *testing.T) {
	// given
	base, err := ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMisdat: %v", err)
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
