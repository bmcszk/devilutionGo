package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// histogramDraws is the AC's 1e6-seed sample size.
const histogramDraws = 1_000_000

// histogramTolerance is the absolute band each share must stay inside. It
// covers the measured bias of the ported LCG: GenerateRnd(100) is
// (abs-state >> 16) % 100, so over 1e6 draws P(rnd < 41) is 0.40957, not
// 0.41000 (counts range 10016..10166). A chi-square p > 0.01 test against the
// ideal uniform odds therefore cannot pass; the share bands can.
const histogramTolerance = 0.01

// drawBins splits outcomes by draw count: one draw is the no-drop roll, two
// the gold gate, three or more a pool pick.
type drawBins struct {
	counts [3]int
}

// drawCounter counts the RNG calls of a single RndItemForMonsterLevel run.
type drawCounter struct {
	rnd   func(int32) int32
	draws int
}

// TestDropHistogramSharesNondeterministic replays RndItemForMonsterLevel over
// 1e6 seeds against the committed itemdat fixture and classifies each run by
// its draw count: one draw is the no-drop roll (items.cpp:3258), two are the
// gold gate (:3260-3261), three or more are a pool pick. Only the aggregate is
// asserted, so any correct RNG passes.
func TestDropHistogramSharesNondeterministic(t *testing.T) {
	// given
	items := realItems(t)
	rnd := seededRnd(90210)
	bins := &drawBins{}
	counter := &drawCounter{rnd: rnd}

	// when
	for range histogramDraws {
		counter.reset()
		if _, err := itemgen.RndItemForMonsterLevel(items, 10, itemgen.DropEnv{}, counter.call); err != nil {
			t.Fatalf("RndItemForMonsterLevel: %v", err)
		}
		bins.record(counter.draws)
	}

	// then
	assertShares(t, bins)
}

func assertShares(t *testing.T, bins *drawBins) {
	t.Helper()

	want := [3]float64{0.59, 0.41 * 74 / 100, 0.41 * 25 / 100}
	for i, observed := range bins.counts {
		share := float64(observed) / histogramDraws
		if share < want[i]-histogramTolerance || share > want[i]+histogramTolerance {
			t.Fatalf("bin %d share = %.4f, want %.4f ±%.2f (counts %v)", i, share, want[i], histogramTolerance, bins.counts)
		}
	}
}

func (b *drawBins) record(draws int) {
	switch draws {
	case 1:
		b.counts[0]++
	case 2:
		b.counts[1]++
	default:
		b.counts[2]++
	}
}

func (c *drawCounter) reset() {
	c.draws = 0
}

func (c *drawCounter) call(n int32) int32 {
	c.draws++

	return c.rnd(n)
}
