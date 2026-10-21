package parity

import "testing"

// AssertTable fails t when gotRows do not match the golden fixture at goldenPath.
func AssertTable(t *testing.T, gotRows [][]string, goldenPath string) {
	t.Helper()
	fixture, err := LoadFixture(goldenPath)
	if err != nil {
		t.Errorf("load fixture: %v", err)
		return
	}
	for _, diff := range CompareTable(gotRows, fixture) {
		t.Errorf("%s", diff)
	}
}
