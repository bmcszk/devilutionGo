package txtdata

import "fmt"

// OverrideStack applies mod tables onto a base table: a mod row whose key
// matches a base row replaces it; an unmatched key appends a new row.
type OverrideStack struct {
	base   *Table
	mods   []*Table // applied in order; later tables win
	keyCol string
}

// NewOverrideStack returns a stack that applies mod tables onto base, keyed
// by the fixed "id" column, which must exist in the base table and in every
// mod table (all ported txtdata tables key on id).
// Mirrors the override application in devilutionX devilutionx.mpq
// txtdata mod loading (SourceD1/txtdata override merge).
func NewOverrideStack(base *Table, mods ...*Table) (*OverrideStack, error) {
	const keyColumn = "id"
	if _, err := base.Column(keyColumn); err != nil {
		return nil, fmt.Errorf("txtdata: base table: %w", err)
	}

	for i, m := range mods {
		if _, err := m.Column(keyColumn); err != nil {
			return nil, fmt.Errorf("txtdata: mod table %d: %w", i, err)
		}
	}

	return &OverrideStack{base: base, mods: mods, keyCol: keyColumn}, nil
}

// Resolve applies every mod table to the base table. The merged table keeps
// the base header, surviving base rows keep base order, and appended rows
// follow in the order their mods introduced them.
func (s *OverrideStack) Resolve() (*Table, error) {
	keyIdx, err := s.base.Column(s.keyCol)
	if err != nil {
		return nil, err
	}

	merged := &Table{Header: s.base.Header}

	seen := map[string]int{}
	if err := mergeRows(merged, seen, s.base.Rows, keyIdx, "base"); err != nil {
		return nil, err
	}

	for _, mod := range s.mods {
		modKeyIdx, err := mod.Column(s.keyCol)
		if err != nil {
			return nil, err
		}

		if err := mergeRows(merged, seen, mod.Rows, modKeyIdx, "mod"); err != nil {
			return nil, err
		}
	}

	return merged, nil
}

// mergeRows appends rows to merged, replacing a row whose key was already
// seen; a key seen for the first time is appended in row order.
func mergeRows(merged *Table, seen map[string]int, rows [][]string, keyIdx int, label string) error {
	for _, row := range rows {
		if keyIdx >= len(row) {
			return fmt.Errorf("txtdata: %s row too short: %v", label, row)
		}

		if rowIdx, ok := seen[row[keyIdx]]; ok {
			merged.Rows[rowIdx] = row
			continue
		}

		seen[row[keyIdx]] = len(merged.Rows)
		merged.Rows = append(merged.Rows, row)
	}

	return nil
}
