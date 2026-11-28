package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audit"
)

// auditManifest is the committed expectation the data audit reads.
const auditManifest = "internal/diablo/gamedata/audit/audit_manifest.json"

// runAudit prints the diffable data-audit report for the committed corpora and
// returns an error (non-zero exit) when the closed world or a table drifted.
func runAudit(paths []string) error {
	_ = paths

	report, err := audit.Audit(".", auditManifest)
	if err != nil {
		return fmt.Errorf("data audit: %w", err)
	}

	if _, err := io.WriteString(os.Stdout, report.Text()); err != nil {
		return fmt.Errorf("print data audit report: %w", err)
	}

	for _, residual := range audit.Residuals() {
		if _, err := fmt.Fprintf(os.Stdout, "# residual: %s\n", residual); err != nil {
			return fmt.Errorf("print data audit residual: %w", err)
		}
	}

	if !report.OK() {
		return fmt.Errorf("data audit: %d problem(s)", len(report.Problems))
	}

	return nil
}
