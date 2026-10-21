package misdat_test

import "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"

// StubWorld stands in for the game world until the real world port lands.
type StubWorld struct {
	spawned []*misdat.MissileRef
	deleted []*misdat.MissileRef
}

func NewStubWorld() *StubWorld { return &StubWorld{} }

func (s *StubWorld) SpawnMissile(kind string, x, y int) *misdat.MissileRef {
	m := newMissileRef(kind, x, y)
	s.spawned = append(s.spawned, m)
	return m
}

func (s *StubWorld) Delete(m *misdat.MissileRef) {
	m.Deleted = true
	s.deleted = append(s.deleted, m)
}

func (s *StubWorld) Missiles() []*misdat.MissileRef { return s.spawned }

func (s *StubWorld) Deleted() []*misdat.MissileRef { return s.deleted }

// Spawn returns the missile under test without recording it, so it never
// shows up in Missiles.
func (*StubWorld) Spawn(kind string, x, y int) *misdat.MissileRef {
	return newMissileRef(kind, x, y)
}

func newMissileRef(kind string, x, y int) *misdat.MissileRef {
	return &misdat.MissileRef{Kind: kind, X: x, Y: y}
}
