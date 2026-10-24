package entities

// Item ports struct Item (items.h:189-279). Field order follows the
// upstream declaration; upstream helper methods (isGold, isWeapon, …) are
// behavior and stay out of the data port.
type Item struct {
	// Seed: randomly generated identifier (items.h:191).
	Seed       uint32
	CreateInfo uint16
	// Itype (items.h:193, ItemType itemdat.h:290).
	Itype    int8
	AnimFlag bool
	Position Point
	DelFlag  bool
	// SelectionRegion — rendering-side enum; kept as raw byte.
	SelectionRegion uint8
	PostDraw        bool
	Identified      bool
	// Magical (items.h:199, item_quality items.h:41).
	Magical uint8
	// Name[IName[ItemNameLength]] (items.h:200-201): base and identified
	// display names, NUL-terminated.
	Name  [ItemNameLength]byte
	IName [ItemNameLength]byte
	// Loc (items.h:202, item_equip_type itemdat.h:98).
	Loc int8
	// Class (items.h:203, item_class itemdat.h:89).
	Class  uint8
	Curs   uint8
	Value  int
	Ivalue int
	MinDam uint8
	MaxDam uint8
	AC     int16
	// Flags (items.h:216, ItemSpecialEffect itemdat.h:386).
	Flags uint32
	// MiscId (items.h:217, item_misc_id itemdat.h:433).
	MiscId int8
	// Spell (items.h:218, SpellID spelldat.h:30).
	Spell int32
	// IDidx (items.h:219, _item_indexes).
	IDidx      int16
	Charges    int
	MaxCharges int
	Durability int
	MaxDur     int
	PLDam      int16
	PLToHit    int16
	PLAC       int16
	PLStr      int16
	PLMag      int16
	PLDex      int16
	PLVit      int16
	PLFR       int16
	PLLR       int16
	PLMR       int16
	PLMana     int16
	PLHP       int16
	PLDamMod   int16
	PLGetHit   int16
	PLLight    int16
	SplLvlAdd  int8
	Request    bool
	// Uid: unique item index into UniqueItemList (items.h:236).
	Uid     int
	FMinDam int16
	FMaxDam int16
	LMinDam int16
	LMaxDam int16
	PLEnAc  int16
	// PrePower/SufPower (items.h:247-248, item_effect_type itemdat.h:519).
	PrePower int8
	SufPower int8
	VAdd1    int
	VMult1   int
	VAdd2    int
	VMult2   int
	MinStr   int8
	MinMag   uint8
	MinDex   int8
	StatFlag bool
	// DamAcFlags (items.h:250, ItemSpecialEffectHf itemdat.h:419).
	DamAcFlags uint8
	Buff       uint32
}
