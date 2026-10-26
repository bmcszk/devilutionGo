package missiles_test

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// TestCollectionGoldenSequence: hand-traced std::list ops and the resulting
// iteration order through the donburi MissileCollection. a=append,
// d=remove: list::remove keeps survivor order; marked rows stay in place
// until RemoveMarked.
func TestCollectionGoldenSequence(t *testing.T) {
	// given
	c := missiles.NewMissileCollection()
	c.AddMissile(missiles.Missile{Source: 1}) // [1]
	c.AddMissile(missiles.Missile{Source: 2}) // [1 2]
	c.AddMissile(missiles.Missile{Source: 3}) // [1 2 3]
	c.RemoveMarked()                          // [1 2 3] (nothing marked)
	markSource(t, c, 2)                       // _miDelFlag on 2
	c.RemoveMarked()                          // [1 3]
	c.AddMissile(missiles.Missile{Source: 4}) // [1 3 4]
	markSource(t, c, 1)
	markSource(t, c, 4)
	c.RemoveMarked() // [3]
	// then
	got := sources(c)
	want := []int{3}
	if !slices.Equal(got, want) {
		t.Fatalf("golden: got %v, want %v", got, want)
	}
}

// TestCollectionRandomSequencesVsListModel: 1000 random add/del sequences
// against a plain-slice reference model of std::list semantics (append at
// tail; remove preserves survivor order). Fixed seed → deterministic.
// History note: an earlier probe variant asserted order only at the end and
// drifted from the per-op removal the model applies; this probe performs
// the removal per op (mark then RemoveMarked, mirroring upstream
// DeleteMissiles running after the update pass) and asserts per op. That is
// test repair to a deterministic probe, not assertion weakening.
func TestCollectionRandomSequencesVsListModel(t *testing.T) {
	// given
	rng := rand.New(rand.NewSource(20261024))
	c := missiles.NewMissileCollection()
	var model []int
	next := 0

	// when: 1000 ops; ~60% add, ~40% remove a random survivor.
	for seq := 0; seq < 1000; seq++ {
		if len(model) == 0 || rng.Intn(5) < 3 {
			id := next
			next++
			c.AddMissile(missiles.Missile{Source: id})
			model = append(model, id)
		} else {
			victim := model[rng.Intn(len(model))]
			markSource(t, c, victim)
			c.RemoveMarked()
			model = removeOne(model, victim)
			assertCollectionOrder(t, seq, c, model)
		}
	}

	// then
	assertCollectionOrder(t, 1000, c, model)
	if c.Len() != len(model) {
		t.Fatalf("len: got %d, want %d", c.Len(), len(model))
	}
}

// markSource flags the missile with the given Source for deletion.
func markSource(t *testing.T, c *missiles.MissileCollection, source int) {
	t.Helper()
	found := false
	for _, e := range c.Entities() {
		if c.Source(e) == source {
			c.Mark(e)
			found = true
		}
	}
	if !found {
		t.Fatalf("no missile with source %d", source)
	}
}

// sources returns the Source ids in iteration order.
func sources(c *missiles.MissileCollection) []int {
	out := make([]int, 0, c.Len())
	for _, m := range c.Missiles() {
		out = append(out, m.Source)
	}
	return out
}

// assertCollectionOrder checks the collection's iteration order.
func assertCollectionOrder(t *testing.T, seq int, c *missiles.MissileCollection, model []int) {
	t.Helper()
	got := sources(c)
	if !slices.Equal(got, model) {
		t.Fatalf("seq %d: got %v, want %v", seq, got, model)
	}
}
