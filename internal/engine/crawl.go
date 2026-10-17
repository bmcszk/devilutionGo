package engine

// Port of diasurgical/DevilutionX Source/crawl.cpp (1:1 semantics).

// Displacement (displacement.hpp:20) — signed offset between two tiles.
type Displacement struct{ DeltaX, DeltaY int }

// DoCrawl (crawl.cpp:16-49) visits every ring from minRadius through
// maxRadius; returning false from visit stops the iteration. Ring order and
// the corner inset are part of the 1:1 contract.
func DoCrawl(minRadius, maxRadius uint, visit func(Displacement) bool) bool {
	for r := int(minRadius); r <= int(maxRadius); r++ {
		if !visit(Displacement{DeltaX: 0, DeltaY: r}) {
			return false
		}
		if r == 0 {
			continue
		}
		if !visit(Displacement{DeltaX: 0, DeltaY: -r}) {
			return false
		}
		for x := 1; x < r; x++ {
			if !visit(Displacement{DeltaX: -x, DeltaY: r}) ||
				!visit(Displacement{DeltaX: x, DeltaY: r}) ||
				!visit(Displacement{DeltaX: -x, DeltaY: -r}) ||
				!visit(Displacement{DeltaX: x, DeltaY: -r}) {
				return false
			}
		}
		if r > 1 {
			d := r - 1
			if !visit(Displacement{DeltaX: -d, DeltaY: d}) ||
				!visit(Displacement{DeltaX: d, DeltaY: d}) ||
				!visit(Displacement{DeltaX: -d, DeltaY: -d}) ||
				!visit(Displacement{DeltaX: d, DeltaY: -d}) {
				return false
			}
		}
		if !visit(Displacement{DeltaX: -r, DeltaY: 0}) ||
			!visit(Displacement{DeltaX: r, DeltaY: 0}) {
			return false
		}
		for y := 1; y < r; y++ {
			if !visit(Displacement{DeltaX: -r, DeltaY: y}) ||
				!visit(Displacement{DeltaX: r, DeltaY: y}) ||
				!visit(Displacement{DeltaX: -r, DeltaY: -y}) ||
				!visit(Displacement{DeltaX: r, DeltaY: -y}) {
				return false
			}
		}
	}
	return true
}

// Crawl (crawl.hpp:49-61) — stops at the first truthy visit result and
// returns it; found is false when every candidate was rejected.
func Crawl(minRadius, maxRadius uint, visit func(Displacement) (Displacement, bool)) (Displacement, bool) {
	var result Displacement
	found := false
	DoCrawl(minRadius, maxRadius, func(displacement Displacement) bool {
		result, found = visit(displacement)
		return !found
	})
	return result, found
}
