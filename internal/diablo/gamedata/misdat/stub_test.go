// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package misdat

// StubWorld stands in for the game world until the real world port lands.
type StubWorld struct {
	spawned []*missileRef
	deleted []*missileRef
}

func NewStubWorld() *StubWorld { return &StubWorld{} }

func (s *StubWorld) SpawnMissile(kind string, x, y int) *missileRef {
	m := newMissileRef(kind, x, y)
	s.spawned = append(s.spawned, m)
	return m
}

func (s *StubWorld) Delete(m *missileRef) {
	m.Deleted = true
	s.deleted = append(s.deleted, m)
}

func (s *StubWorld) Missiles() []*missileRef { return s.spawned }

func (s *StubWorld) Deleted() []*missileRef { return s.deleted }

// Spawn returns the missile under test without recording it, so it never
// shows up in Missiles.
func (*StubWorld) Spawn(kind string, x, y int) *missileRef {
	return newMissileRef(kind, x, y)
}

func newMissileRef(kind string, x, y int) *missileRef {
	return &missileRef{Kind: kind, X: x, Y: y}
}
