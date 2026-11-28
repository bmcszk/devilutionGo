package testsupport

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// MetricTable reads a committed "metric<TAB>value" fixture — the accounting
// rows a replay or parser test pins — and returns them by metric name. Blank
// lines and lines starting with '#' are comments; a row with any other shape
// fails the test, because a fixture that does not parse is a defect.
func MetricTable(t *testing.T, path string) map[string]string {
	t.Helper()

	raw, err := os.ReadFile(path) //nolint:gosec // G304: test-supplied committed fixture path
	if err != nil {
		t.Fatalf("read metric fixture %s: %v", path, err)
	}

	table := make(map[string]string)

	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			t.Fatalf("metric fixture %s line %d = %q, want 2 tab-separated fields", path, i+1, line)
		}

		table[fields[0]] = fields[1]
	}

	return table
}

// MetricInt returns one integer metric of a committed fixture.
func MetricInt(t *testing.T, table map[string]string, name string) int {
	t.Helper()

	raw, ok := table[name]
	if !ok {
		t.Fatalf("metric fixture has no %q row", name)
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("metric %q = %q, want an integer: %v", name, raw, err)
	}

	return value
}
