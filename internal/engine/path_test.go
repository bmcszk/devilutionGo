package engine_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: diasurgical/DevilutionX test/path_test.cpp. Path bytes are
// GetPathDirection codes (path.cpp:178-181), not PathDirs indices:
// 1=up 2=left 3=right 4=down 5=up-left 6=up-right 7=down-right 8=down-left
// in tile coordinates (y grows south).
var stepDelta = [9]engine.Point{
	{},
	{X: 0, Y: -1},
	{X: -1, Y: 0},
	{X: 1, Y: 0},
	{X: 0, Y: 1},
	{X: -1, Y: -1},
	{X: 1, Y: -1},
	{X: 1, Y: 1},
	{X: -1, Y: 1},
}

func alwaysCanStep(engine.Point, engine.Point) bool { return true }

func walkablePosOk(grid [][]bool) func(engine.Point) bool {
	return func(p engine.Point) bool {
		return p.X >= 0 && p.X < len(grid) && p.Y >= 0 && p.Y < len(grid[p.X]) && grid[p.X][p.Y]
	}
}

// makeObstacleGrid is deterministic per seed: terrain comes from the engine LCG.
func makeObstacleGrid(seed uint32, size, wallPercent int) [][]bool {
	engine.SetRndSeed(seed)
	grid := make([][]bool, size)
	for x := range grid {
		grid[x] = make([]bool, size)
		for y := range grid[x] {
			grid[x][y] = engine.GenerateRnd(100) >= int32(wallPercent)
		}
	}
	return grid
}

func checkPath(t *testing.T, start, dest engine.Point, want []int8) {
	t.Helper()
	// 24 is the longest path the vanilla game can express (path_test.cpp:96).
	const maxPathLength = 24
	path := make([]int8, maxPathLength)
	// when
	length := engine.FindPath(alwaysCanStep, func(engine.Point) bool { return true }, start, dest, path)
	// then
	if length < 0 || length > maxPathLength {
		t.Fatalf("FindPath(%v, %v) returned length %d", start, dest, length)
	}
	if length > 0 {
		requirePadding(t, path, length)
	}
	if got := path[:length]; !slices.Equal(got, want) {
		t.Fatalf("FindPath(%v, %v) = %v, want %v", start, dest, got, want)
	}
}

func requirePadding(t *testing.T, path []int8, length int) {
	t.Helper()
	for i := length; i < len(path); i++ {
		if path[i] != -1 {
			t.Fatalf("path[%d] = %d, want -1 padding", i, path[i])
		}
	}
}

func TestPathHeuristics(t *testing.T) {
	// given
	source := engine.Point{X: 25, Y: 32}
	tests := []struct {
		name string
		dest engine.Point
		want int
	}{
		{"same tile", source, 0},
		{"NorthEast", source.Plus(engine.Point{X: 0, Y: -1}), engine.PathAxisAlignedStepCost},
		{"SouthWest", source.Plus(engine.Point{X: 0, Y: 1}), engine.PathAxisAlignedStepCost},
		{"SouthEast", source.Plus(engine.Point{X: 1, Y: 0}), engine.PathAxisAlignedStepCost},
		{"NorthWest", source.Plus(engine.Point{X: -1, Y: 0}), engine.PathAxisAlignedStepCost},
		{"North", source.Plus(engine.Point{X: -1, Y: -1}), engine.PathDiagonalStepCost},
		{"East", source.Plus(engine.Point{X: 1, Y: -1}), engine.PathDiagonalStepCost},
		{"South", source.Plus(engine.Point{X: 1, Y: 1}), engine.PathDiagonalStepCost},
		{"West", source.Plus(engine.Point{X: -1, Y: 1}), engine.PathDiagonalStepCost},
		{"SouthWest+SouthEast", source.Plus(engine.Point{X: 1, Y: 1}), engine.PathDiagonalStepCost},
		{
			"NorthEast+North",
			source.Plus(engine.Point{X: -1, Y: -2}),
			engine.PathAxisAlignedStepCost + engine.PathDiagonalStepCost,
		},
		{"SouthEast+SouthEast", source.Plus(engine.Point{X: 2, Y: 0}), 2 * engine.PathAxisAlignedStepCost},
	}
	for _, tt := range tests {
		if got := engine.TestPathGetHeuristicCost(source, tt.dest); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestFindPath(t *testing.T) {
	// given
	tests := []struct {
		name  string
		start engine.Point
		dest  engine.Point
		want  []int8
	}{
		{"FindPathToSelf", engine.Point{X: 8, Y: 8}, engine.Point{X: 8, Y: 8}, nil},
		{"FindPathTwoStepsUp", engine.Point{X: 8, Y: 8}, engine.Point{X: 8, Y: 6}, []int8{1, 1}},
		{"FindPathTwoStepsLeft", engine.Point{X: 8, Y: 8}, engine.Point{X: 6, Y: 8}, []int8{2, 2}},
		{"FindPathTwoStepsRight", engine.Point{X: 8, Y: 8}, engine.Point{X: 10, Y: 8}, []int8{3, 3}},
		{"FindPathTwoStepsDown", engine.Point{X: 8, Y: 8}, engine.Point{X: 8, Y: 10}, []int8{4, 4}},
		{"FindPathDiagonalsFirst3Left2Up", engine.Point{X: 8, Y: 8}, engine.Point{X: 5, Y: 6}, []int8{5, 5, 2}},
		{"FindPathDiagonalsFirst4Left4Up", engine.Point{X: 8, Y: 8}, engine.Point{X: 4, Y: 4}, []int8{5, 5, 5, 5}},
		{"FindPathDiagonalsFirst2Right4Down", engine.Point{X: 8, Y: 8}, engine.Point{X: 10, Y: 12}, []int8{7, 7, 4, 4}},
		{"FindPathDiagonalsFirst4Right12Down", engine.Point{X: 8, Y: 8}, engine.Point{X: 12, Y: 20},
			[]int8{7, 7, 7, 7, 4, 4, 4, 4, 4, 4, 4, 4}},
		{"LongPathsToBorder", engine.Point{X: 56, Y: 56}, engine.Point{X: 0, Y: 0}, nil},
		{"LongPaths24Units", engine.Point{X: 56, Y: 56}, engine.Point{X: 80, Y: 80}, slices.Repeat([]int8{7}, 24)},
		{"LongPaths25Units", engine.Point{X: 56, Y: 56}, engine.Point{X: 81, Y: 81}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkPath(t, tt.start, tt.dest, tt.want)
		})
	}
}

type findClosestCase struct {
	name            string
	size            int
	start           engine.Point
	minimumRadius   uint
	maximumRadius   uint
	wantFound       bool
	wantPosition    engine.Point
	wantSearchCount func(x, y, size int) int
}

func TestFindClosestValidPosition(t *testing.T) {
	// given
	tests := []findClosestCase{
		{
			name: "NoValidTilesUpToRadius50SkipsExtremeCorners",
			size: 101, start: engine.Point{X: 50, Y: 50}, minimumRadius: 0, maximumRadius: 50,
			wantSearchCount: ringExceptExtremeCorners,
		},
		{
			name: "MinRadiusOneSkipsStartAndCorners",
			size: 5, start: engine.Point{X: 2, Y: 2}, minimumRadius: 1, maximumRadius: 2,
			wantSearchCount: ringExceptStartAndExtremeCorners,
		},
		{
			name: "RadiusZeroSearchesOnlyTheStart",
			size: 3, start: engine.Point{X: 1, Y: 1}, minimumRadius: 0, maximumRadius: 0,
			wantSearchCount: onlyTheStart,
		},
		{
			name: "FixedRadiusMakesSquareWithInsetCorners",
			size: 7, start: engine.Point{X: 3, Y: 3}, minimumRadius: 3, maximumRadius: 3,
			wantSearchCount: squareWithInsetCorners,
		},
		{
			name: "FirstCandidateWithMinRadiusIsAlongPositiveY",
			size: 0, start: engine.Point{X: 50, Y: 50}, minimumRadius: 21, maximumRadius: 50,
			wantFound: true, wantPosition: engine.Point{X: 50, Y: 71},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantSearchCount == nil {
				verifyFirstFind(t, tt)
				return
			}
			verifySearchCounts(t, tt)
		})
	}
}

func ringExceptExtremeCorners(x, y, size int) int {
	if (x == 0 || x == size-1) && (y == 0 || y == size-1) {
		return 0
	}
	return 1
}

func ringExceptStartAndExtremeCorners(x, y, size int) int {
	if x == 2 && y == 2 {
		return 0
	}
	return ringExceptExtremeCorners(x, y, size)
}

func onlyTheStart(x, y, _ int) int {
	if x == 1 && y == 1 {
		return 1
	}
	return 0
}

func squareWithInsetCorners(x, y, size int) int {
	last := size - 1
	insetCorners := (x == 1 || x == last-1) && (y == 1 || y == last-1)
	outerSides := (x == 0 || x == last) && y > 0 && y < last ||
		(y == 0 || y == last) && x > 0 && x < last
	if insetCorners || outerSides {
		return 1
	}
	return 0
}

func verifyFirstFind(t *testing.T, tt findClosestCase) {
	t.Helper()
	// when
	pos, found := engine.FindClosestValidPosition(func(engine.Point) bool { return true },
		tt.start, tt.minimumRadius, tt.maximumRadius)
	// then
	if found != tt.wantFound || pos != tt.wantPosition {
		t.Fatalf("FindClosestValidPosition() = (%v, %v), want (%v, %v)", pos, found, tt.wantPosition, tt.wantFound)
	}
}

func verifySearchCounts(t *testing.T, tt findClosestCase) {
	t.Helper()
	searched := make([][]int, tt.size)
	for x := range searched {
		searched[x] = make([]int, tt.size)
	}
	// when
	_, found := engine.FindClosestValidPosition(func(p engine.Point) bool {
		searched[p.X][p.Y]++
		return false
	}, tt.start, tt.minimumRadius, tt.maximumRadius)
	// then
	if found {
		t.Fatal("FindClosestValidPosition found a position; want none")
	}
	verifySearched(t, searched, tt)
}

func verifySearched(t *testing.T, searched [][]int, tt findClosestCase) {
	t.Helper()
	for x := range searched {
		for y := range searched[x] {
			if want := tt.wantSearchCount(x, y, tt.size); searched[x][y] != want {
				t.Fatalf("searched[%d][%d] = %d, want %d", x, y, searched[x][y], want)
			}
		}
	}
}

func verifySeededGridPath(t *testing.T, seed uint32, size, wallPercent int) bool {
	t.Helper()
	grid := makeObstacleGrid(seed, size, wallPercent)
	start := engine.Point{X: 1, Y: 1}
	dest := engine.Point{X: size - 2, Y: size - 2}
	grid[start.X][start.Y] = true
	grid[dest.X][dest.Y] = true
	path := make([]int8, engine.MaxPathLengthPlayer)
	length := engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	if length == 0 {
		return false
	}
	verifyPathWalk(t, seed, grid, start, dest, path[:length])
	requirePadding(t, path, length)
	return true
}

func verifyPathWalk(t *testing.T, seed uint32, grid [][]bool, start, dest engine.Point, steps []int8) {
	t.Helper()
	walkable := walkablePosOk(grid)
	cur := start
	for i, step := range steps {
		cur = verifyStep(t, seed, i, step, cur, walkable)
	}
	if cur != dest {
		t.Fatalf("seed %d: path ends at %v, want %v", seed, cur, dest)
	}
}

func verifyStep(
	t *testing.T,
	seed uint32,
	i int,
	step int8,
	cur engine.Point,
	walkable func(engine.Point) bool,
) engine.Point {
	t.Helper()
	if step < 1 || step > 8 {
		t.Fatalf("seed %d: step %d is invalid direction byte %d", seed, i, step)
	}
	next := cur.Plus(stepDelta[step])
	if got := engine.GetPathDirection(cur, next); got != step {
		t.Fatalf("seed %d: step %d byte %d != GetPathDirection %d", seed, i, step, got)
	}
	if !walkable(next) {
		t.Fatalf("seed %d: step %d lands on blocked tile %v", seed, i, next)
	}
	return next
}

func TestFindPathDirectionByteParityOnSeededGrids(t *testing.T) {
	// given
	const (
		size        = 40
		gridCount   = 100
		wallPercent = 20
	)
	// when
	found := 0
	for seed := 1; seed <= gridCount; seed++ {
		if verifySeededGridPath(t, uint32(seed), size, wallPercent) {
			found++
		}
	}
	// then
	if found < 90 {
		t.Fatalf("only %d of %d seeded grids yielded a path; grid generation regressed?", found, gridCount)
	}
}
