package engine_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// crawlOrderCases is the visitation-order table: full crawl to radius 2,
// single ring 3, and the ring-2 prefix the order test pins.
var crawlOrderCases = []crawlOrderCase{
	{
		name:      "FullCrawlToRadius2",
		minRadius: 0, maxRadius: 2,
		wantOrder: []engine.Displacement{
			{DeltaX: 0, DeltaY: 0},
			{DeltaX: 0, DeltaY: 1},
			{DeltaX: 0, DeltaY: -1},
			{DeltaX: -1, DeltaY: 0},
			{DeltaX: 1, DeltaY: 0},
			{DeltaX: 0, DeltaY: 2},
			{DeltaX: 0, DeltaY: -2},
			{DeltaX: -1, DeltaY: 2},
			{DeltaX: 1, DeltaY: 2},
			{DeltaX: -1, DeltaY: -2},
			{DeltaX: 1, DeltaY: -2},
			{DeltaX: -1, DeltaY: 1},
			{DeltaX: 1, DeltaY: 1},
			{DeltaX: -1, DeltaY: -1},
			{DeltaX: 1, DeltaY: -1},
			{DeltaX: -2, DeltaY: 0},
			{DeltaX: 2, DeltaY: 0},
			{DeltaX: -2, DeltaY: 1},
			{DeltaX: 2, DeltaY: 1},
			{DeltaX: -2, DeltaY: -1},
			{DeltaX: 2, DeltaY: -1},
		},
	},
	{
		name:      "SingleRingThreeFollowsTheRingShape",
		minRadius: 3, maxRadius: 3,
		wantOrder: crawlCoversRing(3),
	},
	{
		name:      "SingleRingTwoIsTheOrderTestPrefix",
		minRadius: 2, maxRadius: 2,
		wantOrder: []engine.Displacement{
			{DeltaX: 0, DeltaY: 2},
			{DeltaX: 0, DeltaY: -2},
			{DeltaX: -1, DeltaY: 2},
			{DeltaX: 1, DeltaY: 2},
			{DeltaX: -1, DeltaY: -2},
			{DeltaX: 1, DeltaY: -2},
			{DeltaX: -1, DeltaY: 1},
			{DeltaX: 1, DeltaY: 1},
			{DeltaX: -1, DeltaY: -1},
			{DeltaX: 1, DeltaY: -1},
			{DeltaX: -2, DeltaY: 0},
			{DeltaX: 2, DeltaY: 0},
			{DeltaX: -2, DeltaY: 1},
			{DeltaX: 2, DeltaY: 1},
			{DeltaX: -2, DeltaY: -1},
			{DeltaX: 2, DeltaY: -1},
		},
	},
}

// crawlState describes one grid point's coverage and corner exemption.
type crawlState struct {
	added        bool
	exemptCorner bool
}

// crawlStopCase — wantFound stays false whenever the only "valid" tile is one
// the spiral never probes (corner inset), mirroring path_test.cpp:262-271.
type crawlStopCase struct {
	name      string
	minRadius uint
	maxRadius uint
	valid     []engine.Displacement
	wantFound bool
	want      engine.Displacement
}

// crawlResult pins a crawl outcome: the found flag and displacement.
type crawlResult struct {
	found bool
	got   engine.Displacement
}

// crawlOrderCase is one visitation-order expectation.
type crawlOrderCase struct {
	name      string
	minRadius uint
	maxRadius uint
	wantOrder []engine.Displacement
}

func TestDoCrawlBasicCoverage(t *testing.T) {
	// given
	const maxR = 18

	origin := maxR
	size := 2*maxR + 1

	added := make([][]bool, size)
	for x := range added {
		added[x] = make([]bool, size)
	}
	// when
	order := collectCrawl(t, 0, maxR)

	// then
	for _, d := range order {
		x, y := origin+d.DeltaX, origin+d.DeltaY
		if added[x][y] {
			t.Fatalf("displacement %d:%d visited twice", d.DeltaX, d.DeltaY)
		}

		added[x][y] = true
	}

	verifyCrawlCoverage(t, added, maxR)
}

func TestDoCrawlVisitationOrder(t *testing.T) {
	for _, tt := range crawlOrderCases {
		t.Run(tt.name, func(t *testing.T) {
			got := collectCrawl(t, tt.minRadius, tt.maxRadius)
			if !slices.Equal(got, tt.wantOrder) {
				t.Fatalf("engine.DoCrawl(%d, %d) visited %v, want %v", tt.minRadius, tt.maxRadius, got, tt.wantOrder)
			}
		})
	}
}

func TestCrawlStopsAtFirstValidPosition(t *testing.T) {
	tests := []crawlStopCase{
		{
			name:      "StartTileWhenAlwaysValid",
			minRadius: 0, maxRadius: 5,
			want: engine.Displacement{DeltaX: 0, DeltaY: 0},
		},
		{
			name: "MinRadiusSkipsInnerTiles", minRadius: 2, maxRadius: 5,
			valid:     []engine.Displacement{{DeltaX: 0, DeltaY: 2}},
			wantFound: true,
			want:      engine.Displacement{DeltaX: 0, DeltaY: 2},
		},
		{
			name: "LaterRingTile", minRadius: 1, maxRadius: 3,
			valid:     []engine.Displacement{{DeltaX: 1, DeltaY: -2}},
			wantFound: true,
			want:      engine.Displacement{DeltaX: 1, DeltaY: -2},
		},
		{
			name: "CornerRoundedFromSearch", minRadius: 1, maxRadius: 2,
			valid:     []engine.Displacement{{DeltaX: 2, DeltaY: 2}},
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstValid := memberOf(tt.valid)
			got, found := engine.Crawl(tt.minRadius, tt.maxRadius, firstValid)
			assertCrawlResult(t, tt, got, found)
		})
	}
}

func verifyCrawlCoverage(t *testing.T, added [][]bool, maxR int) {
	t.Helper()

	origin := maxR
	for i := -maxR; i <= maxR; i++ {
		for j := -maxR; j <= maxR; j++ {
			assertCrawled(t, crawlState{
				added:        added[i+origin][j+origin],
				exemptCorner: max(i, -i) == maxR && max(j, -j) == maxR,
			}, i, j)
		}
	}
}

// assertCrawled fails when an inner grid point was never crawled; corner
// diagonal points (maxRadial == maxR on both axes) are exempt.
func assertCrawled(t *testing.T, st crawlState, i, j int) {
	t.Helper()

	if st.added || st.exemptCorner {
		return
	}

	t.Fatalf("location %d:%d was never crawled", i, j)
}

// memberOf builds a validity predicate matching any displacement in the list.
func memberOf(valid []engine.Displacement) func(engine.Displacement) (engine.Displacement, bool) {
	return func(d engine.Displacement) (engine.Displacement, bool) {
		if slices.Contains(valid, d) {
			return d, true
		}

		return engine.Displacement{}, false
	}
}

// assertCrawlResult pins the crawl result against the case's expectations.
func assertCrawlResult(t *testing.T, tt crawlStopCase, got engine.Displacement, found bool) {
	t.Helper()
	assertCrawlOutcome(t, tt, crawlResult{found: found, got: got})
}

// assertCrawlOutcome compares a crawlResult with the case expectations.
func assertCrawlOutcome(t *testing.T, tt crawlStopCase, res crawlResult) {
	t.Helper()

	if res.found != tt.wantFound {
		t.Fatalf("crawl found = %v, want %v", res.found, tt.wantFound)
	}

	if res.got != tt.want {
		t.Fatalf("crawl returned %v, want %v", res.got, tt.want)
	}
}

func collectCrawl(t *testing.T, minRadius, maxRadius uint) []engine.Displacement {
	t.Helper()

	var order []engine.Displacement

	engine.DoCrawl(minRadius, maxRadius, func(d engine.Displacement) bool {
		order = append(order, d)
		return true
	})

	return order
}

// crawl covers every tile whose max(|dx|,|dy|) == r except the four extreme
// corners at that radius (crawl.cpp:16-49 rounds the ring by one tile).
func crawlCoversRing(r int) []engine.Displacement {
	var tiles []engine.Displacement

	tiles = append(tiles, engine.Displacement{DeltaX: 0, DeltaY: r}, engine.Displacement{DeltaX: 0, DeltaY: -r})
	for x := 1; x < r; x++ {
		for _, y := range []int{r, -r} {
			tiles = append(tiles, engine.Displacement{DeltaX: -x, DeltaY: y}, engine.Displacement{DeltaX: x, DeltaY: y})
		}
	}

	if r > 1 {
		d := r - 1
		tiles = append(tiles,
			engine.Displacement{DeltaX: -d, DeltaY: d}, engine.Displacement{DeltaX: d, DeltaY: d},
			engine.Displacement{DeltaX: -d, DeltaY: -d}, engine.Displacement{DeltaX: d, DeltaY: -d})
	}

	tiles = append(tiles, engine.Displacement{DeltaX: -r, DeltaY: 0}, engine.Displacement{DeltaX: r, DeltaY: 0})
	for y := 1; y < r; y++ {
		tiles = append(tiles,
			engine.Displacement{DeltaX: -r, DeltaY: y}, engine.Displacement{DeltaX: r, DeltaY: y},
			engine.Displacement{DeltaX: -r, DeltaY: -y}, engine.Displacement{DeltaX: r, DeltaY: -y})
	}

	return tiles
}
