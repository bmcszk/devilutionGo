package parity

import "testing"

// AssertTable reports every difference between gotRows and the golden fixture at goldenPath.
// An unreadable fixture fails the test as well: parity tests must fail, not skip.
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
