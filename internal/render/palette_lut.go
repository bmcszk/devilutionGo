// Package render holds the software palette-lookup core — the 256x256 blend
// look-up table, the 16 light-shade palette variants and the index-buffer blit
// flavours — and the Ebitengine pass that presents those indices with the same
// lookup on the GPU.
//
// The software core is the parity oracle: Presenter's frames are compared byte
// for byte against Surface.ExpandSprite and Surface.ExpandTile, so every GPU
// behaviour has a backend-independent definition to check against.
package render

import "github.com/bmcszk/devilutionGo/internal/assets"

const (
	// paletteSize is the number of entries in a Diablo palette (palette.h:23).
	paletteSize = 256
	// numShades is the number of light tables (lighting_defs.hpp:17).
	numShades = 16
	// numComponents is the length of an RGB triple.
	numComponents = 3
	// noSkip is GenerateBlendedLookupTable's default skip bound
	// (palette_blending.hpp:23): a skipFrom above skipTo skips nothing.
	noSkip = -1
	// fullyLitShade is the light level of an unlit draw: shade 0 is the brightest
	// table (dLight 0 = bright, light_tables.cpp:19-70).
	fullyLitShade = 0
	// kdTreeDepth is the palette k-d tree depth (palette_kd_tree.hpp:34).
	kdTreeDepth = 5
	// kdNumLeaves is the leaf count of that tree.
	kdNumLeaves = 1 << kdTreeDepth
	// rotateForward moves the element at from to to (CycleColors, palette.cpp:54).
	rotateForward rotation = iota
	// rotateBackward moves the element at to to from (CycleColorsReverse,
	// palette.cpp:74).
	rotateBackward
	// keyZero turns a drawn index 0 into a transparent pixel, the sprite keying
	// rule (docs/research/dg-graphics.md G4).
	keyZero keying = iota
	// keepZero keeps a drawn index 0 as opaque black, the tile rule (same trap).
	keepZero
)

// BlendLUT maps the source and destination index of a blended draw to the palette
// index whose colour best matches their 50% blend (paletteTransparencyLookup,
// palette_blending.hpp:16).
type BlendLUT [paletteSize][paletteSize]uint8

// RGB is one palette colour in the order the .pal files store it.
type RGB [numComponents]uint8

// BlendRow is one row of the blend LUT: the lookup indexed by one source colour.
type BlendRow [paletteSize]uint8

// Variants holds the 16 light-shade palette rows: Variants[shade][index] is the
// colour the engine draws palette index index with at that light level.
//
// This is the byte-exact software definition of the 16 palette rows the palette
// pass consumes (docs/research/dg-graphics.md G4/G5, light_tables.cpp:19-70).
type Variants [numShades][paletteSize]RGB

// rotation is the direction a colour range shifts by one.
type rotation int

// keying says whether a drawn index 0 is a hole or a colour.
type keying int

// kdValue is one palette colour with the index it came from
// (palette_kd_tree.hpp:118).
type kdValue struct {
	color RGB
	index uint8
}

// kdLeaf is a leaf's slice of the value array, with the inclusive bounds upstream
// uses; begin > end marks an empty leaf (palette_kd_tree.hpp:74-78).
type kdLeaf struct {
	begin uint8
	end   uint8
}

// paletteKdTree is the 5-level k-d tree GenerateBlendedLookupTable searches for the
// nearest palette colour. The split order, the median rounding and the
// first-found tie-break are all byte-visible in the LUT, so they are ported
// verbatim (palette_kd_tree.hpp, palette_kd_tree.cpp).
type paletteKdTree struct {
	pivots [kdTreeDepth][kdNumLeaves]uint8
	values [paletteSize]kdValue
	leaves [kdNumLeaves]kdLeaf
}

// BuildBlendingLUT builds the 50%-blend look-up table for a palette
// (GenerateBlendedLookupTable, palette_blending.cpp:44-60).
//
// skipFrom and skipTo exclude palette indices from the search, so a blend never
// resolves into the range; pass noSkip for both to blend the whole palette, the
// values LoadPaletteAndInitBlending uses for town, cathedral, catacombs and hell
// (palette.cpp:188-200).
func BuildBlendingLUT(pal assets.Palette, skipFrom, skipTo int) BlendLUT {
	tree := buildPaletteKdTree(pal, skipFrom, skipTo)

	var lut BlendLUT
	for i := range paletteSize {
		writeRow(&lut, tree, pal, i)
	}

	return lut
}

// buildPaletteKdTree indexes the palette colours the blend search may return.
func buildPaletteKdTree(pal assets.Palette, skipFrom, skipTo int) paletteKdTree {
	skipped := func(index int) bool { return index >= skipFrom && index <= skipTo }

	var tree paletteKdTree
	tree.pivots = buildPivots(pal, skipped)
	tree.fillLeaves(pal, skipped)

	return tree
}

// writeRow fills LUT row i: the diagonal is the identity, the lower triangle is the
// mirror of the upper one, and each upper entry is a k-d tree search
// (palette_blending.cpp:46-57).
func writeRow(lut *BlendLUT, tree paletteKdTree, pal assets.Palette, i int) {
	lut[i][i] = uint8(i) //nolint:gosec // i < paletteSize, so the byte cast is exact.

	for j := range i {
		lut[i][j] = lut[j][i]
	}

	for j := i + 1; j < paletteSize; j++ {
		lut[i][j] = tree.nearest(blendColors(pal[i], pal[j]))
	}
}

// blendColors averages two colours with integer halving (BlendColors,
// palette_blending.cpp:32-39).
func blendColors(a, b RGB) RGB {
	return RGB{
		uint8((int(a[0]) + int(b[0])) / 2), //nolint:gosec // both inputs are uint8, so the sum halves to 0..255.
		uint8((int(a[1]) + int(b[1])) / 2), //nolint:gosec // both inputs are uint8, so the sum halves to 0..255.
		uint8((int(a[2]) + int(b[2])) / 2), //nolint:gosec // both inputs are uint8, so the sum halves to 0..255.
	}
}

// buildPivots sets one pivot per tree node, level by level, each to the median of
// the component the level splits on (PopulatePivots, palette_kd_tree.cpp:99-166).
func buildPivots(pal assets.Palette, skipped func(int) bool) [kdTreeDepth][kdNumLeaves]uint8 {
	var pivots [kdTreeDepth][kdNumLeaves]uint8

	for depth := range kdTreeDepth {
		counts, totals := countComponents(pal, pivots, depth, skipped)
		for pos := range 1 << depth {
			pivots[depth][pos] = medianPivot(counts[pos][:], totals[pos])
		}
	}

	return pivots
}

// countComponents tallies the split component of every searchable colour per node
// at depth, descending with the pivots of the levels above it.
func countComponents(
	pal assets.Palette,
	pivots [kdTreeDepth][kdNumLeaves]uint8,
	depth int,
	skipped func(int) bool,
) ([kdNumLeaves][paletteSize]uint16, [kdNumLeaves]uint16) {
	var (
		counts [kdNumLeaves][paletteSize]uint16
		totals [kdNumLeaves]uint16
	)
	coord := depth % numComponents

	for index := range paletteSize {
		if skipped(index) {
			continue
		}

		pos := descendTo(pivots, pal[index], depth)
		counts[pos][pal[index][coord]]++
		totals[pos]++
	}

	return counts, totals
}

// descendTo returns the node position at depth that a colour falls under.
func descendTo(pivots [kdTreeDepth][kdNumLeaves]uint8, color RGB, depth int) int {
	pos := 0
	for level := range depth {
		pos = 2*pos + childSide(color[level%numComponents], pivots[level][pos])
	}

	return pos
}

// childSide returns 1 for the right child and 0 for the left one: a colour equal to
// the pivot goes right (palette_kd_tree.cpp:46).
func childSide(component, pivot uint8) int {
	if component < pivot {
		return 0
	}

	return 1
}

// medianPivot picks a node's split value (GetMedian, palette_kd_tree.cpp:69-88): the
// first component whose cumulative count reaches half the node's colours, plus one,
// so an all-equal node sends every colour to the same side.
func medianPivot(counts []uint16, numValues uint16) uint8 {
	target := (numValues + 1) / 2
	partial := uint16(0)
	index := 0

	for ; partial < target && partial != numValues; index++ {
		partial += counts[index]
	}

	return uint8(index) //nolint:gosec // index stops at paletteSize, so the byte cast is exact.
}

// fillLeaves stores every searchable colour in its leaf, in palette-index order,
// which is the order the nearest-neighbour search visits them in
// (PaletteKdTree::PaletteKdTree, palette_kd_tree.cpp:140-171).
func (t *paletteKdTree) fillLeaves(pal assets.Palette, skipped func(int) bool) {
	leafValues := t.groupIndexesByLeaf(pal, skipped)

	total := 0
	for leaf := range kdNumLeaves {
		total += t.storeLeaf(pal, leaf, leafValues[leaf], total)
	}
}

// groupIndexesByLeaf buckets the searchable palette indexes into leaf slots.
func (t *paletteKdTree) groupIndexesByLeaf(
	pal assets.Palette,
	skipped func(int) bool,
) [kdNumLeaves][]uint8 {
	var leafValues [kdNumLeaves][]uint8

	for index := range paletteSize {
		if skipped(index) {
			continue
		}

		leaf := t.leafIndexFor(pal[index])
		leafValues[leaf] = append(leafValues[leaf], uint8(index)) //nolint:gosec // index < paletteSize.
	}

	return leafValues
}

// storeLeaf writes one leaf's colours into the value array and returns how many it
// added (palette_kd_tree.cpp:152-171).
func (t *paletteKdTree) storeLeaf(pal assets.Palette, leaf int, indexes []uint8, total int) int {
	if len(indexes) == 0 {
		t.leaves[leaf] = kdLeaf{begin: 1, end: 0}

		return 0
	}

	t.leaves[leaf] = kdLeaf{
		begin: uint8(total),                    //nolint:gosec // total <= paletteSize.
		end:   uint8(total + len(indexes) - 1), //nolint:gosec // total+size-1 <= paletteSize-1.
	}

	for offset, index := range indexes {
		t.values[total+offset] = kdValue{color: RGB(pal[index]), index: index}
	}

	return len(indexes)
}

// leafIndexFor returns the leaf slot a colour descends to, in the same order the
// search visits the tree so a leaf's stored range is the range it searches
// (LeafIndexForColor, palette_kd_tree.cpp:36-44).
func (t *paletteKdTree) leafIndexFor(color RGB) int {
	pos := 0

	for level := range kdTreeDepth {
		pos = 2*pos + childSide(color[level%numComponents], t.pivots[level][pos])
	}

	return pos
}

// nearest returns the index of the palette colour closest to target
// (PaletteKdTree::findNearestNeighbor, palette_kd_tree.hpp:103-111).
func (t *paletteKdTree) nearest(target RGB) uint8 {
	best, bestDiff := uint8(0), ^uint32(0)
	t.visit(0, 0, target, &best, &bestDiff)

	return best
}

// visit searches the near half-space first and the far one only when the split
// plane is closer than the best match so far (findNearestNeighborVisit,
// palette_kd_tree.hpp:141-158).
func (t *paletteKdTree) visit(depth, pos int, target RGB, best *uint8, bestDiff *uint32) {
	if depth == kdTreeDepth {
		t.visitLeaf(pos, target, best, bestDiff)

		return
	}

	component := int(target[depth%numComponents])
	pivot := int(t.pivots[depth][pos])
	near, far := 0, 1
	if component >= pivot {
		near, far = 1, 0
	}

	t.visit(depth+1, 2*pos+near, target, best, bestDiff)

	delta := pivot - component
	if uint32(delta*delta) < *bestDiff { //nolint:gosec // |delta| <= 255, so the square cannot overflow.
		t.visit(depth+1, 2*pos+far, target, best, bestDiff)
	}
}

// visitLeaf scans a leaf's colours, keeping the first of equally distant ones
// (findNearestNeighborVisit, palette_kd_tree.hpp:160-179).
func (t *paletteKdTree) visitLeaf(pos int, target RGB, best *uint8, bestDiff *uint32) {
	leaf := t.leaves[pos]
	if leaf.begin > leaf.end {
		return
	}

	for offset := int(leaf.begin); offset <= int(leaf.end); offset++ {
		value := t.values[offset]
		if diff := colorDistance(value.color, target); diff < *bestDiff {
			*best = value.index
			*bestDiff = diff
		}
	}
}

// colorDistance is the squared euclidean distance between two colours
// (getColorDistance, palette_kd_tree.hpp:121-127).
func colorDistance(a, b RGB) uint32 {
	var sum uint32
	for component := range numComponents {
		delta := int(a[component]) - int(b[component])
		sum += uint32(delta * delta) //nolint:gosec // |delta| <= 255, so the square cannot overflow.
	}

	return sum
}
