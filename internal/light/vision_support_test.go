package light_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// The trace scenario's player ids, in the order the fixture comments name them.
const (
	traceMyID       = 0
	traceFriendlyID = 1
	traceHostileID  = 2
	traceOffLevelID = 3
)

// automapCall is one SetAutomapView call: the tile and the explorer recorded.
type automapCall struct {
	position light.Point
	explorer light.Exploration
}

// visionState is the dungeon state a vision test scripts: the maps behind the
// VisionGrid accessors plus the counters that prove a pass wrote nothing.
type visionState struct {
	flags    map[light.Point]light.DungeonFlag
	trans    map[light.Point]int8
	passable func(position light.Point) bool
	inBounds func(position light.Point) bool
	views    []automapCall
	regions  map[int8]bool
	writes   int
	resets   int
}

// tracePlayer is the PlayerView the parity test scripts.
type tracePlayer struct {
	id          int
	myPlayer    bool
	plractive   bool
	onLevel     bool
	lvlChanging bool
	friendly    bool
}

// traceWorld is the parity scenario: the four players, their vision slots and
// the dungeon state of the committed fixture.
type traceWorld struct {
	state   *visionState
	list    *light.VisionList
	players [4]tracePlayer
}

// visionFrame is one parsed trace frame: the expected dFlags, explorer and
// TransList state plus the expected VisionActive flags.
type visionFrame struct {
	flags    [traceSide][traceSide]light.DungeonFlag
	explorer [traceSide][traceSide]int
	regions  []int
	active   [light.MaxVision]bool
}

// traceParser is the row-by-row decoder state of parseVisionTrace.
type traceParser struct {
	frames  map[string]visionFrame
	frame   string
	section string
	row     int
}

// rayTable is the parse state of the ray fixture: the 23 decoded rows, the
// RayLenAdj header and whether that header was seen.
//
// The decisive invariant: ray_len_adj is the LAST header before the rows, so
// its prefix match must be tried before the generic comment skip.
type rayTable struct {
	rays  [23][15]light.Offset
	adj   [23]uint8
	rows  int
	found bool
}

// addLine consumes one fixture line: the RayLenAdj header, a comment or a ray
// row, in that order.
func (t *rayTable) addLine(line string) error {
	if value, ok := strings.CutPrefix(line, "# ray_len_adj:"); ok {
		return t.setAdj(value)
	}
	if strings.HasPrefix(line, "#") {
		return nil
	}

	return t.setRow(line)
}

// newVisionState returns a 25x25 state of open floor with no flags set.
func newVisionState() *visionState {
	state := &visionState{
		flags:    map[light.Point]light.DungeonFlag{},
		trans:    map[light.Point]int8{},
		passable: func(light.Point) bool { return true },
		inBounds: func(position light.Point) bool {
			return position.X >= 0 && position.X < traceSide && position.Y >= 0 && position.Y < traceSide
		},
		regions: map[int8]bool{},
	}

	return state
}

// visionGrid wires the state into the boundary the vision passes use.
func visionGrid(state *visionState) light.VisionGrid {
	return light.VisionGrid{
		Flags: func(position light.Point) light.DungeonFlag { return state.flags[position] },
		SetFlags: func(position light.Point, flags light.DungeonFlag) {
			state.flags[position] = flags
			state.writes++
		},
		PassesLight: state.passable,
		InBounds:    state.inBounds,
		Trans:       func(position light.Point) int8 { return state.trans[position] },
		MarkTransparent: func(region int8) {
			state.regions[region] = true
		},
		ResetTransList: func() {
			state.regions = map[int8]bool{}
			state.resets++
		},
		SetAutomapView: func(position light.Point, explorer light.Exploration) {
			state.views = append(state.views, automapCall{position: position, explorer: explorer})
		},
	}
}

// withVisionBits marks every tile of the 25x25 grid Visible|Lit, the state a
// reveal leaves behind.
func (s *visionState) withVisionBits() {
	for x := range traceSide {
		for y := range traceSide {
			s.flags[light.Point{X: x, Y: y}] = light.FlagVisionBits
		}
	}
}

func (p tracePlayer) ID() int             { return p.id }
func (p tracePlayer) MyPlayer() bool      { return p.myPlayer }
func (p tracePlayer) Plractive() bool     { return p.plractive }
func (p tracePlayer) OnActiveLevel() bool { return p.onLevel }
func (p tracePlayer) LvlChanging() bool   { return p.lvlChanging }
func (p tracePlayer) FriendlyMode() bool  { return p.friendly }

// newTraceWorld builds the scenario the committed trace was produced from.
func newTraceWorld(t *testing.T) *traceWorld {
	t.Helper()

	state := newVisionState()
	state.passable = walledFloor()

	world := &traceWorld{state: state, list: light.NewVisionList()}
	world.players[traceMyID] = tracePlayer{id: traceMyID, myPlayer: true, plractive: true, onLevel: true}
	world.players[traceFriendlyID] = tracePlayer{id: traceFriendlyID, plractive: true, onLevel: true, friendly: true}
	world.players[traceHostileID] = tracePlayer{id: traceHostileID, plractive: true, onLevel: true}
	world.players[traceOffLevelID] = tracePlayer{id: traceOffLevelID, plractive: true, onLevel: true}

	state.trans = transparencyRegions()
	world.list.Activate(light.Point{X: 12, Y: 12}, 6, traceMyID)
	world.list.Activate(light.Point{X: 6, Y: 6}, 5, traceFriendlyID)
	world.list.Activate(light.Point{X: 18, Y: 18}, 5, traceHostileID)
	world.list.Activate(light.Point{X: 20, Y: 6}, 5, traceOffLevelID)

	return world
}

// reveal runs one ProcessVisionList frame over the scenario's players.
func (w *traceWorld) reveal() {
	views := make([]light.PlayerView, 0, len(w.players))
	for _, player := range w.players {
		views = append(views, player)
	}
	w.list.ProcessVisionList(views, visionGrid(w.state))
}

// walledFloor is the BlockLight layout of the committed trace: the walls and
// corner objects the fixture header lists.
func walledFloor() func(position light.Point) bool {
	walls := map[light.Point]bool{}
	for y := 6; y <= 18; y++ {
		walls[light.Point{X: 16, Y: y}] = true
	}
	for x := 3; x <= 10; x++ {
		walls[light.Point{X: x, Y: 7}] = true
	}
	for _, corner := range []light.Point{{X: 10, Y: 16}, {X: 20, Y: 12}, {X: 2, Y: 2}} {
		walls[corner] = true
	}

	return func(position light.Point) bool {
		return withinTraceSide(position) && !walls[position]
	}
}

// transparencyRegions is the dTransVal layout of the committed trace.
func transparencyRegions() map[light.Point]int8 {
	trans := map[light.Point]int8{}
	fillTrans(trans, light.Point{X: 4, Y: 4}, light.Point{X: 8, Y: 8}, 3)
	fillTrans(trans, light.Point{X: 18, Y: 16}, light.Point{X: 22, Y: 20}, 7)
	fillTrans(trans, light.Point{X: 20, Y: 2}, light.Point{X: 22, Y: 4}, 9)

	return trans
}

// fillTrans stamps region over the inclusive rectangle from from to to.
func fillTrans(trans map[light.Point]int8, from, to light.Point, region int8) {
	for x := from.X; x <= to.X; x++ {
		for y := from.Y; y <= to.Y; y++ {
			trans[light.Point{X: x, Y: y}] = region
		}
	}
}

// withinTraceSide reports whether position is inside the trace grid.
func withinTraceSide(position light.Point) bool {
	return position.X >= 0 && position.X < traceSide && position.Y >= 0 && position.Y < traceSide
}

// isInsideSquare reports whether position lies in the (2*radius+1)² square
// centred on centre.
func isInsideSquare(position, centre light.Point, radius int) bool {
	return position.X >= centre.X-radius && position.X <= centre.X+radius &&
		position.Y >= centre.Y-radius && position.Y <= centre.Y+radius
}

// isTrailingBlank reports whether the point at step can be a zero point: the
// table trims its trailing {0,0} run, so every zero point after the first must
// end the row.
func isTrailingBlank(ray, step int, point light.Offset) bool {
	if point != (light.Offset{}) {
		return false
	}
	for _, later := range light.VisionRays[ray][step+1:] {
		if later != (light.Offset{}) {
			return false
		}
	}

	return true
}

// sortedRegions returns the raised transparency regions in ascending order.
func sortedRegions(regions map[int8]bool) []int8 {
	out := make([]int8, 0, len(regions))
	for region := range regions {
		out = append(out, region)
	}
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}

	return out
}

// recordingVisionCallbacks is the DoVision callback set of a test that only
// cares which tiles the rays report, with the given PassesLight rule.
func recordingVisionCallbacks(
	passesLight func(position light.Point) bool,
	visible map[light.Point]int,
) light.VisionCallbacks {
	return light.VisionCallbacks{
		MarkVisible:     func(position light.Point) { visible[position]++ },
		MarkTransparent: func(light.Point) {},
		PassesLight:     passesLight,
		InBounds:        func(position light.Point) bool { return position.X >= 0 && position.Y >= 0 },
	}
}

// readVisionFixture returns the data lines of a committed vision fixture.
func readVisionFixture(t *testing.T, name string) []string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(visionFixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	lines := make([]string, 0, len(raw))
	for _, line := range strings.Split(string(raw), "\n") {
		if trimmed := strings.TrimRight(line, "\r"); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) == 0 {
		t.Fatalf("fixture %s is empty", name)
	}

	return lines
}

// parseVisionRays decodes the ray fixture: the ray_len_adj header line plus 23
// rows of 15 "dx,dy" cells.
func parseVisionRays(lines []string) (rays [23][15]light.Offset, adj [23]uint8, err error) {
	table := rayTable{}
	for _, line := range lines {
		if err := table.addLine(line); err != nil {
			return table.rays, table.adj, err
		}
	}

	return table.rays, table.adj, nil
}

// setAdj decodes the RayLenAdj header.
func (t *rayTable) setAdj(value string) error {
	adj, err := parseRayLenAdj(value)
	if err != nil {
		return err
	}
	t.adj, t.found = adj, true

	return nil
}

// setRow decodes one ray row, rejecting a 24th row.
func (t *rayTable) setRow(line string) error {
	if t.rows >= len(t.rays) {
		return fmt.Errorf("more than %d ray rows", len(t.rays))
	}
	if err := parseRayRow(&t.rays[t.rows], line); err != nil {
		return fmt.Errorf("ray %d: %w", t.rows, err)
	}
	t.rows++

	return nil
}

// parseRayLenAdj decodes the tab-separated RayLenAdj header value.
func parseRayLenAdj(value string) (adj [23]uint8, err error) {
	cells := strings.Fields(strings.TrimSpace(value))
	if len(cells) != len(adj) {
		return adj, fmt.Errorf("%d RayLenAdj cells, want %d", len(cells), len(adj))
	}
	for i, cell := range cells {
		parsed, parseErr := strconv.ParseUint(cell, 10, 8)
		if parseErr != nil {
			return adj, fmt.Errorf("RayLenAdj cell %d %q: %w", i, cell, parseErr)
		}
		adj[i] = uint8(parsed)
	}

	return adj, nil
}

// parseRayRow decodes one tab-separated ray row of "dx,dy" cells.
func parseRayRow(dst *[15]light.Offset, line string) error {
	cells := strings.Split(line, "\t")
	if len(cells) != len(dst) {
		return fmt.Errorf("%d cells, want %d", len(cells), len(dst))
	}
	for i, cell := range cells {
		deltaX, deltaY, err := parseOffset(cell)
		if err != nil {
			return fmt.Errorf("cell %d: %w", i, err)
		}
		dst[i] = light.Offset{DeltaX: deltaX, DeltaY: deltaY}
	}

	return nil
}

// parseOffset decodes a "dx,dy" cell into the two int8 components.
func parseOffset(cell string) (deltaX, deltaY int8, err error) {
	parts := strings.Split(cell, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("cell %q is not dx,dy", cell)
	}
	x, err := strconv.ParseInt(parts[0], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("dx %q: %w", parts[0], err)
	}
	y, err := strconv.ParseInt(parts[1], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("dy %q: %w", parts[1], err)
	}

	return int8(x), int8(y), nil
}

// parseVisionTrace decodes the trace fixture into its named frames. Grid rows
// are indexed [x][y], matching the upstream arrays.
func parseVisionTrace(lines []string) (map[string]visionFrame, error) {
	frames := map[string]visionFrame{}
	parsed := &traceParser{frames: frames}

	for _, line := range lines {
		if err := parsed.readLine(line); err != nil {
			return nil, err
		}
	}
	if err := parsed.finish(); err != nil {
		return nil, err
	}

	return frames, nil
}

// readLine consumes one fixture line, either a section header or a grid row.
func (p *traceParser) readLine(line string) error {
	if strings.HasPrefix(line, "#") {
		return nil
	}

	if name, section, ok := splitTraceSection(line); ok {
		return p.startSection(name, section, line)
	}
	if p.section != traceFlags && p.section != traceExplorer {
		return fmt.Errorf("line %q is not a row of a grid section", line)
	}
	if err := p.readGridRow(line); err != nil {
		return fmt.Errorf("%s.%s row %d: %w", p.frame, p.section, p.row, err)
	}
	p.row++

	return nil
}

// startSection opens a new section and its row counter.
func (p *traceParser) startSection(name, section, line string) error {
	frame, ok := p.frames[name]
	if !ok {
		frame = visionFrame{}
		p.frames[name] = frame
	}
	if err := p.applySection(&frame, section, line); err != nil {
		return err
	}
	p.frames[name] = frame
	p.frame, p.section, p.row = name, section, 0

	return nil
}

// applySection decodes a section that is not a grid of rows.
func (*traceParser) applySection(frame *visionFrame, section, line string) error {
	switch {
	case section == traceFlags, section == traceExplorer:
		return nil
	case strings.HasPrefix(section, traceTransList):
		return applyRegions(frame, strings.TrimPrefix(section, traceTransList), line)
	default:
		return applyActive(frame, section, line)
	}
}

// applyRegions stores the TransList regions of a frame.
func applyRegions(frame *visionFrame, value, line string) error {
	regions, err := parseInts(value)
	if err != nil {
		return fmt.Errorf("%q: %w", line, err)
	}
	frame.regions = regions

	return nil
}

// applyActive stores the VisionActive flags of a frame; a short or long flag
// list is a fixture that no longer matches MAXVISION.
func applyActive(frame *visionFrame, section, line string) error {
	value, ok := strings.CutPrefix(section, traceActive)
	if !ok {
		return fmt.Errorf("unknown section %q", line)
	}
	active, err := parseInts(value)
	if err != nil {
		return fmt.Errorf("%q: %w", line, err)
	}
	if len(active) != light.MaxVision {
		return fmt.Errorf("%q: %d active flags, want %d", line, len(active), light.MaxVision)
	}
	for i, flag := range active {
		frame.active[i] = flag != 0
	}

	return nil
}

// readGridRow decodes one row of the current grid section: one decimal digit
// per tile in the explorer grid, two hex digits per tile in the flag grid.
func (p *traceParser) readGridRow(line string) error {
	if p.row >= traceSide {
		return fmt.Errorf("more than %d rows", traceSide)
	}
	width := traceSide
	if p.section == traceFlags {
		width = 2 * traceSide
	}
	if len(line) != width {
		return fmt.Errorf("%d cells, want %d", len(line), width)
	}

	if p.section == traceExplorer {
		p.setExplorerRow(line)

		return nil
	}

	return p.setFlagRow(line)
}

// setExplorerRow stores one explorer row, one decimal digit per tile.
func (p *traceParser) setExplorerRow(line string) {
	frame := p.frames[p.frame]
	for x := range traceSide {
		frame.explorer[x][p.row] = int(line[x] - '0')
	}
	p.frames[p.frame] = frame
}

// setFlagRow stores one dFlags row, two hex digits per tile.
func (p *traceParser) setFlagRow(line string) error {
	frame := p.frames[p.frame]
	for x := range traceSide {
		cell := line[2*x : 2*x+2]
		flags, err := strconv.ParseUint(cell, 16, 8)
		if err != nil {
			return fmt.Errorf("cell %d %q: %w", x, cell, err)
		}
		frame.flags[x][p.row] = light.DungeonFlag(flags)
	}
	p.frames[p.frame] = frame

	return nil
}

// finish rejects a fixture whose sections are incomplete or inconsistent.
func (p *traceParser) finish() error {
	if len(p.frames) == 0 {
		return errors.New("no frames decoded")
	}
	for name, frame := range p.frames {
		if len(frame.active) != light.MaxVision || frame.explorer == [traceSide][traceSide]int{} &&
			frame.flags == [traceSide][traceSide]light.DungeonFlag{} {
			return fmt.Errorf("frame %q is incomplete", name)
		}
	}

	return nil
}

// splitTraceSection splits a "<frame>.<section>" header line.
func splitTraceSection(line string) (name, section string, ok bool) {
	name, section, ok = strings.Cut(line, ".")

	return name, section, ok
}

// parseInts decodes the whitespace-separated integers of a section value.
func parseInts(value string) ([]int, error) {
	fields := strings.Fields(value)
	out := make([]int, 0, len(fields))
	for _, field := range fields {
		parsed, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field, err)
		}
		out = append(out, parsed)
	}

	return out, nil
}

// assertTraceEqual compares one world against its expected frame: flags,
// automap explorer records, TransList and VisionActive.
func assertTraceEqual(t *testing.T, want visionFrame, world *traceWorld) {
	t.Helper()

	assertFlagsEqual(t, want, world.state)
	assertExplorerEqual(t, want, world.state)
	assertRegionsEqual(t, want, world.state)
	assertActiveEqual(t, want, world.list)
}

// assertFlagsEqual pins dFlags, ignoring the bits outside the trace grid.
func assertFlagsEqual(t *testing.T, want visionFrame, state *visionState) {
	t.Helper()

	for x := range traceSide {
		for y := range traceSide {
			if got := state.flags[light.Point{X: x, Y: y}]; got != want.flags[x][y] {
				t.Fatalf("dFlags[%d][%d] = %#x, want %#x", x, y, got, want.flags[x][y])
			}
		}
	}
}

// assertExplorerEqual pins the automap explorer each tile was last recorded
// with: the fixture stores 0 for "never called" and 1+MAP_EXP_* otherwise.
func assertExplorerEqual(t *testing.T, want visionFrame, state *visionState) {
	t.Helper()

	recorded := lastExplorerPerTile(state.views)
	for x := range traceSide {
		for y := range traceSide {
			if got, wantCell := explorerCell(recorded, x, y, want); got != wantCell {
				t.Fatalf("automap explorer at %v = %d, want %d", light.Point{X: x, Y: y}, got, wantCell)
			}
		}
	}
}

// explorerCell is the fixture encoding of one tile's explorer: 0 for "never
// recorded", 1+MAP_EXP_* for the last recorded kind.
func explorerCell(recorded map[light.Point]light.Exploration, x, y int, want visionFrame) (got, wantCell int) {
	explorer, ok := recorded[light.Point{X: x, Y: y}]
	if !ok {
		return 0, want.explorer[x][y]
	}

	return int(explorer) + 1, want.explorer[x][y]
}

// lastExplorerPerTile keeps only the explorer each tile was last recorded with,
// the value the fixture stores.
func lastExplorerPerTile(views []automapCall) map[light.Point]light.Exploration {
	recorded := make(map[light.Point]light.Exploration, len(views))
	for _, call := range views {
		recorded[call.position] = call.explorer
	}

	return recorded
}

// firstQuadrantFault returns the first point of a ray that leaves the first
// quadrant or ends the walk with a zero point, empty when the ray holds.
func firstQuadrantFault(ray int, points [15]light.Offset) string {
	for step, point := range points {
		if fault := quadrantFault(ray, step, point); fault != "" {
			return fault
		}
	}

	return ""
}

// quadrantFault reports why a table point breaks the one-quadrant invariant,
// or the empty string when the point holds it.
func quadrantFault(ray, step int, point light.Offset) string {
	if point.DeltaX < 0 || point.DeltaY < 0 {
		return fmt.Sprintf("step %d = %v outside the first quadrant", step, point)
	}
	if !isTrailingBlank(ray, step, point) && point == (light.Offset{}) {
		return fmt.Sprintf("ray %d step %d is a zero point before the trailing run", ray, step)
	}

	return ""
}

// assertUnVisionEdges checks the four edge tiles of the rect are cleared and
// the four tiles just past it are untouched.
func assertUnVisionEdges(t *testing.T, state *visionState) {
	t.Helper()

	for _, edge := range []light.Point{{X: 4, Y: 10}, {X: 16, Y: 10}, {X: 10, Y: 4}, {X: 10, Y: 16}} {
		if state.flags[edge]&light.FlagVisionBits != 0 {
			t.Fatalf("edge tile %v not cleared", edge)
		}
	}
	for _, outside := range []light.Point{{X: 3, Y: 10}, {X: 17, Y: 10}, {X: 10, Y: 3}, {X: 10, Y: 17}} {
		if state.flags[outside]&light.FlagVisionBits == 0 {
			t.Fatalf("tile %v outside the rect was cleared", outside)
		}
	}
}

// assertAutomapGated checks the automap hook of one DoVisionFlags call against
// the records it has to have named.
func assertAutomapGated(t *testing.T, got, want []automapCall) {
	t.Helper()

	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("SetAutomapView calls = %v, want %v", got, want)
	}
}

// assertRegionsEqual pins the TransList entries the frame raised.
func assertRegionsEqual(t *testing.T, want visionFrame, state *visionState) {
	t.Helper()

	got := sortedRegions(state.regions)
	if fmt.Sprint(got) != fmt.Sprint(want.regions) {
		t.Fatalf("TransList regions = %v, want %v", got, want.regions)
	}
}

// assertActiveEqual pins the VisionActive flags the frame leaves behind.
func assertActiveEqual(t *testing.T, want visionFrame, list *light.VisionList) {
	t.Helper()

	for id := range want.active {
		if list.Active[id] != want.active[id] {
			t.Fatalf("VisionActive[%d] = %t, want %t", id, list.Active[id], want.active[id])
		}
	}
}
