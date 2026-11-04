package levels

// This file holds the catacombs generator stages GenerateLevel runs between
// CreateDungeon and the pdungeon copy (levels/drlg_l2.cpp:2679-2829).

// catacombTransOrigin is where FixTransparency's doubled 2x2 cursor starts
// (levels/drlg_l2.cpp:2515-2516).
const catacombTransOrigin = 16

// The doubled-transparency wall ids FixTransparency keys on
// (levels/drlg_l2.cpp:2519-2540): the right and bottom walls always copy their
// region id onto their transparent half, the guarded variants only do so next
// to the other half of the pair, and the corner copies both halves plus the
// diagonal.
const (
	catacombTransRightWall     Tile = 10
	catacombTransBottomWall    Tile = 11
	catacombTransRightGuarded  Tile = 14
	catacombTransBottomGuarded Tile = 15
	catacombTransCorner        Tile = 16
)

// l2DirtRules is FixDirtTiles' six closing rules in upstream order
// (levels/drlg_l2.cpp:2548-2567): when the cell holds match and the neighbour
// one step away is not keep, the cell closes into write. The written tiles are
// never a match, so applying the rules in sequence equals applying them to the
// original cell.
var l2DirtRules = [...]struct {
	match Tile
	keep  Tile
	write Tile
	step  Point
}{
	{match: 13, keep: 11, write: 146, step: Point{X: 1}},
	{match: 11, keep: 11, write: 144, step: Point{X: 1}},
	{match: 15, keep: 11, write: 148, step: Point{X: 1}},
	{match: 10, keep: 10, write: 143, step: Point{Y: 1}},
	{match: 13, keep: 10, write: 146, step: Point{Y: 1}},
	{match: 14, keep: 15, write: 147, step: Point{Y: 1}},
}

// l2TilePatternRules is FixTilesPatterns' six neighbour rules in upstream
// order (levels/drlg_l2.cpp:2056-2075). Each fires only for the cell's own
// tile, and no rule writes the cell it tests, so the order is not load bearing.
var l2TilePatternRules = [...]struct {
	match Tile
	want  Tile
	write Tile
	step  Point
}{
	{match: 1, want: 3, write: 1, step: Point{Y: 1}},
	{match: 3, want: 1, write: 3, step: Point{Y: 1}},
	{match: 3, want: 7, write: 3, step: Point{X: 1}},
	{match: 2, want: 3, write: 2, step: Point{X: 1}},
	{match: 11, want: 14, write: 16, step: Point{X: 1}},
	{match: 15, want: 1, write: 8, step: Point{Y: 1}},
}

// createCatacombs is CreateDungeon: it lays out the rooms and halls and turns
// the ASCII layout into the tile layer (levels/drlg_l2.cpp:2445-2508).
func (d *Dungeon) createCatacombs(quests Quests) bool {
	return newCatacombs(d).build(quests)
}

// fixCatacombTilesPatterns is FixTilesPatterns (levels/drlg_l2.cpp:2053-2077).
func (d *Dungeon) fixCatacombTilesPatterns() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.fixCatacombTilesAt(i, j)
		}
	}
}

// fixCatacombTilesAt applies FixTilesPatterns' six neighbour rules at one cell
// (levels/drlg_l2.cpp:2056-2075). The rules read the neighbours to the right
// and below, which walk into pdungeon on the last row and column upstream; the
// writes those rows ask for land in pdungeon, which the final copy overwrites.
func (d *Dungeon) fixCatacombTilesAt(i, j int) {
	for _, rule := range l2TilePatternRules {
		ti, tj := i+rule.step.X, j+rule.step.Y
		if d.Tiles[i][j] != rule.match || d.flatTile(ti, tj) != rule.want {
			continue
		}

		d.Tiles[ti][tj] = rule.write
	}
}

// fixCatacombTransparency copies the region id onto the doubled tiles the wall
// tiles 10, 11, 14, 15 and 16 only half cover (levels/drlg_l2.cpp:2510-2543).
func (d *Dungeon) fixCatacombTransparency() {
	yy := catacombTransOrigin
	for j := range DMAXY {
		xx := catacombTransOrigin
		for i := range DMAXX {
			d.fixCatacombTransparencyAt(i, j, xx, yy)

			xx += 2
		}

		yy += 2
	}
}

// fixCatacombTransparencyAt is the per-tile half of FixTransparency: the doubled
// tiles 10/11/16 always copy their region id, while 14 and 15 only do so when
// the half tile they sit against is the matching wall
// (levels/drlg_l2.cpp:2519-2540).
func (d *Dungeon) fixCatacombTransparencyAt(i, j, xx, yy int) {
	switch d.Tiles[i][j] {
	case catacombTransRightGuarded:
		if d.flatTile(i, j-1) == catacombTransRightWall {
			d.copyTransRight(xx, yy)
		}
	case catacombTransBottomGuarded:
		if d.flatTile(i+1, j) == catacombTransBottomWall {
			d.copyTransDown(xx, yy)
		}
	case catacombTransRightWall:
		d.copyTransRight(xx, yy)
	case catacombTransBottomWall:
		d.copyTransDown(xx, yy)
	case catacombTransCorner:
		d.copyTransRight(xx, yy)
		d.copyTransDown(xx, yy)
	default:
		// every other tile covers its whole 2x2 cell
	}
}

// copyTransRight copies one region id onto the doubled tile to the right.
func (d *Dungeon) copyTransRight(xx, yy int) {
	d.CopyTrans(xx, yy, xx+1, yy)
	d.CopyTrans(xx, yy, xx+1, yy+1)
}

// copyTransDown copies one region id onto the doubled tile below.
func (d *Dungeon) copyTransDown(xx, yy int) {
	d.CopyTrans(xx, yy, xx, yy+1)
	d.CopyTrans(xx, yy, xx+1, yy+1)
}

// fixCatacombDirtTiles closes every dirt tile whose neighbour does not
// continue its face (levels/drlg_l2.cpp:2545-2569).
func (d *Dungeon) fixCatacombDirtTiles() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.fixCatacombDirtAt(i, j)
		}
	}
}

// fixCatacombDirtAt applies FixDirtTiles' six rules at one cell
// (levels/drlg_l2.cpp:2548-2567).
func (d *Dungeon) fixCatacombDirtAt(i, j int) {
	for _, rule := range l2DirtRules {
		if d.Tiles[i][j] != rule.match || d.flatTile(i+rule.step.X, j+rule.step.Y) == rule.keep {
			continue
		}

		d.Tiles[i][j] = rule.write
	}
}

// placeCatacombSetPiece drops the quest set piece into the reserved room
// (levels/drlg_l2.cpp:1638-1655, InitSetPiece).
func (d *Dungeon) placeCatacombSetPiece(opts Options) error {
	path := catacombSetPiecePath(opts.Quests)
	if path == "" {
		return nil
	}

	data, err := dungeonSetPiece(opts, path)
	if err != nil {
		return err
	}

	d.PlaceDunTiles(data, d.SetPieceRoom.Position, catacombFloor)
	d.SetPiece = Rectangle{Position: d.SetPieceRoom.Position, Size: GetDunSize(data)}

	return nil
}

// catacombSetPiecePath is the .dun the catacombs quest set piece loads, in
// upstream's quest order (levels/drlg_l2.cpp:1641-1649).
func catacombSetPiecePath(quests Quests) string {
	switch {
	case quests.Blind:
		return "levels/l2data/blind1.dun"
	case quests.Blood:
		return "levels/l2data/blood1.dun"
	case quests.Schamb:
		return "levels/l2data/bonestr2.dun"
	default:
		return ""
	}
}

// fixCatacombLockout seals the wall tiles that would let a player walk out of
// the level (levels/drlg_l2.cpp:2571-2633).
func (d *Dungeon) fixCatacombLockout() {
	d.closeLockoutEdges()
	d.sealLockoutRows()
	d.sealLockoutColumns()
}

// closeLockoutEdges is FixLockout's first pass: door candidates without floor
// on the in-facing side fall back to the plain wall or floor tile they border
// (levels/drlg_l2.cpp:2573-2582).
func (d *Dungeon) closeLockoutEdges() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.closeLockoutEdgeAt(i, j)
		}
	}
}

func (d *Dungeon) closeLockoutEdgeAt(i, j int) {
	if d.Tiles[i][j] == 4 && d.flatTile(i-1, j) != 3 {
		d.Tiles[i][j] = 1
	}
	if d.Tiles[i][j] == 5 && d.flatTile(i, j-1) != 3 {
		d.Tiles[i][j] = 2
	}
}

// sealLockoutRows is FixLockout's horizontal corridor pass
// (levels/drlg_l2.cpp:2583-2609).
func (d *Dungeon) sealLockoutRows() {
	for j := 1; j < DMAXY-1; j++ {
		for i := 1; i < DMAXX-1; i++ {
			i = d.sealLockoutRow(i, j)
		}
	}
}

// sealLockoutColumns is FixLockout's vertical corridor pass, which walks
// dungeon[j][i] instead (levels/drlg_l2.cpp:2610-2631).
func (d *Dungeon) sealLockoutColumns() {
	for j := 1; j < DMAXX-1; j++ {
		for i := 1; i < DMAXY-1; i++ {
			i = d.sealLockoutColumn(j, i)
		}
	}
}

// sealLockoutRow walks one horizontal corridor of door candidates and turns the
// last tile into a door when the run has none
// (levels/drlg_l2.cpp:2584-2609). It returns the index the scan stops on.
func (d *Dungeon) sealLockoutRow(i, j int) int {
	if !d.lockoutRowCandidate(i, j) {
		return i
	}

	stop, doorOK := d.scanLockoutRow(i, j)
	if !doorOK && !d.Protected.Test(stop-1, j) {
		d.Tiles[stop-1][j] = 5
	}

	return stop
}

// lockoutRowCandidate is FixLockout's horizontal guard: the cell is outside the
// protected mask and is a door candidate with floor on both sides
// (levels/drlg_l2.cpp:2579-2582).
func (d *Dungeon) lockoutRowCandidate(i, j int) bool {
	if d.Protected.Test(i, j) {
		return false
	}

	if d.Tiles[i][j] != 2 && d.Tiles[i][j] != 5 {
		return false
	}

	return d.flatTile(i, j-1) == 3 && d.flatTile(i, j+1) == 3
}

// scanLockoutRow walks the horizontal run of door candidates from i and returns
// the index it stopped on plus whether the run held a vertical door
// (levels/drlg_l2.cpp:2586-2601).
func (d *Dungeon) scanLockoutRow(i, j int) (stop int, doorOK bool) {
	for d.Tiles[i][j] == 2 || d.Tiles[i][j] == 5 {
		if d.flatTile(i, j-1) != 3 || d.flatTile(i, j+1) != 3 {
			break
		}
		if d.Tiles[i][j] == 5 {
			doorOK = true
		}

		i++
	}

	return i, doorOK
}

// sealLockoutColumn is the transposed half of FixLockout: the same scan over
// dungeon[j][i] (levels/drlg_l2.cpp:2610-2631).
func (d *Dungeon) sealLockoutColumn(j, i int) int {
	if !d.lockoutColumnCandidate(j, i) {
		return i
	}

	stop, doorOK := d.scanLockoutColumn(j, i)
	if !doorOK && !d.Protected.Test(j, stop-1) {
		d.Tiles[j][stop-1] = 4
	}

	return stop
}

// lockoutColumnCandidate is the transposed guard of lockoutRowCandidate: door
// candidates 1 and 4 with floor above and below (levels/drlg_l2.cpp:2616-2619).
func (d *Dungeon) lockoutColumnCandidate(j, i int) bool {
	if d.Protected.Test(j, i) {
		return false
	}

	if d.Tiles[j][i] != 1 && d.Tiles[j][i] != 4 {
		return false
	}

	return d.flatTile(j-1, i) == 3 && d.flatTile(j+1, i) == 3
}

// scanLockoutColumn is the transposed half of scanLockoutRow
// (levels/drlg_l2.cpp:2622-2630).
func (d *Dungeon) scanLockoutColumn(j, i int) (stop int, doorOK bool) {
	for d.Tiles[j][i] == 1 || d.Tiles[j][i] == 4 {
		if d.flatTile(j-1, i) != 3 || d.flatTile(j+1, i) != 3 {
			break
		}
		if d.Tiles[j][i] == 4 {
			doorOK = true
		}

		i++
	}

	return i, doorOK
}

// fixCatacombDoors turns the horizontal and vertical door candidates into the
// door tiles their floor neighbour asks for (levels/drlg_l2.cpp:2635-2647).
func (d *Dungeon) fixCatacombDoors() {
	for j := 1; j < DMAXY; j++ {
		for i := 1; i < DMAXX; i++ {
			d.fixCatacombDoorAt(i, j)
		}
	}
}

// fixCatacombDoorAt is FixDoors' two rules: a horizontal candidate hard against
// floor on its left becomes a left-closing door, a vertical one against floor
// above becomes a top-closing door (levels/drlg_l2.cpp:2638-2645).
func (d *Dungeon) fixCatacombDoorAt(i, j int) {
	if d.Tiles[i][j] == 4 && d.Tiles[i][j-1] == 3 {
		d.Tiles[i][j] = 7
	}
	if d.Tiles[i][j] == 5 && d.Tiles[i-1][j] == 3 {
		d.Tiles[i][j] = 9
	}
}

// placeCatacombStairs places the up, down and (on level 5) town-warp stairs and
// points the camera at them, reporting whether the level is usable
// (levels/drlg_l2.cpp:2649-2677).
func (d *Dungeon) placeCatacombStairs(entry Entry) bool {
	up, ok := d.PlaceMiniSet(l2StairsUp, 199, false)
	if !ok {
		return false
	}
	if entry == EntryMain {
		d.ViewPosition = up.MegaToWorld().Add(Point{X: 5, Y: 4})
	}

	down, ok := d.PlaceMiniSet(l2StairsDown, 199, false)
	if !ok {
		return false
	}
	if entry == EntryPrev {
		d.ViewPosition = down.MegaToWorld().Add(Point{X: 4, Y: 6})
	}

	if d.CurrLevel != 5 {
		return true
	}

	warp, ok := d.PlaceMiniSet(l2WarpStairs, 199, false)
	if !ok {
		return false
	}
	if entry == EntryTwarPDn {
		d.ViewPosition = warp.MegaToWorld().Add(Point{X: 5, Y: 4})
	}

	return true
}

// drawCatacombQuests runs DRLG_CheckQuests for the catacombs quest branches
// that write tiles, all anchored on the set piece room's corner
// (levels/drlg_l2.cpp:2818, levels/drlg_quests.cpp:70-131).
func (d *Dungeon) drawCatacombQuests(opts Options) error {
	position := d.SetPieceRoom.Position

	if opts.Quests.Blind {
		d.Tiles[position.X][position.Y+1] = catacombBlindWall
		d.Tiles[position.X+10][position.Y+8] = catacombBlindWall
	}

	if opts.Quests.Blood {
		data, err := dungeonSetPiece(opts, "levels/l2data/blood2.dun")
		if err != nil {
			return err
		}

		d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}
		d.PlaceDunTiles(data, position, 0)
	}

	if opts.Quests.Schamb {
		data, err := dungeonSetPiece(opts, "levels/l2data/bonestr1.dun")
		if err != nil {
			return err
		}

		d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}
		d.PlaceDunTiles(data, position, catacombFloor)
	}

	return nil
}

// catacombSubstitution swaps floor and wall tiles for a random tile of the same
// base type, skipping the set piece room and tiles the same type already
// touches (levels/drlg_l2.cpp:2079-2117).
func (d *Dungeon) catacombSubstitution() {
	for y := range DMAXY {
		for x := range DMAXX {
			d.substituteCatacombTileIfPicked(x, y)
		}
	}
}

// substituteCatacombTileIfPicked is one Substitution draw: set-piece rooms are
// skipped and the rest only substitute on a 1-in-4 frequency
// (levels/drlg_l2.cpp:2087-2100).
func (d *Dungeon) substituteCatacombTileIfPicked(x, y int) {
	if d.SetPieceRoom.Contains(Point{X: x, Y: y}) {
		return
	}

	if !d.Rnd.FlipCoin(4) {
		return
	}

	d.substituteCatacombTile(x, y)
}

// substituteCatacombTile is Substitution's per-cell draw: it picks the draw
// count, resolves it to a tile id and writes it unless the 4x4 neighbourhood
// already holds that id (levels/drlg_l2.cpp:2087-2113).
func (d *Dungeon) substituteCatacombTile(x, y int) {
	base := l2BaseTypes[d.Tiles[x][y]]
	if base == 0 {
		return
	}

	rv := d.Rnd.GenerateRnd(16)
	i := -1
	for rv >= 0 {
		i++
		if i == len(l2BaseTypes) {
			i = 0
		}
		if base == l2BaseTypes[i] {
			rv--
		}
	}

	if !d.substitutedNear(x, y, Tile(i)) {
		d.Tiles[x][y] = Tile(i)
	}
}

// substitutedNear reports whether the 4x4 window around (x,y) already holds
// tile (levels/drlg_l2.cpp:2103-2111).
func (d *Dungeon) substitutedNear(x, y int, tile Tile) bool {
	for j := y - 2; j < y+2; j++ {
		for k := x - 2; k < x+2; k++ {
			if d.flatTile(k, j) == tile {
				return true
			}
		}
	}

	return false
}

// applyCatacombShadowPatterns stamps the two catacombs shadow patterns on every
// 2x2 base-type window (levels/drlg_l2.cpp:1569-1602).
func (d *Dungeon) applyCatacombShadowPatterns() {
	for y := 1; y < DMAXY; y++ {
		for x := 1; x < DMAXX; x++ {
			d.stampCatacombShadowAt(x, y)
		}
	}
}

// stampCatacombShadowAt applies every matching shadow pattern at one window.
// Unlike the cathedral pass it ignores the protected mask and has no break, so
// both patterns can write the same corner (levels/drlg_l2.cpp:1575-1600).
func (d *Dungeon) stampCatacombShadowAt(x, y int) {
	var window [2][2]Tile
	window[0][0] = l2ShadowBaseTypes[d.Tiles[x][y]]
	window[1][0] = l2ShadowBaseTypes[d.Tiles[x-1][y]]
	window[0][1] = l2ShadowBaseTypes[d.Tiles[x][y-1]]
	window[1][1] = l2ShadowBaseTypes[d.Tiles[x-1][y-1]]

	for _, shadow := range l2ShadowPatterns {
		if shadow.matches(window) {
			d.writeCatacombShadowAt(x, y, shadow)
		}
	}
}

// writeCatacombShadowAt writes the three non-zero corners of one matched
// pattern, leaving the zero ones alone (levels/drlg_l2.cpp:1590-1598).
func (d *Dungeon) writeCatacombShadowAt(x, y int, shadow shadowPattern) {
	if shadow.nv1 != 0 {
		d.Tiles[x-1][y-1] = shadow.nv1
	}
	if shadow.nv2 != 0 {
		d.Tiles[x][y-1] = shadow.nv2
	}
	if shadow.nv3 != 0 {
		d.Tiles[x-1][y] = shadow.nv3
	}
}

// placeCatacombMiniSetRandom is the catacombs copy of PlaceMiniSetRandom: it
// scans the level and places miniset wherever it fits, the tile is not already
// near, and the draw passes rndper (levels/drlg_l2.cpp:1604-1631).
func (d *Dungeon) placeCatacombMiniSetRandom(miniset Miniset, rndper int32) {
	sw := miniset.Size.Width
	sh := miniset.Size.Height

	for sy := range DMAXY - sh {
		for sx := range DMAXX - sw {
			d.placeCatacombMiniSetAt(miniset, Point{X: sx, Y: sy}, rndper)
		}
	}
}

// placeCatacombMiniSetAt decides one candidate position
// (levels/drlg_l2.cpp:1610-1630).
func (d *Dungeon) placeCatacombMiniSetAt(miniset Miniset, at Point, rndper int32) {
	if d.SetPieceRoom.Contains(at) {
		return
	}
	if !miniset.Matches(d, at, true) {
		return
	}
	if d.minisetReplacedNear(miniset, at) {
		return
	}
	if d.Rnd.GenerateRnd(100) >= rndper {
		return
	}

	miniset.Place(d, at, false)
}

// minisetReplacedNear reports whether the 3x3 miniset band around at already
// holds the replacement tile (levels/drlg_l2.cpp:1614-1625).
func (d *Dungeon) minisetReplacedNear(miniset Miniset, at Point) bool {
	sw := miniset.Size.Width
	sh := miniset.Size.Height

	for yy := max(at.Y-sh, 0); yy < min(at.Y+2*sh, DMAXY); yy++ {
		for xx := max(at.X-sw, 0); xx < min(at.X+2*sw, DMAXX); xx++ {
			if d.Tiles[xx][yy] == miniset.Replace[0][0] {
				return true
			}
		}
	}

	return false
}

// placeCatacombMiniSetRandom1x1 is the 1x1 case of the catacombs
// PlaceMiniSetRandom (levels/drlg_l2.cpp:1633-1636).
func (d *Dungeon) placeCatacombMiniSetRandom1x1(search, replace Tile, rndper int32) {
	d.placeCatacombMiniSetRandom(Miniset{
		Size:    Size{Width: 1, Height: 1},
		Search:  [6][6]Tile{{search}},
		Replace: [6][6]Tile{{replace}},
	}, rndper)
}
