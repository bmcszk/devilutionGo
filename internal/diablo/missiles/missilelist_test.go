// Package missiles_test ports the AC from devilGo-r462: the missile
// container must reproduce std::list<Missile> iteration order for 1000
// random add/delete sequences (missiles.h:238, missiles.cpp:4184-4187,
// 2801). The Go MissileList (append + in-place remove_if) is checked
// against a reference model of std::list semantics, and both are checked
// against a scripted golden sequence.
package missiles_test

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// golden: hand-traced std::list ops and the resulting iteration order.
// a=append(i), d=remove_if(id==i): list::remove keeps order of survivors.
func TestGoldenSequence(t *testing.T) {
	l := &missiles.MissileList{}
	l.Append(1)                                                     // [1]
	l.Append(2)                                                     // [1 2]
	l.Append(3)                                                     // [1 2 3]
	l.RemoveMarked(func(id int) bool { return id == 2 })            // [1 3]
	l.Append(4)                                                     // [1 3 4]
	l.Append(5)                                                     // [1 3 4 5]
	l.RemoveMarked(func(id int) bool { return id == 1 })            // [3 4 5]
	l.Append(6)                                                     // [3 4 5 6]
	l.RemoveMarked(func(id int) bool { return id == 4 || id == 6 }) // [3 5]
	got := l.IDs()
	want := []int{3, 5}
	if !slices.Equal(got, want) {
		t.Fatalf("golden: got %v, want %v", got, want)
	}
}

// TestRandomSequencesVsListModel: 1000 random add/del sequences; the
// container's iteration order must equal a direct model of std::list
// semantics (append at tail; remove_if preserves survivor order).
func TestRandomSequencesVsListModel(t *testing.T) {
	rng := rand.New(rand.NewSource(988045466)) // fixed seed for reproducibility
	for seq := range 1000 {
		l := &missiles.MissileList{}
		model := runRandomOps(l, rng, 200)
		assertOrder(t, seq, l, model)
	}
}

// runRandomOps drives one 200-op random add/delete sequence on the
// container and its std::list model in lockstep.
func runRandomOps(l *missiles.MissileList, rng *rand.Rand, ops int) []int {
	var model []int
	next := 0
	for range ops {
		if len(model) == 0 || rng.Intn(2) == 0 {
			id := next
			next++
			l.Append(id)
			model = append(model, id)
			continue
		}
		victim := model[rng.Intn(len(model))]
		l.RemoveMarked(func(id int) bool { return id == victim })
		model = removeOne(model, victim)
	}
	return model
}

// removeOne drops the first occurrence of id (model of list::remove).
func removeOne(ids []int, id int) []int {
	var kept []int
	for _, x := range ids {
		if x != id {
			kept = append(kept, x)
		}
	}
	return kept
}

// assertOrder checks the container's iteration order against the model.
func assertOrder(t *testing.T, seq int, l *missiles.MissileList, model []int) {
	t.Helper()
	got := l.IDs()
	if !slices.Equal(got, model) {
		t.Fatalf("seq %d: got %v, want %v", seq, got, model)
	}
}

// TestNewMissileAppendsLikeEmplaceBack: AddMissile's emplace_back port
// appends at the tail even while the list holds marked-for-deletion rows.
func TestNewMissileAppendsLikeEmplaceBack(t *testing.T) {
	l := &missiles.MissileList{}
	l.Append(0)
	l.Append(1)
	l.Mark(1) // _miDelFlag = true, still present (removed only in DeleteMissiles)
	l.Append(2)
	got := l.IDs()
	if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("marked rows stay in place until DeleteMissiles: got %v", got)
	}
	l.RemoveMarked(func(id int) bool { return l.Marked(id) })
	got = l.IDs()
	if len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("after remove_if: got %v, want [0 2]", got)
	}
}
