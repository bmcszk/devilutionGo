package light

// This file is the static light-emitter registry: every AddLight call site in
// upstream, with the radius it asks for and the lifetime it gets. The runtime
// behaviour of those lights is in light_list.go; this file only records what
// each emitter asks for, so a caller can create one without re-deriving its
// radius from the C++ source.

const (
	// FlagAnimated is a light whose radius follows an animation frame
	// (flame traps, firewall, explosions) and is re-set through ChangeLight.
	FlagAnimated SourceFlag = 1 << iota
	// FlagPersistent is a light that lives until its owner is destroyed, not
	// for a fixed number of frames.
	FlagPersistent
	// FlagNetworked is a light created from state the multiplayer delta
	// carries (the monster light is re-placed from CMD_* deltas).
	FlagNetworked
)

const (
	// SrcPlayerTorch is the player's own torch: AddLight(player.position.tile,
	// player._pLightRad) with _pLightRad defaulting to 10 (player.cpp:2334,
	// :2524). items.cpp:2542 re-radii it on equip.
	SrcPlayerTorch SourceType = iota
	// SrcMonsterDefault is the radius-3 light every monster except the Hork
	// Demon gets (monster.cpp:3349-3352).
	SrcMonsterDefault
	// SrcMonsterBerserk is the light a berserked monster gets: 9 in the nest,
	// 3 elsewhere (missiles.cpp:1339).
	SrcMonsterBerserk
	// SrcMonsterData is the radius diablo.cpp:1506 takes from the monster's
	// data light field for berserked monsters on level load.
	SrcMonsterData
	// SrcDiablo is Diablo's radius-8 light (monster.cpp:898).
	SrcDiablo
	// SrcObjectCandle is the radius-5 candle class: OBJ_L1LIGHT, OBJ_SKFIRE,
	// OBJ_CANDLE1, OBJ_CANDLE2, OBJ_BOOKCANDLE (objects.cpp:4208) and the
	// burning crosses (objects.cpp:4260), all via UpdateObjectLight
	// (objects.cpp:1483).
	SrcObjectCandle
	// SrcObjectStoryCandle is the radius-3 story/l5 candle class
	// (objects.cpp:4212).
	SrcObjectStoryCandle
	// SrcObjectTorch is the radius-8 torch class: OBJ_TORCHL, OBJ_TORCHR,
	// OBJ_TORCHL2, OBJ_TORCHR2 (objects.cpp:4241).
	SrcObjectTorch
	// SrcObjectFlameTrap is the radius-1 light of a firing flame trap
	// (objects.cpp:1582), re-radiied from the animation frame at :1596 and
	// :1638 and cleared at :1594.
	SrcObjectFlameTrap
	// SrcMissileRune is the radius-8 light of the rune missiles
	// (missiles.cpp:691).
	SrcMissileRune
	// SrcMissileBigExplosion is the radius-8 light of a big explosion
	// (missiles.cpp:1613) and of immolation (missiles.cpp:1631).
	SrcMissileBigExplosion
	// SrcMissileBolt is the radius-5 light of the charged-bolt bow
	// (missiles.cpp:1726), the elemental arrow (missiles.cpp:1767) and the
	// charged bolt itself (missiles.cpp:2699).
	SrcMissileBolt
	// SrcMissileFirebolt is the radius-8 light of the firebolt
	// (missiles.cpp:1876), magma ball (missiles.cpp:1906), fireball
	// (missiles.cpp:1987), generic magic missile (missiles.cpp:2285),
	// elemental (missiles.cpp:2502), holy bolt (missiles.cpp:2725) and bone
	// spirit (missiles.cpp:2770).
	SrcMissileFirebolt
	// SrcMissileLightning is the radius-4 light of the lightning missile
	// (missiles.cpp:2015).
	SrcMissileLightning
	// SrcMissileFaint is the radius-1 light of the guardian (missiles.cpp:2216),
	// the inferno (missiles.cpp:2667), the town portal (missiles.cpp:3390) and
	// the red portal (missiles.cpp:4173).
	SrcMissileFaint
	// SrcTownPortal is the radius-15 light of the town portal (portal.cpp:59).
	SrcTownPortal
	// SrcFireWall is the firewall's light: radius ExpLight[0] = 2, then the
	// ExpLight ramp re-set every frame (missiles.cpp:3055, :3078-3080).
	SrcFireWall
	// SrcFlameWave is the flame wave's light: the same ExpLight ramp
	// (missiles.cpp:3470, :3490-3491).
	SrcFlameWave
	// SrcWeaponExplosion is the radius-9 light of a weapon explosion
	// (missiles.cpp:3599).
	SrcWeaponExplosion
	// SrcMissileExplosion is the radius-9 light of a missile explosion
	// (missiles.cpp:3623).
	SrcMissileExplosion
)

// lightSourceDefs is the measured inventory: one row per AddLight call site
// shape. Source count and radii come from the grep over the pinned upstream
// (30 AddLight call sites: 23 missiles.cpp, 2 monster.cpp, 2 objects.cpp,
// 1 player.cpp, 1 portal.cpp, 1 diablo.cpp).
var lightSourceDefs = []LightSourceDef{
	{SrcPlayerTorch, 10, FlagPersistent, "player.cpp:2334|player.cpp:2524"},
	{SrcMonsterDefault, 3, FlagPersistent | FlagNetworked, "monster.cpp:3351"},
	{SrcMonsterBerserk, 3, FlagPersistent | FlagNetworked, "missiles.cpp:1339"},
	{SrcMonsterData, 3, FlagPersistent | FlagNetworked, "diablo.cpp:1506"},
	{SrcDiablo, 8, FlagPersistent, "monster.cpp:898"},
	{SrcObjectCandle, 5, FlagPersistent, "objects.cpp:4208|objects.cpp:1483"},
	{SrcObjectStoryCandle, 3, FlagPersistent, "objects.cpp:4212|objects.cpp:1483"},
	{SrcObjectTorch, 8, FlagPersistent, "objects.cpp:4241|objects.cpp:1483"},
	{SrcObjectFlameTrap, 1, FlagAnimated, "objects.cpp:1582|objects.cpp:1596|objects.cpp:1638"},
	{SrcMissileRune, 8, 0, "missiles.cpp:691"},
	{SrcMissileBigExplosion, 8, 0, "missiles.cpp:1613|missiles.cpp:1631"},
	{SrcMissileBolt, 5, 0, "missiles.cpp:1726|missiles.cpp:1767|missiles.cpp:2699"},
	{SrcMissileFirebolt, 8, 0, "missiles.cpp:1876|missiles.cpp:1906|missiles.cpp:1987|" +
		"missiles.cpp:2285|missiles.cpp:2502|missiles.cpp:2725|missiles.cpp:2770"},
	{SrcMissileLightning, 4, 0, "missiles.cpp:2015"},
	{SrcMissileFaint, 1, 0, "missiles.cpp:2216|missiles.cpp:2667|missiles.cpp:3390|missiles.cpp:4173"},
	{SrcTownPortal, 15, FlagPersistent, "portal.cpp:59"},
	{SrcFireWall, 2, FlagAnimated, "missiles.cpp:3055|missiles.cpp:3078"},
	{SrcFlameWave, 2, FlagAnimated, "missiles.cpp:3470|missiles.cpp:3490"},
	{SrcWeaponExplosion, 9, 0, "missiles.cpp:3599"},
	{SrcMissileExplosion, 9, 0, "missiles.cpp:3623"},
}

// SourceType identifies one AddLight call site (or the one caller that creates
// its light through the monster data tables).
type SourceType int

// SourceFlag describes a non-radius property of a light source.
type SourceFlag uint8

// LightSourceDef is one registry row: the light an emitter creates.
type LightSourceDef struct {
	// Type is the row's key.
	Type SourceType
	// Radius is the radius the emitter passes to AddLight. For an animated
	// source it is the first frame of the ramp.
	Radius uint8
	// Flags are the source's non-radius properties.
	Flags SourceFlag
	// Cite is the upstream site(s) the row is measured from, as
	// "file:line" joined by "|".
	Cite string
}

// Has reports whether the source carries flag.
func (d LightSourceDef) Has(flag SourceFlag) bool { return d.Flags&flag != 0 }

// AddSourceLight creates the light a registry row describes, at position, and
// returns its id (NoLight when the pool is full). It is a thin runtime bridge
// between the registry and AddLight; the owner keeps the id to re-radius or
// un-light it.
func (l *LightList) AddSourceLight(sourceType SourceType, position Point) int {
	return l.AddLight(position, GetLightSourceDef(sourceType).Radius)
}

// GetLightSourceDef returns the registry row for a source type; a type outside
// the registry yields the zero row (radius 0, no flags, empty Cite).
func GetLightSourceDef(sourceType SourceType) LightSourceDef {
	for _, def := range lightSourceDefs {
		if def.Type == sourceType {
			return def
		}
	}

	return LightSourceDef{}
}

// LightSourceTypes returns every registered source type in registry order (do
// not mutate).
func LightSourceTypes() []SourceType {
	types := make([]SourceType, 0, len(lightSourceDefs))
	for _, def := range lightSourceDefs {
		types = append(types, def.Type)
	}

	return types
}
