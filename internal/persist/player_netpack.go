package persist

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// Validation indices of PlayerPack.NetValidation, in the pack.h:130-146 order.
const (
	ValStrength = iota
	ValMagic
	ValDexterity
	ValVitality
	ValHitPoints
	ValMaxHP
	ValMana
	ValMaxMana
	ValDamageMod
	ValBaseToBlk
	ValIMinDam
	ValIMaxDam
	ValIAC
	ValIBonusDam
	ValIBonusToHit
	ValIBonusAC
	ValIBonusDamMod
	ValIGetHit
	ValIEnAc
	ValIFMinDam
	ValIFMaxDam
	ValILMinDam
	ValILMaxDam
	// NumNetValidation is the number of validation words.
	NumNetValidation = 23
)

// PlayerNetPack is struct PlayerNetPack (Source/pack.h:90-146): the player
// summary exchanged when a multiplayer game starts.
type PlayerNetPack struct {
	PlrLevel     uint8
	Px           uint8
	Py           uint8
	Dir          uint8
	Name         [32]byte
	Class        uint8
	BaseStr      uint8
	BaseMag      uint8
	BaseDex      uint8
	BaseVit      uint8
	Level        int8
	StatPts      uint8
	Exp          uint32
	HPBase       int32
	MaxHPBase    int32
	ManaBase     int32
	MaxManaBase  int32
	SplLvl       [NetSpellLevels]uint8
	MemSpells    uint64
	InvBody      [7]ItemNetPack
	InvList      [40]ItemNetPack
	InvGrid      [40]int8
	NumInv       uint8
	SpdList      [8]ItemNetPack
	ManaShield   uint8
	Reflections  uint16
	KillLevel    uint8
	FriendlyMode uint8
	IsOnSetLevel uint8
	// Validation carries the pStrength..pILMaxDam words (pack.h:130-146),
	// indexed by the Val* constants.
	Validation [NumNetValidation]int32
}

// NetPlayerExtras are the player fields PackNetPlayer needs that
// entities.Player does not carry: the character level, friendlyMode and the
// playerdat-derived base block chance.
type NetPlayerExtras struct {
	// CharLevel is player.getCharacterLevel().
	CharLevel uint8
	// FriendlyMode is player.friendlyMode (pack.cpp:260).
	FriendlyMode bool
	// BaseToBlk is player.getBaseToBlock() (pack.cpp:272).
	BaseToBlk int32
}

// Encode returns the 1692-byte net-pack image.
func (p PlayerNetPack) Encode() []byte {
	var out recordWriter

	p.encodeHead(&out)
	p.encodeItems(&out)
	p.encodeTail(&out)

	return out.buf
}

// DecodePlayerNetPack reads a net-pack image produced by Encode.
func DecodePlayerNetPack(buf []byte) (PlayerNetPack, error) {
	var packed PlayerNetPack

	in := recordReader{buf: buf}
	packed.decodeHead(&in)
	packed.decodeItems(&in)
	packed.decodeTail(&in)

	if in.err != nil {
		return PlayerNetPack{}, in.err
	}

	return packed, nil
}

// PackNetPlayer packs the player summary (Source/pack.cpp:221-289).
func PackNetPlayer(player entities.Player, extras NetPlayerExtras, env PackEnv) PlayerNetPack {
	var packed PlayerNetPack

	packed.PlrLevel = player.PlrLevel
	packed.Px = uint8(player.Position.Tile.X)
	packed.Py = uint8(player.Position.Tile.Y)
	packed.Dir = player.Direction
	copy(packed.Name[:], player.Name[:])
	packed.Class = player.Class
	packed.BaseStr = uint8(player.BaseStr)
	packed.BaseMag = uint8(player.BaseMag)
	packed.BaseDex = uint8(player.BaseDex)
	packed.BaseVit = uint8(player.BaseVit)
	packed.Level = int8(extras.CharLevel)
	packed.StatPts = uint8(player.StatPts)
	packed.Exp = player.Experience
	packed.HPBase = int32(player.HPBase)
	packed.MaxHPBase = int32(player.MaxHPBase)
	packed.ManaBase = int32(player.ManaBase)
	packed.MaxManaBase = int32(player.MaxManaBase)
	packed.MemSpells = player.MemSpells

	for i := range packed.SplLvl {
		packed.SplLvl[i] = player.SplLvl[i]
	}

	packNetItemBlock(packed.InvBody[:], player.InvBody[:], env)

	packed.NumInv = uint8(player.NumInv)
	packNetItemBlock(packed.InvList[:], player.InvList[:int(packed.NumInv)], env)
	packed.InvGrid = player.InvGrid
	packNetItemBlock(packed.SpdList[:], player.SpdList[:], env)

	fillNetPlayerTail(&packed, player, extras)

	return packed
}

// packNetItemBlock fills a net-pack item array (pack.cpp:239-256).
func packNetItemBlock(dst []ItemNetPack, src []entities.Item, env PackEnv) {
	for i := range dst {
		if i >= len(src) {
			dst[i] = ItemNetPack{Idx: -1}
			continue
		}

		dst[i] = PackNetItem(src[i], NetItemStats{}, env)
	}
}

// fillNetPlayerTail fills the shared tail and validation block
// (pack.cpp:258-289).
func fillNetPlayerTail(packed *PlayerNetPack, player entities.Player, extras NetPlayerExtras) {
	packed.Reflections = player.Reflections
	packed.KillLevel = player.DiabloKillLevel
	if player.ManaShield {
		packed.ManaShield = 1
	}

	if extras.FriendlyMode {
		packed.FriendlyMode = 1
	}

	fillNetValidation(&packed.Validation, player, extras)
}

// fillNetValidation fills the pStrength..pILMaxDam words (pack.cpp:264-288).
func fillNetValidation(validation *[NumNetValidation]int32, player entities.Player, extras NetPlayerExtras) {
	validation[ValStrength] = int32(player.Strength)
	validation[ValMagic] = int32(player.Magic)
	validation[ValDexterity] = int32(player.Dexterity)
	validation[ValVitality] = int32(player.Vitality)
	validation[ValHitPoints] = int32(player.HitPoints)
	validation[ValMaxHP] = int32(player.MaxHP)
	validation[ValMana] = int32(player.Mana)
	validation[ValMaxMana] = int32(player.MaxMana)
	validation[ValDamageMod] = int32(player.DamageMod)
	validation[ValBaseToBlk] = extras.BaseToBlk
	validation[ValIMinDam] = int32(player.IMinDam)
	validation[ValIMaxDam] = int32(player.IMaxDam)
	validation[ValIAC] = int32(player.IAC)
	validation[ValIBonusDam] = int32(player.IBonusDam)
	validation[ValIBonusToHit] = int32(player.IBonusToHit)
	validation[ValIBonusAC] = int32(player.IBonusAC)
	validation[ValIBonusDamMod] = int32(player.IBonusDamMod)
	validation[ValIGetHit] = int32(player.IGetHit)
	validation[ValIEnAc] = int32(player.IEnAc)
	validation[ValIFMinDam] = int32(player.IFMinDam)
	validation[ValIFMaxDam] = int32(player.IFMaxDam)
	validation[ValILMinDam] = int32(player.ILMinDam)
	validation[ValILMaxDam] = int32(player.ILMaxDam)
}
