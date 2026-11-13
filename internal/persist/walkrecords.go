package persist

import "fmt"

// tileWidth is TILE_WIDTH (Source/levels/dun_tile.hpp), the sprite-tile side
// the object width2 centering subtracts.
const tileWidth = 64

// Record sizes of the SaveGameData walk sections this file encodes
// (Source/loadsave.cpp:1703-1822).
const (
	// LightSaveSize is the SaveLighting record length: thirteen int32 fields
	// (loadsave.cpp:1795-1812, memory_map lightning.txt).
	LightSaveSize = 52
	// PortalSaveSize is the SavePortal record length: six int32 fields
	// (loadsave.cpp:1813-1822).
	PortalSaveSize = 24
	// ObjectSaveSize is the SaveObject record length
	// (loadsave.cpp:1703-1746, memory_map object.txt).
	ObjectSaveSize = 120
	// MaxMonsters is MaxMonsters (Source/monster.h:38): the MonsterKillCounts
	// padding bound (loadsave.cpp:2841-2843).
	MaxMonsters = 200
	// MaxObjects is MAXOBJECTS (Source/objects.h:30): the fixed object id
	// table width (loadsave.cpp:2870-2873).
	MaxObjects = 127
	// MaxLights is MAXLIGHTS (Source/lighting.h:41): the fixed light id table
	// width (loadsave.cpp:2874-2875).
	MaxLights = 32
	// MaxPortals is MAXPORTAL (Source/portal.h:19): the fixed portal count
	// (loadsave.cpp:2808-2810).
	MaxPortals = 4
	// bitOne is the upstream uint32 image of a set bool flag.
	bitOne = uint32(1)
)

// LightRecord is struct Light as SaveLighting writes it
// (Source/lighting.h:30-36, loadsave.cpp:1795-1812).
type LightRecord struct {
	X, Y       int32
	Radius     int32
	IsInvalid  bool
	HasChanged bool
	OldX, OldY int32
	OldRadius  int32
	OffsetX    int32
	OffsetY    int32
}

// PortalRecord is struct Portal as SavePortal writes it
// (Source/portal.h:11-17, loadsave.cpp:1813-1822).
type PortalRecord struct {
	Open     bool
	X, Y     int32
	Level    int32
	LType    int32
	SetLevel bool
}

// ObjectRecord is struct Object as SaveObject writes it
// (Source/object.h, loadsave.cpp:1703-1746). Field names mirror the upstream
// members.
type ObjectRecord struct {
	Type            int32
	X, Y            int32
	ApplyLighting   bool
	AnimFlag        bool
	AnimDelay       int32
	AnimCnt         int32
	AnimLen         uint32
	AnimFrame       uint32
	AnimWidth       int32
	DelFlag         bool
	Break           int8
	SolidFlag       bool
	MissFlag        bool
	SelectionRegion uint8
	PreFlag         bool
	TrapFlag        bool
	DoorFlag        bool
	OLid            int32
	ORndSeed        uint32
	Var1            int32
	// IsDynamicLight marks the object type set SaveObject zero-var1s when the
	// level data is saved (loadsave.cpp:1749-1763).
	IsDynamicLight bool
	Var2, Var3     int32
	Var4, Var5     int32
	Var6           uint32
	BookMessage    int32
	Var8           int32
}

// boolField is one bool record field saved as the upstream 0/1 uint32. The
// value rides in the struct, so no helper takes a control flag.
type boolField struct {
	set bool
}

// image returns the 0/1 uint32 upstream writes.
func (f boolField) image() uint32 {
	if f.set {
		return bitOne
	}

	return 0
}

// Encode returns the standing SaveLighting image: the _lid and _lflags fields
// read 0 (loadsave.cpp:1795-1812).
func (l LightRecord) Encode() []byte {
	return l.encodeWithVision(0)
}

// EncodeVision returns the vision SaveLighting image: the _lid and _lflags
// fields read 1 (loadsave.cpp:1803, :1812).
func (l LightRecord) EncodeVision() []byte {
	return l.encodeWithVision(1)
}

// encode writes the record with visionBit in both vision-marked fields.
func (l LightRecord) encodeWithVision(visionBit uint32) []byte {
	var out recordWriter

	out.i32(l.X)
	out.i32(l.Y)
	out.i32(l.Radius)
	out.i32(int32(visionBit)) // _lid
	out.u32(boolField{set: l.IsInvalid}.image())
	out.u32(boolField{set: l.HasChanged}.image())
	out.zeroes(4) // unused
	out.i32(l.OldX)
	out.i32(l.OldY)
	out.i32(l.OldRadius)
	out.i32(l.OffsetX)
	out.i32(l.OffsetY)
	out.u32(visionBit) // _lflags

	return out.buf
}

// DecodeLightRecord parses a SaveLighting image and reports whether it is the
// vision variant.
func DecodeLightRecord(buf []byte) (LightRecord, bool, error) {
	in := recordReader{buf: buf}

	var light LightRecord

	light.X = int32(in.le32())
	light.Y = int32(in.le32())
	light.Radius = int32(in.le32())
	vision := int32(in.le32()) != 0
	light.IsInvalid = in.le32() != 0
	light.HasChanged = in.le32() != 0
	in.skip(4)
	light.OldX = int32(in.le32())
	light.OldY = int32(in.le32())
	light.OldRadius = int32(in.le32())
	light.OffsetX = int32(in.le32())
	light.OffsetY = int32(in.le32())

	if in.err != nil {
		return LightRecord{}, false, in.err
	}

	return light, vision, nil
}

// Encode returns the SavePortal image. A set-level portal keeps the raw
// ltype; a level portal maps crypt/nest to their vanilla-compatible types
// (loadsave.cpp:1820, getHellfireLevelType :1148-1157).
func (p PortalRecord) Encode() []byte {
	var out recordWriter

	out.u32(boolField{set: p.Open}.image())
	out.i32(p.X)
	out.i32(p.Y)
	out.i32(p.Level)
	out.i32(p.levelType())
	out.u32(boolField{set: p.SetLevel}.image())

	return out.buf
}

// DecodePortalRecord parses a SavePortal image.
func DecodePortalRecord(buf []byte) (PortalRecord, error) {
	in := recordReader{buf: buf}

	var portal PortalRecord

	portal.Open = in.le32() != 0
	portal.X = int32(in.le32())
	portal.Y = int32(in.le32())
	portal.Level = int32(in.le32())
	portal.LType = int32(in.le32())
	portal.SetLevel = in.le32() != 0

	if in.err != nil {
		return PortalRecord{}, in.err
	}

	return portal, nil
}

// levelType applies the setlvl branch of loadsave.cpp:1820.
func (p PortalRecord) levelType() int32 {
	if p.SetLevel {
		return p.LType
	}

	return HellfireLevelType(p.LType)
}

// HellfireLevelType is getHellfireLevelType (loadsave.cpp:1148-1157): crypt
// reads as cathedral and nest as caves for vanilla compatibility.
func HellfireLevelType(levelType int32) int32 {
	switch levelType {
	case 5: // DTYPE_CRYPT
		return 1 // DTYPE_CATHEDRAL
	case 6: // DTYPE_NEST
		return 3 // DTYPE_CAVES
	default:
		return levelType
	}
}

// Encode returns the SaveObject image with the dynamic-light var1 zeroing
// (loadsave.cpp:1749-1763). The type is written unchanged; use EncodeForLevel
// on nest/crypt levels.
func (o ObjectRecord) Encode() []byte {
	return o.encodeWithType(o.Type)
}

// EncodeForLevel returns the SaveObject image with the object type remapped
// through ConvertToHellfireObject for the given level type
// (loadsave.cpp:1664-1699, applied at SaveObject :1705).
func (o ObjectRecord) EncodeForLevel(levelType int32) []byte {
	return o.encodeWithType(convertToHellfireObject(o.Type, levelType))
}

// encodeWithType writes the record with the given written type and the dynamic-light
// var1 zeroing (loadsave.cpp:1703-1746, :1749-1763).
func (o ObjectRecord) encodeWithType(writtenType int32) []byte {
	var1 := o.Var1
	if o.IsDynamicLight && var1 != -1 {
		var1 = 0
	}

	var out recordWriter

	out.i32(writtenType)
	out.i32(o.X)
	out.i32(o.Y)
	out.u32(boolField{set: o.ApplyLighting}.image())
	out.u32(boolField{set: o.AnimFlag}.image())
	out.zeroes(4) // _oAnimData pointer
	out.i32(o.AnimDelay)
	out.i32(o.AnimCnt)
	out.u32(o.AnimLen)
	out.u32(o.AnimFrame)
	out.i32(o.AnimWidth)
	out.i32(calculateSpriteTileCenterX(o.AnimWidth))
	out.u32(boolField{set: o.DelFlag}.image())
	out.u8(uint8(o.Break))
	out.zeroes(3) // alignment
	out.u32(boolField{set: o.SolidFlag}.image())
	out.u32(boolField{set: o.MissFlag}.image())
	out.u8(o.SelectionRegion)
	out.zeroes(3) // alignment
	out.u32(boolField{set: o.PreFlag}.image())
	out.u32(boolField{set: o.TrapFlag}.image())
	out.u32(boolField{set: o.DoorFlag}.image())
	out.i32(o.OLid)
	out.u32(o.ORndSeed)
	out.i32(var1)
	out.i32(o.Var2)
	out.i32(o.Var3)
	out.i32(o.Var4)
	out.i32(o.Var5)
	out.u32(o.Var6)
	out.i32(o.BookMessage)
	out.i32(o.Var8)

	return out.buf
}

// DecodeObjectRecord parses a SaveObject image.
func DecodeObjectRecord(buf []byte) (ObjectRecord, error) {
	in := recordReader{buf: buf}

	var object ObjectRecord

	object.Type = int32(in.le32())
	object.X = int32(in.le32())
	object.Y = int32(in.le32())
	object.ApplyLighting = in.le32() != 0
	object.AnimFlag = in.le32() != 0
	in.skip(4)
	object.AnimDelay = int32(in.le32())
	object.AnimCnt = int32(in.le32())
	object.AnimLen = in.le32()
	object.AnimFrame = in.le32()
	object.AnimWidth = int32(in.le32())
	in.skip(4) // width2
	object.DelFlag = in.le32() != 0
	object.Break = int8(in.u8())
	in.skip(3)
	object.SolidFlag = in.le32() != 0
	object.MissFlag = in.le32() != 0
	object.SelectionRegion = in.u8()
	in.skip(3)
	object.PreFlag = in.le32() != 0
	object.TrapFlag = in.le32() != 0
	object.DoorFlag = in.le32() != 0
	object.OLid = int32(in.le32())
	object.ORndSeed = in.le32()
	object.Var1 = int32(in.le32())
	object.Var2 = int32(in.le32())
	object.Var3 = int32(in.le32())
	object.Var4 = int32(in.le32())
	object.Var5 = int32(in.le32())
	object.Var6 = in.le32()
	object.BookMessage = int32(in.le32())
	object.Var8 = int32(in.le32())

	if in.err != nil {
		return ObjectRecord{}, in.err
	}

	return object, nil
}

// MonsterKillCountsBlock writes the kill-count walk section: one big-endian
// int32 per MonstersData entry, zero-padded to MaxMonsters slots for vanilla
// save compatibility (loadsave.cpp:2839-2843, monster.h:38).
func MonsterKillCountsBlock(counts []int32) []byte {
	var out recordWriter

	for _, count := range counts {
		out.be32(uint32(count))
	}

	out.zeroes(4 * (MaxMonsters - len(counts)))

	return out.buf
}

// PremiumTailBlock writes the walk tail: the BE premium count and level, then
// the mode's premium SaveItem records (loadsave.cpp:2918-2921). Upstream
// always writes giNumberOfSmithPremiumItems records — 6 without and 15 with
// Hellfire (:2484, :2489) — indexing the fixed PremiumItems array, so the
// caller must supply exactly that many records.
func PremiumTailBlock(count, level int32, items [][]byte, mode GameMode) ([]byte, error) {
	itemSize := SaveItemSize(mode)
	premiumCount := LevelCountsFor(mode).SmithPremium
	if len(items) != premiumCount {
		return nil, fmt.Errorf("persist: premium items %d != %d fixed records: %w", len(items), premiumCount,
			ErrShortRecord)
	}

	var out recordWriter

	out.be32(uint32(count))
	out.be32(uint32(level))

	for i, item := range items {
		if len(item) != itemSize {
			return nil, fmt.Errorf("persist: premium item %d len %d != %d: %w", i, len(item), itemSize,
				ErrShortRecord)
		}

		out.bytes(item)
	}

	return out.buf, nil
}

// calculateSpriteTileCenterX is CalculateSpriteTileCenterX
// (Source/levels/dun_tile.hpp:137-140): the sprite's centered x offset in a
// tile, written as _oAnimWidth2 for vanilla compatibility (loadsave.cpp:1716).
func calculateSpriteTileCenterX(width int32) int32 {
	return (width - tileWidth) / 2
}

// convertToHellfireObject is ConvertToHellfireObject (loadsave.cpp:1664-1699):
// nest POD/PODEX and crypt urn/L5 types read as their vanilla-compatible ids.
func convertToHellfireObject(objectType, levelType int32) int32 {
	if levelType == 5 { // DTYPE_NEST
		return convertNestObject(objectType)
	}

	if levelType == 6 { // DTYPE_CRYPT
		return convertCryptObject(objectType)
	}

	return objectType
}

// convertNestObject is the DTYPE_NEST switch of loadsave.cpp:1666-1677.
func convertNestObject(objectType int32) int32 {
	switch objectType {
	case 94: // OBJ_POD
		return 58 // OBJ_BARREL
	case 95: // OBJ_PODEX
		return 59 // OBJ_BARRELEX
	default:
		return objectType
	}
}

// convertCryptObject is the DTYPE_CRYPT switch of loadsave.cpp:1679-1697.
func convertCryptObject(objectType int32) int32 {
	switch objectType {
	case 96: // OBJ_URN
		return 58 // OBJ_BARREL
	case 97: // OBJ_URNEX
		return 59 // OBJ_BARRELEX
	case 98: // OBJ_L5BOOKS
		return 83 // OBJ_STORYBOOK
	case 99: // OBJ_L5CANDLE
		return 84 // OBJ_STORYCANDLE
	case 100: // OBJ_L5LDOOR
		return 1 // OBJ_L1LDOOR
	case 101: // OBJ_L5RDOOR
		return 2 // OBJ_L1RDOOR
	case 102: // OBJ_L5LEVER
		return 4 // OBJ_LEVER
	case 103: // OBJ_L5SARC
		return 48 // OBJ_SARC
	default:
		return objectType
	}
}
