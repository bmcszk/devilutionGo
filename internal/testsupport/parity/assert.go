package parity

import "testing"

// AssertTable reports fixture loading and table comparison failures through t.
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
