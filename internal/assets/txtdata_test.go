package assets_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const txtdataFixtures = "../testdata/assets/txtdata"

// accessorValues collects one accessor outcome pair.
type accessorValues[T any] struct {
	value T
	err   error
}

// numericFields is the numeric accessor outcome set for one table.
type numericFields struct {
	id      accessorValues[int]
	small   accessorValues[byte]
	badByte accessorValues[byte]
}

// flagFields is the boolean accessor outcome set for one table.
type flagFields struct {
	trueFlag  accessorValues[bool]
	falseFlag accessorValues[bool]
	badFlag   accessorValues[bool]
}

func TestLoadNormalizesLineEndings(t *testing.T) {
	for _, name := range []string{"bom.tsv", "crlf.tsv", "cr.tsv", "lf_no_trail.tsv"} {
		t.Run(name, func(t *testing.T) {
			// given four exports of one two-row table, differing only in their
			// line endings and byte-order mark
			// (internal/testdata/assets/txtdata/README.md)
			path := filepath.Join(txtdataFixtures, name)

			// when the export is loaded
			table := loadTxtdataPath(t, path)

			// then the header, row count and cell values are identical
			expectSimpleTable(t, table)
		})
	}
}

func TestLoadEmptyIsEmptyTable(t *testing.T) {
	// when a zero-byte export is loaded
	table := loadTxtdataFixture(t, "empty.tsv")

	// then the result is an empty table, not an error
	if table == nil {
		t.Fatal("Load(empty.tsv) = nil table")
	}
	if table.Len() != 0 {
		t.Fatalf("row count = %d, want 0", table.Len())
	}
	if len(table.Header) != 0 {
		t.Fatalf("header = %v, want empty", table.Header)
	}
}

func TestLoadRejectsCorruptTable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		want   string
	}{
		{"ragged row", "id\tname\n1\talpha\textra\n", "3 fields"},
		{"duplicate column", "id\tid\n1\t2\n", "duplicate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given a table source with one defect
			source := tc.source

			// when it is loaded
			table, err := assets.Load(strings.NewReader(source))

			// then it is rejected with a diagnostic naming the defect
			expectRejectedTable(t, table, err, tc.want)
		})
	}
}

func TestRowAccessors(t *testing.T) {
	// given a two-row table covering every accessor type
	table := loadTable(t, "id\tflag\tsmall\tname\n7\t1\t200\talpha\n8\t\t\t256\n")

	// when every accessor reads its column
	numbers := readNumericFields(t, table)
	flags := readFlagFields(t, table)
	unknownErr := readUnknownColumn(t, table)

	// then the conversions and the rejections match the corpus rules
	expectNumericFields(t, numbers)
	expectFlagFields(t, flags)
	if !errors.Is(unknownErr, assets.ErrUnknownColumn) {
		t.Fatalf("String(missing) error = %v, want ErrUnknownColumn", unknownErr)
	}
}

func expectSimpleTable(t *testing.T, table *assets.Table) {
	t.Helper()

	if got, want := strings.Join(table.Header, "|"), "id|name"; got != want {
		t.Fatalf("header = %q, want %q", got, want)
	}
	if got, want := table.Len(), 2; got != want {
		t.Fatalf("row count = %d, want %d", got, want)
	}

	id, idErr := table.Rows[0].Int("id")
	name, nameErr := table.Rows[1].String("name")
	if idErr != nil || nameErr != nil {
		t.Fatalf("accessors = (%d, %q) with errors (%v, %v)", id, name, idErr, nameErr)
	}
	if id != 1 || name != "beta" {
		t.Fatalf("row values = (%d, %q), want (1, beta)", id, name)
	}
}

func expectRejectedTable(t *testing.T, table *assets.Table, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Load accepted corrupt table: %+v", table)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want it to mention %q", err, want)
	}
}

func expectNumericFields(t *testing.T, values numericFields) {
	t.Helper()

	if values.id.err != nil || values.id.value != 7 {
		t.Fatalf("Int(id) = %d, %v; want 7", values.id.value, values.id.err)
	}
	if values.small.err != nil || values.small.value != 200 {
		t.Fatalf("Byte(small) = %d, %v; want 200", values.small.value, values.small.err)
	}
	if values.badByte.err == nil {
		t.Fatal("Byte accepted a non-numeric field")
	}
}

func expectFlagFields(t *testing.T, values flagFields) {
	t.Helper()

	if values.trueFlag.err != nil || !values.trueFlag.value {
		t.Fatalf("Bool(flag) row 0 = %v, %v; want true", values.trueFlag.value, values.trueFlag.err)
	}
	if values.falseFlag.err != nil || values.falseFlag.value {
		t.Fatalf("Bool(flag) row 1 = %v, %v; want false", values.falseFlag.value, values.falseFlag.err)
	}
	if values.badFlag.err == nil {
		t.Fatal("Bool accepted a non-flag literal")
	}
}

func readNumericFields(t *testing.T, table *assets.Table) numericFields {
	t.Helper()

	var values numericFields
	values.id.value, values.id.err = table.Rows[0].Int("id")
	values.small.value, values.small.err = table.Rows[0].Byte("small")
	values.badByte.value, values.badByte.err = table.Rows[1].Byte("name")
	return values
}

func readFlagFields(t *testing.T, table *assets.Table) flagFields {
	t.Helper()

	var values flagFields
	values.trueFlag.value, values.trueFlag.err = table.Rows[0].Bool("flag")
	values.falseFlag.value, values.falseFlag.err = table.Rows[1].Bool("flag")
	values.badFlag.value, values.badFlag.err = table.Rows[0].Bool("id")
	return values
}

func readUnknownColumn(t *testing.T, table *assets.Table) error {
	t.Helper()

	_, err := table.Rows[0].String("missing")
	if err != nil {
		return fmt.Errorf("read missing column: %w", err)
	}
	return nil
}

// loadTable parses an inline table source.
func loadTable(t *testing.T, source string) *assets.Table {
	t.Helper()

	table, err := assets.Load(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return table
}

// loadTxtdataFixture loads a committed .tsv golden by file name.
func loadTxtdataFixture(t *testing.T, name string) *assets.Table {
	t.Helper()
	return loadTxtdataPath(t, filepath.Join(txtdataFixtures, name))
}

// loadTxtdataPath loads a committed .tsv golden and fails when it is missing: a
// missing fixture is a corpus defect, never a skip.
func loadTxtdataPath(t *testing.T, path string) *assets.Table {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture %s: %v", path, err)
	}
	t.Cleanup(func() { _ = file.Close() })

	table, err := assets.Load(file)
	if err != nil {
		t.Fatalf("Load(%s): %v", path, err)
	}
	return table
}
