package engine

// Port of diasurgical/DevilutionX Source/engine/path.cpp (1:1 semantics).

import "slices"

// PathAxisAlignedStepCost (path.cpp:32)
const PathAxisAlignedStepCost = 100

// PathDiagonalStepCost (path.cpp:33)
const PathDiagonalStepCost = 101

// maxPathNodes (path.cpp:37)
const maxPathNodes = 1024

// bucketCapacity (path.cpp:63)
const bucketCapacity = 3 * maxPathNodes / 64

// MaxPathLengthMonsters (path.h:30)
const MaxPathLengthMonsters = 25

// MaxPathLengthPlayer (path.h:31)
const MaxPathLengthPlayer = 100

// PathDirs (path.h:42-53).
var PathDirs = [8]Point{
	{X: -1, Y: -1},
	{X: -1, Y: 1},
	{X: 1, Y: -1},
	{X: 1, Y: 1},
	{X: -1, Y: 0},
	{X: 1, Y: 0},
	{X: 0, Y: -1},
	{X: 0, Y: 1},
}

// pathDirections (path.cpp:178-181)
var pathDirections = [9]int8{5, 1, 6, 2, 0, 3, 8, 4, 7}

// Point (point.hpp:19) — dungeon tile coordinates, y grows south.
type Point struct{ X, Y int }

// Plus (point.hpp:193)
func (p Point) Plus(d Point) Point { return Point{X: p.X + d.X, Y: p.Y + d.Y} }

// point8 (path.cpp:39-40): upstream narrows coordinates to uint8 inside the
// pathfinder, so coordinate arithmetic wraps at 256 exactly as in C++.
type point8 struct{ x, y uint8 }

func packPoint(p point8) uint16 { return uint16(p.x)<<8 | uint16(p.y) }

func unpackPoint(packed uint16) point8 { return point8{x: uint8(packed >> 8), y: uint8(packed)} }

func pointOf(packed uint16) Point { return Point{X: int(packed >> 8), Y: int(packed & 0xFF)} }

// frontierNode (path.cpp:44-49); h caches the comparator's heuristic
// (path.cpp:209-213) and diag the step shape the comparator reads from the
// explored set (path.cpp:215-218) — both are fixed when the entry is pushed,
// and a g-improvement always pushes a fresh entry.
type frontierNode struct {
	position uint16
	f        int16
	h        int16
	diag     bool
}

// exploredEntry (path.cpp:52-57) keyed by the packed (x<<8)|y repr of path.cpp:110-112.
type exploredEntry struct {
	key  uint16
	prev uint16
	g    int32
}

// exploredNodes (path.cpp:61-113) — 64 buckets keyed by (x&7)<<3 | y&7;
// linear scan inside the bucket mirrors the upstream StaticVector lookup.
type exploredNodes struct {
	buckets [64][bucketCapacity]exploredEntry
	sizes   [64]int
}

func bucketIndex(packed uint16) int {
	return (int(packed>>8)&0b111)<<3 | int(packed)&0b111
}

func (e *exploredNodes) find(packed uint16) *exploredEntry {
	i := bucketIndex(packed)
	n := e.sizes[i]
	b := e.buckets[i][:n]
	for j := range b {
		if b[j].key == packed {
			return &b[j]
		}
	}
	return nil
}

func (e *exploredNodes) emplace(packed uint16, prev uint16, g int32) {
	i := bucketIndex(packed)
	e.buckets[i][e.sizes[i]] = exploredEntry{key: packed, prev: prev, g: g}
	e.sizes[i]++
}

func (e *exploredNodes) canInsert(packed uint16) bool {
	return e.sizes[bucketIndex(packed)] < bucketCapacity
}

// isDiagonalStep (path.cpp:118-121)
func isDiagonalStep(a, b point8) bool { return a.x != b.x && a.y != b.y }

// getDistance (path.cpp:126-131)
func getDistance(start, dest point8) int32 {
	if isDiagonalStep(start, dest) {
		return PathDiagonalStepCost
	}
	return PathAxisAlignedStepCost
}

// getHeuristicCost (path.cpp:136-152) — admissible: diagonal steps until a
// shared row/column, then the axis-aligned remainder.
func getHeuristicCost(start, dest point8) int32 {
	dx := absInt(int(start.x) - int(dest.x))
	dy := absInt(int(start.y) - int(dest.y))
	return int32(min(dx, dy))*PathDiagonalStepCost + int32(absInt(dx-dy))*PathAxisAlignedStepCost
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// GetPathDirection (path.cpp:176-182).
func GetPathDirection(start, dest Point) int8 {
	return pathDirections[3*(dest.Y-start.Y)+4+dest.X-start.X]
}

// reconstructPath (path.cpp:155-172).
func reconstructPath(explored *exploredNodes, dest uint16, path []int8) int {
	length := 0
	cur := dest
	for {
		entry := explored.find(cur)
		if entry == nil {
			panic("Failed to reconstruct path")
		}
		if entry.g == 0 {
			break
		}
		if length == len(path) {
			length = 0
			break
		}
		path[length] = GetPathDirection(pointOf(entry.prev), pointOf(cur))
		length++
		cur = entry.prev
	}
	return resetPathTail(path, length)
}

// resetPathTail (path.cpp:170-172) reverses the walked prefix.
func resetPathTail(path []int8, length int) int {
	slices.Reverse(path[:length])
	for i := length; i < len(path); i++ {
		path[i] = -1
	}
	return length
}

// findPath (path.cpp:184-283) returns the step count (0 = no path), or -1
// when the reconstructed path exceeds the buffer.
func findPath(
	canStep func(from, to Point) bool,
	posOk func(Point) bool,
	startPosition, destinationPosition Point,
	path []int8,
) int {
	start := point8{x: uint8(startPosition.X), y: uint8(startPosition.Y)}
	dest := point8{x: uint8(destinationPosition.X), y: uint8(destinationPosition.Y)}
	packedDest := packPoint(dest)

	initialHeuristicCost := getHeuristicCost(start, dest)
	if initialHeuristicCost > PathDiagonalStepCost*int32(len(path)) {
		return 0
	}

	var frontierHeap [maxPathNodes]frontierNode
	frontier := frontierHeap[:0]
	explored := exploredNodes{}
	f := int16(initialHeuristicCost)
	frontier = append(frontier, frontierNode{position: packPoint(start), f: f, h: f})
	explored.emplace(packPoint(start), 0, 0)

	for len(frontier) != 0 {
		cur := frontier[0]
		if cur.position == packedDest {
			return reconstructPath(&explored, cur.position, path)
		}

		frontier = popFrontier(frontier)
		curG := explored.find(cur.position).g
		if curG >= PathDiagonalStepCost*int32(len(path)) {
			continue
		}
		if curG+int32(cur.h) > int32(cur.f) {
			continue
		}

		curPos := unpackPoint(cur.position)
		maxCost := PathDiagonalStepCost * int32(len(path))
		curPacked := cur.position
		for _, d := range PathDirs {
			neighborPos, ok := stepToward(curPos, d)
			if !ok || curG >= maxCost {
				continue
			}
			neighborPacked := packPoint(neighborPos)
			neighborPoint := pointOf(neighborPacked)
			if posOk(neighborPoint) {
				if !canStep(pointOf(curPacked), neighborPoint) {
					continue
				}
			} else if neighborPos != dest {
				// Targeting a non-walkable node is allowed only for the destination.
				continue
			}
			g := curG + getDistance(curPos, neighborPos)
			it := explored.find(neighborPacked)
			if applyGImprovement(&explored, it, neighborPacked, curPacked, g) {
				frontier = enqueueNeighbor(frontier, neighborPacked, g, dest, curPos, neighborPos)
			}
		}
	}
	return 0
}

// popFrontier (path.cpp:234-238) removes the root.
func popFrontier(frontier []frontierNode) []frontierNode {
	frontier[0] = frontier[len(frontier)-1]
	frontier = frontier[:len(frontier)-1]
	frontierSiftDown(frontier, 0)
	return frontier
}

// stepToward (path.cpp:245-247) moves one tile in direction d, refusing to
// leave the uint8 grid.
func stepToward(curPos point8, d Point) (point8, bool) {
	if (curPos.x == 0 && d.X < 0) || (curPos.y == 0 && d.Y < 0) {
		return point8{}, false
	}
	return point8{x: curPos.x + uint8(d.X), y: curPos.y + uint8(d.Y)}, true
}

// applyGImprovement (path.cpp:257-270) records the neighbor's first visit or
// its g-improvement; reports whether the neighbor must be requeued.
func applyGImprovement(explored *exploredNodes, it *exploredEntry, neighborPacked, prev uint16, g int32) bool {
	if it == nil {
		if !explored.canInsert(neighborPacked) {
			return false
		}
		explored.emplace(neighborPacked, prev, g)
		return true
	}
	if it.g <= g {
		return false
	}
	it.prev = prev
	it.g = g
	return true
}

// enqueueNeighbor (path.cpp:271-278) queues an improved neighbor under its
// cached heuristic when the frontier has room.
func enqueueNeighbor(
	frontier []frontierNode,
	neighborPacked uint16,
	g int32,
	dest point8,
	curPos, neighborPos point8,
) []frontierNode {
	if len(frontier) >= maxPathNodes {
		return frontier
	}
	h := getHeuristicCost(neighborPos, dest)
	diag := isDiagonalStep(curPos, neighborPos)
	node := frontierNode{position: neighborPacked, f: int16(g + h), h: int16(h), diag: diag}
	frontier = append(frontier, node)
	frontierSiftUp(frontier, len(frontier)-1)
	return frontier
}

// The frontier is a binary min-heap mirroring the upstream std::push_heap /
// std::pop_heap usage (path.cpp:198-279) under the comparator of path.cpp:202-222:
// lowest f, then lower heuristic, then diagonal-step-first, then coordinate order.
func frontierLess(a, b frontierNode) bool {
	if a.f != b.f {
		return a.f < b.f
	}
	if a.h != b.h {
		return a.h < b.h
	}
	if a.diag != b.diag {
		return a.diag
	}
	return a.position < b.position
}

func frontierSiftUp(frontier []frontierNode, i int) {
	for i != 0 {
		parent := (i - 1) / 2
		if !frontierLess(frontier[i], frontier[parent]) {
			return
		}
		frontier[i], frontier[parent] = frontier[parent], frontier[i]
		i = parent
	}
}

func frontierSiftDown(frontier []frontierNode, i int) {
	for {
		child := frontierMinChild(frontier, i)
		if child == -1 {
			return
		}
		if !frontierLess(frontier[child], frontier[i]) {
			return
		}
		frontier[i], frontier[child] = frontier[child], frontier[i]
		i = child
	}
}

// frontierMinChild returns the index of the smaller of i's two children, or
// -1 when i is a leaf.
func frontierMinChild(frontier []frontierNode, i int) int {
	child := 2*i + 1
	if child >= len(frontier) {
		return -1
	}
	if child+1 < len(frontier) && frontierLess(frontier[child+1], frontier[child]) {
		child++
	}
	return child
}

// testPathGetHeuristicCost (path.cpp:296-300) — upstream testing hook.
func testPathGetHeuristicCost(start, dest Point) int {
	start8 := point8{x: uint8(start.X), y: uint8(start.Y)}
	dest8 := point8{x: uint8(dest.X), y: uint8(dest.Y)}
	return int(getHeuristicCost(start8, dest8))
}

// findClosestValidPosition (path.cpp:285-294)
func findClosestValidPosition(
	posOk func(Point) bool,
	startingPosition Point,
	minimumRadius, maximumRadius uint,
) (Point, bool) {
	candidate, found := crawl(minimumRadius, maximumRadius, func(displacement Displacement) (Displacement, bool) {
		candidatePosition := startingPosition.Plus(Point{X: displacement.DeltaX, Y: displacement.DeltaY})
		if posOk(candidatePosition) {
			return displacement, true
		}
		return Displacement{}, false
	})
	if !found {
		return Point{}, false
	}
	return Point{X: startingPosition.X + candidate.DeltaX, Y: startingPosition.Y + candidate.DeltaY}, true
}
