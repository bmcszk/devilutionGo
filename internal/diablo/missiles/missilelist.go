// Package missiles: the missile container (row list plus payload storage).
package missiles

// Entry is one container row: a missile id and its deletion mark.
type Entry struct {
	ID     int
	DelFlg bool
}

// MissileList holds the rows in std::list<Missile> iteration order.
type MissileList struct {
	rows []Entry
	// payload holds the missiles; removal filters rows only, so the two
	// slices drift apart.
	payload []Missile
}

// Append adds a row at the tail; rows marked for deletion stay in place.
func (l *MissileList) Append(id int) {
	l.rows = append(l.rows, Entry{ID: id})
}

// Mark flags the row with the given id for deletion; the row stays in the
// list until a removal drops it.
func (l *MissileList) Mark(id int) {
	for i := range l.rows {
		if l.rows[i].ID == id {
			l.rows[i].DelFlg = true
			return
		}
	}
}

// Marked reports whether the row with the given id is flagged for deletion.
func (l *MissileList) Marked(id int) bool {
	for i := range l.rows {
		if l.rows[i].ID == id {
			return l.rows[i].DelFlg
		}
	}
	return false
}

// RemoveMarked drops the rows matching pred and keeps the survivors in
// their relative order.
func (l *MissileList) RemoveMarked(pred func(id int) bool) {
	kept := l.rows[:0]
	for _, row := range l.rows {
		if !pred(row.ID) {
			kept = append(kept, row)
		}
	}
	l.rows = kept
}

// AppendMissile appends the missile to the payload and returns a pointer
// to it; a later append can reallocate the payload and invalidate it.
func (l *MissileList) AppendMissile(m Missile) *Missile {
	l.payload = append(l.payload, m)
	l.rows = append(l.rows, Entry{})
	return &l.payload[len(l.payload)-1]
}

// IDs returns the row ids in iteration order.
func (l *MissileList) IDs() []int {
	out := make([]int, len(l.rows))
	for i, row := range l.rows {
		out[i] = row.ID
	}
	return out
}

// Len reports the number of rows.
func (l *MissileList) Len() int { return len(l.rows) }
