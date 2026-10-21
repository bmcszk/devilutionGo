package txtdata_test

import (
	"testing"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

func mustParse(t *testing.T, data []byte) *enginetxtdata.Table {
	t.Helper()
	// given: the caller-supplied TSV payload
	// when: the payload is parsed
	// then: a malformed payload fails the calling test, a valid one yields the table
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return tbl
}
