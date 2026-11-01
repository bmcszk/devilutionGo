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
	// Flags is the row's misdat flags column (tables/misdat.h:119-129) —
	// only the Invisible bit is read, for isDrawn().
	Flags []string
	// AnimFrameAmount is the sprite row's numFrames (animFAmt, misdat.cpp:152);
	// 0 falls back to graphicFrameCount.
	AnimFrameAmount int
	// FrameDelay and FrameLength are the sprite row's per-frame-group
	// timing columns (misdat.cpp:157-160), empty when not loaded.
	FrameDelay  []int
	FrameLength []int
	// Parent is the missile that spawned this one.
	Parent *Missile
}

// AddMissile ports AddMissile (missiles.cpp:2801-2855): append a row, seed the
// fields from the pack, pick the frame group and dispatch the row's addFn.
// It returns nil when the addFn fizzles the spell (parameter.spellFizzled).
//
// Deviations, all waiting on tables/sinks that are not ported yet: the
// Missiles.max_size() capacity rejection (the Go list is unbounded), the
// Missiles.max_size() capacity rejection (the Go list is unbounded) — the
// unique-trans inheritance now reads AddSource and the frame timing comes
// from the row's sprite columns.
func (l *MissileList) AddMissile(w *HitWorld, args AddMissileArgs) *Missile {
	missile := l.AppendMissile(Missile{})
	seedMissile(missile, args)

	inheritUniqueTranslation(w, missile)
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
	m.DrawFlag = isDrawn(args.Flags)
	m.AnimFAmt = args.AnimFrameAmount
	m.AnimDelays = args.FrameDelay
	m.AnimLengths = args.FrameLength
}

// inheritUniqueTranslation ports the unique-trans inheritance
// (missiles.cpp:2827-2832): a monster-side shot from a unique carries its
// palette translation so the missile renders in the unique's colours.
func inheritUniqueTranslation(w *HitWorld, m *Missile) {
	if m.IsTrap() || m.Caster != int(TargetPlayers) {
		return
	}

	if monster := w.AddSource.monsterStats(m.Source); monster.IsUnique {
		m.UniqTrans = monster.UniqTrans + 1
	}
}

// frameAmount reports the sheet's animFAmt: the sprite row's numFrames when
// it was loaded, else the graphicFrameCount table (rotation.go).
func (m *Missile) frameAmount() int {
	if m.AnimFAmt > 0 {
		return m.AnimFAmt
	}

	return graphicFrameCount(m.AnimType)
}

// reloadAnimTiming ports the setAnimation tail: AnimDelay and AnimLen are the
// row's timing columns at the current frame group (misdat.cpp:400-408). Empty
// timing columns (sprite table not loaded) keep the caller's values.
func (m *Missile) reloadAnimTiming() {
	group := m.MFNum
	if group >= 0 && group < len(m.AnimDelays) {
		m.AnimDelay = m.AnimDelays[group]
	}

	if group >= 0 && group < len(m.AnimLengths) {
		m.AnimLen = m.AnimLengths[group]
	}
}

// isDrawn ports MissileData::isDrawn (tables/misdat.h:185-188): the row is
// drawn unless it carries MissileDataFlags::Invisible.
func isDrawn(flags []string) bool {
	for _, flag := range flags {
		if flag == "Invisible" {
			return false
		}
	}

	return true
}

// frameGroupFor ports the frame-group pick (missiles.cpp:2834-2837): a sheet
// with fewer than 8 directions keeps frame group 0, the rest follow midir.
//
// animFAmt comes from the sprite row (frameAmount); MissileGraphicID::None and
// unlisted sheets fall back to graphicFrameCount (rotation.go), which reports
// 1 — the None shape — for sheets the table does not name.
func frameGroupFor(m *Missile, midir Direction) {
	if m.frameAmount() < 8 {
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
	m.reloadAnimTiming()
}
