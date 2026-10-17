// Package missiles ports the missile container (Source/missiles.h:238
// std::list<Missile> Missiles; missiles.cpp:2801 emplace_back;
// missiles.cpp:4184-4187 DeleteMissiles remove_if).
package missiles

// Rows is the parallel Missile payload storage: row i carries the
// Missile instance that entered the container at emplace_back i
// (missiles.cpp:2801). Appends go to the tail like the Entry rows;
// Mark/RemoveMarked act on the linked Entry, not the payload — full
// payload GC lands with the DeleteMissiles port in a later slice.
var Rows []Missile

// Entry is one missile row as the container sees it: the id plus the
// _miDelFlag mark (missiles.h:99). Full Missile fields land with the
// missile-system bean; the container semantics are what must match
// std::list.
type Entry struct {
	ID     int
	DelFlg bool
}

// MissileList reproduces std::list<Missile> iteration order: Append is
// emplace_back (tail), RemoveMarked is remove_if (survivors keep order,
// relative to each other), iteration is head-to-tail.
type MissileList struct {
	rows []Entry
}

// Append ports Missiles.emplace_back(): the new missile goes to the tail,
// even while other rows carry the deletion flag.
func (l *MissileList) Append(id int) {
	l.rows = append(l.rows, Entry{ID: id})
}

// Mark sets _miDelFlag on the row with the given id (missiles.cpp:196 et
// al.): the row stays in place and iterable until DeleteMissiles.
func (l *MissileList) Mark(id int) {
	for i := range l.rows {
		if l.rows[i].ID == id {
			l.rows[i].DelFlg = true
			return
		}
	}
}

// Marked reports whether the row with the given id carries _miDelFlag.
func (l *MissileList) Marked(id int) bool {
	for i := range l.rows {
		if l.rows[i].ID == id {
			return l.rows[i].DelFlg
		}
	}
	return false
}

// RemoveMarked ports DeleteMissiles (missiles.cpp:4186): remove_if with
// the given predicate. std::list::remove_if preserves the order of the
// surviving elements — the in-place filter below does the same.
func (l *MissileList) RemoveMarked(pred func(id int) bool) {
	kept := l.rows[:0]
	for _, row := range l.rows {
		if !pred(row.ID) {
			kept = append(kept, row)
		}
	}
	l.rows = kept
}

// AppendMissile appends a full Missile row (payload + container entry in
// one call — the AddMissile emplace_back shape, missiles.cpp:2801-2830).
func (l *MissileList) AppendMissile(m Missile) *Missile {
	Rows = append(Rows, m)
	l.rows = append(l.rows, Entry{})
	return &Rows[len(Rows)-1]
}

// IDs returns the iteration order (head to tail).
func (l *MissileList) IDs() []int {
	out := make([]int, len(l.rows))
	for i, row := range l.rows {
		out[i] = row.ID
	}
	return out
}

// Len reports the number of rows currently in the list.
func (l *MissileList) Len() int { return len(l.rows) }
