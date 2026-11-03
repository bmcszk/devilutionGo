package kernel

import (
	"slices"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Streams is the kernel's RNG bookkeeping: one seeded engine generator per
// named stream, so a replay draws from the same sequence every run (ADR-015).
// Stream names are sorted, so no map iteration decides a draw order.
type Streams struct {
	names  []string
	byName map[string]*engine.Xoshiro128pp
}

// NewStreams seeds one generator per name; name i gets seed+i so the streams
// never share a sequence.
func NewStreams(seed uint32, names ...string) *Streams {
	sorted := slices.Clone(names)
	slices.Sort(sorted)

	byName := make(map[string]*engine.Xoshiro128pp, len(sorted))
	for i, name := range sorted {
		byName[name] = engine.NewXoshiro128pp(seed + uint32(i))
	}

	return &Streams{names: sorted, byName: byName}
}

// Names returns the stream names in draw order.
func (s *Streams) Names() []string {
	if s == nil {
		return nil
	}

	return slices.Clone(s.names)
}

// Draw advances the named stream once. An unknown stream draws nothing and
// reports false, so a handler stays honest instead of silently inventing a
// generator.
func (s *Streams) Draw(name string) (uint32, bool) {
	if s == nil {
		return 0, false
	}

	gen, ok := s.byName[name]
	if !ok {
		return 0, false
	}

	return gen.Next(), true
}
