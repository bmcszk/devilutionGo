package assets

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Column is the tab separator used by every .tsv table shipped with the game.
const Column = '\t'

// ErrUnknownColumn is returned when a row accessor names a column that the
// header does not declare.
var ErrUnknownColumn = errors.New("assets: unknown txtdata column")

// utf8BOM is stripped from the head of a table when present. The shipped
// tables are ASCII; only the byte-order mark can appear (parser.hpp:ParseTSV
// treats the file as raw bytes).
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Table is one parsed .tsv file: a header row plus typed data rows. Rows keep
// their field order so callers can address columns by name and by index.
type Table struct {
	Header []string
	Rows   []Row

	columns map[string]int
}

// Row is a single data row of a Table. Fields are raw strings; the typed
// accessors convert them on demand and report conversion failures instead of
// silently returning zero values.
type Row struct {
	Fields []string

	table *Table
}

// Load parses a tab-separated table from reader.
//
// The input is normalized before parsing: a leading UTF-8 BOM is dropped and
// CRLF / lone CR line endings are folded to LF, so tables exported on any
// platform parse identically. Empty lines are skipped. The first surviving
// line is the header; every following line must have exactly len(Header)
// fields or Load fails, because a ragged row means the corpus is corrupted and
// no downstream count can be trusted.
//
// An empty input yields an empty table and no error: absence of rows is not a
// parse failure.
func Load(reader io.Reader) (*Table, error) {
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return &Table{}, nil
	}

	table, err := newTable(lines[0])
	if err != nil {
		return nil, err
	}
	if err := table.appendRows(lines[1:]); err != nil {
		return nil, err
	}
	return table, nil
}

// Len reports the number of data rows in the table.
func (t *Table) Len() int {
	if t == nil {
		return 0
	}
	return len(t.Rows)
}

// ColumnIndex resolves a header name to its field position.
func (t *Table) ColumnIndex(name string) (int, error) {
	if t == nil || t.columns == nil {
		return 0, fmt.Errorf("%w: %q", ErrUnknownColumn, name)
	}
	idx, ok := t.columns[name]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownColumn, name)
	}
	return idx, nil
}

// String returns the raw field value of the named column.
func (r Row) String(name string) (string, error) {
	idx, err := r.table.ColumnIndex(name)
	if err != nil {
		return "", err
	}
	return r.Fields[idx], nil
}

// Int returns the named column parsed as a base-10 integer.
func (r Row) Int(name string) (int, error) {
	value, err := r.String(name)
	if err != nil {
		return 0, err
	}
	parsed, convErr := strconv.Atoi(strings.TrimSpace(value))
	if convErr != nil {
		return 0, fmt.Errorf("assets: column %q: parse int %q: %w", name, value, convErr)
	}
	return parsed, nil
}

// Byte returns the named column parsed as an unsigned 8-bit value.
func (r Row) Byte(name string) (byte, error) {
	parsed, err := r.Int(name)
	if err != nil {
		return 0, err
	}
	if parsed < 0 || parsed > 0xFF {
		return 0, fmt.Errorf("assets: column %q: %d out of byte range", name, parsed)
	}
	return byte(parsed), nil
}

// Bool returns the named column parsed as a flag. The corpus writes flags as
// "0"/"1"; "true"/"false" are accepted so hand-written fixtures read naturally.
// An empty field is a false flag.
func (r Row) Bool(name string) (bool, error) {
	value, err := r.String(name)
	if err != nil {
		return false, err
	}
	switch strings.TrimSpace(value) {
	case "", "0", "false", "FALSE":
		return false, nil
	case "1", "true", "TRUE":
		return true, nil
	default:
		return false, fmt.Errorf("assets: column %q: parse bool %q", name, value)
	}
}

// readLines reads the whole table and returns its non-empty lines.
func readLines(reader io.Reader) ([]string, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("assets: read txtdata: %w", err)
	}
	return splitLines(raw), nil
}

// splitLines normalizes line endings and returns the non-empty lines.
func splitLines(raw []byte) []string {
	raw = bytes.TrimPrefix(raw, utf8BOM)
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))

	lines := make([]string, 0, bytes.Count(raw, []byte("\n"))+1)
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// newTable parses the header line and indexes its column names.
func newTable(headerLine string) (*Table, error) {
	table := &Table{Header: strings.Split(headerLine, string(Column))}
	table.columns = make(map[string]int, len(table.Header))
	for i, name := range table.Header {
		if _, dup := table.columns[name]; dup {
			return nil, fmt.Errorf("assets: duplicate txtdata column %q", name)
		}
		table.columns[name] = i
	}
	return table, nil
}

// appendRows parses every data line and asserts the field count per row.
func (t *Table) appendRows(lines []string) error {
	t.Rows = make([]Row, 0, len(lines))
	for i, line := range lines {
		fields := strings.Split(line, string(Column))
		if len(fields) != len(t.Header) {
			// +2: one for the header line, one for 1-based numbering.
			return fmt.Errorf("assets: txtdata line %d has %d fields, header declares %d",
				i+2, len(fields), len(t.Header))
		}
		t.Rows = append(t.Rows, Row{Fields: fields, table: t})
	}
	return nil
}
