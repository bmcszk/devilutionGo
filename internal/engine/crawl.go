package engine

// Port of diasurgical/DevilutionX Source/crawl.cpp (1:1 semantics).

// Displacement (displacement.hpp:20) — signed offset between two tiles.
type Displacement struct{ DeltaX, DeltaY int }

// DoCrawl (crawl.cpp:16-49) visits every ring from minRadius through
// maxRadius; returning false from visit stops the iteration.
func DoCrawl(minRadius, maxRadius uint, visit func(Displacement) bool) bool {
	for r := int(minRadius); r <= int(maxRadius); r++ {
		if !crawlRing(r, visit) {
			return false
		}
	}
	return true
}

// crawlRing (crawl.cpp:19-42) walks one ring.
func crawlRing(r int, visit func(Displacement) bool) bool {
	if !visit(Displacement{0, r}) {
		return false
	}
	if r == 0 {
		return true
	}
	if !visit(Displacement{0, -r}) || !crawlHorizontalEdges(r, visit) {
		return false
	}
	if r > 1 && !crawlInsetCorners(r, visit) {
		return false
	}
	if !visit(Displacement{-r, 0}) || !visit(Displacement{r, 0}) {
		return false
	}
	return crawlVerticalEdges(r, visit)
}

// crawlHorizontalEdges (crawl.cpp:22-27) — ±x for x in 1..r-1 on the top and
// bottom rows.
func crawlHorizontalEdges(r int, visit func(Displacement) bool) bool {
	for x := 1; x < r; x++ {
		if !visit4(visit, Displacement{-x, r}, Displacement{x, r}, Displacement{-x, -r}, Displacement{x, -r}) {
			return false
		}
	}
	return true
}

// crawlVerticalEdges (crawl.cpp:37-42) — ±y for y in 1..r-1 on the left and
// right columns.
func crawlVerticalEdges(r int, visit func(Displacement) bool) bool {
	for y := 1; y < r; y++ {
		if !visit4(visit, Displacement{-r, y}, Displacement{r, y}, Displacement{-r, -y}, Displacement{r, -y}) {
			return false
		}
	}
	return true
}

// crawlInsetCorners (crawl.cpp:28-34) — the (r-1)-inset diagonal corners,
// present only from radius 2 up.
func crawlInsetCorners(r int, visit func(Displacement) bool) bool {
	d := r - 1
	return visit4(visit, Displacement{-d, d}, Displacement{d, d}, Displacement{-d, -d}, Displacement{d, -d})
}

// visit4 (crawl.cpp:23-26 pattern) — short-circuits one 4-symmetry group.
func visit4(visit func(Displacement) bool, a, b, c, d Displacement) bool {
	return visit(a) && visit(b) && visit(c) && visit(d)
}

// Crawl (crawl.hpp:49-61) — stops at the first truthy visit result and
// returns it.
func Crawl(minRadius, maxRadius uint, visit func(Displacement) (Displacement, bool)) (Displacement, bool) {
	var result Displacement
	found := false
	DoCrawl(minRadius, maxRadius, func(displacement Displacement) bool {
		result, found = visit(displacement)
		return !found
	})
	return result, found
}
