// Package missiles: the AddMissile container append and field seeding.
package missiles

// AddMissileArgs is the AddMissile parameter pack (missiles.cpp:2801-2803).
type AddMissileArgs struct {
	// Src is the launch tile (becomes both position.tile and position.start).
	Src TilePosition
	// Dst is the aim tile handed to the addFn.
	Dst TilePosition
	// Midir is the missile direction.
	Midir Direction
	// Type is the misdat row id.
	Type MissileID
	// Caster is the mienemy_type of the shooter.
	Caster int
	// Source is the player/monster index, or -1 for a trap.
	Source int
	// Dam is the pre-rolled damage.
	Dam int
	// SpellLevel is the caster's spell level.
	SpellLevel int
	// Graphic is the row's _miAnimType sheet.
	Graphic GraphicID
	// CastSound is the row's castSound (empty = none).
	CastSound string
	// Parent is the missile that spawned this one.
	Parent *Missile
}

// AddMissile ports AddMissile (missiles.cpp:2801-2855): append a row, seed the
// fields from the pack, pick the frame group and dispatch the row's addFn.
// It returns nil when the addFn fizzles the spell (parameter.spellFizzled).
//
// Deviations, all waiting on tables/sinks that are not ported yet: the
// Missiles.max_size() capacity rejection (the Go list is unbounded), the
// unique-trans inheritance (monstdat row data) and the per-direction
// AnimDelay/AnimLen reload of setAnimation (missile sprite table indices).
func (l *MissileList) AddMissile(w *HitWorld, args AddMissileArgs) *Missile {
	missile := l.AppendMissile(Missile{})
	seedMissile(missile, args)

	frameGroupFor(missile, args.Midir)

	if args.CastSound != "" { // missiles.cpp:2840-2845
		w.ProcessCallbacks.PlaySfxLoc(args.CastSound, args.Src.X, args.Src.Y)
	}

	behavior, ok := BehaviorFor(args.Type)
	if !ok {
		return nil
	}

	parameter := &AddParameter{Dst: args.Dst, Midir: args.Midir, Parent: args.Parent}

	behavior.Add(missile, parameter)
	if parameter.SpellFizzled {
		return nil
	}

	return missile
}

// seedMissile stores the pack into the row (missiles.cpp:2813-2828).
func seedMissile(m *Missile, args AddMissileArgs) {
	m.Type = args.Type
	m.Caster = args.Caster
	m.Source = args.Source
	m.Dam = args.Dam
	m.SpellLevel = args.SpellLevel
	m.Position.Tile = args.Src
	m.Position.Start = args.Src
	m.AnimAdd = 1
	m.AnimType = int(args.Graphic)
	m.Lid = NoLight
	m.LastCollisionTargetHash = 0
}

// frameGroupFor ports the frame-group pick (missiles.cpp:2834-2837): a sheet
// with fewer than 8 directions keeps frame group 0, the rest follow midir.
//
// Deviation: animFAmt comes from the missile sprite table; graphicFrameCount
// (rotation.go) stands in until the table loader lands, and MissileGraphicID
// ::None (no sheet) reads as 0 frames there.
func frameGroupFor(m *Missile, midir Direction) {
	if graphicFrameCount(m.AnimType) < 8 {
		m.setDefaultFrameGroup()
		return
	}

	setDirection(m, midir)
}

// setDirection ports Missile::setDirection(Direction) (missiles.h) — the
// 8-direction counterpart of SetDirection16 (add.go): the frame group becomes
// the direction and the animation restarts on frame 1.
func setDirection(m *Missile, dir Direction) {
	m.MFNum = int(dir)
	m.AnimFrame = 1
	m.AnimCnt = 0
}
