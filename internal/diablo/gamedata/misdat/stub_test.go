package misdat_test

import "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"

// StubWorld stands in for the game world until the real world port lands.
type StubWorld struct {
	spawned []*misdat.MissileRef
	deleted []*misdat.MissileRef
}

func NewStubWorld() *StubWorld { return &StubWorld{} }

func (s *StubWorld) SpawnMissile(kind string, x, y int) *misdat.MissileRef {
	m := &misdat.MissileRef{Kind: kind, X: x, Y: y}
	s.spawned = append(s.spawned, m)
	return m
}

func (s *StubWorld) Delete(m *misdat.MissileRef) {
	m.Deleted = true
	s.deleted = append(s.deleted, m)
}

func (s *StubWorld) Missiles() []*misdat.MissileRef { return s.spawned }

func (s *StubWorld) Deleted() []*misdat.MissileRef { return s.deleted }

// Spawn creates the missile under test without recording it as a behaviour spawn.
func (*StubWorld) Spawn(kind string, x, y int) *misdat.MissileRef {
	return &misdat.MissileRef{Kind: kind, X: x, Y: y}
}
