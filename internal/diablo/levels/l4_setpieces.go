package levels

// This file ports the hell set piece passes: the warlord room, the four Diablo
// quadrants of level 16 and the quest drawing DRLG_CheckQuests does for hell
// (levels/drlg_l4.cpp:170-182, 955-1014, levels/drlg_quests.cpp:34-45).

// initHellSetPiece is InitSetPiece: the warlord room is stamped at the set piece
// room the first layout reserved for it (levels/drlg_l4.cpp:170-182).
//
// The multiplayer vile1.dun branch is not ported: it is guarded by
// UseMultiplayerQuests, which is false for every single-player level, and hell
// has no other set piece.
func (d *Dungeon) initHellSetPiece(opts Options) error {
	if !opts.Quests.Warlord {
		return nil
	}

	data, err := dungeonSetPiece(opts, hellWarlordSetPiece)
	if err != nil {
		return err
	}

	d.PlaceDunTiles(data, d.SetPieceRoom.Position, hellFloor)
	d.SetPiece = Rectangle{Position: d.SetPieceRoom.Position, Size: GetDunSize(data)}

	return nil
}

// protectQuads is ProtectQuads: the 14x14 corner of every level 16 quadrant is
// excluded from the wall and transparency passes (levels/drlg_l4.cpp:955-965).
func (d *Dungeon) protectQuads() {
	for y := range hellQuadSide {
		for x := range hellQuadSide {
			d.Protected.Set(d.HoldL4.X+x, d.HoldL4.Y+y, true)
			d.Protected.Set(DMAXX-1-x-d.HoldL4.X, d.HoldL4.Y+y, true)
			d.Protected.Set(d.HoldL4.X+x, DMAXY-1-y-d.HoldL4.Y, true)
			d.Protected.Set(DMAXX-1-x-d.HoldL4.X, DMAXY-1-y-d.HoldL4.Y, true)
		}
	}
}

// loadPreGateDiabloQuads is LoadDiabQuads(true): the quadrants as they look
// before the gate opens (levels/drlg_l4.cpp:967-989).
func (d *Dungeon) loadPreGateDiabloQuads(opts Options) error {
	return d.loadDiabloQuads(opts, [3]string{
		"levels/l4data/diab2b.dun",
		"levels/l4data/diab3b.dun",
		"levels/l4data/diab4b.dun",
	})
}

// loadPostGateDiabloQuads is LoadDiabQuads(false), the form the level ends in.
func (d *Dungeon) loadPostGateDiabloQuads(opts Options) error {
	return d.loadDiabloQuads(opts, [3]string{
		"levels/l4data/diab2a.dun",
		"levels/l4data/diab3a.dun",
		"levels/l4data/diab4a.dun",
	})
}

// loadDiabloQuads stamps the four quadrants around the hold position: diab1 is
// shared, the other three come in a pre-gate and a post-gate variant
// (levels/drlg_l4.cpp:967-989).
func (d *Dungeon) loadDiabloQuads(opts Options, variants [3]string) error {
	quads := [4]struct {
		path     string
		position Point
	}{
		{hellDiabloQuad1, d.HoldL4.Add(Point{X: 4, Y: 4})},
		{variants[0], Point{X: 27 - d.HoldL4.X, Y: 1 + d.HoldL4.Y}},
		{variants[1], Point{X: 1 + d.HoldL4.X, Y: 27 - d.HoldL4.Y}},
		{variants[2], Point{X: 28 - d.HoldL4.X, Y: 28 - d.HoldL4.Y}},
	}

	for _, quad := range quads {
		data, err := dungeonSetPiece(opts, quad.path)
		if err != nil {
			return err
		}

		d.PlaceDunTiles(data, quad.position, hellFloor)
	}

	return nil
}

// drawHellQuests is DRLG_CheckQuests for the hell quest branch: the warlord room
// is redrawn from warlord2.dun over the set piece room
// (levels/drlg_quests.cpp:34-45).
func (d *Dungeon) drawHellQuests(opts Options) error {
	if !opts.Quests.Warlord {
		return nil
	}

	data, err := dungeonSetPiece(opts, hellWarlordQuestSetPiece)
	if err != nil {
		return err
	}

	position := d.SetPieceRoom.Position
	d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}
	d.PlaceDunTiles(data, position, hellFloor)

	return nil
}
