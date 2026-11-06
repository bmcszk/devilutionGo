package levels

import "fmt"

// This file ports the Hellfire crypt branch of the L1 generator: the stairs,
// the dirt fix, the substitution chain and the chamber set pieces
// (levels/crypt.cpp:668-840, levels/drlg_l1.cpp:1203-1210).
//
// Deferred: InitCryptPieces and SetCryptSetPieceRoom write dSpecial and the
// UberRow/CornerStone globals, which this package does not model and which the
// committed goldens (dungeon, dTransVal, ViewPosition) do not contain.
// PlaceCryptLights writes dLight and lives in crypt_lights.go.

// placeCryptStairs places the crypt's up and down stairs and points the camera
// at them, reporting whether the attempt produced a usable level
// (levels/crypt.cpp:726-749).
func (d *Dungeon) placeCryptStairs(opts Options) bool {
	success := true

	up, ok := d.PlaceMiniSet(cryptStairsUpMiniset(d.CurrLevel), DMAXX*DMAXY, true)
	if !ok {
		success = false
	} else if opts.Entry == EntryMain || opts.Entry == EntryTwarPDn {
		d.ViewPosition = up.MegaToWorld().Add(Point{X: 3, Y: 5})
	}

	// The lowest crypt has no stairs down (levels/crypt.cpp:740-748).
	if d.CurrLevel == 24 {
		return success
	}

	down, ok := d.PlaceMiniSet(cryptStairDown, DMAXX*DMAXY, true)
	if !ok {
		return false
	}

	if opts.Entry == EntryPrev {
		d.ViewPosition = down.MegaToWorld().Add(Point{X: 3, Y: 7})
	}

	return success
}

// cryptStairsUpMiniset is the stairs-up pattern PlaceCryptStairs asks for: the
// town stairs on the first crypt level, the Hellfire stairs everywhere else
// (levels/crypt.cpp:731).
func cryptStairsUpMiniset(level int) Miniset {
	if level == 21 {
		return cryptStairsTown
	}

	return cryptStairsUpHF
}

// fixCryptDirtTiles rewrites every crypt dirt tile to its closed variant
// (levels/crypt.cpp:708-724). Unlike the cathedral pass it reads no
// neighbours.
func (d *Dungeon) fixCryptDirtTiles() {
	replacements := map[Tile]Tile{
		cryptDirtVwall:    cryptDirtVWall2,
		cryptDirtCorner:   cryptDirtCorner2,
		cryptDirtHWallEnd: cryptDirtHWallEnd2,
		cryptDirtVWallEnd: cryptDirtVWallEnd2,
		cryptDirtHwall:    cryptDirtHWall2,
	}

	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			if replacement, ok := replacements[d.Tiles[i][j]]; ok {
				d.Tiles[i][j] = replacement
			}
		}
	}
}

// cryptSubstitution runs the crypt decoration chain: the statues, the wall
// section variants and the per-level cracked/broken/leaking passes
// (levels/crypt.cpp:751-799).
func (d *Dungeon) cryptSubstitution() {
	for _, pair := range cryptStatues {
		d.PlaceMiniSetRandom1x1(pair.search, pair.replace, 10)
	}

	d.PlaceMiniSetRandom1x1(cryptVArch, cryptVArch5, 95)
	d.PlaceMiniSetRandom1x1(cryptHArch, cryptHArch5, 95)
	d.PlaceMiniSetRandom(cryptVWallSection, 100)
	d.PlaceMiniSetRandom(cryptHWallSection, 100)
	d.PlaceMiniSetRandom(cryptFloorLave, 60)
	d.applyCryptShadowsPatterns()

	d.cryptDamage()
}

// cryptDamage runs the level-dependent tail of CryptSubstitution: the order of
// the cracked/broken/leaking passes and their odds differ per crypt level
// (levels/crypt.cpp:770-797).
func (d *Dungeon) cryptDamage() {
	switch d.CurrLevel {
	case 21:
		d.cryptCracked()
		d.cryptBroken(15)
		d.cryptLeaking(5)
		d.applyCryptShadowsPatterns()
		d.cryptFloor()
		d.cryptSubstitutions1(5)
		d.cryptSubstitutions2(20)
	case 22:
		d.cryptFloor()
		d.cryptSubstitutions1(10)
		d.cryptSubstitutions2(20)
		d.cryptCracked()
		d.cryptBroken(20)
		d.cryptLeaking(10)
		d.applyCryptShadowsPatterns()
	case 23:
		d.cryptFloor()
		d.cryptSubstitutions1(15)
		d.cryptSubstitutions2(30)
		d.cryptCracked()
		d.cryptBroken(20)
		d.cryptLeaking(15)
		d.applyCryptShadowsPatterns()
	default:
		d.cryptFloor()
		d.cryptSubstitutions1(20)
		d.cryptSubstitutions2(30)
		d.cryptCracked()
		d.cryptBroken(20)
		d.cryptLeaking(20)
		d.applyCryptShadowsPatterns()
	}
}

// cryptCracked swaps each cracked tile for its damaged variant at the 30% every
// level uses; cryptBroken and cryptLeaking take the per-level odds
// (levels/crypt.cpp:617-639, 772-797).
func (d *Dungeon) cryptCracked() { d.replaceTiles(cryptCrackedTiles, 30) }

func (d *Dungeon) cryptBroken(rndper int32) { d.replaceTiles(cryptBrokenTiles, rndper) }

func (d *Dungeon) cryptLeaking(rndper int32) { d.replaceTiles(cryptLeakingTiles, rndper) }

// cryptSubstitutions1 applies the first variant family
// (levels/crypt.cpp:641-651).
func (d *Dungeon) cryptSubstitutions1(rndper int32) {
	d.replaceTiles(cryptSubstitutions1Tiles, rndper)
}

// cryptSubstitutions2 applies the second variant family
// (levels/crypt.cpp:653-666).
func (d *Dungeon) cryptSubstitutions2(rndper int32) {
	for _, miniset := range cryptSubstitutions2Sets {
		d.PlaceMiniSetRandom(miniset, rndper)
	}

	d.replaceTiles(cryptSubstitution1Floor, rndper)
}

// cryptFloor applies the plain floor variants at the 10% every level uses
// (levels/crypt.cpp:668-673, 772-797).
func (d *Dungeon) cryptFloor() {
	d.replaceTiles(cryptSubstitution2Floor, 10)
}

// replaceTiles applies one search/replace table with the 1x1 random placement
// (levels/crypt.cpp:612-639).
func (d *Dungeon) replaceTiles(table []replaceTile, rndper int32) {
	for _, pair := range table {
		d.PlaceMiniSetRandom1x1(pair.search, pair.replace, rndper)
	}
}

// fillCryptChamber places the crypt's chamber set piece: the uber room on the
// lowest level, the cornerstone room when the quest item is available
// (levels/drlg_l1.cpp:1069-1075, levels/crypt.cpp:676-704).
func (d *Dungeon) fillCryptChamber(opts Options) error {
	switch {
	case d.CurrLevel == 24:
		return d.setCryptRoom(opts)
	case opts.Quests.CornerStone:
		return d.setCornerRoom(opts)
	default:
		return nil
	}
}

// setCryptRoom places the uber room in the selected chamber
// (levels/crypt.cpp:676-693).
func (d *Dungeon) setCryptRoom(opts Options) error {
	data, err := dungeonSetPiece(opts, "nlevels/l5data/uberroom.dun")
	if err != nil {
		return err
	}

	position := d.selectChamber()
	d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}
	d.PlaceDunTiles(data, position, 0)

	return nil
}

// setCornerRoom places the cornerstone room in the selected chamber
// (levels/crypt.cpp:695-704).
func (d *Dungeon) setCornerRoom(opts Options) error {
	data, err := dungeonSetPiece(opts, "nlevels/l5data/cornerstone.dun")
	if err != nil {
		return err
	}

	position := d.selectChamber()
	d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}
	d.PlaceDunTiles(data, position, 0)

	return nil
}

// dungeonSetPiece resolves one .dun resource from the caller's set-piece map.
func dungeonSetPiece(opts Options, path string) ([]byte, error) {
	data, ok := opts.SetPieces[path]
	if !ok {
		return nil, fmt.Errorf("set piece %q: %w", path, ErrMissingSetPiece)
	}

	return data, nil
}
