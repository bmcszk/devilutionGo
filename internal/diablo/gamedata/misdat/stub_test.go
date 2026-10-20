package misdat_test

import "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"

// StubWorld is the scripted world for behaviour tests in this bean; the real
// world port lands with devilGo-mwlm.
type StubWorld struct {
	spawned []*misdat.MissileRef
	deleted []*misdat.MissileRef
}

// NewStubWorld builds an empty scripted world.
func NewStubWorld() *StubWorld { return &StubWorld{} }

// SpawnMissile records a spawn at (x,y).
func (s *StubWorld) SpawnMissile(kind string, x, y int) *misdat.MissileRef {
	m := &misdat.MissileRef{Kind: kind, X: x, Y: y}
	s.spawned = append(s.spawned, m)
	return m
}

// Delete marks a missile deleted.
func (s *StubWorld) Delete(m *misdat.MissileRef) {
	m.Deleted = true
	s.deleted = append(s.deleted, m)
}

// Missiles returns the behaviour-spawned missiles so far.
func (s *StubWorld) Missiles() []*misdat.MissileRef { return s.spawned }

// Deleted returns missiles the behaviours deleted.
func (s *StubWorld) Deleted() []*misdat.MissileRef { return s.deleted }

// Spawn seeds a missile WITHOUT recording it as a behaviour spawn (creates
// the self missile under test).
func (*StubWorld) Spawn(kind string, x, y int) *misdat.MissileRef {
	return &misdat.MissileRef{Kind: kind, X: x, Y: y}
}
