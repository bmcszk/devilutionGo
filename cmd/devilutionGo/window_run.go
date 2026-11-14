//go:build !windowed

package main

import "errors"

// errWindowedTag is the named error the untagged build returns for the
// windowed mode: the binary was built without -tags windowed, so the ebiten
// boot is not linked (ADR-020 two-binary law; devilGo-v9xy display guard).
var errWindowedTag = errors.New(
	"windowed mode requires a build with -tags windowed " +
		"(this binary links no display stack)",
)

// runWindowed reports the missing windowed build; a windowless environment
// fails with this named error, not a panic (v9xy AC).
func runWindowed(_ []string, _ string) error {
	return errWindowedTag
}
