package light_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// diabloPrefix is the layer boundary this package must not cross (ADR-008):
// light is engine-side and owns its level-type enum; diablo packages import
// light, never the reverse.
const diabloPrefix = "github.com/bmcszk/devilutionGo/internal/diablo"

func TestLightPackageImportsNoDiablo(t *testing.T) {
	// given
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob package sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("no .go files found in the package directory")
	}

	// when
	violations := diabloImports(sources)

	// then
	if len(violations) > 0 {
		t.Fatalf("internal/light imports internal/diablo (ADR-008 layering):\n%s",
			strings.Join(violations, "\n"))
	}
}

// diabloImports parses each source file's import block and reports every path
// under the diablo prefix, as "file: path" entries. A parse error is itself a
// violation, so the scan stays fail-closed.
func diabloImports(sources []string) []string {
	fset := token.NewFileSet()

	var violations []string
	for _, src := range sources {
		file, err := parser.ParseFile(fset, src, nil, parser.ImportsOnly)
		if err != nil {
			return []string{fmt.Sprintf("%s: parse error: %v", src, err)}
		}

		for _, imp := range file.Imports {
			violations = appendIfDiablo(violations, src, strings.Trim(imp.Path.Value, `"`))
		}
	}

	return violations
}

// appendIfDiablo appends the import as a violation entry when its path crosses
// the layer boundary.
func appendIfDiablo(violations []string, src, path string) []string {
	if !strings.HasPrefix(path, diabloPrefix) {
		return violations
	}

	return append(violations, fmt.Sprintf("%s: %s", src, path))
}
