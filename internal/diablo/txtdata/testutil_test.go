package txtdata_test

import (
	"testing"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// mustParse parses TSV fixture bytes or fails the test.
func mustParse(t *testing.T, data []byte) *enginetxtdata.Table {
	t.Helper()

	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	return tbl
}
