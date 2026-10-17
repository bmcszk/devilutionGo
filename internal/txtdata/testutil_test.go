package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/txtdata"
)

func mustParse(t *testing.T, data []byte) *txtdata.Table {
	t.Helper()
	tbl, err := txtdata.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return tbl
}
