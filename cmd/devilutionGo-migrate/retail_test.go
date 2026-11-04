package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestRetailAssetsMigration migrates the archives named by DG_ASSET_ARCHIVES when
// the environment has them. Retail archives are environment inputs, never
// committed fixtures, so an absence skips with an explicit message instead of
// failing — the same politeness internal/assets uses for its real-asset test
// (mpq_smoke_test.go:TestRealArchivesDecodeEveryRecord). A record that exists but
// fails to decode is a failure, never a skip.
func TestRetailAssetsMigration(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	configured := os.Getenv(assetArchivesEnv)
	if configured == "" {
		t.Skipf("%s is unset: retail archives are not installed in this environment", assetArchivesEnv)
	}

	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the assets subcommand migrates them
	code := run([]string{"assets", "--out", out}, &stdout, &stderr)

	// then every record decoded through the proven decoders
	if code != exitOK {
		t.Fatalf("exit code = %d (stderr: %s)", code, stderr.String())
	}

	// then the assets dir, the manifest and the CSV tables exist with hashes
	listed := readManifest(t, out)
	if len(listed.Files) == 0 {
		t.Fatal("manifest lists no files")
	}

	if listed.Summary.SHA256 == "" || listed.Summary.Files != len(listed.Files) {
		t.Errorf("summary = %+v, want every listed file", listed.Summary)
	}

	t.Logf("migrated %d record(s) from %d archive(s); summary %s",
		listed.Summary.Files, len(listed.Archives), listed.Summary.SHA256)

	// then a second run is idempotent: same summary, nothing written
	var second bytes.Buffer

	if code := run([]string{"assets", "--out", out}, &second, &stderr); code != exitOK {
		t.Fatalf("second run: exit %d (stderr: %s)", code, stderr.String())
	}

	if again := readManifest(t, out); again.Summary != listed.Summary {
		t.Errorf("summary after rerun = %+v, want %+v", again.Summary, listed.Summary)
	}
}
