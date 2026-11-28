package monsterspawn

// Rnd is the injected random source: upstream's GenerateRnd (random.cpp:96-103).
type Rnd struct {
	// GenerateRnd returns a value in [0,v). PlaceGroup's roll order and count
	// are part of the parity contract, so a scripted source is the test oracle.
	GenerateRnd func(v int32) int32
}

// PlacementWorld is the dungeon state PlaceGroup reads instead of upstream's
// globals: CanPlaceMonster (monster.cpp:282-291) covers dMonster, dPlayer,
// visibility, set pieces and tile solidity, and TransAt is dTransVal.
//
// Tiles placed earlier in the SAME call are not visible through the world:
// PlaceGroup tracks that occupancy itself and the caller applies the returned
// placement in one step (see the package doc).
type PlacementWorld interface {
	// CanPlaceMonster reports whether a monster may stand on the tile.
	CanPlaceMonster(x, y int) bool
	// TransAt is dTransVal[x][y]: the transparency region the scatter stays in
	// (monster.cpp:343).
	TransAt(x, y int) int8
}

// Stats is the monster stat block the placement reads.
type Stats struct {
	// MaxHitPoints is maxHitPoints as InitMonster left it.
	MaxHitPoints int
	// HitPoints is hitPoints as InitMonster left it.
	HitPoints int
	// Intelligence is the monster's intelligence.
	Intelligence int
	// Frames is animInfo.numberOfFrames of the Stand animation: the animation
	// reset rolls GenerateRnd(Frames-1) (monster.cpp:359).
	Frames int
	// Gargoyle reports `ai == MonsterAIID::Gargoyle`: the one AI that keeps its
	// animation instead of resetting to Stand (monster.cpp:357-362).
	Gargoyle bool
}

// Leader is the pack leader PlaceGroup anchors a group around
// (monster.cpp:321-325).
type Leader struct {
	// X, Y are leader->position.tile.
	X, Y int
	// Intelligence is copied into every minion (monster.cpp:354).
	Intelligence int
}

// GroupRequest is one PlaceGroup call's inputs: the caller's counters, the
// monster's base stats and the optional pack leader.
type GroupRequest struct {
	// Num is the requested pack size (`num`).
	Num int
	// Active is ActiveMonsterCount: how many monsters the level already holds.
	Active int
	// Total is totalmonsters: the level's monster budget.
	Total int
	// Base is the placed monster type's stat block.
	Base Stats
	// Leader is the pack leader, or nil for an unled group.
	Leader *Leader
	// Leashed reports the `leashed` argument: minions are bound to the leader
	// and must stay within LeashRadius of the attempt's anchor.
	Leashed bool
}

// Placed is one monster the call placed: the anchor tile plus the state
// PlaceGroup itself wrote (monster.cpp:344-362).
type Placed struct {
	// X, Y is the monster's tile.
	X, Y int
	// Direction is the PlaceMonster direction roll InitMonster receives
	// (monster.cpp:304).
	Direction int
	// MaxHitPoints, HitPoints and Intelligence are the placed monster's stats: a
	// leader's minion gets maxHitPoints doubled, full hitPoints and the leader's
	// intelligence.
	MaxHitPoints, HitPoints, Intelligence int
	// StartFrame is the rolled Stand animation frame, or -1 when the AI keeps
	// its animation (the Gargoyle exception, monster.cpp:357-362).
	StartFrame int
	// Leashed reports whether the minion was leashed to the leader
	// (monster.cpp:355-356).
	Leashed bool
}

// Placement is the outcome of one PlaceGroup call.
type Placement struct {
	// Tiles holds the placed monsters in placement order: the set the caller
	// applies once (upstream applies each as it goes, monster.cpp:345).
	Tiles []Placed
	// Rolls is the number of GenerateRnd calls the call consumed.
	Rolls int
	// Rewinds is the number of monsters a retry unplaced (monster.cpp:312-319).
	Rewinds int
	// PackSize is the leader's packSize, recorded only for a leashed pack
	// (monster.cpp:378-380).
	PackSize int
}

// placer carries one call's inputs and its standing result.
type placer struct {
	world   PlacementWorld
	rnd     Rnd
	req     GroupRequest
	placed  []Placed
	rolls   int
	rewinds int
}

// tile is one dungeon cell.
type tile struct{ x, y int }

// PlaceGroup is PlaceGroup (monster.cpp:308-380): it scatters a monster pack
// of up to req.Num members inside the 16..95 band, retrying the whole attempt
// up to PackRetries times, and returns the placement the caller commits.
//
// The anchor loop has no iteration bound upstream either: a world that never
// accepts an anchor tile loops forever, exactly as the C++ build does.
func PlaceGroup(world PlacementWorld, rnd Rnd, req GroupRequest) Placement {
	placer := placer{world: world, rnd: rnd, req: req}

	return placer.run()
}

// run is the rewind/attempt loop of monster.cpp:312-377.
func (p *placer) run() Placement {
	num := p.req.Num

	for range PackRetries {
		p.rewind()

		xp, yp := p.anchor()
		anchorX, anchorY := xp, yp

		num = p.clamp(num)
		p.scatter(num, tile{x: xp, y: yp}, tile{x: anchorX, y: anchorY})

		if len(p.placed) >= num {
			break
		}
	}

	placement := Placement{Tiles: p.placed, Rolls: p.rolls, Rewinds: p.rewinds}
	if p.req.Leashed && p.req.Leader != nil {
		placement.PackSize = len(p.placed)
	}

	return placement
}

// rewind unplaces the previous attempt's monsters (monster.cpp:313-319).
func (p *placer) rewind() {
	p.rewinds += len(p.placed)
	p.placed = p.placed[:0]
}

// anchor picks the attempt's anchor tile: one direction step from the leader,
// or a retried roll in the 16..95 band (monster.cpp:321-331).
func (p *placer) anchor() (x, y int) {
	if p.req.Leader != nil {
		deltaX, deltaY := displacement(p.roll(DirectionCount))
		leader := p.req.Leader

		return leader.X + deltaX, leader.Y + deltaY
	}

	for {
		xp := p.roll(BandSpan) + BandOrigin
		yp := p.roll(BandSpan) + BandOrigin

		if p.world.CanPlaceMonster(xp, yp) && !p.occupied(tile{x: xp, y: yp}) {
			return xp, yp
		}
	}
}

// clamp is the monster budget clamp (monster.cpp:333-335): the pack never
// overshoots totalmonsters - ActiveMonsterCount. The upstream size_t wrap at
// total < active is not reproduced (package doc).
func (p *placer) clamp(num int) int {
	if num+p.req.Active > p.req.Total {
		return max(p.req.Total-p.req.Active, 0)
	}

	return num
}

// scatter runs one attempt's placement walk (monster.cpp:338-369).
func (p *placer) scatter(num int, spawn, anchor tile) {
	j, attempts := 0, 0

	for j < num && attempts < ScatterIterations {
		if !p.accepts(spawn, anchor) {
			attempts++
			spawn = p.step(spawn)

			continue
		}

		p.place(spawn)
		j++

		// The for-loop increment runs after every body, including the one that
		// completed the pack: its two rolls are part of the call stream.
		spawn = p.step(spawn)
	}
}

// accepts is the placement gate of one candidate tile (monster.cpp:342-347).
func (p *placer) accepts(spawn, anchor tile) bool {
	if !p.world.CanPlaceMonster(spawn.x, spawn.y) || p.occupied(spawn) {
		return false
	}

	if p.world.TransAt(spawn.x, spawn.y) != p.world.TransAt(anchor.x, anchor.y) {
		return false
	}

	if !p.req.Leashed {
		return true
	}

	return abs(spawn.x-anchor.x) < LeashRadius && abs(spawn.y-anchor.y) < LeashRadius
}

// step is the for-loop increment expression of monster.cpp:341, verbatim: both
// axes take the rolled direction's X delta. This is the vanilla bug, kept.
func (p *placer) step(spawn tile) tile {
	deltaX, _ := displacement(p.roll(DirectionCount))
	spawn.x += deltaX

	deltaX, _ = displacement(p.roll(DirectionCount))
	spawn.y += deltaX // vanilla bug, monster.cpp:341: deltaX, not deltaY

	return spawn
}

// place runs PlaceMonster's body plus the leader-minion transform for one tile
// (monster.cpp:292-306, :344-362).
func (p *placer) place(spawn tile) {
	placed := Placed{
		X:            spawn.x,
		Y:            spawn.y,
		Direction:    p.roll(DirectionCount),
		MaxHitPoints: p.req.Base.MaxHitPoints,
		HitPoints:    p.req.Base.HitPoints,
		Intelligence: p.req.Base.Intelligence,
		StartFrame:   -1,
		Leashed:      false,
	}

	if p.req.Leader != nil {
		placed.MaxHitPoints = p.req.Base.MaxHitPoints * MinionHitPointFactor
		placed.HitPoints = placed.MaxHitPoints
		placed.Intelligence = p.req.Leader.Intelligence
		placed.Leashed = p.req.Leashed

		if !p.req.Base.Gargoyle {
			placed.StartFrame = p.roll(p.req.Base.Frames - 1)
		}
	}

	p.placed = append(p.placed, placed)
}

// occupied reports whether this call already placed a monster on the tile.
func (p *placer) occupied(at tile) bool {
	for i := range p.placed {
		if p.placed[i].X == at.x && p.placed[i].Y == at.y {
			return true
		}
	}

	return false
}

// roll draws one value in [0,bound) and counts it.
func (p *placer) roll(bound int) int {
	p.rolls++

	return int(p.rnd.GenerateRnd(int32(bound)))
}

// displacement is Displacement::fromDirection (engine/displacement.hpp:247-276):
// the tile delta of a Direction ordinal (engine/direction.hpp:11-21).
func displacement(direction int) (deltaX, deltaY int) {
	switch direction {
	case 0: // South
		return 1, 1
	case 1: // SouthWest
		return 0, 1
	case 2: // West
		return -1, 1
	case 3: // NorthWest
		return -1, 0
	case 4: // North
		return -1, -1
	case 5: // NorthEast
		return 0, -1
	case 6: // East
		return 1, -1
	case 7: // SouthEast
		return 1, 0
	default: // NoDirection
		return 0, 0
	}
}

// abs is std::abs for the leash comparison (monster.cpp:344).
func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
