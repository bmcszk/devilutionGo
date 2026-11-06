// Command goldens refreshes the committed golden fixtures the parity tests
// read, one subcommand per fixture family.
//
//	go run ./tools/goldens dun /tmp/drlg-out/drlg internal/testdata/drlg
//	go run ./tools/goldens dpiece /tmp/drlg-out/dpiece internal/testdata/drlg/dpiece
//	go run ./tools/goldens town /tmp/drlg-out/town internal/testdata/drlg/town
//	go run ./tools/goldens town-inputs /tmp/raw-dun internal/testdata/drlg/town
//	go run ./tools/goldens til /tmp/raw-til internal/testdata/drlg/assets
//	go run ./tools/goldens assets /path/to/hellfire.mpq internal/testdata/drlg/assets
//	go run ./tools/goldens gamedata
//
// It never runs in go test or CI and never ships: ADR-019 carves fixture
// refreshers out of the user-facing set, and ADR-021 folds them into this one
// command.
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

// refreshers lists every subcommand in the order the usage line prints them.
var refreshers = []refresher{
	{"dun", 2, runDun},
	{"dpiece", 2, runDpiece},
	{"town", 2, runTown},
	{"til", 2, runTil},
	{"assets", 2, runAssets},
	{"town-inputs", 2, runTownInputs},
	{"gamedata", 0, runGamedata},
}

// refresher re-encodes one fixture family. It receives exactly argCount
// positional paths.
type refresher struct {
	name     string
	argCount int
	run      func(paths []string) error
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("goldens failed", "error", err)
		os.Exit(1)
	}
}

// run dispatches one subcommand with its positional paths.
func run(args []string) error {
	ref, ok := lookupRefresher(args)
	if !ok {
		return errUsage()
	}

	return ref.run(args[1:])
}

// lookupRefresher matches a subcommand name plus its positional argument
// count, so `goldens gamedata` cannot receive paths and `goldens dun` cannot
// run short of them.
func lookupRefresher(args []string) (refresher, bool) {
	for _, ref := range refreshers {
		if len(args) == 1+ref.argCount && args[0] == ref.name {
			return ref, true
		}
	}

	return refresher{}, false
}

// errUsage reports the subcommand list on one line: a bare invocation or a
// mistyped name is a typo, not a readable default.
func errUsage() error {
	names := make([]string, len(refreshers))
	for i, ref := range refreshers {
		names[i] = ref.name
	}

	return errors.New("usage: goldens <" + strings.Join(names, "|") + ">")
}

// runDun converts every `.dun`/`.meta` pair into the drlg golden CSVs.
func runDun(paths []string) error {
	artifactDir, corpusDir := paths[0], paths[1]

	bases, err := dunArtifactBases(artifactDir)
	if err != nil {
		return err
	}

	written, err := writeCorpus(artifactDir, corpusDir, bases)
	if err != nil {
		return err
	}

	slog.Info("goldens dun done", "files", written, "corpus", corpusDir)

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
