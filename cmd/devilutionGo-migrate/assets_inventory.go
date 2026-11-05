package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	manifestpkg "github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

const (
	// defaultInventoryPath is the graphics requirements list (ADR-019). The
	// extraction record of the same name is not a valid inventory.
	defaultInventoryPath = "assets/inventory.yaml"
	// defaultInventoryRoot is the assets dir the build embeds: entry paths
	// resolve under it (ADR-005).
	defaultInventoryRoot = "assets/loose"
	// inventorySubcommands names the nested `assets` commands, in usage order.
	inventorySubcommands = "verify, list, convert"
)

const assetsUsage = `assets subcommands:
  extract  extract MPQ archives into the assets dir, CSV tables and manifest.yaml
  verify   check every inventory entry resolves under an asset root (needs no MPQ)
           usage: assets verify [--manifest <yaml>] [--root <dir>|<root>]
  list     print the inventory entries
  convert  convert one legacy .cel/.cl2/.pcx file into .clx
`

const assetConvertUsage = `convert flags: --format cel|cl2|pcx --in <file> --out <file>
  --widths       frame widths, comma separated (cel, cl2)
  --frames       pcx frame count (positive) or frame height (negative)
  --transparent  pcx transparent palette index
`

// nestedAssetsIndex lists the nested `assets` subcommands by name.
var nestedAssetsIndex = map[string]inventoryCommand{
	"verify":  runAssetsVerify,
	"list":    runAssetsList,
	"convert": runAssetsConvert,
}

// inventoryCommand is one nested `assets` subcommand.
type inventoryCommand func(args []string, stdout, stderr io.Writer) int

// runAssetsCommand dispatches an `assets` invocation: a nested subcommand when
// the first argument names one, otherwise the extraction pass the flags ask for.
func runAssetsCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return runAssets(args, stdout, stderr)
	}

	if args[0] == "extract" {
		return runAssets(args[1:], stdout, stderr)
	}

	command, ok := nestedAssetsIndex[args[0]]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets: unknown command %q (want extract, %s)\n\n%s",
			args[0], inventorySubcommands, assetsUsage)

		return exitUsage
	}

	return command(args[1:], stdout, stderr)
}

// loadInventory reads and validates the graphics requirements list.
func loadInventory(path string) (manifestpkg.Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return manifestpkg.Manifest{}, fmt.Errorf("open inventory %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	m, err := manifestpkg.Read(file)
	if err != nil {
		return manifestpkg.Manifest{}, fmt.Errorf("read inventory %s: %w", path, err)
	}

	if err := m.Validate(); err != nil {
		return manifestpkg.Manifest{}, fmt.Errorf("invalid inventory %s: %w", path, err)
	}

	return m, nil
}

// runAssetsVerify reports every required entry that does not resolve under
// root. It needs no MPQ archive: the inventory is the requirement, the assets
// dir is the answer, and nothing here reads a retail file.
func runAssetsVerify(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("assets verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inventoryPath := flags.String("manifest", defaultInventoryPath, "inventory YAML to verify")
	rootFlag := flags.String("root", defaultInventoryRoot, "asset root the entry paths resolve under")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	root := verifyRoot(flags, *rootFlag)

	m, err := loadInventory(*inventoryPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets verify: %v\n", err)

		return exitFailure
	}

	missing := m.Missing(root)
	for _, entry := range missing {
		_, _ = fmt.Fprintf(stdout, "missing %s (requested by %s)\n", entry.Path, entry.Loader)
	}

	if len(missing) > 0 {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets verify: %d of %d entries do not resolve under %s\n",
			len(missing), len(m.Entries), root)

		return exitFailure
	}

	_, _ = fmt.Fprintf(stdout, "devilutionGo-migrate assets verify: %d entries in %d families resolve under %s\n",
		len(m.Entries), m.Summary.Families, root)

	return exitOK
}

// verifyRoot reads the asset root: the positional argument when given, so the
// documented `assets verify <root>` form works, else the --root flag.
func verifyRoot(flags *flag.FlagSet, flagRoot string) string {
	if flags.NArg() > 0 {
		return flags.Arg(0)
	}

	return flagRoot
}

// runAssetsList prints every inventory entry, one per line or as a markdown table.
func runAssetsList(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("assets list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inventoryPath := flags.String("manifest", defaultInventoryPath, "inventory YAML to list")
	markdown := flags.Bool("markdown", false, "print a markdown table instead of tab-separated lines")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	m, err := loadInventory(*inventoryPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets list: %v\n", err)

		return exitFailure
	}

	if *markdown {
		_, _ = fmt.Fprintln(stdout, "| family | path | format | frames | loader |")
		_, _ = fmt.Fprintln(stdout, "| --- | --- | --- | --- | --- |")
	}

	for _, entry := range m.Entries {
		if *markdown {
			_, _ = fmt.Fprintf(stdout, "| %s | %s | %s | %d | %s |\n",
				entry.Family, entry.Path, entry.Format, entry.Frames, entry.Loader)

			continue
		}

		_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%d\t%s\n",
			entry.Family, entry.Path, entry.Format, entry.Frames, entry.Loader)
	}

	return exitOK
}
