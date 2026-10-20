// Mod-override stacking per upstream's built-in-mod mechanism
// (Source/options.cpp:71-80 DiscoverMods; Source/engine/assets.cpp:704-760
// LoadModArchives: mod tables load at priority >= 10000, above retail 1000
// and devilutionx.mpq 9000, so their rows shadow the base tables).
package txtdata

import "fmt"

// OverrideStack applies mod tables onto a base table: a mod row whose key
// matches a base row replaces it; an unmatched key appends a new row
// (dg-requirements-append.md V0.2: rows are appended, some existing rows are
// edited — same header, same column order).
type OverrideStack struct {
	Base   *Table
	Mods   []*Table // in application order; later tables win
	keyCol string
}

// NewOverrideStack keys rows by the named column (upstream loaders key on the
// first column, id/type documentation column, via reader.advance()).
func NewOverrideStack(base *Table, keyColumn string, mods ...*Table) (*OverrideStack, error) {
	if _, err := base.Column(keyColumn); err != nil {
		return nil, fmt.Errorf("txtdata: base table: %w", err)
	}
	for i, m := range mods {
		if _, err := m.Column(keyColumn); err != nil {
			return nil, fmt.Errorf("txtdata: mod table %d: %w", i, err)
		}
	}
	return &OverrideStack{Base: base, Mods: mods, keyCol: keyColumn}, nil
}

// Resolve applies every mod table onto the base and returns the merged table
// with the base header. Row order: surviving base rows in base order, then
// appended rows in the order their mods introduced them.
func (s *OverrideStack) Resolve() (*Table, error) {
	ki, err := s.Base.Column(s.keyCol)
	if err != nil {
		return nil, err
	}
	merged := &Table{Header: s.Base.Header}
	seen := map[string]int{}
	for _, row := range s.Base.Rows {
		if ki >= len(row) {
			return nil, fmt.Errorf("txtdata: base row too short: %v", row)
		}
		seen[row[ki]] = len(merged.Rows)
		merged.Rows = append(merged.Rows, row)
	}
	for _, mod := range s.Mods {
		mi, err := mod.Column(s.keyCol)
		if err != nil {
			return nil, err
		}
		for _, row := range mod.Rows {
			if mi >= len(row) {
				return nil, fmt.Errorf("txtdata: mod row too short: %v", row)
			}
			if idx, ok := seen[row[mi]]; ok {
				merged.Rows[idx] = row
				continue
			}
			seen[row[mi]] = len(merged.Rows)
			merged.Rows = append(merged.Rows, row)
		}
	}
	return merged, nil
}
