// Command dunconv converts dvg_probe `drlg` artifacts (one `.dun` + `.meta`
// pair per level/entry/seed) into the committed drlg golden CSVs under
// internal/testdata/drlg, byte-identically to the corpus the parity tests read.
//
//	go run ./tools/probe/dunconv /tmp/drlg-out/drlg internal/testdata/drlg
//
// The encoder is testsupport.EncodeDunCSV, so a refresh cannot drift from the
// format the tests parse.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("dunconv failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 2 {
		return runDun(args[0], args[1])
	}
	if len(args) == 3 && args[0] == "dpiece" {
		return runDpiece(args[1], args[2])
	}

	return errors.New("usage: dunconv <artifact-dir> <corpus-dir> | dunconv dpiece <artifact-dir> <corpus-dir>")
}

// runDun converts every `.dun`/`.meta` pair into the drlg golden CSVs.
func runDun(artifactDir, corpusDir string) error {
	bases, err := dunArtifactBases(artifactDir)
	if err != nil {
		return err
	}

	written, err := writeCorpus(artifactDir, corpusDir, bases)
	if err != nil {
		return err
	}

	slog.Info("dunconv done", "files", written, "corpus", corpusDir)

	return nil
}

// dunArtifactBases lists the artifact base names ("l5-main-2588") that have
// both a `.dun` and a `.meta` sidecar, sorted for deterministic output.
func dunArtifactBases(artifactDir string) ([]string, error) {
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return nil, fmt.Errorf("read artifact dir: %w", err)
	}

	bases := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".dun") {
			continue
		}
		bases = append(bases, strings.TrimSuffix(name, ".dun"))
	}

	if len(bases) == 0 {
		return nil, fmt.Errorf("no .dun artifacts in %s", artifactDir)
	}

	slices.Sort(bases)

	return bases, nil
}

// writeCorpus converts every artifact base into `<base>.csv` under corpusDir.
func writeCorpus(artifactDir, corpusDir string, bases []string) (int, error) {
	for _, base := range bases {
		csv, err := convertArtifact(artifactDir, base)
		if err != nil {
			return 0, err
		}
		if err := os.WriteFile(filepath.Join(corpusDir, base+".csv"), csv, 0o644); err != nil {
			return 0, fmt.Errorf("write %s.csv: %w", base, err)
		}
	}

	return len(bases), nil
}

// convertArtifact parses one artifact pair and re-encodes it as corpus CSV.
func convertArtifact(artifactDir, base string) ([]byte, error) {
	dun, err := os.ReadFile(filepath.Join(artifactDir, base+".dun"))
	if err != nil {
		return nil, fmt.Errorf("read %s.dun: %w", base, err)
	}
	meta, err := os.ReadFile(filepath.Join(artifactDir, base+".meta"))
	if err != nil {
		return nil, fmt.Errorf("read %s.meta: %w", base, err)
	}

	pair, err := testsupport.ParseDunArtifact(dun, meta)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", base, err)
	}

	return testsupport.EncodeDunCSV(pair), nil
}
