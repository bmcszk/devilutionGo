//go:build ignore

package impure_test

import (
	"os"
	"testing"
)

// TestAllowedFileIO pins the exemption: _test.go files may read the disk.
func TestAllowedFileIO(t *testing.T) {
	if _, err := os.ReadFile("impure.go"); err != nil {
		t.Fatalf("read impure.go: %v", err)
	}
}
