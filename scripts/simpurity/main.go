// Command simpurity is the simulation purity gate (ADR-017): the sim packages
// under internal/game and internal/diablo must replay without a window, a wall
// clock or a filesystem. It parses non-test sources; _test.go files are exempt.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// fixtureLoaderAudit is the committed-golden manifest loader: it serves the
// tests, is not on the simulation path, and must open a file to do its job.
const fixtureLoaderAudit = "internal/diablo/gamedata/audit"

// defaultRoots are the simulation trees the gate covers when no path is given.
var defaultRoots = []string{"internal/game", "internal/diablo"}

// bannedFileImports are the packages that reach the disk directly.
var bannedFileImports = map[string]bool{"os": true, "io/ioutil": true}

// bannedClockFuncs are the wall-clock readings a deterministic sim may not make;
// time.Duration arithmetic stays legal, only reading the clock is not.
var bannedClockFuncs = map[string]bool{"Now": true, "Tick": true, "NewTicker": true, "Since": true}

// violation is one banned import or call: package, file, line and the offender.
type violation struct {
	pkg    string
	file   string
	line   int
	detail string
}

// String renders the finding as package: file:line: offending import or call.
func (v violation) String() string {
	return fmt.Sprintf("%s: %s:%d: %s", v.pkg, v.file, v.line, v.detail)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run reports every violation and returns 1 when the tree is impure, 2 when the
// scan itself failed (fail-closed: an unreadable tree is never a pass).
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("simpurity", flag.ContinueOnError)
	flags.SetOutput(stderr)

	if err := flags.Parse(args); err != nil {
		return 2
	}

	roots := flags.Args()
	if len(roots) == 0 {
		roots = defaultRoots
	}

	found, err := collect(roots)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "simpurity: %v\n", err)

		return 2
	}

	for _, v := range found {
		_, _ = fmt.Fprintln(stdout, v)
	}

	if len(found) > 0 {
		_, _ = fmt.Fprintf(stderr, "simpurity: %d violation(s) in %s\n", len(found), strings.Join(roots, " "))

		return 1
	}

	_, _ = fmt.Fprintf(stdout, "simpurity: %s clean\n", strings.Join(roots, " "))

	return 0
}

// collect scans every root and returns all findings in walk order.
func collect(roots []string) ([]violation, error) {
	var found []violation

	for _, root := range roots {
		rootFound, err := collectRoot(root)
		if err != nil {
			return nil, err
		}

		found = append(found, rootFound...)
	}

	return found, nil
}

// collectRoot walks one tree, skipping testdata, and checks each sim source file.
func collectRoot(root string) ([]violation, error) {
	fset := token.NewFileSet()

	var found []violation

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		fileFound, err := visit(fset, root, path, entry, walkErr)
		if err != nil {
			return err
		}

		found = append(found, fileFound...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}

	return found, nil
}

// visit handles one walked entry: pruned directories, skipped files, and the
// sim sources the purity rules then check.
func visit(
	fset *token.FileSet, root, path string, entry fs.DirEntry, walkErr error,
) ([]violation, error) {
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", path, walkErr)
	}

	if entry.IsDir() {
		if path != root && skipDir(entry.Name()) {
			return nil, filepath.SkipDir
		}

		return nil, nil
	}

	if !isSimSource(path) {
		return nil, nil
	}

	return checkFile(fset, path)
}

// skipDir drops fixtures and hidden directories: they are never sim code.
func skipDir(name string) bool {
	return name == "testdata" || strings.HasPrefix(name, ".")
}

// isSimSource reports the files the gate covers: simulation sources only, so
// _test.go is exempt (tests may script files, clocks and fixtures).
func isSimSource(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// checkFile reports the banned imports and wall-clock calls of one sim file.
func checkFile(fset *token.FileSet, path string) ([]violation, error) {
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("simpurity: parse %s: %w", path, err)
	}

	pkg := filepath.ToSlash(filepath.Dir(path))
	if isFixtureLoader(pkg) {
		return nil, nil
	}

	found, importsTime, err := checkImports(fset, pkg, path, file)
	if err != nil {
		return nil, err
	}

	if importsTime {
		found = append(found, checkClock(fset, pkg, path, file)...)
	}

	return found, nil
}

// checkImports reports banned import paths and tells the caller whether the
// file imports time, which gates the wall-clock scan.
func checkImports(
	fset *token.FileSet, pkg, path string, file *ast.File,
) ([]violation, bool, error) {
	var found []violation

	importsTime := false

	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, false, fmt.Errorf("simpurity: import path in %s: %w", path, err)
		}

		if importPath == "time" {
			importsTime = true
		}

		if reason := bannedImport(importPath); reason != "" {
			found = append(found, violation{
				pkg:    pkg,
				file:   path,
				line:   fset.Position(spec.Pos()).Line,
				detail: fmt.Sprintf("%s imports %q", reason, importPath),
			})
		}
	}

	return found, importsTime, nil
}

// bannedImport names why an import path breaks purity, or "" when it is legal.
func bannedImport(importPath string) string {
	if importPath == "github.com/hajimehoshi/ebiten/v2" ||
		strings.HasPrefix(importPath, "github.com/hajimehoshi/ebiten") {
		return "rendering"
	}

	if hasPathSegment(importPath, "render") {
		return "rendering"
	}

	if bannedFileImports[importPath] {
		return "file io"
	}

	return ""
}

// hasPathSegment reports whether name is one whole slash-separated element.
func hasPathSegment(path, name string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == name {
			return true
		}
	}

	return false
}

// checkClock reports every wall-clock reading in the file.
func checkClock(fset *token.FileSet, pkg, path string, file *ast.File) []violation {
	var found []violation

	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != "time" || !bannedClockFuncs[selector.Sel.Name] {
			return true
		}

		found = append(found, violation{
			pkg:    pkg,
			file:   path,
			line:   fset.Position(selector.Pos()).Line,
			detail: fmt.Sprintf("wall clock time.%s", selector.Sel.Name),
		})

		return true
	})

	return found
}

// isFixtureLoader reports packages that open committed goldens for the tests:
// test scaffolding is not simulation code, so file io is theirs to use.
func isFixtureLoader(pkg string) bool {
	return hasPathSegment(pkg, "testsupport") || pkg == fixtureLoaderAudit
}
