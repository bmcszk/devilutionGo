package engine

// Displacement is a signed offset between two tiles.
type Displacement struct{ DeltaX, DeltaY int }

// doCrawl visits every ring from minRadius through maxRadius; returning false
// from visit stops the iteration early.
func doCrawl(minRadius, maxRadius uint, visit func(Displacement) bool) {
	for r := int(minRadius); r <= int(maxRadius); r++ {
		if !crawlRing(r, visit) {
			return
		}
	}
}

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

func crawlHorizontalEdges(r int, visit func(Displacement) bool) bool {
	for x := 1; x < r; x++ {
		if !visit4(visit, Displacement{-x, r}, Displacement{x, r}, Displacement{-x, -r}, Displacement{x, -r}) {
			return false
		}
	}
	return true
}

func crawlVerticalEdges(r int, visit func(Displacement) bool) bool {
	for y := 1; y < r; y++ {
		if !visit4(visit, Displacement{-r, y}, Displacement{r, y}, Displacement{-r, -y}, Displacement{r, -y}) {
			return false
		}
	}
	return true
}

func crawlInsetCorners(r int, visit func(Displacement) bool) bool {
	d := r - 1
	return visit4(visit, Displacement{-d, d}, Displacement{d, d}, Displacement{-d, -d}, Displacement{d, -d})
}

func visit4(visit func(Displacement) bool, a, b, c, d Displacement) bool {
	return visit(a) && visit(b) && visit(c) && visit(d)
}

// crawl returns the first displacement for which visit reports true.
func crawl(minRadius, maxRadius uint, visit func(Displacement) (Displacement, bool)) (Displacement, bool) {
	var result Displacement
	found := false
	doCrawl(minRadius, maxRadius, func(displacement Displacement) bool {
		result, found = visit(displacement)
		return !found
	})
	return result, found
}
