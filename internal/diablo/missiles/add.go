// Package missiles: the add-time (Add*) bodies of the misdat table rows.
//
// The C++ Add* functions read the Players/Monsters globals, the lighting
// engine and the RNG (missiles.cpp:1772-2800). The players/monsters tables
// are not ported yet and entities.Player imports this package, so the reads
// are injected through AddSource (field of HitWorld) and the ProcessCallbacks
// sink — the hitCallbacks pattern of collision.go.
package missiles

// Item special-effect bits the add bodies query (tables/itemdat.h:386-414).
const (
	ItemSpecialEffectRandomArrowVelocity = 1 << 2
	ItemSpecialEffectQuickAttack         = 1 << 17
	ItemSpecialEffectFastAttack          = 1 << 18
	ItemSpecialEffectFasterAttack        = 1 << 19
	ItemSpecialEffectFastestAttack       = 1 << 20
)

// Hero classes the arrow velocity table branches on
// (tables/playerdat.hpp:21-26, enum class HeroClass).
const (
	HeroClassWarrior   = 0
	HeroClassRogue     = 1
	HeroClassSorcerer  = 2
	HeroClassMonk      = 3
	HeroClassBard      = 4
	HeroClassBarbarian = 5
)

// Missile graphic ids the add bodies name (tables/misdat.h:35-99, enum class
// MissileGraphicID; the ordinal is the missile_sprites.tsv data row index).
const (
	// GraphicMagmaBall is the MagmaBall row's sheet (8 frame groups).
	GraphicMagmaBall GraphicID = 24
	// GraphicAcid is the Acid row's sheet (16 frame groups).
	GraphicAcid GraphicID = 31
	// GraphicBloodStar is the BloodStar row's sheet (1 frame group).
	GraphicBloodStar       GraphicID = 22
	GraphicBloodStarBlue   GraphicID = 40
	GraphicBloodStarYellow GraphicID = 42
	GraphicBloodStarRed    GraphicID = 44
)

// Monster type ordinals the BloodStar animation select tests
// (tables/monstdat.h:262-265). The `enum _monster_id` table starts at
// MT_NZOMBIE = 0, so the ordinal is the monstdat.tsv data row index.
const (
	MonsterTypeSuccubus  = 101
	MonsterTypeSnowWich  = 102
	MonsterTypeHellSpawn = 103
	MonsterTypeSolbrner  = 104
)

// MonsterTypeFamiliar is MT_FAMILIAR (tables/monstdat.h:202) — the ordinal
// shares the MT_NZOMBIE = 0 base (tables/monstdat.h:161).
const MonsterTypeFamiliar = 41

// arrowAttackSpeedFlags is the hellfire attack-speed flag set of AddArrow,
// in upstream test order (missiles.cpp:1791-1792).
var arrowAttackSpeedFlags = []int{
	ItemSpecialEffectQuickAttack,
	ItemSpecialEffectFastAttack,
	ItemSpecialEffectFasterAttack,
	ItemSpecialEffectFastestAttack,
}

// arrowAttackSpeedBonus is the velocity increment per attack-speed flag
// (missiles.cpp:1791-1792: ++, += 2, += 4, += 8).
var arrowAttackSpeedBonus = map[int]int{
	ItemSpecialEffectQuickAttack:   1,
	ItemSpecialEffectFastAttack:    2,
	ItemSpecialEffectFasterAttack:  4,
	ItemSpecialEffectFastestAttack: 8,
}

// AddParameter ports AddMissileParameter (missiles.cpp:2795-2800): the pack
// PlayerLife is the life state AddHealing reads and writes (player.h:236-237,
// Fixed26_6 raw values).
type PlayerLife struct {
	// HitPoints is _pHitPoints.
	HitPoints int
	// MaxHitPoints is _pMaxHP.
	MaxHitPoints int
	// HPBase is _pHPBase.
	HPBase int
	// MaxHPBase is _pMaxHPBase.
	MaxHPBase int
}

// AddMissile hands to the row's addFn, plus the spellFizzled output flag.
type AddParameter struct {
	// Dst is the target tile the motion is aimed at (parameter.dst).
	Dst TilePosition
	// Midir is the missile direction (parameter.midir).
	Midir Direction
	// Parent is the missile that spawned this one (parameter.parent).
	Parent *Missile
	// SpellFizzled tells AddMissile to drop the missile (parameter.spellFizzled).
	SpellFizzled bool
}

// AddPlayerStats is the add-time player read of one player index.
type AddPlayerStats struct {
	// Class is _pClass (tables/playerdat.hpp HeroClass).
	Class int
	// Level is getCharacterLevel().
	Level int
	// Magic is _pMagic.
	Magic int
	// Tile is position.tile (missiles.cpp:1950 reads it for NovaBall).
	Tile TilePosition
	// HasItemFlag answers HasAnyOf(_pIFlags, flag) (tables/itemdat.h:386).
	HasItemFlag func(flag int) bool
}

// AddMonsterStats is the add-time monster read of one monster index.
type AddMonsterStats struct {
	// Type is type().type (tables/monstdat.h:160+).
	Type int
	// Intelligence is intelligence (monster.h:258).
	Intelligence int
	// MinDamage/MaxDamage are the projectile damage range (missiles.cpp:260-264).
	MinDamage int
	MaxDamage int
	// UniqLevel is UniqueMonstersData[uniqtype].mlevel, the base level of a
	// unique monster (monster.h:400-401).
	UniqLevel int
	// Level is monster.data().level (tables/monstdat.h) — the level column
	// ProcessAcidSplate reads to size the puddle damage (missiles.cpp:3648).
	Level int
	// IsUnique is monster.isUnique() — uniqtype != 0 (monster.h).
	IsUnique bool
	// UniqTrans is monster.uniqTrans, the palette row of the unique's
	// translation (missiles.cpp:2828-2831).
	UniqTrans int
}

// AddSource is the injected replacement for the Players/Monsters globals the
// add bodies read.
type AddSource struct {
	PlayerStats  func(index int) AddPlayerStats
	MonsterStats func(index int) AddMonsterStats
}

// NoopAddSource returns the no-op defaults (zero stats, no item flags).
func NoopAddSource() AddSource {
	return AddSource{
		PlayerStats:  func(int) AddPlayerStats { return AddPlayerStats{} },
		MonsterStats: func(int) AddMonsterStats { return AddMonsterStats{} },
	}
}

// AddArrow ports AddArrow (missiles.cpp:1772-1797): velocity from the class
// and level table (plus the hellfire attack-speed flags), the direction-16
// frame, and the 256-tick duration.
func AddArrow(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1775-1777
		dst = shiftTile(dst, p.Midir)
	}

	av := 32
	if m.Caster == int(TargetMonsters) { // cpp:1779
		av = arrowVelocity(w, m, av)
	}

	UpdateMissileVelocity(m, dst, av)                            // cpp:1793
	m.AnimFrame = int(GetDirection16(m.Position.Start, dst)) + 1 // cpp:1794
	m.Duration = 256                                             // cpp:1795
}

// AddFirebolt ports AddFirebolt (missiles.cpp:1860-1894): the spell-level
// velocity cap, the direction-16 frame group, the 8-radius light and the
// unrolled damage (player magic / monster range / trap dice).
func AddFirebolt(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1863-1865
		dst = shiftTile(dst, p.Midir)
	}

	sp := 26
	if m.Caster == int(TargetMonsters) { // cpp:1867-1872
		sp = 16
		if !m.IsTrap() {
			sp += spellLevelVelocity(m.SpellLevel)
		}
	}

	UpdateMissileVelocity(m, dst, sp)                        // cpp:1873
	SetDirection16(m, GetDirection16(m.Position.Start, dst)) // cpp:1874
	m.Duration = 256                                         // cpp:1875
	m.Var1 = m.Position.Start.X                              // cpp:1876
	m.Var2 = m.Position.Start.Y                              // cpp:1877
	m.Lid = addMissileLight(w, m)                            // cpp:1878
	m.Dam = projectileDamage(w, m)                           // cpp:1879-1892
}

// AddMagmaBall ports AddMagmaBall (missiles.cpp:1896-1922): the 3-tick
// velocity lead, the position snap, the hellfire whole-pixel duration
// shortcut and the light/damage block.
func AddMagmaBall(w *HitWorld, m *Missile, p *AddParameter) {
	UpdateMissileVelocity(m, p.Dst, 16) // cpp:1898
	lead := PositionDisplacement{       // cpp:1899-1900
		DeltaX: 3 * m.Position.Velocity.DeltaX,
		DeltaY: 3 * m.Position.Velocity.DeltaY,
	}
	m.Position.Traveled.DeltaX += lead.DeltaX
	m.Position.Traveled.DeltaY += lead.DeltaY
	UpdateMissilePos(w, m)               // cpp:1901
	m.Duration = magmaBallDuration(w, m) // cpp:1902-1905
	m.Var1 = m.Position.Start.X          // cpp:1907
	m.Var2 = m.Position.Start.Y          // cpp:1907
	m.Lid = addMissileLight(w, m)        // cpp:1908
	m.Dam = projectileDamage(w, m)       // cpp:1909-1921
}

// AddGenericMagicMissile ports AddGenericMagicMissile (missiles.cpp:
// 2285-2325) — the BloodStar row's addFn: fixed velocity 16, the
// succubus-family animation override and the player damage formula without a
// roll.
func AddGenericMagicMissile(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:2287-2289
		dst = shiftTile(dst, p.Midir)
	}

	UpdateMissileVelocity(m, dst, 16) // cpp:2290
	m.Duration = 256                  // cpp:2291
	m.Var1 = m.Position.Start.X       // cpp:2292
	m.Var2 = m.Position.Start.Y       // cpp:2293
	m.Lid = addMissileLight(w, m)     // cpp:2294

	if m.Caster != int(TargetMonsters) && m.Source > 0 { // cpp:2295
		applyMagicMissileSheet(w, m)
	}

	// animFAmt comes from the row's sprite (frameAmount), missiles.cpp:2308.
	if m.frameAmount() == 16 { // cpp:2308-2310
		SetDirection16(m, GetDirection16(m.Position.Start, dst))
	}

	m.Dam = magicMissileDamage(w, m) // cpp:2312-2325
}

// SetDirection16 ports Missile::setDirection(Direction16)
// (missiles.h:210-213 → setFrameGroupRaw :180-184): the frame group becomes
// the direction and the animation restarts on frame 1.
//
// The setAnimation tail reloads the frame group's AnimDelay/AnimLen
// (missiles.cpp:90-97 → reloadAnimTiming); a missile without the row's timing
// columns keeps the caller's values.
func SetDirection16(m *Missile, dir Direction16) {
	m.MFNum = int(dir)
	m.reloadAnimTiming()
	m.AnimFrame = 1
	m.AnimCnt = 0
}

// AddAcid ports AddAcid (missiles.cpp:2328-2352): a straight shot at
// velocity 16 (no midir nudge), the intelligence-scaled duration, the
// NO_LIGHT sentinel and the closing PutMissile.
func AddAcid(w *HitWorld, m *Missile, p *AddParameter) {
	UpdateMissileVelocity(m, p.Dst, 16)                        // cpp:2330
	SetDirection16(m, GetDirection16(m.Position.Start, p.Dst)) // cpp:2331
	m.Duration = acidDuration(w, m)                            // cpp:2332-2335
	m.Lid = NoLight                                            // cpp:2336
	m.Var1 = m.Position.Start.X                                // cpp:2337
	m.Var2 = m.Position.Start.Y                                // cpp:2338
	m.Dam = acidDamage(w, m)                                   // cpp:2339-2350
	PutMissile(w, m)                                           // cpp:2351
}

// AddChargedBolt ports AddChargedBolt (missiles.cpp:2697-2713): the random
// burnout radius, the magic-quarter damage (fixed 15 for monsters), the random
// start frame, light 5 and speed 8.
//
// Deviation: the damage roll keys off _micaster only (cpp:2701), so a trap —
// which keeps TARGET_MONSTERS — would read Players[-1]._pMagic. The port asks
// the injected sink instead of indexing the player array out of bounds.
func AddChargedBolt(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	m.Mirnd = uint8(w.ProcessCallbacks.GenerateRnd(15) + 1) // cpp:2700
	m.Dam = chargedBoltDamage(w, m)                         // cpp:2701

	if m.Position.Start == dst { // cpp:2703-2705
		dst = shiftTile(dst, p.Midir)
	}

	m.AnimFrame = randomIntBetween(w, 1, 8)                                        // cpp:2706
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 5) // cpp:2707

	UpdateMissileVelocity(m, dst, 8) // cpp:2709
	m.Var1 = 5                       // cpp:2710
	m.Var2 = int(p.Midir)            // cpp:2711
	m.Duration = 256                 // cpp:2712
}

// AddHolyBolt ports AddHolyBolt (missiles.cpp:2715-2736): speed 16 plus twice
// the spell level (capped at 47, traps keep the base), the 16-step frame
// group, light 8 and the character-level damage roll.
//
// Deviation: upstream reads Players[missile._misource].getCharacterLevel()
// unconditionally (cpp:2725, 2735), so a trap (source -1) would index
// Players[-1] and a monster source would read a player slot. The sibling
// `if (!IsTrap())` guard at cpp:2724 shows traps are expected to reach this
// body. The port asks the injected sink instead: a trap's level contributes 0
// and a monster source reads whatever the sink maps that index to.
func AddHolyBolt(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:2718-2720
		dst = shiftTile(dst, p.Midir)
	}

	sp := 16 // cpp:2721
	if !m.IsTrap() {
		sp += min(m.SpellLevel*2, 47) // cpp:2723
	}

	UpdateMissileVelocity(m, dst, sp)                                              // cpp:2728
	SetDirection16(m, GetDirection16(m.Position.Start, dst))                       // cpp:2729
	m.Duration = 256                                                               // cpp:2730
	m.Var1 = m.Position.Start.X                                                    // cpp:2731
	m.Var2 = m.Position.Start.Y                                                    // cpp:2732
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 8) // cpp:2733
	m.Dam = holyBoltDamage(w, m)                                                   // cpp:2734
}
