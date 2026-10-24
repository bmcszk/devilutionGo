package entities

// Object ports struct Object (objects.h:32-282). Field order follows the
// upstream declaration; upstream methods (GetId, SetMapRange, …) are
// behavior and stay out of the data port. Upstream sprite handles
// (_oAnimData OptionalClxSpriteList) are asset surface and stay out.
type Object struct {
	// Otype (objects.h:33, _object_id tables/objdat.h:42).
	Otype         int8
	ApplyLighting bool
	TrapFlag      bool
	DoorFlag      bool

	Position Point

	AnimFlag uint32
	// AnimDelay: tick length of each frame (objects.h:41).
	AnimDelay int
	// AnimCnt: counts up to AnimDelay each game tick (objects.h:42).
	AnimCnt   int
	AnimLen   uint32
	AnimFrame uint32
	// AnimWidth: unused upstream, equals first sprite width (objects.h:49).
	AnimWidth uint16

	DelFlag   bool
	Break     int8
	SolidFlag bool
	// MissFlag: true allows missiles through (objects.h:54).
	MissFlag bool
	// SelectionRegion — rendering-side enum; kept as raw byte.
	SelectionRegion uint8
	PreFlag         bool
	Olid            int
	// RndSeed: engine state snapshot for container loot
	// (objects.h:73-76, unsigned to avoid UB on read).
	RndSeed uint32

	Var1 int
	Var2 int
	Var3 int
	Var4 int
	Var5 int
	Var6 uint32
	Var8 int

	// BookMessage (objects.h:85, _speech_id textdat.h:18): quest message
	// for spell-book objects.
	BookMessage int16
}
